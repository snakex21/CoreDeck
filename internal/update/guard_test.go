package update

import (
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type fakeProc struct {
	pid    int
	done   chan struct{}
	once   sync.Once
	mu     sync.Mutex
	killed bool
	exe    string
	args   []string
}

func (p *fakeProc) PID() int              { return p.pid }
func (p *fakeProc) Done() <-chan struct{} { return p.done }
func (p *fakeProc) exit()                 { p.once.Do(func() { close(p.done) }) }
func (p *fakeProc) Kill() error {
	p.mu.Lock()
	p.killed = true
	p.mu.Unlock()
	p.exit()
	return nil
}
func (p *fakeProc) wasKilled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.killed
}

type fakeStarter struct {
	mu      sync.Mutex
	started []*fakeProc
	onStart func(p *fakeProc)
}

func (s *fakeStarter) start(exe string, args []string, dir string) (Process, error) {
	s.mu.Lock()
	p := &fakeProc{pid: 1000 + len(s.started), done: make(chan struct{}), exe: exe, args: args}
	s.started = append(s.started, p)
	s.mu.Unlock()
	if s.onStart != nil {
		go s.onStart(p)
	}
	return p, nil
}

func (s *fakeStarter) all() []*fakeProc {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*fakeProc(nil), s.started...)
}

// committedUpdate installs "2.5.0" over "2.4.0" and records a pending swap.
func committedUpdate(t *testing.T, mode string) string {
	t.Helper()
	dir := newAppDir(t)
	stageNew(t, dir)
	files, err := Commit(dir)
	if err != nil {
		t.Fatal(err)
	}
	st := State{Pending: &Pending{Mode: mode, ParentPID: 4242, From: "2.4.0", To: "2.5.0", Files: files,
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano)}}
	if err := SaveState(StatePath(dir), st); err != nil {
		t.Fatal(err)
	}
	return dir
}

func guardEnv(dir string, s *fakeStarter, parentWaits *[]int) GuardEnv {
	return GuardEnv{
		ExeDir:        dir,
		Start:         s.start,
		WaitParent:    func(pid int, _ time.Duration) bool { *parentWaits = append(*parentWaits, pid); return true },
		HealthTimeout: 300 * time.Millisecond,
		PollInterval:  5 * time.Millisecond,
		Logf:          func(f string, a ...any) {},
	}
}

func assertOldVersionRestored(t *testing.T, dir string) {
	t.Helper()
	if readFile(t, filepath.Join(dir, ExeName)) != string(oldExe) || readFile(t, target(dir, HelperRel)) != string(oldHelper) {
		t.Fatal("old version not restored")
	}
	if exists(filepath.Join(dir, ExeName+OldSuffix)) || exists(target(dir, HelperRel)+OldSuffix) {
		t.Fatal(".old files must be moved back")
	}
	st, err := LoadState(StatePath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if st.Pending != nil || st.JustUpdated != "" || st.RolledBack == nil || st.RolledBack.From != "2.4.0" || st.RolledBack.To != "2.5.0" {
		t.Fatalf("state = %#v", st)
	}
}

func TestGuardSuccessWhenMarkerWritten(t *testing.T) {
	dir := committedUpdate(t, ModeUpdate)
	s := &fakeStarter{onStart: func(p *fakeProc) {
		time.Sleep(20 * time.Millisecond)
		_ = WriteMarker(MarkerPath(dir), Marker{PID: p.pid, Version: "2.5.0"})
	}}
	var waited []int
	out, err := RunGuard(guardEnv(dir, s, &waited))
	if err != nil || out != GuardSuccess {
		t.Fatalf("RunGuard = %s, %v", out, err)
	}
	if len(waited) == 0 || waited[0] != 4242 {
		t.Fatalf("guard must wait for the old app first, waited=%v", waited)
	}
	procs := s.all()
	if len(procs) != 1 || procs[0].exe != filepath.Join(dir, ExeName) || len(procs[0].args) != 1 || procs[0].args[0] != FlagUpdated {
		t.Fatalf("started %#v", procs)
	}
	if procs[0].wasKilled() {
		t.Fatal("healthy app must not be killed")
	}
	if readFile(t, filepath.Join(dir, ExeName)) != string(newExe) {
		t.Fatal("new version must stay")
	}
	st, _ := LoadState(StatePath(dir))
	if st.Pending != nil || st.JustUpdated != "2.5.0" || st.RolledBack != nil {
		t.Fatalf("state = %#v", st)
	}
	if exists(MarkerPath(dir)) {
		t.Fatal("marker must be removed")
	}
}

func TestGuardMarkerFromOtherPIDIsIgnored(t *testing.T) {
	dir := committedUpdate(t, ModeUpdate)
	s := &fakeStarter{onStart: func(p *fakeProc) {
		if p.pid == 1000 {
			_ = WriteMarker(MarkerPath(dir), Marker{PID: 7, Version: "2.5.0"})
		}
	}}
	var waited []int
	if out, _ := RunGuard(guardEnv(dir, s, &waited)); out != GuardRolledBack {
		t.Fatalf("outcome = %s", out)
	}
}

func TestGuardChildExitsEarlyRollsBack(t *testing.T) {
	dir := committedUpdate(t, ModeUpdate)
	s := &fakeStarter{onStart: func(p *fakeProc) {
		if p.pid == 1000 {
			p.exit() // the new version crashes on start
		}
	}}
	var waited []int
	out, err := RunGuard(guardEnv(dir, s, &waited))
	if err != nil || out != GuardRolledBack {
		t.Fatalf("RunGuard = %s, %v", out, err)
	}
	assertOldVersionRestored(t, dir)
	procs := s.all()
	if len(procs) != 2 {
		t.Fatalf("expected the new and the restored app to be started, got %d", len(procs))
	}
	if procs[0].wasKilled() {
		t.Fatal("an exited child must not be killed")
	}
	if procs[1].exe != filepath.Join(dir, ExeName) || len(procs[1].args) != 0 {
		t.Fatalf("restored start = %#v", procs[1])
	}
}

func TestGuardTimeoutKillsChildAndRollsBack(t *testing.T) {
	dir := committedUpdate(t, ModeUpdate)
	s := &fakeStarter{} // child never reports and never exits
	var waited []int
	start := time.Now()
	out, err := RunGuard(guardEnv(dir, s, &waited))
	if err != nil || out != GuardRolledBack {
		t.Fatalf("RunGuard = %s, %v", out, err)
	}
	if time.Since(start) < 300*time.Millisecond {
		t.Fatal("guard gave up before the health timeout")
	}
	procs := s.all()
	if !procs[0].wasKilled() {
		t.Fatal("the hanging new version must be killed")
	}
	if procs[1].wasKilled() {
		t.Fatal("restored app must not be killed")
	}
	assertOldVersionRestored(t, dir)
}

func TestGuardKillsParentThatDoesNotExit(t *testing.T) {
	dir := committedUpdate(t, ModeUpdate)
	s := &fakeStarter{onStart: func(p *fakeProc) {
		_ = WriteMarker(MarkerPath(dir), Marker{PID: p.pid})
	}}
	env := guardEnv(dir, s, new([]int))
	calls := 0
	env.WaitParent = func(int, time.Duration) bool { calls++; return calls > 1 }
	var killed []int
	env.KillParent = func(pid int) { killed = append(killed, pid) }
	if out, err := RunGuard(env); err != nil || out != GuardSuccess {
		t.Fatalf("RunGuard = %s, %v", out, err)
	}
	if len(killed) != 1 || killed[0] != 4242 {
		t.Fatalf("killed = %v", killed)
	}
}

func TestGuardRestoreModeJustRestarts(t *testing.T) {
	dir := committedUpdate(t, ModeRestore)
	s := &fakeStarter{}
	var waited []int
	out, err := RunGuard(guardEnv(dir, s, &waited))
	if err != nil || out != GuardRestored {
		t.Fatalf("RunGuard = %s, %v", out, err)
	}
	procs := s.all()
	if len(procs) != 1 || len(procs[0].args) != 0 {
		t.Fatalf("started = %#v", procs)
	}
	if st, _ := LoadState(StatePath(dir)); st.Pending != nil {
		t.Fatal("pending must be cleared")
	}
}

func TestGuardNothingPending(t *testing.T) {
	dir := newAppDir(t)
	s := &fakeStarter{}
	if out, err := RunGuard(guardEnv(dir, s, new([]int))); err != nil || out != GuardNothing || len(s.all()) != 0 {
		t.Fatalf("RunGuard = %s, %v", out, err)
	}
}

func TestValidateGuardStatePath(t *testing.T) {
	dir := t.TempDir()
	if err := ValidateGuardStatePath(dir, StatePath(dir)); err != nil {
		t.Fatal(err)
	}
	if err := ValidateGuardStatePath(dir, filepath.Join(dir, "other.json")); err == nil {
		t.Fatal("foreign state path accepted")
	}
}
