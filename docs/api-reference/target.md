# pkg/target

[Home](../index.md) > [API Reference](cli.md) > pkg/target

Package `target` provides utilities for incremental builds by comparing file modification times.

```go
import "github.com/yaklabco/stave/pkg/target"
```

## Destination vs Sources

### Path

```go
func Path(dst string, sources ...string) (bool, error)
```

Returns true if any source is newer than the destination, or if the destination does not exist.

```go
rebuild, err := target.Path("bin/app", "main.go", "go.mod")
if err != nil {
    return err
}
if rebuild {
    return sh.Run("go", "build", "-o", "bin/app", ".")
}
return nil
```

### Glob

```go
func Glob(dst string, globs ...string) (bool, error)
```

Expand glob patterns and compare against the destination.

```go
rebuild, err := target.Glob("bin/app", "*.go", "internal/**/*.go")
```

Returns an error if a glob pattern matches no files.

### Dir

```go
func Dir(dst string, sources ...string) (bool, error)
```

Recursively compare directories. Returns true if any file in any source directory is newer than the newest file in the destination directory.

```go
rebuild, err := target.Dir("site/", "docs/", "templates/")
```

## Time-Based Comparisons

### PathNewer

```go
func PathNewer(target time.Time, sources ...string) (bool, error)
```

Returns true if any source file is newer than the given time.

```go
cutoff := time.Now().Add(-1 * time.Hour)
changed, err := target.PathNewer(cutoff, "config.yaml", "data.json")
```

### GlobNewer

```go
func GlobNewer(target time.Time, sources ...string) (bool, error)
```

Expand globs and compare against the given time.

```go
changed, err := target.GlobNewer(cutoff, "*.go")
```

### DirNewer

```go
func DirNewer(target time.Time, sources ...string) (bool, error)
```

Recursively check if any file in source directories is newer than the given time.

```go
changed, err := target.DirNewer(cutoff, "src/", "assets/")
```

## Utilities

### OldestModTime

```go
func OldestModTime(targets ...string) (time.Time, error)
```

Find the oldest modification time among all files in the given paths (recursive).

```go
oldest, err := target.OldestModTime("cache/")
```

### NewestModTime

```go
func NewestModTime(targets ...string) (time.Time, error)
```

Find the newest modification time among all files in the given paths (recursive).

```go
newest, err := target.NewestModTime("src/")
```

## Environment Variable Expansion

All path arguments undergo `os.ExpandEnv`:

```go
rebuild, err := target.Path("$GOBIN/app", "main.go")
```

## Example: Incremental Build

```go
func Build() error {
    rebuild, err := target.Dir("bin/myapp", "cmd/", "internal/", "go.mod", "go.sum")
    if err != nil {
        return err
    }

    if !rebuild {
        if st.Verbose() {
            fmt.Println("bin/myapp is up to date")
        }
        return nil
    }

    return sh.RunV("go", "build", "-o", "bin/myapp", "./cmd/myapp")
}
```

## Example: Time-Based Cache Invalidation

```go
func FetchData() error {
    cacheFile := "data/cache.json"
    maxAge := 24 * time.Hour

    info, err := os.Stat(cacheFile)
    if err == nil && time.Since(info.ModTime()) < maxAge {
        return nil  // cache is fresh
    }

    // fetch and write new data
    return fetchAndSave(cacheFile)
}
```

---

## See Also

- [File Targets](../user-guide/file-targets.md) - Usage guide
- [Shell Commands](../user-guide/shell-commands.md) - Running build commands
- [Home](../index.md)

