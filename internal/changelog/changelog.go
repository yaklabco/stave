// Package changelog provides parsing and validation for Keep a Changelog format.
// It supports validating CHANGELOG.md files and checking pre-push requirements.
package changelog

// Changelog represents a parsed CHANGELOG.md file.
type Changelog struct {
	Title    string    // Expected to be "Changelog"
	Headings []Heading // Version headings (Unreleased and releases)
	Links    []Link    // Reference links at bottom of file
	Raw      string    // Original file content
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
func (c *Changelog) HasVersion(version string) bool {
	for _, h := range c.Headings {
		if h.Name == version {
			return true
		}
	}
	return false
}

// HasLinkForVersion returns true if the changelog has a link reference for the given version.
func (c *Changelog) HasLinkForVersion(version string) bool {
	for _, l := range c.Links {
		if l.Name == version {
			return true
		}
	}
	return false
}
