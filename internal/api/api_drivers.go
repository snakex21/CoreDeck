package api

// Driver catalogue read/write.

import (
	"encoding/json"
	"log/slog"

	"launcher/internal/drivers"
)

// GetDriversData returns the driver catalogue JSON (the same nested
// structure as script.js used inline: types → brands → series).
func (a *API) GetDriversData() string {
	if a.DriversData == nil {
		if dd, err := loadCatalogue(a.WebDir, "drivers.json", drivers.LoadDrivers, drivers.LoadFromFS, a.WebFS); err == nil {
			a.DriversData = dd
		} else {
			slog.Error("load drivers", "error", err)
			return `{"error": "Cannot load drivers.json"}`
		}
	}
	if a.DriversData.Types == nil {
		a.DriversData.Types = map[string]drivers.DriverType{}
	}
	data, _ := json.Marshal(a.DriversData)
	return string(data)
}

// SaveDriversData replaces the entire driver catalogue with the
// supplied JSON. The frontend is expected to send the next state
// (a deep clone with the edit applied), which keeps the API simple
// for a three-level tree where partial updates would be fiddly.
func (a *API) SaveDriversData(driversJSON string) string {
	var dd drivers.DriversData
	if err := json.Unmarshal([]byte(driversJSON), &dd); err != nil {
		return `{"error": "Invalid drivers JSON"}`
	}

	if err := drivers.SaveDrivers(a.WebDir, &dd); err != nil {
		return jsonErr(err)
	}
	a.DriversData = &dd
	return `{"success": true}`
}
