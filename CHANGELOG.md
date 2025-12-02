# Changelog

## [2.1.0](https://github.com/yaklabco/stave/compare/stave-v2.0.0...stave-v2.1.0) (2025-12-02)


### Features

* add in dryrun functionality ([2bae053](https://github.com/yaklabco/stave/commit/2bae05363e33bd48d1b6d636f0e632b32baea278))
* allow mage:import alias to be defined for multiple imports ([#398](https://github.com/yaklabco/stave/issues/398)) ([0873a9b](https://github.com/yaklabco/stave/commit/0873a9b4e1c4b05bf27548b43a747005c2f8bd49))
* **config:** add XDG-compliant configuration system ([5222738](https://github.com/yaklabco/stave/commit/52227387ecbfbb0caf1c448b92128938e9269ab2))
* rename templated imports to avoid collisions ([#421](https://github.com/yaklabco/stave/issues/421)) ([40d421b](https://github.com/yaklabco/stave/commit/40d421b7a19a376f9bf5a244b9e6a221a298ff63))
* support trailing line comment for mage:import ([#480](https://github.com/yaklabco/stave/issues/480)) ([9f54e0f](https://github.com/yaklabco/stave/commit/9f54e0f83e2a8d2976c07037ad74aa20c62797a5))


### Bug Fixes

* 60 ([#61](https://github.com/yaklabco/stave/issues/61)) ([a12fd02](https://github.com/yaklabco/stave/commit/a12fd02e5be9f086fccabd7fcd456cf3964bd9af))
* **ci:** correct package-name from 'staff' to 'stave' ([7f24d1f](https://github.com/yaklabco/stave/commit/7f24d1ff62bd91e7dfb08cabf3a19f735d4e8f16))
* **ci:** correct package-name from 'staff' to 'stave' in release-please config ([e895522](https://github.com/yaklabco/stave/commit/e8955221c452e8cbfccf58fb8404a0cd0abdd6b4))
* deterministic compiled mainfile ([#348](https://github.com/yaklabco/stave/issues/348)) ([d9e2e41](https://github.com/yaklabco/stave/commit/d9e2e4152d4123fb9b1ee8d37b4d60277fa0bb0c))


### Code Refactoring

* fix linting issues in production code ([5394e9e](https://github.com/yaklabco/stave/commit/5394e9e21b5bcf53448a9ded4174fb46d15cd962))
* **listGoFiles:** remove go list dependency ([#440](https://github.com/yaklabco/stave/issues/440)) ([85ed9df](https://github.com/yaklabco/stave/commit/85ed9df31e011c531dfbfabe47fd1927bae17d0f))


### Tests

* fix failing tests ([03298a8](https://github.com/yaklabco/stave/commit/03298a8c22ed0c254b43b168ae67dcd001160885))
* fix linting issues in test files ([f5faa03](https://github.com/yaklabco/stave/commit/f5faa034acd59a9fbdb169b5f3d19e4ceee0592b))


### Continuous Integration

* add release-please automation ([3b5014a](https://github.com/yaklabco/stave/commit/3b5014aec4dd41ba5b0ddeeaffdc1089fd5cf361))
* add release-please automation ([df309a1](https://github.com/yaklabco/stave/commit/df309a16cc824018be0922debeda9f4e048fcd57))
* bump `golangci/golangci-lint-action` to `v9` ([e60e455](https://github.com/yaklabco/stave/commit/e60e4552b6aace677977882ef7b95c96637abe82))
* change ci.yml to get Go version from go.mod ([1c86d35](https://github.com/yaklabco/stave/commit/1c86d35d971196a44415068c2911522a257274cb))
* Extend `go-version` with `1.21.x` ([#479](https://github.com/yaklabco/stave/issues/479)) ([0fddccb](https://github.com/yaklabco/stave/commit/0fddccbc366b566bbb0ea972208c54f1066654df))
* migrate from travis to github action ([#391](https://github.com/yaklabco/stave/issues/391)) ([2f1ec40](https://github.com/yaklabco/stave/commit/2f1ec406dfa856a4b8378ef837061abc2a0ce01b))
