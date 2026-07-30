package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"launcher/internal/windowstate"
)

func TestEnvEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "empty", value: "", want: false},
		{name: "one", value: "1", want: true},
		{name: "true", value: "true", want: true},
		{name: "yes", value: " YES ", want: true},
		{name: "on", value: "On", want: true},
		{name: "zero", value: "0", want: false},
		{name: "false", value: "false", want: false},
		{name: "unknown", value: "enabled", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("COREDECK_TEST_FLAG", test.value)
			if got := envEnabled("COREDECK_TEST_FLAG"); got != test.want {
				t.Fatalf("envEnabled() = %v, want %v for %q", got, test.want, test.value)
			}
		})
	}
}

func TestRotateLogKeepsTwoPreviousGenerations(t *testing.T) {
	root := t.TempDir()
	logPath := filepath.Join(root, "launcher.log")
	if err := os.WriteFile(logPath, []byte("current"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath+".1", []byte("previous"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath+".2", []byte("oldest"), 0o644); err != nil {
		t.Fatal(err)
	}

	rotateLog(logPath, 1)

	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Fatalf("current log still exists after rotation: %v", err)
	}
	if got, err := os.ReadFile(logPath + ".1"); err != nil || !bytes.Equal(got, []byte("current")) {
		t.Fatalf("launcher.log.1 = %q, %v; want current", got, err)
	}
	if got, err := os.ReadFile(logPath + ".2"); err != nil || !bytes.Equal(got, []byte("previous")) {
		t.Fatalf("launcher.log.2 = %q, %v; want previous", got, err)
	}
}

func TestRotateLogLeavesSmallFileUntouched(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "launcher.log")
	want := []byte("small")
	if err := os.WriteFile(logPath, want, 0o644); err != nil {
		t.Fatal(err)
	}

	rotateLog(logPath, 1024)

	got, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("small log changed: got %q, want %q", got, want)
	}
	if _, err := os.Stat(logPath + ".1"); !os.IsNotExist(err) {
		t.Fatalf("unexpected backup for small log: %v", err)
	}
}

func TestSeedDataFilesCreatesDefaultsWithoutOverwritingUserData(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	customPrograms := []byte(`{"version":"user","programs":[{"name":"keep-me"}]}`)
	if err := os.WriteFile(filepath.Join(dataDir, "programs.json"), customPrograms, 0o644); err != nil {
		t.Fatal(err)
	}

	seedDataFiles(root)

	gotPrograms, err := os.ReadFile(filepath.Join(dataDir, "programs.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotPrograms, customPrograms) {
		t.Fatalf("seedDataFiles() overwrote user programs.json: %q", gotPrograms)
	}
	for _, name := range []string{"tweaks.json", "drivers.json", "knowledge.json", "settings.json"} {
		data, err := os.ReadFile(filepath.Join(dataDir, name))
		if err != nil {
			t.Fatalf("%s was not seeded: %v", name, err)
		}
		var value any
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatalf("%s contains invalid JSON: %v", name, err)
		}
	}
}

func TestBuildWindowConfigUsesPersistedStateAndDebugFlag(t *testing.T) {
	t.Setenv("COREDECK_WEBVIEW_DEBUG", "1")
	t.Setenv("LAUNCHER_DEBUG", "")
	state := windowstate.State{Width: 1280, Height: 720, Maximized: false}

	got := buildWindowConfig(
		"http://127.0.0.1:8000/index.html",
		`C:\CoreDeck\.webview_data`,
		`C:\CoreDeck\data\state\favicon.ico`,
		`C:\CoreDeck\data\state\window.json`,
		state,
		nil,
	)

	if got.Title != "CoreDeck" || got.Width != 1280 || got.Height != 720 || got.Maximized {
		t.Fatalf("unexpected window config: %#v", got)
	}
	if got.MinWidth != 800 || got.MinHeight != 600 || !got.Debug {
		t.Fatalf("window bounds/debug not applied: %#v", got)
	}
	if got.URL != "http://127.0.0.1:8000/index.html" ||
		got.StatePath != `C:\CoreDeck\data\state\window.json` {
		t.Fatalf("window paths not applied: %#v", got)
	}
}
