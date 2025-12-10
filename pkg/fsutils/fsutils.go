package fsutils

import (
	"fmt"
	"os"
	"path/filepath"
)

func TruePath(path string) (string, error) {
	var prevAbsPath string
	var prevResolvedPath string

	changeFound := true
	for changeFound {
		changeFound = false

		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}
		if absPath != prevAbsPath {
			prevAbsPath = absPath
			changeFound = true
		}

		resolvedPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve symlinks: %w", err)
		}
		if resolvedPath != prevResolvedPath {
			prevResolvedPath = resolvedPath
			changeFound = true
		}

		path = resolvedPath
	}

	return path, nil
}

// MustRead reads a file from the given path or panics if an error occurs.
func MustRead(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return data
}
