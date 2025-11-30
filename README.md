# Staff

A make-like build tool using Go. Write plain Go functions, Staff automatically uses them as runnable targets.

Staff is a fork of [Mage](https://github.com/magefile/mage) by Nate Finch, with additional features and improvements.

## Installation

```bash
go install github.com/yaklabco/staff@latest
```

Or build from source:

```bash
git clone https://github.com/yaklabco/staff.git
cd staff
go run bootstrap.go
```

## Quick Start

Create a `magefile.go` in your project:

```go
//go:build mage

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
staff build    # Run the Build target
staff test     # Run the Test target
staff -l       # List all targets
staff -h build # Show help for Build target
```

## Features

- Write build tasks in Go - no new syntax to learn
- Dependency management with `mg.Deps()`
- Parallel and serial dependency execution
- Namespaced targets for organization
- Cross-platform (Linux, macOS, Windows)
- No external dependencies beyond Go

## Documentation

```bash
staff -h              # Show help
staff -l              # List targets
staff -v <target>     # Verbose mode
staff -t 5m <target>  # Set timeout
```

## Differences from Mage

Staff is built on top of Mage with the following goals:

- [ ] Modernized Go patterns (Go 1.21+)
- [ ] Additional shell helpers
- [ ] Watch mode for file changes
- [ ] Dry-run support
- [ ] Enhanced CLI experience

## Attribution

This project is a fork of [Mage](https://github.com/magefile/mage), originally created by Nate Finch.
Licensed under the Apache License 2.0.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
