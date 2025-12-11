# Stave

<!-- markdownlint-disable MD033 -->
<p align="center">
  <img src="img/stave-logo-251205B.png" alt="Stave logo" width="200">
</p>
<!-- markdownlint-enable MD033 -->

A make-like build tool using Go. Write plain Go functions, Stave automatically uses them as runnable targets.

Stave is a fork of [Mage](https://github.com/magefile/mage) by Nate Finch, with additional features and improvements.

## Installation

### Using Homebrew

```shell
brew tap yaklabco/tap
brew install stave
```

### Using `go install`

```bash
go install github.com/yaklabco/stave@latest
```

### Building & installing from source

```bash
git clone https://github.com/yaklabco/stave.git
cd stave
go run bootstrap.go
```

## Using `stave` in your CI

The most portable, cross-platform way to make `stave` available in your CI workflow is probably to install it via the [Go-based](#using-go-install) install method:

```yaml
    steps:
      # ... prev. steps clipped ...
      - name: Install stave
        run: go install github.com/yaklabco/stave@latest
```

You will, of course, need to install Go in one of the previous steps - for example, via the [setup-go](https://github.com/actions/setup-go?tab=readme-ov-file#quick-start) action.

## Quick Start

Create a `stavefile.go` in your project:

```go
//go:build stave

package main

import "fmt"

// Build compiles the project
func Build() error {
    fmt.Println("Building...")
    return nil
}

// Test runs the test suite
func Test() {
    fmt.Println("Testing...")
}
```

Then run:

```bash
stave build    # Run the Build target
stave test     # Run the Test target
stave -l       # List all targets
stave -h build # Show help for Build target
```

## Features

- Write build tasks in Go - no new syntax to learn
- Dependency management with `st.Deps()`
- Parallel and serial dependency execution
- Namespaced targets for organization
- Cross-platform (Linux, macOS, Windows)
- No external dependencies beyond Go

## Documentation

```bash
stave -h              # Show help
stave -l              # List targets
stave -v <target>     # Verbose mode
stave -t 5m <target>  # Set timeout
```

## Differences from Mage

Stave is built on top of [Mage](https://magefile.org/), with the following goals (checked items are already implemented as of latest release):

- [x] Modernized Go patterns (Go 1.21+)
- [x] Additional shell helpers (`sh.Piper`, `sh.PiperWith`)
- [ ] Watch mode for file changes
- [x] Dry-run support
- [x] Enhanced CLI experience
- [x] Automatic detection of circular dependencies in build targets
- [x] Public Go functions, exported under `pkg/changelog`, for automatically enforcing [_keep-a-changelog_](https://keepachangelog.com/en/1.1.0/)-compliant CHANGELOG formatting; and, separately, for enforcing that every push includes an update to the CHANGELOG (each can be used / not used separately from one another)
- [x] Public Go functions, also exported under `pkg/changelog`, for automatically generating next version & next build-tag based on [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/), using [svu](https://github.com/caarlos0/svu) (included in `stave` as a module dependency; no need to install separately)
- [x] Support for native git-hooks management: no more need to use `husky` or other hooks-management tools; `stave` will manage your hooks for you, and you can specify stavefile targets directly as hooks

## Attribution

This project is a fork of [Mage](https://github.com/magefile/mage), originally created by Nate Finch.
Licensed under the Apache License 2.0.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
