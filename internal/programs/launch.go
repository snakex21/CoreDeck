package programs

import (
	"debug/pe"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"launcher/internal/paths"
	"launcher/internal/uimsg"
)

// LaunchResult is returned to the UI by the launch_program binding.
type LaunchResult struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	// Launched is the absolute path of the started executable.
	Launched string `json:"launched,omitempty"`
	// Candidates lists executables (relative to Folder, forward slashes) when
	// the folder has several plausible ones and the user has to pick.
	Candidates []string `json:"candidates,omitempty"`
	// OpenedFolder is set when there was nothing to run and the program
	// folder was opened in Explorer instead.
	OpenedFolder bool   `json:"openedFolder,omitempty"`
	Folder       string `json:"folder,omitempty"`
	// Code is the UI message key (translations.js), Params its placeholders;
	// Message is the Polish fallback and log text (see internal/uimsg).
	Code   string            `json:"code,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

// IsLaunchable reports whether an install type is started in place (from the
// pendrive) rather than installed: portable programs and tools.
func IsLaunchable(installType string) bool {
	switch strings.ToLower(strings.TrimSpace(installType)) {
	case "portable", "tool":
		return true
	}
	return false
}

// startExecutable starts a program detached from CoreDeck; replaced in tests.
var startExecutable = startDetached

// openFolder opens a folder in Explorer; replaced in tests.
var openFolder = openInExplorer

// LaunchProgram starts a portable program from its catalogue folder.
//
// exeDir is CoreDeck's directory (catalogue paths are resolved against it, so
// a changed drive letter is handled by paths.ResolvePath). file is an
// optional executable chosen by the user, relative to the program folder.
// Without it the configured installerFile is used, then the single obvious
// .exe in the folder. When several are plausible, Candidates is returned and
// nothing starts; when there is none, the folder is opened instead.
func LaunchProgram(exeDir string, p Program, variantIndex int, file string) LaunchResult {
	if variantIndex >= 0 {
		p.SelectedVariant = variantIndex
	}
	cfg := resolveVariant(p, exeDir)
	res := LaunchResult{Name: p.Name}
	if !IsLaunchable(cfg.InstallType) {
		res.setMsg(uimsg.New("launch_not_portable", "Ten wpis nie jest programem portable."))
		return res
	}
	folder := paths.ResolvePath(exeDir, cfg.FolderPath)
	res.Folder = folder
	if info, err := os.Stat(folder); err != nil || !info.IsDir() {
		res.setMsg(uimsg.New("launch_folder_missing", "Nie znaleziono folderu programu: "+folder, "path", folder))
		return res
	}

	var target string
	switch {
	case strings.TrimSpace(file) != "":
		path, err := executableInFolder(folder, file)
		if err != nil {
			res.setMsg(uimsg.From(err))
			return res
		}
		target = path
	case strings.TrimSpace(cfg.InstallerFile) != "":
		path, err := executableInFolder(folder, cfg.InstallerFile)
		if err != nil {
			res.setMsg(uimsg.New("launch_file_missing", "Plik wskazany w katalogu nie istnieje: "+cfg.InstallerFile, "file", cfg.InstallerFile))
			return res
		}
		target = path
	default:
		main, candidates := FindMainExecutable(folder, p.Name, filepath.Base(folder))
		if main == "" && len(candidates) > 0 {
			res.Success = true
			res.Candidates = candidates
			return res
		}
		if main == "" {
			if err := openFolder(folder); err != nil {
				res.setMsg(uimsg.New("launch_failed", "Nie znaleziono pliku .exe, a folderu nie udało się otworzyć: "+err.Error(), "error", err.Error()))
				return res
			}
			res.Success = true
			res.OpenedFolder = true
			res.setMsg(uimsg.New("launch_opened_folder", "W folderze nie ma programu do uruchomienia (np. tylko archiwum lub instalator). Otwarto folder."))
			return res
		}
		target = main
	}

	if err := startExecutable(target); err != nil {
		res.setMsg(launchErrorMessage(err))
		return res
	}
	slog.Info("program launched", "name", p.Name, "path", target)
	res.Success = true
	res.Launched = target
	return res
}

// executableInFolder validates a user/catalogue supplied file: it must stay
// inside folder, exist and be an .exe.
func executableInFolder(folder, file string) (string, error) {
	file = strings.TrimSpace(strings.Trim(file, `"`))
	path := filepath.FromSlash(strings.ReplaceAll(file, `\`, "/"))
	if !filepath.IsAbs(path) {
		path = filepath.Join(folder, path)
	}
	path = filepath.Clean(path)
	rel, err := filepath.Rel(folder, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", uimsg.Errorf("launch_outside_folder", "Plik "+file+" jest poza folderem programu.", "file", file)
	}
	if !strings.EqualFold(filepath.Ext(path), ".exe") {
		return "", uimsg.Errorf("launch_only_exe", "Można uruchomić tylko plik .exe: "+file, "file", file)
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", uimsg.Errorf("launch_file_missing", "Plik wskazany w katalogu nie istnieje: "+file, "file", file)
	}
	return path, nil
}

// setMsg sets the result's UI message.
func (r *LaunchResult) setMsg(m uimsg.Msg) {
	r.Message, r.Code, r.Params = m.Text, m.Code, m.Params
}

func launchErrorMessage(err error) uimsg.Msg {
	if exit, ok := installerExitFromError(err); ok && exit.Outcome == OutcomeCancelled {
		return uimsg.New("launch_cancelled", "Uruchamianie anulowane.")
	}
	m := uimsg.New("launch_failed", "Nie udało się uruchomić programu: "+err.Error(), "error", err.Error())
	if exit, ok := installerExitFromError(err); ok {
		m.Text = strings.Replace(exit.Msg.Text, "Instalator", "Program", 1)
	}
	return m
}

// maxExeDepth limits how deep FindMainExecutable looks (folder\a\b\c\x.exe).
const maxExeDepth = 3

// Names that are never the program itself.
var helperExeMarkers = []string{
	"unins", "uninstall", "setup", "install", "update", "crashreport", "crashpad",
	"crash_handler", "plugin-container", "qtwebengineprocess", "redist", "vcredist",
	"dotnetfx", "keygen", "helper", "elevate", "svc", "service", "notification",
}

type exeCandidate struct {
	rel   string
	score int
}

// FindMainExecutable picks the program's own .exe in a portable folder.
//
// It returns main when one executable is clearly the program (the only one,
// or one that scores well above the rest: shallow, named like the program,
// 64-bit). Otherwise main is empty and candidates (relative paths, forward
// slashes, best first, at most 12) let the user choose.
func FindMainExecutable(folder string, hints ...string) (main string, candidates []string) {
	tokens := nameTokens(hints...)
	var found []exeCandidate
	_ = filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, relErr := filepath.Rel(folder, path)
		if relErr != nil {
			return nil
		}
		depth := strings.Count(rel, string(filepath.Separator))
		if d.IsDir() {
			if path != folder && depth >= maxExeDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(d.Name()), ".exe") {
			return nil
		}
		base := strings.ToLower(strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())))
		score := scoreExecutable(base, depth, tokens)
		if isHelperExecutable(base) {
			// "Display Driver Uninstaller.exe" is DDU itself, not its
			// uninstaller: keep a helper-looking name only when it spells
			// out the program's name.
			if matchingTokens(compactName(base), tokens) < 2 {
				return nil
			}
			score -= 30
		}
		found = append(found, exeCandidate{rel: filepath.ToSlash(rel), score: score})
		return nil
	})
	if len(found) == 0 {
		return "", nil
	}
	sort.SliceStable(found, func(i, j int) bool {
		if found[i].score != found[j].score {
			return found[i].score > found[j].score
		}
		return strings.ToLower(found[i].rel) < strings.ToLower(found[j].rel)
	})
	if len(found) == 1 || found[0].score-found[1].score >= clearWinnerMargin {
		return filepath.Join(folder, filepath.FromSlash(found[0].rel)), nil
	}
	for i, c := range found {
		if i == 12 {
			break
		}
		candidates = append(candidates, c.rel)
	}
	return "", candidates
}

const clearWinnerMargin = 25

func isHelperExecutable(base string) bool {
	for _, marker := range helperExeMarkers {
		if strings.Contains(base, marker) {
			return true
		}
	}
	return false
}

func matchingTokens(compact string, tokens []string) int {
	n := 0
	for _, token := range tokens {
		if len(token) >= 4 && strings.Contains(compact, token) {
			n++
		}
	}
	return n
}

func scoreExecutable(base string, depth int, tokens []string) int {
	score := 100 - 20*depth
	compact := compactName(base)
	stem := strings.TrimRightFunc(strings.TrimSuffix(strings.TrimSuffix(compact, "x64"), "64"), unicode.IsDigit)
	for _, token := range tokens {
		if strings.Contains(compact, token) || (len(stem) >= 3 && strings.Contains(token, stem)) {
			score += 30
			break
		}
	}
	switch {
	case strings.Contains(base, "arm") || strings.HasSuffix(compact, "a64") || strings.HasSuffix(compact, "a32"):
		score -= 40
	case strings.Contains(base, "64"):
		score += 5
	case strings.Contains(base, "32") || strings.Contains(base, "x86"):
		score -= 25
	}
	if strings.Contains(base, "console") || strings.HasSuffix(compact, "cli") || strings.HasSuffix(compact, "cui") || strings.HasSuffix(compact, "cmd") {
		score -= 30
	}
	return score
}

// nameTokens splits program/folder names into lowercase alphanumeric words
// of at least three characters, plus each name squeezed into one word.
func nameTokens(names ...string) []string {
	var tokens []string
	seen := map[string]bool{}
	add := func(t string) {
		if len(t) >= 3 && !seen[t] {
			seen[t] = true
			tokens = append(tokens, t)
		}
	}
	for _, name := range names {
		name = strings.ToLower(name)
		if i := strings.IndexAny(name, "(["); i > 0 {
			add(compactName(name[:i]))
		}
		add(compactName(name))
		for _, word := range strings.FieldsFunc(name, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
			add(word)
		}
	}
	return tokens
}

func compactName(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// isConsoleExecutable reports whether an .exe uses the console subsystem
// (command-line tools such as nvflash), which needs a console window that
// stays open after the tool prints its output.
func isConsoleExecutable(path string) bool {
	f, err := pe.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	const subsystemConsole = 3
	switch h := f.OptionalHeader.(type) {
	case *pe.OptionalHeader64:
		return h.Subsystem == subsystemConsole
	case *pe.OptionalHeader32:
		return h.Subsystem == subsystemConsole
	}
	return false
}

var errNotSupported = errors.New("launching programs is supported only on Windows")
