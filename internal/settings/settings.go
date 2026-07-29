// Package settings manages the settings catalogue: a flat list of
// configuration folders (Logitech G-Hub, MSI Afterburner, OpenRGB…).
// Data is loaded from <exeDir>/data/settings.json.
package settings

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"launcher/internal/safeio"
)

// SettingEntry is a single settings folder — e.g. "MSI Afterburner profiles".
type SettingEntry struct {
	Name       string `json:"name"`
	FolderPath string `json:"folderPath"`
}

// SettingsData is the on-disk format.
type SettingsData struct {
	Version    string         `json:"version"`
	Categories []SettingEntry `json:"categories"`
}

// LoadSettings reads settings.json from <webDir>/data/settings.json.
func LoadSettings(webDir string) (*SettingsData, error) {
	jsonPath := filepath.Join(webDir, "data", "settings.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read settings.json: %w", err)
	}
	var sd SettingsData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, fmt.Errorf("cannot parse settings.json: %w", err)
	}
	return &sd, nil
}

func LoadFromFS(fsys fs.FS) (*SettingsData, error) {
	data, err := fs.ReadFile(fsys, "web/data/settings.json")
	if err != nil {
		return nil, fmt.Errorf("cannot read settings.json from FS: %w", err)
	}
	var sd SettingsData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, fmt.Errorf("cannot parse settings.json from FS: %w", err)
	}
	return &sd, nil
}

// SaveSettings writes settings.json to <webDir>/data/settings.json.
func SaveSettings(webDir string, sd *SettingsData) error {
	jsonPath := filepath.Join(webDir, "data", "settings.json")
	data, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal settings.json: %w", err)
	}
	return safeio.WriteFileAtomic(jsonPath, data, 0o644)
}
