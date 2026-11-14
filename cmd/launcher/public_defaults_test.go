package main

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
)

// The public exe and the author's private launcher are the same binary: all
// user data (programs, tweaks, drivers, settings, KnowledgeBase) is read from
// data\ next to CoreDeck.exe. The embedded web/data/*.json files are only
// empty templates. This test fails if a private catalogue is ever copied into
// the embedded defaults (and so into the public repo and release).
func TestEmbeddedDefaultsAreEmptyTemplates(t *testing.T) {
	entries, err := fs.ReadDir(staticFiles, "web/data")
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		seen++
		b, err := staticFiles.ReadFile("web/data/" + e.Name())
		if err != nil {
			t.Fatal(err)
		}
		if len(b) > 512 {
			t.Errorf("%s is %d bytes; embedded defaults must be empty templates", e.Name(), len(b))
		}
		var doc map[string]any
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Errorf("%s: %v", e.Name(), err)
			continue
		}
		for key, v := range doc {
			switch x := v.(type) {
			case []any:
				if len(x) != 0 {
					t.Errorf("%s: %q has %d entries, want 0", e.Name(), key, len(x))
				}
			case map[string]any:
				if len(x) != 0 && e.Name() != "settings.json" {
					t.Errorf("%s: %q has %d entries, want 0", e.Name(), key, len(x))
				}
			case float64:
				if key == "totalPrograms" && x != 0 {
					t.Errorf("%s: totalPrograms = %v", e.Name(), x)
				}
			}
		}
	}
	if seen != 5 {
		t.Fatalf("expected 5 embedded catalogue templates, found %d", seen)
	}
}
