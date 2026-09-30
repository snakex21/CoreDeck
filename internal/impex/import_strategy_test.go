package impex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"launcher/internal/drivers"
	"launcher/internal/programs"
)

func writeImportFixture(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func writeImportManifest(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeImportFixture(t, path, data)
}

func assertImportFile(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil || string(data) != want {
		t.Fatalf("file %q = %q, %v; want %q", path, data, err, want)
	}
}

func TestProgramImportStrategyProtectsExistingFiles(t *testing.T) {
	for _, strategy := range []string{"skip", "duplicate", "overwrite"} {
		t.Run(strategy, func(t *testing.T) {
			root := t.TempDir()
			pack, target := filepath.Join(root, "pack"), filepath.Join(root, "target")
			oldPath := filepath.Join(target, "App")
			writeImportFixture(t, filepath.Join(oldPath, "app.exe"), []byte("original"))
			writeImportFixture(t, filepath.Join(pack, "programs", "App", "app.exe"), []byte("incoming"))
			writeImportManifest(t, filepath.Join(pack, "programs.json"), programs.ProgramsData{Programs: []programs.Program{{Name: "App", FolderPath: "programs/App"}}})
			original := programs.Program{Name: "App", FolderPath: oldPath, Description: "existing"}
			catalogue := &programs.ProgramsData{Programs: []programs.Program{original}, TotalPrograms: 1}
			result := ImportPack(root, catalogue, nil, ImportPackRequest{PackPath: pack, Strategy: strategy, ImportPrograms: true, ProgramsTargetDir: target})
			if !result.Success {
				t.Fatalf("import: %+v", result)
			}
			if strategy == "overwrite" {
				assertImportFile(t, filepath.Join(oldPath, "app.exe"), "incoming")
				return
			}
			assertImportFile(t, filepath.Join(oldPath, "app.exe"), "original")
			if !reflect.DeepEqual(catalogue.Programs[0], original) {
				t.Fatal("existing catalogue entry changed")
			}
			if strategy == "skip" {
				if result.ProgramCount != 0 || len(catalogue.Programs) != 1 {
					t.Fatalf("skipped item reported imported: %+v", result)
				}
			} else {
				if result.ProgramCount != 1 || len(catalogue.Programs) != 2 {
					t.Fatalf("duplicate not imported: %+v", result)
				}
				added := catalogue.Programs[1]
				if added.Name != "App_imported" || added.FolderPath == oldPath {
					t.Fatalf("duplicate not isolated: %+v", added)
				}
				assertImportFile(t, filepath.Join(added.FolderPath, "app.exe"), "incoming")
			}
		})
	}
}

func TestDriverImportStrategyProtectsExistingFiles(t *testing.T) {
	for _, strategy := range []string{"skip", "duplicate", "overwrite"} {
		t.Run(strategy, func(t *testing.T) {
			root := t.TempDir()
			pack, target := filepath.Join(root, "pack"), filepath.Join(root, "target")
			oldPath := filepath.Join(target, "gpu", "vendor", "Series")
			writeImportFixture(t, filepath.Join(oldPath, "driver.inf"), []byte("original"))
			writeImportFixture(t, filepath.Join(pack, "drivers", "gpu", "driver.inf"), []byte("incoming"))
			incoming := drivers.DriverType{Title: "New", Brands: map[string]drivers.DriverBrand{"vendor": {Series: []drivers.DriverSeries{{Name: "Series", FolderPath: "drivers/gpu"}}}}}
			writeImportManifest(t, filepath.Join(pack, "drivers.json"), drivers.DriversData{Types: map[string]drivers.DriverType{"gpu": incoming}})
			original := drivers.DriverType{Title: "Existing", Brands: map[string]drivers.DriverBrand{"vendor": {Series: []drivers.DriverSeries{{Name: "Series", FolderPath: oldPath}}}}}
			catalogue := &drivers.DriversData{Types: map[string]drivers.DriverType{"gpu": original}}
			result := ImportPack(root, nil, catalogue, ImportPackRequest{PackPath: pack, Strategy: strategy, ImportDrivers: true, DriversTargetDir: target})
			if !result.Success {
				t.Fatalf("import: %+v", result)
			}
			if strategy == "overwrite" {
				assertImportFile(t, filepath.Join(oldPath, "driver.inf"), "incoming")
				return
			}
			assertImportFile(t, filepath.Join(oldPath, "driver.inf"), "original")
			if !reflect.DeepEqual(catalogue.Types["gpu"], original) {
				t.Fatal("existing catalogue entry changed")
			}
			if strategy == "skip" {
				if result.DriverCount != 0 || len(catalogue.Types) != 1 {
					t.Fatalf("skipped item reported imported: %+v", result)
				}
			} else {
				added, ok := catalogue.Types["gpu_imported"]
				if !ok || result.DriverCount != 1 {
					t.Fatalf("duplicate not imported: %+v", result)
				}
				newPath := added.Brands["vendor"].Series[0].FolderPath
				if newPath == oldPath {
					t.Fatal("duplicate reuses existing directory")
				}
				assertImportFile(t, filepath.Join(newPath, "driver.inf"), "incoming")
			}
		})
	}
}

func TestZipOperationsPreserveOtherScratchDirectories(t *testing.T) {
	for _, operation := range []string{"preview", "import"} {
		t.Run(operation, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("TMPDIR", root)
			t.Setenv("TMP", root)
			t.Setenv("TEMP", root)
			shared := filepath.Join(root, "launcher_"+operation+"_tmp", "keep.txt")
			writeImportFixture(t, shared, []byte("other operation"))
			pack := filepath.Join(root, "pack")
			writeImportManifest(t, filepath.Join(pack, "programs.json"), programs.ProgramsData{Programs: []programs.Program{{Name: "App"}}})
			archive := filepath.Join(root, "pack.zip")
			if err := zipDirectory(pack, archive); err != nil {
				t.Fatal(err)
			}
			if operation == "preview" {
				if result := PreviewPack(archive); !result.Success {
					t.Fatalf("preview: %+v", result)
				}
			} else {
				if result := ImportPack(root, &programs.ProgramsData{}, nil, ImportPackRequest{PackPath: archive, ImportPrograms: true}); !result.Success {
					t.Fatalf("import: %+v", result)
				}
			}
			assertImportFile(t, shared, "other operation")
		})
	}
}

func TestDuplicateImportPreservesUncataloguedDestination(t *testing.T) {
	root := t.TempDir()
	pack, target := filepath.Join(root, "pack"), filepath.Join(root, "target")
	writeImportFixture(t, filepath.Join(pack, "payload", "data.bin"), []byte("incoming"))
	writeImportFixture(t, filepath.Join(target, "App_imported", "keep.txt"), []byte("unrelated program"))
	writeImportFixture(t, filepath.Join(target, "gpu_imported", "keep.txt"), []byte("unrelated driver"))
	writeImportManifest(t, filepath.Join(pack, "programs.json"), programs.ProgramsData{Programs: []programs.Program{{Name: "App", FolderPath: "payload"}}})
	writeImportManifest(t, filepath.Join(pack, "drivers.json"), drivers.DriversData{Types: map[string]drivers.DriverType{"gpu": {Brands: map[string]drivers.DriverBrand{"vendor": {Series: []drivers.DriverSeries{{Name: "Series", FolderPath: "payload"}}}}}}})
	pd := &programs.ProgramsData{Programs: []programs.Program{{Name: "App"}}}
	dd := &drivers.DriversData{Types: map[string]drivers.DriverType{"gpu": {Title: "existing"}}}
	result := ImportPack(root, pd, dd, ImportPackRequest{PackPath: pack, Strategy: "duplicate", ImportPrograms: true, ImportDrivers: true, ProgramsTargetDir: target, DriversTargetDir: target})
	if !result.Success || result.ProgramCount != 1 || result.DriverCount != 1 {
		t.Fatalf("import: %+v", result)
	}
	assertImportFile(t, filepath.Join(target, "App_imported", "keep.txt"), "unrelated program")
	assertImportFile(t, filepath.Join(target, "gpu_imported", "keep.txt"), "unrelated driver")
	assertImportFile(t, filepath.Join(pd.Programs[1].FolderPath, "data.bin"), "incoming")
	series := dd.Types["gpu_imported_2"].Brands["vendor"].Series
	if len(series) != 1 {
		t.Fatalf("duplicate did not choose a free driver destination: %+v", dd.Types)
	}
	assertImportFile(t, filepath.Join(series[0].FolderPath, "data.bin"), "incoming")
}

func TestZipOperationsCleanScratchAfterFailure(t *testing.T) {
	for _, operation := range []string{"preview", "import"} {
		t.Run(operation, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("TMPDIR", root)
			t.Setenv("TMP", root)
			t.Setenv("TEMP", root)
			archive := filepath.Join(root, "broken.zip")
			writeImportFixture(t, archive, []byte("not a ZIP"))
			if operation == "preview" {
				if result := PreviewPack(archive); result.Success {
					t.Fatal("invalid ZIP accepted")
				}
			} else {
				if result := ImportPack(root, &programs.ProgramsData{}, nil, ImportPackRequest{PackPath: archive, ImportPrograms: true}); result.Success {
					t.Fatal("invalid ZIP accepted")
				}
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 || entries[0].Name() != "broken.zip" {
				t.Fatalf("scratch files leaked after failure: %v", entries)
			}
		})
	}
}

func TestImportProtectsOrphanDestinations(t *testing.T) {
	for _, strategy := range []string{"skip", "duplicate"} {
		t.Run(strategy, func(t *testing.T) {
			root := t.TempDir()
			pack, target := filepath.Join(root, "pack"), filepath.Join(root, "target")
			writeImportFixture(t, filepath.Join(pack, "payload", "data.bin"), []byte("incoming"))
			writeImportFixture(t, filepath.Join(target, "App", "keep.txt"), []byte("uncatalogued program"))
			writeImportFixture(t, filepath.Join(target, "gpu", "vendor", "Series", "keep.txt"), []byte("uncatalogued driver"))
			writeImportManifest(t, filepath.Join(pack, "programs.json"), programs.ProgramsData{Programs: []programs.Program{{Name: "App", FolderPath: "payload"}}})
			writeImportManifest(t, filepath.Join(pack, "drivers.json"), drivers.DriversData{Types: map[string]drivers.DriverType{"gpu": {Brands: map[string]drivers.DriverBrand{"vendor": {Series: []drivers.DriverSeries{{Name: "Series", FolderPath: "payload"}}}}}}})
			pd, dd := &programs.ProgramsData{}, &drivers.DriversData{}
			result := ImportPack(root, pd, dd, ImportPackRequest{PackPath: pack, Strategy: strategy, ImportPrograms: true, ImportDrivers: true, ProgramsTargetDir: target, DriversTargetDir: target})
			if !result.Success {
				t.Fatalf("import: %+v", result)
			}
			assertImportFile(t, filepath.Join(target, "App", "keep.txt"), "uncatalogued program")
			assertImportFile(t, filepath.Join(target, "gpu", "vendor", "Series", "keep.txt"), "uncatalogued driver")
			if strategy == "skip" {
				if result.ProgramCount != 0 || result.DriverCount != 0 {
					t.Fatalf("skip added conflicting entries: %+v", result)
				}
			} else {
				if len(pd.Programs) != 1 || pd.Programs[0].Name != "App_imported" {
					t.Fatalf("duplicate name: %+v", pd.Programs)
				}
				assertImportFile(t, filepath.Join(pd.Programs[0].FolderPath, "data.bin"), "incoming")
				series := dd.Types["gpu_imported"].Brands["vendor"].Series
				if len(series) != 1 {
					t.Fatalf("duplicate did not choose a free driver destination: %+v", dd.Types)
				}
				assertImportFile(t, filepath.Join(series[0].FolderPath, "data.bin"), "incoming")
			}
		})
	}
}
