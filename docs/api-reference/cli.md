# CLI Reference

Complete reference for the Stave command-line interface.

## Synopsis

```
stave [flags] [target] [args...]
```

## Description

Stave is a Go-native build tool. It parses Go files with the `stave` build tag, extracts exported functions as targets, and executes them.

When run without arguments and a default target is defined, Stave runs the default target. Otherwise, it lists available targets.

## Flags

### General Flags

#### -l, --list

List all available targets in the current directory.

```bash
stave -l
stave --list
```

#### -i, --info

Show detailed help for a specific target, including its full documentation and arguments.

```bash
stave -i build
stave --info build
```

#### -v, --verbose

Show verbose output when running targets. This displays command execution details and dependency progress.

```bash
stave -v build
stave --verbose build
```

#### -d, --debug

Enable debug messages for troubleshooting Stave itself.

```bash
stave -d build
stave --debug build
```

#### -f, --force

Force recreation of the compiled stavefile binary, bypassing the cache.

```bash
stave -f build
stave --force build
```

#### -t, --timeout

Set a timeout for running targets. Uses Go duration format.

```bash
stave -t 5m build
stave --timeout 30s test
stave -t 1h30m deploy
```

Duration format examples:
- `300ms` - 300 milliseconds
- `5s` - 5 seconds
- `5m` - 5 minutes
- `1h30m` - 1 hour 30 minutes

#### -C, --dir

Directory to read stavefiles from. Defaults to current directory.

```bash
stave -C /path/to/project build
stave --dir ./subproject test
```

#### -w, --workdir

Working directory where stavefiles will run. Defaults to the same as `--dir`.

```bash
stave -w /path/to/workdir build
stave --workdir ./output deploy
```

#### --version

Display version information and exit.

```bash
stave --version
```

#### --help

Display help information.

```bash
stave --help
stave -h build  # Help for specific target
```

### Build Flags

#### --keep

Keep the generated main file after compiling. Useful for debugging.

```bash
stave --keep build
ls stave_output_file.go  # Inspect generated code
```

#### --gocmd

Specify the Go command to use for compilation.

```bash
stave --gocmd go1.22 build
stave --gocmd /usr/local/go/bin/go test
```

Default: `go` or `STAVEFILE_GOCMD` environment variable.

### Dry-Run Mode

#### --dryrun

Print commands instead of executing them. Useful for testing what a target would do.

```bash
stave --dryrun build
```

Output:

```
DRYRUN: go build -o myapp .
```

Note: The stavefile is still compiled and executed; only `sh.Run*` commands within it are simulated.

### Compilation Flags

These flags are used when creating standalone binaries.

#### --compile

Output a static binary to the specified path instead of running the target.

```bash
stave --compile ./mybuild build
```

This creates a standalone binary that can be distributed and run without Stave installed.

#### --goos

Set GOOS for cross-compilation. Only valid with `--compile`.

```bash
stave --compile ./mybuild-linux --goos linux build
```

#### --goarch

Set GOARCH for cross-compilation. Only valid with `--compile`.

```bash
stave --compile ./mybuild-arm64 --goos linux --goarch arm64 build
```

#### --ldflags

Set ldflags for the compiled binary. Only valid with `--compile`.

```bash
stave --compile ./mybuild --ldflags "-s -w" build
```

### Initialization and Maintenance

#### --init

Create a starting stavefile template if no stave files exist.

```bash
stave --init
```

Creates `stavefile.go` with a basic template.

#### --clean

Remove all cached compiled binaries from the cache directory.

```bash
stave --clean
```

## Subcommands

### stave config

Manage Stave configuration.

```bash
stave config          # Show effective configuration
stave config init     # Create default config file
stave config show     # Same as 'stave config'
stave config path     # Show config file locations
```

See [Configuration](../user-guide/configuration.md) for details.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments or target not found |

Custom exit codes can be returned using `st.Fatal(code, message)` in stavefiles.

## Examples

### Basic Usage

```bash
# List targets
stave -l

# Run default target
stave

# Run specific target
stave build

# Run with arguments
stave deploy production
```

### Verbose and Debug

```bash
# See what's happening
stave -v build

# Debug Stave internals
stave -d build

# Both
stave -v -d build
```

### Cross-Platform Build

```bash
# Build for Linux
stave --compile dist/app-linux --goos linux --goarch amd64 build

# Build for macOS ARM
stave --compile dist/app-darwin --goos darwin --goarch arm64 build

# Build for Windows
stave --compile dist/app.exe --goos windows --goarch amd64 build
```

### CI/CD Pipeline

```bash
# Force rebuild in CI
stave -f test

# With timeout
stave -t 10m test

# Dry run to verify
stave --dryrun deploy
```

### Working with Different Directories

```bash
# Build from another directory
stave -C ./services/api build

# Run in specific working directory
stave -w ./output generate

# Combine both
stave -C ./build -w ./dist package
```

## Environment Variables

The following environment variables affect Stave behavior:

| Variable | Description |
|----------|-------------|
| `STAVEFILE_CACHE` | Cache directory for compiled binaries |
| `STAVEFILE_GOCMD` | Default Go command |
| `STAVEFILE_VERBOSE` | Enable verbose mode |
| `STAVEFILE_DEBUG` | Enable debug mode |
| `STAVEFILE_HASHFAST` | Use fast hashing |
| `STAVEFILE_IGNOREDEFAULT` | Ignore default target |
| `STAVEFILE_ENABLE_COLOR` | Enable colored output |
| `STAVEFILE_TARGET_COLOR` | ANSI color for targets |

See [Configuration](../user-guide/configuration.md) for details.

