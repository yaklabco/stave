# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

- `--dryrun` mode.

- Automated detection of circular dependencies among stavefile targets.

- Detailed API reference documentation as well as an architecture overview for contributors.

- Pretty-printed debug logs, both in "outer" Stave execution _and_ in execution of compiled stavefile.

- `--exec` flag to execute arbitrary command-lines under Stave.

- `CHANGELOG.md`! (And first formally-versioned release of Stave.)

### Changed

- Added parallelism-by-default to use of Go tools from inside Stave.

- Parallelized tests where possible, including locking mechanism to prevent parallel tests in same `testdata/(xyz/)` subdir.

[unreleased]: https://github.com/yaklabco/stave/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/yaklabco/stave/compare/v0.1.3...v0.2.0
[0.1.3]: https://github.com/yaklabco/stave/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/yaklabco/stave/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/yaklabco/stave/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/yaklabco/stave/releases/tag/v0.1.0
