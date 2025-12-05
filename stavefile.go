//go:build stave

// This is the build script for Stave. The install target is all you really need.
// The release target is for generating official releases and is really only
// useful to project admins.
package main

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/yaklabco/stave/cmd/stave/version"
	"github.com/yaklabco/stave/internal/dryrun"
	"github.com/yaklabco/stave/pkg/sh"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/ui"
)

var Aliases = map[string]interface{}{
	"Speak": Say,
}

// Default target to run when none is specified.
var Default = All

func All() error {
	st.Deps(Init, Test)
	st.Deps(Build)

	return nil
}

// Init installs required tools and sets up git hooks and modules.
func Init() error { // stave:help=Install dev tools (Brewfile), setup husky hooks, and tidy modules
	// Install tools from Brewfile.
	if err := sh.Run("brew", "bundle", "--file=Brewfile"); err != nil {
		return err
	}

	// Install npm.
	if os.Getenv("CI") == "" {
		if err := sh.Run("npm", "ci"); err != nil {
			if err := sh.Run("npm", "install"); err != nil {
				return err
			}
		}
	} else {
		slog.Debug("in CI; skipping explicit npm installation")
	}

	// Set up husky git hooks.
	if err := sh.Run("git", "config", "core.hooksPath", ".husky"); err != nil {
		return err
	}
	if err := sh.Run("chmod", "+x", ".husky/pre-push"); err != nil {
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
		_, _ = fmt.Println(titleStyle.Render("golangci-lint output"))
		_, _ = fmt.Println(blockStyle.Render(out))
		_, _ = fmt.Println()
		return err
	}

	return nil
}

// Lint runs golangci-lint after markdownlint and init.
func Lint() { // stave:help=Run linters and auto-fix issues
	st.Deps(Init, Markdownlint, LintGo)
}

// Test aggregate target runs Lint and TestGo.
func Test() { // stave:help=Run lint and Go tests with coverage
	st.Deps(Init, Lint, TestGo)
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
	fmt.Printf("STAVE_NUM_PROCESSORS=%q\n", os.Getenv("STAVE_NUM_PROCESSORS"))
	fmt.Printf("GOMAXPROCS=%q\n", os.Getenv("GOMAXPROCS"))
}

// setSkipSVUChangelogCheck sets the SKIP_SVU_CHANGELOG_CHECK environment variable.
func setSkipSVUChangelogCheck() error {
	// Set SKIP_SVU_CHANGELOG_CHECK env var.
	return os.Setenv("SKIP_SVU_CHANGELOG_CHECK", "1")
}

// getRepoRoot returns the absolute path to the repository root (git top-level).
func getRepoRoot() (string, error) {
	out, err := sh.Output("git", "rev-parse", "--show-toplevel")
	if err != nil {
		slog.Warn("error running `git rev-parse --show-toplevel`", slog.Any("error", err))

		// Fallback to current working dir on failure
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		return cwd, nil
	}

	return strings.TrimSpace(out), nil
}

// Say says something.
func Say(msg string, i int, b bool, d time.Duration) error {
	_, err := fmt.Printf("%v(%T) %v(%T) %v(%T) %v(%T)\n", msg, msg, i, i, b, b, d, d)
	return err
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

var releaseTag = regexp.MustCompile(`^v1\.[0-9]+\.[0-9]+$`)

// origRelease generates a new release. Expects a version tag in v1.x.x format.
// It is the original `Release` target for Mage.
func origRelease(tag string) error {
	if _, err := exec.LookPath("goreleaser"); err != nil {
		return fmt.Errorf("can't find goreleaser: %w", err)
	}
	if !releaseTag.MatchString(tag) {
		return errors.New("TAG environment variable must be in semver v1.x.x format, but was " + tag)
	}

	if err := sh.RunV("git", "tag", "-a", tag, "-m", tag); err != nil {
		return err
	}
	if err := sh.RunV("git", "push", "origin", tag); err != nil {
		return err
	}
	var releaseErr error
	defer func() {
		if releaseErr != nil {
			_ = sh.RunV("git", "tag", "--delete", tag)            //nolint:errcheck // This is last-ditch cleanup.
			_ = sh.RunV("git", "push", "--delete", "origin", tag) //nolint:errcheck // This is last-ditch cleanup.
		}
	}()
	releaseErr = sh.RunV("goreleaser")

	return releaseErr
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
