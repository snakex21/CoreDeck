package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestRecordWritesValidJSONLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "audit.jsonl")
	logger := New(path)
	logger.now = func() time.Time {
		return time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	}

	if err := logger.Record("tweak.apply", "show-extensions", "completed", "tweak-1", map[string]any{
		"success": true,
	}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		t.Fatal("audit log is empty")
	}
	var entry Entry
	if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSONL entry: %v", err)
	}
	if entry.Action != "tweak.apply" || entry.SessionID != "tweak-1" ||
		entry.Timestamp != "2026-07-30T12:00:00Z" {
		t.Fatalf("unexpected entry: %#v", entry)
	}
}

func TestRecordSerializesConcurrentWriters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	logger := New(path)
	var group sync.WaitGroup
	for i := 0; i < 20; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := logger.Record("install.programs", "", "completed", "", nil); err != nil {
				t.Errorf("Record() error = %v", err)
			}
		}()
	}
	group.Wait()

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("invalid entry %d: %v", count, err)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 20 {
		t.Fatalf("entry count = %d, want 20", count)
	}
}

func TestRecordRejectsIncompleteEntry(t *testing.T) {
	logger := New(filepath.Join(t.TempDir(), "audit.jsonl"))
	if err := logger.Record("", "", "completed", "", nil); err == nil {
		t.Fatal("Record() should reject an empty action")
	}
}
