# File Targets

[Home](../index.md) > [User Guide](stavefiles.md) > File Targets

The `pkg/target` package provides utilities for incremental builds by comparing file modification times.

## Concept

Incremental builds skip work when outputs are newer than inputs. The `target` package answers the question: "Is the destination older than any source?"

## target.Path

Compare a destination file against source files:

```go
func Build() error {
    rebuild, err := target.Path("bin/app", "main.go", "go.mod", "go.sum")
    if err != nil {
        return err
    }
    if !rebuild {
        fmt.Println("bin/app is up to date")
        return nil
    }
    return sh.Run("go", "build", "-o", "bin/app", ".")
}
```

`target.Path` returns `true` if:

- The destination does not exist, or
- Any source is newer than the destination

## target.Glob

Use glob patterns for sources:

```go
func Build() error {
    rebuild, err := target.Glob("bin/app", "*.go", "go.*")
    if err != nil {
        return err
    }
    if !rebuild {
        return nil
    }
    return sh.Run("go", "build", "-o", "bin/app", ".")
}
```

Glob patterns follow `filepath.Glob` syntax. An error is returned if a pattern matches no files.

## target.Dir

Recursively compare directories:

```go
func BuildDocs() error {
    rebuild, err := target.Dir("site/", "docs/", "templates/")
    if err != nil {
        return err
    }
    if !rebuild {
        fmt.Println("docs are up to date")
        return nil
    }
    return sh.Run("mkdocs", "build")
}
```

`target.Dir` walks source directories and compares the newest source file against the newest file in the destination directory.

## Time-Based Variants

For comparing against an explicit timestamp:

### target.PathNewer

```go
cutoff := time.Now().Add(-24 * time.Hour)
changed, err := target.PathNewer(cutoff, "file1.go", "file2.go")
```

Returns true if any source is newer than `cutoff`.

### target.GlobNewer

```go
changed, err := target.GlobNewer(cutoff, "*.go")
```

### target.DirNewer

```go
changed, err := target.DirNewer(cutoff, "src/")
```

## Utilities

### target.OldestModTime

Find the oldest modification time in a set of paths:

```go
oldest, err := target.OldestModTime("dir1/", "dir2/")
```

### target.NewestModTime

Find the newest modification time:

```go
newest, err := target.NewestModTime("src/")
```

## Environment Variable Expansion

All path arguments undergo `$VAR` expansion:

```go
rebuild, err := target.Path("$GOBIN/app", "main.go")
```

## Practical Example

A typical incremental build target:

```go
func Build() error {
    sources := []string{"cmd/", "internal/", "go.mod", "go.sum"}
    output := "bin/myapp"

    rebuild, err := target.Dir(output, sources...)
    if err != nil {
        return err
    }

    if !rebuild {
        if st.Verbose() {
            fmt.Println("build: up to date")
        }
        return nil
    }

    return sh.RunV("go", "build", "-o", output, "./cmd/myapp")
}
```

---

## See Also

- [pkg/target API](../api-reference/target.md) - Function reference
- [Shell Commands](shell-commands.md) - Running build commands
- [Home](../index.md)
