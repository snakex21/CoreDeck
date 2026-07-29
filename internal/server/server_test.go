package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestUserPagesAreServedWithSandboxPolicy(t *testing.T) {
	userDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(userDir, "page.html"), []byte("<h1>test</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	static := fstest.MapFS{"index.html": {Data: []byte("index")}}
	srv, port, err := Start(static, userDir)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown(context.Background())

	resp, err := http.Get("http://127.0.0.1:" + fmt.Sprint(port) + "/user_pages/page.html")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	policy := resp.Header.Get("Content-Security-Policy")
	if !strings.Contains(policy, "sandbox") || strings.Contains(policy, "allow-same-origin") {
		t.Fatalf("unexpected user page sandbox policy: %q", policy)
	}
}

func TestProgramIconsAreServedFromDataDirectory(t *testing.T) {
	dataDir := t.TempDir()
	userDir := filepath.Join(dataDir, "user_pages")
	iconsDir := filepath.Join(dataDir, "icons")
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(iconsDir, "custom.svg"), []byte("<svg>private-icon</svg>"), 0o644); err != nil {
		t.Fatal(err)
	}

	static := fstest.MapFS{
		"index.html":        {Data: []byte("index")},
		"icons/default.png": {Data: []byte("fallback")},
	}
	srv, port, err := Start(static, userDir)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown(context.Background())

	resp, err := http.Get("http://127.0.0.1:" + fmt.Sprint(port) + "/icons/custom.svg")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "private-icon") {
		t.Fatalf("expected external program icon, got %q", body)
	}
}
