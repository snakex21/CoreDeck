package programs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"launcher/internal/jsonutil"
	"launcher/internal/paths"
	"launcher/internal/safeio"
)

type Variant struct {
	Name          string `json:"name"`
	FolderPath    string `json:"folderPath"`
	InstallType   string `json:"installType,omitempty"`   // opcjonalnie nadpisuje główny typ
	InstallerFile string `json:"installerFile,omitempty"` // konkretny .exe/.msi do uruchomienia
	InstallerArgs string `json:"installerArgs,omitempty"` // opcjonalne argumenty instalatora
	RunFromTemp   bool   `json:"runFromTemp,omitempty"`   // skopiuj folder do TEMP, uruchom kopię, potem usuń
}

type ProgramLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Kind  string `json:"kind,omitempty"`
}

type Program struct {
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	Category        string         `json:"category"`
	Icon            string         `json:"icon"`
	FolderPath      string         `json:"folderPath"`
	Tags            []string       `json:"tags"`
	Type            string         `json:"type,omitempty"`
	LocalPath       string         `json:"localPath,omitempty"`
	Details         string         `json:"details,omitempty"`
	Links           []ProgramLink  `json:"links,omitempty"`
	ExternalURL     string         `json:"externalUrl,omitempty"`
	InstallType     string         `json:"installType"`
	InstallerFile   string         `json:"installerFile,omitempty"` // konkretny .exe/.msi do uruchomienia
	InstallerArgs   string         `json:"installerArgs,omitempty"` // opcjonalne argumenty instalatora
	RunFromTemp     bool           `json:"runFromTemp,omitempty"`   // skopiuj folder do TEMP, uruchom kopię, potem usuń
	Selected        bool           `json:"selected"`
	ProfileRelevant []string       `json:"profileRelevant"`
	Variants        []Variant      `json:"variants,omitempty"`
	DefaultVariant  int            `json:"defaultVariant,omitempty"`
	SelectedVariant int            `json:"selectedVariant,omitempty"` // indeks wybranego wariantu z UI
	ProfileVariants map[string]int `json:"profileVariants,omitempty"` // wariant per profil (np. {"laptop": 1, "ai_box": 0})
}

type ProgramsData struct {
	Version       string             `json:"version"`
	TotalPrograms int                `json:"totalPrograms"`
	Profiles      map[string]Profile `json:"profiles"`
	Programs      []Program          `json:"programs"`
}

type Profile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type InstallResult struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Path    string `json:"path,omitempty"`
	// Outcome is one of OutcomeOK, OutcomeOKNote, OutcomeCancelled,
	// OutcomeError. Cancelled duplicates OutcomeCancelled for older UIs.
	Outcome   string `json:"outcome,omitempty"`
	Cancelled bool   `json:"cancelled,omitempty"`
	ExitCode  int    `json:"exitCode,omitempty"`
}

type InstallProgress struct {
	Current int             `json:"current"`
	Total   int             `json:"total"`
	Program string          `json:"program"`
	Status  string          `json:"status"`
	Results []InstallResult `json:"results"`
	Done    bool            `json:"done"`
}

func LoadPrograms(webDir string) (*ProgramsData, error) {
	jsonPath := filepath.Join(webDir, "data", "programs.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read programs.json: %w", err)
	}
	var pd ProgramsData
	if err := jsonutil.Unmarshal(data, &pd); err != nil {
		return nil, fmt.Errorf("cannot parse programs.json: %w", err)
	}
	return &pd, nil
}

// SavePrograms writes the programs data back to the external file
// <webDir>/data/programs.json, creating the data directory if needed.
func SavePrograms(webDir string, data *ProgramsData) error {
	if data != nil {
		data.TotalPrograms = len(data.Programs)
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Join(webDir, "data")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return safeio.WriteFileAtomic(filepath.Join(dir, "programs.json"), out, 0o644)
}

func LoadFromFS(fsys fs.FS) (*ProgramsData, error) {
	data, err := fs.ReadFile(fsys, "web/data/programs.json")
	if err != nil {
		return nil, fmt.Errorf("cannot read programs.json from FS: %w", err)
	}
	var pd ProgramsData
	if err := jsonutil.Unmarshal(data, &pd); err != nil {
		return nil, fmt.Errorf("cannot parse programs.json: %w", err)
	}
	return &pd, nil
}

func GetProgramsForProfile(pd *ProgramsData, profileName string) []Program {
	var result []Program
	for _, p := range pd.Programs {
		for _, pr := range p.ProfileRelevant {
			if pr == profileName || pr == "common" {
				result = append(result, p)
				break
			}
		}
	}
	return result
}

func InstallPrograms(exeDir string, programs []Program, targetDir string, progressCb func(InstallProgress), cancelCheck func() bool) []InstallResult {
	total := len(programs)

	type job struct {
		index int
		p     Program
		cfg   installConfig
		path  string
	}

	var copyJobs []job
	var interactiveJobs []job

	for i, p := range programs {
		cfg := resolveVariant(p, exeDir)
		path := paths.ResolvePath(exeDir, cfg.FolderPath)
		j := job{index: i, p: p, cfg: cfg, path: path}

		switch cfg.InstallType {
		case "portable", "tool", "addon", "collection":
			copyJobs = append(copyJobs, j)
		default:
			interactiveJobs = append(interactiveJobs, j)
		}
	}

	var results []InstallResult
	var mu sync.Mutex
	done := 0

	reportProgress := func(name, status string) {
		mu.Lock()
		snapshot := make([]InstallResult, len(results))
		copy(snapshot, results)
		cur := done
		mu.Unlock()
		progressCb(InstallProgress{
			Current: cur,
			Total:   total,
			Program: name,
			Status:  status,
			Results: snapshot,
		})
	}

	// --- Phase 1: Parallel copies (portable, tool, addon, collection) ---
	var shortcutDirs []string // collect for batch shortcut creation
	if len(copyJobs) > 0 {
		slog.Info("install phase 1: parallel copies", "count", len(copyJobs))
		sem := make(chan struct{}, 6) // 6 concurrent copies
		var wg sync.WaitGroup

		for _, j := range copyJobs {
			if cancelCheck != nil && cancelCheck() {
				break
			}
			wg.Add(1)
			go func(j job) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				if cancelCheck != nil && cancelCheck() {
					return
				}

				result := installPortable(j.path, targetDir, j.p.Name)

				mu.Lock()
				results = append(results, result)
				done++
				if result.Success {
					shortcutDirs = append(shortcutDirs, filepath.Join(targetDir, j.p.Name)+"\x00"+j.p.Name)
				}
				mu.Unlock()
				reportProgress(j.p.Name, "done")
			}(j)
		}
		wg.Wait()

		// Batch shortcut creation — one PowerShell call for ALL shortcuts
		batchCreateShortcuts(shortcutDirs)
	}

	// --- Phase 1.5: Pre-stage TEMP copies for installers (parallel) ---
	type staged struct {
		installerPath string
		tempDir       string
	}
	stagedMap := make(map[int]staged)

	type stageJob struct {
		idx           int
		installerDir  string
		installerFile string
		name          string
	}
	var stageJobs []stageJob

	for _, j := range interactiveJobs {
		if j.cfg.InstallType == "installer" && j.cfg.RunFromTemp {
			ip := resolveInstallerPath(j.path, j.cfg.InstallerFile)
			if strings.TrimSpace(j.cfg.InstallerFile) != "" && ip == "" {
				continue // will error in phase 2
			}
			if ip == "" {
				ip = findInstaller(j.path)
			}
			if ip == "" {
				continue
			}
			stageJobs = append(stageJobs, stageJob{idx: j.index, installerDir: j.path, installerFile: ip, name: j.p.Name})
		}
	}

	if len(stageJobs) > 0 {
		slog.Info("install phase 1.5: pre-staging TEMP copies", "count", len(stageJobs))
		sem := make(chan struct{}, 4)
		var wg sync.WaitGroup

		for _, sj := range stageJobs {
			wg.Add(1)
			go func(sj stageJob) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				tempDir, err := os.MkdirTemp("", "launcher-installer-*")
				if err != nil {
					return
				}
				if err := copyDirFast(sj.installerDir, tempDir); err != nil {
					_ = os.RemoveAll(tempDir)
					return
				}
				rel, err := filepath.Rel(sj.installerDir, sj.installerFile)
				if err != nil || strings.HasPrefix(rel, "..") {
					_ = os.RemoveAll(tempDir)
					return
				}
				tempPath := filepath.Join(tempDir, rel)

				mu.Lock()
				stagedMap[sj.idx] = staged{installerPath: tempPath, tempDir: tempDir}
				mu.Unlock()
			}(sj)
		}
		wg.Wait()
	}

	// --- Phase 2: Parallel interactive (installer, script, other) ---
	if len(interactiveJobs) > 0 {
		slog.Info("install phase 2: sequential interactive", "count", len(interactiveJobs))
		sem := make(chan struct{}, 1) // 1 installer/script/manual at a time
		var wg sync.WaitGroup

		for _, j := range interactiveJobs {
			if cancelCheck != nil && cancelCheck() {
				break
			}
			wg.Add(1)
			go func(j job) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				if cancelCheck != nil && cancelCheck() {
					return
				}

				reportProgress(j.p.Name, "installing")

				var result InstallResult
				result.Name = j.p.Name
				result.Type = j.cfg.InstallType

				switch j.cfg.InstallType {
				case "installer":
					// Use pre-staged TEMP if available, otherwise run normally
					if s, ok := stagedMap[j.index]; ok {
						result = runInstallerStaged(s.installerPath, s.tempDir, j.p.Name, j.cfg.InstallerArgs)
					} else {
						result = runInstaller(j.path, j.p.Name, j.cfg.InstallerFile, j.cfg.InstallerArgs, j.cfg.RunFromTemp)
					}
				case "script":
					result = runScript(j.path, j.p.Name, j.cfg.InstallerFile, j.cfg.InstallerArgs)
				case "manual", "other":
					result = installManual(j.path, targetDir, j.p.Name, j.cfg.InstallType)
				default:
					result = InstallResult{
						Name:    j.p.Name,
						Success: false,
						Message: "Nieznany typ: " + j.cfg.InstallType + " - otwórz ręcznie",
						Type:    j.cfg.InstallType,
					}
				}

				mu.Lock()
				results = append(results, result)
				done++
				mu.Unlock()
				reportProgress(j.p.Name, "done")
			}(j)
		}
		wg.Wait()
	}

	progressCb(InstallProgress{
		Current: total,
		Total:   total,
		Status:  "finished",
		Results: results,
		Done:    true,
	})

	return results
}

// runInstallerStaged runs an installer from a pre-staged TEMP directory.
// The tempDir is always cleaned up after the installer exits.
func runInstallerStaged(installerPath, tempDir, name, installerArgs string) InstallResult {
	defer func() { _ = os.RemoveAll(tempDir) }()

	args := splitCommandLine(installerArgs)
	cmd, err := installerCommand(installerPath, args)
	if err != nil {
		return InstallResult{Name: name, Success: false, Message: err.Error(), Type: "installer"}
	}
	cmd.Dir = filepath.Dir(installerPath)

	if err = cmd.Start(); err != nil {
		return installerResult(name, installerPath, true, err)
	}

	slog.Info("staged installer started", "name", name, "path", installerPath)
	err = cmd.Wait()
	result := installerResult(name, installerPath, true, err)
	slog.Info("staged installer finished", "name", name, "outcome", result.Outcome, "code", result.ExitCode)
	return result
}

type installConfig struct {
	FolderPath    string
	InstallType   string
	InstallerFile string
	InstallerArgs string
	RunFromTemp   bool
}

// resolveVariant zwraca konfigurację instalacji dla wybranego wariantu.
// Jeśli program nie ma wariantów, zwraca główne wartości.
func resolveVariant(p Program, exeDir string) installConfig {
	cfg := installConfig{
		FolderPath:    p.FolderPath,
		InstallType:   p.InstallType,
		InstallerFile: p.InstallerFile,
		InstallerArgs: p.InstallerArgs,
		RunFromTemp:   p.RunFromTemp,
	}
	if len(p.Variants) == 0 {
		return cfg
	}

	// Wybierz wariant: SelectedVariant > DefaultVariant > 0
	idx := p.SelectedVariant
	if idx < 0 || idx >= len(p.Variants) {
		idx = p.DefaultVariant
	}
	if idx < 0 || idx >= len(p.Variants) {
		idx = 0
	}

	v := p.Variants[idx]
	if v.FolderPath != "" {
		cfg.FolderPath = v.FolderPath
	}
	if v.InstallType != "" {
		cfg.InstallType = v.InstallType
	}
	if v.InstallerFile != "" {
		cfg.InstallerFile = v.InstallerFile
	}
	if v.InstallerArgs != "" {
		cfg.InstallerArgs = v.InstallerArgs
	}
	if v.RunFromTemp {
		cfg.RunFromTemp = true
	}
	return cfg
}

func installPortable(srcDir, targetDir, name string) InstallResult {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return InstallResult{
			Name:    name,
			Success: false,
			Message: "Folder nie znaleziony: " + srcDir,
			Type:    "portable",
		}
	}

	destDir := filepath.Join(targetDir, name)
	if err := copyDirFast(srcDir, destDir); err != nil {
		return InstallResult{
			Name:    name,
			Success: false,
			Message: "Błąd kopiowania: " + err.Error(),
			Type:    "portable",
		}
	}

	return InstallResult{
		Name:    name,
		Success: true,
		Message: "Skopiowano do " + destDir,
		Type:    "portable",
		Path:    destDir,
	}
}

func installManual(srcDir, targetDir, name, installType string) InstallResult {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return InstallResult{
			Name:    name,
			Success: false,
			Message: "Folder nie znaleziony: " + srcDir,
			Type:    installType,
		}
	}

	destDir := filepath.Join(targetDir, name)
	if err := copyDirFast(srcDir, destDir); err != nil {
		return InstallResult{
			Name:    name,
			Success: false,
			Message: "Błąd kopiowania: " + err.Error(),
			Type:    installType,
		}
	}

	return InstallResult{
		Name:    name,
		Success: true,
		Message: "Skopiowano do " + destDir + " — wymagana ręczna instalacja/konfiguracja. Otwórz folder i wykonaj instrukcję ze szczegółów programu.",
		Type:    installType,
		Path:    destDir,
	}
}

func runInstaller(installerDir, name, installerFile, installerArgs string, runFromTemp bool) InstallResult {
	if _, err := os.Stat(installerDir); os.IsNotExist(err) {
		return InstallResult{
			Name:    name,
			Success: false,
			Message: "Folder nie znaleziony: " + installerDir,
			Type:    "installer",
		}
	}

	installerPath := resolveInstallerPath(installerDir, installerFile)
	if strings.TrimSpace(installerFile) != "" && installerPath == "" {
		return InstallResult{
			Name:    name,
			Success: false,
			Message: "Wskazany instalator nie istnieje albo nie jest obsługiwanym typem (.exe/.msi/.appx/.bat/.cmd/.ps1): " + installerFile,
			Type:    "installer",
		}
	}

	if installerPath == "" {
		installerPath = findInstaller(installerDir)
	}

	if installerPath == "" {
		return InstallResult{
			Name:    name,
			Success: false,
			Message: "Nie znaleziono instalatora w " + installerDir,
			Type:    "installer",
		}
	}

	cleanup := func() {}
	if runFromTemp {
		var tempDir string
		var err error
		installerPath, tempDir, err = copyInstallerDirToTemp(installerDir, installerPath, name)
		if err != nil {
			return InstallResult{
				Name:    name,
				Success: false,
				Message: "Błąd kopiowania instalatora do TEMP: " + err.Error(),
				Type:    "installer",
			}
		}
		cleanup = func() { _ = os.RemoveAll(tempDir) }
		defer cleanup()
	}

	args := splitCommandLine(installerArgs)
	cmd, err := installerCommand(installerPath, args)
	if err != nil {
		return InstallResult{Name: name, Success: false, Message: err.Error(), Type: "installer"}
	}
	cmd.Dir = filepath.Dir(installerPath)

	if err = cmd.Start(); err != nil {
		return installerResult(name, installerPath, runFromTemp, err)
	}

	slog.Info("installer started", "name", name, "path", installerPath)
	err = cmd.Wait()
	result := installerResult(name, installerPath, runFromTemp, err)
	slog.Info("installer finished", "name", name, "outcome", result.Outcome, "code", result.ExitCode)
	return result
}

func resolveInstallerPath(installerDir, installerFile string) string {
	installerFile = strings.TrimSpace(installerFile)
	if installerFile == "" {
		return ""
	}

	var path string
	if filepath.IsAbs(installerFile) {
		path = installerFile
	} else {
		path = filepath.Join(installerDir, installerFile)
	}

	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return ""
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".exe", ".msi", ".msixbundle", ".appx", ".appxbundle", ".bat", ".cmd", ".ps1":
		return path
	}
	return ""
}

func copyInstallerDirToTemp(installerDir, installerPath, name string) (string, string, error) {
	tempDir, err := os.MkdirTemp("", "launcher-installer-*")
	if err != nil {
		return "", "", err
	}

	if err := copyDir(installerDir, tempDir); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", "", err
	}

	rel, err := filepath.Rel(installerDir, installerPath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		_ = os.RemoveAll(tempDir)
		return "", "", fmt.Errorf("instalator %s nie znajduje się w folderze %s", installerPath, installerDir)
	}

	tempInstallerPath := filepath.Join(tempDir, rel)
	slog.Info("installer copied to temp", "name", name, "source", installerPath, "temp", tempInstallerPath)
	return tempInstallerPath, tempDir, nil
}

func splitCommandLine(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var args []string
	var b strings.Builder
	inQuote := false
	for _, r := range s {
		if r == '"' {
			inQuote = !inQuote
			continue
		}
		if (r == ' ' || r == '\t') && !inQuote {
			if b.Len() > 0 {
				args = append(args, b.String())
				b.Reset()
			}
			continue
		}
		b.WriteRune(r)
	}
	if b.Len() > 0 {
		args = append(args, b.String())
	}
	return args
}

func installerDoneMessage(installerPath string, ranFromTemp bool) string {
	msg := "Instalator zakończony: " + filepath.Base(installerPath)
	if ranFromTemp {
		msg += " (uruchomiony z TEMP, kopia usunięta)"
	}
	return msg
}

func findInstaller(installerDir string) string {
	type candidate struct {
		path  string
		score int
	}
	var candidates []candidate

	filepath.WalkDir(installerDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if strings.Contains(name, "uninstall") || strings.Contains(name, "unins") {
			return nil
		}

		score := 0
		switch {
		case strings.HasSuffix(name, ".msi"):
			score = 90
		case strings.HasSuffix(name, ".ps1"):
			score = 85
		case strings.HasSuffix(name, ".cmd"), strings.HasSuffix(name, ".bat"):
			score = 80
		case strings.HasSuffix(name, ".exe"):
			score = 70
		default:
			return nil
		}
		if strings.Contains(name, "setup") || strings.Contains(name, "install") || strings.Contains(name, "installer") || strings.Contains(name, "ninite") {
			score += 20
		}
		candidates = append(candidates, candidate{path: path, score: score})
		return nil
	})

	if len(candidates) == 0 {
		return ""
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].path < candidates[j].path
		}
		return candidates[i].score > candidates[j].score
	})
	return candidates[0].path
}

func runScript(scriptDir, name, scriptFile, scriptArgs string) InstallResult {
	if _, err := os.Stat(scriptDir); os.IsNotExist(err) {
		return InstallResult{Name: name, Success: false, Message: "Folder nie znaleziony: " + scriptDir, Type: "script"}
	}

	scriptPath := resolveScriptPath(scriptDir, scriptFile)
	if strings.TrimSpace(scriptFile) != "" && scriptPath == "" {
		return InstallResult{Name: name, Success: false, Message: "Wskazany skrypt nie istnieje albo ma nieobsługiwany typ (.ps1/.bat/.cmd/.reg): " + scriptFile, Type: "script"}
	}
	if scriptPath == "" {
		scriptPath = findScript(scriptDir)
	}
	if scriptPath == "" {
		return InstallResult{Name: name, Success: false, Message: "Nie znaleziono skryptu w " + scriptDir, Type: "script"}
	}

	lower := strings.ToLower(scriptPath)
	var cmd *exec.Cmd
	args := splitCommandLine(scriptArgs)
	switch {
	case strings.HasSuffix(lower, ".ps1"):
		cmd = powershellFileCommand(scriptPath, args)
	case strings.HasSuffix(lower, ".bat"), strings.HasSuffix(lower, ".cmd"):
		cmd = batchCommand(scriptPath, args)
	case strings.HasSuffix(lower, ".reg"):
		cmd = exec.Command("reg", "import", scriptPath)
	default:
		return InstallResult{Name: name, Success: false, Message: "Nieobsługiwany skrypt: " + filepath.Base(scriptPath), Type: "script"}
	}

	cmd.Dir = filepath.Dir(scriptPath)
	if err := cmd.Start(); err != nil {
		return InstallResult{Name: name, Success: false, Message: "Błąd uruchamiania skryptu: " + err.Error(), Type: "script"}
	}
	if err := cmd.Wait(); err != nil {
		return InstallResult{Name: name, Success: false, Message: commandFailureMessage("Skrypt zakończył się błędem", err), Type: "script"}
	}
	return InstallResult{Name: name, Success: true, Message: "Skrypt wykonany: " + filepath.Base(scriptPath), Type: "script"}
}

func installerCommand(installerPath string, args []string) (*exec.Cmd, error) {
	lower := strings.ToLower(installerPath)
	switch {
	case strings.HasSuffix(lower, ".msi"):
		return exec.Command("msiexec", append([]string{"/i", installerPath}, args...)...), nil
	case strings.HasSuffix(lower, ".msixbundle"), strings.HasSuffix(lower, ".appx"), strings.HasSuffix(lower, ".appxbundle"):
		escapedPath := strings.ReplaceAll(installerPath, "'", "''")
		return exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass",
			"-Command", fmt.Sprintf("Add-AppxPackage -LiteralPath '%s'", escapedPath)), nil
	case strings.HasSuffix(lower, ".ps1"):
		return powershellFileCommand(installerPath, args), nil
	case strings.HasSuffix(lower, ".bat"), strings.HasSuffix(lower, ".cmd"):
		return batchCommand(installerPath, args), nil
	case strings.HasSuffix(lower, ".exe"):
		return exec.Command(installerPath, args...), nil
	default:
		return nil, fmt.Errorf("Nieobsługiwany typ instalatora: %s", filepath.Ext(installerPath))
	}
}

// CALL zabezpiecza ścieżki BAT/CMD zawierające spacje i znaki specjalne,
// np. katalog "Windows & Office", przed ponowną interpretacją przez cmd.exe.
func batchCommand(scriptPath string, args []string) *exec.Cmd {
	comspec := strings.TrimSpace(os.Getenv("ComSpec"))
	if comspec == "" {
		comspec = "cmd.exe"
	}
	cmdArgs := []string{"/d", "/c", "call", scriptPath}
	cmdArgs = append(cmdArgs, args...)
	return exec.Command(comspec, cmdArgs...)
}

func powershellFileCommand(scriptPath string, args []string) *exec.Cmd {
	cmdArgs := []string{"-NoLogo", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath}
	cmdArgs = append(cmdArgs, args...)
	return exec.Command("powershell.exe", cmdArgs...)
}

func commandFailureMessage(prefix string, err error) string {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Sprintf("%s (%s).", prefix, formatExitCode(exitErr.ExitCode()))
	}
	return prefix + ": " + err.Error()
}

func resolveScriptPath(scriptDir, scriptFile string) string {
	scriptFile = strings.TrimSpace(scriptFile)
	if scriptFile == "" {
		return ""
	}
	path := scriptFile
	if !filepath.IsAbs(path) {
		path = filepath.Join(scriptDir, path)
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return ""
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ps1", ".bat", ".cmd", ".reg":
		return path
	default:
		return ""
	}
}

func findScript(scriptDir string) string {
	priority := map[string]int{".ps1": 40, ".bat": 30, ".cmd": 20, ".reg": 10}
	type candidate struct {
		path  string
		score int
	}
	var candidates []candidate
	filepath.WalkDir(scriptDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		score, ok := priority[ext]
		if !ok {
			return nil
		}
		name := strings.ToLower(d.Name())
		if strings.Contains(name, "install") || strings.Contains(name, "setup") || strings.Contains(name, "run") || strings.Contains(name, "start") {
			score += 20
		}
		candidates = append(candidates, candidate{path: path, score: score})
		return nil
	})
	if len(candidates) == 0 {
		return ""
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].path < candidates[j].path
		}
		return candidates[i].score > candidates[j].score
	})
	return candidates[0].path
}

// copyDirFast uses robocopy (native Windows) for dramatically faster directory
// copying with multi-threading. Falls back to Go copyDir if robocopy is unavailable.
func copyDirFast(src, dst string) error {
	cmd := exec.Command("robocopy", src, dst, "/E", "/MT:16", "/R:1", "/W:1", "/NFL", "/NDL", "/NJH", "/NJS", "/NP", "/NS", "/NC")
	hidden(cmd)
	_ = cmd.Run()

	if cmd.ProcessState != nil {
		code := cmd.ProcessState.ExitCode()
		// robocopy: 0-7 = success, 8+ = error
		if code < 8 {
			return nil
		}
		slog.Warn("robocopy failed, falling back to Go copy", "exit", code)
	}
	// Fallback
	return copyDir(src, dst)
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
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// batchCreateShortcuts creates desktop shortcuts for multiple programs in a single
// PowerShell call. Each entry in dirs is "destDir\x00name".
func batchCreateShortcuts(dirs []string) {
	if len(dirs) == 0 {
		return
	}

	home, _ := os.UserHomeDir()
	desktop := filepath.Join(home, "Desktop")
	if _, err := os.Stat(desktop); err != nil {
		return
	}

	var sb strings.Builder
	sb.WriteString("$ws = New-Object -ComObject WScript.Shell\n")
	any := false

	for _, entry := range dirs {
		parts := strings.SplitN(entry, "\x00", 2)
		if len(parts) != 2 {
			continue
		}
		dir, name := parts[0], parts[1]

		var exePath string
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(strings.ToLower(d.Name()), ".exe") {
				exePath = path
				return filepath.SkipAll
			}
			return nil
		})
		if exePath == "" {
			continue
		}

		shortcutPath := filepath.Join(desktop, name+".lnk")
		sb.WriteString(fmt.Sprintf("$s = $ws.CreateShortcut('%s'); $s.TargetPath = '%s'; $s.WorkingDirectory = '%s'; $s.Save()\n",
			shortcutPath, exePath, filepath.Dir(exePath)))
		any = true
	}

	if !any {
		return
	}

	slog.Info("batch creating shortcuts", "count", len(dirs))
	hidden(exec.Command("powershell", "-NoProfile", "-Command", sb.String())).Run()
}

func createShortcut(programDir, name string) {
	home, _ := os.UserHomeDir()
	desktop := filepath.Join(home, "Desktop")
	if _, err := os.Stat(desktop); err != nil {
		return
	}
	shortcutPath := filepath.Join(desktop, name+".lnk")

	// Find main .exe
	var exePath string
	filepath.WalkDir(programDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".exe") {
			exePath = path
			return filepath.SkipAll
		}
		return nil
	})

	if exePath == "" {
		return
	}

	// Create shortcut using PowerShell
	ps := fmt.Sprintf(`$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut('%s'); $s.TargetPath = '%s'; $s.WorkingDirectory = '%s'; $s.Save()`,
		shortcutPath, exePath, filepath.Dir(exePath))
	hidden(exec.Command("powershell", "-NoProfile", "-Command", ps)).Run()
}
