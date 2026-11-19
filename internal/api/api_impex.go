package api

// Import/export of program+driver packs and the export history.

import (
	"encoding/json"

	"launcher/internal/drivers"
	"launcher/internal/impex"
	"launcher/internal/programs"
	"launcher/internal/uimsg"
)

// ValidateExport sprawdza czy foldery istnieją i szacuje rozmiar przed eksportem.
func (a *API) ValidateExport(requestJSON string) string {
	var req impex.ValidateExportRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return `{"error": "Invalid request JSON"}`
	}
	if a.ProgramsData == nil {
		a.GetProgramsData()
	}
	if a.DriversData == nil {
		a.GetDriversData()
	}
	result := impex.ValidateExport(a.ExeDir, a.ProgramsData, a.DriversData, req)
	data, _ := json.Marshal(result)
	return string(data)
}

// ExportPack exports selected programs and drivers to a folder or ZIP.
func (a *API) ExportPack(requestJSON string) string {
	var req impex.ExportPackRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return `{"error": "Invalid request JSON"}`
	}

	// Ensure data is loaded
	if a.ProgramsData == nil {
		a.GetProgramsData()
	}
	if a.DriversData == nil {
		a.GetDriversData()
	}

	result := impex.ExportPack(a.ExeDir, a.ProgramsData, a.DriversData, req)

	// Zapisz wpis w historii eksportów (tylko przy sukcesie).
	if result.Success {
		impex.AppendHistory(a.ExeDir, impex.ExportRecord{
			TargetPath:   result.ExportedPath,
			Format:       req.Format,
			ProgramCount: result.ProgramCount,
			DriverCount:  result.DriverCount,
			IncludeFiles: req.IncludeFiles,
		})
	}

	data, _ := json.Marshal(result)
	return string(data)
}

// ListExportHistory zwraca listę ostatnio wyeksportowanych paczek.
func (a *API) ListExportHistory() string {
	records := impex.LoadHistory(a.ExeDir)
	data, _ := json.Marshal(records)
	return string(data)
}

// RemoveExportHistory usuwa jeden wpis z historii eksportów (plik na dysku zostaje).
func (a *API) RemoveExportHistory(idJSON string) string {
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(idJSON), &payload); err != nil || payload.ID == "" {
		return jsonResult(false, uimsg.New("msg_history_no_id", "Brak ID wpisu"))
	}
	if impex.RemoveHistoryEntry(a.ExeDir, payload.ID) {
		return jsonResult(true, uimsg.New("msg_history_removed", "Wpis usunięty z historii"))
	}
	return jsonResult(false, uimsg.New("msg_history_not_found", "Nie znaleziono wpisu"))
}

// ClearExportHistory czyści całą historię eksportów.
func (a *API) ClearExportHistory() string {
	impex.ClearHistory(a.ExeDir)
	return jsonResult(true, uimsg.New("msg_history_cleared", "Historia wyczyszczona"))
}

// ImportPack imports a folder or ZIP package into programs.json/drivers.json.
func (a *API) ImportPack(requestJSON string) string {
	var req impex.ImportPackRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return `{"error": "Invalid request JSON"}`
	}

	if a.ProgramsData == nil {
		a.GetProgramsData()
	}
	if a.DriversData == nil {
		a.GetDriversData()
	}

	result := impex.ImportPack(a.ExeDir, a.ProgramsData, a.DriversData, req)
	if result.ProgramCount > 0 && a.ProgramsData != nil {
		if err := programs.SavePrograms(a.WebDir, a.ProgramsData); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, impex.LineOf(uimsg.New("msg_impex_save_programs_failed", "Zapis programs.json: "+err.Error(), "error", err.Error())))
		}
	}
	if result.DriverCount > 0 && a.DriversData != nil {
		if err := drivers.SaveDrivers(a.WebDir, a.DriversData); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, impex.LineOf(uimsg.New("msg_impex_save_drivers_failed", "Zapis drivers.json: "+err.Error(), "error", err.Error())))
		}
	}

	data, _ := json.Marshal(result)
	return string(data)
}

func (a *API) PreviewPack(packPath string) string {
	result := impex.PreviewPack(packPath)
	data, _ := json.Marshal(result)
	return string(data)
}
