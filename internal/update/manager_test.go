package update

import (
	"errors"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestManager(t *testing.T, dir string, f *fakeRelease) (*Manager, *atomic.Int32, *atomic.Int32) {
	t.Helper()
	var guards, quits atomic.Int32
	m := &Manager{
		ExeDir:       dir,
		ExePath:      filepath.Join(dir, ExeName),
		Current:      "2.4.0",
		Policy:       f.policy(),
		Client:       f.srv.Client(),
		PID:          func() int { return 4242 },
		StartGuard:   func(string) error { guards.Add(1); return nil },
		Quit:         func() { quits.Add(1) },
		RestartDelay: 10 * time.Millisecond,
	}
	return m, &guards, &quits
}

func waitProgress(t *testing.T, m *Manager, id string, final ...string) Progress {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		p := m.Progress(id)
		for _, f := range final {
			if p.State == f {
				return p
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("session did not finish: %#v", m.Progress(id))
	return Progress{}
}

func TestManagerInstallEndToEnd(t *testing.T) {
	dir := newAppDir(t)
	z := goodZip(t)
	f := newFakeRelease(t, z, sumsFor(z))
	m, guards, quits := newTestManager(t, dir, f)

	res := m.Check(t.Context(), ModeManual)
	if !res.UpdateAvailable || res.LatestVersion != "2.5.0" || res.Error != "" {
		t.Fatalf("check = %#v", res)
	}
	id, err := m.StartUpdate("2.5.0")
	if err != nil {
		t.Fatal(err)
	}
	p := waitProgress(t, m, id, StateRestarting, StateError)
	if p.State != StateRestarting || p.Percent != 100 {
		t.Fatalf("progress = %#v", p)
	}
	if guards.Load() != 1 {
		t.Fatal("guard not started")
	}
	deadline := time.Now().Add(2 * time.Second)
	for quits.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if quits.Load() != 1 {
		t.Fatal("app must quit after restarting")
	}
	if readFile(t, filepath.Join(dir, ExeName)) != string(newExe) || readFile(t, filepath.Join(dir, ExeName+OldSuffix)) != string(oldExe) {
		t.Fatal("files not swapped")
	}
	b := ListBackups(dir)
	if len(b) != 1 || b[0].Version != "2.4.0" || readFile(t, filepath.Join(b[0].Dir, ExeName)) != string(oldExe) {
		t.Fatalf("backup = %#v", b)
	}
	st, _ := LoadState(StatePath(dir))
	if st.Pending == nil || st.Pending.Mode != ModeUpdate || st.Pending.ParentPID != 4242 || st.Pending.From != "2.4.0" || st.Pending.To != "2.5.0" {
		t.Fatalf("pending = %#v", st.Pending)
	}
	if _, err := m.StartUpdate("2.5.0"); err == nil {
		t.Fatal("a second update must be refused while restarting")
	}
}

func TestManagerInstallBadChecksumReportsCode(t *testing.T) {
	dir := newAppDir(t)
	before := snapshot(t, dir)
	z := goodZip(t)
	f := newFakeRelease(t, z, strings.Replace(sumsFor(z), sha(z), sha([]byte("x")), 1))
	m, guards, _ := newTestManager(t, dir, f)
	id, err := m.StartUpdate("2.5.0")
	if err != nil {
		t.Fatal(err)
	}
	p := waitProgress(t, m, id, StateRestarting, StateError)
	if p.State != StateError || p.ErrorCode != CodeBadChecksum || p.Error == "" {
		t.Fatalf("progress = %#v", p)
	}
	if guards.Load() != 0 {
		t.Fatal("guard must not start")
	}
	assertSnapshot(t, before, snapshot(t, dir))
	if len(ListBackups(dir)) != 0 {
		t.Fatal("no backup before verification succeeded")
	}
	// A new attempt is allowed after a failure.
	if _, err := m.StartUpdate("2.5.0"); err != nil {
		t.Fatalf("retry refused: %v", err)
	}
}

func TestManagerGuardStartFailureRestoresFiles(t *testing.T) {
	dir := newAppDir(t)
	z := goodZip(t)
	f := newFakeRelease(t, z, sumsFor(z))
	m, _, quits := newTestManager(t, dir, f)
	m.StartGuard = func(string) error { return errors.New("blocked") }
	id, err := m.StartUpdate("2.5.0")
	if err != nil {
		t.Fatal(err)
	}
	p := waitProgress(t, m, id, StateRestarting, StateError)
	if p.State != StateError || p.ErrorCode != CodeInstall {
		t.Fatalf("progress = %#v", p)
	}
	if readFile(t, filepath.Join(dir, ExeName)) != string(oldExe) || exists(filepath.Join(dir, ExeName+OldSuffix)) {
		t.Fatal("files must be restored")
	}
	if st, _ := LoadState(StatePath(dir)); st.Pending != nil {
		t.Fatal("pending must be cleared")
	}
	time.Sleep(30 * time.Millisecond)
	if quits.Load() != 0 {
		t.Fatal("app must keep running")
	}
}

func TestManagerStartUpdateValidation(t *testing.T) {
	dir := newAppDir(t)
	z := goodZip(t)
	f := newFakeRelease(t, z, sumsFor(z))
	m, _, _ := newTestManager(t, dir, f)
	for _, v := range []string{"", "x", "2.4.0", "2.3.9"} {
		if _, err := m.StartUpdate(v); err == nil {
			t.Errorf("StartUpdate(%q) accepted", v)
		}
	}
	m.ExePath = filepath.Join(dir, "CoreDeck (1).exe")
	if _, err := m.StartUpdate("2.5.0"); err == nil {
		t.Error("renamed executable accepted")
	}
	m.ExePath = filepath.Join(dir, ExeName)
	id, _ := m.StartUpdate("9.9.9") // not published
	if p := waitProgress(t, m, id, StateError, StateRestarting); p.ErrorCode != CodeNoAsset {
		t.Fatalf("progress = %#v", p)
	}
	if p := m.Progress("nope"); p.State != StateError {
		t.Fatalf("unknown session = %#v", p)
	}
}

func TestManagerAutoCheckHonoursSettings(t *testing.T) {
	dir := newAppDir(t)
	z := goodZip(t)
	f := newFakeRelease(t, z, sumsFor(z))
	m, _, _ := newTestManager(t, dir, f)
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	m.Now = func() time.Time { return now }

	if res := m.Check(t.Context(), ModeAuto); !res.UpdateAvailable {
		t.Fatalf("first auto check = %#v", res)
	}
	if got := m.Settings().LastCheck; got != now.Format(time.RFC3339) {
		t.Fatalf("lastCheck = %q", got)
	}
	if res := m.Check(t.Context(), ModeAuto); res.UpdateAvailable || f.hits["api"] != 1 {
		t.Fatalf("auto check within 24 h must not contact GitHub: %#v hits=%d", res, f.hits["api"])
	}
	if res := m.Check(t.Context(), ModeManual); !res.UpdateAvailable || f.hits["api"] != 2 {
		t.Fatalf("manual check = %#v", res)
	}

	now = now.Add(25 * time.Hour)
	if err := m.Skip("v2.5.0"); err != nil {
		t.Fatal(err)
	}
	if res := m.Check(t.Context(), ModeAuto); res.UpdateAvailable || !res.Skipped {
		t.Fatalf("skipped version offered: %#v", res)
	}
	if res := m.Check(t.Context(), ModeManual); !res.UpdateAvailable || !res.Skipped {
		t.Fatalf("manual check must ignore skip: %#v", res)
	}

	now = now.Add(25 * time.Hour)
	if err := m.SetSettings(`{"autoCheck":false,"skippedVersion":""}`); err != nil {
		t.Fatal(err)
	}
	hits := f.hits["api"]
	if res := m.Check(t.Context(), ModeAuto); res.UpdateAvailable || f.hits["api"] != hits {
		t.Fatalf("disabled auto check ran: %#v", res)
	}
}

func TestSettingsPersistenceAndValidation(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{ExeDir: dir, Current: "2.4.0"}
	if s := m.Settings(); s != DefaultSettings() {
		t.Fatalf("defaults = %#v", s)
	}
	if err := m.SetSettings(`{"channel":"test","autoCheck":false,"lastCheck":"ignored"}`); err != nil {
		t.Fatal(err)
	}
	s := LoadSettings(SettingsPath(dir))
	if s.Channel != ChannelTest || s.AutoCheck || s.LastCheck != "" {
		t.Fatalf("saved = %#v", s)
	}
	for _, bad := range []string{`{"channel":"nightly"}`, `{"skippedVersion":"abc"}`, `not json`, `{"autoCheck":"yes"}`} {
		if err := m.SetSettings(bad); err == nil {
			t.Errorf("SetSettings(%s) accepted", bad)
		}
	}
	if err := m.Skip("nope"); err == nil {
		t.Error("invalid skip accepted")
	}
	if filepath.Base(filepath.Dir(SettingsPath(dir))) != "state" {
		t.Fatal("settings must live in data/state")
	}
}

func TestManagerReady(t *testing.T) {
	dir := newAppDir(t)

	// Run started by the guard: writes the marker with its PID.
	st := State{Pending: &Pending{Mode: ModeUpdate, From: "2.4.0", To: "2.5.0", StartedAt: time.Now().UTC().Format(time.RFC3339Nano)}}
	if err := SaveState(StatePath(dir), st); err != nil {
		t.Fatal(err)
	}
	m := &Manager{ExeDir: dir, Current: "2.5.0", LaunchedAfterUpdate: true, PID: func() int { return 77 }}
	if info := m.Ready(); info.JustUpdated != "2.5.0" || info.RolledBack != nil {
		t.Fatalf("ready = %#v", info)
	}
	if mk, ok := readMarker(MarkerPath(dir)); !ok || mk.PID != 77 {
		t.Fatalf("marker = %#v %v", mk, ok)
	}
	if info := m.Ready(); info.JustUpdated != "" {
		t.Fatal("reported twice")
	}

	// Next start after an automatic rollback: reported once, then cleared.
	if err := SaveState(StatePath(dir), State{RolledBack: &RolledBack{From: "2.4.0", To: "2.5.0"}}); err != nil {
		t.Fatal(err)
	}
	m2 := &Manager{ExeDir: dir, Current: "2.4.0"}
	if info := m2.Ready(); info.RolledBack == nil || info.RolledBack.To != "2.5.0" {
		t.Fatalf("ready = %#v", info)
	}
	if s, _ := LoadState(StatePath(dir)); s.RolledBack != nil {
		t.Fatal("rolledBack must be cleared after reporting")
	}
	m3 := &Manager{ExeDir: dir, Current: "2.4.0"}
	if info := m3.Ready(); info.RolledBack != nil || info.JustUpdated != "" {
		t.Fatalf("reported again: %#v", info)
	}
}

func TestManagerManualRollback(t *testing.T) {
	dir := newAppDir(t)
	z := goodZip(t)
	f := newFakeRelease(t, z, sumsFor(z))
	m, guards, _ := newTestManager(t, dir, f)
	if ok, _ := m.RollbackInfo(); ok {
		t.Fatal("no backup yet")
	}
	if err := m.Rollback(); err == nil {
		t.Fatal("rollback without backup must fail")
	}
	if err := CreateBackup(dir, "2.4.0", time.Now()); err != nil {
		t.Fatal(err)
	}
	// Simulate the updated install running 2.5.0.
	writeFile(t, filepath.Join(dir, ExeName), newExe)
	writeFile(t, target(dir, HelperRel), newHelper)
	m.Current = "2.5.0"

	ok, v := m.RollbackInfo()
	if !ok || v != "2.4.0" {
		t.Fatalf("RollbackInfo = %v %q", ok, v)
	}
	if err := m.Rollback(); err != nil {
		t.Fatal(err)
	}
	if readFile(t, filepath.Join(dir, ExeName)) != string(oldExe) || readFile(t, target(dir, HelperRel)) != string(oldHelper) {
		t.Fatal("backup not restored")
	}
	if readFile(t, filepath.Join(dir, ExeName+OldSuffix)) != string(newExe) {
		t.Fatal("running binary must become the guard (.old)")
	}
	st, _ := LoadState(StatePath(dir))
	if guards.Load() != 1 || st.Pending == nil || st.Pending.Mode != ModeRestore || st.Pending.To != "2.4.0" {
		t.Fatalf("guard=%d pending=%#v", guards.Load(), st.Pending)
	}
}
