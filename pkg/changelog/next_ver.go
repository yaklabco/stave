package changelog

import (
	"errors"
	"fmt"
	"strings"

	"github.com/caarlos0/svu/v3/pkg/svu"
)

// ErrEmptyVersion is returned when tools returns an empty version.
var ErrEmptyVersion = errors.New("tool returned empty version")

// NextVersion returns the next semantic version.
// It strips the leading 'v' prefix to match CHANGELOG heading format.
func NextVersion() (string, error) {
	out, err := svu.Next(svu.Always())
	if err != nil {
		return "", fmt.Errorf("svu.Next: %w", err)
	}

	version := strings.TrimSpace(out)
	if version == "" {
		return "", fmt.Errorf("svu.Next: %w", ErrEmptyVersion)
	}

	// Strip leading 'v' to match CHANGELOG format
	version = strings.TrimPrefix(version, "v")

	return version, nil
}
