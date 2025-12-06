package changelog

import (
	"errors"
	"regexp"
	"strings"
)

// ChangelogTitle is the expected title for a changelog.
const ChangelogTitle = "Changelog"

// Regular expressions compiled once at package init.
var (
	// titlePattern matches the changelog title line.
	titlePattern = regexp.MustCompile(`^#\s+Changelog\s*$`)

	// headingPattern matches version headings like "## [Unreleased]" or "## [1.2.3] - 2025-01-01".
	headingPattern = regexp.MustCompile(`^##\s+\[(Unreleased|[0-9]+\.[0-9]+\.[0-9]+)\](\s+-\s+([0-9]{4}-[0-9]{2}-[0-9]{2}))?`)

	// Link pattern: "[1.2.3]: https://..." or "[Unreleased]: https://..."
	linkPattern = regexp.MustCompile(`^\[([0-9]+\.[0-9]+\.[0-9]+|[Uu]nreleased)\]:\s*(https?://.+)$`)

	// Semver pattern for identifying release versions.
	semverPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
)

// ErrEmptyContent is returned when parsing empty content.
var ErrEmptyContent = errors.New("changelog content is empty")

// ErrMissingTitle is returned when the changelog lacks a "# Changelog" title.
var ErrMissingTitle = errors.New("changelog must have '# Changelog' title")

// Parse parses CHANGELOG.md content and returns a Changelog struct.
func Parse(content string) (*Changelog, error) {
	if strings.TrimSpace(content) == "" {
		return nil, ErrEmptyContent
	}

	cl := &Changelog{
		Headings: []Heading{},
		Links:    []Link{},
	}

	lines := strings.Split(content, "\n")
	foundTitle := false

	for i, line := range lines {
		lineNum := i + 1 // 1-indexed

		// Check for title
		if titlePattern.MatchString(line) {
			foundTitle = true
			cl.Title = ChangelogTitle
			continue
		}

		// Check for heading
		if h, ok := parseHeading(line, lineNum); ok {
			cl.Headings = append(cl.Headings, *h)
			continue
		}

		// Check for link
		if l, ok := parseLink(line, lineNum); ok {
			cl.Links = append(cl.Links, *l)
			continue
		}
	}

	if !foundTitle {
		return nil, ErrMissingTitle
	}

	return cl, nil
}

// parseHeading parses a version heading line.
// Returns nil and false if the line is not a heading.
func parseHeading(line string, lineNum int) (*Heading, bool) {
	matches := headingPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, false
	}

	heading := &Heading{
		Name: matches[1],
		Line: lineNum,
	}

	// Check if it's a release version (semver)
	if semverPattern.MatchString(heading.Name) {
		heading.IsRelease = true
	}

	// Extract date if present (group 3)
	if len(matches) > 3 && matches[3] != "" {
		heading.Date = matches[3]
	}

	return heading, true
}

// parseLink parses a reference link line.
// Returns nil and false if the line is not a link.
func parseLink(line string, lineNum int) (*Link, bool) {
	matches := linkPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, false
	}

	link := &Link{
		Name: matches[1],
		URL:  matches[2],
		Line: lineNum,
	}

	return link, true
}
