package fsutils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// LookPathIn searches for an executable in a custom PATH string.
func LookPathIn(file, pathList string) (string, error) {
	// 1. If the file contains a path separator, try to resolve it directly.
	if strings.ContainsRune(file, filepath.Separator) {
		err := findExecutable(file)
		if err != nil {
			return "", err
		}
		return file, nil
	}

	// 2. If pathList is empty, default to checking the current directory on Windows,
	// or return an error on Unix-like systems (matching Go's standard library behavior).
	if pathList == "" {
		if runtime.GOOS != "windows" {
			return "", &os.PathError{Op: "lookpath", Path: file, Err: os.ErrNotExist}
		}
		pathList = "."
	}

	// 3. Iterate through the custom path list
	for _, dir := range filepath.SplitList(pathList) {
		if dir == "" {
			// Unix treating empty path element as "."
			dir = "."
		}
		path := filepath.Join(dir, file)
		if err := findExecutable(path); err == nil {
			return path, nil
		}
	}

	return "", &os.PathError{Op: "lookpath", Path: file, Err: os.ErrNotExist}
}

func findExecutable(file string) error {
	theDescriptor, err := os.Stat(file)
	if err != nil {
		return err
	}

	// Ensure it's not a directory and has executable permissions
	if m := theDescriptor.Mode(); !m.IsDir() && m&0111 != 0 {
		return nil
	}

	// Windows handles executable extensions (.exe, .bat, etc.) automatically
	// via standard library mechanisms, but if you need strict Windows validation,
	// you can check filepath.Ext against PATHEXT. For standard Unix/Linux,
	// checking the permission bits above is sufficient.
	if runtime.GOOS == "windows" && !theDescriptor.IsDir() {
		return nil
	}

	return os.ErrPermission
}
