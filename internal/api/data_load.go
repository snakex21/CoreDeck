package api

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// loadCatalogue loads a catalogue from data\<name> next to the exe (the
// user's programs, tweaks, drivers, settings and KnowledgeBase live there).
// Only when that file does not exist is the embedded default used; the
// embedded defaults are empty templates, so the public build carries no
// private data. A file that exists but cannot be read or parsed is an error:
// silently falling back to the empty default would let the next save
// overwrite the user's data.
func loadCatalogue[T any](
	webDir, name string,
	external func(string) (*T, error),
	embedded func(fs.FS) (*T, error),
	fsys fs.FS,
) (*T, error) {
	path := filepath.Join(webDir, "data", name)
	_, statErr := os.Stat(path)
	if statErr == nil {
		v, err := external(webDir)
		if err != nil {
			return nil, fmt.Errorf("%s exists but cannot be loaded (fix or restore it; the empty default is not used to avoid overwriting it): %w", path, err)
		}
		return v, nil
	}
	if !errors.Is(statErr, fs.ErrNotExist) {
		return nil, fmt.Errorf("cannot access %s: %w", path, statErr)
	}
	if fsys == nil {
		return nil, fmt.Errorf("%s not found and no embedded default", path)
	}
	return embedded(fsys)
}
