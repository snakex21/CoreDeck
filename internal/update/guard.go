package update

// The guard is the OLD binary (<exeDir>\CoreDeck.exe.old) started with
// --update-guard <statefile> right after a commit, so it does not depend on
// the new code. It waits for the old app to exit, starts the new
// CoreDeck.exe --updated and waits for the health marker written by the new
// app (update_ready). If the new process exits first or the timeout passes,
// it kills only that child, restores the .old files, records rolledBack and
// starts the restored version.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Command-line flags handled by main().
const (
	FlagGuard   = "--update-guard"
	FlagUpdated = "--updated"
)

// Process is a started child process.
type Process interface {
	PID() int
	Done() <-chan struct{} // closed when the process has exited
	Kill() error
}

// GuardEnv holds everything the guard touches outside the file system, so
// tests can run it with fake processes.
type GuardEnv struct {
	ExeDir        string
	Start         func(exe string, args []string, dir string) (Process, error)
	WaitParent    func(pid int, timeout time.Duration) bool // true once exited
	KillParent    func(pid int)
	Now           func() time.Time
	HealthTimeout time.Duration
	ParentTimeout time.Duration
	PollInterval  time.Duration
	// SettleDelay is waited after the old app exited, so its WebView2
	// processes release the shared profile before the new app starts.
	SettleDelay time.Duration
	Logf        func(format string, args ...any)
}

// GuardOutcome is the result of RunGuard.
type GuardOutcome string

const (
	GuardNothing    GuardOutcome = "nothing"
	GuardSuccess    GuardOutcome = "success"
	GuardRolledBack GuardOutcome = "rolled-back"
	GuardRestored   GuardOutcome = "restored"
)

func (e *GuardEnv) defaults() {
	if e.Now == nil {
		e.Now = time.Now
	}
	if e.HealthTimeout <= 0 {
		e.HealthTimeout = 60 * time.Second
	}
	if e.ParentTimeout <= 0 {
		e.ParentTimeout = 30 * time.Second
	}
	if e.PollInterval <= 0 {
		e.PollInterval = 250 * time.Millisecond
	}
	if e.Logf == nil {
		e.Logf = func(string, ...any) {}
	}
	if e.KillParent == nil {
		e.KillParent = func(int) {}
	}
	if e.WaitParent == nil {
		e.WaitParent = func(int, time.Duration) bool { return true }
	}
}

// ValidateGuardStatePath accepts only <exeDir>\.coredeck-update\state.json.
func ValidateGuardStatePath(exeDir, statePath string) error {
	want := filepath.Clean(StatePath(exeDir))
	got, err := filepath.Abs(statePath)
	if err != nil || !strings.EqualFold(filepath.Clean(got), want) {
		return fmt.Errorf("unexpected guard state file %q", statePath)
	}
	return nil
}

// RunGuard finishes a pending swap described in the state file.
func RunGuard(env GuardEnv) (GuardOutcome, error) {
	env.defaults()
	statePath := StatePath(env.ExeDir)
	st, err := LoadState(statePath)
	if err != nil {
		return GuardNothing, fmt.Errorf("cannot read update state: %w", err)
	}
	p := st.Pending
	if p == nil {
		env.Logf("guard: nothing pending")
		return GuardNothing, nil
	}
	env.Logf("guard: mode=%s from=%s to=%s parent=%d", p.Mode, p.From, p.To, p.ParentPID)

	if p.ParentPID > 0 && p.ParentPID != os.Getpid() {
		if !env.WaitParent(p.ParentPID, env.ParentTimeout) {
			env.Logf("guard: old app %d still running after %s, terminating it", p.ParentPID, env.ParentTimeout)
			env.KillParent(p.ParentPID)
			env.WaitParent(p.ParentPID, 5*time.Second)
		}
	}

	if env.SettleDelay > 0 {
		time.Sleep(env.SettleDelay)
	}

	exe := target(env.ExeDir, ExeName)
	if p.Mode == ModeRestore {
		st.Pending = nil
		if err := SaveState(statePath, st); err != nil {
			env.Logf("guard: cannot save state: %v", err)
		}
		if _, err := env.Start(exe, nil, env.ExeDir); err != nil {
			env.Logf("guard: cannot start restored version: %v", err)
			return GuardRestored, err
		}
		return GuardRestored, nil
	}

	marker := MarkerPath(env.ExeDir)
	_ = os.Remove(marker)
	child, startErr := env.Start(exe, []string{FlagUpdated}, env.ExeDir)
	healthy := false
	if startErr != nil {
		env.Logf("guard: cannot start new version: %v", startErr)
	} else {
		healthy = waitHealthy(env, child, marker)
	}
	_ = os.Remove(marker)

	if healthy {
		env.Logf("guard: version %s started correctly", p.To)
		st.Pending = nil
		st.JustUpdated = p.To
		st.RolledBack = nil
		return GuardSuccess, SaveState(statePath, st)
	}

	if child != nil {
		select {
		case <-child.Done():
		default:
			env.Logf("guard: terminating new version (pid %d)", child.PID())
			_ = child.Kill()
			select {
			case <-child.Done():
			case <-time.After(10 * time.Second):
			}
		}
	}
	restoreErr := RestoreOld(env.ExeDir, p.Files)
	if restoreErr != nil {
		env.Logf("guard: restore failed: %v", restoreErr)
	}
	st.Pending = nil
	st.JustUpdated = ""
	st.RolledBack = &RolledBack{From: p.From, To: p.To}
	saveErr := SaveState(statePath, st)
	if _, err := env.Start(exe, nil, env.ExeDir); err != nil {
		env.Logf("guard: cannot start restored version: %v", err)
		return GuardRolledBack, errors.Join(restoreErr, saveErr, err)
	}
	env.Logf("guard: rolled back to %s", p.From)
	return GuardRolledBack, errors.Join(restoreErr, saveErr)
}

func waitHealthy(env GuardEnv, child Process, marker string) bool {
	deadline := env.Now().Add(env.HealthTimeout)
	tick := time.NewTicker(env.PollInterval)
	defer tick.Stop()
	for {
		if m, ok := readMarker(marker); ok && m.PID == child.PID() {
			return true
		}
		select {
		case <-child.Done():
			// one last look: the app may have written the marker and exited
			if m, ok := readMarker(marker); ok && m.PID == child.PID() {
				return true
			}
			env.Logf("guard: new version exited before reporting a healthy start")
			return false
		case <-tick.C:
		}
		if !env.Now().Before(deadline) {
			env.Logf("guard: no health report within %s", env.HealthTimeout)
			return false
		}
	}
}
