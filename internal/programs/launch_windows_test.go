//go:build windows

package programs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeExitScript(t *testing.T, code string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "Installer & co")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "setup.cmd"), []byte("@echo off\r\nexit /b "+code+"\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// A real installer process that the user cancels (exit code 1602) must be
// reported as a cancelled installation, not as an error.
func TestRunInstallerUserCancelIsNotAnError(t *testing.T) {
	dir := writeExitScript(t, "1602")
	res := runInstaller(dir, "ByClick Downloader", "setup.cmd", "", false)
	if !res.Cancelled || res.Outcome != OutcomeCancelled || res.Success {
		t.Fatalf("got %+v; want cancelled", res)
	}
	if res.Message != "Instalacja anulowana" || strings.Contains(res.Message, "exit status") {
		t.Fatalf("message %q", res.Message)
	}

	// Same through the staged (TEMP copy) path used by runFromTemp entries.
	res = runInstallerStaged(filepath.Join(dir, "setup.cmd"), t.TempDir(), "ByClick Downloader", "")
	if !res.Cancelled || res.Message != "Instalacja anulowana" {
		t.Fatalf("staged: got %+v", res)
	}
}

func TestRunInstallerRebootRequiredIsSuccess(t *testing.T) {
	dir := writeExitScript(t, "3010")
	res := runInstaller(dir, "Runtime", "setup.cmd", "", false)
	if !res.Success || res.Outcome != OutcomeOKNote || !strings.Contains(res.Message, "Uruchom ponownie") {
		t.Fatalf("got %+v", res)
	}
}

// unusedDriveLetter returns a drive letter with nothing mounted, standing in
// for the pendrive's old letter (J: before it became L:).
func unusedDriveLetter(t *testing.T) string {
	t.Helper()
	for c := 'Z'; c >= 'D'; c-- {
		root := string(c) + `:\`
		if _, err := os.Stat(root); os.IsNotExist(err) {
			return string(c) + ":"
		}
	}
	t.Skip("no free drive letter")
	return ""
}

// The catalogue still holds an absolute path on the old letter; the program
// must start from the same folder on the drive CoreDeck runs from now.
func TestLaunchPortableAfterDriveLetterChange(t *testing.T) {
	started := stubLaunch(t)
	root, exeDir := portableLayout(t, "OCCT", "OCCT.exe")
	current := filepath.VolumeName(root)
	oldLetter := unusedDriveLetter(t)
	oldPath := oldLetter + strings.TrimPrefix(filepath.Join(root, "Centrum", "OCCT"), current)

	p := Program{Name: "OCCT", InstallType: "portable", FolderPath: oldPath}
	res := LaunchProgram(exeDir, p, -1, "")
	want := filepath.Join(root, "Centrum", "OCCT", "OCCT.exe")
	if !res.Success || res.Launched != want {
		t.Fatalf("old path %q: got %+v, started %v; want %q", oldPath, res, *started, want)
	}
}

func TestIsConsoleExecutable(t *testing.T) {
	self, err := os.Executable() // go test binaries are console programs
	if err != nil {
		t.Fatal(err)
	}
	if !isConsoleExecutable(self) {
		t.Fatalf("%s should be a console executable", self)
	}
	explorer := filepath.Join(os.Getenv("WINDIR"), "explorer.exe")
	if _, err := os.Stat(explorer); err == nil && isConsoleExecutable(explorer) {
		t.Fatalf("explorer.exe is a GUI program")
	}
	if isConsoleExecutable(filepath.Join(t.TempDir(), "missing.exe")) {
		t.Fatal("missing file is not a console program")
	}
}
