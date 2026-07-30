package impex

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"launcher/internal/drivers"
	"launcher/internal/programs"
)

func TestHumanSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{bytes: 0, want: "0 B"},
		{bytes: 999, want: "999 B"},
		{bytes: 1024, want: "1 KB"},
		{bytes: 1536, want: "2 KB"},
		{bytes: 3 * 1024 * 1024, want: "3.0 MB"},
		{bytes: 2 * 1024 * 1024 * 1024, want: "2.0 GB"},
	}
	for _, test := range tests {
		if got := humanSize(test.bytes); got != test.want {
			t.Errorf("humanSize(%d) = %q, want %q", test.bytes, got, test.want)
		}
	}
}

func TestDirSizeCountsNestedFiles(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "one.bin"), make([]byte, 17), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "two.bin"), make([]byte, 25), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := dirSize(root); got != 42 {
		t.Fatalf("dirSize() = %d, want 42", got)
	}
}

func TestValidateExportCountsItemsSizesAndMissingPaths(t *testing.T) {
	root := t.TempDir()
	programDir := filepath.Join(root, "program")
	driverDir := filepath.Join(root, "driver")
	for path, size := range map[string]int{
		filepath.Join(programDir, "app.exe"):       11,
		filepath.Join(driverDir, "driver-package"): 13,
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	programsData := &programs.ProgramsData{Programs: []programs.Program{
		{Name: "Present", FolderPath: programDir},
		{Name: "Missing", FolderPath: filepath.Join(root, "missing")},
	}}
	driversData := &drivers.DriversData{Types: map[string]drivers.DriverType{
		"gpu": {
			Title: "GPU",
			Brands: map[string]drivers.DriverBrand{
				"vendor": {
					Title:  "Vendor",
					Series: []drivers.DriverSeries{{Name: "Series", FolderPath: driverDir}},
				},
			},
		},
	}}

	got := ValidateExport(root, programsData, driversData, ValidateExportRequest{
		ProgramNames: []string{"Present", "Missing"},
		DriverTypes:  []string{"gpu"},
		IncludeFiles: true,
	})

	if !got.Success || got.ProgramCount != 2 || got.DriverCount != 1 {
		t.Fatalf("unexpected validation result: %#v", got)
	}
	if got.EstimatedSize != 24 {
		t.Fatalf("EstimatedSize = %d, want 24", got.EstimatedSize)
	}
	if len(got.Warnings) != 1 || got.Warnings[0].Item != "Missing" ||
		got.Warnings[0].Query != "missing_folder" {
		t.Fatalf("unexpected warnings: %#v", got.Warnings)
	}
}

func TestExportPreviewAndImportRoundTrip(t *testing.T) {
	root := t.TempDir()
	packDir := filepath.Join(root, "pack")
	sourcePrograms := &programs.ProgramsData{
		Version: "1.0",
		Programs: []programs.Program{{
			Name:        "Rufus",
			Category:    "Tools",
			InstallType: "portable",
		}},
	}
	sourceDrivers := &drivers.DriversData{
		Version: "1.0",
		Types: map[string]drivers.DriverType{
			"gpu": {
				Title:  "GPU",
				Brands: map[string]drivers.DriverBrand{},
			},
		},
	}

	exported := ExportPack(root, sourcePrograms, sourceDrivers, ExportPackRequest{
		ProgramNames: []string{"Rufus"},
		DriverTypes:  []string{"gpu"},
		Format:       "folder",
		TargetPath:   packDir,
	})
	if !exported.Success || exported.ProgramCount != 1 || exported.DriverCount != 1 {
		t.Fatalf("ExportPack() = %#v", exported)
	}

	preview := PreviewPack(packDir)
	if !preview.Success || preview.ProgramCount != 1 || preview.DriverCount != 1 {
		t.Fatalf("PreviewPack() = %#v", preview)
	}
	if len(preview.Programs) != 1 || preview.Programs[0].Name != "Rufus" {
		t.Fatalf("unexpected preview programs: %#v", preview.Programs)
	}

	targetPrograms := &programs.ProgramsData{}
	targetDrivers := &drivers.DriversData{}
	imported := ImportPack(root, targetPrograms, targetDrivers, ImportPackRequest{
		PackPath:       packDir,
		Strategy:       "overwrite",
		ImportPrograms: true,
		ImportDrivers:  true,
	})
	if !imported.Success || imported.ProgramCount != 1 || imported.DriverCount != 1 {
		t.Fatalf("ImportPack() = %#v", imported)
	}
	if len(targetPrograms.Programs) != 1 || targetPrograms.Programs[0].Name != "Rufus" {
		t.Fatalf("unexpected imported programs: %#v", targetPrograms.Programs)
	}
	if _, ok := targetDrivers.Types["gpu"]; !ok {
		t.Fatalf("GPU driver type was not imported: %#v", targetDrivers.Types)
	}
}

func TestUnzipRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	zipPath := filepath.Join(root, "malicious.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("../escape.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("escape")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	err = unzip(zipPath, filepath.Join(root, "output"))
	if err == nil || !strings.Contains(err.Error(), "ZIP") {
		t.Fatalf("unzip() error = %v, want traversal rejection", err)
	}
	if _, err := os.Stat(filepath.Join(root, "escape.txt")); !os.IsNotExist(err) {
		t.Fatalf("traversal file exists outside output: %v", err)
	}
}

func TestHistoryLifecycleDeduplicatesSortsAndLimits(t *testing.T) {
	root := t.TempDir()
	base := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	for index := 0; index < historyMaxEntries+5; index++ {
		AppendHistory(root, ExportRecord{
			ID:         "record-" + humanSize(int64(index)),
			CreatedAt:  base.Add(time.Duration(index) * time.Minute).Format(time.RFC3339),
			TargetPath: filepath.Join(root, humanSize(int64(index))),
		})
	}

	records := LoadHistory(root)
	if len(records) != historyMaxEntries {
		t.Fatalf("history length = %d, want %d", len(records), historyMaxEntries)
	}
	if records[0].CreatedAt < records[len(records)-1].CreatedAt {
		t.Fatalf("history is not sorted newest first: %#v", records)
	}

	duplicateID := records[0].ID
	records = AppendHistory(root, ExportRecord{
		ID:         duplicateID,
		CreatedAt:  base.Add(24 * time.Hour).Format(time.RFC3339),
		TargetPath: "updated",
	})
	matches := 0
	for _, record := range records {
		if record.ID == duplicateID {
			matches++
		}
	}
	if matches != 1 || records[0].TargetPath != "updated" {
		t.Fatalf("history deduplication failed: matches=%d first=%#v", matches, records[0])
	}

	if !RemoveHistoryEntry(root, duplicateID) {
		t.Fatal("RemoveHistoryEntry() did not remove an existing ID")
	}
	if RemoveHistoryEntry(root, "missing") {
		t.Fatal("RemoveHistoryEntry() removed a missing ID")
	}
	ClearHistory(root)
	if got := LoadHistory(root); len(got) != 0 {
		t.Fatalf("ClearHistory() left %d entries", len(got))
	}
}
