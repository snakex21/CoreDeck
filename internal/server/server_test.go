package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBindLoopbackNeverListensOnAllInterfaces(t *testing.T) {
	reservation, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := reservation.Addr().(*net.TCPAddr).Port
	if err := reservation.Close(); err != nil {
		t.Fatal(err)
	}

	listener, boundPort, err := bindLoopback(port, port)
	if err != nil {
		t.Fatalf("bindLoopback() error = %v", err)
	}
	defer listener.Close()

	address, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("listener address has type %T, want *net.TCPAddr", listener.Addr())
	}
	if boundPort != port || address.Port != port {
		t.Fatalf("bound port = %d / %d, want %d", boundPort, address.Port, port)
	}
	if address.IP.IsUnspecified() || !address.IP.IsLoopback() || !address.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		t.Fatalf("listener bound to %q, want exactly 127.0.0.1", address.IP)
	}
}

func TestBindLoopbackRejectsInvalidRange(t *testing.T) {
	if listener, port, err := bindLoopback(2, 1); err == nil || listener != nil || port != 0 {
		t.Fatalf("bindLoopback(2, 1) = (%v, %d, %v), want an error", listener, port, err)
	}
}

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
	if cache := resp.Header.Get("Cache-Control"); !strings.Contains(cache, "max-age=86400") {
		t.Fatalf("expected icon cache policy, got %q", cache)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "private-icon") {
		t.Fatalf("expected external program icon, got %q", body)
	}
}

func TestNoCORSAndOnlyLoopbackHosts(t *testing.T) {
	static := fstest.MapFS{"index.html": {Data: []byte("index")}}
	srv, port, err := Start(static, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown(context.Background())
	base := "http://127.0.0.1:" + fmt.Sprint(port) + "/index.html"

	for _, host := range []string{"", "localhost:" + fmt.Sprint(port), "127.0.0.1:" + fmt.Sprint(port)} {
		req, _ := http.NewRequest(http.MethodGet, base, nil)
		req.Header.Set("Origin", "https://example.com")
		if host != "" {
			req.Host = host
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("host %q: status %d, want 200", host, resp.StatusCode)
		}
		if v := resp.Header.Get("Access-Control-Allow-Origin"); v != "" {
			t.Fatalf("host %q: unexpected Access-Control-Allow-Origin %q", host, v)
		}
	}

	for _, host := range []string{"evil.example:" + fmt.Sprint(port), "localhost", "127.0.0.1:1", "[::1]:" + fmt.Sprint(port)} {
		req, _ := http.NewRequest(http.MethodGet, base, nil)
		req.Host = host
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("host %q: status %d, want 403", host, resp.StatusCode)
		}
	}
}

func TestKnowledgeBaseIsServedFromDataDirectory(t *testing.T) {
	dataDir := t.TempDir()
	userDir := filepath.Join(dataDir, "user_pages")
	kbDir := filepath.Join(dataDir, "KnowledgeBase", "dane")
	for _, d := range []string{userDir, kbDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(dataDir, "KnowledgeBase", "index.html"), []byte("<h1>private-kb</h1>"), 0o644)
	os.WriteFile(filepath.Join(kbDir, "ai-dane.js"), []byte("var x=1"), 0o644)
	os.WriteFile(filepath.Join(dataDir, "secret.json"), []byte("secret"), 0o644)

	srv, port, err := Start(fstest.MapFS{"index.html": {Data: []byte("index")}}, userDir)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown(context.Background())
	base := "http://127.0.0.1:" + fmt.Sprint(port)
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

	for path, want := range map[string]string{
		"/KnowledgeBase/index.html?returnProgram=x": "private-kb",
		"/KnowledgeBase/":                           "private-kb",
		"/KnowledgeBase/dane/ai-dane.js":            "var x=1",
	} {
		resp, err := client.Get(base + path)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), want) {
			t.Errorf("%s: status %d body %q", path, resp.StatusCode, body)
		}
	}
	for _, path := range []string{"/KnowledgeBase/../secret.json", "/KnowledgeBase/%2e%2e/secret.json", "/KnowledgeBase/missing.html", "/KnowledgeBase/dane"} {
		resp, err := http.Get(base + path) // follows the mux's path-cleaning redirect
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if string(body) == "secret" || resp.StatusCode == http.StatusOK {
			t.Errorf("%s: status %d body %q", path, resp.StatusCode, body)
		}
	}
}
