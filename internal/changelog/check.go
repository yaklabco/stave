package changelog

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	// ZeroSHA represents a deleted/new ref.
	ZeroSHA = "0000000000000000000000000000000000000000"

	// ChangelogFile is the expected changelog filename.
	ChangelogFile = "CHANGELOG.md"
)

// CheckResult holds the outcome of pre-push checks.
type CheckResult struct {
	ChangelogValid     bool     // CHANGELOG.md format is valid
	ChangelogUpdated   bool     // CHANGELOG.md was modified in the push
	NextVersionPresent bool     // svu next-version exists in changelog
	Errors             []string // Accumulated error messages
	Skipped            bool     // Check was skipped
	SkipReason         string   // Reason for skipping
}

// HasErrors returns true if any errors were found.
func (r *CheckResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// Error returns a combined error or nil.
func (r *CheckResult) Error() error {
	if !r.HasErrors() {
		return nil
	}
	return errors.New(strings.Join(r.Errors, "; "))
}

// PushRef represents a single ref being pushed.
type PushRef struct {
	LocalRef  string
	LocalSHA  string
	RemoteRef string
	RemoteSHA string
}

// PrePushCheckOptions configures pre-push behavior.
type PrePushCheckOptions struct {
	GitOps        GitOps    // Git operations interface
	RemoteName    string    // Name of remote (e.g., "origin")
	Refs          []PushRef // Refs being pushed (from stdin)
	ChangelogPath string    // Path to CHANGELOG.md (default: CHANGELOG.md)
	SkipSVUCheck  bool      // Skip svu next-version verification
}

// PrePushCheck runs all pre-push validations.
// This mirrors the behavior of the bash pre-push hook.
func PrePushCheck(opts PrePushCheckOptions) (*CheckResult, error) {
	result := &CheckResult{
		Errors: make([]string, 0),
	}

	// Check for bypass environment variable
	if os.Getenv("BYPASS_CHANGELOG_CHECK") == "1" {
		result.Skipped = true
		result.SkipReason = "BYPASS_CHANGELOG_CHECK=1"
		return result, nil
	}

	changelogPath := opts.ChangelogPath
	if changelogPath == "" {
		changelogPath = ChangelogFile
	}

	// For git diff comparison, we always use the relative filename
	changelogFilename := ChangelogFile

	// Read and parse changelog
	content, err := os.ReadFile(changelogPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", changelogPath, err)
	}

	cl, err := Parse(string(content))
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("parsing %s: %s", changelogPath, err))
		return result, nil
	}

	// Validate changelog format
	validationResult := cl.Validate()
	if validationResult.HasErrors() {
		result.ChangelogValid = false
		for _, e := range validationResult.Errors {
			if e.Line > 0 {
				result.Errors = append(result.Errors, fmt.Sprintf("%s line %d: %s", changelogPath, e.Line, e.Message))
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", changelogPath, e.Message))
			}
		}
	} else {
		result.ChangelogValid = true
	}

	// Check if changelog was updated in any of the refs being pushed
	sawBranchPush := false
	missingChangelog := false

	for _, ref := range opts.Refs {
		// Skip deleted refs
		if ref.LocalSHA == ZeroSHA {
			continue
		}

		// Skip tag pushes
		if strings.HasPrefix(ref.RemoteRef, "refs/tags/") {
			continue
		}

		// Track if we saw any branch pushes
		if strings.HasPrefix(ref.RemoteRef, "refs/heads/") {
			sawBranchPush = true
		}

		// Skip main/master branches
		if ref.RemoteRef == "refs/heads/main" || ref.RemoteRef == "refs/heads/master" {
			continue
		}

		// Determine base for diff
		base := ref.RemoteSHA
		if base == ZeroSHA {
			// New branch - find merge base with default branch
			base = findDefaultBase(opts.GitOps, opts.RemoteName, ref.LocalSHA)
		}

		if base == "" {
			// Can't determine base - be conservative and require changelog
			missingChangelog = true
			result.Errors = append(result.Errors, fmt.Sprintf("missing %s update for push to %s", changelogFilename, ref.RemoteRef))
			continue
		}

		// Get changed files
		files, err := opts.GitOps.ChangedFiles(base, ref.LocalSHA)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to get changed files: %s", err))
			continue
		}

		if len(files) == 0 {
			// No changed files - be conservative
			missingChangelog = true
			result.Errors = append(result.Errors, fmt.Sprintf("missing %s update for push to %s", changelogFilename, ref.RemoteRef))
			continue
		}

		if !ContainsFile(files, changelogFilename) {
			missingChangelog = true
			result.Errors = append(result.Errors, fmt.Sprintf("missing %s update for push to %s", changelogFilename, ref.RemoteRef))
		}
	}

	result.ChangelogUpdated = !missingChangelog

	// Skip SVU check if:
	// - Explicitly opted out via opts.SkipSVUCheck
	// - SKIP_SVU_CHANGELOG_CHECK=1
	// - GORELEASER_CURRENT_TAG or GORELEASER are set
	// - Only tag pushes (no branch pushes)
	skipSVU := opts.SkipSVUCheck ||
		os.Getenv("SKIP_SVU_CHANGELOG_CHECK") == "1" ||
		os.Getenv("GORELEASER_CURRENT_TAG") != "" ||
		os.Getenv("GORELEASER") != "" ||
		!sawBranchPush

	if skipSVU {
		result.Skipped = true
		result.SkipReason = "svu check skipped (release/tag-only/opt-out)"
		result.NextVersionPresent = true // Skip means we don't fail for this
		return result, nil
	}

	// Verify next version exists in changelog
	nextVersion, err := NextVersion()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to get next version from svu: %s", err))
		return result, nil
	}

	switch {
	case !cl.HasVersion(nextVersion):
		result.NextVersionPresent = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s is missing release heading for version [%s]", changelogPath, nextVersion))
	case !cl.HasLinkForVersion(nextVersion):
		result.NextVersionPresent = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s is missing link reference for [%s]", changelogPath, nextVersion))
	default:
		result.NextVersionPresent = true
	}

	return result, nil
}

// findDefaultBase attempts to find the merge base with the default branch.
func findDefaultBase(gitOps GitOps, remoteName, localSHA string) string {
	// Try main first, then master
	for _, branch := range []string{"main", "master"} {
		ref := fmt.Sprintf("refs/remotes/%s/%s", remoteName, branch)
		if gitOps.RefExists(ref) {
			base, err := gitOps.MergeBase(localSHA, ref)
			if err == nil && base != "" {
				return base
			}
		}
	}
	return ""
}

// ValidateFile reads and validates a CHANGELOG.md file.
// This is a convenience function for the stavefile targets.
func ValidateFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	cl, err := Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	result := cl.Validate()
	if result.HasErrors() {
		return result.Error()
	}

	return nil
}

// VerifyNextVersion checks that the svu next-version exists in the changelog.
// This is a convenience function for the stavefile targets.
func VerifyNextVersion(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	cl, err := Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	nextVersion, err := NextVersion()
	if err != nil {
		return fmt.Errorf("getting next version: %w", err)
	}

	if !cl.HasVersion(nextVersion) {
		return fmt.Errorf("%s is missing release heading for version [%s]", path, nextVersion)
	}

	if !cl.HasLinkForVersion(nextVersion) {
		return fmt.Errorf("%s is missing link reference for [%s]", path, nextVersion)
	}

	return nil
}
