package windowstate

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "window.json")
	want := State{Width: 1280, Height: 760, Maximized: false}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if got := Load(path, State{}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestLoadUsesFallbackForInvalidOrUnsafeDimensions(t *testing.T) {
	fallback := State{Width: 1400, Height: 900, Maximized: true}
	path := filepath.Join(t.TempDir(), "window.json")

	if got := Load(path, fallback); !reflect.DeepEqual(got, fallback) {
		t.Fatalf("missing file: got %#v, want %#v", got, fallback)
	}
	if err := os.WriteFile(path, []byte(`{"width":100,"height":99999,"maximized":false}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Load(path, fallback)
	if got.Width != fallback.Width || got.Height != fallback.Height || got.Maximized {
		t.Fatalf("invalid dimensions were not normalized: %#v", got)
	}
}
