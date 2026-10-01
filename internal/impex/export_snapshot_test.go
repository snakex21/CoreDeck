package impex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"launcher/internal/drivers"
	"launcher/internal/programs"
)

func catalogueSnapshot(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func readExportManifest(t *testing.T, path string, value any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, value); err != nil {
		t.Fatal(err)
	}
}

func TestExportDoesNotMutateLiveProgramVariants(t *testing.T) {
	for _, format := range []string{"folder", "zip"} {
		t.Run(format, func(t *testing.T) {
			root := t.TempDir()
			writeImportFixture(t, filepath.Join(root, "source", "app.exe"), []byte("fixture"))
			pd := &programs.ProgramsData{Programs: []programs.Program{{
				Name: "App", FolderPath: "source", Variants: []programs.Variant{{Name: "v1", FolderPath: "source"}},
			}}}
			before := catalogueSnapshot(t, pd)
			// A second export must still find the original source paths.
			for attempt := 0; attempt < 2; attempt++ {
				target := filepath.Join(root, fmt.Sprintf("export-%d", attempt))
				result := ExportPack(root, pd, nil, ExportPackRequest{
					ProgramNames: []string{"App"}, IncludeFiles: true, Format: format, TargetPath: target,
				})
				if !result.Success || len(result.Errors) != 0 || result.ProgramCount != 1 {
					t.Fatalf("export: %+v", result)
				}
				if after := catalogueSnapshot(t, pd); before != after {
					t.Fatalf("live catalogue changed from %s to %s", before, after)
				}
				if format == "zip" {
					target = t.TempDir()
					if err := unzip(result.ExportedPath, target); err != nil {
						t.Fatal(err)
					}
				}
				var exported programs.ProgramsData
				readExportManifest(t, filepath.Join(target, "programs.json"), &exported)
				if len(exported.Programs) != 1 || len(exported.Programs[0].Variants) != 1 ||
					exported.Programs[0].FolderPath != "programs/App/v1" || exported.Programs[0].Variants[0].FolderPath != "programs/App/v1" {
					t.Fatalf("portable export paths missing: %+v", exported)
				}
				assertImportFile(t, filepath.Join(target, "programs", "App", "v1", "app.exe"), "fixture")
			}
		})
	}
}

func TestExportDoesNotPruneLiveDriverSeries(t *testing.T) {
	for _, format := range []string{"folder", "zip"} {
		t.Run(format, func(t *testing.T) {
			root := t.TempDir()
			writeImportFixture(t, filepath.Join(root, "source", "driver.inf"), []byte("fixture"))
			dd := &drivers.DriversData{Types: map[string]drivers.DriverType{"gpu": {
				Brands: map[string]drivers.DriverBrand{"vendor": {
					Series: []drivers.DriverSeries{{Name: "one", FolderPath: "source"}, {Name: "two", FolderPath: "source"}},
				}},
			}}}
			before := catalogueSnapshot(t, dd)
			for attempt := 0; attempt < 2; attempt++ {
				target := filepath.Join(root, fmt.Sprintf("export-%d", attempt))
				result := ExportPack(root, nil, dd, ExportPackRequest{
					DriverTypes: []string{"gpu"}, IncludeFiles: true, Format: format, TargetPath: target,
					DriverSelections: map[string]DriverSelection{"gpu": {Brands: map[string]BrandSelection{"vendor": {Series: []string{"two"}}}}},
				})
				if !result.Success || len(result.Errors) != 0 || result.DriverCount != 1 {
					t.Fatalf("export: %+v", result)
				}
				if after := catalogueSnapshot(t, dd); before != after {
					t.Fatalf("live catalogue changed from %s to %s", before, after)
				}
				if format == "zip" {
					target = t.TempDir()
					if err := unzip(result.ExportedPath, target); err != nil {
						t.Fatal(err)
					}
				}
				var exported drivers.DriversData
				readExportManifest(t, filepath.Join(target, "drivers.json"), &exported)
				series := exported.Types["gpu"].Brands["vendor"].Series
				if len(series) != 1 || series[0].Name != "two" || series[0].FolderPath != "drivers/gpu/vendor/two" {
					t.Fatalf("export selection or portable path missing: %+v", series)
				}
				assertImportFile(t, filepath.Join(target, "drivers", "gpu", "vendor", "two", "driver.inf"), "fixture")
			}
		})
	}
}

func TestMetadataExportPreservesEmptyAndNilSeries(t *testing.T) {
	dd := &drivers.DriversData{Types: map[string]drivers.DriverType{"gpu": {
		Brands: map[string]drivers.DriverBrand{
			"empty": {Series: []drivers.DriverSeries{}},
			"nil":   {Series: nil},
		},
	}}}
	before := catalogueSnapshot(t, dd)
	root := t.TempDir()
	result := ExportPack(root, nil, dd, ExportPackRequest{DriverTypes: []string{"gpu"}, TargetPath: root})
	if !result.Success || len(result.Errors) != 0 {
		t.Fatalf("export: %+v", result)
	}
	var exported drivers.DriversData
	readExportManifest(t, filepath.Join(root, "drivers.json"), &exported)
	if got := catalogueSnapshot(t, &exported); got != before {
		t.Fatalf("metadata export changed empty values: got %s, want %s", got, before)
	}
	if catalogueSnapshot(t, dd) != before {
		t.Fatal("metadata export changed the live catalogue")
	}
}
