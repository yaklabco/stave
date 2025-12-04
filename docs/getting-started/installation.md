# Installation

This guide covers various methods to install Stave on your system.

## Prerequisites

- Go 1.21 or later

## Install with Go

The simplest way to install Stave is using `go install`:

```bash
go install github.com/yaklabco/stave@latest
```

This installs the `stave` binary to your `$GOBIN` directory (typically `$GOPATH/bin` or `$HOME/go/bin`).

Verify the installation:

```bash
stave --version
```

## Build from Source

To build Stave from source:

```bash
git clone https://github.com/yaklabco/stave.git
cd stave
go run bootstrap.go
```

The `bootstrap.go` script compiles and installs Stave to your `$GOBIN` directory.

Alternatively, if you already have Stave installed:

```bash
git clone https://github.com/yaklabco/stave.git
cd stave
stave install
```

## Verify Installation

After installation, verify Stave is working:

```bash
# Check version
stave --version

# Show help
stave --help
```

## Cache Directory

Stave caches compiled binaries to speed up subsequent runs. The default cache location is:

| Platform | Location |
|----------|----------|
| Linux    | `~/.stavefile/` |
| macOS    | `~/Library/Caches/stave/` |
| Windows  | `%HOMEDRIVE%%HOMEPATH%\stavefile\` |

You can override this with the `STAVEFILE_CACHE` environment variable:

```bash
export STAVEFILE_CACHE=/path/to/cache
```

Or in a configuration file (see [Configuration](../user-guide/configuration.md)).

## Updating Stave

To update to the latest version:

```bash
go install github.com/yaklabco/stave@latest
```

Or if building from source:

```bash
cd stave
git pull
go run bootstrap.go
```

## Uninstalling

To uninstall Stave:

```bash
# Remove the binary
rm $(which stave)

# Optionally, remove the cache directory
rm -rf ~/.stavefile  # Linux
rm -rf ~/Library/Caches/stave  # macOS
```

## Next Steps

- [Quickstart](quickstart.md) - Create your first stavefile
- [Migration from Mage](migration-from-mage.md) - If you're coming from Mage

