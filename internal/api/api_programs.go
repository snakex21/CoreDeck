package api

// Program catalogue, async install sessions and program CRUD.

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"launcher/internal/paths"
	"launcher/internal/programs"
	"launcher/internal/uimsg"
)

func (a *API) GetProgramsData() string {
	if a.ProgramsData == nil {
		if pd, err := loadCatalogue(a.WebDir, "programs.json", programs.LoadPrograms, programs.LoadFromFS, a.WebFS); err == nil {
			a.ProgramsData = pd
		} else {
			slog.Error("load programs", "error", err)
			return `{"error": "Cannot load programs.json"}`
		}
	}
	data, _ := json.Marshal(a.ProgramsData)
	return string(data)
}

func (a *API) GetProgramsForProfile(profileName string) string {
	if a.ProgramsData == nil {
		a.GetProgramsData()
	}
	if a.ProgramsData == nil {
		return `[]`
	}
	result := programs.GetProgramsForProfile(a.ProgramsData, profileName)
	data, _ := json.Marshal(result)
	return string(data)
}

// --- Async install with live progress ---

type installSession struct {
	mu        sync.Mutex
	progress  programs.InstallProgress
	done      bool
	final     []programs.InstallResult
	err       string
	cancelled bool
}

var installSessions sync.Map
var sessionCounter int64

func (a *API) InstallSelectedPrograms(programsJSON string, targetDir string) string {
	if a.ProgramsData == nil {
		a.GetProgramsData()
	}
	if a.ProgramsData == nil {
		return `{"error": "No programs data"}`
	}

	var selected []programs.Program
	if err := json.Unmarshal([]byte(programsJSON), &selected); err != nil {
		return `{"error": "Invalid programs JSON"}`
	}

	if targetDir == "" {
		targetDir = defaultTargetDir()
	}
	os.MkdirAll(targetDir, 0755)

	sessionID := fmt.Sprintf("inst-%d", atomic.AddInt64(&sessionCounter, 1))
	session := &installSession{}
	installSessions.Store(sessionID, session)

	exeDir := a.ExeDir
	programNames := make([]string, 0, len(selected))
	for _, program := range selected {
		programNames = append(programNames, program.Name)
	}
	a.recordAudit("install.programs", targetDir, "started", sessionID, map[string]any{
		"programs": programNames,
		"count":    len(programNames),
	})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				session.mu.Lock()
				session.err = fmt.Sprintf("install panic: %v", r)
				session.done = true
				session.mu.Unlock()
				a.recordAudit("install.programs", targetDir, "failed", sessionID, map[string]any{
					"error": fmt.Sprintf("panic: %v", r),
				})
			}
		}()

		cancelCheck := func() bool {
			session.mu.Lock()
			defer session.mu.Unlock()
			return session.cancelled
		}

		results := programs.InstallPrograms(exeDir, selected, targetDir, func(p programs.InstallProgress) {
			session.mu.Lock()
			session.progress = p
			session.mu.Unlock()
		}, cancelCheck)

		session.mu.Lock()
		session.final = results
		session.done = true
		cancelled := session.cancelled
		session.mu.Unlock()
		slog.Info("install session completed", "id", sessionID, "results", len(results))
		succeeded := 0
		for _, result := range results {
			if result.Success {
				succeeded++
			}
		}
		status := "completed"
		if cancelled {
			status = "cancelled"
		}
		a.recordAudit("install.programs", targetDir, status, sessionID, map[string]any{
			"requested": len(selected),
			"completed": len(results),
			"succeeded": succeeded,
			"failed":    len(results) - succeeded,
		})

		// Self-cleanup: remove session after 30s (in case user closed overlay)
		time.AfterFunc(30*time.Second, func() { installSessions.Delete(sessionID) })
	}()

	data, _ := json.Marshal(map[string]string{"sessionId": sessionID})
	return string(data)
}

func (a *API) CheckInstallProgress(sessionID string) string {
	val, ok := installSessions.Load(sessionID)
	if !ok {
		return `{"error":"Invalid session","done":true}`
	}
	session := val.(*installSession)

	session.mu.Lock()
	defer session.mu.Unlock()

	p := session.progress
	resp := map[string]interface{}{
		"done":      session.done,
		"cancelled": session.cancelled,
		"current":   p.Current,
		"total":     p.Total,
		"program":   p.Program,
		"status":    p.Status,
	}

	if session.done && len(session.final) > 0 {
		resp["results"] = session.final
	} else if len(p.Results) > 0 {
		resp["results"] = p.Results
	}

	if session.err != "" {
		resp["error"] = session.err
	}

	data, _ := json.Marshal(resp)

	if session.done {
		installSessions.Delete(sessionID)
	}

	return string(data)
}

// defaultTargetDir is <system drive>\Programs. filepath.Join("C:", "Programs")
// would give the drive-relative "C:Programs", so the separator is explicit.
func defaultTargetDir() string {
	drive := os.Getenv("SYSTEMDRIVE")
	if drive == "" {
		drive = "C:"
	}
	return filepath.Join(drive+string(filepath.Separator), "Programs")
}

// LaunchProgram starts a portable program (or tool) straight from the
// catalogue folder. The program is looked up by name in the catalogue, so the
// page can only choose which entry and, optionally, which .exe inside that
// entry's folder (file, relative to the folder) to start.
func (a *API) LaunchProgram(name string, variantIndex int, file string) string {
	if a.ProgramsData == nil {
		a.GetProgramsData()
	}
	if a.ProgramsData == nil {
		return jsonMsg(uimsg.New("msg_prog_no_catalog", "Brak katalogu programów"))
	}
	for _, p := range a.ProgramsData.Programs {
		if p.Name != name {
			continue
		}
		result := programs.LaunchProgram(a.ExeDir, p, variantIndex, file)
		status := "failed"
		switch {
		case result.Launched != "":
			status = "launched"
		case len(result.Candidates) > 0:
			status = "choose"
		case result.OpenedFolder:
			status = "opened_folder"
		}
		a.recordAudit("program.launch", name, status, "", map[string]any{"path": result.Launched})
		data, _ := json.Marshal(result)
		return string(data)
	}
	return jsonMsg(uimsg.New("msg_prog_not_in_catalog", "Nie znaleziono programu w katalogu: "+name, "name", name))
}

func (a *API) OpenTargetFolder(path string) error {
	if path == "" {
		path = defaultTargetDir()
	}
	os.MkdirAll(path, 0755)
	return exec.Command("explorer", path).Start()
}

// ScanInstallerCandidates scans a folder for installer/script files
// and returns them as a JSON array of relative paths.
func (a *API) ScanInstallerCandidates(folderPath string) string {
	resolved := paths.ResolvePath(a.ExeDir, folderPath)

	type candidate struct {
		Path  string `json:"path"`
		Name  string `json:"name"`
		Score int    `json:"score"`
	}

	var candidates []candidate

	filepath.WalkDir(resolved, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if strings.Contains(name, "uninstall") || strings.Contains(name, "unins") || strings.Contains(name, "postinstall") {
			return nil
		}

		score := 0
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".msi":
			score = 90
		case ".msixbundle", ".appx", ".appxbundle":
			score = 85
		case ".ps1":
			score = 85
		case ".cmd", ".bat":
			score = 80
		case ".exe":
			score = 70
		default:
			return nil
		}
		if strings.Contains(name, "setup") || strings.Contains(name, "install") || strings.Contains(name, "installer") {
			score += 20
		}

		rel, err := filepath.Rel(resolved, path)
		if err != nil {
			rel = d.Name()
		}
		candidates = append(candidates, candidate{
			Path:  filepath.ToSlash(rel),
			Name:  d.Name(),
			Score: score,
		})
		return nil
	})

	// Sort by score descending, then by name
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].Score > candidates[i].Score ||
				(candidates[j].Score == candidates[i].Score && candidates[j].Name < candidates[i].Name) {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	data, _ := json.Marshal(candidates)
	return string(data)
}

// AddProgram adds or updates a program in the external programs.json and
// persists it to <exeDir>/data/programs.json. If a program with the same
// name already exists it is replaced, otherwise the program is appended.
func (a *API) AddProgram(programJSON string) string {
	if a.ProgramsData == nil {
		a.GetProgramsData()
	}
	if a.ProgramsData == nil {
		return `{"error": "No programs data"}`
	}

	var p programs.Program
	if err := json.Unmarshal([]byte(programJSON), &p); err != nil {
		return `{"error": "Invalid program JSON"}`
	}
	if p.Name == "" {
		return jsonMsg(uimsg.New("msg_prog_name_required", "Nazwa programu jest wymagana"))
	}

	replaced := false
	for i := range a.ProgramsData.Programs {
		if a.ProgramsData.Programs[i].Name == p.Name {
			a.ProgramsData.Programs[i] = p
			replaced = true
			break
		}
	}
	if !replaced {
		a.ProgramsData.Programs = append(a.ProgramsData.Programs, p)
	}

	if err := programs.SavePrograms(a.WebDir, a.ProgramsData); err != nil {
		return jsonErr(err)
	}
	return `{"success": true}`
}

// DeleteProgram removes a program by name from the external programs.json
// and persists the change to <exeDir>/data/programs.json.
func (a *API) DeleteProgram(name string) string {
	if a.ProgramsData == nil {
		a.GetProgramsData()
	}
	if a.ProgramsData == nil {
		return `{"error": "No programs data"}`
	}

	filtered := make([]programs.Program, 0, len(a.ProgramsData.Programs))
	for _, p := range a.ProgramsData.Programs {
		if p.Name != name {
			filtered = append(filtered, p)
		}
	}
	a.ProgramsData.Programs = filtered

	if err := programs.SavePrograms(a.WebDir, a.ProgramsData); err != nil {
		return jsonErr(err)
	}
	return `{"success": true}`
}
