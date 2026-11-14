package system

import (
	"runtime"
	"strings"
	"testing"
)

// TestPortablePathsFollowDriveLetterChange simulates the pendrive being
// re-mounted under a different letter: every default must move with the exe
// and none may keep pointing at the old drive.
func TestPortablePathsFollowDriveLetterChange(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("drive letters are Windows-only")
	}
	before := PortablePathsFor(`L:\_Launcher`)
	after := PortablePathsFor(`J:\_Launcher`)

	wantBefore := PortablePaths{
		ExeDir:   `L:\_Launcher`,
		Drive:    "L:",
		Root:     `L:\`,
		Export:   `L:\Export`,
		Import:   `L:\Import`,
		Programs: `L:\Centrum Narzędzi i Oprogramowania`,
		Drivers:  `L:\Centrum Narzędzi i Oprogramowania\Drivers`,
	}
	if before != wantBefore {
		t.Fatalf("L: layout:\n got %+v\nwant %+v", before, wantBefore)
	}

	for name, p := range map[string]string{
		"exeDir": after.ExeDir, "root": after.Root, "export": after.Export,
		"import": after.Import, "programs": after.Programs, "drivers": after.Drivers,
	} {
		if !strings.HasPrefix(p, `J:\`) {
			t.Errorf("%s = %q, want it on J: after the drive letter changed", name, p)
		}
		if strings.Contains(strings.ToUpper(p), `L:\`) {
			t.Errorf("%s = %q still references the old L: drive", name, p)
		}
	}
	if after.Drive != "J:" {
		t.Errorf("drive = %q, want J:", after.Drive)
	}
}
