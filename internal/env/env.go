package env

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/samber/lo"
)

func GetMap() map[string]string {
	return ToMap(os.Environ())
}

const keyValueParts = 2 // Number of parts in a key=value pair.

func ToMap(assignments []string) map[string]string {
	return lo.FromPairs(lo.FilterMap(assignments, func(item string, _ int) (lo.Entry[string, string], bool) {
		parts := strings.SplitN(item, "=", keyValueParts)
		if len(parts) != keyValueParts {
			return lo.Entry[string, string]{}, false
		}

		return lo.Entry[string, string]{Key: parts[0], Value: parts[1]}, true
	}))
}

func ToAssignments(envMap map[string]string) []string {
	return lo.MapToSlice(envMap, func(k, v string) string {
		return k + "=" + v
	})
}

// ErrInvalidBool is returned when a string cannot be parsed as a boolean.
var ErrInvalidBool = errors.New("invalid boolean value")

// ParseBool interprets a string as a boolean.
// It trims leading and trailing whitespace, then lowercases the value
// before matching.
//
// Accepted values (case-insensitive, after trimming):
//   - "true", "yes", "1"  -> true
//   - "false", "no", "0"  -> false
//   - "" (empty)          -> false, nil error
//   - any other non-empty -> false, ErrInvalidBool
func ParseBool(value string) (bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return false, nil
	}

	switch strings.ToLower(value) {
	case "true", "yes", "1":
		return true, nil
	case "false", "no", "0":
		return false, nil
	default:
		return false, fmt.Errorf("%w: %q", ErrInvalidBool, value)
	}
}

// ParseBoolEnv reads an environment variable and parses it as a boolean
// using ParseBool. Unset variables are treated the same as empty strings.
func ParseBoolEnv(envVar string) (bool, error) {
	v := os.Getenv(envVar)
	return ParseBool(v)
}

// FailsafeParseBoolEnv reads an environment variable and parses it as a boolean.
// It returns defaultValue if the variable is unset, empty, or contains an invalid
// value. This provides a fail-safe default where invalid configuration does not
// accidentally enable or disable features, depending on the chosen default.
func FailsafeParseBoolEnv(envVar string, defaultValue bool) bool {
	v, ok := os.LookupEnv(envVar)
	if !ok || v == "" {
		return defaultValue
	}

	b, err := ParseBool(v)
	if err != nil {
		return defaultValue
	}

	return b
}

