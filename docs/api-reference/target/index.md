# pkg/target

The `target` package provides utilities for checking file modification times, enabling incremental builds.

```go
import "github.com/yaklabco/stave/pkg/target"
```

## Overview

The `target` package helps you implement incremental builds by checking whether source files have been modified since a target file was created. This lets you skip expensive operations when nothing has changed.

## Basic Usage

### Path

```go
func Path(dst string, sources ...string) (bool, error)
```

Reports whether any source file is newer than the destination file. Returns `true` if the destination doesn't exist or any source is newer.

```go
func Build() error {
    rebuild, err := target.Path("app", "main.go", "util.go")
    if err != nil {
        return err
    }
    if !rebuild {
        fmt.Println("app is up to date")
        return nil
    }
    return sh.Run("go", "build", "-o", "app", ".")
}
```

### Glob

```go
func Glob(dst string, globs ...string) (bool, error)
```

Like `Path`, but accepts glob patterns for sources.

```go
func Build() error {
    rebuild, err := target.Glob("app", "*.go", "internal/**/*.go")
    if err != nil {
        return err
    }
    if !rebuild {
        return nil
    }
    return sh.Run("go", "build", "-o", "app", "./...")
}
```

Glob syntax follows Go's `filepath.Glob`:
- `*` matches any sequence of non-separator characters
- `**` matches any sequence including separators (in Stave's implementation)
- `?` matches any single non-separator character
- `[abc]` matches any character in the set

### Dir

```go
func Dir(dst string, sources ...string) (bool, error)
```

Like `Path`, but recursively checks directories. For directory sources, compares against the newest file in the source. For directory destinations, compares against the newest file in the destination.

```go
func BuildDocs() error {
    rebuild, err := target.Dir("site/", "docs/", "templates/")
    if err != nil {
        return err
    }
    if !rebuild {
        return nil
    }
    return sh.Run("mkdocs", "build")
}
```

## Time-Based Functions

These functions check against a specific time rather than a destination file.

### PathNewer

```go
func PathNewer(target time.Time, sources ...string) (bool, error)
```

Reports whether any source is newer than the given time.

```go
func CheckSources() error {
    cutoff := time.Now().Add(-24 * time.Hour)
    modified, err := target.PathNewer(cutoff, "src/main.go")
    if err != nil {
        return err
    }
    if modified {
        fmt.Println("Files modified in last 24 hours")
    }
    return nil
}
```

### GlobNewer

```go
func GlobNewer(target time.Time, globs ...string) (bool, error)
```

Like `PathNewer`, but accepts glob patterns.

```go
func CheckGoFiles() error {
    lastBuild := getLastBuildTime()
    modified, err := target.GlobNewer(lastBuild, "**/*.go")
    if err != nil {
        return err
    }
    return nil
}
```

### DirNewer

```go
func DirNewer(target time.Time, sources ...string) (bool, error)
```

Like `PathNewer`, but recursively checks directories.

```go
func CheckSourceDirs() error {
    lastDeploy := getLastDeployTime()
    modified, err := target.DirNewer(lastDeploy, "src/", "lib/")
    if err != nil {
        return err
    }
    if modified {
        fmt.Println("Source directories have changes since last deploy")
    }
    return nil
}
```

## Utility Functions

### OldestModTime

```go
func OldestModTime(targets ...string) (time.Time, error)
```

Returns the oldest modification time among all files in the given paths (recursive for directories).

```go
func GetOldestSource() (time.Time, error) {
    return target.OldestModTime("src/", "lib/")
}
```

### NewestModTime

```go
func NewestModTime(targets ...string) (time.Time, error)
```

Returns the newest modification time among all files in the given paths (recursive for directories).

```go
func GetNewestSource() (time.Time, error) {
    return target.NewestModTime("src/", "lib/")
}
```

## Environment Variable Expansion

All path arguments support environment variable expansion:

```go
func Build() error {
    rebuild, err := target.Path("$OUTPUT/app", "$SRC/*.go")
    if err != nil {
        return err
    }
    // ...
}
```

## Common Patterns

### Skip If Up-to-Date

```go
func Build() error {
    rebuild, err := target.Glob("bin/app", "**/*.go")
    if err != nil {
        return err
    }
    if !rebuild {
        fmt.Println("bin/app is up to date")
        return nil
    }
    return sh.Run("go", "build", "-o", "bin/app", "./cmd/app")
}
```

### Multiple Outputs

```go
func Generate() error {
    // Check if any generated file is older than any source
    sources := []string{"schema.graphql", "templates/*.tmpl"}
    outputs := []string{"generated/types.go", "generated/resolvers.go"}
    
    srcTime, err := target.NewestModTime(sources...)
    if err != nil {
        return err
    }
    
    for _, out := range outputs {
        newer, err := target.PathNewer(srcTime, out)
        if err != nil || newer {
            // Need to regenerate
            return sh.Run("go", "generate", "./...")
        }
    }
    
    fmt.Println("Generated files are up to date")
    return nil
}
```

### Watch-Style Rebuild

```go
func Watch() error {
    var lastCheck time.Time
    for {
        modified, err := target.DirNewer(lastCheck, "src/")
        if err != nil {
            return err
        }
        if modified {
            lastCheck = time.Now()
            if err := Build(); err != nil {
                fmt.Println("Build failed:", err)
            }
        }
        time.Sleep(time.Second)
    }
}
```

### Conditional Docker Build

```go
func BuildDocker() error {
    // Rebuild if Dockerfile or any Go file changed
    rebuild, err := target.Glob(".docker-built", "Dockerfile", "**/*.go", "go.mod", "go.sum")
    if err != nil {
        return err
    }
    if !rebuild {
        fmt.Println("Docker image is up to date")
        return nil
    }
    
    if err := sh.Run("docker", "build", "-t", "myapp", "."); err != nil {
        return err
    }
    
    // Touch a marker file to track build time
    return os.WriteFile(".docker-built", nil, 0644)
}
```

## Error Handling

The functions return errors for:
- Source files that don't exist (not found)
- Permission issues
- Invalid glob patterns (empty results)

The destination not existing is not an error - it returns `true` (needs rebuild).

```go
func Build() error {
    rebuild, err := target.Path("app", "nonexistent.go")
    if err != nil {
        // nonexistent.go doesn't exist
        return fmt.Errorf("source check failed: %w", err)
    }
    // ...
}
```

