//go:build stave

// This is the build script for Stave. The install target is all you really need.
// The release target is for generating official releases and is really only
// useful to project admins.
package main

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/yaklabco/stave/cmd/stave/version"
	"github.com/yaklabco/stave/config"
	"github.com/yaklabco/stave/pkg/changelog"
	"github.com/yaklabco/stave/pkg/fsutils"
	"github.com/yaklabco/stave/pkg/sh"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/stave"
	"github.com/yaklabco/stave/pkg/stave/prettylog"
	"github.com/yaklabco/stave/pkg/ui"
	"github.com/yaklabco/stave/pkg/watch"
)

const (
	changelogFilename = "CHANGELOG.md"

	buildCacheDirName    = ".buildcache"
	releaseNotesFilename = "release-notes.md"
)

func init() {
	logHandler := prettylog.SetupPrettyLogger(os.Stdout)
	if st.Debug() {
		logHandler.SetLevel(log.DebugLevel)
	}
}

// *********************************************************************
// * Aliases maps target aliases to their implementations.
// * (This is a stave convention - stavefiles define this global to create target aliases.)
// *

var Aliases = map[string]any{
	"LCL":   Prep.LinkifyChangelog,
	"RN":    Prep.ReleaseNotes,
	"Speak": Debug.Say,
}

// *
// *********************************************************************

// *********************************************************************
// * Default target
// * (This is a stave convention - stavefiles define this global to set the default target.)
// *

var Default = All

// *
// *********************************************************************

// *********************************************************************
// * Default namespace
// *

// All runs init, test:all, and build in sequence
func All() error {
	st.Deps(Init, Test.All)
	st.Deps(Build)

	return nil
}

// Init installs required tools and sets up git hooks and modules
func Init() {
	st.Deps(Prereq.Brew, Setup.Hooks, Prereq.Go)
}

// Build builds artifacts via goreleaser snapshot build
func Build() error {
	st.Deps(Init)

	if err := sh.RunV("goreleaser", "check"); err != nil {
		return err
	}

	return sh.RunV("goreleaser", "--parallelism", numProcsAsString(), "build", "--snapshot", "--clean")
}

// Release tags the next version and runs goreleaser release
func Release() error {
	if err := releasePrepper(tagAndPush); err != nil {
		return err
	}

	return sh.Run("goreleaser", "--parallelism", numProcsAsString(), "release", "--clean", "--release-notes="+filepath.Join(buildCacheDirName, releaseNotesFilename))
}

// Snapshot is like Release except it runs locally and does not push a new tag;
// useful for debugging the release process.
func Snapshot() error {
	noopTaggingFunc := func(string) error { return nil }
	if err := releasePrepper(noopTaggingFunc); err != nil {
		return err
	}

	return sh.Run("goreleaser", "--parallelism", numProcsAsString(), "release", "--snapshot", "--clean", "--release-notes="+filepath.Join(buildCacheDirName, releaseNotesFilename))
}

// Install builds and installs stave to GOBIN with version info embedded
func Install() error {
	name := "stave"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	gocmd := st.GoCmd()
	// use GOBIN if set in the environment, otherwise fall back to first path
	// in GOPATH environment string
	bin, err := sh.Output(gocmd, "env", "GOBIN")
	if err != nil {
		return fmt.Errorf("can't determine GOBIN: %w", err)
	}
	if bin == "" {
		gopath, err := sh.Output(gocmd, "env", "GOPATH")
		if err != nil {
			return fmt.Errorf("can't determine GOPATH: %w", err)
		}
		paths := strings.Split(gopath, string([]rune{os.PathListSeparator}))
		bin = filepath.Join(paths[0], "bin")
	}
	// specifically don't mkdirall, if you have an invalid gopath in the first
	// place, that's not on us to fix.
	if err := os.Mkdir(bin, 0o700); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create %q: %w", bin, err)
	}
	path := filepath.Join(bin, name)

	// we use go build here because if someone built with go get, then `go
	// install` turns into a no-op, and `go install -a` fails on people's
	// machines that have go installed in a non-writeable directory (such as
	// normal OS installs in /usr/bin)
	return sh.RunV(gocmd, "build", "-o", path, "-ldflags="+flags(), "github.com/yaklabco/stave")
}

// Clean removes the dist directory created by goreleaser
func Clean() error {
	return sh.Rm("dist")
}

// *
// * Default namespace
// *********************************************************************

// *********************************************************************
// * Prereq namespace
// *

type Prereq st.Namespace

// Go tidies modules and runs go generate
func (Prereq) Go() error {
	st.Deps(Prereq.Brew)

	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return err
	}

	if err := sh.Run("go", "generate", "./..."); err != nil {
		return err
	}

	return sh.Run("go", "mod", "tidy")
}

// Brew installs tools from Brewfile via Homebrew
func (Prereq) Brew() error {
	return sh.Run("brew", "bundle", "--file=Brewfile")
}

// *
// * Prereq namespace
// *********************************************************************

// *********************************************************************
// * Setup namespace
// *

type Setup st.Namespace

// Hooks configures git hooks to use stave targets
func (Setup) Hooks() error {
	st.Deps(Prereq.Brew)

	cs := ui.GetFangScheme()
	successStyle := lipgloss.NewStyle().Foreground(cs.Flag)
	labelStyle := lipgloss.NewStyle().Foreground(cs.Base)
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(cs.Program)

	// Ensure stave.yaml exists with hooks config
	if err := ensureStaveYAML(); err != nil {
		return err
	}

	// Install stave hooks
	if err := sh.Run("stave", "--hooks", "install"); err != nil {
		return fmt.Errorf("failed to install stave hooks: %w", err)
	}

	// Get configured hooks from config
	configuredHooks := findStaveHooks()
	hooksSuffix := ""
	if len(configuredHooks) > 0 {
		hooksSuffix = " (" + strings.Join(configuredHooks, ", ") + ")"
	}

	outputf("%s %s %s%s\n",
		successStyle.Render("⚙️"),
		labelStyle.Render("Git hooks configured:"),
		valueStyle.Render("Stave"),
		hooksSuffix,
	)
	if st.Verbose() {
		outputf("  %s %s\n", labelStyle.Render("Directory:"), valueStyle.Render(filepath.Join(".git", "hooks")+string(filepath.Separator)))
		outputf("  %s %s\n", labelStyle.Render("Config:"), valueStyle.Render("stave.yaml"))
	}
	return nil
}

// *
// * Setup namespace
// *********************************************************************

// *********************************************************************
// * Lint namespace
// *

type Lint st.Namespace

// Default equivalent to lint:all
func (Lint) Default() error {
	st.Deps(Lint.All)

	return nil
}

// All runs lint:go after lint:markdown and init
func (Lint) All() error {
	st.Deps(Init, Lint.Markdown, Lint.Go)

	return nil
}

// Markdown runs markdownlint-cli2 on all tracked Markdown files
func (Lint) Markdown() error {
	st.Deps(Init)

	markdownFilesList, err := sh.Output("git", "ls-files", "--cached", "--others", "--exclude-standard", "--", "*.md")
	if err != nil {
		return err
	}

	markdownFilesList = strings.TrimSpace(markdownFilesList)
	if markdownFilesList == "" {
		slog.Info("No Markdown files found to lint. Skipping.")
		return nil
	}

	files := lo.Filter(strings.Split(markdownFilesList, "\n"), func(s string, _ int) bool {
		return !lo.IsEmpty(s)
	})

	return sh.Run("markdownlint-cli2", files...)
}

// Go runs golangci-lint with auto-fix enabled
func (Lint) Go() error {
	st.Deps(Init)

	args := []string{"run", "--allow-parallel-runners", "--build-tags='!ignore'", "--fix"}

	_ = sh.Run("golangci-lint", args...) //nolint:errcheck // Intentional; re-run without `--fix` on next line.

	out, err := sh.Output("golangci-lint", lo.Slice(args, 0, len(args)-1)...)
	if err != nil {
		titleStyle, blockStyle := ui.GetBlockStyles()
		outputln(titleStyle.Render("golangci-lint output"))
		outputln(blockStyle.Render(out))
		outputln("")
		return err
	}

	slog.Debug("golangci-lint completed successfully")

	return nil
}

// *
// * Lint namespace
// *********************************************************************

// *********************************************************************
// * Check namespace
// *

type Check st.Namespace

// Changelog validates the changelog's format against 'Keep a Changelog' conventions
func (Check) Changelog() error {
	if err := changelog.ValidateFile(changelogFilename); err != nil {
		return fmt.Errorf("changelog validation failed: %w", err)
	}

	slog.Info("changelog validation passed")

	return nil
}

// GitStateClean checks that the git state is clean, i.e., that there are no
// uncommitted changes or untracked files.
func (Check) GitStateClean() error {
	out, err := sh.Output("git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check git state: %w", err)
	}

	if strings.TrimSpace(out) != "" {
		return fmt.Errorf("git state is not clean; do you have uncommitted changes or untracked files?\n%s", out)
	}

	return nil
}

// PrePush runs pre-push validations including changelog checks
func (Check) PrePush(remoteName, _remoteURL string) error {
	st.Deps(Prep.LinkifyChangelog)
	st.Deps(Lint.All)
	st.Deps(Check.GitStateClean)

	pushRefs, err := changelog.ReadPushRefs(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read push refs: %w", err)
	}

	if len(pushRefs) == 0 {
		slog.Warn("no refs pushed, skipping changelog pre-push check")
		return nil
	}

	slog.Info("about to run changelog pre-push check", slog.String("remote_name", remoteName), slog.Any("push_refs", pushRefs))
	result, err := changelog.PrePushCheck(changelog.PrePushCheckOptions{
		RemoteName:    remoteName,
		ChangelogPath: changelogFilename,
		Refs:          pushRefs,
	})
	if err != nil {
		return fmt.Errorf("changelog pre-push check failed: %w", err)
	}

	if result.HasErrors() {
		return fmt.Errorf("changelog pre-push check failed: %s", result.Errors)
	}

	if !result.ChangelogValid {
		return errors.New("changelog pre-push check failed: changelog is not valid")
	}

	if !result.ChangelogUpdated {
		return errors.New("changelog pre-push check failed: changelog has not been updated")
	}

	slog.Info("changelog next-version verification passed")

	return nil
}

// *
// * Check namespace
// *********************************************************************

// *********************************************************************
// * Prep namespace
// *

type Prep st.Namespace

// LinkifyChangelog ensures that heading links in changelog have Link Reference Definitions
func (Prep) LinkifyChangelog() error {
	if err := changelog.Linkify(changelogFilename); err != nil {
		return fmt.Errorf("changelog linkification failed: %w", err)
	}

	slog.Info("changelog linkification complete")

	return nil
}

// ReleaseNotes generates release notes from the changelog.
func (Prep) ReleaseNotes() error {
	st.Deps(Prep.BuildCacheDir)

	err := changelog.ExtractSection(changelogFilename, filepath.Join(buildCacheDirName, releaseNotesFilename), "")
	if err != nil {
		return fmt.Errorf("failed to extract release notes: %w", err)
	}

	return nil
}

// BuildCacheDir creates the build cache directory.
func (Prep) BuildCacheDir() error {
	err := os.MkdirAll(buildCacheDirName, 0o700)
	if err != nil {
		return fmt.Errorf("failed to create .buildcache directory: %w", err)
	}

	return nil
}

// *
// * Prep namespace
// *********************************************************************

// *********************************************************************
// * Test namespace
// *

type Test st.Namespace

// Default equivalent to test:all
func (Test) Default() error {
	st.Deps(Test.All)

	return nil
}

// All aggregate target runs lint:all and test:go
func (Test) All() error {
	// Run Init first (handles setup messages like hooks configured)
	st.Deps(Init)

	// Print test header (unless in quiet/CI mode)
	if !isQuietMode() {
		outputln("🧪 Running tests...")
	}

	startTime := time.Now()

	st.Deps(Lint.All, Test.Go)

	// Print success message with timing (unless in quiet/CI mode)
	if !isQuietMode() {
		outputf("👌 All tests ran successfully (%s)\n", time.Since(startTime).Round(time.Millisecond))
	}

	return nil
}

// Go runs Go tests with coverage and produces coverage.out and coverage.html
func (Test) Go() error {
	st.Deps(Init)

	nProcsStr := numProcsAsString()
	if err := sh.RunWithV(
		map[string]string{
			st.DryRunPossibleEnv:     "",
			stave.HooksAreRunningEnv: "",
		},
		"go", "tool", "gotestsum", "-f", "pkgname-and-test-fails",
		"--",
		"-v", "-p", nProcsStr, "-parallel", nProcsStr, "./...", "-count", "1",
		"-coverprofile=coverage.out", "-covermode=atomic",
	); err != nil {
		return err
	}

	return sh.Run("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
}

// *
// * Test namespace
// *********************************************************************

// *********************************************************************
// * Debug namespace
// *

type Debug st.Namespace

// Parallelism prints parallelism environment variables (debugging utility)
func (Debug) Parallelism() {
	outputf("STAVE_NUM_PROCESSORS=%q\n", os.Getenv("STAVE_NUM_PROCESSORS"))
	outputf("GOMAXPROCS=%q\n", os.Getenv("GOMAXPROCS"))
}

// DumpStdin reads stdin and dumps each line via spew (debugging utility)
func (Debug) DumpStdin() error {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		spew.Dump(line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading from stdin: %w", err)
	}

	return nil
}

// Say prints arguments with their types (example target demonstrating args)
func (Debug) Say(msg string, i int, b bool, d time.Duration) error {
	outputf("%v(%T) %v(%T) %v(%T) %v(%T)\n", msg, msg, i, i, b, b, d, d)

	return nil
}

// WatchFile watches the file specified in its single argument, and `cat`s its content any time it changes.
func (Debug) WatchFile(file string) {
	st.Deps(
		Build,
	)
	watch.Deps(
		Lint.Go,
		st.F(Debug.WatchDir, filepath.Dir(file)),
	)

	watch.Watch(file)

	contents := fsutils.MustRead(file)
	outputln(string(contents))
}

// WatchDir watches the directory specified in its single argument, and re-runs `ls` any time anything contained therein changes.
func (Debug) WatchDir(dir string) error {
	st.Deps(Build)
	watch.Deps(Lint.Go)

	watch.Watch(dir + "/**")

	output, err := sh.Output("ls", dir)
	if err != nil {
		return err
	}
	outputf("%s\n", output)

	return nil
}

// *
// * Debug namespace
// *********************************************************************

// *********************************************************************
// * utility functions
// *

// outputf writes a formatted string to stdout.
// Uses fmt.Fprintf for output (avoids forbidigo which bans fmt.Print* patterns).
func outputf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

// outputln writes a string to stdout with a trailing newline.
func outputln(s string) {
	_, _ = fmt.Fprintln(os.Stdout, s)
}

// isQuietMode returns true if output should be suppressed (CI environments).
// Check STAVE_QUIET=1 first, then common CI environment variables.
func isQuietMode() bool {
	if os.Getenv("STAVE_QUIET") == "1" {
		return true
	}
	// Common CI environment variables
	ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "CIRCLECI", "BUILDKITE"}
	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}

func flags() string {
	timestamp := time.Now().Format(time.RFC3339)
	theHash := hash()
	theTag := tag()
	if theTag == "" {
		theTag = "dev"
	}
	return fmt.Sprintf(
		`-X "github.com/yaklabco/stave/cmd/stave/version.BuildDate=%s"`+` `+
			`-X "github.com/yaklabco/stave/cmd/stave/version.Commit=%s"`+` `+
			`-X "github.com/yaklabco/stave/cmd/stave/version.Version=%s"`,
		timestamp, theHash, theTag,
	)
}

// tag returns the git tag for the current branch or "" if none.
func tag() string {
	// value, _ := sh.Output("git", "describe", "--tags")

	value := version.Version

	return value
}

// hash returns the git hash for the current repo or "" if none.
func hash() string {
	// value, _ := sh.Output("git", "rev-parse", "--short", "HEAD")

	value := version.Commit

	return value
}

// setSkipNextVerChangelogCheck sets the STAVEFILE_SKIP_NEXTVER_CHANGELOG_CHECK environment variable.
func setSkipNextVerChangelogCheck() error {
	// Set STAVEFILE_SKIP_NEXTVER_CHANGELOG_CHECK env var.
	return os.Setenv("STAVEFILE_SKIP_NEXTVER_CHANGELOG_CHECK", "1")
}

// hookSystem represents the active git hook system.
// findStaveHooks returns a list of hook names configured in stave.yaml.
func findStaveHooks() []string {
	cfg, err := config.Load(nil)
	if err != nil || cfg.Hooks == nil {
		return nil
	}
	return cfg.Hooks.HookNames()
}

// ensureStaveYAML creates stave.yaml with default hooks config if it doesn't exist.
func ensureStaveYAML() error {
	const staveYAML = "stave.yaml"

	// Check if file exists
	if _, err := os.Stat(staveYAML); err == nil {
		return nil
	}

	// Create default config
	const defaultConfig = `# Stave configuration
# See: https://github.com/yaklabco/stave

# Use hash_fast for faster hook execution (skips GOCACHE check)
hash_fast: true

# Git hooks configuration
hooks:
  pre-push:
    - target: Test
`
	const configFilePerm = 0o600
	if err := os.WriteFile(staveYAML, []byte(defaultConfig), configFilePerm); err != nil {
		return fmt.Errorf("failed to create stave.yaml: %w", err)
	}

	return nil
}

// releasePrepper prepares a release by generating release notes and tagging.
// It computes the next version tag using `changelog.NextTag` and invokes the
// provided `taggingFunc` with the computed tag. Dependencies such as release
// notes generation and initialization are handled internally.
func releasePrepper(taggingFunc func(nextTag string) error) error {
	st.Deps(Prep.ReleaseNotes)

	if err := setSkipNextVerChangelogCheck(); err != nil {
		return err
	}

	st.Deps(Init)

	nextTag, err := changelog.NextTag()
	if err != nil {
		return err
	}

	if !strings.HasPrefix(nextTag, "v") {
		nextTag = "v" + nextTag
	}

	slog.Info("computed next tag", slog.String("next_tag", nextTag))

	return taggingFunc(nextTag)
}

// tagAndPush creates a Git tag and pushes all tags to the remote repository.
// It takes the tag name as a string argument and returns an error if any command fails.
func tagAndPush(nextTag string) error {
	if err := sh.Run("git", "tag", nextTag); err != nil {
		return err
	}

	return sh.Run("git", "push", "--tags")
}

// numProcsAsString returns the number of processor cores as a string.
// It checks the environment variable "STAVE_NUM_PROCESSORS" or defaults to "1".
func numProcsAsString() string {
	return cmp.Or(os.Getenv("STAVE_NUM_PROCESSORS"), "1")
}

// *
// * utility functions
// *********************************************************************
