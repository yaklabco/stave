# Stave

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
- Dependency management with `mg.Deps()`
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

## Environment Variables

Stave supports the following environment variables (with backward compatibility for Mage's `MAGEFILE_*` equivalents):

| Variable | Description |
|----------|-------------|
| `STAVEFILE_VERBOSE` | Enable verbose output |
| `STAVEFILE_DEBUG` | Enable debug messages |
| `STAVEFILE_CACHE` | Custom cache directory (default: `~/.stavefile`) |
| `STAVEFILE_GOCMD` | Custom Go binary path |
| `STAVEFILE_HASHFAST` | Use fast hashing for rebuild detection |
| `STAVEFILE_IGNOREDEFAULT` | Ignore default target |
| `STAVEFILE_ENABLE_COLOR` | Enable colored output |
| `STAVEFILE_TARGET_COLOR` | ANSI color for target names |

## Naming Changes from Mage

| Mage | Stave |
|------|-------|
| `magefile.go` | `stavefile.go` |
| `magefiles/` directory | `stavefiles/` directory |
| `//go:build mage` | `//go:build stave` |
| `mage:import` | `stave:import` |
| `MAGEFILE_*` env vars | `STAVEFILE_*` env vars (with backward compat) |

## Attribution

This project is a fork of [Mage](https://github.com/magefile/mage), originally created by Nate Finch.
Licensed under the Apache License 2.0.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
