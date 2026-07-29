package api

// Tweak catalogue, async tweak sessions, cancellation and revert.

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"launcher/internal/tweaks"
)

func (a *API) configureTweakStateDir() {
	stateDir := filepath.Join(a.ExeDir, "data", "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		slog.Warn("create tweak state directory", "path", stateDir, "error", err)
	}

	legacyPath := filepath.Join(a.ExeDir, ".launcher_tweaks_backup.json")
	currentPath := filepath.Join(stateDir, ".launcher_tweaks_backup.json")
	if _, err := os.Stat(currentPath); os.IsNotExist(err) {
		if _, err := os.Stat(legacyPath); err == nil {
			if err := os.Rename(legacyPath, currentPath); err != nil {
				slog.Warn("migrate tweak backup", "from", legacyPath, "to", currentPath, "error", err)
			} else {
				slog.Info("migrated tweak backup", "path", currentPath)
			}
		}
	}

	tweaks.SetStateDir(stateDir)
}

func (a *API) GetTweaksData() string {
	if a.TweaksData == nil {
		if td, err := tweaks.LoadTweaks(a.WebDir); err == nil { // external <exeDir>/data/tweaks.json
			a.TweaksData = td
		} else if td, err := tweaks.LoadFromFS(a.WebFS); err == nil { // embedded fallback
			a.TweaksData = td
		} else {
			slog.Error("load tweaks", "error", err)
			return `{"error": "Cannot load tweaks.json"}`
		}
	}
	data, _ := json.Marshal(a.TweaksData)
	return string(data)
}

func (a *API) GetTweaksForProfile(profileName string) string {
	if a.TweaksData == nil {
		a.GetTweaksData()
	}
	if a.TweaksData == nil {
		return `[]`
	}
	result := tweaks.GetTweaksForProfile(a.TweaksData, profileName)
	data, _ := json.Marshal(result)
	return string(data)
}

func (a *API) GetTweakStates() string {
	a.configureTweakStateDir()
	if a.TweaksData == nil {
		a.GetTweaksData()
	}
	if a.TweaksData == nil {
		return `{"error": "No tweaks data"}`
	}

	states := tweaks.CheckTweaksStates(a.TweaksData.Tweaks)
	data, _ := json.Marshal(states)
	return string(data)
}

type tweakStateScanSession struct {
	mu        sync.Mutex
	completed int
	total     int
	done      bool
	states    []tweaks.TweakState
	err       string
}

var tweakStateScanSessions sync.Map
var tweakStateScanCounter int64
var tweakStateScanCache struct {
	sync.Mutex
	states          []tweaks.TweakState
	updatedAt       time.Time
	activeSessionID string
	generation      uint64
}

const tweakStateScanCacheTTL = 10 * time.Minute

// StartTweakStateScan starts the expensive system-state scan outside the
// WebView binding callback. This keeps the UI thread responsive while registry,
// AppX, service and optional-feature checks run in the background.
func (a *API) StartTweakStateScan(force bool) string {
	a.configureTweakStateDir()
	if a.TweaksData == nil {
		a.GetTweaksData()
	}
	if a.TweaksData == nil {
		return `{"error": "No tweaks data"}`
	}

	items := append([]tweaks.Tweak(nil), a.TweaksData.Tweaks...)

	tweakStateScanCache.Lock()
	if !force && len(tweakStateScanCache.states) > 0 &&
		time.Since(tweakStateScanCache.updatedAt) < tweakStateScanCacheTTL {
		states := append([]tweaks.TweakState(nil), tweakStateScanCache.states...)
		tweakStateScanCache.Unlock()
		return completedTweakStateScanSession(states)
	}
	if tweakStateScanCache.activeSessionID != "" {
		sessionID := tweakStateScanCache.activeSessionID
		tweakStateScanCache.Unlock()
		return tweakStateScanSessionResponse(sessionID)
	}

	sessionID := fmt.Sprintf("tweak-scan-%d", atomic.AddInt64(&tweakStateScanCounter, 1))
	session := &tweakStateScanSession{total: len(items)}
	tweakStateScanSessions.Store(sessionID, session)
	scanGeneration := tweakStateScanCache.generation
	tweakStateScanCache.activeSessionID = sessionID
	tweakStateScanCache.Unlock()

	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				session.mu.Lock()
				session.err = fmt.Sprintf("tweak scan panic: %v", recovered)
				session.done = true
				session.mu.Unlock()
			}
			tweakStateScanCache.Lock()
			if tweakStateScanCache.activeSessionID == sessionID {
				tweakStateScanCache.activeSessionID = ""
			}
			tweakStateScanCache.Unlock()
			time.AfterFunc(30*time.Second, func() { tweakStateScanSessions.Delete(sessionID) })
		}()

		states := tweaks.CheckTweaksStatesWithProgress(items, func(completed, total int) {
			session.mu.Lock()
			session.completed = completed
			session.total = total
			session.mu.Unlock()
		})

		session.mu.Lock()
		session.states = states
		session.completed = len(items)
		session.done = true
		session.mu.Unlock()

		tweakStateScanCache.Lock()
		if tweakStateScanCache.generation == scanGeneration {
			tweakStateScanCache.states = append([]tweaks.TweakState(nil), states...)
			tweakStateScanCache.updatedAt = time.Now()
		}
		tweakStateScanCache.Unlock()
	}()

	return tweakStateScanSessionResponse(sessionID)
}

func (a *API) WarmTweakStateCache() {
	_ = a.StartTweakStateScan(false)
}

func completedTweakStateScanSession(states []tweaks.TweakState) string {
	sessionID := fmt.Sprintf("tweak-scan-%d", atomic.AddInt64(&tweakStateScanCounter, 1))
	session := &tweakStateScanSession{
		completed: len(states),
		total:     len(states),
		done:      true,
		states:    states,
	}
	tweakStateScanSessions.Store(sessionID, session)
	time.AfterFunc(30*time.Second, func() { tweakStateScanSessions.Delete(sessionID) })
	return tweakStateScanSessionResponse(sessionID)
}

func tweakStateScanSessionResponse(sessionID string) string {
	data, _ := json.Marshal(map[string]string{"sessionId": sessionID})
	return string(data)
}

func invalidateTweakStateCache() {
	tweakStateScanCache.Lock()
	tweakStateScanCache.states = nil
	tweakStateScanCache.updatedAt = time.Time{}
	tweakStateScanCache.activeSessionID = ""
	tweakStateScanCache.generation++
	tweakStateScanCache.Unlock()
}

func (a *API) CheckTweakStateScan(sessionID string) string {
	value, ok := tweakStateScanSessions.Load(sessionID)
	if !ok {
		return `{"error":"Invalid scan session","done":true}`
	}
	session := value.(*tweakStateScanSession)

	session.mu.Lock()
	response := struct {
		Done      bool                `json:"done"`
		Completed int                 `json:"completed"`
		Total     int                 `json:"total"`
		States    []tweaks.TweakState `json:"states,omitempty"`
		Error     string              `json:"error,omitempty"`
	}{
		Done:      session.done,
		Completed: session.completed,
		Total:     session.total,
		Error:     session.err,
	}
	if session.done {
		response.States = session.states
	}
	session.mu.Unlock()

	data, _ := json.Marshal(response)
	if response.Done {
		tweakStateScanSessions.Delete(sessionID)
	}
	return string(data)
}

// --- Async tweak sessions ---

type tweakSession struct {
	mu        sync.Mutex
	progress  tweaks.TweakProgress
	done      bool
	final     []tweaks.TweakResult
	err       string
	cancelled bool
}

var tweakSessions sync.Map
var tweakSessionCounter int64

func (a *API) ApplyTweaks(tweaksJSON string) string {
	a.configureTweakStateDir()
	if a.TweaksData == nil {
		a.GetTweaksData()
	}
	if a.TweaksData == nil {
		return `{"error": "No tweaks data"}`
	}

	var selected []tweaks.Tweak
	if err := json.Unmarshal([]byte(tweaksJSON), &selected); err != nil {
		return `{"error": "Invalid tweaks JSON"}`
	}

	sessionID := fmt.Sprintf("tweak-%d", atomic.AddInt64(&tweakSessionCounter, 1))
	session := &tweakSession{}
	tweakSessions.Store(sessionID, session)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				session.mu.Lock()
				session.err = fmt.Sprintf("tweak panic: %v", r)
				session.done = true
				session.mu.Unlock()
			}
		}()

		cancelCheck := func() bool {
			session.mu.Lock()
			defer session.mu.Unlock()
			return session.cancelled
		}

		results := tweaks.ApplyTweaks(selected, func(p tweaks.TweakProgress) {
			session.mu.Lock()
			session.progress = p
			session.mu.Unlock()
		}, cancelCheck)

		session.mu.Lock()
		session.final = results
		session.done = true
		session.mu.Unlock()
		invalidateTweakStateCache()
		slog.Info("tweak session completed", "id", sessionID, "results", len(results))

		// Self-cleanup: remove session after 30s (in case user closed overlay)
		time.AfterFunc(30*time.Second, func() { tweakSessions.Delete(sessionID) })
	}()

	data, _ := json.Marshal(map[string]string{"sessionId": sessionID})
	return string(data)
}

func (a *API) CheckTweakProgress(sessionID string) string {
	val, ok := tweakSessions.Load(sessionID)
	if !ok {
		return `{"error":"Invalid session","done":true}`
	}
	session := val.(*tweakSession)

	session.mu.Lock()
	defer session.mu.Unlock()

	p := session.progress
	resp := map[string]interface{}{
		"done":      session.done,
		"cancelled": session.cancelled,
		"current":   p.Current,
		"total":     p.Total,
		"program":   p.Name,
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
		tweakSessions.Delete(sessionID)
	}

	return string(data)
}

// CancelOperation marks an install or tweak session as cancelled.
// The goroutine checks this flag and stops processing remaining items.
func (a *API) CancelOperation(sessionID string) string {
	if val, ok := installSessions.Load(sessionID); ok {
		s := val.(*installSession)
		s.mu.Lock()
		s.cancelled = true
		s.mu.Unlock()
		slog.Info("install session cancelled", "id", sessionID)
		return `{"cancelled":true}`
	}
	if val, ok := tweakSessions.Load(sessionID); ok {
		s := val.(*tweakSession)
		s.mu.Lock()
		s.cancelled = true
		s.mu.Unlock()
		slog.Info("tweak session cancelled", "id", sessionID)
		return `{"cancelled":true}`
	}
	return `{"error":"Session not found"}`
}

func (a *API) RevertTweak(id string) string {
	a.configureTweakStateDir()
	result := tweaks.RevertTweak(id)
	invalidateTweakStateCache()
	data, _ := json.Marshal(result)
	return string(data)
}
