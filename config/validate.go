package config

import (
	"fmt"
	"io"
	"strings"
)

// validTargetColors is the set of valid ANSI color names for target output.
//
//nolint:gochecknoglobals // package-level lookup table for color validation
var validTargetColors = map[string]bool{
	"black":         true,
	"red":           true,
	"green":         true,
	"yellow":        true,
	"blue":          true,
	"magenta":       true,
	"cyan":          true,
	"white":         true,
	"brightblack":   true,
	"brightred":     true,
	"brightgreen":   true,
	"brightyellow":  true,
	"brightblue":    true,
	"brightmagenta": true,
	"brightcyan":    true,
	"brightwhite":   true,
}

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("config: %s: %s", e.Field, e.Message)
}

// ValidationWarning represents a non-fatal configuration issue.
type ValidationWarning struct {
	Field   string
	Message string
}

func (w ValidationWarning) String() string {
	return fmt.Sprintf("config warning: %s: %s", w.Field, w.Message)
}

// ValidationResults holds the results of configuration validation.
type ValidationResults struct {
	Errors   []ValidationError
	Warnings []ValidationWarning
}

// HasErrors returns true if there are validation errors.
func (r ValidationResults) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are validation warnings.
func (r ValidationResults) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// ErrorMessage returns a combined error message for all validation errors.
func (r ValidationResults) ErrorMessage() string {
	if !r.HasErrors() {
		return ""
	}
	msgs := make([]string, 0, len(r.Errors))
	for _, e := range r.Errors {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

// WriteWarnings writes all warnings to the given writer.
func (r ValidationResults) WriteWarnings(w io.Writer) {
	for _, warn := range r.Warnings {
		_, _ = fmt.Fprintln(w, warn.String())
	}
}

// Validate checks the configuration for errors and warnings.
// It returns errors for invalid values that would cause runtime issues,
// and warnings for issues that can be safely ignored.
func (c *Config) Validate() ValidationResults {
	var result ValidationResults

	// Validate target_color
	if c.TargetColor != "" {
		normalized := strings.ToLower(c.TargetColor)
		if !validTargetColors[normalized] {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "target_color",
				Message: fmt.Sprintf("invalid color %q, must be one of: %s", c.TargetColor, validColorList()),
			})
		}
	}

	return result
}

// validColorList returns a comma-separated list of valid colors.
func validColorList() string {
	colors := []string{
		"Black", "Red", "Green", "Yellow", "Blue", "Magenta", "Cyan", "White",
		"BrightBlack", "BrightRed", "BrightGreen", "BrightYellow",
		"BrightBlue", "BrightMagenta", "BrightCyan", "BrightWhite",
	}
	return strings.Join(colors, ", ")
}
