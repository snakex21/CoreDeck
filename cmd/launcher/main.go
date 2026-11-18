package main

import (
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"launcher/internal/api"
	"launcher/internal/audit"
	"launcher/internal/llm"
	"launcher/internal/paths"
	"launcher/internal/programs"
	"launcher/internal/server"
	"launcher/internal/system"
	"launcher/internal/update"
	"launcher/internal/version"
	"launcher/internal/webview"
	"launcher/internal/windowstate"
)

func main() {
	// The update guard runs from the OLD binary (CoreDeck.exe.old) and must not
	// elevate, open a window or start the server.
	if statePath, ok := updateGuardArg(os.Args[1:]); ok {
		os.Exit(runUpdateGuard(getExeDir(), statePath))
	}
	// A run started by the guard keeps the guard's token: relaunching through
	// UAC would make the guard see an early exit and roll the update back.
	launchedAfterUpdate := hasArg(os.Args[1:], update.FlagUpdated)
	if !launchedAfterUpdate && ensureElevated() {
		return
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	exeDir := getExeDir()
	setupLogging(exeDir)
	slog.Info("CoreDeck", "version", version.Current)
	slog.Info("exeDir", "path", exeDir)

	// Hardware data is static for the lifetime of the app. Warm it in the
	// background while the main page starts, so the configurator opens instantly.
	go system.Detect()

	update.CleanupStale(exeDir, time.Now())
	seedDataFiles(exeDir)
	migrateDriveLetters(exeDir)

	// Extract the embedded icon into the runtime state directory.
	iconPath := extractIcon(exeDir)

	webFS, err := fs.Sub(staticFiles, "web")
	if err != nil {
		fatalExit("Nie można wczytać plików interfejsu.", err)
	}

	userPagesDir := filepath.Join(exeDir, "data", "user_pages")
	os.MkdirAll(userPagesDir, 0755)

	srv, port, err := server.Start(webFS, userPagesDir)
	if err != nil {
		fatalExit("Nie można uruchomić lokalnego serwera programu.", err)
	}
	defer srv.Close()

	llmClient := llm.NewLLMClient(exeDir)
	defer llmClient.Stop()

	dataDir := filepath.Join(exeDir, "data", "webview")
	migrateLegacyWebviewData(exeDir, dataDir)
	os.MkdirAll(dataDir, 0755)
	// WebView2 profile (theme, language, local storage) next to the exe
	// instead of %APPDATA%; must be set before the window is created.
	setupWebviewProfile(dataDir, legacyWebviewRoot())

	profilesStore, err := programs.NewProfilesStore(dataDir)
	if err != nil {
		slog.Warn("profiles store init failed", "error", err)
	}

	appAPI := &api.API{
		ExeDir:        exeDir,
		WebDir:        exeDir,
		WebFS:         staticFiles,
		ProfilesStore: profilesStore,
		LLMClient:     llmClient,
		AuditLog:      audit.New(filepath.Join(exeDir, "data", "logs", "audit.jsonl")),
	}
	appAPI.ConfigureUpdates(launchedAfterUpdate)
	// Load the small catalogue synchronously, then warm the expensive state scan
	// shortly after startup. Manual refresh and post-apply scans still bypass it.
	_ = appAPI.GetTweaksData()
	go func() {
		time.Sleep(600 * time.Millisecond)
		appAPI.WarmTweakStateCache()
	}()

	url := webview.BuildURL(port, exeDir)
	windowStatePath := filepath.Join(exeDir, "data", "state", "window.json")
	windowState := windowstate.Load(windowStatePath, windowstate.State{
		Width:     1400,
		Height:    900,
		Maximized: true,
	})

	cfg := buildWindowConfig(url, dataDir, iconPath, windowStatePath, windowState, appAPI)

	if err := webview.NewWindow(cfg); err != nil {
		fatalExit("Nie można otworzyć okna programu. Sprawdź instalację Microsoft Edge WebView2.", err)
	}

	slog.Info("launcher finished")
}

func buildWindowConfig(
	url string,
	dataDir string,
	iconPath string,
	statePath string,
	state windowstate.State,
	appAPI *api.API,
) webview.WindowConfig {
	return webview.WindowConfig{
		Title:     "CoreDeck",
		URL:       url,
		Width:     state.Width,
		Height:    state.Height,
		MinWidth:  800,
		MinHeight: 600,
		Debug:     envEnabled("COREDECK_WEBVIEW_DEBUG") || envEnabled("LAUNCHER_DEBUG"),
		Maximized: state.Maximized,
		DataDir:   dataDir,
		API:       appAPI,
		IconPath:  iconPath,
		StatePath: statePath,
	}
}

func hasArg(args []string, flag string) bool {
	for _, a := range args {
		if strings.EqualFold(a, flag) {
			return true
		}
	}
	return false
}

// updateGuardArg returns the state file passed as --update-guard <statefile>.
func updateGuardArg(args []string) (string, bool) {
	for i, a := range args {
		if strings.EqualFold(a, update.FlagGuard) {
			if i+1 < len(args) {
				return args[i+1], true
			}
			return "", true
		}
	}
	return "", false
}

// runUpdateGuard finishes a pending self-update (see internal/update/guard.go)
// and returns the process exit code.
func runUpdateGuard(exeDir, statePath string) int {
	env := update.RealGuardEnv(exeDir)
	if err := update.ValidateGuardStatePath(exeDir, statePath); err != nil {
		env.Logf("guard: %v", err)
		return 2
	}
	env.OpenParentEarly()
	outcome, err := update.RunGuard(env)
	env.Logf("guard: finished (%s)", outcome)
	if err != nil {
		env.Logf("guard: %v", err)
		return 1
	}
	return 0
}

func envEnabled(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// seedDataFiles writes the embedded data files to <exeDir>/data/<name>.json
// only when they don't already exist, so user edits are never overwritten.
func seedDataFiles(exeDir string) {
	dataDir := filepath.Join(exeDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		slog.Warn("seed: cannot create data dir", "dir", dataDir, "error", err)
		return
	}
	for _, name := range []string{"programs.json", "tweaks.json", "drivers.json", "knowledge.json", "settings.json"} {
		dst := filepath.Join(dataDir, name)
		if _, err := os.Stat(dst); err == nil {
			continue // exists — keep user edits
		}
		data, err := staticFiles.ReadFile("web/data/" + name)
		if err != nil {
			slog.Warn("seed: embedded default missing", "file", name, "error", err)
			continue
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			slog.Warn("seed: cannot write default", "file", dst, "error", err)
		}
	}
}

// migrateDriveLetters repairs catalogue paths saved as absolute paths on the
// pendrive's previous letter (e.g. J:\... after it became L:). Tweaks are
// not touched: their commands are approved by hash.
func migrateDriveLetters(exeDir string) {
	for _, name := range []string{"programs.json", "drivers.json", "settings.json"} {
		file := filepath.Join(exeDir, "data", name)
		rewrites, err := paths.MigrateDriveLetters(exeDir, file)
		if err != nil {
			slog.Warn("drive letter migration skipped", "file", file, "error", err)
			continue
		}
		for _, rw := range rewrites {
			slog.Info("drive letter migration", "file", name, "old", rw.Old, "new", rw.New)
		}
	}
}

func getExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func migrateLegacyWebviewData(exeDir, dataDir string) {
	legacyDir := filepath.Join(exeDir, ".webview_data")
	if _, err := os.Stat(dataDir); err == nil {
		return
	}
	info, err := os.Stat(legacyDir)
	if err != nil || !info.IsDir() {
		return
	}
	if err := os.MkdirAll(filepath.Dir(dataDir), 0755); err != nil {
		slog.Warn("legacy WebView data migration failed", "error", err)
		return
	}
	if err := os.Rename(legacyDir, dataDir); err != nil {
		slog.Warn("legacy WebView data migration failed", "error", err)
	}
}

func setupLogging(exeDir string) {
	logDir := filepath.Join(exeDir, "data", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}
	logPath := filepath.Join(logDir, "launcher.log")
	rotateLog(logPath, 2*1024*1024)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	level := slog.LevelInfo
	var output io.Writer = f
	if os.Getenv("LAUNCHER_DEBUG") != "" {
		level = slog.LevelDebug
		output = io.MultiWriter(f, os.Stderr)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{Level: level})))
	slog.Info("logging enabled", "path", logPath, "level", level.String())
}

func rotateLog(logPath string, maxBytes int64) {
	info, err := os.Stat(logPath)
	if err != nil || info.Size() < maxBytes {
		return
	}
	oldest := logPath + ".2"
	previous := logPath + ".1"
	_ = os.Remove(oldest)
	_ = os.Rename(previous, oldest)
	_ = os.Rename(logPath, previous)
}

func fatalExit(message string, err error) {
	slog.Error(message, "error", err)
	showFatalError(message + "\n\n" + err.Error() + "\n\nSzczegóły zapisano w data\\logs\\launcher.log")
	os.Exit(1)
}

// extractIcon writes the embedded favicon.ico to data/state/favicon.ico so it
// can be loaded by Win32 API without cluttering the application root.
func extractIcon(exeDir string) string {
	stateDir := filepath.Join(exeDir, "data", "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		slog.Warn("failed to create runtime state directory", "error", err)
		return ""
	}
	iconPath := filepath.Join(stateDir, "favicon.ico")
	data, err := staticFiles.ReadFile("web/favicon.ico")
	if err != nil {
		slog.Warn("favicon.ico not found in embed", "error", err)
		return ""
	}
	// Only write if different (avoid unnecessary disk write)
	if existing, err := os.ReadFile(iconPath); err == nil && len(existing) == len(data) {
		return iconPath
	}
	if err := os.WriteFile(iconPath, data, 0644); err != nil {
		slog.Warn("failed to extract favicon.ico", "error", err)
		return ""
	}
	slog.Info("extracted favicon.ico for window icon", "path", iconPath)
	return iconPath
}
