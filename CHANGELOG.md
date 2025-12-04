# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.1.0](https://github.com/yaklabco/stave/compare/stave-v2.0.0...stave-v2.1.0) (2025-12-04)


### Features

* add in dryrun functionality ([2bae053](https://github.com/yaklabco/stave/commit/2bae05363e33bd48d1b6d636f0e632b32baea278))
* allow mage:import alias to be defined for multiple imports ([#398](https://github.com/yaklabco/stave/issues/398)) ([0873a9b](https://github.com/yaklabco/stave/commit/0873a9b4e1c4b05bf27548b43a747005c2f8bd49))
* cobra-ify! ([ce22920](https://github.com/yaklabco/stave/commit/ce229203c142bd5f83d1e6450380e5059ffae3bf))
* **config:** add XDG-compliant configuration system ([5222738](https://github.com/yaklabco/stave/commit/52227387ecbfbb0caf1c448b92128938e9269ab2))
* cycle detection ([a02086a](https://github.com/yaklabco/stave/commit/a02086af84ce18562b1cc9c26d4c9bce994963a6))
* logging revamp ([0700a0f](https://github.com/yaklabco/stave/commit/0700a0fb9a13b12a6bf7dd73097d6314a4a3c3d8))
* modernize stave ([3aa9175](https://github.com/yaklabco/stave/commit/3aa917586ebcc6da0fc2f6869d7f0a86336d8b1e))
* rename templated imports to avoid collisions ([#421](https://github.com/yaklabco/stave/issues/421)) ([40d421b](https://github.com/yaklabco/stave/commit/40d421b7a19a376f9bf5a244b9e6a221a298ff63))
* support trailing line comment for mage:import ([#480](https://github.com/yaklabco/stave/issues/480)) ([9f54e0f](https://github.com/yaklabco/stave/commit/9f54e0f83e2a8d2976c07037ad74aa20c62797a5))
* transition `stringer` to use the `//go:generate` pattern ([946e922](https://github.com/yaklabco/stave/commit/946e922174c9df9a0278ade1c325a50df251473f))


### Bug Fixes

* 60 ([#61](https://github.com/yaklabco/stave/issues/61)) ([a12fd02](https://github.com/yaklabco/stave/commit/a12fd02e5be9f086fccabd7fcd456cf3964bd9af))
* **ci:** correct package-name from 'staff' to 'stave' ([7f24d1f](https://github.com/yaklabco/stave/commit/7f24d1ff62bd91e7dfb08cabf3a19f735d4e8f16))
* **ci:** correct package-name from 'staff' to 'stave' in release-please config ([e895522](https://github.com/yaklabco/stave/commit/e8955221c452e8cbfccf58fb8404a0cd0abdd6b4))
* correct some mistaken references to the repo URL ([eb2310b](https://github.com/yaklabco/stave/commit/eb2310bdab4de50036fa4e2eb7de44a60c2f02db))
* deterministic compiled mainfile ([#348](https://github.com/yaklabco/stave/issues/348)) ([d9e2e41](https://github.com/yaklabco/stave/commit/d9e2e4152d4123fb9b1ee8d37b4d60277fa0bb0c))
* metadata injection in `install` stavefile target ([9ab5a3e](https://github.com/yaklabco/stave/commit/9ab5a3e95e5b18da9f3402a3e993acb4dc315626))
* replace `lowerFirstWord(...)` with idiomatic, regex-free implementation ([8f9fcf0](https://github.com/yaklabco/stave/commit/8f9fcf00b0fdc02b77a8c9210efbd6e49276986f))
* various issues in the interaction log logging, unit-tests, and verbosity-meets-debugging ([82a3ca8](https://github.com/yaklabco/stave/commit/82a3ca87ae3d22e229523fcfcd03592153d18e5b))


### Code Refactoring

* fix linting issues in production code ([5394e9e](https://github.com/yaklabco/stave/commit/5394e9e21b5bcf53448a9ded4174fb46d15cd962))
* **listGoFiles:** remove go list dependency ([#440](https://github.com/yaklabco/stave/issues/440)) ([85ed9df](https://github.com/yaklabco/stave/commit/85ed9df31e011c531dfbfabe47fd1927bae17d0f))


### Tests

* `stave_out` -&gt; `stave_test_out` (+ add to .gitignore) ([11c3948](https://github.com/yaklabco/stave/commit/11c39489b3f83bb00b709cb2a52f5cd5664e3a4e))
* fix failing tests ([03298a8](https://github.com/yaklabco/stave/commit/03298a8c22ed0c254b43b168ae67dcd001160885))
* fix linting issues in test files ([f5faa03](https://github.com/yaklabco/stave/commit/f5faa034acd59a9fbdb169b5f3d19e4ceee0592b))
* parallelize ([f9d01f6](https://github.com/yaklabco/stave/commit/f9d01f658a77ce4db89d8741c28518030b453174))


### Continuous Integration

* add release-please automation ([3b5014a](https://github.com/yaklabco/stave/commit/3b5014aec4dd41ba5b0ddeeaffdc1089fd5cf361))
* add release-please automation ([df309a1](https://github.com/yaklabco/stave/commit/df309a16cc824018be0922debeda9f4e048fcd57))
* bring in CI from `goctx` with the fancy caching ([5572597](https://github.com/yaklabco/stave/commit/557259788504c29dadbcdab3d478a1040f22b2fa))
* bump `golangci/golangci-lint-action` to `v9` ([e60e455](https://github.com/yaklabco/stave/commit/e60e4552b6aace677977882ef7b95c96637abe82))
* change ci.yml to get Go version from go.mod ([1c86d35](https://github.com/yaklabco/stave/commit/1c86d35d971196a44415068c2911522a257274cb))
* Extend `go-version` with `1.21.x` ([#479](https://github.com/yaklabco/stave/issues/479)) ([0fddccb](https://github.com/yaklabco/stave/commit/0fddccbc366b566bbb0ea972208c54f1066654df))
* migrate from travis to github action ([#391](https://github.com/yaklabco/stave/issues/391)) ([2f1ec40](https://github.com/yaklabco/stave/commit/2f1ec406dfa856a4b8378ef837061abc2a0ce01b))
* modify workflows for self-install ([879fd6b](https://github.com/yaklabco/stave/commit/879fd6b0b65cfa0880d1ae36e032c559906ad07d))
* move to new parallelism handling ([b974e5b](https://github.com/yaklabco/stave/commit/b974e5b055ee65edab716952e776f6c7e48c1979))
* remove now-unnecessary hack in `Self-install` step ([5e2c81d](https://github.com/yaklabco/stave/commit/5e2c81da0738b4041cae2fb7adcdf19c9105d5c4))

## [Unreleased]

## [0.1.0] - 2025-12-02

### Added

- `--dryrun` mode.

- Automated detection of circular dependencies among stavefile targets.

- `CHANGELOG.md`! (And first formally-versioned release of Stave.)

[unreleased]: https://github.com/yaklabco/stave/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/yaklabco/stave/releases/tag/v0.1.0
