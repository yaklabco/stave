# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Native Git hooks management via `stave hooks` CLI commands (`init`, `install`, `uninstall`, `list`, `run`).

- Declarative hooks configuration in `stave.yaml` with support for `target`, `args`, and `passStdin` options.

- CHANGELOG.md validation package (`internal/changelog`) implementing Keep a Changelog format validation in Go.

- `ValidateChangelog` and `PrePushCheck` targets in stavefile for pre-push changelog enforcement.

- `Hooks` target to switch between Husky and native Stave hooks systems.

- Hook system detection to preserve user's choice during `Init`.

- User init script support (`~/.config/stave/hooks/init.sh`) for PATH setup in GUI clients.

- Debug logging throughout hooks internals via `--debug` flag.

- Documentation for hooks in `docs/user-guide/hooks.md`.

### Changed

- Centralized `SimpleConsoleLogger` in `internal/log` package.

- Renamed `Staventa`/`BrightStaventa` to standard ANSI names `Magenta`/`BrightMagenta`.

- Exported `ExitStatuser` interface in `pkg/st` for cross-package use.

### Fixed

- Correct comment group length check in parser (`== 9` to `== 0`) that caused import tag parsing failures.

- Replace invalid `%#w` format verb with `%w` in error wrapping.

## [0.1.0] - 2025-12-02

### Added

- `--dryrun` mode.

- Automated detection of circular dependencies among stavefile targets.

- Detailed API reference documentation as well as an architecture overview for contributors.

- Pretty-printed debug logs, both in "outer" Stave execution _and_ in execution of compiled stavefile.

- `--exec` flag to execute arbitrary command-lines under Stave.

- `CHANGELOG.md`! (And first formally-versioned release of Stave.)

### Changed

- Added parallelism-by-default to use of Go tools from inside Stave.

- Parallelized tests where possible, including locking mechanism to prevent parallel tests in same `testdata/(xyz/)` subdir.

[unreleased]: https://github.com/yaklabco/stave/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/yaklabco/stave/releases/tag/v0.1.0
