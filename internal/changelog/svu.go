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

	out, err := sh.Output(svuPath, "next", "--force-patch-increment")
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

// findSVUBinary locates svu in PATH or GOBIN.
func findSVUBinary() (string, error) {
	// Try PATH first
	if path, err := exec.LookPath("svu"); err == nil {
		return path, nil
	}

	// Try GOBIN
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		svuPath := filepath.Join(gobin, "svu")
		if _, err := os.Stat(svuPath); err == nil {
			return svuPath, nil
		}
	}

	// Try GOPATH/bin
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		svuPath := filepath.Join(gopath, "bin", "svu")
		if _, err := os.Stat(svuPath); err == nil {
			return svuPath, nil
		}
	}

	// Try go env GOBIN
	if goBin, err := sh.Output("go", "env", "GOBIN"); err == nil {
		goBin = strings.TrimSpace(goBin)
		if goBin != "" {
			svuPath := filepath.Join(goBin, "svu")
			if _, err := os.Stat(svuPath); err == nil {
				return svuPath, nil
			}
		}
	}

	// Try go env GOPATH
	if gopath, err := sh.Output("go", "env", "GOPATH"); err == nil {
		gopath = strings.TrimSpace(gopath)
		if gopath != "" {
			svuPath := filepath.Join(gopath, "bin", "svu")
			if _, err := os.Stat(svuPath); err == nil {
				return svuPath, nil
			}
		}
	}

	return "", ErrSVUNotFound
}
