package main

// Portable WebView2 profile.
//
// webview_go hard-codes the WebView2 profile to %APPDATA%\<exe name>, which
// leaves the theme, language and other local-storage settings on the PC
// instead of on the portable drive. The patched copy in third_party/webview_go
// reads WEBVIEW_GO_USER_DATA_FOLDER instead; CoreDeck points it at
// data\webview next to the exe. An existing profile from %APPDATA% is copied
// (never moved) the first time, without the caches; the old one stays put.

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const webviewProfileEnv = "WEBVIEW_GO_USER_DATA_FOLDER"

// webviewProfileSkip lists directories that are pure caches or telemetry;
// WebView2 recreates them and they would only bloat the portable drive.
var webviewProfileSkip = map[string]bool{
	"cache": true, "code cache": true, "gpucache": true, "grshadercache": true,
	"shadercache": true, "gpupersistentcache": true, "dawngraphitecache": true,
	"dawnwebgpucache": true, "component_crx_cache": true, "extensions_crx_cache": true,
	"browsermetrics": true, "crashpad": true, "subresource filter": true,
	"speech recognition": true, "hyphen-data": true, "widevinecdm": true,
	"optimization_guide_model_store": true,
}

// setupWebviewProfile selects the WebView2 profile folder and migrates the
// legacy one. It returns the folder in use.
func setupWebviewProfile(profileDir, legacyRoot string) string {
	if v := strings.TrimSpace(os.Getenv(webviewProfileEnv)); v != "" {
		slog.Info("webview profile from environment", "path", v)
		return v
	}
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		slog.Warn("webview profile: cannot create folder, using the default", "path", profileDir, "error", err)
		return ""
	}
	if copied, err := migrateWebviewProfile(legacyRoot, profileDir); err != nil {
		slog.Warn("webview profile migration failed (starting with a fresh profile)", "from", legacyRoot, "error", err)
	} else if copied > 0 {
		slog.Info("webview profile copied next to the exe", "from", legacyRoot, "to", profileDir, "files", copied)
	}
	if err := os.Setenv(webviewProfileEnv, profileDir); err != nil {
		slog.Warn("webview profile: cannot set "+webviewProfileEnv, "error", err)
		return ""
	}
	return profileDir
}

// legacyWebviewRoot is where unpatched webview_go keeps the profile.
func legacyWebviewRoot() string {
	appData := os.Getenv("APPDATA")
	exe, err := os.Executable()
	if appData == "" || err != nil {
		return ""
	}
	return filepath.Join(appData, filepath.Base(exe))
}

// migrateWebviewProfile copies <legacyRoot>\EBWebView to
// <profileDir>\EBWebView when the portable profile does not exist yet. The
// copy goes to a temporary folder first and is renamed into place only when
// complete, so an interrupted copy is retried on the next start. Files that
// cannot be read (locked by a running WebView2) are skipped.
func migrateWebviewProfile(legacyRoot, profileDir string) (int, error) {
	if legacyRoot == "" {
		return 0, nil
	}
	dst := filepath.Join(profileDir, "EBWebView")
	if _, err := os.Stat(dst); err == nil {
		return 0, nil // portable profile already exists
	}
	src := filepath.Join(legacyRoot, "EBWebView")
	if info, err := os.Stat(src); err != nil || !info.IsDir() {
		return 0, nil // nothing to migrate
	}
	if sameDir(src, dst) {
		return 0, nil
	}

	tmp := dst + ".migrating"
	if err := os.RemoveAll(tmp); err != nil {
		return 0, err
	}
	copied := 0
	walkErr := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == src {
				return err
			}
			slog.Debug("webview profile: skip unreadable", "path", path, "error", err)
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != src && webviewProfileSkip[strings.ToLower(d.Name())] {
				return fs.SkipDir
			}
			return os.MkdirAll(filepath.Join(tmp, rel), 0o755)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if err := copyFile(path, filepath.Join(tmp, rel)); err != nil {
			slog.Debug("webview profile: skip file", "path", path, "error", err)
			return nil
		}
		copied++
		return nil
	})
	if walkErr != nil {
		_ = os.RemoveAll(tmp)
		return 0, walkErr
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.RemoveAll(tmp)
		return 0, fmt.Errorf("activate copied profile: %w", err)
	}
	note := fmt.Sprintf("Copied from %s by CoreDeck. The original was left in place and can be deleted once everything works.\r\n", src)
	_ = os.WriteFile(filepath.Join(profileDir, "MIGRATED_FROM_APPDATA.txt"), []byte(note), 0o644)
	return copied, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		_ = os.Remove(dst)
		return err
	}
	return out.Close()
}

func sameDir(a, b string) bool {
	aa, err1 := filepath.Abs(a)
	bb, err2 := filepath.Abs(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return strings.EqualFold(filepath.Clean(aa), filepath.Clean(bb))
}
