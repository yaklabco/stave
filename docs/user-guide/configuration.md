# Configuration

Stave can be configured through configuration files and environment variables.

## Configuration Sources

Configuration is loaded in the following order (later sources override earlier):

1. **Defaults** - Built-in default values
2. **User config** - `~/.config/stave/config.yaml`
3. **Project config** - `./stave.yaml` in the current directory
4. **Environment variables** - `STAVEFILE_*` variables

## Configuration File Locations

### User Configuration

The user configuration file provides personal defaults across all projects.

| Platform | Location |
|----------|----------|
| Linux    | `~/.config/stave/config.yaml` |
| macOS    | `~/.config/stave/config.yaml` |
| Windows  | `%APPDATA%\stave\config.yaml` |

The XDG_CONFIG_HOME environment variable overrides the default location on all platforms.

### Project Configuration

Place a `stave.yaml` file in your project root to configure Stave for that project:

```yaml
# stave.yaml
verbose: true
go_cmd: go1.21
target_color: Green
```

## Configuration Options

### cache_dir

Directory where Stave caches compiled binaries.

```yaml
cache_dir: ~/.cache/stave
```

Environment variable: `STAVEFILE_CACHE`

Default:
- Linux: `~/.cache/stave`
- macOS: `~/Library/Caches/stave`
- Windows: `%LOCALAPPDATA%\cache\stave`

### go_cmd

The Go command to use for compilation.

```yaml
go_cmd: go
```

Environment variable: `STAVEFILE_GOCMD`

Default: `go`

Use this to specify a different Go version:

```yaml
go_cmd: go1.22
```

### verbose

Enable verbose output when running targets.

```yaml
verbose: false
```

Environment variable: `STAVEFILE_VERBOSE`

Default: `false`

### debug

Enable debug messages for troubleshooting.

```yaml
debug: false
```

Environment variable: `STAVEFILE_DEBUG`

Default: `false`

### hash_fast

Use quick hashing instead of relying on GOCACHE.

```yaml
hash_fast: false
```

Environment variable: `STAVEFILE_HASHFAST`

Default: `false`

When enabled:
- Faster startup (skips GOCACHE check)
- May miss transitive dependency changes
- Use `-f` flag to force rebuild when needed

### ignore_default

Ignore the default target in stavefiles.

```yaml
ignore_default: false
```

Environment variable: `STAVEFILE_IGNOREDEFAULT`

Default: `false`

When enabled, running `stave` without arguments lists targets instead of running the default.

### enable_color

Enable colored output in terminal.

```yaml
enable_color: false
```

Environment variable: `STAVEFILE_ENABLE_COLOR`

Default: `false`

### target_color

ANSI color name for colorizing target names in output.

```yaml
target_color: Cyan
```

Environment variable: `STAVEFILE_TARGET_COLOR`

Default: `Cyan`

Valid colors:
- `Black`, `Red`, `Green`, `Yellow`, `Blue`, `Magenta`, `Cyan`, `White`
- `BrightBlack`, `BrightRed`, `BrightGreen`, `BrightYellow`
- `BrightBlue`, `BrightMagenta`, `BrightCyan`, `BrightWhite`

## Environment Variables

All configuration options can be set via environment variables with the `STAVEFILE_` prefix:

| Config Option | Environment Variable |
|---------------|---------------------|
| cache_dir | `STAVEFILE_CACHE` |
| go_cmd | `STAVEFILE_GOCMD` |
| verbose | `STAVEFILE_VERBOSE` |
| debug | `STAVEFILE_DEBUG` |
| hash_fast | `STAVEFILE_HASHFAST` |
| ignore_default | `STAVEFILE_IGNOREDEFAULT` |
| enable_color | `STAVEFILE_ENABLE_COLOR` |
| target_color | `STAVEFILE_TARGET_COLOR` |

Boolean values accept: `1`, `true`, `TRUE`, `True` for true; anything else is false.

## The stave config Command

Stave includes a `config` subcommand for managing configuration:

### Show Effective Configuration

```bash
stave config
# or
stave config show
```

Output:

```
# Effective Stave Configuration
# Loaded from: /home/user/.config/stave/config.yaml

cache_dir: /home/user/.cache/stave
go_cmd: go
verbose: false
debug: false
hash_fast: false
ignore_default: false
enable_color: false
target_color: Cyan
```

### Create Default Configuration

```bash
stave config init
```

Creates `~/.config/stave/config.yaml` with default values.

### Show Configuration Paths

```bash
stave config path
```

Output:

```
Configuration Paths:
  User config:    /home/user/.config/stave/config.yaml
  Config dir:     /home/user/.config/stave
  Cache dir:      /home/user/.cache/stave
  Data dir:       /home/user/.local/share/stave

Active config file: /home/user/.config/stave/config.yaml
```

## Example Configurations

### Development Setup

```yaml
# ~/.config/stave/config.yaml
verbose: true
debug: false
enable_color: true
target_color: Green
```

### CI/CD Pipeline

```bash
# In CI environment
export STAVEFILE_VERBOSE=1
export STAVEFILE_HASHFAST=1
export STAVEFILE_CACHE=/tmp/stave-cache
```

### Project-Specific

```yaml
# ./stave.yaml
go_cmd: go1.22
verbose: false
```

## Accessing Configuration in Stavefiles

Use the `st` package to access runtime configuration:

```go
//go:build stave

package main

import (
    "fmt"
    "github.com/yaklabco/stave/pkg/st"
)

func Info() {
    fmt.Printf("Verbose: %v\n", st.Verbose())
    fmt.Printf("Debug: %v\n", st.Debug())
    fmt.Printf("GoCmd: %s\n", st.GoCmd())
    fmt.Printf("CacheDir: %s\n", st.CacheDir())
}
```

Available functions:
- `st.Verbose()` - Returns true if verbose mode is enabled
- `st.Debug()` - Returns true if debug mode is enabled
- `st.GoCmd()` - Returns the configured Go command
- `st.CacheDir()` - Returns the cache directory path
- `st.HashFast()` - Returns true if fast hashing is enabled
- `st.IgnoreDefault()` - Returns true if default target is ignored
- `st.EnableColor()` - Returns true if color output is enabled
- `st.TargetColor()` - Returns the ANSI escape code for target color

