package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// User-visible backend messages reach the UI as a stable code (a key of
// translations.js) with params, next to the Polish text kept for logs and as
// a fallback (internal/uimsg; the UI translates them with tMsg / tErr). This
// test keeps it that way for the Go code the UI talks to (internal/** and
// cmd/launcher, without _test.go files):
//
//  1. a struct with a `Message string` field also has a `Code` field, and one
//     with `Detail string` a `DetailCode` field;
//  2. a composite literal that sets Message (Detail) also sets Code
//     (DetailCode), and an assignment to x.Message (x.Detail) is paired with
//     one to x.Code (x.DetailCode) in the same statement list;
//  3. every code passed to uimsg.New / uimsg.Errorf, or written as a
//     `Code:` / `DetailCode:` literal, exists in pl and en (count_* codes as
//     plural families), and the params given to uimsg.New / uimsg.Errorf
//     are exactly the key's {placeholders};
//  4. backstop: a string literal that looks Polish (diacritics or common
//     Polish words) appears only inside a uimsg.New / uimsg.Errorf call, in a
//     slog / log call, or on backendTextAllowlist with a reason.
//
// cmd/disk-helper is not scanned: it is a command-line tool for maintainers
// and its output never reaches the UI.

var polishWords = regexp.MustCompile(`\b(Nie|nie|Brak|brak|Nazwa|Wybierz|Historia|wpisu|strony|Skopiowano|Skrypt|Wygenerowano|Zastosowano|Uruchom|uruchom|Eksport zako|Import zako)\b`)

// backendTextAllowlist: Polish literals that stay. file is relative to the
// repository root; contains is a substring of the literal, or "func:Name" to
// allow every literal inside that function.
var backendTextAllowlist = []struct{ file, contains, reason string }{
	{"cmd/launcher/main.go", "Nie można wczytać plików interfejsu", "fatal start-up error in a native message box, before any UI or translation exists"},
	{"cmd/launcher/main.go", "Nie można uruchomić lokalnego serwera", "fatal start-up error (native message box)"},
	{"cmd/launcher/main.go", "Nie można otworzyć okna programu", "fatal start-up error (native message box)"},
	{"cmd/launcher/main.go", "Szczegóły zapisano", "fatal start-up error (native message box)"},
	{"cmd/launcher/fatal_windows.go", "CoreDeck — błąd", "title of the fatal start-up message box"},
	{"internal/api/api_pages.go", "Własne Strony", "default category when the UI sends none (the UI always sends t('default_pages_cat'))"},
	{"internal/api/api_pages.go", "własna strona", "catalogue tag stored in programs.json (data)"},
	{"internal/api/api_pages.go", "func:pageTemplate", "starter content of the user's own HTML page file (user-editable data, written once)"},
	{"internal/linkreview/window.go", "Kontrola linku", "toolbar injected into the separate link-review window, which has no access to translations (TODO)"},
	{"internal/system/portable.go", "Centrum Narzędzi i Oprogramowania", "on-disk folder name"},
	{"internal/tweaks/tweaks.go", "odmowa dostępu", "matches Polish Windows command output"},
	{"internal/tweaks/tweaks.go", "nie może odnaleźć", "matches Polish Windows command output"},
	{"internal/tweaks/tweaks.go", "nie moze odnalezc", "matches Polish Windows command output (without diacritics)"},
	{"internal/impex/impex.go", "func:generateReadme", "README.txt written into the export pack (file content)"},
}

// Struct types that keep a Message without a Code.
var backendStructAllowlist = map[string]string{
	"internal/models/models.go LoadResult": "unused type, never sent to the UI",
}

type goSource struct {
	rel  string
	file *ast.File
	fset *token.FileSet
}

func loadBackendSources(t *testing.T) []goSource {
	t.Helper()
	root := filepath.Join("..", "..")
	fset := token.NewFileSet()
	var out []goSource
	for _, dir := range []string{"internal", "cmd/launcher"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "web" || d.Name() == "testdata" {
					return filepath.SkipDir
				}
				if dir == "cmd/launcher" && p != filepath.Join(root, dir) {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return nil
			}
			f, err := parser.ParseFile(fset, p, nil, 0)
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(root, p)
			out = append(out, goSource{rel: filepath.ToSlash(rel), file: f, fset: fset})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(out) < 40 {
		t.Fatalf("only %d Go files scanned; layout changed?", len(out))
	}
	return out
}

// codedPairs: a field that carries a message and the field naming its code.
var codedPairs = map[string]string{"Message": "Code", "Detail": "DetailCode"}

func TestBackendMessagesHaveCodes(t *testing.T) {
	for _, src := range loadBackendSources(t) {
		pos := func(n ast.Node) string { return src.rel + ":" + strconv.Itoa(src.fset.Position(n.Pos()).Line) }
		ast.Inspect(src.file, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.TypeSpec:
				st, ok := n.Type.(*ast.StructType)
				if !ok {
					return true
				}
				fields := map[string]string{}
				for _, f := range st.Fields.List {
					for _, name := range f.Names {
						fields[name.Name] = exprString(f.Type)
					}
				}
				for msg, code := range codedPairs {
					if fields[msg] == "string" && fields[code] == "" {
						if _, ok := backendStructAllowlist[src.rel+" "+n.Name.Name]; !ok {
							t.Errorf("%s: struct %s has %s but no %s field (UI message code)", pos(n), n.Name.Name, msg, code)
						}
					}
				}
			case *ast.CompositeLit:
				keys := map[string]ast.Expr{}
				for _, e := range n.Elts {
					if kv, ok := e.(*ast.KeyValueExpr); ok {
						if id, ok := kv.Key.(*ast.Ident); ok {
							keys[id.Name] = kv.Value
						}
					}
				}
				for msg, code := range codedPairs {
					v, ok := keys[msg]
					if !ok || isEmptyString(v) {
						continue
					}
					if _, ok := keys[code]; !ok {
						t.Errorf("%s: literal sets %s without %s", pos(n), msg, code)
					}
				}
			case *ast.BlockStmt:
				checkAssignments(t, pos, n.List)
			case *ast.CaseClause:
				checkAssignments(t, pos, n.Body)
			}
			return true
		})
	}
}

// checkAssignments: x.Message = ... needs x.Code = ... in the same list.
func checkAssignments(t *testing.T, pos func(ast.Node) string, list []ast.Stmt) {
	assigned := map[string]bool{}
	type pending struct {
		node ast.Node
		want string
	}
	var need []pending
	for _, stmt := range list {
		as, ok := stmt.(*ast.AssignStmt)
		if !ok {
			continue
		}
		for _, lhs := range as.Lhs {
			sel, ok := lhs.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			target := exprString(sel.X) + "." + sel.Sel.Name
			assigned[target] = true
			if code, ok := codedPairs[sel.Sel.Name]; ok {
				need = append(need, pending{as, exprString(sel.X) + "." + code})
			}
		}
	}
	for _, p := range need {
		if !assigned[p.want] {
			t.Errorf("%s: message assigned without %s", pos(p.node), p.want)
		}
	}
}

func TestBackendMessageCodesAreDefined(t *testing.T) {
	all := parseTranslations(t)
	codes := 0
	check := func(where, code string, params []string, checkParams bool) {
		codes++
		for _, lang := range []string{"pl", "en"} {
			key := code
			if strings.HasPrefix(code, "count_") {
				key = code + "_other"
			}
			text, ok := all[lang][key]
			if !ok {
				t.Errorf("%s: code %q has no key %q in %s", where, code, key, lang)
				continue
			}
			if !checkParams {
				continue
			}
			sort.Strings(params)
			var want []string
			for _, p := range params {
				want = append(want, "{"+p+"}")
			}
			if got := placeholders(text); got != strings.Join(want, ",") {
				t.Errorf("%s: code %q gets params %v but %s text has %q", where, code, params, lang, got)
			}
		}
	}
	for _, src := range loadBackendSources(t) {
		pos := func(n ast.Node) string { return src.rel + ":" + strconv.Itoa(src.fset.Position(n.Pos()).Line) }
		ast.Inspect(src.file, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				if !isUimsgCall(n) || len(n.Args) < 2 {
					return true
				}
				code, ok := stringLit(n.Args[0])
				if !ok {
					return true // a variable code is checked where its literal is written
				}
				var params []string
				literalParams := true
				for i := 2; i < len(n.Args); i += 2 {
					name, ok := stringLit(n.Args[i])
					if !ok {
						literalParams = false
						break
					}
					params = append(params, name)
				}
				check(pos(n), code, params, literalParams)
			case *ast.KeyValueExpr:
				id, ok := n.Key.(*ast.Ident)
				if !ok || (id.Name != "Code" && id.Name != "DetailCode") {
					return true
				}
				if code, ok := stringLit(n.Value); ok && code != "" {
					check(pos(n), code, nil, false)
				}
			}
			return true
		})
	}
	if codes < 100 {
		t.Fatalf("only %d message codes found; scanner broken?", codes)
	}
}

func TestBackendHasNoUncodedPolishText(t *testing.T) {
	used := make([]bool, len(backendTextAllowlist))
	for _, src := range loadBackendSources(t) {
		tags := map[token.Pos]bool{}
		ast.Inspect(src.file, func(n ast.Node) bool {
			if f, ok := n.(*ast.Field); ok && f.Tag != nil {
				tags[f.Tag.Pos()] = true
			}
			return true
		})
		for _, decl := range src.file.Decls {
			funcName := ""
			if fd, ok := decl.(*ast.FuncDecl); ok {
				funcName = fd.Name.Name
			}
			var stack []ast.Node
			ast.Inspect(decl, func(n ast.Node) bool {
				if n == nil {
					stack = stack[:len(stack)-1]
					return true
				}
				stack = append(stack, n)
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING || tags[lit.Pos()] {
					return true
				}
				text, err := strconv.Unquote(lit.Value)
				if err != nil || len([]rune(text)) <= 2 || !(polishLetters.MatchString(text) || polishWords.MatchString(text)) {
					return true // short literals: transliteration tables ("ą" -> "a")
				}
				for _, outer := range stack {
					if call, ok := outer.(*ast.CallExpr); ok && (isUimsgCall(call) || isLogCall(call)) {
						return true
					}
				}
				for i, a := range backendTextAllowlist {
					if a.file != src.rel {
						continue
					}
					if strings.Contains(text, a.contains) || (funcName != "" && a.contains == "func:"+funcName) {
						used[i] = true
						return true
					}
				}
				t.Errorf("%s:%d Polish text without a UI message code (use uimsg.New): %q",
					src.rel, src.fset.Position(lit.Pos()).Line, shorten(text))
				return true
			})
		}
	}
	for i, a := range backendTextAllowlist {
		if !used[i] {
			t.Errorf("allowlist entry %s %q no longer matches anything; remove it", a.file, a.contains)
		}
	}
}

// The scanner itself on a small sample.
func TestBackendTextScannerSample(t *testing.T) {
	src := `package x
func f() {
	_ = uimsg.New("msg_a", "Błąd: "+e, "error", e)
	slog.Warn("nie udało się", "k", v)
	_ = "Zażółć"
	_ = "ok"
	_ = "ą"
}`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var flagged []string
	var stack []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		stack = append(stack, n)
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		text, _ := strconv.Unquote(lit.Value)
		if len([]rune(text)) <= 2 || !(polishLetters.MatchString(text) || polishWords.MatchString(text)) {
			return true
		}
		for _, outer := range stack {
			if call, ok := outer.(*ast.CallExpr); ok && (isUimsgCall(call) || isLogCall(call)) {
				return true
			}
		}
		flagged = append(flagged, text)
		return true
	})
	if strings.Join(flagged, "|") != "Zażółć" {
		t.Errorf("flagged %q; want only Zażółć", flagged)
	}
}

func isUimsgCall(call *ast.CallExpr) bool {
	name := exprString(call.Fun)
	return name == "uimsg.New" || name == "uimsg.Errorf"
}

func isLogCall(call *ast.CallExpr) bool {
	name := exprString(call.Fun)
	return strings.HasPrefix(name, "slog.") || strings.HasPrefix(name, "log.") || strings.HasPrefix(name, "a.logger.")
}

func stringLit(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	return s, err == nil
}

func isEmptyString(e ast.Expr) bool {
	s, ok := stringLit(e)
	return ok && s == ""
}

func exprString(e ast.Expr) string {
	switch e := e.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + exprString(e.X)
	case *ast.IndexExpr:
		return exprString(e.X) + "[]"
	case *ast.CallExpr:
		return exprString(e.Fun) + "()"
	}
	return ""
}
