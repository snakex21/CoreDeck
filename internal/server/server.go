package server

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// bindLoopback explicitly binds to 127.0.0.1 (never 0.0.0.0), so the local
// HTTP server is unreachable from LAN.
func bindLoopback(from, to int) (net.Listener, int, error) {
	var lastErr error
	for port := from; port <= to; port++ {
		listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
		if err == nil {
			return listener, port, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		return nil, 0, fmt.Errorf("invalid port range %d-%d", from, to)
	}
	return nil, 0, fmt.Errorf("all loopback ports in range %d-%d are busy: %w", from, to, lastErr)
}

func Start(staticFS fs.FS, userPagesDir string) (*http.Server, int, error) {
	fileServer := http.FileServer(http.FS(staticFS))

	listener, port, err := bindLoopback(8000, 8100)
	if err != nil {
		return nil, 0, fmt.Errorf("no free port: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", &quietHandler{fileServer: fileServer, userPagesDir: userPagesDir, port: port})

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		slog.Info("HTTP server", "port", port)
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP error", "error", err)
		}
	}()

	return srv, port, nil
}

type quietHandler struct {
	fileServer   http.Handler
	userPagesDir string
	port         int
}

// allowedHost accepts only the loopback names the WebView uses. Together with
// the absence of CORS headers (same-origin only) this keeps web pages opened
// in a regular browser, including DNS-rebinding attempts, from reading the
// UI, user pages or icons served here.
func allowedHost(host string, port int) bool {
	p := strconv.Itoa(port)
	switch strings.ToLower(host) {
	case "localhost:" + p, "127.0.0.1:" + p:
		return true
	}
	return false
}

func (h *quietHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !allowedHost(r.Host, h.port) {
		http.Error(w, "forbidden host", http.StatusForbidden)
		return
	}
	// No Access-Control-Allow-Origin: everything the UI loads (pages, scripts,
	// workers, icons, user pages in a sandboxed iframe as plain documents)
	// comes from this same origin.
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	setContentType(w, r.URL.Path)

	if r.URL.Path == "/favicon.ico" {
		// The CoreDeck brand mark is needed before the loading gate can appear.
		w.Header().Set("Cache-Control", "private, max-age=86400")
	}

	if strings.HasPrefix(r.URL.Path, "/user_pages/") && h.userPagesDir != "" {
		w.Header().Set("Content-Security-Policy", "sandbox allow-scripts allow-forms allow-modals allow-popups")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		rel := strings.TrimPrefix(r.URL.Path, "/user_pages/")
		rel = filepath.Clean(filepath.FromSlash(rel))
		if rel != "." && !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel) {
			fullPath := filepath.Join(h.userPagesDir, rel)
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				http.ServeFile(w, r, fullPath)
				return
			}
		}
	}

	// Program icons belong to the user's catalogue, not to the embedded
	// application UI. Serve them from data/icons while keeping embedded
	// UI fallbacks such as icons/default.png available below.
	if strings.HasPrefix(r.URL.Path, "/icons/") && h.userPagesDir != "" {
		// Icons change rarely and are requested in large batches. Allow WebView
		// to reuse decoded files instead of reading every icon on every render.
		w.Header().Set("Cache-Control", "private, max-age=86400")
		rel := strings.TrimPrefix(r.URL.Path, "/icons/")
		rel = filepath.Clean(filepath.FromSlash(rel))
		if rel != "." && !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel) {
			iconsDir := filepath.Join(filepath.Dir(h.userPagesDir), "icons")
			fullPath := filepath.Join(iconsDir, rel)
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				w.Header().Set("X-Content-Type-Options", "nosniff")
				http.ServeFile(w, r, fullPath)
				return
			}
		}
	}

	h.fileServer.ServeHTTP(w, r)
}

func setContentType(w http.ResponseWriter, path string) {
	switch {
	case strings.HasSuffix(path, ".wasm"):
		w.Header().Set("Content-Type", "application/wasm")
	case strings.HasSuffix(path, ".onnx"), strings.HasSuffix(path, ".bin"), strings.HasSuffix(path, ".gguf"):
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}
