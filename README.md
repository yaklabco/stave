# Stave

<!-- markdownlint-disable no-inline-html -->
<p align="center">
  <img src="img/stave-logo-251205B.png" alt="Stave logo" width="400" />
</p>
<!-- markdownlint-enable no-inline-html -->

A make-like build tool using Go. Write plain Go functions, and Stave will automatically use them as runnable targets.

Stave was forked from [mage](https://github.com/magefile/mage) (originally created by Nate Finch), and adds [various features and improvements](#differences-from-mage).

Documentation for `stave` is available [here](./docs/index.md).

<!-- markdownlint-disable no-inline-html no-multiple-blanks no-trailing-spaces -->
 

<p align="center">
  <a href="img/stave-l-in-goctx.png" target="_blank">
    <img src="img/stave-l-in-goctx.png" alt="`stave -l` output in the goctx project" width="67%" />
  </a>
  <br />
  <small>(screenshot: output of <code>stave -l</code> in the <a href="https://github.com/preminger/goctx">goctx</a> project)</small>
</p>

 
<!-- markdownlint-enable no-inline-html no-multiple-blanks no-trailing-spaces -->

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

You will, of course, need to install Go in one of the previous steps - for example, by using the [setup-go](https://github.com/actions/setup-go?tab=readme-ov-file#quick-start) GitHub action.

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
stave -h              # Show help
stave -h build # Show help for Build target
stave -v <target>     # Verbose mode
stave -t 5m <target>  # Set timeout
```

## Documentation

Full documentation is available in the [docs/](./docs/index.md) folder.

## Differences from `mage`

Stave was forked from [mage](https://magefile.org/), with the following goals (checked items are already implemented as of the latest release of `stave`).

For more details on any of these features, please see the [stave documentation](./docs/index.md).

- [x] Modernized Go patterns (Go 1.24+)
- [x] Additional shell helpers (`sh.Piper`, `sh.PiperWith`)
- [x] Watch-mode, to re-run build targets when watched files change
- [x] Dry-run support (print the command lines that would be executed, but don't run them)
- [x] Modernized CLI, using the wonderful tools developed by the folks at <https://github.com/charmbracelet>, including pretty-printed `-l`/`--list` output
- [x] Command-line completion of targets (via `stave completion <shell_name>`)
- [x] Automatic detection of circular dependencies in build targets
- [x] Support for native git-hooks management: no more need to use `husky` or other hooks-management tools; `stave` will manage your hooks for you, and you can specify stavefile targets directly as hooks
- [x] Public functions, exported under `pkg/changelog`, for automatically generating next version & next build-tag based on [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/), using [svu](https://github.com/caarlos0/svu) (included in `stave` as a module dependency; no need to install separately)
- [x] Public functions, also exported under `pkg/changelog`, for automatically enforcing [_keep-a-changelog_](https://keepachangelog.com/en/1.1.0/)-compliant CHANGELOG formatting; for enforcing that every push includes an update to the CHANGELOG; and for automatically linkifying your CHANGELOG headings to point to corresponding tag diffs (each can be used / not used separately from one another)

## Attribution

This project is a fork of [mage](https://github.com/magefile/mage), originally created by Nate Finch.
Licensed under the Apache License 2.0.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
