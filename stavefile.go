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
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/yaklabco/stave/cmd/stave/version"
	"github.com/yaklabco/stave/config"
	"github.com/yaklabco/stave/internal/changelog"
	"github.com/yaklabco/stave/internal/dryrun"
	"github.com/yaklabco/stave/pkg/sh"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/ui"
)

// outputf writes a formatted string to stdout.
// Uses fmt.Fprintf for output (avoids forbidigo which bans fmt.Print* patterns).
func outputf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

// outputln writes a string to stdout with a trailing newline.
func outputln(s string) {
	_, _ = fmt.Fprintln(os.Stdout, s)
}

// isQuietMode returns true if output should be suppressed (CI environments).
// Checks STAVE_QUIET=1 first, then common CI environment variables.
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

// Aliases maps target aliases to their implementations.
// This is a stave convention - stavefiles define this global to create target aliases.
//

var Aliases = map[string]interface{}{
	"Speak": Say,
}

// Default target to run when none is specified.
// This is a stave convention - stavefiles define this global to set the default target.
//

var Default = All

func All() error {
	st.Deps(Init, Test)
	st.Deps(Build)

	return nil
}

// Init installs required tools and sets up git hooks and modules.
func Init() error { // stave:help=Install dev tools (Brewfile), setup hooks (respects current choice), and tidy modules
	// Install tools from Brewfile.
	if err := sh.Run("brew", "bundle", "--file=Brewfile"); err != nil {
		return err
	}

	if err := setupHooksStave(); err != nil {
		return err
	}

	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return err
	}

	if err := sh.Run("go", "generate", "./..."); err != nil {
		return err
	}

	return sh.Run("go", "mod", "tidy")
}

func SwitchHooks() error { // stave:help=Switch git hooks system to stave
	return setupHooksStave()
}

// findHuskyHooks returns a list of hook names configured in .husky directory.
func findHuskyHooks() []string {
	knownHooks := []string{
		"pre-commit", "prepare-commit-msg", "commit-msg", "post-commit",
		"pre-push", "pre-rebase", "post-checkout", "post-merge",
	}
	var found []string
	for _, hook := range knownHooks {
		hookPath := filepath.Join(".husky", hook)
		if info, err := os.Stat(hookPath); err == nil && !info.IsDir() {
			found = append(found, hook)
		}
	}
	return found
}

// hookSystem represents the active git hook system.
type hookSystem string

const (
	hookSystemNone  hookSystem = "none"
	hookSystemHusky hookSystem = "husky"
	hookSystemStave hookSystem = "stave"
)

// detectActiveHookSystem determines which hook system is currently configured.
func detectActiveHookSystem() hookSystem {
	// Check if core.hooksPath is set to .husky
	hooksPath, err := sh.Output("git", "config", "--get", "core.hooksPath")
	if err == nil && strings.TrimSpace(hooksPath) == ".husky" {
		return hookSystemHusky
	}

	// Check if there are Stave-managed hooks in .git/hooks
	hooksDir := filepath.Join(".git", "hooks")
	for _, hook := range []string{"pre-commit", "pre-push", "commit-msg", "prepare-commit-msg"} {
		hookPath := filepath.Join(hooksDir, hook)
		if content, err := os.ReadFile(hookPath); err == nil {
			if strings.Contains(string(content), "Installed by Stave") {
				return hookSystemStave
			}
		}
	}

	return hookSystemNone
}

func setupHooksStave() error {
	cs := ui.GetFangScheme()
	successStyle := lipgloss.NewStyle().Foreground(cs.Flag)
	labelStyle := lipgloss.NewStyle().Foreground(cs.Base)
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(cs.Program)

	// Ensure stave.yaml exists with hooks config
	if err := ensureStaveYAML(); err != nil {
		return err
	}

	// Remove husky hooks path config (ignore error - may not be set)
	//nolint:errcheck // Intentionally ignoring - config key may not exist
	sh.Run("git", "config", "--unset", "core.hooksPath")

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

// Markdownlint runs markdownlint-cli2 on all tracked Markdown files.
func Markdownlint() error { // stave:help=Run markdownlint on Markdown files
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

// LintGo runs golangci-lint with auto-fix and parallel runner options enabled.
func LintGo() error {
	st.Deps(Init)
	out, err := sh.Output("golangci-lint", "run", "--fix", "--allow-parallel-runners", "--build-tags='!ignore'")
	if err != nil {
		titleStyle, blockStyle := ui.GetBlockStyles()
		outputln(titleStyle.Render("golangci-lint output"))
		outputln(blockStyle.Render(out))
		outputln("")
		return err
	}

	return nil
}

// Lint runs golangci-lint after markdownlint and init.
func Lint() { // stave:help=Run linters and auto-fix issues
	st.Deps(Init, Markdownlint, LintGo)
}

// Test aggregate target runs Lint and TestGo.
func Test() error { // stave:help=Run lint and Go tests with coverage
	// Run Init first (handles setup messages like hooks configured)
	st.Deps(Init)

	// Print test header (unless in quiet/CI mode)
	if !isQuietMode() {
		outputln("🧪 Running tests (Test: Lint, TestGo)")
	}

	startTime := time.Now()

	st.Deps(Lint, TestGo)

	// Print success message with timing (unless in quiet/CI mode)
	if !isQuietMode() {
		outputf("👌 All tests ran successfully (%s)\n", time.Since(startTime).Round(time.Millisecond))
	}

	return nil
}

// ValidateChangelog validates CHANGELOG.md format against 'Keep a Changelog' conventions.
func ValidateChangelog() error { // stave:help=Validate CHANGELOG.md format
	if err := changelog.ValidateFile("CHANGELOG.md"); err != nil {
		return fmt.Errorf("CHANGELOG.md validation failed: %w", err)
	}
	slog.Info("CHANGELOG.md validation passed")
	return nil
}

// DumpStdin reads lines from stdin and dumps them until stdin is closed.
// It uses spew.Dump() for output and returns an error if reading fails.
func DumpStdin() error {
	// Read lines from stdin and spew.Dump() them until stdin is closed.
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

// PrePushCheck runs all pre-push validations for branch pushes.
// This is the Go equivalent of the .githooks/pre-push bash script.
func PrePushCheck(remoteName, _remoteURL string) error { // stave:help=Run pre-push changelog validations
	pushRefs, err := changelog.ReadPushRefs(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read push refs: %w", err)
	}

	// Check that changelog has changed on current branch

	result, err := changelog.PrePushCheck(changelog.PrePushCheckOptions{
		RemoteName:    remoteName,
		ChangelogPath: "CHANGELOG.md",
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

	slog.Info("CHANGELOG.md next-version verification passed")

	return nil
}

// TestGo runs Go tests with coverage and produces coverage.out and coverage.html.
func TestGo() error { // stave:help=Run Go tests with coverage (coverage.out, coverage.html)
	st.Deps(Init)

	nCoresStr := cmp.Or(os.Getenv("STAVE_NUM_PROCESSORS"), "1")

	// Unset STAVEFILE_DRYRUN_POSSIBLE - which will be set by this point, normally -
	// so that tests *of* the dryrun functionality work as though they were run
	// from a bare `go test` command-line.
	if err := os.Unsetenv(dryrun.PossibleEnv); err != nil {
		return err
	}

	if err := sh.RunV(
		"go", "tool", "gotestsum", "-f", "pkgname-and-test-fails",
		"--",
		"-v", "-p", nCoresStr, "-parallel", nCoresStr, "./...", "-count", "1",
		"-coverprofile=coverage.out", "-covermode=atomic",
	); err != nil {
		return err
	}

	return sh.Run("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
}

// Build artifacts via goreleaser snapshot build.
func Build() error { // stave:help=Build artifacts using goreleaser (snapshot)
	st.Deps(Init)

	nCoresStr := cmp.Or(os.Getenv("STAVE_NUM_PROCESSORS"), "1")

	if err := sh.RunV("goreleaser", "check"); err != nil {
		return err
	}

	return sh.RunV("goreleaser", "--parallelism", nCoresStr, "build", "--snapshot", "--clean")
}

// Release tags the next version with svu and runs goreleaser release.
func Release() error { // stave:help=Create and push a new tag with svu, then goreleaser
	if err := setSkipSVUChangelogCheck(); err != nil {
		return err
	}

	st.Deps(Init)

	goBin, err := sh.Output("go", "env", "GOBIN")
	if err != nil {
		return err
	}
	goBin = strings.TrimSpace(goBin)

	if goBin == "" {
		goPath, err := sh.Output("go", "env", "GOPATH")
		if err != nil {
			return err
		}

		goBin = filepath.Join(strings.TrimSpace(goPath), "bin")
	}

	svuPath := filepath.Join(goBin, "svu")
	slog.Debug("svu binary path", slog.String("path", svuPath))
	nextVersion, err := sh.Output(svuPath, "next", "--force-patch-increment")
	if err != nil {
		return err
	}

	nextVersion = strings.TrimSpace(nextVersion)
	if nextVersion == "" {
		return errors.New("svu returned empty version")
	}

	slog.Info("computed next version", slog.String("version", nextVersion))

	if err := sh.Run("git", "tag", nextVersion); err != nil {
		return err
	}

	if err := sh.Run("git", "push", "--tags"); err != nil {
		return err
	}

	nCoresStr := cmp.Or(os.Getenv("STAVE_NUM_PROCESSORS"), "1")

	return sh.Run("goreleaser", "--parallelism", nCoresStr, "release", "--clean")
}

func ParallelismCheck() {
	outputf("STAVE_NUM_PROCESSORS=%q\n", os.Getenv("STAVE_NUM_PROCESSORS"))
	outputf("GOMAXPROCS=%q\n", os.Getenv("GOMAXPROCS"))
}

// setSkipSVUChangelogCheck sets the SKIP_SVU_CHANGELOG_CHECK environment variable.
func setSkipSVUChangelogCheck() error {
	// Set SKIP_SVU_CHANGELOG_CHECK env var.
	return os.Setenv("SKIP_SVU_CHANGELOG_CHECK", "1")
}

// Say says something.
func Say(msg string, i int, b bool, d time.Duration) error {
	outputf("%v(%T) %v(%T) %v(%T) %v(%T)\n", msg, msg, i, i, b, b, d, d)
	return nil
}

// Install runs "go install" for stave. This also generates version info for the binary.
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
	const binDirPerm = 0o700
	if err := os.Mkdir(bin, binDirPerm); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create %q: %w", bin, err)
	}
	path := filepath.Join(bin, name)

	// we use go build here because if someone built with go get, then `go
	// install` turns into a no-op, and `go install -a` fails on people's
	// machines that have go installed in a non-writeable directory (such as
	// normal OS installs in /usr/bin)
	return sh.RunV(gocmd, "build", "-o", path, "-ldflags="+flags(), "github.com/yaklabco/stave")
}

// Clean removes the temporarily generated files from Release.
func Clean() error {
	return sh.Rm("dist")
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
