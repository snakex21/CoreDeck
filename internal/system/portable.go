package system

import "path/filepath"

// ProgramsDirName is the top-level catalogue folder of the portable layout
// (<drive>\_Launcher\CoreDeck.exe next to <drive>\Centrum Narzędzi i Oprogramowania).
const ProgramsDirName = "Centrum Narzędzi i Oprogramowania"

// PortablePaths are the default import/export locations offered by the UI.
// They are always derived from where CoreDeck.exe runs right now, so moving
// the pendrive to another drive letter (e.g. L: -> J:) never leaves defaults
// pointing at a drive that no longer exists.
type PortablePaths struct {
	ExeDir   string `json:"exeDir"`
	Drive    string `json:"drive"`
	Root     string `json:"root"`
	Export   string `json:"export"`
	Import   string `json:"import"`
	Programs string `json:"programs"`
	Drivers  string `json:"drivers"`
}

// PortablePathsFor computes PortablePaths for an executable directory.
func PortablePathsFor(exeDir string) PortablePaths {
	drive := DriveOf(exeDir)
	root := drive + string(filepath.Separator)
	programs := filepath.Join(root, ProgramsDirName)
	return PortablePaths{
		ExeDir:   filepath.Clean(exeDir),
		Drive:    drive,
		Root:     root,
		Export:   filepath.Join(root, "Export"),
		Import:   filepath.Join(root, "Import"),
		Programs: programs,
		Drivers:  filepath.Join(programs, "Drivers"),
	}
}
