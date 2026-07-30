package knowledge

import (
	"reflect"
	"testing"
	"testing/fstest"
)

func TestSaveAndLoadKnowledgeRoundTrip(t *testing.T) {
	root := t.TempDir()
	want := &KnowledgeData{
		Version: "1.0",
		Groups:  []Group{{ID: "docs", Name: "Documentation"}},
		Categories: []Category{{
			ID:      "go",
			GroupID: "docs",
			Name:    "Go",
			Links:   []Link{{Name: "Go", URL: "https://go.dev", Starred: true}},
		}},
	}

	if err := Save(root, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch:\ngot  %#v\nwant %#v", got, want)
	}
}

func TestLoadKnowledgeFromEmbeddedFS(t *testing.T) {
	fsys := fstest.MapFS{
		"web/data/knowledge.json": &fstest.MapFile{
			Data: []byte(`{"version":"1.0","groups":[],"categories":[]}`),
		},
	}
	got, err := LoadFromFS(fsys)
	if err != nil {
		t.Fatalf("LoadFromFS() error = %v", err)
	}
	if got.Version != "1.0" || got.Groups == nil || got.Categories == nil {
		t.Fatalf("unexpected data: %#v", got)
	}
}

func TestReviewDataDefaultsAndRoundTrip(t *testing.T) {
	root := t.TempDir()
	got, err := LoadReviewData(root)
	if err != nil {
		t.Fatalf("LoadReviewData() error = %v", err)
	}
	if got.Version != "1.0" || got.Decisions == nil {
		t.Fatalf("unexpected defaults: %#v", got)
	}

	key := ReviewKey("docs", "https://go.dev")
	got.Decisions[key] = ReviewDecision{
		Category: "docs",
		URL:      "https://go.dev",
		Action:   "keep",
	}
	if err := SaveReviewData(root, got); err != nil {
		t.Fatalf("SaveReviewData() error = %v", err)
	}
	reloaded, err := LoadReviewData(root)
	if err != nil {
		t.Fatalf("LoadReviewData() after save error = %v", err)
	}
	if !reflect.DeepEqual(reloaded, got) {
		t.Fatalf("round trip mismatch:\ngot  %#v\nwant %#v", reloaded, got)
	}
}
