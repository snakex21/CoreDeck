package api

import (
	"encoding/json"
	"testing"

	"launcher/internal/tweaks"
)

func TestCompletedTweakStateScanSession(t *testing.T) {
	startedJSON := completedTweakStateScanSession([]tweaks.TweakState{
		{ID: "example", Active: true, Known: true},
	})
	var started struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.Unmarshal([]byte(startedJSON), &started); err != nil {
		t.Fatalf("invalid start response: %v", err)
	}
	if started.SessionID == "" {
		t.Fatal("missing session id")
	}

	var result struct {
		Done      bool                `json:"done"`
		Completed int                 `json:"completed"`
		Total     int                 `json:"total"`
		States    []tweaks.TweakState `json:"states"`
	}
	if err := json.Unmarshal([]byte((&API{}).CheckTweakStateScan(started.SessionID)), &result); err != nil {
		t.Fatalf("invalid scan response: %v", err)
	}
	if !result.Done || result.Completed != 1 || result.Total != 1 || len(result.States) != 1 {
		t.Fatalf("unexpected completed scan response: %+v", result)
	}
}

func TestInvalidateTweakStateCache(t *testing.T) {
	tweakStateScanCache.Lock()
	tweakStateScanCache.states = []tweaks.TweakState{{ID: "cached"}}
	before := tweakStateScanCache.generation
	tweakStateScanCache.Unlock()

	invalidateTweakStateCache()

	tweakStateScanCache.Lock()
	defer tweakStateScanCache.Unlock()
	if len(tweakStateScanCache.states) != 0 {
		t.Fatal("expected cached states to be cleared")
	}
	if tweakStateScanCache.generation != before+1 {
		t.Fatalf("expected generation %d, got %d", before+1, tweakStateScanCache.generation)
	}
}
