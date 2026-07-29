//go:build !windows

package api

// Stubs for non-Windows platforms. The launcher is Windows-only, but
// these keep the package compilable on other OSes.

func runFolderDialog(title, initialDir string) (string, error) {
	return "", nil
}

func runFileDialog(title, initialDir, filter string, save bool) (string, error) {
	return "", nil
}

func runSaveDialog(title, initialDir, filter, defaultName string) (string, error) {
	return "", nil
}
