package impex

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"launcher/internal/drivers"
	"launcher/internal/paths"
	"launcher/internal/programs"
	"launcher/internal/uimsg"
)

// Line is one log or error line shown in the UI. Code is its UI message key
// (translations.js), Params its placeholders; Message is the Polish fallback
// and log text (see internal/uimsg).
type Line struct {
	Message string            `json:"message"`
	Code    string            `json:"code,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
}

func LineOf(m uimsg.Msg) Line {
	return Line{Message: m.Text, Code: m.Code, Params: m.Params}
}

// errorLine reports a failed part of an export/import: the error's own
// message when it has a code, otherwise "<part>: <error>".
func errorLine(code, prefix string, err error) Line {
	if m := uimsg.From(err); m.Code != "" {
		return LineOf(m)
	}
	return LineOf(uimsg.New(code, prefix+err.Error(), "error", err.Error()))
}

func itoa(n int) string { return fmt.Sprintf("%d", n) }

type ExportPackRequest struct {
	ProgramNames     []string                   `json:"programNames"`
	DriverTypes      []string                   `json:"driverTypes"`
	DriverSelections map[string]DriverSelection `json:"driverSelections,omitempty"` // type → brands/series
	IncludeFiles     bool                       `json:"includeFiles"`
	Format           string                     `json:"format"` // "folder" albo "zip"
	TargetPath       string                     `json:"targetPath"`
}

// DriverSelection opisuje które brand-y i serie eksportować w danym typie.
type DriverSelection struct {
	Brands map[string]BrandSelection `json:"brands,omitempty"` // brand → series names; puste = wszystkie
}

type BrandSelection struct {
	Series []string `json:"series,omitempty"` // puste = wszystkie
}

type ExportPackResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	// Code / Params: UI message key and placeholders (see Line).
	Code          string            `json:"code,omitempty"`
	Params        map[string]string `json:"params,omitempty"`
	ExportedPath  string            `json:"exportedPath"`
	ProgramCount  int               `json:"programCount"`
	DriverCount   int               `json:"driverCount"`
	EstimatedSize int64             `json:"estimatedSize"` // bajty
	Log           []Line            `json:"log,omitempty"`
	Errors        []Line            `json:"errors,omitempty"`
}

func (r *ExportPackResult) setMsg(m uimsg.Msg) {
	r.Message, r.Code, r.Params = m.Text, m.Code, m.Params
}

// ValidateExportRequest to samo co ExportPackRequest — sprawdza czy foldery istnieją.
type ValidateExportRequest struct {
	ProgramNames     []string                   `json:"programNames"`
	DriverTypes      []string                   `json:"driverTypes"`
	DriverSelections map[string]DriverSelection `json:"driverSelections,omitempty"`
	IncludeFiles     bool                       `json:"includeFiles"`
}

type ValidateExportResult struct {
	Success       bool              `json:"success"`
	Message       string            `json:"message"`
	Code          string            `json:"code,omitempty"`
	Params        map[string]string `json:"params,omitempty"`
	EstimatedSize int64             `json:"estimatedSize"`
	Warnings      []ValidateWarning `json:"warnings,omitempty"`
	ProgramCount  int               `json:"programCount"`
	DriverCount   int               `json:"driverCount"`
}

func (r *ValidateExportResult) setMsg(m uimsg.Msg) {
	r.Message, r.Code, r.Params = m.Text, m.Code, m.Params
}

type ValidateWarning struct {
	Item  string `json:"item"`
	Type  string `json:"type"` // "program" lub "driver"
	Path  string `json:"path"`
	Query string `json:"query"` // "missing_folder" itp.
}

type ImportPackRequest struct {
	PackPath          string   `json:"packPath"`
	ProgramsTargetDir string   `json:"programsTargetDir"`
	DriversTargetDir  string   `json:"driversTargetDir"`
	Strategy          string   `json:"strategy"` // skip, overwrite, duplicate
	ImportPrograms    bool     `json:"importPrograms"`
	ImportDrivers     bool     `json:"importDrivers"`
	SelectedPrograms  []string `json:"selectedPrograms"`
	SelectedDrivers   []string `json:"selectedDrivers"`
}

type ImportPackResult struct {
	Success      bool              `json:"success"`
	Message      string            `json:"message"`
	Code         string            `json:"code,omitempty"`
	Params       map[string]string `json:"params,omitempty"`
	ProgramCount int               `json:"programCount"`
	DriverCount  int               `json:"driverCount"`
	Skipped      []string          `json:"skipped,omitempty"`
	Overwritten  []string          `json:"overwritten,omitempty"`
	Errors       []Line            `json:"errors,omitempty"`
}

func (r *ImportPackResult) setMsg(m uimsg.Msg) {
	r.Message, r.Code, r.Params = m.Text, m.Code, m.Params
}

func (p *PackPreview) setMsg(m uimsg.Msg) { p.Message, p.Code, p.Params = m.Text, m.Code, m.Params }

type PackPreview struct {
	Success      bool              `json:"success"`
	Message      string            `json:"message"`
	Code         string            `json:"code,omitempty"`
	Params       map[string]string `json:"params,omitempty"`
	ProgramCount int               `json:"programCount"`
	DriverCount  int               `json:"driverCount"`
	Programs     []PackProgramInfo `json:"programs,omitempty"`
	Drivers      []PackDriverInfo  `json:"drivers,omitempty"`
}

type PackProgramInfo struct {
	Name        string `json:"name"`
	InstallType string `json:"installType"`
	Variants    int    `json:"variants"`
	Category    string `json:"category"`
}

type PackDriverInfo struct {
	Key    string `json:"key"`
	Title  string `json:"title"`
	Brands int    `json:"brands"`
	Series int    `json:"series"`
}

// ValidateExport sprawdza czy foldery programów i sterowników istnieją
// i szacuje rozmiar paczki przed właściwym eksportem.
func ValidateExport(exeDir string, programsData *programs.ProgramsData, driversData *drivers.DriversData, req ValidateExportRequest) ValidateExportResult {
	result := ValidateExportResult{Success: true}
	if programsData != nil && len(req.ProgramNames) > 0 {
		nameSet := make(map[string]bool)
		for _, n := range req.ProgramNames {
			nameSet[n] = true
		}
		for _, p := range programsData.Programs {
			if !nameSet[p.Name] {
				continue
			}
			result.ProgramCount++
			if !req.IncludeFiles {
				continue
			}
			if len(p.Variants) > 0 {
				for _, v := range p.Variants {
					raw := firstNonEmpty(v.FolderPath, p.FolderPath)
					resolved := paths.ResolvePath(exeDir, raw)
					if resolved == "" || !exists(resolved) {
						result.Warnings = append(result.Warnings, ValidateWarning{
							Item:  p.Name + " → " + firstNonEmpty(v.Name, "?"),
							Type:  "program",
							Path:  raw,
							Query: "missing_folder",
						})
					} else {
						result.EstimatedSize += dirSize(resolved)
					}
				}
			} else {
				resolved := paths.ResolvePath(exeDir, p.FolderPath)
				if resolved == "" || !exists(resolved) {
					result.Warnings = append(result.Warnings, ValidateWarning{
						Item:  p.Name,
						Type:  "program",
						Path:  p.FolderPath,
						Query: "missing_folder",
					})
				} else {
					result.EstimatedSize += dirSize(resolved)
				}
			}
		}
	}

	if driversData != nil && len(req.DriverTypes) > 0 {
		typeSet := make(map[string]bool)
		for _, tk := range req.DriverTypes {
			typeSet[tk] = true
		}
		for typeKey, dt := range driversData.Types {
			if !typeSet[typeKey] {
				continue
			}
			result.DriverCount++
			if !req.IncludeFiles {
				continue
			}
			sel := req.DriverSelections[typeKey]
			for brandKey, brand := range dt.Brands {
				brandSel := sel.Brands[brandKey]
				brandSeriesSet := stringSet(brandSel.Series)
				allSeries := len(brandSel.Series) == 0
				for _, s := range brand.Series {
					if !allSeries && !brandSeriesSet[s.Name] {
						continue
					}
					resolved := paths.ResolvePath(exeDir, s.FolderPath)
					if resolved == "" || !exists(resolved) {
						result.Warnings = append(result.Warnings, ValidateWarning{
							Item:  (dt.Title) + " → " + firstNonEmpty(brand.Title, brandKey) + " → " + s.Name,
							Type:  "driver",
							Path:  s.FolderPath,
							Query: "missing_folder",
						})
					} else {
						result.EstimatedSize += dirSize(resolved)
					}
				}
			}
		}
	}

	if result.ProgramCount == 0 && result.DriverCount == 0 {
		result.Success = false
		result.setMsg(uimsg.New("msg_impex_nothing_selected", "Nie wybrano żadnych programów ani sterowników"))
		return result
	}

	if len(result.Warnings) > 0 {
		size := humanSize(result.EstimatedSize)
		result.setMsg(uimsg.New("count_msg_impex_validate_warnings", fmt.Sprintf("OK — %d zagrożeń (brakujące foldery). Rozmiar: %s", len(result.Warnings), size),
			"n", itoa(len(result.Warnings)), "size", size))
	} else {
		size := humanSize(result.EstimatedSize)
		result.setMsg(uimsg.New("msg_impex_validate_ok", "OK — rozmiar szacowany: "+size, "size", size))
	}
	return result
}

func dirSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

func humanSize(b int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.0f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func generateReadme(targetDir string, programCount, driverCount int, format string) {
	lines := []string{
		"Paczka Launchera",
		"================",
		"",
		fmt.Sprintf("Utworzono:    %s", time.Now().Format("2006-01-02 15:04")),
		fmt.Sprintf("Format:       %s", strings.ToUpper(format)),
		fmt.Sprintf("Programy:     %d", programCount),
		fmt.Sprintf("Sterowniki:   %d", driverCount),
		"",
		"Import",
		"------",
		"Aby zaimportować tę paczkę:",
		"1. Otwórz Launcher -> Zarządzanie -> Import/Export -> Import.",
		"2. Wybierz folder lub ZIP tej paczki.",
		"3. Kliknij 'Sprawdź' - zobaczysz podgląd zawartości.",
		"4. Wybierz strategię kolizji (nadpisz / pomiń / duplikat).",
		"5. Kliknij 'Importuj'.",
		"",
		"Pliki w paczce",
		"--------------",
		"  programs.json   - katalog programów",
		"  drivers.json    - katalog sterowników",
		"  README.txt      - ten plik",
		"",
	}
	_ = os.WriteFile(filepath.Join(targetDir, "README.txt"), []byte(strings.Join(lines, "\n")), 0644)
}

// ExportPack eksportuje wybrane programy i sterowniki do folderu lub ZIP.
func ExportPack(exeDir string, programsData *programs.ProgramsData, driversData *drivers.DriversData, req ExportPackRequest) ExportPackResult {
	result := ExportPackResult{Success: false}

	if req.TargetPath == "" {
		result.setMsg(uimsg.New("msg_impex_no_target", "Nie podano ścieżki docelowej"))
		return result
	}

	// 1. Przygotuj folder tymczasowy
	tmpDir := req.TargetPath
	if req.Format == "zip" {
		tmpDir = filepath.Join(os.TempDir(), "launcher_export_tmp")
		os.RemoveAll(tmpDir)
	}

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		result.setMsg(uimsg.New("msg_impex_mkdir_failed", "Nie można utworzyć folderu: "+err.Error(), "error", err.Error()))
		return result
	}

	// 2. Eksportuj programy
	if len(req.ProgramNames) > 0 && programsData != nil {
		n := len(req.ProgramNames)
		result.Log = append(result.Log, LineOf(uimsg.New("count_msg_impex_exporting_programs", fmt.Sprintf("Eksportuję %d programów...", n), "n", itoa(n))))
		if err := exportPrograms(exeDir, programsData, req.ProgramNames, tmpDir, req.IncludeFiles); err != nil {
			result.Errors = append(result.Errors, errorLine("msg_impex_programs_error", "Programy: ", err))
		} else {
			result.ProgramCount = n
			result.Log = append(result.Log, LineOf(uimsg.New("count_msg_impex_exported_programs", fmt.Sprintf("OK — wyeksportowano %d programów", n), "n", itoa(n))))
		}
	}

	// 3. Eksportuj sterowniki
	if len(req.DriverTypes) > 0 && driversData != nil {
		n := len(req.DriverTypes)
		result.Log = append(result.Log, LineOf(uimsg.New("count_msg_impex_exporting_drivers", fmt.Sprintf("Eksportuję %d typów sterowników...", n), "n", itoa(n))))
		if err := exportDrivers(exeDir, driversData, req.DriverTypes, req.DriverSelections, tmpDir, req.IncludeFiles); err != nil {
			result.Errors = append(result.Errors, errorLine("msg_impex_drivers_error", "Sterowniki: ", err))
		} else {
			result.DriverCount = n
			result.Log = append(result.Log, LineOf(uimsg.New("count_msg_impex_exported_drivers", fmt.Sprintf("OK — wyeksportowano %d typów sterowników", n), "n", itoa(n))))
		}
	}

	// 4. Generuj README.txt
	if result.ProgramCount > 0 || result.DriverCount > 0 {
		generateReadme(tmpDir, result.ProgramCount, result.DriverCount, req.Format)
		result.Log = append(result.Log, LineOf(uimsg.New("msg_impex_readme_generated", "Wygenerowano README.txt")))
	}

	// 5. Jeśli format ZIP, spakuj
	finalPath := tmpDir
	if req.Format == "zip" {
		zipPath := req.TargetPath
		if !strings.HasSuffix(strings.ToLower(zipPath), ".zip") {
			zipPath += ".zip"
		}
		if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
			result.setMsg(uimsg.New("msg_impex_zip_mkdir_failed", "Nie można utworzyć folderu ZIP: "+err.Error(), "error", err.Error()))
			os.RemoveAll(tmpDir)
			return result
		}
		if err := zipDirectory(tmpDir, zipPath); err != nil {
			result.setMsg(uimsg.New("msg_impex_zip_failed", "Błąd pakowania ZIP: "+err.Error(), "error", err.Error()))
			os.RemoveAll(tmpDir)
			return result
		}
		finalPath = zipPath
		os.RemoveAll(tmpDir)
	}

	result.Success = true
	result.setMsg(uimsg.New("msg_impex_export_done",
		fmt.Sprintf("Eksport zakończony. Programy: %d, typy sterowników: %d.", result.ProgramCount, result.DriverCount),
		"programs", itoa(result.ProgramCount), "drivers", itoa(result.DriverCount)))
	result.ExportedPath = finalPath
	return result
}

// ImportPack imports a folder or ZIP created by ExportPack into the current catalogues.
// It mutates programsData/driversData; the caller is responsible for saving them.
func ImportPack(exeDir string, programsData *programs.ProgramsData, driversData *drivers.DriversData, req ImportPackRequest) ImportPackResult {
	result := ImportPackResult{Success: false}
	if req.PackPath == "" {
		result.setMsg(uimsg.New("msg_impex_no_pack", "Nie podano paczki do importu"))
		return result
	}
	if req.Strategy == "" {
		req.Strategy = "overwrite"
	}

	packDir := req.PackPath
	if strings.HasSuffix(strings.ToLower(req.PackPath), ".zip") {
		tmpDir, err := os.MkdirTemp("", "coredeck-import-*")
		if err != nil {
			result.setMsg(tempDirFailed(err))
			return result
		}
		defer os.RemoveAll(tmpDir)
		if err := unzip(req.PackPath, tmpDir); err != nil {
			result.setMsg(unzipFailed(err))
			return result
		}
		packDir = tmpDir
	}

	if req.ImportPrograms {
		count, err := importPrograms(packDir, programsData, req)
		if err != nil {
			result.Errors = append(result.Errors, errorLine("msg_impex_programs_error", "Programy: ", err))
		} else {
			result.ProgramCount = count
		}
	}

	if req.ImportDrivers {
		count, err := importDrivers(packDir, driversData, req)
		if err != nil {
			result.Errors = append(result.Errors, errorLine("msg_impex_drivers_error", "Sterowniki: ", err))
		} else {
			result.DriverCount = count
		}
	}

	result.Success = len(result.Errors) == 0
	if result.Success {
		result.setMsg(uimsg.New("msg_impex_import_done",
			fmt.Sprintf("Import zakończony. Programy: %d, typy sterowników: %d.", result.ProgramCount, result.DriverCount),
			"programs", itoa(result.ProgramCount), "drivers", itoa(result.DriverCount)))
	} else {
		result.setMsg(uimsg.New("msg_impex_import_errors", "Import zakończony z błędami"))
	}
	return result
}

func PreviewPack(packPath string) PackPreview {
	preview := PackPreview{Success: false}
	if strings.TrimSpace(packPath) == "" {
		preview.setMsg(uimsg.New("msg_impex_no_pack_path", "Nie podano ścieżki paczki"))
		return preview
	}

	packDir := packPath
	if strings.HasSuffix(strings.ToLower(packPath), ".zip") {
		tmpDir, err := os.MkdirTemp("", "coredeck-preview-*")
		if err != nil {
			preview.setMsg(tempDirFailed(err))
			return preview
		}
		defer os.RemoveAll(tmpDir)
		if err := unzip(packPath, tmpDir); err != nil {
			preview.setMsg(unzipFailed(err))
			return preview
		}
		packDir = tmpDir
	}

	programsPath := filepath.Join(packDir, "programs.json")
	if data, err := os.ReadFile(programsPath); err == nil {
		var pd programs.ProgramsData
		if err := json.Unmarshal(data, &pd); err == nil {
			for _, p := range pd.Programs {
				preview.Programs = append(preview.Programs, PackProgramInfo{
					Name:        p.Name,
					InstallType: p.InstallType,
					Variants:    len(p.Variants),
					Category:    p.Category,
				})
			}
			preview.ProgramCount = len(preview.Programs)
		}
	}

	driversPath := filepath.Join(packDir, "drivers.json")
	if data, err := os.ReadFile(driversPath); err == nil {
		var dd drivers.DriversData
		if err := json.Unmarshal(data, &dd); err == nil {
			for key, dt := range dd.Types {
				series := 0
				for _, brand := range dt.Brands {
					series += len(brand.Series)
				}
				preview.Drivers = append(preview.Drivers, PackDriverInfo{
					Key:    key,
					Title:  firstNonEmpty(dt.Title, key),
					Brands: len(dt.Brands),
					Series: series,
				})
			}
			preview.DriverCount = len(preview.Drivers)
		}
	}

	if preview.ProgramCount == 0 && preview.DriverCount == 0 {
		preview.setMsg(uimsg.New("msg_impex_pack_empty", "Paczka nie zawiera programs.json ani drivers.json albo pliki są niepoprawne"))
		return preview
	}

	preview.Success = true
	preview.setMsg(uimsg.New("msg_impex_preview_found",
		fmt.Sprintf("Znaleziono w paczce. Programy: %d, typy sterowników: %d.", preview.ProgramCount, preview.DriverCount),
		"programs", itoa(preview.ProgramCount), "drivers", itoa(preview.DriverCount)))
	return preview
}

func tempDirFailed(err error) uimsg.Msg {
	return uimsg.New("msg_impex_temp_failed", "Nie można utworzyć katalogu tymczasowego: "+err.Error(), "error", err.Error())
}

func unzipFailed(err error) uimsg.Msg {
	if m := uimsg.From(err); m.Code != "" {
		return m
	}
	return uimsg.New("msg_impex_unzip_failed", "Nie można rozpakować ZIP: "+err.Error(), "error", err.Error())
}

func importPrograms(packDir string, pd *programs.ProgramsData, req ImportPackRequest) (int, error) {
	if pd == nil {
		return 0, uimsg.Errorf("msg_impex_no_programs_catalog", "Brak katalogu programów")
	}
	data, err := os.ReadFile(filepath.Join(packDir, "programs.json"))
	if err != nil {
		return 0, err
	}
	var imported programs.ProgramsData
	if err := json.Unmarshal(data, &imported); err != nil {
		return 0, err
	}
	selected := stringSet(req.SelectedPrograms)
	importAll := len(selected) == 0

	count := 0
	for _, p := range imported.Programs {
		if !importAll && !selected[p.Name] {
			continue
		}
		// Resolve catalogue conflicts before any file operation: skip must not
		// replace existing files, and duplicate needs its own destination.
		conflict := req.ProgramsTargetDir != "" && exists(filepath.Join(req.ProgramsTargetDir, safeName(p.Name)))
		for _, existing := range pd.Programs {
			if existing.Name == p.Name {
				conflict = true
				break
			}
		}
		if conflict {
			switch req.Strategy {
			case "skip":
				continue
			case "duplicate":
				base := p.Name + "_imported"
				p.Name = uniqueProgramName(pd, base)
				for i := 2; req.ProgramsTargetDir != "" && exists(filepath.Join(req.ProgramsTargetDir, safeName(p.Name))); i++ {
					p.Name = uniqueProgramName(pd, fmt.Sprintf("%s_%d", base, i))
				}
			}
		}
		p.Selected = false
		if req.ProgramsTargetDir != "" {
			rewriteAndCopyProgramPaths(packDir, req.ProgramsTargetDir, &p)
		}
		mergeProgram(pd, p, req.Strategy)
		count++
	}
	pd.TotalPrograms = len(pd.Programs)
	return count, nil
}

func importDrivers(packDir string, dd *drivers.DriversData, req ImportPackRequest) (int, error) {
	if dd == nil {
		return 0, uimsg.Errorf("msg_impex_no_drivers_catalog", "Brak katalogu sterowników")
	}
	data, err := os.ReadFile(filepath.Join(packDir, "drivers.json"))
	if err != nil {
		return 0, err
	}
	var imported drivers.DriversData
	if err := json.Unmarshal(data, &imported); err != nil {
		return 0, err
	}
	selected := stringSet(req.SelectedDrivers)
	importAll := len(selected) == 0
	if dd.Types == nil {
		dd.Types = map[string]drivers.DriverType{}
	}

	count := 0
	for typeKey, dt := range imported.Types {
		if !importAll && !selected[typeKey] {
			continue
		}
		destKey := typeKey
		_, conflict := dd.Types[destKey]
		conflict = conflict || (req.DriversTargetDir != "" && exists(filepath.Join(req.DriversTargetDir, safeName(destKey))))
		if conflict {
			switch req.Strategy {
			case "skip":
				continue
			case "duplicate":
				base := typeKey + "_imported"
				destKey = uniqueDriverTypeKey(dd, base)
				for i := 2; req.DriversTargetDir != "" && exists(filepath.Join(req.DriversTargetDir, safeName(destKey))); i++ {
					destKey = uniqueDriverTypeKey(dd, fmt.Sprintf("%s_%d", base, i))
				}
			}
		}
		if req.DriversTargetDir != "" {
			rewriteAndCopyDriverPaths(packDir, req.DriversTargetDir, destKey, &dt)
		}
		dd.Types[destKey] = dt
		count++
	}
	return count, nil
}

func rewriteAndCopyProgramPaths(packDir, targetDir string, p *programs.Program) {
	programTarget := filepath.Join(targetDir, safeName(p.Name))
	if len(p.Variants) > 0 {
		for i := range p.Variants {
			v := &p.Variants[i]
			old := v.FolderPath
			if old == "" {
				old = p.FolderPath
			}
			src := filepath.Join(packDir, filepath.FromSlash(old))
			dst := filepath.Join(programTarget, safeName(firstNonEmpty(v.Name, fmt.Sprintf("variant-%d", i+1))))
			if exists(src) {
				os.RemoveAll(dst)
				_ = copyDir(src, dst)
			}
			v.FolderPath = dst
			if i == p.DefaultVariant || (p.DefaultVariant == 0 && i == 0) {
				p.FolderPath = dst
			}
		}
		if p.FolderPath == "" && len(p.Variants) > 0 {
			p.FolderPath = p.Variants[0].FolderPath
		}
		return
	}
	src := filepath.Join(packDir, filepath.FromSlash(p.FolderPath))
	if exists(src) {
		os.RemoveAll(programTarget)
		_ = copyDir(src, programTarget)
		p.FolderPath = programTarget
	}
}

func rewriteAndCopyDriverPaths(packDir, targetDir, typeKey string, dt *drivers.DriverType) {
	for brandKey, brand := range dt.Brands {
		for i := range brand.Series {
			series := &brand.Series[i]
			src := filepath.Join(packDir, filepath.FromSlash(series.FolderPath))
			dst := filepath.Join(targetDir, safeName(typeKey), safeName(brandKey), safeName(series.Name))
			if exists(src) {
				os.RemoveAll(dst)
				_ = copyDir(src, dst)
				series.FolderPath = dst
			}
		}
		dt.Brands[brandKey] = brand
	}
}

func mergeProgram(pd *programs.ProgramsData, p programs.Program, strategy string) {
	for i, existing := range pd.Programs {
		if existing.Name != p.Name {
			continue
		}
		switch strategy {
		case "skip":
			return
		case "duplicate":
			p.Name = uniqueProgramName(pd, p.Name+"_imported")
			pd.Programs = append(pd.Programs, p)
			return
		default:
			pd.Programs[i] = p
			return
		}
	}
	pd.Programs = append(pd.Programs, p)
}

func uniqueProgramName(pd *programs.ProgramsData, base string) string {
	existsName := func(name string) bool {
		for _, p := range pd.Programs {
			if p.Name == name {
				return true
			}
		}
		return false
	}
	name := base
	for i := 2; existsName(name); i++ {
		name = fmt.Sprintf("%s_%d", base, i)
	}
	return name
}

func uniqueDriverTypeKey(dd *drivers.DriversData, base string) string {
	key := base
	for i := 2; ; i++ {
		if _, ok := dd.Types[key]; !ok {
			return key
		}
		key = fmt.Sprintf("%s_%d", base, i)
	}
}

func stringSet(values []string) map[string]bool {
	m := map[string]bool{}
	for _, v := range values {
		m[v] = true
	}
	return m
}

func exportPrograms(exeDir string, pd *programs.ProgramsData, names []string, targetDir string, includeFiles bool) error {
	// Filtruj programy
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	filtered := []programs.Program{}
	for _, p := range pd.Programs {
		if nameSet[p.Name] {
			// Export rewrites variant paths; keep the live catalogue's slice
			// independent from the portable snapshot being prepared below.
			p.Variants = slices.Clone(p.Variants)
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 {
		return uimsg.Errorf("msg_impex_no_programs_found", "Nie znaleziono żadnego programu do eksportu")
	}

	if includeFiles {
		programsDir := filepath.Join(targetDir, "programs")
		if err := os.MkdirAll(programsDir, 0755); err != nil {
			return err
		}
		for pi := range filtered {
			p := &filtered[pi]
			programSlug := safeName(p.Name)
			if len(p.Variants) > 0 {
				for vi := range p.Variants {
					v := &p.Variants[vi]
					srcPath := paths.ResolvePath(exeDir, firstNonEmpty(v.FolderPath, p.FolderPath))
					if srcPath == "" || !exists(srcPath) {
						slog.Warn("program variant folder not found", "name", p.Name, "variant", v.Name, "path", v.FolderPath)
						continue
					}
					variantSlug := safeName(firstNonEmpty(v.Name, fmt.Sprintf("variant-%d", vi+1)))
					relPath := filepath.ToSlash(filepath.Join("programs", programSlug, variantSlug))
					destPath := filepath.Join(targetDir, relPath)
					if err := copyDir(srcPath, destPath); err != nil {
						slog.Warn("failed to copy program variant", "name", p.Name, "variant", v.Name, "error", err)
						continue
					}
					v.FolderPath = relPath
					if vi == p.DefaultVariant || (p.DefaultVariant == 0 && vi == 0) {
						p.FolderPath = relPath
					}
				}
				if p.FolderPath == "" && len(p.Variants) > 0 {
					p.FolderPath = p.Variants[0].FolderPath
				}
				continue
			}

			srcPath := paths.ResolvePath(exeDir, p.FolderPath)
			if srcPath == "" || !exists(srcPath) {
				slog.Warn("program folder not found", "name", p.Name, "path", p.FolderPath)
				continue
			}
			relPath := filepath.ToSlash(filepath.Join("programs", programSlug))
			destPath := filepath.Join(targetDir, relPath)
			if err := copyDir(srcPath, destPath); err != nil {
				slog.Warn("failed to copy program", "name", p.Name, "error", err)
				continue
			}
			p.FolderPath = relPath
		}
	}

	// Zapisz programs.json po ewentualnym przepisaniu folderPath na ścieżki paczki.
	exportData := programs.ProgramsData{
		Version:       pd.Version,
		TotalPrograms: len(filtered),
		Profiles:      pd.Profiles,
		Programs:      filtered,
	}
	out, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(targetDir, "programs.json"), out, 0644); err != nil {
		return err
	}

	return nil
}

func exportDrivers(exeDir string, dd *drivers.DriversData, typeKeys []string, selections map[string]DriverSelection, targetDir string, includeFiles bool) error {
	// Filtruj sterowniki
	typeSet := make(map[string]bool)
	for _, tk := range typeKeys {
		typeSet[tk] = true
	}
	filteredTypes := make(map[string]drivers.DriverType)
	for tk, dt := range dd.Types {
		if typeSet[tk] {
			// Both path rewriting and selection filtering mutate this tree.
			// Copy its map and nested slices before preparing the export.
			if dt.Brands != nil {
				brands := make(map[string]drivers.DriverBrand, len(dt.Brands))
				for key, brand := range dt.Brands {
					brand.Series = slices.Clone(brand.Series)
					brands[key] = brand
				}
				dt.Brands = brands
			}
			filteredTypes[tk] = dt
		}
	}

	if len(filteredTypes) == 0 {
		return uimsg.Errorf("msg_impex_no_drivers_found", "Nie znaleziono żadnego typu sterownika do eksportu")
	}

	if includeFiles {
		driversDir := filepath.Join(targetDir, "drivers")
		if err := os.MkdirAll(driversDir, 0755); err != nil {
			return err
		}
		for typeKey, dt := range filteredTypes {
			sel := selections[typeKey]
			for brandKey, brand := range dt.Brands {
				brandSel := sel.Brands[brandKey]
				brandSeriesSet := stringSet(brandSel.Series)
				allSeries := len(brandSel.Series) == 0

				for si := range brand.Series {
					series := &brand.Series[si]
					if !allSeries && !brandSeriesSet[series.Name] {
						// Ta seria nie została zaznaczona — usuń z filteredTypes
						continue
					}
					srcPath := paths.ResolvePath(exeDir, series.FolderPath)
					if srcPath == "" || !exists(srcPath) {
						slog.Warn("driver folder not found", "type", typeKey, "brand", brandKey, "series", series.Name, "path", series.FolderPath)
						continue
					}
					relPath := filepath.ToSlash(filepath.Join("drivers", safeName(typeKey), safeName(brandKey), safeName(series.Name)))
					destPath := filepath.Join(targetDir, relPath)
					if err := copyDir(srcPath, destPath); err != nil {
						slog.Warn("failed to copy driver", "type", typeKey, "brand", brandKey, "series", series.Name, "error", err)
						continue
					}
					series.FolderPath = relPath
				}

				// Jeśli seria nie została zaznaczona — usuń ją z eksportu
				if !allSeries && len(brandSel.Series) > 0 {
					filtered := brand.Series[:0]
					for _, s := range brand.Series {
						if brandSeriesSet[s.Name] {
							filtered = append(filtered, s)
						}
					}
					brand.Series = filtered
				}
				dt.Brands[brandKey] = brand
			}
			filteredTypes[typeKey] = dt
		}
	}

	// Zapisz drivers.json po ewentualnym przepisaniu folderPath na ścieżki paczki.
	exportData := drivers.DriversData{
		Version: dd.Version,
		Types:   filteredTypes,
	}
	out, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(targetDir, "drivers.json"), out, 0644); err != nil {
		return err
	}

	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func safeName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "item"
	}
	re := regexp.MustCompile(`[<>:"/\\|?*]+`)
	s = re.ReplaceAllString(s, "_")
	s = strings.Trim(s, " .")
	if s == "" {
		return "item"
	}
	return s
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

func zipDirectory(src, dst string) error {
	outFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	defer w.Close()

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		if rel == "." {
			return nil
		}
		if d.IsDir() {
			_, err := w.Create(rel + "/")
			return err
		}
		f, err := w.Create(rel)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = f.Write(data)
		return err
	})
}

func unzip(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	cleanDst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cleanDst, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		path := filepath.Join(cleanDst, f.Name)
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(absPath, cleanDst+string(os.PathSeparator)) && absPath != cleanDst {
			return uimsg.Errorf("msg_impex_zip_bad_path", "Nieprawidłowa ścieżka w ZIP: "+f.Name, "path", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(absPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(absPath)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		closeErr := out.Close()
		rc.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
