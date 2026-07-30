package drivers

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestSaveAndLoadDriversRoundTrip(t *testing.T) {
	root := t.TempDir()
	want := &DriversData{
		Version: "1.0",
		Types: map[string]DriverType{
			"gpu": {
				Title: "GPU",
				Brands: map[string]DriverBrand{
					"nvidia": {
						Title: "NVIDIA",
						Series: []DriverSeries{{
							Name:       "GeForce",
							FolderPath: `Drivers\NVIDIA`,
						}},
					},
				},
			},
		},
	}

	if err := SaveDrivers(root, want); err != nil {
		t.Fatalf("SaveDrivers() error = %v", err)
	}
	got, err := LoadDrivers(root)
	if err != nil {
		t.Fatalf("LoadDrivers() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch:\ngot  %#v\nwant %#v", got, want)
	}
}

func TestLoadDriversInitializesMissingTypes(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "data", "drivers.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"version":"1.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadDrivers(root)
	if err != nil {
		t.Fatalf("LoadDrivers() error = %v", err)
	}
	if got.Types == nil {
		t.Fatal("LoadDrivers() left Types nil")
	}
}

func TestLoadDriversFromEmbeddedFS(t *testing.T) {
	fsys := fstest.MapFS{
		"web/data/drivers.json": &fstest.MapFile{
			Data: []byte(`{"version":"1.0","types":{}}`),
		},
	}
	got, err := LoadFromFS(fsys)
	if err != nil {
		t.Fatalf("LoadFromFS() error = %v", err)
	}
	if got.Version != "1.0" || got.Types == nil {
		t.Fatalf("unexpected data: %#v", got)
	}
}

func TestSaveDriversRejectsNilData(t *testing.T) {
	if err := SaveDrivers(t.TempDir(), nil); err == nil {
		t.Fatal("SaveDrivers(nil) should return an error")
	}
}
