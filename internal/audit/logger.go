package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const maxAuditBytes int64 = 5 * 1024 * 1024

type Entry struct {
	Timestamp string         `json:"timestamp"`
	Action    string         `json:"action"`
	Target    string         `json:"target,omitempty"`
	Status    string         `json:"status"`
	SessionID string         `json:"sessionId,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

type Logger struct {
	path string
	mu   sync.Mutex
	now  func() time.Time
}

func New(path string) *Logger {
	return &Logger{path: path, now: time.Now}
}

func (l *Logger) Record(action, target, status, sessionID string, details map[string]any) error {
	if l == nil || l.path == "" {
		return nil
	}
	if action == "" || status == "" {
		return fmt.Errorf("audit action and status are required")
	}

	entry := Entry{
		Timestamp: l.now().UTC().Format(time.RFC3339Nano),
		Action:    action,
		Target:    target,
		Status:    status,
		SessionID: sessionID,
		Details:   details,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}
	data = append(data, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return err
	}
	if info, err := os.Stat(l.path); err == nil && info.Size()+int64(len(data)) > maxAuditBytes {
		backup := l.path + ".1"
		_ = os.Remove(backup)
		if err := os.Rename(l.path, backup); err != nil {
			return fmt.Errorf("rotate audit log: %w", err)
		}
	}

	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	return file.Sync()
}
