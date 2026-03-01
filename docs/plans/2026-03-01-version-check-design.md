# Version Check Design

**Date**: 2026-03-01
**Status**: Draft

## Problem

Stave users have no way to know when a new version is available or what it
contains. They must manually check GitHub releases or remember to run
`brew upgrade`. This means improvements and bug fixes go unnoticed.

## Solution

After every target execution, stave performs a non-blocking version check
against the GitHub Releases API. When a newer version exists, it prints a
single styled notification line. The notification appears once per new
version, then goes silent until the next release.

Users can also run `stave --check-update` to force a fresh check and view
the full changelog for the latest release.

## Design Principles

1. **Never slow down work.** The check runs after the target finishes, uses
   cached results, and has a short HTTP timeout.
2. **Never be noisy.** One notification per new version. No errors on
   network failure. Silent in CI.
3. **Respect the user.** Configurable interval, disableable, no telemetry.

## Behavior

### Auto-check (after target execution)

```
$ stave build
... build output ...

  stave v0.13.0 available (you have v0.12.0) — stave --check-update for details

$ stave test:all
... test output ...
(no notification — v0.13.0 was already shown)

$ stave lint:go
... lint output ...
(still silent — same version)
```

When v0.14.0 is released, the notification appears once more.

### Explicit check (`stave --check-update`)

```
$ stave --check-update

  UPDATE AVAILABLE

  v0.13.0  (you have v0.12.0)
  Released 2026-02-28

  ## [0.13.0] - 2026-02-28

  ### Added

  - Watch mode dependency graph visualization.
  - Shell completion for fish targets.

  ### Fixed

  - Race condition in parallel dep execution.

  https://github.com/yaklabco/stave/releases/tag/v0.13.0
```

The explicit command always fetches fresh (ignores cache TTL) and renders the
full changelog body with lipgloss styling using the existing Fang color scheme.

### When no update is available

```
$ stave --check-update

  stave v0.12.0 is the latest version.
```

### Suppression rules

The auto-check is **skipped** when any of these are true:

- Running in CI (detected via `CI`, `GITHUB_ACTIONS`, `GITLAB_CI`,
  `JENKINS_URL`, `CIRCLECI`, `BUILDKITE` environment variables)
- The `update_check.enabled` config is `false`
- The `STAVE_NO_UPDATE_CHECK` environment variable is set
- The cached result is fresh (within the configured interval)
- The cached result shows the latest version has already been notified
- The current version is `dev` (development build)

The `--check-update` flag bypasses all suppression except `enabled: false`.

## Data Source

**GitHub Releases API**: `GET https://api.github.com/repos/yaklabco/stave/releases/latest`

Single unauthenticated request. Returns:

```json
{
  "tag_name": "v0.13.0",
  "body": "## [0.13.0] - 2026-02-28\n\n### Added\n...",
  "published_at": "2026-02-28T18:23:30Z",
  "html_url": "https://github.com/yaklabco/stave/releases/tag/v0.13.0"
}
```

Rate limit: 60 requests/hour per IP (irrelevant with caching).

HTTP timeout: 3 seconds. Any error (network, timeout, non-200, parse
failure) is silently ignored for auto-checks.

## Cache

### Location

`{XDG_CACHE_HOME}/stave/update-check.json`

Uses the existing `config.ResolveXDGPaths().CacheDir()` infrastructure.
The cache file lives alongside the compiled binary cache.

### Schema

```json
{
  "latest_version": "0.13.0",
  "release_url": "https://github.com/yaklabco/stave/releases/tag/v0.13.0",
  "changelog_body": "## [0.13.0] - 2026-02-28\n\n### Added\n...",
  "published_at": "2026-02-28T18:23:30Z",
  "checked_at": "2026-03-01T10:15:00Z",
  "notified_version": "0.13.0"
}
```

| Field              | Purpose                                         |
| ------------------ | ----------------------------------------------- |
| `latest_version`   | Semver from the latest release tag               |
| `release_url`      | Link to the GitHub release page                  |
| `changelog_body`   | Raw markdown body from the release               |
| `published_at`     | When the release was published                   |
| `checked_at`       | When the last API call was made                  |
| `notified_version` | The version that was last shown to the user       |

### Cache logic

```
if now - checked_at < interval:
    use cached result
else:
    fetch from API, update cache

if latest_version > current_version AND latest_version != notified_version:
    show notification
    set notified_version = latest_version
```

## Configuration

### stave.yaml / config.yaml

```yaml
update_check:
  enabled: true         # default: true
  interval: 24h         # default: 24h, accepts Go duration strings
```

### Environment variable

`STAVE_NO_UPDATE_CHECK=1` disables the auto-check. This is the escape hatch
for environments where `stave.yaml` is not available.

### Config struct addition

```go
type UpdateCheckConfig struct {
    Enabled  bool          `mapstructure:"enabled"`
    Interval time.Duration `mapstructure:"interval"`
}
```

Added to the existing `Config` struct as:

```go
UpdateCheck UpdateCheckConfig `mapstructure:"update_check"`
```

Defaults: `Enabled: true`, `Interval: 24 * time.Hour`.

## Package Layout

### New package: `pkg/update`

```
pkg/update/
    update.go       # Core logic: Check(), Notify(), cache read/write
    update_test.go  # Tests with HTTP test server
    github.go       # GitHub API client (minimal, single-purpose)
    github_test.go  # GitHub client tests
    cache.go        # Cache file operations
    cache_test.go   # Cache tests
```

**Why a new package?** This is a self-contained feature with its own I/O
(network + filesystem), its own data types, and no dependency on the
compilation pipeline. It belongs alongside `pkg/changelog` and `pkg/target`
as a peer utility package.

**Dependencies from within stave:**
- `config` (for cache dir path, update check config, CI detection)
- `cmd/stave/version` (for current version)
- `pkg/ui` (for Fang color scheme)

**External dependencies:**
- `net/http` (stdlib)
- `encoding/json` (stdlib)
- `github.com/Masterminds/semver/v3` (already a transitive dep, needs
  direct import in go.mod)

### CI detection

Add `InCI() bool` to the `pkg/env` package. This extracts the CI detection
logic from `stavefile.go` into a reusable location. The stavefile's
`isQuietMode()` can then call `env.InCI()` as well.

## Integration Point

### Where the check runs

In `pkg/stave/main.go`, after the target binary has finished executing.
The check runs synchronously but is fast due to caching. The flow:

```
target execution completes
    |
    v
update.CheckAndNotify(ctx, currentVersion, stderr)
    |
    +-- load config
    +-- if disabled or CI: return
    +-- read cache
    +-- if cache fresh: use cached
    +-- else: fetch (3s timeout), update cache
    +-- if newer version exists and not yet notified:
    |       print one-liner to stderr
    |       update notified_version in cache
    +-- return (all errors swallowed)
```

### Where `--check-update` runs

In `cmd/stave/stave.go`, as a pseudo-flag handler alongside `--config`,
`--hooks`, and `--init`. When `--check-update` is set, it calls
`update.ExplicitCheck(ctx, currentVersion, stdout)` which:

1. Fetches fresh from GitHub (ignores cache TTL)
2. Updates the cache
3. Renders the full changelog with lipgloss styling
4. Returns exit code 0

## Output Styling

### One-liner (auto-check, stderr)

Uses the existing Fang color scheme:
- Version numbers: `colorScheme.QuotedString` (same as version display)
- Arrow/separator: `colorScheme.Base`
- Command hint: `colorScheme.Program`

Indented with 2-space left margin for visual separation from target output.

### Full changelog (--check-update, stdout)

Uses `ui.GetBlockStyles()` for the title and `ui.GetFangScheme()` for
inline styling. The changelog body is rendered as-is (it's already
markdown from the release). The release URL is appended at the bottom.

### Respects NO_COLOR

All styling checks `NO_COLOR` and `STAVEFILE_ENABLE_COLOR` via the
existing color detection path.

## Error Handling

| Scenario                  | Behavior                        |
| ------------------------- | ------------------------------- |
| No network                | Silent skip (auto), error (explicit) |
| GitHub API error           | Silent skip (auto), error (explicit) |
| Malformed JSON response    | Silent skip (auto), error (explicit) |
| Cache file unreadable      | Treat as cold cache, fetch fresh |
| Cache file unwritable      | Skip cache update, still show notification |
| Cache dir doesn't exist    | Create it (already exists for binary cache) |
| Current version is "dev"   | Skip auto-check entirely         |
| Semver parse failure       | Silent skip                      |
| HTTP timeout (3s)          | Silent skip (auto), error (explicit) |

The explicit check (`--check-update`) surfaces errors because the user
explicitly asked for the check. Auto-check never surfaces errors.

## Testing Strategy

### Unit tests

- **Cache**: Read/write/expiry logic with temp files
- **GitHub client**: HTTP test server returning various responses
  (success, error, malformed, timeout)
- **Version comparison**: Current vs. latest, equal, newer, older, dev builds
- **Notification logic**: First time, already notified, new version after
  notified, cache expired
- **CI detection**: Various env var combinations
- **Config integration**: Enabled/disabled, custom intervals, env override

### Integration test

- End-to-end test using a mock HTTP server that verifies the full flow:
  cold cache -> fetch -> notify -> warm cache -> skip -> new version -> notify

### What we do NOT test

- Actual GitHub API calls (flaky, rate-limited)
- Lipgloss rendering (visual, tested by charmbracelet)

## Pressure-Testing Assumptions

### "The GitHub Releases API is stable and sufficient"

**Risk**: GitHub changes the API or rate-limits more aggressively.
**Mitigation**: We only use `tag_name`, `body`, `published_at`, and
`html_url` - the most stable fields. The 60 req/hr limit with 24h caching
means a single user makes ~1 request/day. Even in a team of 60 sharing an
IP, that's well within limits.

### "Semver comparison is straightforward"

**Risk**: Pre-release versions (e.g., `v0.13.0-rc.1`), non-semver tags.
**Mitigation**: Use `Masterminds/semver` which handles pre-release correctly.
If parsing fails, skip silently. Pre-release versions from GitHub should
NOT trigger notifications (they're not "latest" releases).

### "The cache file won't cause issues"

**Risk**: Concurrent stave processes writing the cache simultaneously.
**Mitigation**: Use atomic write (write to temp file, rename). JSON is
small (~1KB). Last writer wins is acceptable - worst case, a notification
shows twice.

### "3-second timeout is appropriate"

**Risk**: Slow networks cause 3-second delays after every target.
**Mitigation**: The check only fires once per interval (default 24h).
A 3-second delay once per day is imperceptible. For truly constrained
environments, disable via config.

### "Printing to stderr is correct for the auto-notification"

**Risk**: Some tools parse stave's stdout. Notifications on stderr won't
interfere. But some users redirect stderr.
**Mitigation**: Stderr is the correct channel for advisory output that
isn't part of the command's primary output. This matches `gh`, `brew`,
and other well-behaved CLI tools.

### "One notification per version is the right frequency"

**Risk**: User sees the notification, forgets about it, never updates.
**Mitigation**: The `--check-update` command is mentioned in the one-liner.
We don't nag - that's a feature. Users who want reminders can set a shorter
interval or clear the cache.
