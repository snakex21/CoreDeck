package settings

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestSaveAndLoadSettingsRoundTrip(t *testing.T) {
	root := t.TempDir()
	want := &SettingsData{
		Version: "1.0",
		Categories: []SettingEntry{{
			Name:       "Fan Control",
			FolderPath: `Settings\FanControl`,
		}},
	}

	if err := SaveSettings(root, want); err != nil {
		t.Fatalf("SaveSettings() error = %v", err)
	}
	got, err := LoadSettings(root)
	if err != nil {
		t.Fatalf("LoadSettings() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch:\ngot  %#v\nwant %#v", got, want)
	}
}

func TestLoadSettingsRejectsMalformedJSON(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "data", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"version":`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSettings(root); err == nil {
		t.Fatal("LoadSettings() should reject malformed JSON")
	}
}

func TestLoadSettingsFromEmbeddedFS(t *testing.T) {
	fsys := fstest.MapFS{
		"web/data/settings.json": &fstest.MapFile{
			Data: []byte(`{"version":"1.0","categories":[]}`),
		},
	}
	got, err := LoadFromFS(fsys)
	if err != nil {
		t.Fatalf("LoadFromFS() error = %v", err)
	}
	if got.Version != "1.0" || got.Categories == nil {
		t.Fatalf("unexpected data: %#v", got)
	}
}
