package changelog

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	// testValidChangelog is a minimal valid changelog for testing.
	testValidChangelog = `# Changelog

## [Unreleased]
`
	// testMergeBase is a test merge base commit hash.
	testMergeBase = "abc123"
)

func TestPrePushCheck_BypassEnv(t *testing.T) {
	t.Setenv("BYPASS_CHANGELOG_CHECK", "1")

	result, err := PrePushCheck(PrePushCheckOptions{})
	if err != nil {
		t.Fatalf("PrePushCheck() error = %v", err)
	}
	if !result.Skipped {
		t.Error("Expected check to be skipped")
	}
	if result.SkipReason != "BYPASS_CHANGELOG_CHECK=1" {
		t.Errorf("SkipReason = %q, want BYPASS_CHANGELOG_CHECK=1", result.SkipReason)
	}
}

func TestPrePushCheck_ValidChangelog(t *testing.T) {
	// Create temp dir with valid changelog
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	validChangelog := `# Changelog

## [Unreleased]

## [1.0.0] - 2025-01-01

### Added
- Initial release

[unreleased]: https://github.com/org/repo/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/org/repo/releases/tag/v1.0.0
`
	if err := os.WriteFile(changelogPath, []byte(validChangelog), 0o600); err != nil {
		t.Fatalf("Failed to write changelog: %v", err)
	}

	// Mock that returns CHANGELOG.md in changed files
	mock := &mockGitOps{
		changedFiles:  []string{"README.md", "CHANGELOG.md", "main.go"},
		mergeBase:     testMergeBase,
		refExists:     map[string]bool{"refs/remotes/origin/main": true},
		currentBranch: "feature-branch",
	}

	result, err := PrePushCheck(PrePushCheckOptions{
		GitOps:               mock,
		RemoteName:           "origin",
		ChangelogPath:        changelogPath,
		SkipNextVersionCheck: true, // Skip next-version check for this test
		Refs: []PushRef{
			{
				LocalRef:  "refs/heads/feature",
				LocalSHA:  "def456",
				RemoteRef: "refs/heads/feature",
				RemoteSHA: testMergeBase,
			},
		},
	})
	if err != nil {
		t.Fatalf("PrePushCheck() error = %v", err)
	}
	if !result.ChangelogValid {
		t.Error("Expected ChangelogValid = true")
	}
	if !result.ChangelogUpdated {
		t.Error("Expected ChangelogUpdated = true")
	}
	if result.HasErrors() {
		t.Errorf("Unexpected errors: %v", result.Errors)
	}
}

func TestPrePushCheck_MissingChangelog(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	validChangelog := `# Changelog

## [Unreleased]

[unreleased]: https://github.com/org/repo/compare/v1.0.0...HEAD
`
	if err := os.WriteFile(changelogPath, []byte(validChangelog), 0o600); err != nil {
		t.Fatalf("Failed to write changelog: %v", err)
	}

	// Mock that does NOT include CHANGELOG.md in changed files
	mock := &mockGitOps{
		changedFiles:  []string{"README.md", "main.go"},
		mergeBase:     testMergeBase,
		refExists:     map[string]bool{"refs/remotes/origin/main": true},
		currentBranch: "feature-branch",
	}

	result, err := PrePushCheck(PrePushCheckOptions{
		GitOps:               mock,
		RemoteName:           "origin",
		ChangelogPath:        changelogPath,
		SkipNextVersionCheck: true,
		Refs: []PushRef{
			{
				LocalRef:  "refs/heads/feature",
				LocalSHA:  "def456",
				RemoteRef: "refs/heads/feature",
				RemoteSHA: testMergeBase,
			},
		},
	})
	if err != nil {
		t.Fatalf("PrePushCheck() error = %v", err)
	}
	if !result.ChangelogValid {
		t.Error("Expected ChangelogValid = true (format is valid)")
	}
	if result.ChangelogUpdated {
		t.Error("Expected ChangelogUpdated = false")
	}
	if !result.HasErrors() {
		t.Error("Expected errors for missing changelog update")
	}
}

func TestPrePushCheck_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Release without date
	invalidChangelog := `# Changelog

## [Unreleased]

## [1.0.0]

### Added
- Initial release
`
	if err := os.WriteFile(changelogPath, []byte(invalidChangelog), 0o600); err != nil {
		t.Fatalf("Failed to write changelog: %v", err)
	}

	mock := &mockGitOps{
		changedFiles: []string{"CHANGELOG.md"},
		mergeBase:    testMergeBase,
		refExists:    map[string]bool{"refs/remotes/origin/main": true},
	}

	result, err := PrePushCheck(PrePushCheckOptions{
		GitOps:               mock,
		RemoteName:           "origin",
		ChangelogPath:        changelogPath,
		SkipNextVersionCheck: true,
		Refs: []PushRef{
			{
				LocalRef:  "refs/heads/feature",
				LocalSHA:  "def456",
				RemoteRef: "refs/heads/feature",
				RemoteSHA: testMergeBase,
			},
		},
	})
	if err != nil {
		t.Fatalf("PrePushCheck() error = %v", err)
	}
	if result.ChangelogValid {
		t.Error("Expected ChangelogValid = false")
	}
	if !result.HasErrors() {
		t.Error("Expected errors for invalid format")
	}
}

func TestPrePushCheck_SkipTagPush(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	if err := os.WriteFile(changelogPath, []byte(testValidChangelog), 0o600); err != nil {
		t.Fatalf("Failed to write changelog: %v", err)
	}

	mock := &mockGitOps{
		changedFiles: []string{}, // No changelog change
	}

	result, err := PrePushCheck(PrePushCheckOptions{
		GitOps:               mock,
		RemoteName:           "origin",
		ChangelogPath:        changelogPath,
		SkipNextVersionCheck: true,
		Refs: []PushRef{
			{
				LocalRef:  "refs/tags/v1.0.0",
				LocalSHA:  "def456",
				RemoteRef: "refs/tags/v1.0.0",
				RemoteSHA: ZeroSHA,
			},
		},
	})
	if err != nil {
		t.Fatalf("PrePushCheck() error = %v", err)
	}
	// Tag pushes should be skipped, no errors about missing changelog
	if result.HasErrors() {
		t.Errorf("Tag push should not cause errors: %v", result.Errors)
	}
}

func TestPrePushCheck_SkipMainBranch(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	if err := os.WriteFile(changelogPath, []byte(testValidChangelog), 0o600); err != nil {
		t.Fatalf("Failed to write changelog: %v", err)
	}

	mock := &mockGitOps{
		changedFiles: []string{}, // No changelog change
	}

	result, err := PrePushCheck(PrePushCheckOptions{
		GitOps:               mock,
		RemoteName:           "origin",
		ChangelogPath:        changelogPath,
		SkipNextVersionCheck: true,
		Refs: []PushRef{
			{
				LocalRef:  "refs/heads/main",
				LocalSHA:  "def456",
				RemoteRef: "refs/heads/main",
				RemoteSHA: testMergeBase,
			},
		},
	})
	if err != nil {
		t.Fatalf("PrePushCheck() error = %v", err)
	}
	// Main branch pushes should be skipped
	if result.HasErrors() {
		t.Errorf("Main branch push should not cause errors: %v", result.Errors)
	}
}

func TestPrePushCheck_SkipNextVerCheckEnvVars(t *testing.T) {
	tests := []struct {
		name   string
		envKey string
		envVal string
	}{
		{"STAVEFILE_SKIP_NEXTVER_CHANGELOG_CHECK", "STAVEFILE_SKIP_NEXTVER_CHANGELOG_CHECK", "1"},
		{"GORELEASER_CURRENT_TAG", "GORELEASER_CURRENT_TAG", "v1.0.0"},
		{"GORELEASER", "GORELEASER", "true"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

			if err := os.WriteFile(changelogPath, []byte(testValidChangelog), 0o600); err != nil {
				t.Fatalf("Failed to write changelog: %v", err)
			}

			t.Setenv(testCase.envKey, testCase.envVal)

			mock := &mockGitOps{
				changedFiles: []string{"CHANGELOG.md"},
				mergeBase:    testMergeBase,
				refExists:    map[string]bool{"refs/remotes/origin/main": true},
			}

			result, err := PrePushCheck(PrePushCheckOptions{
				GitOps:               mock,
				RemoteName:           "origin",
				ChangelogPath:        changelogPath,
				SkipNextVersionCheck: false, // Not skipping via options
				Refs: []PushRef{
					{
						LocalRef:  "refs/heads/feature",
						LocalSHA:  "def456",
						RemoteRef: "refs/heads/feature",
						RemoteSHA: testMergeBase,
					},
				},
			})
			if err != nil {
				t.Fatalf("PrePushCheck() error = %v", err)
			}
			if !result.Skipped {
				t.Error("Expected check to be skipped due to env var")
			}
		})
	}
}

func TestPrePushCheck_NewBranch(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	if err := os.WriteFile(changelogPath, []byte(testValidChangelog), 0o600); err != nil {
		t.Fatalf("Failed to write changelog: %v", err)
	}

	mock := &mockGitOps{
		changedFiles: []string{"CHANGELOG.md"},
		mergeBase:    testMergeBase,
		refExists:    map[string]bool{"refs/remotes/origin/main": true},
	}

	result, err := PrePushCheck(PrePushCheckOptions{
		GitOps:               mock,
		RemoteName:           "origin",
		ChangelogPath:        changelogPath,
		SkipNextVersionCheck: true,
		Refs: []PushRef{
			{
				LocalRef:  "refs/heads/new-feature",
				LocalSHA:  "def456",
				RemoteRef: "refs/heads/new-feature",
				RemoteSHA: ZeroSHA, // New branch
			},
		},
	})
	if err != nil {
		t.Fatalf("PrePushCheck() error = %v", err)
	}
	if result.HasErrors() {
		t.Errorf("New branch with changelog update should pass: %v", result.Errors)
	}
}

func TestValidateFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("valid changelog", func(t *testing.T) {
		path := filepath.Join(tmpDir, "valid.md")
		content := `# Changelog

## [Unreleased]

## [1.0.0] - 2025-01-01

[unreleased]: https://example.com
[1.0.0]: https://example.com
`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		if err := ValidateFile(path); err != nil {
			t.Errorf("ValidateFile() error = %v, want nil", err)
		}
	})

	t.Run("invalid changelog", func(t *testing.T) {
		path := filepath.Join(tmpDir, "invalid.md")
		content := `# Changelog

## [1.0.0]
`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		if err := ValidateFile(path); err == nil {
			t.Error("ValidateFile() error = nil, want error")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		if err := ValidateFile(filepath.Join(tmpDir, "nonexistent.md")); err == nil {
			t.Error("ValidateFile() error = nil, want error for missing file")
		}
	})
}

func TestCheckResult_Error(t *testing.T) {
	tests := []struct {
		name    string
		result  *CheckResult
		wantNil bool
	}{
		{
			name:    "no errors",
			result:  &CheckResult{Errors: []string{}},
			wantNil: true,
		},
		{
			name:    "with errors",
			result:  &CheckResult{Errors: []string{"error1", "error2"}},
			wantNil: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.result.Error()
			if testCase.wantNil && err != nil {
				t.Errorf("Error() = %v, want nil", err)
			}
			if !testCase.wantNil && err == nil {
				t.Error("Error() = nil, want error")
			}
		})
	}
}

func TestFindDefaultBase(t *testing.T) {
	t.Run("main exists", func(t *testing.T) {
		mock := &mockGitOps{
			mergeBase: testMergeBase,
			refExists: map[string]bool{"refs/remotes/origin/main": true},
		}
		base := findDefaultBase(mock, "origin", "def456")
		if base != testMergeBase {
			t.Errorf("findDefaultBase() = %q, want abc123", base)
		}
	})

	t.Run("master fallback", func(t *testing.T) {
		mock := &mockGitOps{
			mergeBase: testMergeBase,
			refExists: map[string]bool{"refs/remotes/origin/master": true},
		}
		base := findDefaultBase(mock, "origin", "def456")
		if base != testMergeBase {
			t.Errorf("findDefaultBase() = %q, want abc123", base)
		}
	})

	t.Run("no default branch", func(t *testing.T) {
		mock := &mockGitOps{
			mergeBase: testMergeBase,
			refExists: map[string]bool{},
		}
		base := findDefaultBase(mock, "origin", "def456")
		if base != "" {
			t.Errorf("findDefaultBase() = %q, want empty", base)
		}
	})
}
