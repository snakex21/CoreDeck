package update

// Installation in two phases, so that an interrupted download never breaks
// the program:
//
//  1. Prepare downloads SHA256SUMS.txt and CoreDeck-Portable.zip into
//     <exeDir>\.coredeck-update\staging\, checks the archive size against the
//     API asset size and its SHA-256 against SHA256SUMS.txt, then extracts
//     ONLY the managed files (CoreDeck.exe, data/tools/disk-helper.exe) and
//     checks the extracted CoreDeck.exe against its own SHA256SUMS.txt line.
//     Any unsafe entry name (.., absolute, drive letter) rejects the whole
//     archive. On any error the staging directory is removed and no program
//     file is touched.
//  2. Commit renames every target to <name>.old and moves the staged file in
//     (MoveFileEx, same volume; a running .exe may be renamed). Any failure
//     undoes every earlier step.

import (
	"archive/zip"
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	WorkDirName     = ".coredeck-update"
	ExeName         = "CoreDeck.exe"
	HelperRel       = "data/tools/disk-helper.exe"
	OldSuffix       = ".old"
	RollbackSuffix  = ".rollback-tmp"
	MaxEntrySize    = 150 << 20
	BackupsToKeep   = 2
	backupMetaName  = "backup.json"
	stagingDirName  = "staging"
	backupDirName   = "backup"
	stateFileName   = "state.json"
	markerFileName  = "health.json"
	guardLogName    = "guard.log"
	maxGuardLogSize = 256 << 10
)

// ManagedFiles are the only files an update may replace (slash separated,
// relative to exeDir). The first one is mandatory.
var ManagedFiles = []string{ExeName, HelperRel}

func WorkDir(exeDir string) string    { return filepath.Join(exeDir, WorkDirName) }
func StagingDir(exeDir string) string { return filepath.Join(WorkDir(exeDir), stagingDirName) }
func BackupRoot(exeDir string) string { return filepath.Join(WorkDir(exeDir), backupDirName) }
func StatePath(exeDir string) string  { return filepath.Join(WorkDir(exeDir), stateFileName) }
func MarkerPath(exeDir string) string { return filepath.Join(WorkDir(exeDir), markerFileName) }

func target(exeDir, rel string) string { return filepath.Join(exeDir, filepath.FromSlash(rel)) }
func staged(exeDir, rel string) string {
	return filepath.Join(StagingDir(exeDir), filepath.FromSlash(rel))
}

// RemoveStaging deletes the staging directory (errors ignored).
func RemoveStaging(exeDir string) { _ = os.RemoveAll(StagingDir(exeDir)) }

// ---------------------------------------------------------------------------
// SHA256SUMS.txt
// ---------------------------------------------------------------------------

// Sums maps a lower-case, slash separated file name to its lower-case hex
// SHA-256.
type Sums map[string]string

// ParseSums reads "<sha256>  <name>" / "<sha256> *<name>" lines.
func ParseSums(text string) (Sums, error) {
	out := Sums{}
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(sc.Text(), "\ufeff"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, newErr(CodeBadChecksum, "malformed SHA256SUMS line %q", line)
		}
		sum := strings.ToLower(fields[0])
		if len(sum) != 64 {
			return nil, newErr(CodeBadChecksum, "malformed SHA256SUMS line %q", line)
		}
		if _, err := hex.DecodeString(sum); err != nil {
			return nil, newErr(CodeBadChecksum, "malformed SHA256SUMS line %q", line)
		}
		name := strings.TrimSpace(line[len(fields[0]):])
		name = strings.TrimPrefix(name, "*")
		key := sumsKey(name)
		if prev, ok := out[key]; ok && prev != sum {
			return nil, newErr(CodeBadChecksum, "conflicting SHA256SUMS entries for %s", name)
		}
		out[key] = sum
	}
	if err := sc.Err(); err != nil {
		return nil, newErr(CodeBadChecksum, "cannot read SHA256SUMS: %v", err)
	}
	return out, nil
}

func sumsKey(name string) string {
	return strings.ToLower(strings.TrimPrefix(strings.ReplaceAll(strings.TrimSpace(name), `\`, "/"), "./"))
}

// Lookup returns the checksum for the first name that is listed.
func (s Sums) Lookup(names ...string) (string, bool) {
	for _, n := range names {
		if v, ok := s[sumsKey(n)]; ok {
			return v, true
		}
	}
	return "", false
}

// ---------------------------------------------------------------------------
// Archive entries
// ---------------------------------------------------------------------------

// ErrUnsafePath rejects an archive containing a dangerous entry name.
var ErrUnsafePath = errors.New("unsafe path in update archive")

// NormalizeEntry turns a zip entry name into a clean, relative, slash
// separated path. Backslashes are accepted as separators (Compress-Archive
// writes them). It returns isDir for directory entries.
func NormalizeEntry(name string) (clean string, isDir bool, err error) {
	if name == "" || len(name) > 260 {
		return "", false, ErrUnsafePath
	}
	n := strings.ReplaceAll(name, `\`, "/")
	for i := 0; i < len(n); i++ {
		if c := n[i]; c < 0x20 || c == 0x7f || c == ':' {
			return "", false, ErrUnsafePath // control chars, drive letters, ADS
		}
	}
	if strings.HasPrefix(n, "/") {
		return "", false, ErrUnsafePath
	}
	if strings.HasSuffix(n, "/") {
		isDir = true
		n = strings.TrimSuffix(n, "/")
	}
	if n == "" {
		return "", false, ErrUnsafePath
	}
	for _, part := range strings.Split(n, "/") {
		if part == "" || part == "." || part == ".." {
			return "", false, ErrUnsafePath
		}
		// Windows strips trailing dots and spaces: "CoreDeck.exe." == "CoreDeck.exe".
		if last := part[len(part)-1]; last == '.' || last == ' ' {
			return "", false, ErrUnsafePath
		}
	}
	return n, isDir, nil
}

func managedIndex(clean string) int {
	for i, m := range ManagedFiles {
		if strings.EqualFold(m, clean) {
			return i
		}
	}
	return -1
}

// ExtractManaged extracts only ManagedFiles from zipPath into destDir.
func ExtractManaged(zipPath, destDir string) ([]string, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil && !errors.Is(err, zip.ErrInsecurePath) {
		return nil, newErr(CodeInstall, "cannot open update archive: %v", err)
	}
	if zr == nil {
		return nil, newErr(CodeInstall, "cannot open update archive: %v", err)
	}
	defer zr.Close()

	seen := make([]bool, len(ManagedFiles))
	type pick struct {
		f   *zip.File
		idx int
	}
	var picks []pick
	for _, f := range zr.File {
		clean, isDir, err := NormalizeEntry(f.Name)
		if err != nil {
			return nil, newErr(CodeInstall, "%v: %q", ErrUnsafePath, f.Name)
		}
		if isDir {
			continue
		}
		idx := managedIndex(clean)
		if idx < 0 {
			continue
		}
		if seen[idx] {
			return nil, newErr(CodeInstall, "duplicate entry %q in update archive", f.Name)
		}
		seen[idx] = true
		if f.UncompressedSize64 > MaxEntrySize {
			return nil, newErr(CodeBadSize, "entry %q is too large", f.Name)
		}
		picks = append(picks, pick{f, idx})
	}
	if !seen[0] {
		return nil, newErr(CodeNoAsset, "%s is missing from the update archive", ExeName)
	}
	var out []string
	for _, p := range picks {
		rel := ManagedFiles[p.idx]
		if err := extractOne(p.f, filepath.Join(destDir, filepath.FromSlash(rel))); err != nil {
			return nil, err
		}
		out = append(out, rel)
	}
	return out, nil
}

func extractOne(f *zip.File, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return wrapErr(CodeInstall, err)
	}
	rc, err := f.Open()
	if err != nil {
		return newErr(CodeInstall, "cannot read %q: %v", f.Name, err)
	}
	defer rc.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
	if err != nil {
		return wrapErr(CodeInstall, err)
	}
	n, err := io.Copy(out, io.LimitReader(rc, MaxEntrySize+1))
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return newErr(CodeInstall, "cannot extract %q: %v", f.Name, err)
	}
	if n > MaxEntrySize || uint64(n) != f.UncompressedSize64 {
		return newErr(CodeBadSize, "entry %q has an unexpected size", f.Name)
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ---------------------------------------------------------------------------
// Phase 1: prepare
// ---------------------------------------------------------------------------

// Progress callback: stage is "downloading" or "verifying", percent 0..100
// within that stage.
type ProgressFunc func(stage string, percent int)

// PrepareOptions configures Prepare.
type PrepareOptions struct {
	ExeDir   string
	Release  *Release
	Client   *http.Client
	Policy   Policy
	Progress ProgressFunc
}

// Prepare downloads, verifies and stages a release. On error the staging
// directory is removed and nothing else is changed.
func Prepare(ctx context.Context, o PrepareOptions) (files []string, err error) {
	progress := o.Progress
	if progress == nil {
		progress = func(string, int) {}
	}
	if o.Release == nil {
		return nil, newErr(CodeNoAsset, "no release selected")
	}
	zipAsset, ok := o.Release.Asset(AssetZip)
	if !ok {
		return nil, newErr(CodeNoAsset, "release has no %s", AssetZip)
	}
	sumsAsset, ok := o.Release.Asset(AssetSums)
	if !ok {
		return nil, newErr(CodeNoAsset, "release has no %s", AssetSums)
	}
	if zipAsset.Size <= 0 || zipAsset.Size > MaxDownloadSize {
		return nil, newErr(CodeBadSize, "%s has an invalid size (%d bytes)", AssetZip, zipAsset.Size)
	}
	if sumsAsset.Size > MaxSumsSize {
		return nil, newErr(CodeBadSize, "%s is too large", AssetSums)
	}
	if err := o.Policy.CheckAssetURL(zipAsset.BrowserDownloadURL); err != nil {
		return nil, err
	}
	if err := o.Policy.CheckAssetURL(sumsAsset.BrowserDownloadURL); err != nil {
		return nil, err
	}

	RemoveStaging(o.ExeDir)
	stage := StagingDir(o.ExeDir)
	if err := os.MkdirAll(stage, 0o755); err != nil {
		return nil, wrapErr(CodeInstall, err)
	}
	defer func() {
		if err != nil {
			RemoveStaging(o.ExeDir)
		}
	}()

	progress("downloading", 0)
	var sumsBuf strings.Builder
	expectSums := int64(-1)
	if sumsAsset.Size > 0 {
		expectSums = sumsAsset.Size
	}
	if _, err := o.Policy.download(ctx, o.Client, sumsAsset.BrowserDownloadURL, MaxSumsSize, expectSums, &sumsBuf, nil); err != nil {
		return nil, err
	}
	sums, err := ParseSums(sumsBuf.String())
	if err != nil {
		return nil, err
	}
	wantZip, ok := sums.Lookup(AssetZip)
	if !ok {
		return nil, newErr(CodeBadChecksum, "%s is not listed in %s", AssetZip, AssetSums)
	}

	zipPath := filepath.Join(stage, AssetZip)
	zf, err := os.OpenFile(zipPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, wrapErr(CodeInstall, err)
	}
	h := sha256.New()
	_, err = o.Policy.download(ctx, o.Client, zipAsset.BrowserDownloadURL, MaxDownloadSize, zipAsset.Size,
		io.MultiWriter(zf, h), func(done, total int64) {
			if total > 0 {
				progress("downloading", int(done*100/total))
			}
		})
	if cerr := zf.Close(); err == nil && cerr != nil {
		err = wrapErr(CodeInstall, cerr)
	}
	if err != nil {
		return nil, err
	}

	progress("verifying", 0)
	if got := hex.EncodeToString(h.Sum(nil)); got != wantZip {
		return nil, newErr(CodeBadChecksum, "%s checksum mismatch", AssetZip)
	}
	files, err = ExtractManaged(zipPath, stage)
	if err != nil {
		return nil, err
	}
	_ = os.Remove(zipPath)
	progress("verifying", 50)

	wantExe, ok := sums.Lookup(ExeName, "CoreDeck-Portable/"+ExeName)
	if !ok {
		return nil, newErr(CodeBadChecksum, "%s is not listed in %s", ExeName, AssetSums)
	}
	if got, err := fileSHA256(staged(o.ExeDir, ExeName)); err != nil || got != wantExe {
		return nil, newErr(CodeBadChecksum, "%s checksum mismatch", ExeName)
	}
	if want, ok := sums.Lookup(HelperRel, filepath.Base(HelperRel)); ok {
		if _, statErr := os.Stat(staged(o.ExeDir, HelperRel)); statErr == nil {
			if got, err := fileSHA256(staged(o.ExeDir, HelperRel)); err != nil || got != want {
				return nil, newErr(CodeBadChecksum, "%s checksum mismatch", filepath.Base(HelperRel))
			}
		}
	}
	progress("verifying", 100)
	return files, nil
}

// ---------------------------------------------------------------------------
// Phase 2: commit
// ---------------------------------------------------------------------------

// CommittedFile records one swapped file; HadOld is false when the file did
// not exist before (rollback then removes it).
type CommittedFile struct {
	Rel    string `json:"rel"`
	HadOld bool   `json:"hadOld"`
}

// commitHook is a test seam called after <rel>.old was made and before the
// new file is moved in.
var commitHook func(rel string) error

// Commit swaps the staged files in. Any error restores every earlier step.
func Commit(exeDir string) ([]CommittedFile, error) {
	var todo []string
	for _, rel := range ManagedFiles {
		if fi, err := os.Stat(staged(exeDir, rel)); err == nil && fi.Mode().IsRegular() {
			todo = append(todo, rel)
		}
	}
	if len(todo) == 0 || !strings.EqualFold(todo[0], ExeName) {
		return nil, newErr(CodeInstall, "nothing staged for installation")
	}
	// Pre-flight: stale .old copies must be removable before anything moves
	// (e.g. a guard from an earlier update may still run from CoreDeck.exe.old).
	for _, rel := range todo {
		old := target(exeDir, rel) + OldSuffix
		if err := os.Remove(old); err != nil && !os.IsNotExist(err) {
			return nil, newErr(CodeInstall, "previous update files are still in use (%s): %v", filepath.Base(old), err)
		}
	}

	var done []CommittedFile
	undo := func() {
		for i := len(done) - 1; i >= 0; i-- {
			c := done[i]
			t := target(exeDir, c.Rel)
			if err := moveFile(t, staged(exeDir, c.Rel)); err != nil {
				_ = os.Remove(t)
			}
			if c.HadOld {
				_ = moveFile(t+OldSuffix, t)
			}
		}
	}
	for _, rel := range todo {
		t := target(exeDir, rel)
		c := CommittedFile{Rel: rel}
		if err := os.MkdirAll(filepath.Dir(t), 0o755); err != nil {
			undo()
			return nil, wrapErr(CodeInstall, err)
		}
		if _, err := os.Lstat(t); err == nil {
			if err := moveFile(t, t+OldSuffix); err != nil {
				undo()
				return nil, newErr(CodeInstall, "cannot rename %s: %v", filepath.Base(t), err)
			}
			c.HadOld = true
		}
		var err error
		if commitHook != nil {
			err = commitHook(rel)
		}
		if err == nil {
			err = moveFile(staged(exeDir, rel), t)
		}
		if err != nil {
			if c.HadOld {
				_ = moveFile(t+OldSuffix, t)
			}
			undo()
			return nil, newErr(CodeInstall, "cannot install %s: %v", filepath.Base(t), err)
		}
		done = append(done, c)
	}
	RemoveStaging(exeDir)
	return done, nil
}

// RestoreOld undoes a committed swap: every file goes back to its .old copy;
// files that did not exist before are moved away. It continues past errors
// and returns the first one.
func RestoreOld(exeDir string, files []CommittedFile) error {
	var first error
	keep := func(err error) {
		if err != nil && first == nil {
			first = err
		}
	}
	for i := len(files) - 1; i >= 0; i-- {
		c := files[i]
		if managedIndex(c.Rel) < 0 {
			continue // never trust a path from the state file blindly
		}
		t := target(exeDir, ManagedFiles[managedIndex(c.Rel)])
		tmp := t + RollbackSuffix
		_ = os.Remove(tmp)
		_, statErr := os.Lstat(t)
		hadCurrent := statErr == nil
		if !c.HadOld {
			if hadCurrent {
				keep(moveFile(t, tmp))
				_ = os.Remove(tmp)
			}
			continue
		}
		if _, err := os.Lstat(t + OldSuffix); err != nil {
			keep(fmt.Errorf("missing %s: %w", filepath.Base(t+OldSuffix), err))
			continue
		}
		if hadCurrent {
			if err := moveFile(t, tmp); err != nil {
				keep(err)
				continue
			}
		}
		if err := moveFile(t+OldSuffix, t); err != nil {
			if hadCurrent {
				_ = moveFile(tmp, t)
			}
			keep(err)
			continue
		}
		_ = os.Remove(tmp) // a still running .exe cannot be deleted; CleanupStale retries
	}
	return first
}

// CleanupStale removes leftovers of an earlier update (.old, .rollback-tmp,
// staging) unless a guard may still be using them. Errors are ignored.
func CleanupStale(exeDir string, now time.Time) {
	st, _ := LoadState(StatePath(exeDir))
	if st.Pending != nil && !st.Pending.Stale(now) {
		return
	}
	if st.Pending != nil {
		st.Pending = nil
		_ = SaveState(StatePath(exeDir), st)
	}
	for _, rel := range ManagedFiles {
		t := target(exeDir, rel)
		_ = os.Remove(t + OldSuffix)
		_ = os.Remove(t + RollbackSuffix)
	}
	RemoveStaging(exeDir)
}

// ---------------------------------------------------------------------------
// Backups
// ---------------------------------------------------------------------------

// BackupInfo describes one backup directory.
type BackupInfo struct {
	Version   string `json:"version"`
	CreatedAt string `json:"createdAt"`
	Dir       string `json:"-"`
}

// CreateBackup copies the current managed files to backup\<version>\ and
// keeps only the newest BackupsToKeep backups.
func CreateBackup(exeDir, version string, now time.Time) error {
	v, err := ParseVersion(version)
	if err != nil {
		return newErr(CodeInstall, "invalid current version %q", version)
	}
	dir := filepath.Join(BackupRoot(exeDir), v.String())
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return wrapErr(CodeInstall, err)
	}
	for i, rel := range ManagedFiles {
		src := target(exeDir, rel)
		if _, err := os.Stat(src); err != nil {
			if i == 0 {
				_ = os.RemoveAll(dir)
				return newErr(CodeInstall, "cannot back up %s: %v", rel, err)
			}
			continue
		}
		if err := copyFile(src, filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			_ = os.RemoveAll(dir)
			return newErr(CodeInstall, "cannot back up %s: %v", rel, err)
		}
	}
	meta := BackupInfo{Version: v.String(), CreatedAt: now.UTC().Format(time.RFC3339Nano)}
	if err := writeJSONAtomic(filepath.Join(dir, backupMetaName), meta); err != nil {
		_ = os.RemoveAll(dir)
		return wrapErr(CodeInstall, err)
	}
	backups := ListBackups(exeDir)
	for i := BackupsToKeep; i < len(backups); i++ {
		_ = os.RemoveAll(backups[i].Dir)
	}
	return nil
}

// ListBackups returns valid backups, newest first.
func ListBackups(exeDir string) []BackupInfo {
	root := BackupRoot(exeDir)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var out []BackupInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		var meta BackupInfo
		if readJSON(filepath.Join(dir, backupMetaName), &meta) != nil {
			continue
		}
		v, err := ParseVersion(meta.Version)
		if err != nil || v.String() != e.Name() {
			continue
		}
		if fi, err := os.Stat(filepath.Join(dir, ExeName)); err != nil || !fi.Mode().IsRegular() {
			continue
		}
		meta.Version = v.String()
		meta.Dir = dir
		out = append(out, meta)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return out
}

// LatestBackup returns the newest backup of a version other than current.
func LatestBackup(exeDir, current string) (BackupInfo, bool) {
	for _, b := range ListBackups(exeDir) {
		if c, ok := CompareStrings(b.Version, current); ok && c == 0 {
			continue
		}
		return b, true
	}
	return BackupInfo{}, false
}

// StageBackup copies a backup into the staging directory for Commit.
func StageBackup(exeDir string, b BackupInfo) error {
	RemoveStaging(exeDir)
	for i, rel := range ManagedFiles {
		src := filepath.Join(b.Dir, filepath.FromSlash(rel))
		if _, err := os.Stat(src); err != nil {
			if i == 0 {
				RemoveStaging(exeDir)
				return newErr(CodeInstall, "backup is incomplete: %v", err)
			}
			continue
		}
		if err := copyFile(src, staged(exeDir, rel)); err != nil {
			RemoveStaging(exeDir)
			return wrapErr(CodeInstall, err)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp := dst + ".partial"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if err == nil {
		err = out.Sync()
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = moveFile(tmp, dst)
	}
	if err != nil {
		_ = os.Remove(tmp)
	}
	return err
}
