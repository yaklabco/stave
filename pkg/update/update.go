package update

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/yaklabco/stave/config"
	"github.com/yaklabco/stave/pkg/env"
	"github.com/yaklabco/stave/pkg/ui"
)

// Params holds the parameters for an update check.
type Params struct {
	// CurrentVersion is the version of the running binary (e.g. "v0.12.0").
	CurrentVersion string
	// CacheDir is the directory where the update check cache file is stored.
	CacheDir string
	// Output is the writer where notification messages are written.
	Output io.Writer
	// Config holds the update check configuration (enabled flag and interval).
	Config config.UpdateCheckConfig
	// ClientOptions allows overriding the GitHub client (for testing).
	ClientOptions []GitHubClientOption
}

// CheckAndNotify performs a silent, best-effort update check after target execution.
// All errors are swallowed — this function never affects the user's workflow.
func CheckAndNotify(ctx context.Context, params Params) {
	if env.InCI() {
		return
	}

	if !params.Config.Enabled {
		return
	}

	if params.CurrentVersion == "" || params.CurrentVersion == "dev" {
		return
	}

	// Check cache first.
	cached := ReadCache(params.CacheDir)
	if cached != nil && !cached.IsExpired(params.Config.Interval) {
		// Cache is fresh — use cached data.
		if isNewer(cached.LatestVersion, params.CurrentVersion) && cached.NotifiedVersion != cached.LatestVersion {
			printNotification(params.Output, cached.LatestVersion, params.CurrentVersion)
			cached.NotifiedVersion = cached.LatestVersion
			_ = WriteCache(params.CacheDir, cached) //nolint:errcheck // Best-effort cache update.
		}

		return
	}

	// Cache is missing or expired — fetch from GitHub.
	release, err := fetchRelease(ctx, params.ClientOptions)
	if err != nil {
		return
	}

	entry := &CacheEntry{
		CheckedAt:     time.Now(),
		LatestVersion: release.TagName,
		ReleaseURL:    release.HTMLURL,
		ReleaseBody:   release.Body,
	}

	if isNewer(release.TagName, params.CurrentVersion) {
		printNotification(params.Output, release.TagName, params.CurrentVersion)
		entry.NotifiedVersion = release.TagName
	}

	_ = WriteCache(params.CacheDir, entry) //nolint:errcheck // Best-effort cache write.
}

// ExplicitCheck performs a version check that always fetches from GitHub.
// Unlike CheckAndNotify, errors are returned to the caller.
func ExplicitCheck(ctx context.Context, params Params) error {
	release, err := fetchRelease(ctx, params.ClientOptions)
	if err != nil {
		return err
	}

	// Update cache with fresh data.
	entry := &CacheEntry{
		CheckedAt:     time.Now(),
		LatestVersion: release.TagName,
		ReleaseURL:    release.HTMLURL,
		ReleaseBody:   release.Body,
	}
	_ = WriteCache(params.CacheDir, entry) //nolint:errcheck // Best-effort cache write.

	if isNewer(release.TagName, params.CurrentVersion) {
		printFullUpdateInfo(params.Output, params.CurrentVersion, release)
	} else {
		fmt.Fprintf(params.Output, "\n  You are running the latest version of stave (%s).\n", params.CurrentVersion)
	}

	return nil
}

// fetchRelease creates a GitHubClient with the given options and fetches the
// latest release.
func fetchRelease(ctx context.Context, opts []GitHubClientOption) (*Release, error) {
	client := NewGitHubClient(opts...)
	return client.FetchLatestRelease(ctx)
}

// isNewer returns true if latest is a newer semver than current.
// It returns false on any parse error.
func isNewer(latest, current string) bool {
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	latestVer, err := semver.NewVersion(latest)
	if err != nil {
		return false
	}

	currentVer, err := semver.NewVersion(current)
	if err != nil {
		return false
	}

	return latestVer.GreaterThan(currentVer)
}

// printNotification writes a one-line update notification to the given writer.
func printNotification(output io.Writer, latestVersion, currentVersion string) {
	scheme := ui.GetFangScheme()
	versionStyle := lipgloss.NewStyle().Foreground(scheme.QuotedString)
	baseStyle := lipgloss.NewStyle().Foreground(scheme.Base)
	cmdStyle := lipgloss.NewStyle().Foreground(scheme.Program)

	fmt.Fprintf(output, "\n  🎁 %s %s %s %s\n",
		versionStyle.Render("stave "+latestVersion+" available"),
		baseStyle.Render("(you have "+currentVersion+")"),
		baseStyle.Render("\u2014"),
		cmdStyle.Render("stave --check-update for details"),
	)
}

// printFullUpdateInfo writes a detailed update notification with changelog
// to the given writer.
func printFullUpdateInfo(output io.Writer, currentVersion string, release *Release) {
	titleStyle, _ := ui.GetBlockStyles()
	scheme := ui.GetFangScheme()
	versionStyle := lipgloss.NewStyle().Foreground(scheme.QuotedString)
	baseStyle := lipgloss.NewStyle().Foreground(scheme.Base)

	fmt.Fprintln(output)
	fmt.Fprintln(output, titleStyle.Render("🎁 Update Available"))
	fmt.Fprintf(output, "  %s  %s\n", versionStyle.Render(release.TagName), baseStyle.Render("(you have "+currentVersion+")"))
	fmt.Fprintln(output)

	if release.Body != "" {
		renderMarkdown(output, release.Body)
	}

	fmt.Fprintf(output, "  %s\n\n", release.HTMLURL)
}

// wordWrapWidth is the column width for glamour markdown rendering.
const wordWrapWidth = 80

// changelogStyle returns a glamour style based on terminal background,
// with heading prefixes (##, ###) removed for cleaner changelog output.
func changelogStyle() ansi.StyleConfig {
	style := styles.DarkStyleConfig
	if !lipgloss.HasDarkBackground(os.Stdin, os.Stdout) {
		style = styles.LightStyleConfig
	}

	// Remove markdown heading prefixes — they look noisy in a changelog.
	style.H2.Prefix = ""
	style.H3.Prefix = ""
	style.H4.Prefix = ""
	style.H5.Prefix = ""
	style.H6.Prefix = ""

	return style
}

// renderMarkdown renders markdown to styled terminal output via glamour.
// Falls back to raw text if rendering fails.
func renderMarkdown(output io.Writer, body string) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(changelogStyle()),
		glamour.WithWordWrap(wordWrapWidth),
	)
	if err != nil {
		fmt.Fprintln(output, strings.TrimSpace(body))
		fmt.Fprintln(output)

		return
	}

	rendered, err := renderer.Render(body)
	if err != nil {
		fmt.Fprintln(output, strings.TrimSpace(body))
		fmt.Fprintln(output)

		return
	}

	fmt.Fprint(output, rendered)
}
