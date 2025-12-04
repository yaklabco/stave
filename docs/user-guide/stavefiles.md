# Stavefiles

[Home](../index.md) > [User Guide](stavefiles.md) > Stavefiles

A stavefile is a Go source file containing target functions that Stave can execute.

## Build Tag

Stavefiles must include the `stave` build tag:

```go
//go:build stave

package main
```

The build tag prevents normal `go build` from compiling the file while allowing Stave to discover it.

## Package Declaration

The package must be `main`:

```go
package main
```

Stave generates a `main()` function that dispatches to your targets.

## File Naming

Any `.go` file with the `stave` build tag is treated as a stavefile. The conventional name is `stavefile.go`, but any name works:

```
stavefile.go       # conventional
build.go           # also valid if it has the build tag
tasks.go           # also valid if it has the build tag
```

## Multiple Files

Multiple stavefiles in the same directory are compiled together:

```
stavefile.go       # defines Build, Test
deploy.go          # defines Deploy, Rollback (with stave tag)
```

All exported functions from all files become available targets.

## The stavefiles/ Directory

If a `stavefiles/` subdirectory exists, Stave uses it as the source directory. In this case, all `.go` files in `stavefiles/` are included regardless of build tags:

```
project/
├── stavefiles/
│   ├── stavefile.go      # main targets
│   ├── helpers.go        # helper functions (no tag needed)
│   └── deploy.go         # more targets
└── main.go               # your application
```

This convention separates build logic from application code.

## Helper Files

Helper functions that are not targets can be placed in:

1. **Same file**: Unexported functions are not targets.
2. **Separate file with tag**: Include the `stave` tag; unexported functions remain private.
3. **stavefiles/ directory**: All files are compiled together; no tag required.

Example helper:

```go
//go:build stave

package main

// unexported - not a target
func runTests(pkg string) error {
    return sh.Run("go", "test", pkg)
}

// Exported - this IS a target
func Test() error {
    return runTests("./...")
}
```

## Imports

Standard Go imports work as expected:

```go
//go:build stave

package main

import (
    "fmt"
    "os"

    "github.com/yaklabco/stave/pkg/sh"
    "github.com/yaklabco/stave/pkg/st"
)
```

---

## See Also

- [Targets](targets.md) - Defining target functions
- [Quickstart](../getting-started/quickstart.md) - First stavefile tutorial
- [Home](../index.md)

