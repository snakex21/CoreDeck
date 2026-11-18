package programs

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func touch(t *testing.T, root string, rels ...string) {
	t.Helper()
	for _, rel := range rels {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("MZ"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFindMainExecutable(t *testing.T) {
	cases := []struct {
		name       string
		files      []string
		program    string
		main       string
		candidates []string
	}{
		{"single exe", []string{"Prime95/prime95.exe"}, "Prime95", "Prime95/prime95.exe", nil},
		{"uninstaller ignored", []string{"unins000.exe", "WhoCrashedEx.exe"}, "WhoCrashed", "WhoCrashedEx.exe", nil},
		{"named like the program", []string{"CRU.exe", "reset-all.exe", "restart.exe", "restart64.exe"}, "CRU (Custom Resolution Utility)", "CRU.exe", nil},
		{"64-bit preferred", []string{"DiskInfo32.exe", "DiskInfo64.exe", "DiskInfoA64.exe"}, "CrystalDiskInfo", "DiskInfo64.exe", nil},
		{"program whose name looks like a helper", []string{"Display Driver Uninstaller.exe"}, "DDU (Display Driver Uninstaller)", "Display Driver Uninstaller.exe", nil},
		{"shallow subfolder", []string{"Samsung SSD Integrated Software Installer.exe", "Samsung Magician/SamsungMagician.exe", "Samsung Magician/SamsungMagicianSVC.exe", "Samsung Magician/DiskClean.exe"}, "Samsung Magician", "Samsung Magician/SamsungMagician.exe", nil},
		{"only installers", []string{"Everything-x64-Setup.exe", "Everything-x86-Setup.exe"}, "Everything", "", nil},
		{"only an archive", []string{"HWiNFO.zip"}, "HWiNFO", "", nil},
		{"ambiguous", []string{"SDI_x64_R2111.exe", "SDI_x64_R2201.exe"}, "Snappy Driver Installer", "", []string{"SDI_x64_R2111.exe", "SDI_x64_R2201.exe"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			touch(t, dir, c.files...)
			main, candidates := FindMainExecutable(dir, c.program)
			want := ""
			if c.main != "" {
				want = filepath.Join(dir, filepath.FromSlash(c.main))
			}
			if main != want || !reflect.DeepEqual(candidates, c.candidates) {
				t.Fatalf("got (%q, %v); want (%q, %v)", main, candidates, want, c.candidates)
			}
		})
	}
}

// stubLaunch records what LaunchProgram would start instead of starting it.
func stubLaunch(t *testing.T) *[]string {
	t.Helper()
	var started []string
	prevStart, prevOpen := startExecutable, openFolder
	startExecutable = func(p string) error { started = append(started, p); return nil }
	openFolder = func(p string) error { started = append(started, "folder:"+p); return nil }
	t.Cleanup(func() { startExecutable, openFolder = prevStart, prevOpen })
	return &started
}

// portableLayout builds <root>\_Launcher plus <root>\Centrum\<app>\files.
func portableLayout(t *testing.T, app string, files ...string) (root, exeDir string) {
	t.Helper()
	root = t.TempDir()
	exeDir = filepath.Join(root, "_Launcher")
	if err := os.MkdirAll(exeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	touch(t, filepath.Join(root, "Centrum", app), files...)
	return root, exeDir
}

func TestLaunchProgramPortableRelativePath(t *testing.T) {
	started := stubLaunch(t)
	root, exeDir := portableLayout(t, "OCCT", "OCCT.exe")
	p := Program{Name: "OCCT", InstallType: "portable", FolderPath: "../Centrum/OCCT"}

	res := LaunchProgram(exeDir, p, -1, "")
	want := filepath.Join(root, "Centrum", "OCCT", "OCCT.exe")
	if !res.Success || res.Launched != want || len(*started) != 1 || (*started)[0] != want {
		t.Fatalf("got %+v, started %v; want %q", res, *started, want)
	}
}

func TestLaunchProgramUsesConfiguredFile(t *testing.T) {
	started := stubLaunch(t)
	root, exeDir := portableLayout(t, "DeOldify", "DeOldify.NET.win.exe", "other.exe")
	p := Program{Name: "DeOldify.NET", InstallType: "portable", FolderPath: "../Centrum/DeOldify", InstallerFile: "DeOldify.NET.win.exe"}

	res := LaunchProgram(exeDir, p, -1, "")
	if want := filepath.Join(root, "Centrum", "DeOldify", "DeOldify.NET.win.exe"); !res.Success || res.Launched != want {
		t.Fatalf("got %+v, started %v", res, *started)
	}
}

func TestLaunchProgramAsksWhenAmbiguousThenRunsChoice(t *testing.T) {
	started := stubLaunch(t)
	root, exeDir := portableLayout(t, "FMOD", "FMOD Studio.exe", "fmodstudio.exe")
	p := Program{Name: "FMOD Studio", InstallType: "portable", FolderPath: "../Centrum/FMOD"}

	res := LaunchProgram(exeDir, p, -1, "")
	if !res.Success || res.Launched != "" || len(res.Candidates) != 2 || len(*started) != 0 {
		t.Fatalf("expected a choice, got %+v, started %v", res, *started)
	}
	choice := res.Candidates[1]
	res = LaunchProgram(exeDir, p, -1, choice)
	if want := filepath.Join(root, "Centrum", "FMOD", filepath.FromSlash(choice)); res.Launched != want {
		t.Fatalf("choice not launched: %+v (want %q)", res, want)
	}
}

func TestLaunchProgramRejectsFilesOutsideFolder(t *testing.T) {
	started := stubLaunch(t)
	root, exeDir := portableLayout(t, "App", "app.exe")
	touch(t, root, "evil.exe")
	p := Program{Name: "App", InstallType: "portable", FolderPath: "../Centrum/App"}

	for _, file := range []string{"../../evil.exe", filepath.Join(root, "evil.exe"), "app.txt"} {
		if res := LaunchProgram(exeDir, p, -1, file); res.Success || res.Launched != "" {
			t.Fatalf("%q must be rejected, got %+v", file, res)
		}
	}
	if len(*started) != 0 {
		t.Fatalf("nothing may start, started %v", *started)
	}
}

func TestLaunchProgramOpensFolderWhenNothingToRun(t *testing.T) {
	started := stubLaunch(t)
	root, exeDir := portableLayout(t, "HWiNFO", "hwi.zip")
	p := Program{Name: "HWiNFO", InstallType: "portable", FolderPath: "../Centrum/HWiNFO"}

	res := LaunchProgram(exeDir, p, -1, "")
	folder := filepath.Join(root, "Centrum", "HWiNFO")
	if !res.Success || !res.OpenedFolder || len(*started) != 1 || (*started)[0] != "folder:"+folder {
		t.Fatalf("got %+v, started %v", res, *started)
	}
}

func TestLaunchProgramRefusesInstallers(t *testing.T) {
	started := stubLaunch(t)
	_, exeDir := portableLayout(t, "Setup", "setup.exe")
	p := Program{Name: "Setup", InstallType: "installer", FolderPath: "../Centrum/Setup"}
	if res := LaunchProgram(exeDir, p, -1, ""); res.Success || len(*started) != 0 {
		t.Fatalf("installer must go through InstallPrograms, got %+v", res)
	}
}

func TestLaunchProgramVariant(t *testing.T) {
	started := stubLaunch(t)
	root, exeDir := portableLayout(t, "Tool", "v1/tool.exe", "v2/tool.exe")
	p := Program{Name: "Tool", InstallType: "portable", FolderPath: "../Centrum/Tool/v1", Variants: []Variant{
		{Name: "v1", FolderPath: "../Centrum/Tool/v1"},
		{Name: "v2", FolderPath: "../Centrum/Tool/v2"},
	}}
	res := LaunchProgram(exeDir, p, 1, "")
	if want := filepath.Join(root, "Centrum", "Tool", "v2", "tool.exe"); res.Launched != want {
		t.Fatalf("got %+v, started %v; want %q", res, *started, want)
	}
}
