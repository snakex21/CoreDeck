package api

// Self-updater bindings. The logic lives in internal/update; this file only
// adapts it to the JSON-string API used by the WebView.

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"

	launcherupdate "launcher/internal/update"
	"launcher/internal/version"
)

// updateRuntime is process-wide: there is one API and one window per process.
var updateRuntime struct {
	mu                  sync.Mutex
	manager             *launcherupdate.Manager
	launchedAfterUpdate bool
	quit                func()
}

// ConfigureUpdates must be called before the window opens. launchedAfterUpdate
// is true when the process was started with --updated by the update guard.
func (a *API) ConfigureUpdates(launchedAfterUpdate bool) {
	updateRuntime.mu.Lock()
	defer updateRuntime.mu.Unlock()
	updateRuntime.launchedAfterUpdate = launchedAfterUpdate
	updateRuntime.manager = nil
}

// SetQuitHook registers how the updater closes the app (the WebView window
// registers a Terminate call here).
func (a *API) SetQuitHook(fn func()) {
	updateRuntime.mu.Lock()
	defer updateRuntime.mu.Unlock()
	updateRuntime.quit = fn
}

func quitApp() {
	updateRuntime.mu.Lock()
	fn := updateRuntime.quit
	updateRuntime.mu.Unlock()
	slog.Info("update: closing the application for restart")
	if fn != nil {
		fn()
		// Fallback in case the window does not close.
		time.AfterFunc(10*time.Second, func() { os.Exit(0) })
		return
	}
	os.Exit(0)
}

func (a *API) updater() *launcherupdate.Manager {
	updateRuntime.mu.Lock()
	defer updateRuntime.mu.Unlock()
	if updateRuntime.manager == nil {
		m := launcherupdate.NewManager(a.ExeDir, version.Current, updateRuntime.launchedAfterUpdate)
		m.Quit = quitApp
		updateRuntime.manager = m
	}
	return updateRuntime.manager
}

func toJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return jsonError(err.Error())
	}
	return string(data)
}

type updateErrResp struct {
	Error     string `json:"error"`
	ErrorCode string `json:"errorCode,omitempty"`
}

type successResp struct {
	Success bool `json:"success"`
}

func (a *API) GetAppVersion() string {
	return version.Current
}

// GetUpdateSettings → {channel, autoCheck, skippedVersion, lastCheck}.
func (a *API) GetUpdateSettings() string {
	return toJSON(a.updater().Settings())
}

// SetUpdateSettings(json) → {success} or {error}.
func (a *API) SetUpdateSettings(raw string) string {
	if err := a.updater().SetSettings(raw); err != nil {
		return jsonError(err.Error())
	}
	return toJSON(successResp{Success: true})
}

// CheckForUpdates(mode "auto"|"manual") → update.CheckResult.
func (a *API) CheckForUpdates(mode string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return toJSON(a.updater().Check(ctx, mode))
}

// SkipUpdateVersion(version) → {success} or {error}.
func (a *API) SkipUpdateVersion(v string) string {
	if err := a.updater().Skip(v); err != nil {
		return jsonError(err.Error())
	}
	return toJSON(successResp{Success: true})
}

// StartUpdate(version) → {sessionId} or {error, errorCode}.
func (a *API) StartUpdate(v string) string {
	id, err := a.updater().StartUpdate(v)
	if err != nil {
		slog.Warn("update: cannot start", "version", v, "error", err)
		return toJSON(updateErrResp{Error: err.Error(), ErrorCode: launcherupdate.ErrorCode(err)})
	}
	slog.Info("update: started", "version", v, "session", id)
	return toJSON(struct {
		SessionID string `json:"sessionId"`
	}{id})
}

// UpdateProgress(sessionId) → {state, percent, error, errorCode}.
func (a *API) UpdateProgress(id string) string {
	p := a.updater().Progress(id)
	if p.State == launcherupdate.StateError {
		slog.Warn("update: failed", "session", id, "code", p.ErrorCode, "error", p.Error)
	}
	return toJSON(p)
}

// UpdateReady → {justUpdated, rolledBack:{from,to}|null}; writes the health
// marker in a run started with --updated.
func (a *API) UpdateReady() string {
	info := a.updater().Ready()
	if info.JustUpdated != "" || info.RolledBack != nil {
		slog.Info("update: ready", "justUpdated", info.JustUpdated, "rolledBack", info.RolledBack != nil)
	}
	return toJSON(info)
}

// GetRollbackInfo → {available, version}.
func (a *API) GetRollbackInfo() string {
	ok, v := a.updater().RollbackInfo()
	return toJSON(struct {
		Available bool   `json:"available"`
		Version   string `json:"version"`
	}{ok, v})
}

// RollbackUpdate → {success} or {error}. On success the app restarts.
func (a *API) RollbackUpdate() string {
	if err := a.updater().Rollback(); err != nil {
		slog.Warn("update: rollback failed", "error", err)
		return toJSON(updateErrResp{Error: err.Error(), ErrorCode: launcherupdate.ErrorCode(err)})
	}
	slog.Info("update: restoring previous version")
	return toJSON(successResp{Success: true})
}
