package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// execProcess adapts exec.Cmd to Process.
type execProcess struct {
	cmd  *exec.Cmd
	done chan struct{}
}

func (p *execProcess) PID() int              { return p.cmd.Process.Pid }
func (p *execProcess) Done() <-chan struct{} { return p.done }
func (p *execProcess) Kill() error           { return p.cmd.Process.Kill() }

// StartProcess starts exe and reaps it in the background.
func StartProcess(exe string, args []string, dir string) (Process, error) {
	cmd, err := startDetached(func() *exec.Cmd {
		c := exec.Command(exe, args...)
		c.Dir = dir
		return c
	})
	if err != nil {
		return nil, err
	}
	p := &execProcess{cmd: cmd, done: make(chan struct{})}
	go func() {
		_ = cmd.Wait()
		close(p.done)
	}()
	return p, nil
}

// StartGuardProcess launches <exeDir>\CoreDeck.exe.old --update-guard <state>.
func StartGuardProcess(exeDir string) error {
	guard := target(exeDir, ExeName) + OldSuffix
	cmd, err := startDetached(func() *exec.Cmd {
		c := exec.Command(guard, FlagGuard, StatePath(exeDir))
		c.Dir = exeDir
		return c
	})
	if err != nil {
		return err
	}
	return cmd.Process.Release()
}

// startDetached starts the command outside the parent's job object when
// that is allowed (so it survives the parent), otherwise normally.
func startDetached(build func() *exec.Cmd) (*exec.Cmd, error) {
	cmd := build()
	if configureDetached(cmd, true) {
		if err := cmd.Start(); err == nil {
			return cmd, nil
		}
		cmd = build()
		configureDetached(cmd, false)
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

// parentWatcher keeps one handle per PID, opened on first use (while the
// parent still runs), so a reused PID can never be waited on or killed.
type parentWatcher struct {
	mu    sync.Mutex
	procs map[int]*watched
}

type watched struct {
	proc *os.Process
	done chan struct{}
}

func (w *parentWatcher) get(pid int) *watched {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.procs == nil {
		w.procs = map[int]*watched{}
	}
	if x, ok := w.procs[pid]; ok {
		return x
	}
	x := &watched{done: make(chan struct{})}
	proc, err := os.FindProcess(pid)
	if err != nil {
		close(x.done) // not running (or not accessible): treat as exited
	} else {
		x.proc = proc
		go func() {
			_, _ = proc.Wait()
			close(x.done)
		}()
	}
	w.procs[pid] = x
	return x
}

func (w *parentWatcher) wait(pid int, timeout time.Duration) bool {
	x := w.get(pid)
	select {
	case <-x.done:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (w *parentWatcher) kill(pid int) {
	if x := w.get(pid); x.proc != nil {
		_ = x.proc.Kill()
	}
}

// RealGuardEnv wires the guard to real processes and a log file in the
// work directory.
func RealGuardEnv(exeDir string) GuardEnv {
	w := &parentWatcher{}
	return GuardEnv{
		ExeDir:      exeDir,
		Start:       StartProcess,
		WaitParent:  w.wait,
		KillParent:  w.kill,
		SettleDelay: 1500 * time.Millisecond,
		Logf:        guardLogger(exeDir),
	}
}

// OpenParentEarly opens the parent handle immediately (call first in the
// guard, before anything slow).
func (e GuardEnv) OpenParentEarly() {
	st, err := LoadState(StatePath(e.ExeDir))
	if err != nil || st.Pending == nil || st.Pending.ParentPID <= 0 || e.WaitParent == nil {
		return
	}
	e.WaitParent(st.Pending.ParentPID, 0)
}

func guardLogger(exeDir string) func(string, ...any) {
	path := filepath.Join(WorkDir(exeDir), guardLogName)
	var mu sync.Mutex
	return func(format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		if fi, err := os.Stat(path); err == nil && fi.Size() > maxGuardLogSize {
			_ = os.Remove(path)
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return
		}
		defer f.Close()
		fmt.Fprintf(f, "%s %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
	}
}
