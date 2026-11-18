package update

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Progress states reported by update_progress.
const (
	StateDownloading = "downloading"
	StateVerifying   = "verifying"
	StateInstalling  = "installing"
	StateRestarting  = "restarting"
	StateError       = "error"
)

// Progress is the JSON returned by update_progress.
type Progress struct {
	State     string `json:"state"`
	Percent   int    `json:"percent"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"errorCode,omitempty"`
}

// ReadyInfo is the JSON returned by update_ready.
type ReadyInfo struct {
	JustUpdated string      `json:"justUpdated"`
	RolledBack  *RolledBack `json:"rolledBack"`
}

// Manager drives checks, installs and rollbacks for the running app.
type Manager struct {
	ExeDir  string
	ExePath string // running executable
	Current string
	Policy  Policy
	Client  *http.Client

	// LaunchedAfterUpdate is true when this run was started with --updated.
	LaunchedAfterUpdate bool

	Now          func() time.Time
	PID          func() int
	StartGuard   func(exeDir string) error
	Quit         func() // terminates the app; called RestartDelay after "restarting"
	RestartDelay time.Duration

	mu            sync.Mutex
	settingsMu    sync.Mutex
	sessions      map[string]*session
	busy          bool
	readyReported bool
}

type session struct {
	mu sync.Mutex
	p  Progress
}

func (s *session) set(state string, percent int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.p.State, s.p.Percent = state, percent
}

func (s *session) fail(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.p.State = StateError
	s.p.Error = err.Error()
	s.p.ErrorCode = ErrorCode(err)
}

func (s *session) get() Progress {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.p
}

// NewManager builds a production manager.
func NewManager(exeDir, current string, launchedAfterUpdate bool) *Manager {
	exe, _ := os.Executable()
	return &Manager{
		ExeDir:              exeDir,
		ExePath:             exe,
		Current:             current,
		Policy:              DefaultPolicy(),
		LaunchedAfterUpdate: launchedAfterUpdate,
	}
}

func (m *Manager) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now()
}

func (m *Manager) pid() int {
	if m.PID != nil {
		return m.PID()
	}
	return os.Getpid()
}

func (m *Manager) restartDelay() time.Duration {
	if m.RestartDelay > 0 {
		return m.RestartDelay
	}
	return 1500 * time.Millisecond
}

// ---------------------------------------------------------------------------
// Settings and checks
// ---------------------------------------------------------------------------

func (m *Manager) Settings() Settings {
	m.settingsMu.Lock()
	defer m.settingsMu.Unlock()
	return LoadSettings(SettingsPath(m.ExeDir))
}

func (m *Manager) updateSettings(fn func(*Settings) error) error {
	m.settingsMu.Lock()
	defer m.settingsMu.Unlock()
	s := LoadSettings(SettingsPath(m.ExeDir))
	if err := fn(&s); err != nil {
		return err
	}
	return SaveSettings(SettingsPath(m.ExeDir), s)
}

// SetSettings applies a JSON patch from the UI.
func (m *Manager) SetSettings(raw string) error {
	return m.updateSettings(func(s *Settings) error {
		n, err := ApplyPatch(*s, raw)
		if err != nil {
			return err
		}
		*s = n
		return nil
	})
}

// Skip records a version the automatic check should not offer.
func (m *Manager) Skip(version string) error {
	v, err := normalizeOptionalVersion(version)
	if err != nil {
		return err
	}
	return m.updateSettings(func(s *Settings) error {
		s.SkippedVersion = v
		return nil
	})
}

// Check runs an automatic or manual check.
func (m *Manager) Check(ctx context.Context, mode string) CheckResult {
	if mode != ModeAuto {
		mode = ModeManual
	}
	s := m.Settings()
	res := CheckResult{CurrentVersion: m.Current, Channel: NormalizeChannel(s.Channel)}
	if mode == ModeAuto && !ShouldAutoCheck(s, m.now()) {
		return res
	}
	_ = m.updateSettings(func(cur *Settings) error {
		cur.LastCheck = m.now().UTC().Format(time.RFC3339)
		return nil
	})
	releases, err := FetchReleases(ctx, m.Client, m.Policy)
	if err != nil {
		res.Error = err.Error()
		res.ErrorCode = ErrorCode(err)
		return res
	}
	return BuildResult(releases, m.Current, s, mode)
}

// ---------------------------------------------------------------------------
// Install
// ---------------------------------------------------------------------------

func newSessionID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func (m *Manager) checkLayout() error {
	if m.ExePath == "" {
		return newErr(CodeInstall, "cannot locate the running program")
	}
	if !strings.EqualFold(filepath.Base(m.ExePath), ExeName) ||
		!strings.EqualFold(filepath.Clean(filepath.Dir(m.ExePath)), filepath.Clean(m.ExeDir)) {
		return newErr(CodeInstall, "the program file must be named %s to update itself", ExeName)
	}
	return nil
}

func (m *Manager) begin() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.busy {
		return newErr(CodeInstall, "an update is already in progress")
	}
	m.busy = true
	return nil
}

func (m *Manager) end() {
	m.mu.Lock()
	m.busy = false
	m.mu.Unlock()
}

// StartUpdate begins installing the given version in the background.
func (m *Manager) StartUpdate(version string) (string, error) {
	want, err := ParseVersion(version)
	if err != nil {
		return "", newErr(CodeNoAsset, "invalid version %q", version)
	}
	cur, err := ParseVersion(m.Current)
	if err != nil {
		return "", newErr(CodeInstall, "invalid current version %q", m.Current)
	}
	if Compare(want, cur) <= 0 {
		return "", newErr(CodeInstall, "version %s is not newer than %s", want, cur)
	}
	if err := m.checkLayout(); err != nil {
		return "", err
	}
	if err := m.begin(); err != nil {
		return "", err
	}
	id := newSessionID()
	s := &session{p: Progress{State: StateDownloading}}
	m.mu.Lock()
	if m.sessions == nil {
		m.sessions = map[string]*session{}
	}
	m.sessions[id] = s
	m.mu.Unlock()

	go func() {
		err := m.install(context.Background(), want.String(), s)
		if err != nil {
			s.fail(err)
			m.end()
		}
	}()
	return id, nil
}

// Progress returns the state of a session.
func (m *Manager) Progress(id string) Progress {
	m.mu.Lock()
	s := m.sessions[id]
	m.mu.Unlock()
	if s == nil {
		return Progress{State: StateError, Error: "unknown update session", ErrorCode: CodeInstall}
	}
	return s.get()
}

func (m *Manager) install(ctx context.Context, version string, s *session) error {
	settings := m.Settings()
	releases, err := FetchReleases(ctx, m.Client, m.Policy)
	if err != nil {
		return err
	}
	rel := FindRelease(releases, settings.Channel, version)
	if rel == nil {
		return newErr(CodeNoAsset, "release %s was not found on the %s channel", version, NormalizeChannel(settings.Channel))
	}
	_, err = Prepare(ctx, PrepareOptions{
		ExeDir: m.ExeDir, Release: rel, Client: m.Client, Policy: m.Policy,
		Progress: func(stage string, pct int) {
			if stage == StateDownloading {
				s.set(StateDownloading, pct*85/100)
			} else {
				s.set(StateVerifying, 85+pct*10/100)
			}
		},
	})
	if err != nil {
		return err
	}
	s.set(StateInstalling, 96)
	if err := CreateBackup(m.ExeDir, m.Current, m.now()); err != nil {
		RemoveStaging(m.ExeDir)
		return err
	}
	return m.swapAndRestart(ModeUpdate, m.Current, rel.version.String(), s)
}

// swapAndRestart commits staged files, records the pending swap, starts the
// guard from the old binary and schedules the app exit.
func (m *Manager) swapAndRestart(mode, from, to string, s *session) error {
	files, err := Commit(m.ExeDir)
	if err != nil {
		RemoveStaging(m.ExeDir)
		return err
	}
	statePath := StatePath(m.ExeDir)
	st, _ := LoadState(statePath)
	st.Pending = &Pending{
		Mode: mode, ParentPID: m.pid(), From: from, To: to, Files: files,
		StartedAt: m.now().UTC().Format(time.RFC3339Nano),
	}
	_ = os.Remove(MarkerPath(m.ExeDir))
	err = SaveState(statePath, st)
	if err == nil {
		start := m.StartGuard
		if start == nil {
			start = StartGuardProcess
		}
		err = start(m.ExeDir)
	}
	if err != nil {
		restoreErr := RestoreOld(m.ExeDir, files)
		st.Pending = nil
		_ = SaveState(statePath, st)
		return wrapErr(CodeInstall, errors.Join(fmt.Errorf("cannot start the update helper: %w", err), restoreErr))
	}
	if s != nil {
		s.set(StateRestarting, 100)
	}
	time.AfterFunc(m.restartDelay(), func() {
		if m.Quit != nil {
			m.Quit()
		}
	})
	return nil
}

// ---------------------------------------------------------------------------
// Health and rollback
// ---------------------------------------------------------------------------

// Ready is called by the UI once it has loaded. In a run started with
// --updated it writes the health marker for the guard. It reports
// justUpdated / rolledBack once.
func (m *Manager) Ready() ReadyInfo {
	m.mu.Lock()
	if m.readyReported {
		m.mu.Unlock()
		return ReadyInfo{}
	}
	m.readyReported = true
	m.mu.Unlock()

	statePath := StatePath(m.ExeDir)
	st, err := LoadState(statePath)
	if err != nil {
		return ReadyInfo{}
	}
	if st.Pending != nil {
		if m.LaunchedAfterUpdate && st.Pending.Mode == ModeUpdate {
			_ = WriteMarker(MarkerPath(m.ExeDir), Marker{PID: m.pid(), Version: m.Current})
			go m.clearJustUpdatedLater(m.Current)
			return ReadyInfo{JustUpdated: m.Current}
		}
		return ReadyInfo{}
	}
	info := ReadyInfo{JustUpdated: st.JustUpdated, RolledBack: st.RolledBack}
	if st.JustUpdated != "" || st.RolledBack != nil {
		st.JustUpdated, st.RolledBack = "", nil
		_ = SaveState(statePath, st)
	}
	return info
}

// clearJustUpdatedLater waits for the guard to finish and then clears the
// justUpdated flag it records, since this run has already reported it.
func (m *Manager) clearJustUpdatedLater(version string) {
	statePath := StatePath(m.ExeDir)
	deadline := m.now().Add(2 * time.Minute)
	for m.now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		st, err := LoadState(statePath)
		if err != nil || st.Pending != nil {
			continue
		}
		if st.JustUpdated == version {
			st.JustUpdated = ""
			_ = SaveState(statePath, st)
		}
		return
	}
}

// RollbackInfo reports the newest backup that can be restored.
func (m *Manager) RollbackInfo() (bool, string) {
	b, ok := LatestBackup(m.ExeDir, m.Current)
	if !ok {
		return false, ""
	}
	return true, b.Version
}

// Rollback restores the newest backup and restarts through the guard.
func (m *Manager) Rollback() error {
	b, ok := LatestBackup(m.ExeDir, m.Current)
	if !ok {
		return newErr(CodeInstall, "no previous version is available")
	}
	if err := m.checkLayout(); err != nil {
		return err
	}
	if err := m.begin(); err != nil {
		return err
	}
	if err := StageBackup(m.ExeDir, b); err != nil {
		m.end()
		return err
	}
	if err := m.swapAndRestart(ModeRestore, m.Current, b.Version, nil); err != nil {
		m.end()
		return err
	}
	return nil
}
