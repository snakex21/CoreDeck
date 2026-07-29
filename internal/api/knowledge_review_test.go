package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"launcher/internal/knowledge"
)

func TestReviewKnowledgeLinkPersistsDecisionAndProtectsFeaturedRemoval(t *testing.T) {
	webDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(webDir, "data"), 0755); err != nil {
		t.Fatal(err)
	}
	data := &knowledge.KnowledgeData{Categories: []knowledge.Category{{
		Name: "Test",
		Links: []knowledge.Link{
			{Name: "Normal", URL: "https://example.com"},
			{Name: "Featured", URL: "https://featured.example.com", Starred: true},
		},
	}}}
	if err := knowledge.Save(webDir, data); err != nil {
		t.Fatal(err)
	}
	client := &API{WebDir: webDir, KnowledgeData: data}

	keep := client.ReviewKnowledgeLink(`{"action":"keep","category":"Test","url":"https://example.com"}`)
	var keepResponse struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal([]byte(keep), &keepResponse); err != nil || !keepResponse.Success {
		t.Fatalf("keep response = %s, err = %v", keep, err)
	}
	reviews, err := knowledge.LoadReviewData(webDir)
	if err != nil {
		t.Fatal(err)
	}
	if decision := reviews.Decisions[knowledge.ReviewKey("Test", "https://example.com")]; decision.Action != "keep" {
		t.Fatalf("saved action = %q, want keep", decision.Action)
	}
	stored := client.GetKnowledgeLinkReviewDecision("Test", "https://example.com")
	var storedResponse struct {
		Found    bool                     `json:"found"`
		Decision knowledge.ReviewDecision `json:"decision"`
	}
	if err := json.Unmarshal([]byte(stored), &storedResponse); err != nil || !storedResponse.Found || storedResponse.Decision.Action != "keep" {
		t.Fatalf("stored decision = %s, err = %v", stored, err)
	}

	blocked := client.ReviewKnowledgeLink(`{"action":"remove","category":"Test","url":"https://featured.example.com"}`)
	var blockedResponse struct {
		RequiresConfirmation bool `json:"requiresConfirmation"`
	}
	if err := json.Unmarshal([]byte(blocked), &blockedResponse); err != nil || !blockedResponse.RequiresConfirmation {
		t.Fatalf("unconfirmed featured removal = %s, err = %v", blocked, err)
	}
	if got := len(client.KnowledgeData.Categories[0].Links); got != 2 {
		t.Fatalf("link count after blocked removal = %d, want 2", got)
	}

	removed := client.ReviewKnowledgeLink(`{"action":"remove","category":"Test","url":"https://featured.example.com","confirmed":true}`)
	var removeResponse struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal([]byte(removed), &removeResponse); err != nil || !removeResponse.Success {
		t.Fatalf("confirmed removal response = %s, err = %v", removed, err)
	}
	if got := len(client.KnowledgeData.Categories[0].Links); got != 1 {
		t.Fatalf("link count after confirmed removal = %d, want 1", got)
	}
}
