package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestMigrateWebviewProfileCopiesWithoutCaches(t *testing.T) {
	legacy := t.TempDir()
	profile := filepath.Join(t.TempDir(), "data", "webview")
	src := filepath.Join(legacy, "EBWebView")
	writeFile(t, filepath.Join(src, "Local State"), "state")
	writeFile(t, filepath.Join(src, "Default", "Local Storage", "leveldb", "000003.log"), "theme=dark")
	writeFile(t, filepath.Join(src, "Default", "Preferences"), "{}")
	writeFile(t, filepath.Join(src, "Default", "Cache", "Cache_Data", "f_000001"), "cached")
	writeFile(t, filepath.Join(src, "GrShaderCache", "data_0"), "shader")
	writeFile(t, filepath.Join(src, "Crashpad", "settings.dat"), "crash")

	n, err := migrateWebviewProfile(legacy, profile)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("copied %d files, want 3", n)
	}
	dst := filepath.Join(profile, "EBWebView")
	for _, rel := range []string{"Local State", "Default/Preferences", "Default/Local Storage/leveldb/000003.log"} {
		if b, err := os.ReadFile(filepath.Join(dst, filepath.FromSlash(rel))); err != nil || len(b) == 0 {
			t.Errorf("%s not copied: %v", rel, err)
		}
	}
	for _, rel := range []string{"Default/Cache", "GrShaderCache", "Crashpad"} {
		if _, err := os.Stat(filepath.Join(dst, filepath.FromSlash(rel))); err == nil {
			t.Errorf("cache %s should not be copied", rel)
		}
	}
	// The original is left untouched (copy, not move).
	if _, err := os.Stat(filepath.Join(src, "Default", "Cache", "Cache_Data", "f_000001")); err != nil {
		t.Fatal("legacy profile was modified")
	}
	if _, err := os.Stat(filepath.Join(profile, "EBWebView.migrating")); err == nil {
		t.Fatal("temporary folder left behind")
	}
	if _, err := os.Stat(filepath.Join(profile, "MIGRATED_FROM_APPDATA.txt")); err != nil {
		t.Fatal("migration note missing")
	}

	// Second start: portable profile exists, nothing is copied again.
	writeFile(t, filepath.Join(src, "Default", "Preferences"), `{"changed":true}`)
	if n, err := migrateWebviewProfile(legacy, profile); err != nil || n != 0 {
		t.Fatalf("second migration copied %d files (err %v)", n, err)
	}
	if b, _ := os.ReadFile(filepath.Join(dst, "Default", "Preferences")); string(b) != "{}" {
		t.Fatal("existing portable profile was overwritten")
	}
}

func TestMigrateWebviewProfileNothingToDo(t *testing.T) {
	profile := t.TempDir()
	if n, err := migrateWebviewProfile("", profile); n != 0 || err != nil {
		t.Fatalf("empty legacy root: %d %v", n, err)
	}
	if n, err := migrateWebviewProfile(filepath.Join(t.TempDir(), "missing"), profile); n != 0 || err != nil {
		t.Fatalf("missing legacy profile: %d %v", n, err)
	}
	if _, err := os.Stat(filepath.Join(profile, "EBWebView")); err == nil {
		t.Fatal("EBWebView must not be created without a source")
	}
}

func TestSetupWebviewProfileSetsEnvironment(t *testing.T) {
	t.Setenv(webviewProfileEnv, "")
	profile := filepath.Join(t.TempDir(), "data", "webview")
	if got := setupWebviewProfile(profile, ""); got != profile {
		t.Fatalf("got %q, want %q", got, profile)
	}
	if os.Getenv(webviewProfileEnv) != profile {
		t.Fatal("environment variable not set for the patched webview_go")
	}
	// An explicit override wins.
	t.Setenv(webviewProfileEnv, `X:\custom`)
	if got := setupWebviewProfile(profile, ""); got != `X:\custom` {
		t.Fatalf("override ignored: %q", got)
	}
}

func TestPatchedWebviewReadsProfileVariable(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "third_party", "webview_go", "libs", "webview", "include", "webview.h"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `L"WEBVIEW_GO_USER_DATA_FOLDER"`) {
		t.Fatal("third_party/webview_go is not the patched copy")
	}
	mod, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mod), "replace github.com/webview/webview_go => ./third_party/webview_go") {
		t.Fatal("go.mod does not use the patched webview_go")
	}
}
