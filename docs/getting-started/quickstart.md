# Quickstart

[Home](../index.md) > Getting Started > Quickstart

This guide walks through creating a stavefile and running targets.

## Create a Stavefile

Create `stavefile.go` in your project root:

```go
//go:build stave

package main

import (
    "fmt"

    "github.com/yaklabco/stave/pkg/sh"
    "github.com/yaklabco/stave/pkg/st"
)

// Build compiles the project.
func Build() error {
    st.Deps(Generate)
    fmt.Println("Building...")
    return sh.Run("go", "build", "-o", "myapp", ".")
}

// Test runs the test suite.
func Test() error {
    return sh.RunV("go", "test", "./...")
}

// Generate runs code generators.
func Generate() error {
    return sh.Run("go", "generate", "./...")
}

// Clean removes build artifacts.
func Clean() error {
    return sh.Rm("myapp")
}
```

The `//go:build stave` directive marks this file for Stave. The package must be `main`.

## List Targets

```bash
stave -l
```

Output:

```text
Targets:
  build       compiles the project.
  clean       removes build artifacts.
  generate    runs code generators.
  test        runs the test suite.
```

## Run a Target

```bash
stave build
```

Stave compiles the stavefile into a binary (cached for subsequent runs), then executes the `Build` function. Because `Build` calls `st.Deps(Generate)`, the `Generate` target runs first.

## Set a Default Target

Add a `Default` variable to run a target when no arguments are provided:

```go
var Default = Build
```

Now `stave` with no arguments runs `Build`.

## Add Parallel Dependencies

```go
func All() {
    st.Deps(Build, Test)  // Build and Test run in parallel
}
```

Each dependency runs exactly once, even if referenced multiple times in the dependency graph.

---

## See Also

- [Installation](installation.md) - Installing Stave
- [Stavefiles](../user-guide/stavefiles.md) - File conventions
- [Dependencies](../user-guide/dependencies.md) - Dependency execution
- [Home](../index.md)
