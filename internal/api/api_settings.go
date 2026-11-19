package api

// Settings catalogue read/write.

import (
	"encoding/json"
	"log/slog"

	"launcher/internal/settings"
)

// GetSettingsData returns the settings catalogue JSON.
func (a *API) GetSettingsData() string {
	if a.SettingsData == nil {
		if sd, err := loadCatalogue(a.WebDir, "settings.json", settings.LoadSettings, settings.LoadFromFS, a.WebFS); err == nil {
			a.SettingsData = sd
		} else {
			slog.Error("load settings", "error", err)
			return `{"error": "Cannot load settings.json"}`
		}
	}
	data, _ := json.Marshal(a.SettingsData)
	return string(data)
}

// SaveSettingsData replaces the entire settings catalogue.
func (a *API) SaveSettingsData(settingsJSON string) string {
	var sd settings.SettingsData
	if err := json.Unmarshal([]byte(settingsJSON), &sd); err != nil {
		return `{"error": "Invalid settings JSON"}`
	}
	if err := settings.SaveSettings(a.WebDir, &sd); err != nil {
		return jsonErr(err)
	}
	a.SettingsData = &sd
	return `{"success": true}`
}
