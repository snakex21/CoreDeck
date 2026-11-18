package paths

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// unusedDrive returns a drive letter with nothing mounted (the pendrive's
// old letter in these tests).
func unusedDrive(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("drive letters are Windows-only")
	}
	for c := 'Z'; c >= 'D'; c-- {
		if _, err := os.Stat(string(c) + `:\`); os.IsNotExist(err) {
			return string(c) + ":"
		}
	}
	t.Skip("no free drive letter")
	return ""
}

// onOldDrive rewrites an existing absolute path onto the old letter.
func onOldDrive(t *testing.T, oldLetter, abs string) string {
	t.Helper()
	return oldLetter + strings.TrimPrefix(abs, filepath.VolumeName(abs))
}

func TestResolveAbsoluteOnOldDriveLetter(t *testing.T) {
	old := unusedDrive(t)
	root, exeDir := layout(t, "Centrum/App")
	want := filepath.Join(root, "Centrum", "App")
	// driveRoot = the test root, standing in for L:\ ; the catalogue says J:\Centrum\App.
	p, ok := resolve(exeDir, root, old+`\Centrum\App`)
	if p != want || !ok {
		t.Fatalf("got %q, %v; want %q", p, ok, want)
	}
	// Through the public resolver, with the real current drive.
	if got := ResolvePath(exeDir, onOldDrive(t, old, want)); got != want {
		t.Fatalf("ResolvePath: got %q; want %q", got, want)
	}
	// Missing everywhere: the original path is returned, not found.
	if p, ok := resolve(exeDir, root, old+`\Centrum\Nope`); ok || !strings.HasPrefix(p, old) {
		t.Fatalf("missing: got %q, %v", p, ok)
	}
}

func TestMigrateDriveLetters(t *testing.T) {
	old := unusedDrive(t)
	root, exeDir := layout(t, "Centrum/App", "Centrum/Drivers & Co")
	app := filepath.Join(root, "Centrum", "App")
	drivers := filepath.Join(root, "Centrum", "Drivers & Co")
	existing := os.Getenv("WINDIR") // exists: must stay as is
	nowhere := old + `\Centrum\Gone`

	catalogue := map[string]any{
		"programs": []any{
			map[string]any{"name": "App", "folderPath": onOldDrive(t, old, app), "details": "Kopiuj z " + onOldDrive(t, old, app) + " ręcznie"},
			map[string]any{"name": "Drivers", "folderPath": filepath.ToSlash(onOldDrive(t, old, drivers))},
			map[string]any{"name": "System", "folderPath": existing},
			map[string]any{"name": "Gone", "folderPath": nowhere},
			map[string]any{"name": "Relative", "folderPath": "../Centrum/App"},
		},
	}
	data, _ := json.MarshalIndent(catalogue, "", "  ")
	file := filepath.Join(exeDir, "data", "programs.json")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, data, 0o644); err != nil {
		t.Fatal(err)
	}

	rewrites, err := MigrateDriveLetters(exeDir, file)
	if err != nil {
		t.Fatal(err)
	}
	if len(rewrites) != 2 {
		t.Fatalf("want 2 rewrites, got %+v", rewrites)
	}

	var got struct {
		Programs []struct {
			Name       string `json:"name"`
			FolderPath string `json:"folderPath"`
			Details    string `json:"details"`
		} `json:"programs"`
	}
	raw, _ := os.ReadFile(file)
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"App":      "../Centrum/App",
		"Drivers":  "../Centrum/Drivers & Co",
		"System":   existing,
		"Gone":     nowhere,
		"Relative": "../Centrum/App",
	}
	for _, p := range got.Programs {
		if p.FolderPath != want[p.Name] {
			t.Errorf("%s: folderPath %q; want %q", p.Name, p.FolderPath, want[p.Name])
		}
		if p.Name == "App" && !strings.Contains(p.Details, onOldDrive(t, old, app)) {
			t.Errorf("prose mentioning a path must not change: %q", p.Details)
		}
	}
	// The migrated relative paths resolve on the current drive.
	if p := ResolvePath(exeDir, want["Drivers"]); p != drivers {
		t.Errorf("resolve after migration: %q; want %q", p, drivers)
	}

	backups, _ := filepath.Glob(file + ".drive-*.bak")
	if len(backups) != 1 {
		t.Fatalf("want one backup, got %v", backups)
	}
	if b, _ := os.ReadFile(backups[0]); string(b) != string(data) {
		t.Fatal("backup must hold the original file")
	}

	// Second run: nothing left to do, file untouched.
	again, err := MigrateDriveLetters(exeDir, file)
	if err != nil || len(again) != 0 {
		t.Fatalf("second run: %+v, %v", again, err)
	}
}

func TestMigrateDriveLettersMissingFile(t *testing.T) {
	if rw, err := MigrateDriveLetters(t.TempDir(), filepath.Join(t.TempDir(), "none.json")); err != nil || rw != nil {
		t.Fatalf("got %v, %v", rw, err)
	}
}
