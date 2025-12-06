// Package changelog provides parsing and validation for Keep a Changelog format.
// It supports validating CHANGELOG.md files and checking pre-push requirements.
package changelog

import (
	"slices"
	"strings"
)

// Changelog represents a parsed CHANGELOG.md file.
type Changelog struct {
	Title    string    // Expected to be "Changelog"
	Headings []Heading // Version headings (Unreleased and releases)
	Links    []Link    // Reference links at bottom of file
}

// Heading represents a version heading like "## [1.2.3] - 2025-01-01".
type Heading struct {
	Name      string // "Unreleased" or semver "1.2.3"
	Date      string // "2025-01-01" or empty for Unreleased
	Line      int    // 1-indexed line number
	IsRelease bool   // true if semver version, false for Unreleased
}

// Link represents a reference link like "[1.2.3]: https://...".
type Link struct {
	Name string // "Unreleased" or semver "1.2.3"
	URL  string // Full URL
	Line int    // 1-indexed line number
}

// HasVersion returns true if the changelog has a heading for the given version.
// Comparison is case-insensitive to match validation behavior.
func (c *Changelog) HasVersion(version string) bool {
	return slices.ContainsFunc(c.Headings, func(h Heading) bool {
		return strings.EqualFold(h.Name, version)
	})
}

// HasLinkForVersion returns true if the changelog has a link reference for the given version.
// Comparison is case-insensitive to match validation behavior.
func (c *Changelog) HasLinkForVersion(version string) bool {
	return slices.ContainsFunc(c.Links, func(l Link) bool {
		return strings.EqualFold(l.Name, version)
	})
}
