package changelog

import (
	"testing"
)

func TestChangelog_Validate(t *testing.T) {
	tests := []struct {
		name       string
		changelog  *Changelog
		wantErrors int
		wantWarns  int
	}{
		{
			name: "valid changelog",
			changelog: &Changelog{
				Title: "Changelog",
				Headings: []Heading{
					{Name: "Unreleased", Line: 3, IsRelease: false},
					{Name: "1.2.3", Date: "2025-01-15", Line: 7, IsRelease: true},
					{Name: "1.2.2", Date: "2025-01-01", Line: 12, IsRelease: true},
				},
				Links: []Link{
					{Name: "unreleased", URL: "https://github.com/org/repo/compare/v1.2.3...HEAD", Line: 20},
					{Name: "1.2.3", URL: "https://github.com/org/repo/compare/v1.2.2...v1.2.3", Line: 21},
					{Name: "1.2.2", URL: "https://github.com/org/repo/releases/tag/v1.2.2", Line: 22},
				},
			},
			wantErrors: 0,
			wantWarns:  0,
		},
		{
			name: "missing title",
			changelog: &Changelog{
				Title: "",
				Headings: []Heading{
					{Name: "Unreleased", Line: 1, IsRelease: false},
				},
				Links: []Link{},
			},
			wantErrors: 1,
			wantWarns:  0,
		},
		{
			name: "release without date",
			changelog: &Changelog{
				Title: "Changelog",
				Headings: []Heading{
					{Name: "Unreleased", Line: 3, IsRelease: false},
					{Name: "1.0.0", Date: "", Line: 7, IsRelease: true},
				},
				Links: []Link{
					{Name: "1.0.0", URL: "https://example.com", Line: 15},
				},
			},
			wantErrors: 1,
			wantWarns:  0,
		},
		{
			name: "missing link for release",
			changelog: &Changelog{
				Title: "Changelog",
				Headings: []Heading{
					{Name: "Unreleased", Line: 3, IsRelease: false},
					{Name: "1.0.0", Date: "2025-01-01", Line: 7, IsRelease: true},
				},
				Links: []Link{}, // No links
			},
			wantErrors: 1,
			wantWarns:  0,
		},
		{
			name: "orphan link (warning)",
			changelog: &Changelog{
				Title: "Changelog",
				Headings: []Heading{
					{Name: "Unreleased", Line: 3, IsRelease: false},
				},
				Links: []Link{
					{Name: "1.0.0", URL: "https://example.com", Line: 10},
				},
			},
			wantErrors: 0,
			wantWarns:  1,
		},
		{
			name: "unreleased without link is OK",
			changelog: &Changelog{
				Title: "Changelog",
				Headings: []Heading{
					{Name: "Unreleased", Line: 3, IsRelease: false},
				},
				Links: []Link{}, // No links is fine for Unreleased only
			},
			wantErrors: 0,
			wantWarns:  0,
		},
		{
			name: "case insensitive link matching",
			changelog: &Changelog{
				Title: "Changelog",
				Headings: []Heading{
					{Name: "Unreleased", Line: 3, IsRelease: false},
					{Name: "1.0.0", Date: "2025-01-01", Line: 7, IsRelease: true},
				},
				Links: []Link{
					{Name: "UNRELEASED", URL: "https://example.com/head", Line: 15},
					{Name: "1.0.0", URL: "https://example.com/1.0.0", Line: 16},
				},
			},
			wantErrors: 0,
			wantWarns:  0,
		},
		{
			name: "multiple errors",
			changelog: &Changelog{
				Title: "",
				Headings: []Heading{
					{Name: "1.0.0", Date: "", Line: 3, IsRelease: true},
					{Name: "0.9.0", Date: "", Line: 7, IsRelease: true},
				},
				Links: []Link{},
			},
			wantErrors: 5, // missing title + 2 missing dates + 2 missing links
			wantWarns:  0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.changelog.Validate()
			if len(result.Errors) != testCase.wantErrors {
				t.Errorf("Errors count = %d, want %d", len(result.Errors), testCase.wantErrors)
				for i, e := range result.Errors {
					t.Logf("  Error[%d]: line %d: %s", i, e.Line, e.Message)
				}
			}
			if len(result.Warnings) != testCase.wantWarns {
				t.Errorf("Warnings count = %d, want %d", len(result.Warnings), testCase.wantWarns)
				for i, w := range result.Warnings {
					t.Logf("  Warning[%d]: line %d: %s", i, w.Line, w.Message)
				}
			}
		})
	}
}

func TestValidationResult_HasErrors(t *testing.T) {
	tests := []struct {
		name   string
		result ValidationResult
		want   bool
	}{
		{
			name:   "no errors",
			result: ValidationResult{Errors: []ValidationError{}},
			want:   false,
		},
		{
			name:   "with errors",
			result: ValidationResult{Errors: []ValidationError{{Message: "test"}}},
			want:   true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.result.HasErrors(); got != testCase.want {
				t.Errorf("HasErrors() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestValidationResult_Error(t *testing.T) {
	tests := []struct {
		name    string
		result  ValidationResult
		wantNil bool
		wantMsg string
	}{
		{
			name:    "no errors returns nil",
			result:  ValidationResult{Errors: []ValidationError{}},
			wantNil: true,
		},
		{
			name: "single error without line",
			result: ValidationResult{
				Errors: []ValidationError{{Message: "test error"}},
			},
			wantMsg: "test error",
		},
		{
			name: "single error with line",
			result: ValidationResult{
				Errors: []ValidationError{{Line: 5, Message: "test error"}},
			},
			wantMsg: "line 5: test error",
		},
		{
			name: "multiple errors",
			result: ValidationResult{
				Errors: []ValidationError{
					{Line: 5, Message: "error one"},
					{Line: 10, Message: "error two"},
				},
			},
			wantMsg: "line 5: error one; line 10: error two",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.result.Error()
			if testCase.wantNil {
				if err != nil {
					t.Errorf("Error() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Error() = nil, want error")
			}
			if err.Error() != testCase.wantMsg {
				t.Errorf("Error() = %q, want %q", err.Error(), testCase.wantMsg)
			}
		})
	}
}
