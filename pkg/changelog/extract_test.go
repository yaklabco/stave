package changelog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractSection(t *testing.T) {
	content := `# Changelog
All notable changes to this project will be documented in this file.

## [Unreleased]
### Added
- Some unreleased feature.

## [1.1.0] - 2025-01-02
### Changed
- Some change in 1.1.0.

## [1.0.0] - 2025-01-01
### Added
- Initial release.

[Unreleased]: https://github.com/user/repo/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/user/repo/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/user/repo/releases/tag/v1.0.0
`
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "CHANGELOG.md")
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	tests := []struct {
		name           string
		section        string
		expectedOutput string
		expectErr      bool
	}{
		{
			name:    "extract most recent numbered",
			section: "",
			expectedOutput: `## [1.1.0] - 2025-01-02
### Changed
- Some change in 1.1.0.
`,
		},
		{
			name:    "extract specific version",
			section: "1.0.0",
			expectedOutput: `## [1.0.0] - 2025-01-01
### Added
- Initial release.
`,
		},
		{
			name:      "non-existent version",
			section:   "2.0.0",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputFile := filepath.Join(tmpDir, "output_"+tt.name+".md")
			err := ExtractSection(inputFile, outputFile, tt.section)
			if (err != nil) != tt.expectErr {
				t.Errorf("ExtractSection() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr {
				got, err := os.ReadFile(outputFile)
				if err != nil {
					t.Fatalf("failed to read output file: %v", err)
				}
				if string(got) != tt.expectedOutput {
					t.Errorf("ExtractSection() got =\n%q\nwant =\n%q", string(got), tt.expectedOutput)
				}
			}
		})
	}
}
