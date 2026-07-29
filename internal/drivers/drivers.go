// Package drivers manages the driver catalogue: a three-level tree of
// hardware type → brand → series. Data is loaded from an external
// JSON file next to the launcher exe (<exeDir>/data/drivers.json) and
// persisted back to the same location.
package drivers

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"launcher/internal/jsonutil"
	"launcher/internal/safeio"
)

// DriverSeries is a single driver entry — e.g. "GeForce 4000 & 5000"
// pointing at a folder of installers.
type DriverSeries struct {
	Name       string `json:"name"`
	FolderPath string `json:"folderPath"`
}

// DriverBrand groups a set of series under one vendor (NVIDIA, AMD…).
type DriverBrand struct {
	Title  string         `json:"title"`
	Series []DriverSeries `json:"series"`
}

// DriverType groups brands under one hardware category
// (GPU, motherboard, headphones…).
type DriverType struct {
	Title  string                 `json:"title"`
	Brands map[string]DriverBrand `json:"brands"`
}

// DriversData is the on-disk format. The JSON keys under "types" and
// "brands" act as stable identifiers — keep them short, lower-case,
// without spaces.
type DriversData struct {
	Version string                `json:"version"`
	Types   map[string]DriverType `json:"types"`
}

// LoadDrivers reads drivers.json from <webDir>/data/drivers.json
// (the external, user-editable copy next to the launcher exe).
func LoadDrivers(webDir string) (*DriversData, error) {
	jsonPath := filepath.Join(webDir, "data", "drivers.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read drivers.json: %w", err)
	}
	var dd DriversData
	if err := jsonutil.Unmarshal(data, &dd); err != nil {
		return nil, fmt.Errorf("cannot parse drivers.json: %w", err)
	}
	if dd.Types == nil {
		dd.Types = map[string]DriverType{}
	}
	return &dd, nil
}

// LoadFromFS reads the embedded web/data/drivers.json (used as a
// fallback when the external file is missing or corrupt).
func LoadFromFS(fsys fs.FS) (*DriversData, error) {
	data, err := fs.ReadFile(fsys, "web/data/drivers.json")
	if err != nil {
		return nil, fmt.Errorf("cannot read drivers.json from FS: %w", err)
	}
	var dd DriversData
	if err := jsonutil.Unmarshal(data, &dd); err != nil {
		return nil, fmt.Errorf("cannot parse drivers.json: %w", err)
	}
	if dd.Types == nil {
		dd.Types = map[string]DriverType{}
	}
	return &dd, nil
}

// SaveDrivers writes the drivers catalogue back to the external file
// at <webDir>/data/drivers.json. The data directory is created if it
// does not exist.
func SaveDrivers(webDir string, data *DriversData) error {
	if data == nil {
		return fmt.Errorf("drivers data is nil")
	}
	if data.Types == nil {
		data.Types = map[string]DriverType{}
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Join(webDir, "data")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return safeio.WriteFileAtomic(filepath.Join(dir, "drivers.json"), out, 0o644)
}
