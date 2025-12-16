# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.3] - 2025-12-16

### Removed

- Extra printing of errors in main.go (`ExecuteWithFang(...)` already pretty-prints error; eliminates duplicate error printing).

## [0.3.2] - 2025-12-16

### Changed

- Organized targets in project's own stavefile.go using namespaces.

### Fixed

- Remove extra padding added to SYNOPSIS header (added in word-wrapping feature) to prevent line overflow.

## [0.3.1] - 2025-12-15

### Added

- Word-wrapping in stave `-l`/`--list` output.

## [0.3.0] - 2025-12-15

### Added

- Enhanced `stave -l` output with Lipgloss styling and table formatting.

- Color auto-detection via `st.ColorEnabled()` respecting `NO_COLOR` standard.

### Changed

- List output (`-l`) now handled by stave binary, not compiled output.

### Removed

- Unused `st.EnableColor()` function (use `st.ColorEnabled()` instead).

- Dead list code from compiled mainfile template.

## [0.2.8] - 2025-12-12

### Changed

- Fix some inaccuracies in CLI usage strings.

## [0.2.7] - 2025-12-12

### Changed

- Docs: more updates & improvements to documentation.

## [0.2.6] - 2025-12-11

### Changed

- Lots of updates to documentation. See [docs/index.md](./docs/index.md) and links therein.

- Bump all updatable Go dependencies to their latest versions as of this date.

## [0.2.5] - 2025-12-10

### Added

- Section on using `stave` in CI has been added to [the README](./README.md#using-stave-in-your-ci).

## [0.2.4] - 2025-12-10

### Added

- Installation instructions for installing via Homebrew.

### Changed

- Simplified how env vars are nullified in `TestGo` build target.

## [0.2.3] - 2025-12-10

### Changed

- Replace dependency on `goctx`'s `fsutils` with an "in-house" `fsutils`.

## [0.2.2] - 2025-12-10

### Added

- Added `changelog.NextTag()`, which returns the next version prefixed with "v" (in contrast to `changelog.NextVersion()`, which strips the `v`).

## [0.2.1] - 2025-12-09

### Fixed

- Maintenance release to ensure proper propagation to `sum.golang.org`.

## [0.2.0] - 2025-12-09

### Added

- New `sh.Piper(...)` and `sh.PiperWith(...)` functions.

### Changed

- Refactored `internal/env` -> `pkg/env` to expose `env` functions publicly.

## [0.1.3] - 2025-12-09

### Changed

- When calculating the next version to-be-released, call `svu` code programmatically instead of running the executable.

## [0.1.2] - 2025-12-08

### Changed

- Upgraded `caarlos0/svu` to `v3`, and removed deprecated `--force-patch-increment` flag from all its invocations.

## [0.1.1] - 2025-12-08

### Changed

- Drop minimum Go version to `1.24.11` (was: `1.25.4`) (by consuming `v0.14.3` of `goctx` instead of the older `v0.14.2`, which, despite being older, had a _higher_ minimum Go version requirement).

## [0.1.0] - 2025-12-08

### Added

- Git hooks management. Stave can manage your git hooks, implementing both native hooks management, and `husky`-based hooks management for support of legacy projects. See [docs/user-guide/hooks.md](./docs/user-guide/hooks.md) for details.

- Public Go functions, exported as `pkg/changelog`, for automatically enforcing [_keep-a-changelog_](https://keepachangelog.com/en/1.1.0/)-compliant CHANGELOG formatting; and, separately, for enforcing that every push includes an update to the CHANGELOG (each can be used / not used separately from one another). Also, `changelog.NextVersion()`, which automatically calculates next release version based on [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#summary).

- `--dryrun` mode.

- Automated detection of circular dependencies among stavefile targets.

- Detailed API reference documentation as well as an architecture overview for contributors.

- Pretty-printed debug logs, both in "outer" Stave execution _and_ in execution of compiled stavefile.

- `--exec` flag to execute arbitrary command-lines under Stave.

- `CHANGELOG.md`! (And first formally-versioned release of Stave.)

### Changed

- Added parallelism-by-default to use of Go tools from inside Stave.

- Parallelized tests where possible, including locking mechanism to prevent parallel tests in same `testdata/(xyz/)` subdir.

[unreleased]: https://github.com/yaklabco/stave/compare/v0.3.3...HEAD
[0.3.3]: https://github.com/yaklabco/stave/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/yaklabco/stave/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/yaklabco/stave/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/yaklabco/stave/compare/v0.2.8...v0.3.0
[0.2.8]: https://github.com/yaklabco/stave/compare/v0.2.7...v0.2.8
[0.2.7]: https://github.com/yaklabco/stave/compare/v0.2.6...v0.2.7
[0.2.6]: https://github.com/yaklabco/stave/compare/v0.2.5...v0.2.6
[0.2.5]: https://github.com/yaklabco/stave/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/yaklabco/stave/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/yaklabco/stave/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/yaklabco/stave/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/yaklabco/stave/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/yaklabco/stave/compare/v0.1.3...v0.2.0
[0.1.3]: https://github.com/yaklabco/stave/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/yaklabco/stave/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/yaklabco/stave/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/yaklabco/stave/releases/tag/v0.1.0
