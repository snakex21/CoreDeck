package update

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func prepare(t *testing.T, dir string, f *fakeRelease) ([]string, error) {
	t.Helper()
	return Prepare(t.Context(), PrepareOptions{ExeDir: dir, Release: f.release(t), Client: f.srv.Client(), Policy: f.policy()})
}

func TestPrepareStagesOnlyManagedFiles(t *testing.T) {
	dir := newAppDir(t)
	before := snapshot(t, dir)
	z := goodZip(t)
	f := newFakeRelease(t, z, sumsFor(z))

	files, err := prepare(t, dir, f)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(files, ",") != "CoreDeck.exe,data/tools/disk-helper.exe" {
		t.Fatalf("files = %v", files)
	}
	if readFile(t, staged(dir, ExeName)) != string(newExe) || readFile(t, staged(dir, HelperRel)) != string(newHelper) {
		t.Fatal("staged content mismatch")
	}
	for _, unwanted := range []string{"release-manifest.json", AssetZip} {
		if exists(filepath.Join(StagingDir(dir), unwanted)) {
			t.Fatalf("%s must not stay in staging", unwanted)
		}
	}
	assertSnapshot(t, before, snapshot(t, dir))
}

func assertAbortedCleanly(t *testing.T, dir string, before map[string]string, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	if ErrorCode(err) != code {
		t.Fatalf("code = %s (%v), want %s", ErrorCode(err), err, code)
	}
	if exists(StagingDir(dir)) {
		t.Fatal("staging directory must be removed")
	}
	assertSnapshot(t, before, snapshot(t, dir))
}

func TestPrepareBadChecksumChangesNothing(t *testing.T) {
	dir := newAppDir(t)
	before := snapshot(t, dir)
	z := goodZip(t)
	sums := strings.Replace(sumsFor(z), sha(z), strings.Repeat("0", 64), 1)
	_, err := prepare(t, dir, newFakeRelease(t, z, sums))
	assertAbortedCleanly(t, dir, before, err, CodeBadChecksum)
}

func TestPrepareZipNotListedInSums(t *testing.T) {
	dir := newAppDir(t)
	before := snapshot(t, dir)
	z := goodZip(t)
	_, err := prepare(t, dir, newFakeRelease(t, z, sha(newExe)+"  CoreDeck.exe\n"))
	assertAbortedCleanly(t, dir, before, err, CodeBadChecksum)
}

func TestPrepareBadExtractedExeChecksum(t *testing.T) {
	dir := newAppDir(t)
	before := snapshot(t, dir)
	z := goodZip(t)
	sums := sha(z) + "  CoreDeck-Portable.zip\n" + sha([]byte("other")) + "  CoreDeck.exe\n"
	_, err := prepare(t, dir, newFakeRelease(t, z, sums))
	assertAbortedCleanly(t, dir, before, err, CodeBadChecksum)

	// The exe line is mandatory.
	_, err = prepare(t, dir, newFakeRelease(t, z, sha(z)+"  CoreDeck-Portable.zip\n"))
	assertAbortedCleanly(t, dir, before, err, CodeBadChecksum)
}

func TestPrepareBadSizeChangesNothing(t *testing.T) {
	dir := newAppDir(t)
	before := snapshot(t, dir)
	z := goodZip(t)
	f := newFakeRelease(t, z, sumsFor(z))
	f.assetSize = int64(len(z)) + 1 // API claims a different size
	_, err := prepare(t, dir, f)
	assertAbortedCleanly(t, dir, before, err, CodeBadSize)

	f.assetSize = MaxDownloadSize + 1
	_, err = prepare(t, dir, f)
	assertAbortedCleanly(t, dir, before, err, CodeBadSize)
}

func TestPrepareMissingAssets(t *testing.T) {
	dir := newAppDir(t)
	before := snapshot(t, dir)
	z := goodZip(t)
	f := newFakeRelease(t, z, sumsFor(z))
	f.releases = func(base string) []map[string]any {
		return []map[string]any{{"tag_name": "v2.5.0", "assets": []map[string]any{
			{"name": AssetSums, "size": 10, "browser_download_url": base + "/dl/" + AssetSums},
		}}}
	}
	_, err := prepare(t, dir, f)
	assertAbortedCleanly(t, dir, before, err, CodeNoAsset)

	// An archive without CoreDeck.exe is rejected too.
	z2 := makeZip(t, zipEntry{`data\tools\disk-helper.exe`, newHelper})
	_, err = prepare(t, dir, newFakeRelease(t, z2, sumsFor(z2)))
	assertAbortedCleanly(t, dir, before, err, CodeNoAsset)
}

func TestPrepareRejectsForeignAssetURL(t *testing.T) {
	dir := newAppDir(t)
	z := goodZip(t)
	f := newFakeRelease(t, z, sumsFor(z))
	f.releases = func(base string) []map[string]any {
		return []map[string]any{{"tag_name": "v2.5.0", "assets": []map[string]any{
			{"name": AssetZip, "size": len(z), "browser_download_url": "https://evil.example/" + AssetZip},
			{"name": AssetSums, "size": 10, "browser_download_url": base + "/dl/" + AssetSums},
		}}}
	}
	if _, err := prepare(t, dir, f); ErrorCode(err) != CodeNoAsset {
		t.Fatalf("expected no_asset, got %v", err)
	}
	if f.hits["/dl/"+AssetSums] != 0 {
		t.Fatal("nothing may be downloaded when an asset URL is rejected")
	}
}

func TestNormalizeEntry(t *testing.T) {
	good := map[string]string{
		"CoreDeck.exe":               "CoreDeck.exe",
		`data\tools\disk-helper.exe`: "data/tools/disk-helper.exe",
		"data/tools/":                "data/tools",
	}
	for in, want := range good {
		got, _, err := NormalizeEntry(in)
		if err != nil || got != want {
			t.Errorf("NormalizeEntry(%q) = %q, %v", in, got, err)
		}
	}
	for _, in := range []string{
		"", `..\CoreDeck.exe`, "../CoreDeck.exe", `data\..\..\x.exe`, "/CoreDeck.exe", `\CoreDeck.exe`,
		`C:\Windows\x.exe`, "C:CoreDeck.exe", "CoreDeck.exe:ads", "data//x", "./CoreDeck.exe",
		"CoreDeck.exe.", "CoreDeck.exe ", "a\x00b",
	} {
		if _, _, err := NormalizeEntry(in); err == nil {
			t.Errorf("NormalizeEntry(%q) accepted", in)
		}
	}
}

func TestZipTraversalRejected(t *testing.T) {
	for _, evil := range []string{`..\..\evil.exe`, "/abs/evil.exe", `C:\evil.exe`, `data\..\..\CoreDeck.exe`} {
		t.Run(evil, func(t *testing.T) {
			dir := newAppDir(t)
			before := snapshot(t, dir)
			z := makeZip(t,
				zipEntry{"CoreDeck.exe", newExe},
				zipEntry{evil, []byte("evil")},
			)
			_, err := prepare(t, dir, newFakeRelease(t, z, sumsFor(z)))
			if err == nil || !strings.Contains(err.Error(), "unsafe path") {
				t.Fatalf("expected unsafe path, got %v", err)
			}
			assertAbortedCleanly(t, dir, before, err, CodeInstall)
			if exists(filepath.Join(filepath.Dir(dir), "evil.exe")) {
				t.Fatal("file written outside the app directory")
			}
		})
	}
}

func TestZipDuplicateManagedEntryRejected(t *testing.T) {
	dir := newAppDir(t)
	z := makeZip(t, zipEntry{"CoreDeck.exe", newExe}, zipEntry{"coredeck.EXE", []byte("second")})
	if _, err := prepare(t, dir, newFakeRelease(t, z, sumsFor(z))); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestParseSums(t *testing.T) {
	s, err := ParseSums("\ufeff" + strings.Repeat("A", 64) + "  CoreDeck-Portable.zip\r\n" +
		strings.Repeat("b", 64) + " *data\\tools\\disk-helper.exe\n# comment\n\n")
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := s.Lookup("coredeck-portable.zip"); !ok || v != strings.Repeat("a", 64) {
		t.Fatalf("zip = %q %v", v, ok)
	}
	if _, ok := s.Lookup(HelperRel); !ok {
		t.Fatal("helper not found")
	}
	for _, bad := range []string{"xyz  file", strings.Repeat("a", 64), strings.Repeat("g", 64) + "  f",
		strings.Repeat("a", 64) + "  f\n" + strings.Repeat("b", 64) + "  f\n"} {
		if _, err := ParseSums(bad); ErrorCode(err) != CodeBadChecksum {
			t.Errorf("ParseSums(%q) = %v", bad, err)
		}
	}
}

func stageNew(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, staged(dir, ExeName), newExe)
	writeFile(t, staged(dir, HelperRel), newHelper)
}

func TestCommitSwapsFiles(t *testing.T) {
	dir := newAppDir(t)
	stageNew(t, dir)
	files, err := Commit(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || !files[0].HadOld || !files[1].HadOld {
		t.Fatalf("files = %#v", files)
	}
	exe := filepath.Join(dir, ExeName)
	helper := target(dir, HelperRel)
	if readFile(t, exe) != string(newExe) || readFile(t, helper) != string(newHelper) {
		t.Fatal("new files not in place")
	}
	if readFile(t, exe+OldSuffix) != string(oldExe) || readFile(t, helper+OldSuffix) != string(oldHelper) {
		t.Fatal(".old copies missing")
	}
	if exists(StagingDir(dir)) {
		t.Fatal("staging must be removed after commit")
	}
	if readFile(t, filepath.Join(dir, "data", "settings.json")) != `{"user":true}` {
		t.Fatal("user data touched")
	}

	// RestoreOld brings the previous version back.
	if err := RestoreOld(dir, files); err != nil {
		t.Fatal(err)
	}
	if readFile(t, exe) != string(oldExe) || readFile(t, helper) != string(oldHelper) || exists(exe+OldSuffix) {
		t.Fatal("restore failed")
	}
}

func TestCommitNewHelperIsRemovedOnRestore(t *testing.T) {
	dir := newAppDir(t)
	_ = os.Remove(target(dir, HelperRel))
	stageNew(t, dir)
	files, err := Commit(dir)
	if err != nil {
		t.Fatal(err)
	}
	if files[1].HadOld {
		t.Fatal("helper did not exist before")
	}
	if err := RestoreOld(dir, files); err != nil {
		t.Fatal(err)
	}
	if exists(target(dir, HelperRel)) {
		t.Fatal("helper added by the update must be removed on rollback")
	}
}

func TestCommitFailureMidwayRollsBack(t *testing.T) {
	dir := newAppDir(t)
	before := snapshot(t, dir)
	stageNew(t, dir)
	commitHook = func(rel string) error {
		if rel == HelperRel {
			return errors.New("injected failure")
		}
		return nil
	}
	defer func() { commitHook = nil }()

	if _, err := Commit(dir); err == nil || ErrorCode(err) != CodeInstall {
		t.Fatalf("expected install error, got %v", err)
	}
	assertSnapshot(t, before, snapshot(t, dir))
	if exists(filepath.Join(dir, ExeName+OldSuffix)) || exists(target(dir, HelperRel)+OldSuffix) {
		t.Fatal(".old files must be gone after rollback")
	}
	// the staged files are back in staging, so nothing was lost
	if readFile(t, staged(dir, ExeName)) != string(newExe) {
		t.Fatal("staged exe lost")
	}
}

func TestCommitRequiresStagedExe(t *testing.T) {
	dir := newAppDir(t)
	writeFile(t, staged(dir, HelperRel), newHelper)
	if _, err := Commit(dir); err == nil {
		t.Fatal("commit without CoreDeck.exe must fail")
	}
}

func TestBackupsKeepLastTwo(t *testing.T) {
	dir := newAppDir(t)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, v := range []string{"2.2.0", "2.3.0", "2.4.0"} {
		writeFile(t, filepath.Join(dir, ExeName), []byte("exe "+v))
		if err := CreateBackup(dir, v, base.Add(time.Duration(i)*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	list := ListBackups(dir)
	if len(list) != 2 || list[0].Version != "2.4.0" || list[1].Version != "2.3.0" {
		t.Fatalf("backups = %#v", list)
	}
	if readFile(t, filepath.Join(list[0].Dir, ExeName)) != "exe 2.4.0" ||
		readFile(t, filepath.Join(list[0].Dir, filepath.FromSlash(HelperRel))) != string(oldHelper) {
		t.Fatal("backup content")
	}
	if b, ok := LatestBackup(dir, "2.4.0"); !ok || b.Version != "2.3.0" {
		t.Fatalf("LatestBackup must skip the running version: %#v", b)
	}
	if b, ok := LatestBackup(dir, "2.5.0"); !ok || b.Version != "2.4.0" {
		t.Fatalf("LatestBackup = %#v", b)
	}
	if err := CreateBackup(dir, "../evil", base); err == nil {
		t.Fatal("invalid version accepted as backup name")
	}
}

func TestCleanupStale(t *testing.T) {
	dir := newAppDir(t)
	now := time.Now()
	old := filepath.Join(dir, ExeName+OldSuffix)
	writeFile(t, old, oldExe)
	writeFile(t, target(dir, HelperRel)+OldSuffix, oldHelper)

	// A running guard (fresh pending swap) keeps the .old files.
	st := State{Pending: &Pending{Mode: ModeUpdate, StartedAt: now.UTC().Format(time.RFC3339Nano)}}
	if err := SaveState(StatePath(dir), st); err != nil {
		t.Fatal(err)
	}
	CleanupStale(dir, now)
	if !exists(old) {
		t.Fatal(".old removed while the guard may still need it")
	}
	// Later the leftovers go away.
	CleanupStale(dir, now.Add(PendingStaleAfter+time.Minute))
	if exists(old) || exists(target(dir, HelperRel)+OldSuffix) {
		t.Fatal("stale .old not removed")
	}
	if s, _ := LoadState(StatePath(dir)); s.Pending != nil {
		t.Fatal("stale pending must be cleared")
	}
}
