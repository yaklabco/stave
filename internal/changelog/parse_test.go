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
		validate func(t *testing.T, parsedChangelog *Changelog)
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
			validate: func(t *testing.T, parsedChangelog *Changelog) {
				t.Helper()
				if parsedChangelog.Title != "Changelog" {
					t.Errorf("Title = %q, want Changelog", parsedChangelog.Title)
				}
				if len(parsedChangelog.Headings) != 1 {
					t.Fatalf("Headings count = %d, want 1", len(parsedChangelog.Headings))
				}
				if parsedChangelog.Headings[0].Name != "Unreleased" {
					t.Errorf("Heading name = %q, want Unreleased", parsedChangelog.Headings[0].Name)
				}
				if parsedChangelog.Headings[0].IsRelease {
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
			validate: func(t *testing.T, parsedChangelog *Changelog) {
				t.Helper()
				if parsedChangelog.Title != "Changelog" {
					t.Errorf("Title = %q, want Changelog", parsedChangelog.Title)
				}
				if len(parsedChangelog.Headings) != 3 {
					t.Fatalf("Headings count = %d, want 3", len(parsedChangelog.Headings))
				}

				// Check Unreleased heading
				if parsedChangelog.Headings[0].Name != "Unreleased" {
					t.Errorf("Heading[0].Name = %q, want Unreleased", parsedChangelog.Headings[0].Name)
				}
				if parsedChangelog.Headings[0].IsRelease {
					t.Error("Unreleased should not be a release")
				}

				// Check 1.2.3 heading
				if parsedChangelog.Headings[1].Name != "1.2.3" {
					t.Errorf("Heading[1].Name = %q, want 1.2.3", parsedChangelog.Headings[1].Name)
				}
				if parsedChangelog.Headings[1].Date != "2025-01-15" {
					t.Errorf("Heading[1].Date = %q, want 2025-01-15", parsedChangelog.Headings[1].Date)
				}
				if !parsedChangelog.Headings[1].IsRelease {
					t.Error("1.2.3 should be a release")
				}

				// Check links
				if len(parsedChangelog.Links) != 3 {
					t.Fatalf("Links count = %d, want 3", len(parsedChangelog.Links))
				}
				if parsedChangelog.Links[0].Name != "unreleased" {
					t.Errorf("Links[0].Name = %q, want unreleased", parsedChangelog.Links[0].Name)
				}
				if parsedChangelog.Links[1].Name != "1.2.3" {
					t.Errorf("Links[1].Name = %q, want 1.2.3", parsedChangelog.Links[1].Name)
				}
			},
		},
		{
			name: "release without date",
			content: `# Changelog

## [1.0.0]
`,
			validate: func(t *testing.T, parsedChangelog *Changelog) {
				t.Helper()
				if len(parsedChangelog.Headings) != 1 {
					t.Fatalf("Headings count = %d, want 1", len(parsedChangelog.Headings))
				}
				if parsedChangelog.Headings[0].Date != "" {
					t.Errorf("Date should be empty, got %q", parsedChangelog.Headings[0].Date)
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			parsedChangelog, err := Parse(testCase.content)
			if testCase.wantErr != nil {
				if err == nil {
					t.Fatalf("Parse() error = nil, want %v", testCase.wantErr)
				}
				if !errors.Is(err, testCase.wantErr) {
					t.Errorf("Parse() error = %v, want %v", err, testCase.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() error = %v, want nil", err)
			}
			if testCase.validate != nil {
				testCase.validate(t, parsedChangelog)
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, ok := parseHeading(testCase.line, 1)
			if ok != testCase.wantOK {
				t.Errorf("parseHeading() ok = %v, want %v", ok, testCase.wantOK)
				return
			}
			if !testCase.wantOK {
				return
			}
			if got.Name != testCase.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, testCase.want.Name)
			}
			if got.Date != testCase.want.Date {
				t.Errorf("Date = %q, want %q", got.Date, testCase.want.Date)
			}
			if got.IsRelease != testCase.want.IsRelease {
				t.Errorf("IsRelease = %v, want %v", got.IsRelease, testCase.want.IsRelease)
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, ok := parseLink(testCase.line, 1)
			if ok != testCase.wantOK {
				t.Errorf("parseLink() ok = %v, want %v", ok, testCase.wantOK)
				return
			}
			if !testCase.wantOK {
				return
			}
			if got.Name != testCase.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, testCase.want.Name)
			}
			if got.URL != testCase.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, testCase.want.URL)
			}
		})
	}
}

func TestChangelog_HasVersion(t *testing.T) {
	testChangelog := &Changelog{
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

	for _, testCase := range tests {
		t.Run(testCase.version, func(t *testing.T) {
			if got := testChangelog.HasVersion(testCase.version); got != testCase.want {
				t.Errorf("HasVersion(%q) = %v, want %v", testCase.version, got, testCase.want)
			}
		})
	}
}

func TestChangelog_HasLinkForVersion(t *testing.T) {
	testChangelog := &Changelog{
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

	for _, testCase := range tests {
		t.Run(testCase.version, func(t *testing.T) {
			if got := testChangelog.HasLinkForVersion(testCase.version); got != testCase.want {
				t.Errorf("HasLinkForVersion(%q) = %v, want %v", testCase.version, got, testCase.want)
			}
		})
	}
}
