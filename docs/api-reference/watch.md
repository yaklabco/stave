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
func Deps(fns ...interface{})
```

Registers watch-specific dependencies for the current target. These dependencies are tracked and run in a way that respects the watch mode's cancellable context.

```go
func Build() {
    watch.Deps(Init, Generate)
    // ...
}
```

## Watch-Aware Shell Helpers

These functions are counterparts to the `sh` package functions but are aware of the cancellable context used in watch mode. When a re-run is triggered, any ongoing command started with these helpers will be terminated.

### Run

```go
func Run(cmd string, args ...string) error
```

### RunV

```go
func RunV(cmd string, args ...string) error
```

### RunWith

```go
func RunWith(env map[string]string, cmd string, args ...string) error
```

### RunWithV

```go
func RunWithV(env map[string]string, cmd string, args ...string) error
```

### Output

```go
func Output(cmd string, args ...string) (string, error)
```

### OutputWith

```go
func OutputWith(env map[string]string, cmd string, args ...string) (string, error)
```

### Piper

```go
func Piper(stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) error
```

### PiperWith

```go
func PiperWith(env map[string]string, stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) error
```

### Exec

```go
func Exec(env map[string]string, stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) (bool, error)
```

### RunCmd

```go
func RunCmd(cmd string, args ...string) func(args ...string) error
```

### OutCmd

```go
func OutCmd(cmd string, args ...string) func(args ...string) (string, error)
```

### Rm

```go
func Rm(path string) error
```

### Copy

```go
func Copy(dst, src string) error
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
