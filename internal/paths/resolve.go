// Package paths holds the single resolver that turns a path stored in a
// catalogue (programs.json, drivers.json, a workflow step) into an absolute
// path on the drive CoreDeck runs from.
package paths

import (
	"os"
	"path/filepath"
	"strings"

	"launcher/internal/system"
)

// ResolvePath maps a catalogue path to an absolute, cleaned path.
//
// Absolute paths are returned cleaned. Relative paths are tried against these
// candidates, and the first one that exists wins:
//
//  1. <exeDir>\<raw>               the literal relative meaning
//  2. <exeDir>\data\<raw>          catalogue assets kept in data\
//  3. <drive>\<raw without ..\>    portable layout <drive>\_Launcher\CoreDeck.exe
//  4. <exeDir>\..\<raw without ..\>
//  5. <exeDir>\<raw without ..\>
//
// When none exists, candidate 1 is returned so callers can show what was
// tried; they must still check that the path exists before using it. An
// empty input returns "".
func ResolvePath(exeDir, raw string) string {
	p, _ := resolve(exeDir, system.DriveOf(exeDir)+string(filepath.Separator), raw)
	return p
}

func resolve(exeDir, driveRoot, raw string) (string, bool) {
	raw = strings.Trim(strings.TrimSpace(raw), `"`)
	if raw == "" {
		return "", false
	}
	rel := filepath.FromSlash(strings.ReplaceAll(raw, `\`, "/"))
	if filepath.IsAbs(rel) {
		p := filepath.Clean(rel)
		return p, exists(p)
	}

	stripped := rel
	parent := ".." + string(filepath.Separator)
	for strings.HasPrefix(stripped, parent) {
		stripped = stripped[len(parent):]
	}

	candidates := []string{
		filepath.Join(exeDir, rel),
		filepath.Join(exeDir, "data", rel),
		filepath.Join(driveRoot, stripped),
		filepath.Join(exeDir, "..", stripped),
		filepath.Join(exeDir, stripped),
	}
	for _, c := range candidates {
		if exists(c) {
			return c, true
		}
	}
	return candidates[0], false
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
