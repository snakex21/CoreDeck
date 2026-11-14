package paths

import (
	"os"
	"path/filepath"
	"testing"
)

// layout builds <root>\_Launcher (the exe dir) plus the given directories
// relative to <root>, and returns (root, exeDir).
func layout(t *testing.T, dirs ...string) (string, string) {
	t.Helper()
	root := t.TempDir()
	exeDir := filepath.Join(root, "_Launcher")
	for _, d := range append([]string{"_Launcher"}, dirs...) {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(d)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root, exeDir
}

func TestResolveEmpty(t *testing.T) {
	if p, ok := resolve(`C:\x`, `C:\`, "   "); p != "" || ok {
		t.Fatalf("empty input: got %q, %v", p, ok)
	}
	if ResolvePath(`C:\x`, "") != "" {
		t.Fatal("ResolvePath(\"\") must be empty")
	}
}

func TestResolveAbsolute(t *testing.T) {
	root, exeDir := layout(t, "Tools")
	abs := filepath.Join(root, "Tools")
	p, ok := resolve(exeDir, root, abs+string(filepath.Separator)+".")
	if p != abs || !ok {
		t.Fatalf("absolute existing: got %q, %v", p, ok)
	}
	missing := filepath.Join(root, "nope")
	if p, ok := resolve(exeDir, root, `"`+missing+`"`); p != missing || ok {
		t.Fatalf("absolute missing (quoted): got %q, %v", p, ok)
	}
}

func TestResolvePortableLayout(t *testing.T) {
	// Typical catalogue entry: "../Centrum/App" next to <drive>\_Launcher.
	root, exeDir := layout(t, "Centrum/App")
	want := filepath.Join(root, "Centrum", "App")
	for _, raw := range []string{"../Centrum/App", `..\Centrum\App`, "Centrum/App", " ../Centrum/App "} {
		p, ok := resolve(exeDir, root, raw)
		if p != want || !ok {
			t.Errorf("%q: got %q, %v; want %q", raw, p, ok, want)
		}
	}
}

func TestResolveLiteralWinsOverDrive(t *testing.T) {
	// Both <exeDir>\App and <drive>\App exist: the literal relative meaning wins.
	root, exeDir := layout(t, "App", "_Launcher/App")
	p, ok := resolve(exeDir, root, "App")
	if want := filepath.Join(exeDir, "App"); p != want || !ok {
		t.Fatalf("got %q, %v; want %q", p, ok, want)
	}
}

func TestResolveDataDir(t *testing.T) {
	root, exeDir := layout(t, "_Launcher/data/user_pages")
	p, ok := resolve(exeDir, root, "user_pages")
	if want := filepath.Join(exeDir, "data", "user_pages"); p != want || !ok {
		t.Fatalf("got %q, %v; want %q", p, ok, want)
	}
}

func TestResolveDriveRootWhenExeIsDeeper(t *testing.T) {
	// exe in <drive>\Apps\_Launcher, catalogue still written for the
	// <drive>\_Launcher layout: "..\Centrum" must fall back to the drive root.
	root := t.TempDir()
	exeDir := filepath.Join(root, "Apps", "_Launcher")
	for _, d := range []string{exeDir, filepath.Join(root, "Centrum")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	p, ok := resolve(exeDir, root, `..\Centrum`)
	if want := filepath.Join(root, "Centrum"); p != want || !ok {
		t.Fatalf("got %q, %v; want %q", p, ok, want)
	}
}

func TestResolveParentWithoutDrive(t *testing.T) {
	// Candidate 4 (<exeDir>\..\stripped) when the drive root does not have it.
	root := t.TempDir()
	exeDir := filepath.Join(root, "Apps", "_Launcher")
	for _, d := range []string{exeDir, filepath.Join(root, "Apps", "Centrum")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	p, ok := resolve(exeDir, filepath.Join(root, "elsewhere"), `../../Centrum`)
	if want := filepath.Join(root, "Apps", "Centrum"); p != want || !ok {
		t.Fatalf("got %q, %v; want %q", p, ok, want)
	}
}

func TestResolveMissingReturnsLiteral(t *testing.T) {
	root, exeDir := layout(t)
	p, ok := resolve(exeDir, root, "../Missing/App")
	if want := filepath.Join(root, "Missing", "App"); p != want || ok {
		t.Fatalf("got %q, %v; want %q (not found)", p, ok, want)
	}
}
