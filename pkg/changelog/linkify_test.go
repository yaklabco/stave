package changelog

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangelog_Linkify(t *testing.T) {
	content := `# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.3.4] - 2025-12-16

### Changed

- Outdated outputln string in main app stavefile.go file.

## [0.3.3] - 2025-12-16

### Removed

- Extra printing of errors in main.go

[0.3.3]: https://github.com/yaklabco/stave/compare/v0.3.2...v0.3.3
`
	// Expected: [unreleased] and [0.3.4] are missing links.
	// [unreleased] should compare v0.3.4...HEAD
	// [0.3.4] should compare v0.3.3...v0.3.4
	// They should be inserted before [0.3.3] link because their headings are before 0.3.3 heading.

	expected := `# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.3.4] - 2025-12-16

### Changed

- Outdated outputln string in main app stavefile.go file.

## [0.3.3] - 2025-12-16

### Removed

- Extra printing of errors in main.go

[unreleased]: https://github.com/yaklabco/stave/compare/v0.3.4...HEAD
[0.3.4]: https://github.com/yaklabco/stave/compare/v0.3.3...v0.3.4
[0.3.3]: https://github.com/yaklabco/stave/compare/v0.3.2...v0.3.3
`

	got, err := LinkifyContent(content)
	if err != nil {
		t.Fatalf("LinkifyContent failed: %v", err)
	}

	if got != expected {
		t.Errorf("LinkifyContent mismatch\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestChangelog_Linkify_NoExistingLinks(t *testing.T) {
	content := `# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.3.4] - 2025-12-16

## [0.1.0] - 2025-12-08

[0.1.0]: https://github.com/yaklabco/stave/releases/tag/v0.1.0
`

	_, err := LinkifyContent(content)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not determine")
}

func TestChangelog_Linkify_UpdateUnreleased(t *testing.T) {
	content := `# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.3.5] - 2025-12-19

## [0.3.4] - 2025-12-16

[unreleased]: https://github.com/yaklabco/stave/compare/v0.3.4...HEAD
[0.3.4]: https://github.com/yaklabco/stave/compare/v0.3.3...v0.3.4
`
	// [0.3.5] is missing. It is the topmost version under [Unreleased].
	// [unreleased] should be updated to compare v0.3.5...HEAD.
	expected := `# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.3.5] - 2025-12-19

## [0.3.4] - 2025-12-16

[unreleased]: https://github.com/yaklabco/stave/compare/v0.3.5...HEAD
[0.3.5]: https://github.com/yaklabco/stave/compare/v0.3.4...v0.3.5
[0.3.4]: https://github.com/yaklabco/stave/compare/v0.3.3...v0.3.4
`

	got, err := LinkifyContent(content)
	if err != nil {
		t.Fatalf("LinkifyContent failed: %v", err)
	}

	if got != expected {
		t.Errorf("LinkifyContent mismatch\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestLinkify_PreservesPermissions(t *testing.T) {
	// Create a temporary file with specific permissions
	tmpFile, err := os.CreateTemp("", "CHANGELOG_test_*.md")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(
		func() {
			_ = tmpFile.Close()
		},
	)

	content := `# Changelog

## [Unreleased]

[unreleased]: https://github.com/user/repo/compare/v0.1.0...main
[0.1.0]: https://github.com/user/repo/releases/tag/v0.1.0
`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tmpFile.Close())

	// Set custom permissions
	expectedMode := os.FileMode(0600)
	if err := os.Chmod(tmpFile.Name(), expectedMode); err != nil {
		t.Fatal(err)
	}

	// Linkify should preserve permissions even if no changes are made
	// In our case, we'll trigger a change to ensure WriteFile is called.
	contentWithNewVersion := `# Changelog

## [Unreleased]

## [0.2.0] - 2025-12-19

## [0.1.0] - 2025-12-08

[unreleased]: https://github.com/user/repo/compare/v0.1.0...main
[0.1.0]: https://github.com/user/repo/releases/tag/v0.1.0
`
	if err := os.WriteFile(tmpFile.Name(), []byte(contentWithNewVersion), expectedMode); err != nil {
		t.Fatal(err)
	}

	err = Linkify(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != expectedMode {
		t.Errorf("Expected permissions %o, got %o", expectedMode, info.Mode().Perm())
	}
}
