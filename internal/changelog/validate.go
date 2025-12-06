package changelog

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError represents a single validation error with line context.
type ValidationError struct {
	Line    int // 1-indexed line number (0 if not applicable)
	Message string
}

// ValidationWarning represents a non-fatal validation issue.
type ValidationWarning struct {
	Line    int
	Message string
}

// ValidationResult holds the outcomes of changelog validation.
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationWarning
}

// HasErrors returns true if there are any validation errors.
func (r ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are any validation warnings.
func (r ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// Error returns a combined error message or nil if no errors.
func (r ValidationResult) Error() error {
	if !r.HasErrors() {
		return nil
	}

	var msgs []string
	for _, e := range r.Errors {
		if e.Line > 0 {
			msgs = append(msgs, fmt.Sprintf("line %d: %s", e.Line, e.Message))
		} else {
			msgs = append(msgs, e.Message)
		}
	}
	return errors.New(strings.Join(msgs, "; "))
}

// Validate checks the changelog against Keep a Changelog conventions.
// Returns a ValidationResult containing any errors and warnings found.
func (c *Changelog) Validate() ValidationResult {
	result := ValidationResult{
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationWarning, 0),
	}

	// Rule 1: Title must be "Changelog"
	if c.Title != ChangelogTitle {
		result.Errors = append(result.Errors, ValidationError{
			Message: "changelog must have '# Changelog' title",
		})
	}

	// Build sets for cross-referencing
	headingNames := make(map[string]int) // name -> line number
	linkNames := make(map[string]int)    // name -> line number

	for _, heading := range c.Headings {
		// Normalize to lowercase for comparison
		key := strings.ToLower(heading.Name)
		headingNames[key] = heading.Line

		// Rule 2: Release versions must have dates
		if heading.IsRelease && heading.Date == "" {
			result.Errors = append(result.Errors, ValidationError{
				Line:    heading.Line,
				Message: fmt.Sprintf("release '%s' must include a date 'YYYY-MM-DD'", heading.Name),
			})
		}
	}

	for _, link := range c.Links {
		key := strings.ToLower(link.Name)
		linkNames[key] = link.Line
	}

	// Rule 3: Each version heading (except possibly Unreleased) must have a link
	for _, heading := range c.Headings {
		key := strings.ToLower(heading.Name)
		if _, hasLink := linkNames[key]; !hasLink {
			// Unreleased may not have a link if there are no releases yet
			if key != "unreleased" {
				result.Errors = append(result.Errors, ValidationError{
					Line:    heading.Line,
					Message: fmt.Sprintf("missing link reference for heading '[%s]'", heading.Name),
				})
			}
		}
	}

	// Rule 4: Links without corresponding headings are warnings
	for _, link := range c.Links {
		key := strings.ToLower(link.Name)
		if _, hasHeading := headingNames[key]; !hasHeading {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Line:    link.Line,
				Message: fmt.Sprintf("link reference '[%s]:' exists without corresponding heading", link.Name),
			})
		}
	}

	return result
}
