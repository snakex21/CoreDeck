package main

import (
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Every user-visible text in the UI goes through i18n keys (t(), tn(),
// data-i18n*). This test looks for Polish text (diacritics are a reliable
// marker) left in the embedded pages and scripts outside js/i18n/:
//
//   - JS string and template literals, except in comments, inside console.*
//     calls and as the fallback argument of tr(key, 'fallback') wrappers;
//   - HTML text nodes not covered by a data-i18n ancestor (the Polish text may
//     stay as initial content of a data-i18n element);
//   - title / placeholder / aria-label / alt attributes without the matching
//     data-i18n-title / -placeholder / -aria / -alt.

var polishLetters = regexp.MustCompile(`[ąćęłńóśźżĄĆĘŁŃÓŚŹŻ]`)

// hardcodedTextAllowlist lists the only places where Polish literals may stay.
// file is relative to web/, contains is a substring of the literal.
var hardcodedTextAllowlist = []struct{ file, contains, reason string }{
	{"js/workers/ai-worker.js", "Masz dostępny katalog programów", "system prompt sent to the local LLM, never displayed"},
	{"js/workers/knowledge-advisor-worker.js", "Jesteś doradcą stron", "system prompt sent to the local LLM, never displayed"},
	{"js/pages/launcher.js", "../Centrum Narzędzi i Oprogramowania/", "on-disk folder paths of the setup workflow steps (data; the step texts use keys)"},
	{"js/pages/configurator/tweaks.js", "Narzędzia i inne", "debloat group id from tweaks.json (data key; label is cfg_debloat_group_tools_other)"},
}

type textHit struct {
	line int
	text string
}

func skipUIPath(p string) bool {
	return strings.HasPrefix(p, "js/i18n/") || strings.HasPrefix(p, "lmstudio/")
}

func TestHardcodedTextUsesI18nKeys(t *testing.T) {
	webFS, err := fs.Sub(staticFiles, "web")
	if err != nil {
		t.Fatal(err)
	}
	used := make([]bool, len(hardcodedTextAllowlist))
	checked := 0
	err = fs.WalkDir(webFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || skipUIPath(p) {
			return err
		}
		ext := strings.ToLower(path.Ext(p))
		if ext != ".js" && ext != ".html" {
			return nil
		}
		data, err := fs.ReadFile(webFS, p)
		if err != nil {
			return err
		}
		checked++
		src := strings.ReplaceAll(string(data), "\r\n", "\n")
		var hits []textHit
		if ext == ".js" {
			hits = scanJSLiterals(src, 0)
		} else {
			hits = scanHTMLText(src)
		}
	next:
		for _, h := range hits {
			for i, a := range hardcodedTextAllowlist {
				if a.file == p && strings.Contains(h.text, a.contains) {
					used[i] = true
					continue next
				}
			}
			t.Errorf("%s:%d hard-coded user-visible text (use an i18n key): %q", p, h.line, shorten(h.text))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked < 20 {
		t.Fatalf("only %d UI files checked; embed layout changed?", checked)
	}
	for i, a := range hardcodedTextAllowlist {
		if !used[i] {
			t.Errorf("allowlist entry %s %q no longer matches anything; remove it", a.file, a.contains)
		}
	}
}

var (
	keyCall  = regexp.MustCompile(`(?:^|[^\w.$])(t|tr|tn)\(\s*'([a-z0-9_]+)'\s*[,)]`)
	keyAttr  = regexp.MustCompile(`data-i18n(?:-placeholder|-title|-aria|-alt)?="([a-z0-9_]+)"`)
	keyCalls = 0
)

// Every literal key passed to t()/tr()/tn() or data-i18n* exists in pl and en
// (otherwise the user sees the raw key or a Polish fallback).
func TestHardcodedTextKeysAreDefined(t *testing.T) {
	all := parseTranslations(t)
	webFS, err := fs.Sub(staticFiles, "web")
	if err != nil {
		t.Fatal(err)
	}
	keyCalls = 0
	err = fs.WalkDir(webFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || skipUIPath(p) {
			return err
		}
		if ext := path.Ext(p); ext != ".js" && ext != ".html" {
			return nil
		}
		data, err := fs.ReadFile(webFS, p)
		if err != nil {
			return err
		}
		src := string(data)
		check := func(pos int, fn, key string) {
			keyCalls++
			for _, lang := range []string{"pl", "en"} {
				table := all[lang]
				if fn == "tn" {
					if _, ok := table[key+"_other"]; !ok {
						t.Errorf("%s:%d tn('%s') has no %s_other in %s", p, lineAt(src, pos), key, key, lang)
					}
					continue
				}
				if _, ok := table[key]; !ok {
					t.Errorf("%s:%d key %q missing in %s", p, lineAt(src, pos), key, lang)
				}
			}
		}
		for _, m := range keyCall.FindAllStringSubmatchIndex(src, -1) {
			check(m[0], src[m[2]:m[3]], src[m[4]:m[5]])
		}
		for _, m := range keyAttr.FindAllStringSubmatchIndex(src, -1) {
			check(m[0], "t", src[m[2]:m[3]])
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if keyCalls < 200 {
		t.Fatalf("only %d key references found; regexp broken?", keyCalls)
	}
}

// The scanner itself: a planted literal must be reported, fallbacks and
// comments must not.
func TestHardcodedTextScanner(t *testing.T) {
	js := strings.Join([]string{
		`// komentarz z ąę`,
		`/* blok ść */`,
		`console.warn('Nie udało się', x);`,
		`const a = tr('key', 'Zażółć fallback');`,
		`const r = /[ąę]/g;`,
		"const b = `<b>${t('x')}</b>`;",
		`uiToast('Błąd zapisu');`,
		"el.innerHTML = `<p>${n} plików</p>`;",
		`const c = cond ? t('a') : "Gęś";`,
	}, "\n")
	hits := scanJSLiterals(js, 0)
	var got []string
	for _, h := range hits {
		got = append(got, h.text)
	}
	want := []string{"Błąd zapisu", "<p>${…} plików</p>", "Gęś"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("JS hits = %q; want %q", got, want)
	}

	html := strings.Join([]string{
		`<!-- komentarz ąę -->`,
		`<h1 data-i18n="k">Zażółć</h1>`,
		`<p>Gęślą jaźń</p>`,
		`<div data-i18n="k2"><b>pogrubione ść</b></div>`,
		`<input placeholder="Szukaj…ść" data-i18n-placeholder="p">`,
		`<button title="Usuń">x</button>`,
		`<img alt="Obraz ł" data-i18n-alt="a">`,
		`<script>uiToast('Źle');</script>`,
		`<span aria-label="Zamknij okno ó"></span>`,
	}, "\n")
	got = nil
	for _, h := range scanHTMLText(html) {
		got = append(got, h.text)
	}
	want = []string{"Gęślą jaźń", "title=Usuń", "Źle", "aria-label=Zamknij okno ó"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("HTML hits = %q; want %q", got, want)
	}
}

func shorten(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 90 {
		return string(r[:90]) + "…"
	}
	return s
}

func lineAt(src string, pos int) int {
	return strings.Count(src[:pos], "\n") + 1
}

// ---------------------------------------------------------------- JS scanner

type jsFrame struct {
	open   byte   // '(', '[', '{' or '$' (template expression)
	callee string // for '(': the identifier chain before it, e.g. "console.log"
	arg    int    // for '(': index of the current argument
}

type jsScanner struct {
	src     string
	pos     int
	base    int // line offset (for inline <script>)
	stack   []jsFrame
	prev    string // previous significant token ("" at start)
	prevID  string // identifier chain ending at prev ("" if prev is not an identifier)
	hits    []textHit
	lastPos int
}

func scanJSLiterals(src string, lineOffset int) []textHit {
	s := &jsScanner{src: src, base: lineOffset}
	s.run(false)
	return s.hits
}

// regexAllowed: after these tokens a '/' starts a regex literal.
func (s *jsScanner) regexAllowed() bool {
	if s.prev == "" {
		return true
	}
	if s.prevID != "" {
		switch s.prev {
		case "return", "typeof", "case", "do", "else", "in", "of", "new", "delete", "void", "throw", "instanceof", "yield", "await":
			return true
		}
		return false
	}
	switch s.prev {
	case ")", "]", "}", "num", "str":
		return false
	}
	return true
}

// literal reports a string literal (or template text) if it is user-visible.
func (s *jsScanner) literal(start int, text string) {
	if !polishLetters.MatchString(text) {
		return
	}
	for i := len(s.stack) - 1; i >= 0; i-- {
		f := s.stack[i]
		if f.open == '(' && (f.callee == "console" || strings.HasPrefix(f.callee, "console.")) {
			return
		}
	}
	if n := len(s.stack); n > 0 {
		f := s.stack[n-1]
		if f.open == '(' && f.arg >= 1 && (f.callee == "tr" || strings.HasSuffix(f.callee, ".tr")) {
			return // tr(key, 'fallback')
		}
	}
	s.hits = append(s.hits, textHit{line: s.base + lineAt(s.src, start), text: text})
}

// run scans code; with inTemplateExpr it returns at the '}' closing a ${…}.
func (s *jsScanner) run(inTemplateExpr bool) {
	depth := len(s.stack)
	for s.pos < len(s.src) {
		c := s.src[s.pos]
		switch {
		case c == '/' && strings.HasPrefix(s.src[s.pos:], "//"):
			if i := strings.IndexByte(s.src[s.pos:], '\n'); i >= 0 {
				s.pos += i
			} else {
				s.pos = len(s.src)
			}
		case c == '/' && strings.HasPrefix(s.src[s.pos:], "/*"):
			if i := strings.Index(s.src[s.pos+2:], "*/"); i >= 0 {
				s.pos += i + 4
			} else {
				s.pos = len(s.src)
			}
		case c == '/' && s.regexAllowed():
			s.skipRegex()
			s.prev, s.prevID = "str", ""
		case c == '\'' || c == '"':
			start := s.pos
			text := s.readString(c)
			s.literal(start, text)
			s.prev, s.prevID = "str", ""
		case c == '`':
			start := s.pos
			text := s.readTemplate()
			s.literal(start, text)
			s.prev, s.prevID = "str", ""
		case c == '(' || c == '[' || c == '{':
			f := jsFrame{open: c}
			if c == '(' {
				f.callee = s.prevID
			}
			s.stack = append(s.stack, f)
			s.pos++
			s.prev, s.prevID = string(c), ""
		case c == ')' || c == ']' || c == '}':
			if c == '}' && inTemplateExpr && len(s.stack) == depth {
				s.pos++
				return
			}
			if len(s.stack) > 0 {
				s.stack = s.stack[:len(s.stack)-1]
			}
			s.pos++
			s.prev, s.prevID = string(c), ""
		case c == ',':
			if n := len(s.stack); n > 0 && s.stack[n-1].open == '(' {
				s.stack[n-1].arg++
			}
			s.pos++
			s.prev, s.prevID = ",", ""
		case isIdentStart(c):
			start := s.pos
			for s.pos < len(s.src) && isIdentPart(s.src[s.pos]) {
				s.pos++
			}
			word := s.src[start:s.pos]
			if s.prev == "." && s.prevID != "" {
				s.prevID += "." + word
			} else {
				s.prevID = word
			}
			s.prev = word
		case c >= '0' && c <= '9':
			for s.pos < len(s.src) && (isIdentPart(s.src[s.pos]) || s.src[s.pos] == '.') {
				s.pos++
			}
			s.prev, s.prevID = "num", ""
		case c == '.':
			s.pos++
			if s.prevID != "" {
				s.prev = "." // keep prevID to extend the chain
			} else {
				s.prev = "."
			}
		case c == '?' && strings.HasPrefix(s.src[s.pos:], "?."):
			s.pos += 2
			s.prev = "."
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			s.pos++
		default:
			s.pos++
			s.prev, s.prevID = string(c), ""
		}
	}
}

func isIdentStart(c byte) bool {
	return c == '_' || c == '$' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || c >= '0' && c <= '9'
}

func (s *jsScanner) readString(q byte) string {
	var b strings.Builder
	s.pos++
	for s.pos < len(s.src) {
		c := s.src[s.pos]
		if c == '\\' && s.pos+1 < len(s.src) {
			b.WriteByte(s.src[s.pos+1])
			s.pos += 2
			continue
		}
		s.pos++
		if c == q || c == '\n' {
			break
		}
		b.WriteByte(c)
	}
	return b.String()
}

// readTemplate returns the template text with every ${…} replaced by "${…}";
// expressions are scanned recursively (they may hold their own literals).
func (s *jsScanner) readTemplate() string {
	var b strings.Builder
	s.pos++
	for s.pos < len(s.src) {
		c := s.src[s.pos]
		switch {
		case c == '\\' && s.pos+1 < len(s.src):
			b.WriteByte(s.src[s.pos+1])
			s.pos += 2
		case c == '`':
			s.pos++
			return b.String()
		case c == '$' && s.pos+1 < len(s.src) && s.src[s.pos+1] == '{':
			s.pos += 2
			s.stack = append(s.stack, jsFrame{open: '$'})
			savedPrev, savedID := s.prev, s.prevID
			s.prev, s.prevID = "", ""
			s.run(true)
			s.stack = s.stack[:len(s.stack)-1]
			s.prev, s.prevID = savedPrev, savedID
			b.WriteString("${…}")
		default:
			b.WriteByte(c)
			s.pos++
		}
	}
	return b.String()
}

func (s *jsScanner) skipRegex() {
	s.pos++
	inClass := false
	for s.pos < len(s.src) {
		c := s.src[s.pos]
		s.pos++
		switch {
		case c == '\\':
			s.pos++
		case c == '[':
			inClass = true
		case c == ']':
			inClass = false
		case c == '/' && !inClass:
			for s.pos < len(s.src) && isIdentPart(s.src[s.pos]) {
				s.pos++ // flags
			}
			return
		case c == '\n':
			return // not a regex after all
		}
	}
}

// -------------------------------------------------------------- HTML scanner

var (
	htmlTag  = regexp.MustCompile(`(?s)^<(/?)([a-zA-Z][a-zA-Z0-9-]*)((?:[^>"']|"[^"]*"|'[^']*')*)>`)
	htmlAttr = regexp.MustCompile(`([a-zA-Z_:][-a-zA-Z0-9_:.]*)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'>]+)))?`)
)

var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true, "hr": true, "img": true,
	"input": true, "link": true, "meta": true, "source": true, "track": true, "wbr": true,
}

// Attributes that are shown to the user, and the data-i18n-* that replaces each.
var translatedAttrs = map[string]string{
	"title": "data-i18n-title", "placeholder": "data-i18n-placeholder",
	"aria-label": "data-i18n-aria", "alt": "data-i18n-alt",
}

type htmlFrame struct {
	name       string
	translated bool // this element or an ancestor has data-i18n
}

func scanHTMLText(src string) []textHit {
	var hits []textHit
	var stack []htmlFrame
	covered := func() bool { return len(stack) > 0 && stack[len(stack)-1].translated }
	pos := 0
	for pos < len(src) {
		lt := strings.IndexByte(src[pos:], '<')
		if lt < 0 {
			lt = len(src) - pos
		}
		if text := src[pos : pos+lt]; polishLetters.MatchString(text) && !covered() {
			hits = append(hits, textHit{line: lineAt(src, pos), text: strings.TrimSpace(text)})
		}
		pos += lt
		if pos >= len(src) {
			break
		}
		if strings.HasPrefix(src[pos:], "<!--") {
			end := strings.Index(src[pos:], "-->")
			if end < 0 {
				break
			}
			pos += end + 3
			continue
		}
		m := htmlTag.FindStringSubmatchIndex(src[pos:])
		if m == nil {
			pos++ // a stray '<' (e.g. "<!DOCTYPE"), skip it
			if strings.HasPrefix(src[pos:], "!") {
				if end := strings.IndexByte(src[pos:], '>'); end >= 0 {
					pos += end + 1
				}
			}
			continue
		}
		tagStart := pos
		closing := m[3] > m[2]
		name := strings.ToLower(src[pos+m[4] : pos+m[5]])
		attrText := src[pos+m[6] : pos+m[7]]
		pos += m[1]
		if closing {
			for i := len(stack) - 1; i >= 0; i-- {
				if stack[i].name == name {
					stack = stack[:i]
					break
				}
			}
			continue
		}
		attrs := map[string]string{}
		for _, a := range htmlAttr.FindAllStringSubmatch(attrText, -1) {
			attrs[strings.ToLower(a[1])] = a[2] + a[3] + a[4]
		}
		_, hasI18n := attrs["data-i18n"]
		for attr, marker := range translatedAttrs {
			v, ok := attrs[attr]
			if !ok || !polishLetters.MatchString(v) {
				continue
			}
			if _, ok := attrs[marker]; ok {
				continue
			}
			if attr == "placeholder" && hasI18n && name == "input" {
				continue // applyTranslations sets the placeholder of <input data-i18n>
			}
			hits = append(hits, textHit{line: lineAt(src, tagStart), text: attr + "=" + v})
		}
		for attr, v := range attrs {
			if strings.HasPrefix(attr, "on") {
				for _, h := range scanJSLiterals(v, lineAt(src, tagStart)-1) {
					hits = append(hits, h)
				}
			}
		}
		if name == "script" || name == "style" {
			end := strings.Index(strings.ToLower(src[pos:]), "</"+name)
			if end < 0 {
				end = len(src) - pos
			}
			if name == "script" {
				hits = append(hits, scanJSLiterals(src[pos:pos+end], lineAt(src, pos)-1)...)
			}
			pos += end
			continue
		}
		if voidElements[name] || strings.HasSuffix(strings.TrimSpace(attrText), "/") {
			continue
		}
		// <title> is replaced by applyTranslations() (document.title = t('title')).
		stack = append(stack, htmlFrame{name: name, translated: hasI18n || name == "title" || covered()})
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].line < hits[j].line })
	return hits
}
