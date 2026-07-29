package main

import (
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"launcher/internal/api"
	"launcher/internal/llm"
	"launcher/internal/programs"
	"launcher/internal/server"
	"launcher/internal/system"
	"launcher/internal/version"
	"launcher/internal/webview"
)

func main() {
	if ensureElevated() {
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

	seedDataFiles(exeDir)

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

	dataDir := filepath.Join(exeDir, ".webview_data")
	os.MkdirAll(dataDir, 0755)

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
	}
	// Load the small catalogue synchronously, then warm the expensive state scan
	// shortly after startup. Manual refresh and post-apply scans still bypass it.
	_ = appAPI.GetTweaksData()
	go func() {
		time.Sleep(600 * time.Millisecond)
		appAPI.WarmTweakStateCache()
	}()

	url := webview.BuildURL(port, exeDir)

	cfg := webview.WindowConfig{
		Title:     "CoreDeck",
		URL:       url,
		Width:     1400,
		Height:    900,
		MinWidth:  800,
		MinHeight: 600,
		Debug:     false,
		Maximized: true,
		DataDir:   dataDir,
		API:       appAPI,
		IconPath:  iconPath,
	}

	if err := webview.NewWindow(cfg); err != nil {
		fatalExit("Nie można otworzyć okna programu. Sprawdź instalację Microsoft Edge WebView2.", err)
	}

	slog.Info("launcher finished")
}

// seedDataFiles writes the embedded data files to <exeDir>/data/<name>.json
// only when they don't already exist, so user edits are never overwritten.
func seedDataFiles(exeDir string) {
	dataDir := filepath.Join(exeDir, "data")
	os.MkdirAll(dataDir, 0755)
	for _, name := range []string{"programs.json", "tweaks.json", "drivers.json", "knowledge.json", "settings.json"} {
		dst := filepath.Join(dataDir, name)
		if _, err := os.Stat(dst); err == nil {
			continue // exists — keep user edits
		}
		data, err := staticFiles.ReadFile("web/data/" + name)
		if err != nil {
			continue
		}
		os.WriteFile(dst, data, 0644)
	}
}

func getExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
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
