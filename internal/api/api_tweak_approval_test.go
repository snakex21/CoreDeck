package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"launcher/internal/tweaks"
)

func approvalAPI(t *testing.T) *API {
	t.Helper()
	return &API{
		ExeDir: t.TempDir(),
		TweaksData: &tweaks.TweaksData{Tweaks: []tweaks.Tweak{
			{ID: "a", Name: "A", Type: "powershell", Command: "Write-Output a"},
			{ID: "b", Name: "B", Type: "registry", Command: `reg add "HKCU\Software\X" /v Y /d 1 /f`},
		}},
	}
}

func preview(t *testing.T, a *API, ids ...string) []TweakPreview {
	t.Helper()
	raw, _ := json.Marshal(ids)
	var out struct {
		Tweaks []TweakPreview `json:"tweaks"`
		Error  string         `json:"error"`
	}
	if err := json.Unmarshal([]byte(a.PreviewTweaks(string(raw))), &out); err != nil {
		t.Fatal(err)
	}
	if out.Error != "" {
		t.Fatalf("preview error: %s", out.Error)
	}
	return out.Tweaks
}

func TestPreviewShowsCatalogueCommand(t *testing.T) {
	a := approvalAPI(t)
	p := preview(t, a, "a", "b", "a")
	if len(p) != 2 || p[0].Command != "Write-Output a" || p[0].Approved || p[1].ID != "b" {
		t.Fatalf("unexpected preview: %+v", p)
	}
	if p[0].Hash != TweakScriptHash(a.TweaksData.Tweaks[0]) {
		t.Fatal("hash mismatch")
	}
	var bad map[string]any
	_ = json.Unmarshal([]byte(a.PreviewTweaks(`["nope"]`)), &bad)
	if bad["error"] == nil {
		t.Fatal("unknown id must be an error")
	}
}

func TestApplyRefusesUnapprovedTweaks(t *testing.T) {
	a := approvalAPI(t)
	// Legacy payload (tweak objects) carrying a different command: rejected
	// before anything runs, and only the catalogue entry would ever be used.
	resp := a.ApplyTweaks(`[{"id":"a","type":"powershell","command":"Stop-Computer -Force"}]`)
	var out struct {
		Error         string   `json:"error"`
		NeedsApproval []string `json:"needsApproval"`
		SessionID     string   `json:"sessionId"`
	}
	if err := json.Unmarshal([]byte(resp), &out); err != nil {
		t.Fatal(err)
	}
	if out.SessionID != "" || len(out.NeedsApproval) != 1 || out.NeedsApproval[0] != "a" {
		t.Fatalf("expected approval request, got %s", resp)
	}
	// A hash for a different script does not approve this one.
	resp = a.ApplyTweaks(`{"ids":["a"],"approve":["deadbeef"]}`)
	if err := json.Unmarshal([]byte(resp), &out); err != nil || len(out.NeedsApproval) != 1 {
		t.Fatalf("wrong hash must not approve: %s", resp)
	}
	if _, err := os.Stat(filepath.Join(a.ExeDir, "data", "state", "tweak_approvals.json")); err == nil {
		t.Fatal("nothing should be stored before approval")
	}
}

func TestApprovalIsStoredAndInvalidatedByChange(t *testing.T) {
	a := approvalAPI(t)
	h := TweakScriptHash(a.TweaksData.Tweaks[0])
	list, pending, err := a.approvedTweaks(applyRequest{IDs: []string{"a"}, Approve: []string{h}})
	if err != nil || len(pending) != 0 || len(list) != 1 || list[0].Command != "Write-Output a" {
		t.Fatalf("approve: %v %v %v", list, pending, err)
	}
	if p := preview(t, a, "a", "b"); !p[0].Approved || p[1].Approved {
		t.Fatalf("approval not stored per tweak: %+v", p)
	}
	// Re-applying the same script needs no new confirmation.
	if _, pending, err := a.approvedTweaks(applyRequest{IDs: []string{"a"}}); err != nil || len(pending) != 0 {
		t.Fatalf("stored approval ignored: %v %v", pending, err)
	}
	// An imported/edited command needs a new approval.
	a.TweaksData.Tweaks[0].Command = "Write-Output changed"
	if _, pending, _ := a.approvedTweaks(applyRequest{IDs: []string{"a"}}); len(pending) != 1 {
		t.Fatal("changed command must need approval again")
	}
}

func TestParseApplyRequest(t *testing.T) {
	req, err := parseApplyRequest(`{"ids":["x","y"],"approve":["h"]}`)
	if err != nil || len(req.IDs) != 2 || req.Approve[0] != "h" {
		t.Fatalf("new format: %+v %v", req, err)
	}
	req, err = parseApplyRequest(`[{"id":"x","command":"ignored"}]`)
	if err != nil || len(req.IDs) != 1 || req.IDs[0] != "x" || len(req.Approve) != 0 {
		t.Fatalf("legacy format: %+v %v", req, err)
	}
	if _, err := parseApplyRequest(`nope`); err == nil {
		t.Fatal("invalid JSON accepted")
	}
}
