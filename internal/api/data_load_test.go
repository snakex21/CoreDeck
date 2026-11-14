package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"launcher/internal/knowledge"
)

func embeddedKnowledge() fstest.MapFS {
	return fstest.MapFS{"web/data/knowledge.json": {Data: []byte(`{"version":"1.0","categories":[],"groups":[]}`)}}
}

func TestLoadCatalogueMissingUsesEmbeddedDefault(t *testing.T) {
	dir := t.TempDir()
	kd, err := loadCatalogue(dir, "knowledge.json", knowledge.Load, knowledge.LoadFromFS, embeddedKnowledge())
	if err != nil || kd == nil || len(kd.Categories) != 0 {
		t.Fatalf("got %+v, %v", kd, err)
	}
}

func TestLoadCatalogueExternalWins(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "data"), 0o755)
	private := `{"version":"1.0","categories":[{"id":"c","name":"Private","links":[]}],"groups":[]}`
	os.WriteFile(filepath.Join(dir, "data", "knowledge.json"), []byte(private), 0o644)
	kd, err := loadCatalogue(dir, "knowledge.json", knowledge.Load, knowledge.LoadFromFS, embeddedKnowledge())
	if err != nil || len(kd.Categories) != 1 || kd.Categories[0].Name != "Private" {
		t.Fatalf("got %+v, %v", kd, err)
	}
}

func TestLoadCatalogueCorruptFileIsNotReplacedByDefault(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "data"), 0o755)
	path := filepath.Join(dir, "data", "knowledge.json")
	os.WriteFile(path, []byte(`{"categories": [ broken`), 0o644)
	if _, err := loadCatalogue(dir, "knowledge.json", knowledge.Load, knowledge.LoadFromFS, embeddedKnowledge()); err == nil {
		t.Fatal("a corrupt data file must be an error, not the empty default")
	}

	// Through the API: an error is returned and the file is left untouched.
	a := &API{ExeDir: dir, WebDir: dir, WebFS: embeddedKnowledge()}
	var resp map[string]any
	if err := json.Unmarshal([]byte(a.GetKnowledgeData()), &resp); err != nil || resp["error"] == nil {
		t.Fatalf("expected error response, got %v (%v)", resp, err)
	}
	if a.KnowledgeData != nil {
		t.Fatal("empty default must not be loaded over a corrupt file")
	}
	if b, _ := os.ReadFile(path); !strings.Contains(string(b), "broken") {
		t.Fatal("corrupt file was modified")
	}
}
