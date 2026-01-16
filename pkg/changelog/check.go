package changelog

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
	NextVersionPresent bool     // next version to-be-released exists in changelog
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
	GitOps               GitOps    // Git operations interface
	RemoteName           string    // Name of remote (e.g., "origin")
	Refs                 []PushRef // Refs being pushed (from stdin)
	ChangelogPath        string    // Path to CHANGELOG.md (default: CHANGELOG.md)
	SkipNextVersionCheck bool      // Skip next-version verification
}

// PrePushCheck runs all pre-push validations.
// This mirrors the behavior of the bash pre-push hook.
func PrePushCheck(opts PrePushCheckOptions) (*CheckResult, error) {
	result := &CheckResult{}

	if opts.GitOps == nil {
		opts.GitOps = &ShellGitOps{}
	}

	if os.Getenv("BYPASS_CHANGELOG_CHECK") == "1" {
		result.Skipped = true
		result.SkipReason = "BYPASS_CHANGELOG_CHECK=1"
		return result, nil
	}

	changelogPath := resolveChangelogPath(opts.ChangelogPath)

	changelog, err := readAndParseChangelog(changelogPath, result)
	if err != nil {
		if errors.Is(err, errParseFailure) {
			return result, nil // Parse error already recorded
		}
		return nil, err
	}

	validateChangelogFormat(changelog, changelogPath, result)

	sawBranchPush := checkRefChanges(opts, result)

	if shouldSkipNextVersionCheck(opts.SkipNextVersionCheck, sawBranchPush) {
		result.Skipped = true
		result.SkipReason = "next-version check skipped (release/tag-only/opt-out)"
		result.NextVersionPresent = true
		return result, nil
	}

	verifyNextVersionInChangelog(changelog, changelogPath, result)

	return result, nil
}

func ReadPushRefs(stdin io.Reader) ([]PushRef, error) {
	scanner := bufio.NewScanner(stdin)

	pushRefs := make([]PushRef, 0)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) != 4 { //nolint:mnd // Git has a format with 4 fields here, it is what it is.
			return nil, fmt.Errorf("invalid push-refs line: %q", line)
		}

		pushRefs = append(pushRefs, PushRef{
			LocalRef:  fields[0],
			LocalSHA:  fields[1],
			RemoteRef: fields[2],
			RemoteSHA: fields[3],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading from stdin: %w", err)
	}

	return pushRefs, nil
}

// resolveChangelogPath returns the changelog path with a default fallback.
func resolveChangelogPath(path string) string {
	if path == "" {
		return ChangelogFile
	}
	return path
}

// errParseFailure is returned when changelog parsing fails but errors are recorded in result.
var errParseFailure = errors.New("changelog parse failure")

// readAndParseChangelog reads and parses the changelog file.
// Returns errParseFailure if parsing fails, with error recorded in result.
func readAndParseChangelog(path string, result *CheckResult) (*Changelog, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	changelog, err := Parse(string(content))
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("parsing %s: %s", path, err))
		return nil, errParseFailure
	}
	return changelog, nil
}

// validateChangelogFormat validates the changelog and records any errors.
func validateChangelogFormat(changelog *Changelog, path string, result *CheckResult) {
	validationResult := changelog.Validate()
	if validationResult.HasErrors() {
		result.ChangelogValid = false
		for _, validationErr := range validationResult.Errors {
			if validationErr.Line > 0 {
				result.Errors = append(
					result.Errors,
					fmt.Sprintf("%s line %d: %s", path, validationErr.Line, validationErr.Message))
			} else {
				result.Errors = append(
					result.Errors, fmt.Sprintf("%s: %s", path, validationErr.Message))
			}
		}
	} else {
		result.ChangelogValid = true
	}
}

// checkRefChanges checks each ref being pushed for changelog updates.
// Returns true if any branch pushes were seen.
func checkRefChanges(opts PrePushCheckOptions, result *CheckResult) bool {
	sawBranchPush := false
	missingChangelog := false

	for _, ref := range opts.Refs {
		if skipRef(ref) {
			continue
		}

		if strings.HasPrefix(ref.RemoteRef, "refs/heads/") {
			sawBranchPush = true
		}

		if isMainBranch(ref.RemoteRef) {
			continue
		}

		if !checkRefForChangelog(opts, ref, result) {
			missingChangelog = true
		}
	}

	result.ChangelogUpdated = !missingChangelog
	return sawBranchPush
}

// skipRef returns true if this ref should be skipped (deleted or tag).
func skipRef(ref PushRef) bool {
	return ref.LocalSHA == ZeroSHA || strings.HasPrefix(ref.RemoteRef, "refs/tags/")
}

// isMainBranch returns true if the ref is main or master branch.
func isMainBranch(remoteRef string) bool {
	return remoteRef == "refs/heads/main" || remoteRef == "refs/heads/master"
}

// checkRefForChangelog checks if changelog was updated for a specific ref.
// Returns false if changelog is missing.
func checkRefForChangelog(opts PrePushCheckOptions, ref PushRef, result *CheckResult) bool {
	base := ref.RemoteSHA
	if base == ZeroSHA {
		base = FindDefaultBase(opts.GitOps, opts.RemoteName, ref.LocalSHA)
	}

	if base == "" {
		result.Errors = append(
			result.Errors,
			fmt.Sprintf("missing %s update for push to %s", ChangelogFile, ref.RemoteRef))
		return false
	}

	files, err := opts.GitOps.ChangedFiles(base, ref.LocalSHA)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to get changed files: %s", err))
		return false
	}

	if len(files) == 0 || !ContainsFile(files, ChangelogFile) {
		result.Errors = append(
			result.Errors,
			fmt.Sprintf("missing %s update for push to %s", ChangelogFile, ref.RemoteRef))
		return false
	}

	return true
}

// shouldSkipNextVersionCheck returns true if the next-version version check should be skipped.
func shouldSkipNextVersionCheck(optOut bool, sawBranchPush bool) bool {
	return optOut ||
		os.Getenv("STAVEFILE_SKIP_NEXTVER_CHANGELOG_CHECK") == "1" ||
		os.Getenv("GORELEASER_CURRENT_TAG") != "" ||
		os.Getenv("GORELEASER") != "" ||
		!sawBranchPush
}

// verifyNextVersionInChangelog checks that the next version to-be-released exists in the changelog.
func verifyNextVersionInChangelog(changelog *Changelog, path string, result *CheckResult) {
	nextVersion, err := NextVersion()
	if err != nil {
		result.Errors = append(
			result.Errors,
			fmt.Sprintf("failed to get next version: %s", err))
		return
	}

	switch {
	case !changelog.HasVersion(nextVersion):
		result.NextVersionPresent = false
		result.Errors = append(
			result.Errors,
			fmt.Sprintf("%s is missing release heading for version [%s]", path, nextVersion))
	case !changelog.HasLinkForVersion(nextVersion):
		result.NextVersionPresent = false
		result.Errors = append(
			result.Errors,
			fmt.Sprintf("%s is missing link reference for [%s]", path, nextVersion))
	default:
		result.NextVersionPresent = true
	}
}

// FindDefaultBase attempts to find the merge base with the default branch.
func FindDefaultBase(gitOps GitOps, remoteName, localSHA string) string {
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

	parsedChangelog, err := Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	result := parsedChangelog.Validate()
	if result.HasErrors() {
		return result.Error()
	}

	return nil
}

// VerifyNextVersion checks that the next version to-be-released exists in the changelog.
// This is a convenience function for the stavefile targets.
func VerifyNextVersion(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	parsedChangelog, err := Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	nextVersion, err := NextVersion()
	if err != nil {
		return fmt.Errorf("getting next version: %w", err)
	}

	if !parsedChangelog.HasVersion(nextVersion) {
		return fmt.Errorf("%s is missing release heading for version [%s]", path, nextVersion)
	}

	if !parsedChangelog.HasLinkForVersion(nextVersion) {
		return fmt.Errorf("%s is missing link reference for [%s]", path, nextVersion)
	}

	return nil
}
