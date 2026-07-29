package safeio

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileAtomic zapisuje nową wersję przez plik tymczasowy i pozostawia
// poprzednią poprawną wersję jako <nazwa>.bak. Jest to szczególnie ważne na
// pamięciach USB, które mogą zostać odłączone w trakcie zapisu.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}
	defer cleanup()

	if err := tmp.Chmod(perm); err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	backupPath := path + ".bak"
	hadOriginal := false
	if _, err := os.Stat(path); err == nil {
		hadOriginal = true
		_ = os.Remove(backupPath)
		if err := os.Rename(path, backupPath); err != nil {
			return fmt.Errorf("cannot create backup: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		if hadOriginal {
			_ = os.Rename(backupPath, path)
		}
		return fmt.Errorf("cannot replace file: %w", err)
	}
	return nil
}
