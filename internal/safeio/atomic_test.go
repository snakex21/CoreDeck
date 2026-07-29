package safeio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicKeepsPreviousVersionAsBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteFileAtomic(path, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	backup, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != "new" {
		t.Fatalf("current file = %q, want new", current)
	}
	if string(backup) != "old" {
		t.Fatalf("backup file = %q, want old", backup)
	}
}
