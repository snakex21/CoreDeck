package knowledge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReviewDecision preserves a user's review choice independently from any
// external source refresh. It is intentionally keyed by category + URL.
type ReviewDecision struct {
	Category  string `json:"category"`
	URL       string `json:"url"`
	Action    string `json:"action"` // keep, flag, remove
	UpdatedAt string `json:"updatedAt"`
}

// ReviewData is stored alongside knowledge.json at runtime.
type ReviewData struct {
	Version   string                    `json:"version"`
	Decisions map[string]ReviewDecision `json:"decisions"`
}

func ReviewKey(category, rawURL string) string {
	return category + "\n" + rawURL
}

func LoadReviewData(webDir string) (*ReviewData, error) {
	path := filepath.Join(webDir, "data", "knowledge_reviews.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &ReviewData{Version: "1.0", Decisions: map[string]ReviewDecision{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read knowledge_reviews.json: %w", err)
	}
	var reviews ReviewData
	if err := json.Unmarshal(data, &reviews); err != nil {
		return nil, fmt.Errorf("cannot parse knowledge_reviews.json: %w", err)
	}
	if reviews.Decisions == nil {
		reviews.Decisions = map[string]ReviewDecision{}
	}
	return &reviews, nil
}

func SaveReviewData(webDir string, reviews *ReviewData) error {
	path := filepath.Join(webDir, "data", "knowledge_reviews.json")
	data, err := json.MarshalIndent(reviews, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal knowledge_reviews.json: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
