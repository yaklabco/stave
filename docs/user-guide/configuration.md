# Configuration

[Home](../index.md) > [User Guide](stavefiles.md) > Configuration

Stave supports layered configuration through files and environment variables.

## Precedence

Configuration sources are applied in order (later overrides earlier):

1. Built-in defaults
2. User config file (`~/.config/stave/config.yaml`)
3. Project config file (`./stave.yaml`)
4. Environment variables (`STAVEFILE_*`)

## User Configuration

The user config file location follows XDG conventions:

| Platform | Path                          |
| -------- | ----------------------------- |
| Linux    | `~/.config/stave/config.yaml` |
| macOS    | `~/.config/stave/config.yaml` |
| Windows  | `%APPDATA%\stave\config.yaml` |

Create a default config file:

```bash
stave --config init
```

Example `config.yaml`:

```yaml
go_cmd: go
verbose: false
debug: false
hash_fast: false
ignore_default: false
enable_color: false
target_color: Cyan
```

## Project Configuration

Place `stave.yaml` in your project root to share settings across the team:

```yaml
verbose: true
enable_color: true
target_color: Green
```

Project config overrides user config.

## Configuration Options

| Option           | Type   | Default   | Description                       |
| ---------------- | ------ | --------- | --------------------------------- |
| `cache_dir`      | string | XDG cache | Directory for compiled binaries   |
| `go_cmd`         | string | `go`      | Go command for compilation        |
| `verbose`        | bool   | `false`   | Print verbose output              |
| `debug`          | bool   | `false`   | Print debug messages              |
| `hash_fast`      | bool   | `false`   | Skip GOCACHE, hash files directly |
| `ignore_default` | bool   | `false`   | Ignore default target             |
| `enable_color`   | bool   | `false`   | Enable colored output             |
| `target_color`   | string | `Cyan`    | ANSI color for target names       |

### Boolean values

Boolean options accept a small set of string values. Input is trimmed and matched case-insensitively:

- **True values**: `true`, `yes`, `1`
- **False values**: `false`, `no`, `0`

Empty strings are treated as `false` for configuration values. Any other non-empty value is considered invalid and falls back to the default value for that option.

## Environment Variables

Environment variables override all config files:

| Variable                  | Corresponds To   |
| ------------------------- | ---------------- |
| `STAVEFILE_CACHE`         | `cache_dir`      |
| `STAVEFILE_GOCMD`         | `go_cmd`         |
| `STAVEFILE_VERBOSE`       | `verbose`        |
| `STAVEFILE_DEBUG`         | `debug`          |
| `STAVEFILE_HASHFAST`      | `hash_fast`      |
| `STAVEFILE_IGNOREDEFAULT` | `ignore_default` |
| `STAVEFILE_ENABLE_COLOR`  | `enable_color`   |
| `STAVEFILE_TARGET_COLOR`  | `target_color`   |

Boolean environment variables use the same value semantics as configuration options:

- **True values**: `true`, `yes`, `1`
- **False values**: `false`, `no`, `0`

Unset or empty variables do not override configuration and therefore behave like the existing config value (which is `false` by default, unless otherwise documented).

## Parallelism Control

`STAVE_NUM_PROCESSORS` controls parallelism:

```bash
STAVE_NUM_PROCESSORS=4 stave build
```

This sets `runtime.GOMAXPROCS` and is passed to the compiled stavefile. Use it to limit CPU usage in CI or constrained environments.

## Quiet Mode

Decorative CLI output (hook run messages, test headers, success messages) is automatically suppressed in CI environments. Stave detects CI via:

- `CI`
- `GITHUB_ACTIONS`
- `GITLAB_CI`
- `JENKINS_URL`
- `CIRCLECI`
- `BUILDKITE`

To force quiet mode outside CI:

```bash
STAVE_QUIET=1 stave test
```

## Color Output

Stave automatically detects terminal color support for built-in commands (`stave -l`, `stave --version`). Colors are enabled by default when:

- Standard output is a TTY
- The `TERM` environment variable indicates color support (not `dumb`, `vt100`, `cygwin`, etc.)

### Disabling Color

Use the standard `NO_COLOR` environment variable to disable color output:

```bash
NO_COLOR=1 stave -l
```

See [no-color.org](https://no-color.org/) for the specification. When `NO_COLOR` is set to any value, color output is disabled regardless of terminal capabilities.

### Target Colors

Customize the ANSI color for target names in list output using `STAVEFILE_TARGET_COLOR` or `target_color` in config:

```bash
STAVEFILE_TARGET_COLOR=Green stave -l
```

Supported colors: `Black`, `Red`, `Green`, `Yellow`, `Blue`, `Magenta`, `Cyan`, `White`, and their `Hi` variants (e.g., `HiCyan`).

### Compiled Stavefiles

When running compiled stavefiles (not `stave -l`), color support uses opt-in behavior for backward compatibility. Set `STAVEFILE_ENABLE_COLOR=true` to enable colors in the compiled stavefile's own output.

## Git Hooks Configuration

Configure Git hooks to run Stave targets automatically:

```yaml
hooks:
  pre-commit:
    - target: fmt
    - target: lint
      args: ["--fast"]
  pre-push:
    - target: test
      args: ["./..."]
  commit-msg:
    - target: validate-commit-message
      passStdin: true
```

Each hook entry supports:

| Option      | Type     | Description                                |
| ----------- | -------- | ------------------------------------------ |
| `target`    | string   | Stave target name to run (required)        |
| `args`      | []string | Additional arguments for the target        |
| `workdir`   | string   | Working directory for the target           |
| `passStdin` | bool     | Forward stdin from Git to the target       |

After configuring hooks, install them with:

```bash
stave --hooks install
```

See [Git Hooks](hooks.md) for complete documentation.

## stave --config Subcommands

### stave --config

Display effective configuration:

```bash
stave --config
```

### stave --config init

Create a default user config file:

```bash
stave --config init
```

### stave --config path

Show configuration paths:

```bash
stave --config path
```

Output:

```text
Configuration Paths:
  User config:    /home/user/.config/stave/config.yaml
  Config dir:     /home/user/.config/stave
  Cache dir:      /home/user/.cache/stave
  Data dir:       /home/user/.local/share/stave

No config file currently loaded (using defaults)
```

## Cache Directory

Compiled stavefiles are cached for performance. The cache location:

| Platform | Default Path                 |
| -------- | ---------------------------- |
| Linux    | `~/.cache/stave`             |
| macOS    | `~/Library/Caches/stave`     |
| Windows  | `%LOCALAPPDATA%\cache\stave` |

Override with `cache_dir` in config or `STAVEFILE_CACHE` environment variable.

Clean the cache:

```bash
stave --clean
```

---

## See Also

- [CLI Reference](../api-reference/cli.md) - Command-line flags
- [Git Hooks](hooks.md) - Git hook management
- [Advanced Topics](advanced.md) - CI integration, debugging
- [Home](../index.md)
