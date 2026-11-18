package main

import "testing"

func TestUpdateGuardArg(t *testing.T) {
	if p, ok := updateGuardArg([]string{"--update-guard", `C:\app\.coredeck-update\state.json`}); !ok || p != `C:\app\.coredeck-update\state.json` {
		t.Fatalf("got %q %v", p, ok)
	}
	if p, ok := updateGuardArg([]string{"--UPDATE-GUARD"}); !ok || p != "" {
		t.Fatalf("flag without a path must still be handled as the guard: %q %v", p, ok)
	}
	if _, ok := updateGuardArg([]string{"--updated"}); ok {
		t.Fatal("--updated is not the guard flag")
	}
	if !hasArg([]string{"x", "--Updated"}, "--updated") || hasArg(nil, "--updated") {
		t.Fatal("hasArg")
	}
}

func TestRunUpdateGuardRejectsForeignStateFile(t *testing.T) {
	dir := t.TempDir()
	if code := runUpdateGuard(dir, dir+`\elsewhere.json`); code != 2 {
		t.Fatalf("exit code = %d", code)
	}
}
