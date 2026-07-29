package api

import (
	"path/filepath"
	"testing"
)

func TestUserPageDeletePathStaysInsidePagesDirectory(t *testing.T) {
	webDir := t.TempDir()
	got, err := userPageDeletePath(webDir, "user_pages/pages/poradnik.html")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(webDir, "data", "user_pages", "pages", "poradnik.html")
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func TestUserPageDeletePathRejectsTraversal(t *testing.T) {
	webDir := t.TempDir()
	for _, path := range []string{
		"user_pages/pages/../../programs.json",
		"pages/../settings.json",
		`user_pages/pages/C:\Windows\system.html`,
		"not-a-page.txt",
	} {
		if _, err := userPageDeletePath(webDir, path); err == nil {
			t.Fatalf("expected %q to be rejected", path)
		}
	}
}
