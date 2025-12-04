# Stave Documentation

Stave is a Go-native, make-like build tool. Write plain Go functions, and Stave automatically exposes them as runnable targets.

Stave is a fork of [Mage](https://github.com/magefile/mage) with additional features and improvements.

## Why Stave?

- **Pure Go** - No new syntax to learn. Your build scripts are just Go code.
- **No dependencies** - Stave compiles your build scripts into a binary with zero runtime dependencies.
- **Cross-platform** - Works on Linux, macOS, and Windows.
- **Fast** - Compiled binaries are cached for instant subsequent runs.
- **Parallel** - Dependencies run in parallel by default with `st.Deps()`.

## Quick Example

Create a `stavefile.go`:

```go
//go:build stave

package main

import (
    "fmt"
    "github.com/yaklabco/stave/pkg/sh"
    "github.com/yaklabco/stave/pkg/st"
)

// Build compiles the application.
func Build() error {
    st.Deps(Lint, Test)
    fmt.Println("Building...")
    return sh.Run("go", "build", "-o", "myapp", ".")
}

// Lint runs the linter.
func Lint() error {
    return sh.Run("golangci-lint", "run")
}

// Test runs the test suite.
func Test() error {
    return sh.Run("go", "test", "./...")
}
```

Then run:

```bash
stave build    # Runs Lint and Test in parallel, then Build
stave -l       # List all available targets
stave -h build # Show help for Build target
```

## Documentation Sections

### Getting Started

- [Installation](getting-started/installation.md) - Install Stave on your system
- [Quickstart](getting-started/quickstart.md) - Create your first stavefile
- [Migration from Mage](getting-started/migration-from-mage.md) - Guide for Mage users

### User Guide

- [Stavefiles](user-guide/stavefiles.md) - How stavefiles work
- [Targets](user-guide/targets.md) - Defining build targets
- [Dependencies](user-guide/dependencies.md) - Managing target dependencies
- [Namespaces](user-guide/namespaces.md) - Organizing targets into groups
- [Arguments](user-guide/arguments.md) - Passing arguments to targets
- [Configuration](user-guide/configuration.md) - Configuration files and environment variables
- [Shell Commands](user-guide/shell-commands.md) - Running external commands
- [File Targets](user-guide/file-targets.md) - Incremental builds

### API Reference

- [CLI Reference](api-reference/cli.md) - Command-line options
- [pkg/st](api-reference/st/index.md) - Core Stave API
- [pkg/sh](api-reference/sh/index.md) - Shell command helpers
- [pkg/target](api-reference/target/index.md) - File target utilities

### Contributing

- [Development Setup](contributing/development-setup.md) - Set up your development environment
- [Architecture](contributing/architecture.md) - Codebase overview
- [Testing](contributing/testing.md) - Running and writing tests
- [Pull Requests](contributing/pull-requests.md) - Contribution workflow

## Getting Help

- [GitHub Issues](https://github.com/yaklabco/stave/issues) - Report bugs or request features
- [GitHub Discussions](https://github.com/yaklabco/stave/discussions) - Ask questions

## License

Apache License 2.0 - see [LICENSE](../LICENSE) for details.

