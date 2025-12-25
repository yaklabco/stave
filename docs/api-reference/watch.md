# pkg/watch

[Home](../index.md) > [API Reference](cli.md) > pkg/watch

Package `watch` provides functions for monitoring file changes and automatically re-running targets.

```go
import "github.com/yaklabco/stave/pkg/watch"
```

## Core Functions

### Watch

```go
func Watch(patterns ...string)
```

Registers glob patterns to watch for the current target. When any file matching these patterns changes, the current target's context is cancelled, and the target is re-run.

In order for a target to actually be watched, it must be explicitly requested on the command line (or be a dependency of a requested target that activates watch mode).

Patterns can include wildcards:

- `*` matches any character except path separators
- `**` matches any character including path separators
- `?` matches any single character except path separators
- `[seq]` matches any character in `seq`
- `[!seq]` matches any character not in `seq`

```go
func WatchDir(dir string) error {
    watch.Watch(fmt.Sprintf("%s/**", dir))
    // ...
}
```

### Deps

```go
func Deps(fns ...any)
```

Registers watch-specific dependencies for the current target. These dependencies are tracked and run in a way that respects the watch mode's cancellable context.

```go
func Build() {
    watch.Deps(Init, Generate)
    // ...
}
```

## Utility Functions

### IsOverallWatchMode

```go
func IsOverallWatchMode() bool
```

Returns whether the current execution is in watch mode.

---

## See Also

- [Watch Mode](../user-guide/watch.md) - Usage guide
- [pkg/sh](sh.md) - Standard shell helpers
- [Home](../index.md)
