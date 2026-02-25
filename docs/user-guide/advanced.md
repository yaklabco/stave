# Advanced Topics

[Home](../index.md) > [User Guide](stavefiles.md) > Advanced Topics

This page covers cross-compilation, dry-run mode, CI integration, and debugging.

## Cross-Compilation

### Compile to Static Binary

Produce a standalone binary instead of running targets:

```bash
stave --compile=./build/stave-linux --goos=linux --goarch=amd64
```

The compiled binary can run on machines without Go installed.

### Flags

| Flag              | Description                                     |
| ----------------- | ----------------------------------------------- |
| `--compile=PATH`  | Output path for compiled binary                 |
| `--goos=OS`       | Target operating system                         |
| `--goarch=ARCH`   | Target architecture                             |
| `--ldflags=FLAGS` | Linker flags (e.g., `-s -w` for smaller binary) |

### Example

Build for multiple platforms:

```go
func Release() error {
    platforms := []struct{ os, arch string }{
        {"linux", "amd64"},
        {"darwin", "amd64"},
        {"darwin", "arm64"},
        {"windows", "amd64"},
    }

    for _, p := range platforms {
        output := fmt.Sprintf("dist/app-%s-%s", p.os, p.arch)
        if p.os == "windows" {
            output += ".exe"
        }
        err := sh.RunWith(
            map[string]string{"GOOS": p.os, "GOARCH": p.arch, "CGO_ENABLED": "0"},
            "go", "build", "-o", output, ".",
        )
        if err != nil {
            return err
        }
    }
    return nil
}
```

## Dry-Run Mode

Preview commands without executing them:

```bash
stave --dryrun build
```

### Behavior

- All `sh.Run*` functions print `DRYRUN: command args...` instead of executing
- `sh.Rm` and `sh.Copy` also print instead of acting
- The stavefile itself still runs (only shell commands are skipped)

## direnv Integration

Delegate environment variable management to [direnv](https://direnv.net/) directly from Stave:

```bash
stave --direnv [direnv-subcommand] [args...]
```

### Examples

#### Load environment from `.envrc`

If you are using `direnv` to manage your environment variables, you can run `direnv` commands through Stave:

```bash
stave --direnv allow
stave --direnv reload
```

This is particularly useful in CI environments where you might want to leverage `direnv`'s environment-loading capabilities without having to install the `direnv` binary separately, as Stave includes `direnv` as a library.

### Behavior

- Stave delegates the execution to the embedded `direnv` library.
- Environment variables set via `direnv` are made available to the subsequent execution within that process (though typically you use `--direnv` as its own command).
- The `--direnv` flag acts as a "pseudo-flag" that triggers a dedicated execution mode, similar to `--hooks` or `--config`.

## CI Integration

### GitHub Actions

```yaml
name: Build
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Install Stave
        run: go install github.com/yaklabco/stave@latest

      - name: Build
        run: stave build

      - name: Test
        run: stave test
```

### Caching Compiled Stavefiles

Cache the Stave binary cache to speed up CI:

```yaml
- uses: actions/cache@v4
  with:
    path: ~/.cache/stave
    key: stave-${{ runner.os }}-${{ hashFiles('stavefile.go') }}
```

### Parallelism Control

Limit parallelism in resource-constrained environments:

```yaml
- name: Build
  env:
    STAVE_NUM_PROCESSORS: 2
  run: stave build
```

### Disabling Hooks in CI

Git hooks are typically not needed in CI (tests run explicitly). Disable them:

```yaml
- name: Build
  env:
    STAVE_HOOKS: "0"
  run: |
    stave build
    stave test
```

### GitLab CI

```yaml
build:
  image: golang:1.21
  script:
    - go install github.com/yaklabco/stave@latest
    - stave build
    - stave test
  cache:
    paths:
      - ~/.cache/stave
```

## Debugging

### Verbose Mode

Print target execution and command details:

```bash
stave -v build
```

Or set the environment variable:

```bash
STAVEFILE_VERBOSE=true stave build
```

### Debug Mode

Print internal Stave operations (parsing, compilation, caching):

```bash
stave -d build
```

Or:

```bash
STAVEFILE_DEBUG=true stave build
```

### Keep Generated Files

Retain the generated mainfile for inspection:

```bash
stave --keep build
```

The generated file is `stave_output_file_<hash>_<pid>.go` in the stavefile directory.

### Force Recompilation

Bypass the cache and recompile:

```bash
stave -f build
```

### Common Issues

This section will be expanded as issues are reported.

## Git Hooks

### Hook Debugging

Run hooks manually to debug issues:

```bash
stave --hooks run pre-commit
```

Enable debug output in hook scripts:

```bash
STAVE_HOOKS=debug git commit -m "test"
```

### Skipping Hooks Temporarily

Set `STAVE_HOOKS=0` to disable all hooks:

```bash
STAVE_HOOKS=0 git commit -m "WIP"
```

This is preferable to `git commit --no-verify` as it still allows the hook script to run (and log that hooks are disabled).

### User Init Script

If hooks fail in GUI clients (SourceTree, VS Code, etc.) due to missing `stave` on PATH, create a user init script:

```bash
mkdir -p ~/.config/stave/hooks
cat > ~/.config/stave/hooks/init.sh << 'EOF'
export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"
EOF
```

This script is sourced by all hook scripts before running Stave.

---

## See Also

- [Configuration](configuration.md) - Config files and environment variables
- [Git Hooks](hooks.md) - Complete hooks documentation
- [CLI Reference](../api-reference/cli.md) - All command-line flags
- [Shell Commands](shell-commands.md) - Command execution details
- [Home](../index.md)
