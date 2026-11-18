package update

import (
	"os"
	"time"
)

// Guard modes.
const (
	ModeUpdate  = "update"  // health check + automatic rollback
	ModeRestore = "restore" // manual rollback to a backup: just restart
)

// PendingStaleAfter: a pending guard older than this is considered dead.
const PendingStaleAfter = 10 * time.Minute

// Pending describes a swap that a guard process still has to finish.
type Pending struct {
	Mode      string          `json:"mode"`
	ParentPID int             `json:"parentPid"`
	From      string          `json:"from"`
	To        string          `json:"to"`
	Files     []CommittedFile `json:"files"`
	StartedAt string          `json:"startedAt"`
}

// Stale reports whether the guard for this swap can no longer be running.
func (p *Pending) Stale(now time.Time) bool {
	t, err := time.Parse(time.RFC3339Nano, p.StartedAt)
	if err != nil {
		return true
	}
	return now.Sub(t) > PendingStaleAfter || t.After(now.Add(time.Minute))
}

// RolledBack is reported once to the UI after an automatic rollback.
type RolledBack struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// State is <exeDir>\.coredeck-update\state.json, shared by the app and the
// guard process.
type State struct {
	Pending     *Pending    `json:"pending,omitempty"`
	JustUpdated string      `json:"justUpdated,omitempty"`
	RolledBack  *RolledBack `json:"rolledBack,omitempty"`
}

// LoadState returns an empty state when the file is missing.
func LoadState(path string) (State, error) {
	var st State
	if err := readJSON(path, &st); err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		}
		return State{}, err
	}
	return st, nil
}

// SaveState writes the state atomically.
func SaveState(path string, st State) error { return writeJSONAtomic(path, st) }

// Marker is written by the new app (update_ready) to prove a healthy start.
type Marker struct {
	PID     int    `json:"pid"`
	Version string `json:"version"`
}

// WriteMarker writes the health marker.
func WriteMarker(path string, m Marker) error { return writeJSONAtomic(path, m) }

func readMarker(path string) (Marker, bool) {
	var m Marker
	if readJSON(path, &m) != nil {
		return m, false
	}
	return m, true
}
