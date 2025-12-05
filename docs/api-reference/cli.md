# CLI Reference

[Home](../index.md) > [API Reference](cli.md) > CLI

Command-line interface for Stave.

## Synopsis

```bash
stave [flags] [target] [arguments...]
```

## Global Flags

| Flag        | Short | Default         | Description                                  |
| ----------- | ----- | --------------- | -------------------------------------------- |
| `--force`   | `-f`  | `false`         | Force recompilation of stavefile             |
| `--debug`   | `-d`  | `false`         | Print debug messages                         |
| `--verbose` | `-v`  | `false`         | Print verbose output during execution        |
| `--list`    | `-l`  | `false`         | List available targets                       |
| `--info`    | `-i`  | `false`         | Show documentation for a target              |
| `--timeout` | `-t`  | `0`             | Timeout for target execution (e.g., `5m30s`) |
| `--dir`     | `-C`  | `.`             | Directory containing stavefiles              |
| `--workdir` | `-w`  | same as `--dir` | Working directory for target execution       |
| `--gocmd`   |       | `go`            | Go command for compilation                   |
| `--keep`    |       | `false`         | Keep generated mainfile after compilation    |
| `--dryrun`  |       | `false`         | Print commands instead of executing          |
| `--clean`   |       | `false`         | Remove cached compiled binaries              |
| `--init`    |       | `false`         | Create a starter stavefile                   |

## Compilation Flags

Used with `--compile`:

| Flag              | Description                                  |
| ----------------- | -------------------------------------------- |
| `--compile=PATH`  | Compile stavefile to a static binary at PATH |
| `--goos=OS`       | Target OS for cross-compilation              |
| `--goarch=ARCH`   | Target architecture for cross-compilation    |
| `--ldflags=FLAGS` | Linker flags passed to `go build`            |

## Subcommands

### stave config

Manage configuration.

```bash
stave config [subcommand]
```

| Subcommand | Description                                          |
| ---------- | ---------------------------------------------------- |
| (none)     | Show effective configuration                         |
| `init`     | Create default user config file                      |
| `show`     | Show effective configuration (same as no subcommand) |
| `path`     | Show configuration file paths                        |

## Usage Examples

### List Targets

```bash
stave -l
```

### Run a Target

```bash
stave build
```

### Run with Arguments

```bash
stave deploy production true
```

### Show Target Documentation

```bash
stave -i build
```

### Verbose Execution

```bash
stave -v test
```

### Set Timeout

```bash
stave -t 5m build
```

### Dry Run

```bash
stave --dryrun deploy
```

### Force Recompilation

```bash
stave -f build
```

### Cross-Compile Stavefile

```bash
stave --compile=./build/stave-linux --goos=linux --goarch=amd64
```

### Use Different Directory

```bash
stave -C ./build build
```

### Initialize New Project

```bash
stave --init
```

### Clean Cache

```bash
stave --clean
```

## Exit Codes

| Code | Meaning                                         |
| ---- | ----------------------------------------------- |
| 0    | Success                                         |
| 1    | General error (target failed)                   |
| 2    | Usage error (invalid arguments, unknown target) |

Targets can return custom exit codes using `st.Fatal(code, msg)`.

## Environment Variables

Flags can also be set via environment variables:

| Variable               | Equivalent Flag   |
| ---------------------- | ----------------- |
| `STAVEFILE_VERBOSE`    | `--verbose`       |
| `STAVEFILE_DEBUG`      | `--debug`         |
| `STAVEFILE_GOCMD`      | `--gocmd`         |
| `STAVEFILE_CACHE`      | Cache directory   |
| `STAVEFILE_DRYRUN`     | `--dryrun`        |
| `STAVE_NUM_PROCESSORS` | Parallelism limit |

See [Configuration](../user-guide/configuration.md) for the full list.

---

## See Also

- [Configuration](../user-guide/configuration.md) - Config files and environment variables
- [Advanced Topics](../user-guide/advanced.md) - Cross-compilation, dry-run, CI
- [Home](../index.md)
