package changelog

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  error
		validate func(t *testing.T, cl *Changelog)
	}{
		{
			name:    "empty content",
			content: "",
			wantErr: ErrEmptyContent,
		},
		{
			name:    "whitespace only",
			content: "   \n\t\n  ",
			wantErr: ErrEmptyContent,
		},
		{
			name:    "missing title",
			content: "## [Unreleased]\n\n### Added\n- Feature",
			wantErr: ErrMissingTitle,
		},
		{
			name: "valid minimal changelog",
			content: `# Changelog

## [Unreleased]
`,
			validate: func(t *testing.T, cl *Changelog) {
				t.Helper()
				if cl.Title != "Changelog" {
					t.Errorf("Title = %q, want Changelog", cl.Title)
				}
				if len(cl.Headings) != 1 {
					t.Fatalf("Headings count = %d, want 1", len(cl.Headings))
				}
				if cl.Headings[0].Name != "Unreleased" {
					t.Errorf("Heading name = %q, want Unreleased", cl.Headings[0].Name)
				}
				if cl.Headings[0].IsRelease {
					t.Error("Unreleased should not be marked as release")
				}
			},
		},
		{
			name: "valid full changelog",
			content: `# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- New feature

## [1.2.3] - 2025-01-15

### Fixed
- Bug fix

## [1.2.2] - 2025-01-01

### Changed
- Something

[unreleased]: https://github.com/org/repo/compare/v1.2.3...HEAD
[1.2.3]: https://github.com/org/repo/compare/v1.2.2...v1.2.3
[1.2.2]: https://github.com/org/repo/releases/tag/v1.2.2
`,
			validate: func(t *testing.T, cl *Changelog) {
				t.Helper()
				if cl.Title != "Changelog" {
					t.Errorf("Title = %q, want Changelog", cl.Title)
				}
				if len(cl.Headings) != 3 {
					t.Fatalf("Headings count = %d, want 3", len(cl.Headings))
				}

				// Check Unreleased heading
				if cl.Headings[0].Name != "Unreleased" {
					t.Errorf("Heading[0].Name = %q, want Unreleased", cl.Headings[0].Name)
				}
				if cl.Headings[0].IsRelease {
					t.Error("Unreleased should not be a release")
				}

				// Check 1.2.3 heading
				if cl.Headings[1].Name != "1.2.3" {
					t.Errorf("Heading[1].Name = %q, want 1.2.3", cl.Headings[1].Name)
				}
				if cl.Headings[1].Date != "2025-01-15" {
					t.Errorf("Heading[1].Date = %q, want 2025-01-15", cl.Headings[1].Date)
				}
				if !cl.Headings[1].IsRelease {
					t.Error("1.2.3 should be a release")
				}

				// Check links
				if len(cl.Links) != 3 {
					t.Fatalf("Links count = %d, want 3", len(cl.Links))
				}
				if cl.Links[0].Name != "unreleased" {
					t.Errorf("Links[0].Name = %q, want unreleased", cl.Links[0].Name)
				}
				if cl.Links[1].Name != "1.2.3" {
					t.Errorf("Links[1].Name = %q, want 1.2.3", cl.Links[1].Name)
				}
			},
		},
		{
			name: "release without date",
			content: `# Changelog

## [1.0.0]
`,
			validate: func(t *testing.T, cl *Changelog) {
				t.Helper()
				if len(cl.Headings) != 1 {
					t.Fatalf("Headings count = %d, want 1", len(cl.Headings))
				}
				if cl.Headings[0].Date != "" {
					t.Errorf("Date should be empty, got %q", cl.Headings[0].Date)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, err := Parse(tt.content)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Parse() error = nil, want %v", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Parse() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() error = %v, want nil", err)
			}
			if tt.validate != nil {
				tt.validate(t, cl)
			}
		})
	}
}

func TestParseHeading(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		want   *Heading
		wantOK bool
	}{
		{
			name:   "unreleased",
			line:   "## [Unreleased]",
			want:   &Heading{Name: "Unreleased", Line: 1, IsRelease: false},
			wantOK: true,
		},
		{
			name:   "release with date",
			line:   "## [1.2.3] - 2025-01-15",
			want:   &Heading{Name: "1.2.3", Date: "2025-01-15", Line: 1, IsRelease: true},
			wantOK: true,
		},
		{
			name:   "release without date",
			line:   "## [2.0.0]",
			want:   &Heading{Name: "2.0.0", Date: "", Line: 1, IsRelease: true},
			wantOK: true,
		},
		{
			name:   "not a heading - wrong prefix",
			line:   "# [1.0.0] - 2025-01-01",
			want:   nil,
			wantOK: false,
		},
		{
			name:   "not a heading - section",
			line:   "### Added",
			want:   nil,
			wantOK: false,
		},
		{
			name:   "not a heading - plain text",
			line:   "Some text",
			want:   nil,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseHeading(tt.line, 1)
			if ok != tt.wantOK {
				t.Errorf("parseHeading() ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if !tt.wantOK {
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Date != tt.want.Date {
				t.Errorf("Date = %q, want %q", got.Date, tt.want.Date)
			}
			if got.IsRelease != tt.want.IsRelease {
				t.Errorf("IsRelease = %v, want %v", got.IsRelease, tt.want.IsRelease)
			}
		})
	}
}

func TestParseLink(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		want   *Link
		wantOK bool
	}{
		{
			name:   "version link",
			line:   "[1.2.3]: https://github.com/org/repo/compare/v1.2.2...v1.2.3",
			want:   &Link{Name: "1.2.3", URL: "https://github.com/org/repo/compare/v1.2.2...v1.2.3", Line: 1},
			wantOK: true,
		},
		{
			name:   "unreleased link lowercase",
			line:   "[unreleased]: https://github.com/org/repo/compare/v1.2.3...HEAD",
			want:   &Link{Name: "unreleased", URL: "https://github.com/org/repo/compare/v1.2.3...HEAD", Line: 1},
			wantOK: true,
		},
		{
			name:   "unreleased link capitalized",
			line:   "[Unreleased]: https://github.com/org/repo/compare/v1.2.3...HEAD",
			want:   &Link{Name: "Unreleased", URL: "https://github.com/org/repo/compare/v1.2.3...HEAD", Line: 1},
			wantOK: true,
		},
		{
			name:   "http link",
			line:   "[1.0.0]: http://example.com/releases/1.0.0",
			want:   &Link{Name: "1.0.0", URL: "http://example.com/releases/1.0.0", Line: 1},
			wantOK: true,
		},
		{
			name:   "not a link - regular text",
			line:   "Some text",
			want:   nil,
			wantOK: false,
		},
		{
			name:   "not a link - markdown link",
			line:   "[text](https://example.com)",
			want:   nil,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseLink(tt.line, 1)
			if ok != tt.wantOK {
				t.Errorf("parseLink() ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if !tt.wantOK {
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
			}
		})
	}
}

func TestChangelog_HasVersion(t *testing.T) {
	cl := &Changelog{
		Headings: []Heading{
			{Name: "Unreleased"},
			{Name: "1.2.3"},
			{Name: "1.2.2"},
		},
	}

	tests := []struct {
		version string
		want    bool
	}{
		{"Unreleased", true},
		{"1.2.3", true},
		{"1.2.2", true},
		{"1.0.0", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := cl.HasVersion(tt.version); got != tt.want {
				t.Errorf("HasVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestChangelog_HasLinkForVersion(t *testing.T) {
	cl := &Changelog{
		Links: []Link{
			{Name: "unreleased"},
			{Name: "1.2.3"},
			{Name: "1.2.2"},
		},
	}

	tests := []struct {
		version string
		want    bool
	}{
		{"unreleased", true},
		{"1.2.3", true},
		{"1.2.2", true},
		{"1.0.0", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := cl.HasLinkForVersion(tt.version); got != tt.want {
				t.Errorf("HasLinkForVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}
