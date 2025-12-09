package changelog

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yaklabco/stave/pkg/sh"
)

// ErrSVUNotFound is returned when svu binary cannot be located.
var ErrSVUNotFound = errors.New("svu not found: install via 'brew install svu' or ensure it's in PATH")

// ErrSVUEmptyVersion is returned when svu returns an empty version.
var ErrSVUEmptyVersion = errors.New("svu returned empty version")

// NextVersion returns the next semantic version from svu.
// It strips the leading 'v' prefix to match CHANGELOG heading format.
func NextVersion() (string, error) {
	svuPath, err := findSVUBinary()
	if err != nil {
		return "", err
	}

	out, err := sh.Output(svuPath, "next")
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(out)
	if version == "" {
		return "", ErrSVUEmptyVersion
	}

	// Strip leading 'v' to match CHANGELOG format
	version = strings.TrimPrefix(version, "v")
	return version, nil
}

// findSVUBinary locates svu in PATH or common Go binary locations.
func findSVUBinary() (string, error) {
	// Try PATH first
	if path, err := exec.LookPath("svu"); err == nil {
		return path, nil
	}

	// Build candidate directories from environment and go env
	candidates := []string{
		os.Getenv("GOBIN"),
		filepath.Join(os.Getenv("GOPATH"), "bin"),
	}

	// Add go env values (these may differ from environment variables)
	if goBin, err := sh.Output("go", "env", "GOBIN"); err == nil {
		candidates = append(candidates, strings.TrimSpace(goBin))
	}
	if goPath, err := sh.Output("go", "env", "GOPATH"); err == nil {
		candidates = append(candidates, filepath.Join(strings.TrimSpace(goPath), "bin"))
	}

	// Check each candidate directory for svu binary
	for _, dir := range candidates {
		if dir == "" {
			continue
		}
		svuPath := filepath.Join(dir, "svu")
		if _, err := os.Stat(svuPath); err == nil {
			return svuPath, nil
		}
	}

	return "", ErrSVUNotFound
}
