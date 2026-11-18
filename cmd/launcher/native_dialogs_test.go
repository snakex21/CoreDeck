package main

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// nativeDialog matches alert(, confirm(, prompt( and window.alert( etc. but
// not names that merely end in them (uiConfirm(, showPrompt(, obj.confirm().
var nativeDialog = regexp.MustCompile(`(^|[^\w.$])(window\.)?(alert|confirm|prompt)\s*\(`)

// In WebView2 the browser's own alert/confirm/prompt show a "Message from
// localhost:8000" box. The UI must use js/shared/dialogs.js instead.
func TestUIHasNoNativeDialogs(t *testing.T) {
	webFS, err := fs.Sub(staticFiles, "web")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	err = fs.WalkDir(webFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".js" && ext != ".html" {
			return nil
		}
		if path == "js/shared/dialogs.js" {
			return nil // defines the replacements and mentions the originals
		}
		data, err := fs.ReadFile(webFS, path)
		if err != nil {
			return err
		}
		checked++
		for i, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") {
				continue
			}
			if m := nativeDialog.FindString(line); m != "" {
				t.Errorf("%s:%d uses a native browser dialog (%s); use uiAlert/uiConfirm/uiPrompt/uiToast: %s",
					path, i+1, strings.TrimSpace(m), trimmed)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked < 10 {
		t.Fatalf("only %d UI files checked; embed layout changed?", checked)
	}
}

// Every page that can show a dialog or toast must load the shared helpers
// before its own scripts.
func TestPagesLoadDialogHelpers(t *testing.T) {
	for _, page := range []string{"index.html", "configurator.html", "manage.html", "knowledge.html", "drivers.html", "settings.html"} {
		data, err := staticFiles.ReadFile("web/" + page)
		if err != nil {
			t.Fatal(err)
		}
		html := string(data)
		helper := strings.Index(html, `src="js/shared/dialogs.js"`)
		if helper < 0 {
			t.Errorf("%s does not load js/shared/dialogs.js", page)
			continue
		}
		if first := strings.Index(html, `src="js/pages/`); first >= 0 && first < helper {
			t.Errorf("%s loads page scripts before js/shared/dialogs.js", page)
		}
	}
}
