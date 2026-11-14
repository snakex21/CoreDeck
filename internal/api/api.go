package api

// Shared core: the API struct bound to the WebView, path helpers, system
// detection and the JSON error envelope used by every other api_*.go file.

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"launcher/internal/audit"
	"launcher/internal/drivers"
	"launcher/internal/knowledge"
	"launcher/internal/llm"
	"launcher/internal/models"
	"launcher/internal/programs"
	"launcher/internal/settings"
	"launcher/internal/system"
	"launcher/internal/tweaks"
)

type API struct {
	ExeDir        string
	WebDir        string
	WebFS         fs.FS
	ProgramsData  *programs.ProgramsData
	TweaksData    *tweaks.TweaksData
	DriversData   *drivers.DriversData
	SettingsData  *settings.SettingsData
	KnowledgeData *knowledge.KnowledgeData
	ProfilesStore *programs.ProfilesStore
	LLMClient     *llm.LLMClient
	AuditLog      *audit.Logger
	knowledgeMu   sync.Mutex
}

// ResolvePath maps a path stored in the catalogue (often relative like
// "../Centrum...") to the absolute Windows path that the launcher would
// actually use. The first candidate that exists on disk is returned; if
// none exist, the most likely candidate (<drive>\<pathPart>) is returned
// anyway so the caller can see what would be tried.
func (a *API) ResolvePath(rawPath string) string {
	pathPart := strings.ReplaceAll(rawPath, "/", "\\")
	pathPart = strings.TrimSpace(pathPart)
	for strings.HasPrefix(pathPart, "..\\") {
		pathPart = pathPart[3:]
	}
	drive := system.DriveOf(a.ExeDir)
	candidates := []string{
		filepath.Join(drive+"\\", pathPart),
		absJoin(a.ExeDir, "data", rawPath),
		absJoin(a.ExeDir, "..", pathPart),
		absJoin(a.ExeDir, pathPart),
		absJoin(a.WebDir, rawPath),
	}
	for _, cand := range candidates {
		normalized := filepath.Clean(cand)
		if _, err := os.Stat(normalized); err == nil {
			return normalized
		}
	}
	return filepath.Clean(candidates[0])
}

func (a *API) OpenFolder(rawPath string) error {
	resolved := a.ResolvePath(rawPath)
	slog.Info("open_folder", "raw", rawPath, "resolved", resolved)
	return exec.Command("explorer", resolved).Start()
}

// OpenURL opens a URL in the system's default browser.
func (a *API) OpenURL(url string) error {
	slog.Info("open_url", "url", url)
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

func (a *API) GetAvailableModels() []models.ModelInfo {
	llmDir := filepath.Join(a.ExeDir, "data", "tools", "llama-cpp")
	_, serverErr := os.Stat(filepath.Join(llmDir, "llama-server.exe"))
	_, modelErr := os.Stat(filepath.Join(llmDir, "LFM2.5-350M-Q4_K_M.gguf"))
	available := serverErr == nil && modelErr == nil
	return []models.ModelInfo{
		{ID: "llm", Name: "LLM (LFM2.5-350M)", Size: "219MB", Available: available},
	}
}

func absJoin(base string, parts ...string) string {
	return filepath.Clean(filepath.Join(append([]string{base}, parts...)...))
}

// --- System Detection ---

func (a *API) GetSystemInfo() string {
	info := system.Detect()
	data, _ := json.Marshal(info)
	return string(data)
}

// errResp is the JSON envelope returned to the frontend when an API call
// fails. Marshalling it via json.Marshal guarantees correct escaping of
// quotes, backslashes and control characters inside err.Error() — something
// the previous string concatenation (`{"error": "`+err+`"}`) got wrong.
type errResp struct {
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
}

// jsonError builds a safe `{"error": "..."}` string from any message.
func jsonError(msg string) string {
	b, _ := json.Marshal(errResp{Error: msg})
	return string(b)
}

// GetPortablePaths returns the default import/export folders derived from the
// executable's current drive (see system.PortablePathsFor).
func (a *API) GetPortablePaths() string {
	data, _ := json.Marshal(system.PortablePathsFor(a.ExeDir))
	return string(data)
}
