# Stave

<!-- markdownlint-disable MD033 -->
<p align="center">
  <img src="../pics/stave-logo-251205.png" alt="Stave logo" width="200">
</p>
<!-- markdownlint-enable MD033 -->

Stave is a make-like build tool using Go. Write plain Go functions, and Stave automatically exposes them as runnable targets. It is a fork of [Mage](https://github.com/magefile/mage) with additional features including dry-run mode, XDG-compliant configuration, and parallelism control.

## Installation

```bash
go install github.com/yaklabco/stave@latest
```

## Quick Example

Create `stavefile.go`:

```go
//go:build stave

package main

import "github.com/yaklabco/stave/pkg/st"

// Build compiles the project.
func Build() error {
    st.Deps(Generate)
    return sh.Run("go", "build", "./...")
}

// Generate runs code generation.
func Generate() error {
    return sh.Run("go", "generate", "./...")
}
```

Run targets:

```bash
stave build    # Run Build target
stave -l       # List available targets
```

## Documentation

### Getting Started

- [Installation](getting-started/installation.md) - Install Stave via `go install` or from source
- [Quickstart](getting-started/quickstart.md) - Create your first stavefile
- [Migration from Mage](getting-started/migration-from-mage.md) - Guide for existing Mage users

### User Guide

- [Stavefiles](user-guide/stavefiles.md) - File conventions and build tags
- [Targets](user-guide/targets.md) - Defining target functions
- [Dependencies](user-guide/dependencies.md) - `st.Deps` and execution order
- [Namespaces](user-guide/namespaces.md) - Organizing targets with `st.Namespace`
- [Arguments](user-guide/arguments.md) - Typed target arguments
- [Configuration](user-guide/configuration.md) - Config files and environment variables
- [Shell Commands](user-guide/shell-commands.md) - Running external commands with `pkg/sh`
- [File Targets](user-guide/file-targets.md) - Incremental builds with `pkg/target`
- [Git Hooks](user-guide/hooks.md) - Native Git hook management
- [Advanced Topics](user-guide/advanced.md) - Cross-compilation, dry-run, CI, debugging

### API Reference

- [CLI Reference](api-reference/cli.md) - Command-line flags and subcommands
- [pkg/st](api-reference/st.md) - Dependency management and runtime utilities
- [pkg/sh](api-reference/sh.md) - Shell command execution
- [pkg/target](api-reference/target.md) - File modification time utilities

### Contributing

- [Development Setup](contributing/development.md) - Setting up a development environment
- [Architecture](contributing/architecture.md) - Codebase structure and design

## License

Apache License 2.0 - see [LICENSE](../LICENSE) for details.
