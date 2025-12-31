# Migration from `mage`

[Home](../index.md) > Getting Started > Migration from `mage`

Stave is a fork of `mage`. Existing magefiles require minimal changes to work with `stave`.

## Migration Steps

### 1. Rename Files

Rename `magefile.go` to `stavefile.go` (optional but recommended for clarity).

### 2. Update Build Tag

Change the build tag from `mage` to `stave`:

```go
// Before
//go:build mage

// After
//go:build stave
```

### 3. Update Imports

Replace `mage` imports with `stave` equivalents:

| Mage Import                       | Stave Import                           |
| --------------------------------- | -------------------------------------- |
| `github.com/magefile/mage/mg`     | `github.com/yaklabco/stave/pkg/st`     |
| `github.com/magefile/mage/sh`     | `github.com/yaklabco/stave/pkg/sh`     |
| `github.com/magefile/mage/target` | `github.com/yaklabco/stave/pkg/target` |

### 4. Rename Package References

Update function calls:

| Mage                 | Stave                |
| -------------------- | -------------------- |
| `mg.Deps()`          | `st.Deps()`          |
| `mg.CtxDeps()`       | `st.CtxDeps()`       |
| `mg.SerialDeps()`    | `st.SerialDeps()`    |
| `mg.SerialCtxDeps()` | `st.SerialCtxDeps()` |
| `mg.F()`             | `st.F()`             |
| `mg.Namespace`       | `st.Namespace`       |
| `mg.Fatal()`         | `st.Fatal()`         |
| `mg.Fatalf()`        | `st.Fatalf()`        |

## Feature Comparison

Stave maintains compatibility with Mage features and adds:

| Feature                  | Mage | Stave |
|--------------------------| ---- | ----- |
| Target functions         | Yes  | Yes   |
| Dependencies (`Deps`)    | Yes  | Yes   |
| Namespaces               | Yes  | Yes   |
| Target arguments         | Yes  | Yes   |
| Aliases                  | Yes  | Yes   |
| `stave:import`           | Yes  | Yes   |
| Namespace defaults       | No   | Yes   |
| Dry-run mode             | No   | Yes   |
| XDG configuration        | No   | Yes   |
| `stave --config` command | No   | Yes   |
| `STAVE_NUM_PROCESSORS`   | No   | Yes   |

## Stave-Only Features

### Modernized Go Patterns and Compatibility

- Stave targets are just Go functions with the `//go:build stave` build tag.
- The codebase and templates follow modern Go patterns (errors wrapping, contexts, generics where appropriate).
- Minimum supported Go version: 1.24 (see `go.mod`). Using newer Go versions is recommended.

### Dry-Run Mode

Preview commands without execution:

```bash
stave --dryrun build
```

All `sh.Run*` functions print commands prefixed with `DRYRUN:` instead of executing them.

### Configuration Files

Stave supports layered configuration:

1. User config: `~/.config/stave/config.yaml`
2. Project config: `./stave.yaml`
3. Environment variables: `STAVEFILE_*`

See [Configuration](../user-guide/configuration.md) for details.

### Parallelism Control

Set `STAVE_NUM_PROCESSORS` to control parallelism across Stave and downstream tools:

```bash
STAVE_NUM_PROCESSORS=4 stave build
```

---

## See Also

- [Quickstart](quickstart.md) - First stavefile tutorial
- [Configuration](../user-guide/configuration.md) - Configuration system
- [Home](../index.md)
