# Migration from Mage

This guide helps Mage users migrate to Stave. The transition is straightforward since Stave maintains compatibility with Mage's core concepts.

## Overview

Stave is a fork of Mage with additional features. Most Mage concepts translate directly:

| Mage | Stave |
|------|-------|
| `magefile.go` | `stavefile.go` |
| `//go:build mage` | `//go:build stave` |
| `github.com/magefile/mage/mg` | `github.com/yaklabco/stave/pkg/st` |
| `github.com/magefile/mage/sh` | `github.com/yaklabco/stave/pkg/sh` |
| `github.com/magefile/mage/target` | `github.com/yaklabco/stave/pkg/target` |
| `mg.Deps()` | `st.Deps()` |
| `mg.SerialDeps()` | `st.SerialDeps()` |
| `mg.CtxDeps()` | `st.CtxDeps()` |
| `mg.F()` | `st.F()` |
| `mg.Namespace` | `st.Namespace` |
| `mg.Verbose()` | `st.Verbose()` |
| `MAGEFILE_*` | `STAVEFILE_*` |

## Step-by-Step Migration

### 1. Rename Your Magefile

```bash
mv magefile.go stavefile.go
```

If you have a `magefiles/` directory:

```bash
mv magefiles/ stavefiles/
```

### 2. Update Build Tags

Change the build tag from `mage` to `stave`:

```go
// Before
//go:build mage

// After
//go:build stave
```

### 3. Update Imports

Replace Mage imports with Stave equivalents:

```go
// Before
import (
    "github.com/magefile/mage/mg"
    "github.com/magefile/mage/sh"
    "github.com/magefile/mage/target"
)

// After
import (
    "github.com/yaklabco/stave/pkg/st"
    "github.com/yaklabco/stave/pkg/sh"
    "github.com/yaklabco/stave/pkg/target"
)
```

### 4. Replace Package References

Update function calls from `mg` to `st`:

```go
// Before
mg.Deps(Build, Test)
mg.SerialDeps(Clean, Build)
mg.CtxDeps(ctx, Build)
mg.F(Deploy, "prod")
if mg.Verbose() { ... }

// After
st.Deps(Build, Test)
st.SerialDeps(Clean, Build)
st.CtxDeps(ctx, Build)
st.F(Deploy, "prod")
if st.Verbose() { ... }
```

### 5. Update Namespace Declarations

```go
// Before
type Build mg.Namespace

// After
type Build st.Namespace
```

### 6. Update Environment Variables

If you use Mage environment variables, update them:

| Mage | Stave |
|------|-------|
| `MAGEFILE_CACHE` | `STAVEFILE_CACHE` |
| `MAGEFILE_VERBOSE` | `STAVEFILE_VERBOSE` |
| `MAGEFILE_DEBUG` | `STAVEFILE_DEBUG` |
| `MAGEFILE_GOCMD` | `STAVEFILE_GOCMD` |
| `MAGEFILE_HASHFAST` | `STAVEFILE_HASHFAST` |
| `MAGEFILE_IGNOREDEFAULT` | `STAVEFILE_IGNOREDEFAULT` |
| `MAGEFILE_ENABLE_COLOR` | `STAVEFILE_ENABLE_COLOR` |
| `MAGEFILE_TARGET_COLOR` | `STAVEFILE_TARGET_COLOR` |

### 7. Update CI/CD Scripts

Replace `mage` commands with `stave`:

```bash
# Before
mage build
mage -v test

# After
stave build
stave -v test
```

### 8. Install Stave

```bash
go install github.com/yaklabco/stave@latest
```

## Complete Example

### Before (Mage)

```go
//go:build mage

package main

import (
    "fmt"

    "github.com/magefile/mage/mg"
    "github.com/magefile/mage/sh"
)

var Default = Build

type Build mg.Namespace

func (Build) Binary() error {
    mg.Deps(Lint, Test)
    return sh.Run("go", "build", "-o", "app", ".")
}

func (Build) Docker() error {
    mg.SerialDeps(Build{}.Binary)
    return sh.Run("docker", "build", "-t", "myapp", ".")
}

func Lint() error {
    if mg.Verbose() {
        fmt.Println("Running linter...")
    }
    return sh.Run("golangci-lint", "run")
}

func Test() error {
    return sh.Run("go", "test", "./...")
}
```

### After (Stave)

```go
//go:build stave

package main

import (
    "fmt"

    "github.com/yaklabco/stave/pkg/sh"
    "github.com/yaklabco/stave/pkg/st"
)

var Default = Build

type Build st.Namespace

func (Build) Binary() error {
    st.Deps(Lint, Test)
    return sh.Run("go", "build", "-o", "app", ".")
}

func (Build) Docker() error {
    st.SerialDeps(Build{}.Binary)
    return sh.Run("docker", "build", "-t", "myapp", ".")
}

func Lint() error {
    if st.Verbose() {
        fmt.Println("Running linter...")
    }
    return sh.Run("golangci-lint", "run")
}

func Test() error {
    return sh.Run("go", "test", "./...")
}
```

## Automated Migration Script

For larger codebases, you can use sed to automate some changes:

```bash
#!/bin/bash

# Rename files
find . -name "magefile.go" -exec bash -c 'mv "$0" "${0/magefile/stavefile}"' {} \;
[ -d "magefiles" ] && mv magefiles stavefiles

# Update build tags
find . -name "*.go" -exec sed -i '' 's/\/\/go:build mage/\/\/go:build stave/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\/\/ +build mage/\/\/ +build stave/g' {} \;

# Update imports
find . -name "*.go" -exec sed -i '' 's|github.com/magefile/mage/mg|github.com/yaklabco/stave/pkg/st|g' {} \;
find . -name "*.go" -exec sed -i '' 's|github.com/magefile/mage/sh|github.com/yaklabco/stave/pkg/sh|g' {} \;
find . -name "*.go" -exec sed -i '' 's|github.com/magefile/mage/target|github.com/yaklabco/stave/pkg/target|g' {} \;

# Update package references
find . -name "*.go" -exec sed -i '' 's/mg\./st./g' {} \;

# Update env vars in scripts
find . -name "*.sh" -exec sed -i '' 's/MAGEFILE_/STAVEFILE_/g' {} \;
find . -name "*.yml" -exec sed -i '' 's/MAGEFILE_/STAVEFILE_/g' {} \;
find . -name "*.yaml" -exec sed -i '' 's/MAGEFILE_/STAVEFILE_/g' {} \;

# Update commands in scripts
find . -name "*.sh" -exec sed -i '' 's/mage /stave /g' {} \;
find . -name "*.yml" -exec sed -i '' 's/mage /stave /g' {} \;
find . -name "*.yaml" -exec sed -i '' 's/mage /stave /g' {} \;

echo "Migration complete. Please review changes and run: go mod tidy"
```

## Feature Parity

Stave maintains feature parity with Mage:

| Feature | Mage | Stave |
|---------|------|-------|
| Parallel dependencies | Yes | Yes |
| Serial dependencies | Yes | Yes |
| Context-aware deps | Yes | Yes |
| Namespaces | Yes | Yes |
| Aliases | Yes | Yes |
| Default target | Yes | Yes |
| Verbose mode | Yes | Yes |
| Target arguments | Yes | Yes |
| Cross-compilation | Yes | Yes |
| Dry-run mode | Partial | Yes |
| Configuration files | No | Yes |

## New Features in Stave

Stave adds several features not in Mage:

1. **Configuration Files** - YAML config at user and project level
2. **Enhanced Dry-Run** - Complete dry-run support for all shell commands
3. **XDG Compliance** - Proper cache and config directory locations
4. **Colorized Output** - Configurable terminal colors
5. **Modern CLI** - Built with Cobra/Fang for better UX

## Troubleshooting

### "package mg is not in GOROOT"

You still have Mage imports. Update them to Stave equivalents.

### "undefined: mg.Deps"

Replace `mg.` with `st.` throughout your code.

### Tests fail after migration

1. Clear the binary cache: `stave --clean`
2. Force rebuild: `stave -f test`
3. Update `go.mod`: `go mod tidy`

### Environment variables not working

Mage's `MAGEFILE_*` variables don't work with Stave. Update to `STAVEFILE_*`.

## Getting Help

If you encounter issues migrating:

1. Check the [Stave documentation](../index.md)
2. [Open an issue](https://github.com/yaklabco/stave/issues)
3. Compare behavior with the original Mage to identify differences

