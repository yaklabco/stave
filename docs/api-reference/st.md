# pkg/st

[Home](../index.md) > [API Reference](cli.md) > pkg/st

Package `st` provides dependency management, error handling, and runtime utilities for stavefiles.

```go
import "github.com/yaklabco/stave/pkg/st"
```

## Dependency Functions

### Deps

```go
func Deps(fns ...interface{})
```

Run dependencies in parallel. Each dependency runs exactly once per Stave invocation.

```go
func Build() {
    st.Deps(Generate, Compile)
}
```

### CtxDeps

```go
func CtxDeps(ctx context.Context, fns ...interface{})
```

Run dependencies in parallel with a context.

```go
func Build(ctx context.Context) {
    st.CtxDeps(ctx, Generate, Compile)
}
```

### SerialDeps

```go
func SerialDeps(fns ...interface{})
```

Run dependencies sequentially.

```go
func Deploy() {
    st.SerialDeps(Build, Test, Push)
}
```

### SerialCtxDeps

```go
func SerialCtxDeps(ctx context.Context, fns ...interface{})
```

Run dependencies sequentially with a context.

## Function Wrapper

### F

```go
func F(target interface{}, args ...interface{}) Fn
```

Wrap a target function with arguments for use in `Deps`.

```go
func Deploy(env string) error { /* ... */ }

func DeployAll() {
    st.Deps(
        st.F(Deploy, "staging"),
        st.F(Deploy, "production"),
    )
}
```

### Fn Interface

```go
type Fn interface {
    Name() string
    ID() string
    Run(ctx context.Context) error
}
```

Interface implemented by `F()` results. Used for dependency identity and execution.

## Error Functions

### Fatal

```go
func Fatal(code int, args ...interface{}) error
```

Return an error that causes Stave to exit with the given code.

```go
if missing {
    return st.Fatal(2, "required file not found")
}
```

### Fatalf

```go
func Fatalf(code int, format string, args ...interface{}) error
```

Formatted version of `Fatal`.

```go
return st.Fatalf(1, "build failed: %v", err)
```

### ExitStatus

```go
func ExitStatus(err error) int
```

Extract exit code from an error. Returns 0 for nil, 1 for errors without a code.

```go
code := st.ExitStatus(err)
```

## Runtime Query Functions

### Verbose

```go
func Verbose() bool
```

Returns true if `-v` is set or `STAVEFILE_VERBOSE` is a true value. Accepted true values are `true`, `yes`, and `1` (case-insensitive, leading/trailing whitespace ignored).

```go
if st.Verbose() {
    fmt.Println("Building with verbose output...")
}
```

### Debug

```go
func Debug() bool
```

Returns true if `-d` is set or `STAVEFILE_DEBUG` is a true value (`true`, `yes`, or `1`, case-insensitive).

### Info

```go
func Info() bool
```

Returns true if `-i` is set or `STAVEFILE_INFO` is a true value (`true`, `yes`, or `1`, case-insensitive).

### GoCmd

```go
func GoCmd() string
```

Returns the Go command to use. Default is `"go"`, overridden by `STAVEFILE_GOCMD`.

### CacheDir

```go
func CacheDir() string
```

Returns the cache directory for compiled binaries.

### HashFast

```go
func HashFast() bool
```

Returns true if `STAVEFILE_HASHFAST` is a true value (`true`, `yes`, or `1`, case-insensitive).

### IgnoreDefault

```go
func IgnoreDefault() bool
```

Returns true if `STAVEFILE_IGNOREDEFAULT` is a true value (`true`, `yes`, or `1`, case-insensitive).

### IsOverallWatchMode

```go
func IsOverallWatchMode() bool
```

Returns whether the current execution is in watch mode.

### SetOverallWatchMode

```go
func SetOverallWatchMode(b bool)
```

Sets whether we are in overall watch mode.

### SetOutermostTarget

```go
func SetOutermostTarget(name string)
```

Sets the name of the outermost target being run.

### GetOutermostTarget

```go
func GetOutermostTarget() string
```

Returns the name of the outermost target.

### ColorEnabled

```go
func ColorEnabled() bool
```

Returns true if color output should be enabled using auto-detection. This is used by Stave's built-in commands (`stave -l`, `stave --version`). Auto-detection respects:

- `NO_COLOR` environment variable (disables color when set to any value)
- `TERM` environment variable (disables color for terminals like `dumb`, `vt100`, `cygwin`)

### TerminalSupportsColor

```go
func TerminalSupportsColor(term string) bool
```

Returns true if the given TERM value is not in the known-no-color blacklist. Returns false for terminals like `dumb`, `vt100`, `cygwin`, `xterm-mono`. An empty string returns true (letting Lipgloss handle further TTY detection).

### NoColorTERMs

```go
func NoColorTERMs() []string
```

Returns a sorted list of TERM values that do not support color output. Used internally and by the generated mainfile template.

### TargetColor

```go
func TargetColor() string
```

Returns the ANSI color code for target names. Respects `STAVEFILE_TARGET_COLOR` environment variable if set.

### TargetStyle

```go
func TargetStyle() lipgloss.Style
```

Returns a Lipgloss style configured with the user's target color. This is the preferred way to style target names when using Charmbracelet/Lipgloss, as it respects `STAVEFILE_TARGET_COLOR` and integrates cleanly with other Lipgloss styles.

## Types

### Namespace

```go
type Namespace struct{}
```

Marker type for creating namespaced targets.

```go
type Build st.Namespace

func (Build) Docker() error { /* ... */ }
```

---

## See Also

- [Dependencies](../user-guide/dependencies.md) - Using Deps
- [Namespaces](../user-guide/namespaces.md) - Using Namespace
- [Targets](../user-guide/targets.md) - Target definitions
- [Home](../index.md)
