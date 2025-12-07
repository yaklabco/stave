# Stave

<!-- markdownlint-disable MD033 -->
<p align="center">
  <img src="img/stave-logo-251205B.png" alt="Stave logo" width="200">
</p>
<!-- markdownlint-enable MD033 -->

A make-like build tool using Go. Write plain Go functions, Stave automatically uses them as runnable targets.

Stave is a fork of [Mage](https://github.com/magefile/mage) by Nate Finch, with additional features and improvements.

## Installation

```bash
go install github.com/yaklabco/stave@latest
```

Or build from source:

```bash
git clone https://github.com/yaklabco/stave.git
cd stave
go run bootstrap.go
```

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

## Differences from Stave

Stave is built on top of Stave with the following goals:

- [x] Modernized Go patterns (Go 1.21+)
- [ ] Additional shell helpers
- [ ] Watch mode for file changes
- [x] Dry-run support
- [x] Enhanced CLI experience
- [x] Automatic detection of circular dependencies in build targets

## Attribution

This project is a fork of [Mage](https://github.com/magefile/mage), originally created by Nate Finch.
Licensed under the Apache License 2.0.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
