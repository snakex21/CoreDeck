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
