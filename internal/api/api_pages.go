package api

import (
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"launcher/internal/uimsg"
)

// PageRequest jest wysyłany z Manage → "Własne Strony".
type PageRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Icon        string `json:"icon"`
}

// CreatePage tworzy nową stronę HTML w <webDir>/data/user_pages/pages/ oraz wpis programu
// typu "link" wskazujący na ten plik. Strona pokazuje się w launcherze jak
// zwykły program — kliknięcie otwiera wygenerowany HTML w webview.
//
// Dzięki temu użytkownik może dodawać dowolne własne strony (poradniki,
// bazy wiedzy, listy, instrukcje) bez limitu wyglądu — plik HTML jest w pełni
// edytowalny ręcznie (Notatnik, VS Code) lub przez AI (asystenta kodu).
func (a *API) CreatePage(requestJSON string) string {
	var req PageRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return `{"error": "Invalid request JSON"}`
	}
	if strings.TrimSpace(req.Name) == "" {
		return jsonMsg(uimsg.New("msg_page_name_required", "Nazwa strony jest wymagana"))
	}
	if req.Category == "" {
		req.Category = "Własne Strony"
	}
	if req.Icon == "" {
		req.Icon = "icons/page.svg"
	}

	slug := slugify(req.Name)
	if slug == "" {
		slug = fmt.Sprintf("strona-%d", time.Now().Unix())
	}

	pagesDir := filepath.Join(a.WebDir, "data", "user_pages", "pages")
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		return jsonMsg(uimsg.New("msg_page_mkdir_failed", "Nie można utworzyć folderu pages: "+err.Error(), "error", err.Error()))
	}

	fileName := slug + ".html"
	// Unikaj nadpisania istniejącego pliku — dodaj sufiks jeśli trzeba.
	fullPath := filepath.Join(pagesDir, fileName)
	if _, err := os.Stat(fullPath); err == nil {
		fileName = fmt.Sprintf("%s-%d.html", slug, time.Now().Unix())
		fullPath = filepath.Join(pagesDir, fileName)
	}

	html := pageTemplate(req.Name, req.Description)
	if err := os.WriteFile(fullPath, []byte(html), 0644); err != nil {
		return jsonMsg(uimsg.New("msg_page_write_failed", "Nie można zapisać pliku HTML: "+err.Error(), "error", err.Error()))
	}

	// Zbuduj wpis programu type:link.
	prog := map[string]any{
		"name":        req.Name,
		"description": req.Description,
		"category":    req.Category,
		"icon":        req.Icon,
		"folderPath":  "",
		"tags":        []string{"własna strona", "html"},
		"type":        "link",
		"localPath":   "user_pages/pages/" + fileName,
		"installType": "other",
		"selected":    false,
	}
	progJSON, _ := json.Marshal(prog)
	res := a.AddProgram(string(progJSON))
	var addResp map[string]any
	if err := json.Unmarshal([]byte(res), &addResp); err == nil {
		if msg, ok := addResp["error"].(string); ok && msg != "" {
			_ = os.Remove(fullPath)
			return jsonMsg(uimsg.New("msg_page_add_entry_failed", "Nie można dodać wpisu programu: "+msg, "error", msg))
		}
	}

	slog.Info("create_page", "name", req.Name, "file", fileName, "add_result", res)
	return `{"success": true, "fileName": "user_pages/pages/` + fileName + `"}`
}

// DeletePage usuwa plik HTML strony oraz powiązany wpis programu.
func (a *API) DeletePage(requestJSON string) string {
	var payload struct {
		Name      string `json:"name"`
		LocalPath string `json:"localPath"`
	}
	if err := json.Unmarshal([]byte(requestJSON), &payload); err != nil {
		return `{"error": "Invalid request JSON"}`
	}
	if payload.Name == "" {
		return jsonMsg(uimsg.New("msg_page_name_missing", "Brak nazwy strony"))
	}

	// Usuń plik HTML jeśli wskazano lokalPath.
	if payload.LocalPath != "" {
		htmlPath, err := userPageDeletePath(a.WebDir, payload.LocalPath)
		if err != nil {
			return jsonErr(err)
		}
		if err := os.Remove(htmlPath); err != nil && !os.IsNotExist(err) {
			slog.Warn("delete_page html", "path", htmlPath, "error", err)
		}
	}

	// Usuń wpis programu.
	res := a.DeleteProgram(payload.Name)
	slog.Info("delete_page", "name", payload.Name, "del_result", res)
	return `{"success": true}`
}

var errBadPagePath = uimsg.Errorf("msg_page_bad_path", "Nieprawidłowa ścieżka strony")

func userPageDeletePath(webDir, localPath string) (string, error) {
	clean := strings.TrimPrefix(localPath, "user_pages/pages/")
	clean = strings.TrimPrefix(clean, "pages/") // legacy support
	if filepath.IsAbs(clean) || filepath.VolumeName(clean) != "" {
		return "", errBadPagePath
	}
	pagesDir := filepath.Join(webDir, "data", "user_pages", "pages")
	htmlPath := filepath.Join(pagesDir, clean)
	rel, err := filepath.Rel(pagesDir, htmlPath)
	if err != nil || rel == "." || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", errBadPagePath
	}
	if !strings.EqualFold(filepath.Ext(htmlPath), ".html") {
		return "", uimsg.Errorf("msg_page_only_html", "Można usuwać tylko pliki HTML własnych stron")
	}
	return htmlPath, nil
}

// slugify zamienia nazwę czytelną dla człowieka na bezpieczną nazwę pliku.
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	// polskie znaki → ASCII
	repl := strings.NewReplacer(
		"ą", "a", "ć", "c", "ę", "e", "ł", "l", "ń", "n",
		"ó", "o", "ś", "s", "ż", "z", "ź", "z",
	)
	s = repl.Replace(s)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// pageTemplate zwraca startowy HTML nowej strony. Utrzymuje spójność
// wyglądu z resztą launchera (zmienne CSS z css/app.css), ale jest w pełni
// niezależny — użytkownik może go dowolnie zmienić bez ograniczeń.
func pageTemplate(title, description string) string {
	if description == "" {
		description = "Opis strony — edytuj ten plik HTML, żeby dodać treść."
	}
	safeTitle := html.EscapeString(title)
	safeDescription := html.EscapeString(description)
	return `<!DOCTYPE html>
<html lang="pl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + safeTitle + `</title>
    <link rel="stylesheet" href="/css/app.css">
    <style>
        body {
            background: var(--bg-primary);
            background-image:
                radial-gradient(circle at 20% 50%, rgba(29, 185, 84, 0.1) 0%, transparent 50%),
                radial-gradient(circle at 80% 80%, rgba(77, 159, 255, 0.1) 0%, transparent 50%);
            background-attachment: fixed;
            color: var(--text-primary);
            font-family: inherit;
            line-height: 1.7;
        }
        .page-container {
            margin-top: var(--header-height);
            max-width: 1000px;
            margin-left: auto;
            margin-right: auto;
            padding: 32px;
        }
        .page-hero {
            text-align: center;
            padding: 40px 20px;
            margin-bottom: 32px;
            background: linear-gradient(135deg, rgba(29, 185, 84, 0.08) 0%, rgba(77, 159, 255, 0.08) 100%);
            border-radius: var(--border-radius-xl);
            border: 1px solid var(--border-color);
        }
        .page-hero h1 {
            font-size: clamp(1.8rem, 4vw, 2.6rem);
            font-weight: 800;
            background: var(--gradient-primary);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            margin: 0 0 8px;
        }
        .page-hero p {
            color: var(--text-muted);
            font-size: 1rem;
            margin: 0;
        }
        .page-section {
            background: var(--bg-elevated);
            border: 1px solid var(--border-color);
            border-radius: var(--border-radius-lg);
            padding: 24px;
            margin-bottom: 20px;
        }
        .page-section h2 {
            color: var(--accent-primary);
            font-size: 1.2rem;
            margin: 0 0 12px;
        }
        .page-section p {
            color: var(--text-secondary);
            margin: 0 0 8px;
        }
        /* Tu dodawaj własne sekcje, tabele, listy, obrazy itd. — brak limitów. */
    </style>
</head>
<body>
    <header class="subpage-header">
        <div class="header-brand">
            <h1>` + safeTitle + `</h1>
            <span class="header-subtitle">Własna strona</span>
        </div>
        <div class="header-controls">
            <a href="/index.html" class="subpage-back-btn">← Wróć do launchera</a>
        </div>
    </header>

    <div class="page-container">
        <div class="page-hero">
            <h1>` + safeTitle + `</h1>
            <p>` + safeDescription + `</p>
        </div>

        <div class="page-section">
            <h2>Nowa sekcja</h2>
            <p>` + safeDescription + `</p>
            <p>Edytuj ten plik HTML, aby dodać własną treść. Możesz dodawać sekcje, tabele, listy, obrazy, linki — wszystko. Brak ograniczeń.</p>
        </div>

        <!-- Kopiuj .page-section, żeby dodać kolejne sekcje. -->
    </div>
</body>
</html>
`
}
