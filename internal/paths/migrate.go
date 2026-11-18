package paths

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"launcher/internal/safeio"
	"launcher/internal/system"
)

// driveLetterPath matches a whole JSON string that is an absolute path on a
// drive letter, e.g. `J:\Centrum\App` or `J:/Centrum/App`.
var driveLetterPath = regexp.MustCompile(`^[A-Za-z]:[\\/]`)

// Rewrite is one path changed by MigrateDriveLetters.
type Rewrite struct {
	Old string `json:"old"`
	New string `json:"new"`
}

// MigrateDriveLetters is a one-off repair for catalogues written while the
// pendrive had another letter. In the JSON file it rewrites every string
// that is an absolute path which no longer exists, when the same path does
// exist on CoreDeck's current drive. The new value is relative to exeDir
// ("../Centrum/App"), so the next letter change needs no repair at all.
//
// Paths that exist, paths on the current drive and paths found nowhere are
// left untouched. The file is only rewritten when something changed; the
// original is kept as <file>.drive-<timestamp>.bak.
func MigrateDriveLetters(exeDir, file string) ([]Rewrite, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var tree any
	if err := json.Unmarshal(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")), &tree); err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(file), err)
	}

	driveRoot := system.DriveOf(exeDir) + string(filepath.Separator)
	seen := map[string]bool{}
	var rewrites []Rewrite
	walkStrings(tree, func(s string) {
		if seen[s] || !driveLetterPath.MatchString(s) {
			return
		}
		seen[s] = true
		abs := filepath.Clean(filepath.FromSlash(strings.ReplaceAll(s, `\`, "/")))
		if exists(abs) {
			return
		}
		moved, ok := onDrive(abs, driveRoot)
		if !ok {
			return
		}
		rewrites = append(rewrites, Rewrite{Old: s, New: portableForm(exeDir, moved)})
	})
	if len(rewrites) == 0 {
		return nil, nil
	}

	updated := raw
	for _, rw := range rewrites {
		updated = replaceJSONString(updated, rw.Old, rw.New)
	}
	// Make sure every rewrite landed and the file is still valid JSON.
	var check any
	if err := json.Unmarshal(bytes.TrimPrefix(updated, []byte("\xef\xbb\xbf")), &check); err != nil {
		return nil, fmt.Errorf("%s: migration produced invalid JSON: %w", filepath.Base(file), err)
	}
	remaining := map[string]bool{}
	walkStrings(check, func(s string) { remaining[s] = true })
	for _, rw := range rewrites {
		if remaining[rw.Old] {
			return nil, fmt.Errorf("%s: could not rewrite %q", filepath.Base(file), rw.Old)
		}
	}

	backup := fmt.Sprintf("%s.drive-%s.bak", file, time.Now().Format("20060102-150405"))
	if err := os.WriteFile(backup, raw, 0o644); err != nil {
		return nil, fmt.Errorf("backup %s: %w", backup, err)
	}
	if err := safeio.WriteFileAtomic(file, updated, 0o644); err != nil {
		return nil, err
	}
	return rewrites, nil
}

// portableForm returns target relative to exeDir with forward slashes (the
// catalogue convention), or the absolute path when no relative form exists.
func portableForm(exeDir, target string) string {
	rel, err := filepath.Rel(exeDir, target)
	if err != nil || filepath.IsAbs(rel) {
		return target
	}
	return filepath.ToSlash(rel)
}

func walkStrings(v any, fn func(string)) {
	switch t := v.(type) {
	case string:
		fn(t)
	case []any:
		for _, item := range t {
			walkStrings(item, fn)
		}
	case map[string]any:
		for _, item := range t {
			walkStrings(item, fn)
		}
	}
}

// replaceJSONString swaps the encoded form of old for new. Files written by
// Go escape <, > and & while hand-edited ones may not, so both are tried.
func replaceJSONString(data []byte, oldValue, newValue string) []byte {
	for _, escapeHTML := range []bool{true, false} {
		from := encodeJSONString(oldValue, escapeHTML)
		if bytes.Contains(data, from) {
			data = bytes.ReplaceAll(data, from, encodeJSONString(newValue, escapeHTML))
		}
	}
	return data
}

func encodeJSONString(s string, escapeHTML bool) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(escapeHTML)
	_ = enc.Encode(s)
	return bytes.TrimRight(buf.Bytes(), "\n")
}
