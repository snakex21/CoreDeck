// Package knowledge manages the knowledge base: a collection of
// user-curated links organized by categories. Data is loaded from
// <exeDir>/data/knowledge.json.
package knowledge

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"launcher/internal/safeio"
)

// Link is a single bookmark entry.
type Link struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Tag     string `json:"tag,omitempty"`
	Note    string `json:"note,omitempty"`
	Section string `json:"section,omitempty"`
	Starred bool   `json:"starred,omitempty"`
}

// Group is the first level of the knowledge taxonomy. Its identity and icon
// are data, so the UI does not need label-specific branching.
type Group struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon,omitempty"`
	Order int    `json:"order,omitempty"`
}

// Category is the second level of the knowledge taxonomy.
type Category struct {
	ID      string `json:"id"`
	GroupID string `json:"groupId"`
	Name    string `json:"name"`
	Icon    string `json:"icon,omitempty"`
	Order   int    `json:"order,omitempty"`
	Links   []Link `json:"links"`
}

// KnowledgeData is the on-disk format.
type KnowledgeData struct {
	Version    string     `json:"version"`
	Groups     []Group    `json:"groups"`
	Categories []Category `json:"categories"`
}

// Load reads knowledge.json from <webDir>/data/knowledge.json.
func Load(webDir string) (*KnowledgeData, error) {
	jsonPath := filepath.Join(webDir, "data", "knowledge.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read knowledge.json: %w", err)
	}
	var kd KnowledgeData
	if err := json.Unmarshal(data, &kd); err != nil {
		return nil, fmt.Errorf("cannot parse knowledge.json: %w", err)
	}
	return &kd, nil
}

func LoadFromFS(fsys fs.FS) (*KnowledgeData, error) {
	data, err := fs.ReadFile(fsys, "web/data/knowledge.json")
	if err != nil {
		return nil, fmt.Errorf("cannot read knowledge.json from FS: %w", err)
	}
	var kd KnowledgeData
	if err := json.Unmarshal(data, &kd); err != nil {
		return nil, fmt.Errorf("cannot parse knowledge.json from FS: %w", err)
	}
	return &kd, nil
}

// Save writes knowledge.json to <webDir>/data/knowledge.json.
func Save(webDir string, kd *KnowledgeData) error {
	jsonPath := filepath.Join(webDir, "data", "knowledge.json")
	data, err := json.MarshalIndent(kd, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal knowledge.json: %w", err)
	}
	return safeio.WriteFileAtomic(jsonPath, data, 0o644)
}
