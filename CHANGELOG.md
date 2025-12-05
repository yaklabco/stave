# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-12-02

### Added

- `--dryrun` mode.

- Automated detection of circular dependencies among stavefile targets.

- Detailed API reference documentation as well as an architecture overview for contributors.

- Pretty-printed debug logs, both in "outer" Stave execution _and_ in execution of compiled stavefile.

- `CHANGELOG.md`! (And first formally-versioned release of Stave.)

### Changed

- Added parallelism-by-default to use of Go tools from inside Stave.

- Parallelized tests where possible, including locking mechanism to prevent parallel tests in same `testdata/(xyz/)` subdir.

[unreleased]: https://github.com/yaklabco/stave/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/yaklabco/stave/releases/tag/v0.1.0
