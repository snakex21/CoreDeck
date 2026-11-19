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
	"sync"

	"launcher/internal/audit"
	"launcher/internal/drivers"
	"launcher/internal/knowledge"
	"launcher/internal/llm"
	"launcher/internal/models"
	"launcher/internal/paths"
	"launcher/internal/programs"
	"launcher/internal/settings"
	"launcher/internal/system"
	"launcher/internal/tweaks"
	"launcher/internal/uimsg"
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

// ResolvePath maps a catalogue path to the absolute path the launcher would
// use (see paths.ResolvePath). Bound to the UI as resolve_path.
func (a *API) ResolvePath(rawPath string) string {
	return paths.ResolvePath(a.ExeDir, rawPath)
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
	// Code is the UI message key (translations.js) of Error/Message, Params
	// its placeholders; the text stays as the Polish fallback (internal/uimsg).
	Code   string            `json:"code,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

// jsonError builds a safe `{"error": "..."}` string from a technical message
// (shown as it is). User-facing messages go through jsonMsg / jsonErr.
func jsonError(msg string) string {
	b, _ := json.Marshal(errResp{Error: msg})
	return string(b)
}

// jsonMsg builds `{"error": text, "code": ..., "params": ...}`.
func jsonMsg(m uimsg.Msg) string {
	b, _ := json.Marshal(errResp{Error: m.Text, Code: m.Code, Params: m.Params})
	return string(b)
}

// jsonErr builds the error response of err, with its UI code when it has one.
func jsonErr(err error) string {
	return jsonMsg(uimsg.From(err))
}

// jsonMsgExtra is jsonMsg with additional top-level fields.
func jsonMsgExtra(m uimsg.Msg, extra map[string]any) string {
	out := map[string]any{"error": m.Text, "code": m.Code}
	if len(m.Params) > 0 {
		out["params"] = m.Params
	}
	for k, v := range extra {
		out[k] = v
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// jsonResult builds `{"success": ok, "message": text, "code": ...}`.
func jsonResult(ok bool, m uimsg.Msg) string {
	b, _ := json.Marshal(errResp{Success: ok, Message: m.Text, Code: m.Code, Params: m.Params})
	return string(b)
}

// GetPortablePaths returns the default import/export folders derived from the
// executable's current drive (see system.PortablePathsFor).
func (a *API) GetPortablePaths() string {
	data, _ := json.Marshal(system.PortablePathsFor(a.ExeDir))
	return string(data)
}
