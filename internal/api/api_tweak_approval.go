package api

// Script preview and explicit approval for tweaks.
//
// The tweak catalogue (data\tweaks.json) is user-supplied: it is imported,
// copied from another machine or edited by hand, and every tweak runs its
// command elevated. Before a tweak runs for the first time, or after its
// command changed, the UI must show the exact command and the user must
// approve it. Approvals are stored per tweak ID as the SHA-256 of what runs
// (type + command) in data\state\tweak_approvals.json, so an edited or
// re-imported tweak needs a new approval.
//
// ApplyTweaks never trusts commands sent by the page: it looks every tweak up
// by ID in the loaded catalogue, so the command that runs is the one that was
// previewed.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"launcher/internal/safeio"
	"launcher/internal/tweaks"
)

var tweakApprovalMu sync.Mutex

// TweakScriptHash identifies what a tweak executes.
func TweakScriptHash(t tweaks.Tweak) string {
	sum := sha256.Sum256([]byte(t.Type + "\n" + t.Command))
	return hex.EncodeToString(sum[:])
}

func (a *API) tweakApprovalPath() string {
	return filepath.Join(a.ExeDir, "data", "state", "tweak_approvals.json")
}

func (a *API) loadTweakApprovals() map[string]string {
	approvals := map[string]string{}
	data, err := os.ReadFile(a.tweakApprovalPath())
	if err != nil {
		return approvals
	}
	if err := json.Unmarshal(data, &approvals); err != nil || approvals == nil {
		return map[string]string{}
	}
	return approvals
}

func (a *API) saveTweakApprovals(approvals map[string]string) error {
	data, err := json.MarshalIndent(approvals, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(a.tweakApprovalPath()), 0o755); err != nil {
		return err
	}
	return safeio.WriteFileAtomic(a.tweakApprovalPath(), data, 0o644)
}

// TweakPreview is one entry of the confirmation dialog.
type TweakPreview struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Risk     string `json:"risk,omitempty"`
	Command  string `json:"command"`
	Hash     string `json:"hash"`
	Approved bool   `json:"approved"`
}

// catalogueTweaks returns the catalogue entries for ids, in order.
func (a *API) catalogueTweaks(ids []string) ([]tweaks.Tweak, error) {
	if a.TweaksData == nil {
		a.GetTweaksData()
	}
	if a.TweaksData == nil {
		return nil, fmt.Errorf("no tweaks data")
	}
	byID := make(map[string]tweaks.Tweak, len(a.TweaksData.Tweaks))
	for _, t := range a.TweaksData.Tweaks {
		byID[t.ID] = t
	}
	out := make([]tweaks.Tweak, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		t, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("unknown tweak %q", id)
		}
		out = append(out, t)
	}
	return out, nil
}

// PreviewTweaks returns the commands the given tweak IDs (JSON array) would
// run and whether each one was already approved in this exact form.
func (a *API) PreviewTweaks(idsJSON string) string {
	var ids []string
	if err := json.Unmarshal([]byte(idsJSON), &ids); err != nil {
		return jsonError("invalid tweak id list")
	}
	list, err := a.catalogueTweaks(ids)
	if err != nil {
		return jsonErr(err)
	}
	tweakApprovalMu.Lock()
	approvals := a.loadTweakApprovals()
	tweakApprovalMu.Unlock()

	previews := make([]TweakPreview, 0, len(list))
	for _, t := range list {
		h := TweakScriptHash(t)
		previews = append(previews, TweakPreview{
			ID: t.ID, Name: t.Name, Type: t.Type, Risk: t.Risk,
			Command: t.Command, Hash: h, Approved: approvals[t.ID] == h,
		})
	}
	data, _ := json.Marshal(map[string]any{"tweaks": previews})
	return string(data)
}

// applyRequest is what the page sends to ApplyTweaks. Older pages sent the
// tweak objects themselves; only their IDs are used.
type applyRequest struct {
	IDs     []string `json:"ids"`
	Approve []string `json:"approve"` // script hashes the user just confirmed
}

func parseApplyRequest(raw string) (applyRequest, error) {
	var req applyRequest
	if err := json.Unmarshal([]byte(raw), &req); err == nil && req.IDs != nil {
		return req, nil
	}
	var legacy []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		return req, fmt.Errorf("invalid tweaks JSON")
	}
	req = applyRequest{IDs: make([]string, 0, len(legacy))}
	for _, t := range legacy {
		req.IDs = append(req.IDs, t.ID)
	}
	return req, nil
}

// approvedTweaks resolves the request against the catalogue and checks every
// tweak is approved (stored, or confirmed in this request). Newly confirmed
// hashes are stored. It returns the IDs that still need approval, if any.
func (a *API) approvedTweaks(req applyRequest) ([]tweaks.Tweak, []string, error) {
	list, err := a.catalogueTweaks(req.IDs)
	if err != nil {
		return nil, nil, err
	}
	confirmed := make(map[string]bool, len(req.Approve))
	for _, h := range req.Approve {
		confirmed[h] = true
	}

	tweakApprovalMu.Lock()
	defer tweakApprovalMu.Unlock()
	approvals := a.loadTweakApprovals()
	var pending []string
	changed := false
	for _, t := range list {
		h := TweakScriptHash(t)
		if approvals[t.ID] == h {
			continue
		}
		if confirmed[h] {
			approvals[t.ID] = h
			changed = true
			continue
		}
		pending = append(pending, t.ID)
	}
	if len(pending) > 0 {
		return nil, pending, nil
	}
	if changed {
		if err := a.saveTweakApprovals(approvals); err != nil {
			return nil, nil, fmt.Errorf("save tweak approvals: %w", err)
		}
	}
	return list, nil, nil
}
