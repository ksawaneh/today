package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// WriteFileAtomic writes data to path by writing to a temp file in the same
// directory, fsyncing, and then renaming into place.
//
// On Unix, rename is atomic. On Windows, rename does not overwrite existing
// files; in that case we fall back to removing the destination first (not
// atomic, but best-effort).
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}

	tmpPath := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if err := tmp.Chmod(perm); err != nil {
		cleanup()
		return fmt.Errorf("chmod %s: %w", tmpPath, err)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write %s: %w", tmpPath, err)
	}

	if err := tmp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("fsync %s: %w", tmpPath, err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close %s: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		// Windows cannot rename over an existing destination.
		if runtime.GOOS == "windows" {
			if _, statErr := os.Stat(path); statErr == nil {
				if rmErr := os.Remove(path); rmErr == nil {
					if renameErr := os.Rename(tmpPath, path); renameErr == nil {
						return syncDir(dir)
					}
				}
			}
		}
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename %s -> %s: %w", tmpPath, path, err)
	}

	return syncDir(dir)
}

// BestEffortBackup tries to write a `.bak` alongside path with the current
// contents, without failing the calling operation.
func BestEffortBackup(path string, perm os.FileMode) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	_ = WriteFileAtomic(path+".bak", data, perm)
}

func syncDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer f.Close()
	_ = f.Sync()
	return nil
}
