# pkg/st

The `st` package provides the core API for writing stavefiles.

```go
import "github.com/yaklabco/stave/pkg/st"
```

## Overview

The `st` package provides:
- Dependency management (`Deps`, `SerialDeps`, `CtxDeps`)
- Function wrapping for deps with arguments (`F`)
- Error handling utilities (`Fatal`, `Fatalf`, `ExitStatus`)
- Runtime information (`Verbose`, `Debug`, `GoCmd`, etc.)
- Namespace support for organizing targets

## Dependency Functions

### Deps

```go
func Deps(fns ...interface{})
```

Runs the given functions as dependencies in parallel. Each dependency runs exactly once per Stave invocation, regardless of how many times it's requested.

```go
func Build() error {
    st.Deps(Lint, Test)  // Run Lint and Test in parallel
    return sh.Run("go", "build", "./...")
}
```

Valid function signatures:
- `func()`
- `func() error`
- `func(context.Context)`
- `func(context.Context) error`

### SerialDeps

```go
func SerialDeps(fns ...interface{})
```

Like `Deps`, but runs each dependency sequentially instead of in parallel.

```go
func Deploy() error {
    st.SerialDeps(Build, Test, Push)  // Run in order
    return nil
}
```

### CtxDeps

```go
func CtxDeps(ctx context.Context, fns ...interface{})
```

Like `Deps`, but passes a context to dependencies that accept one.

```go
func Build(ctx context.Context) error {
    st.CtxDeps(ctx, Lint, Test)
    return sh.Run("go", "build", "./...")
}
```

### SerialCtxDeps

```go
func SerialCtxDeps(ctx context.Context, fns ...interface{})
```

Like `SerialDeps`, but passes a context to dependencies.

## Function Wrapper

### F

```go
func F(target interface{}, args ...interface{}) Fn
```

Wraps a function with arguments for use with `Deps`. Allows calling the same function with different arguments as separate dependencies.

```go
func ProcessFile(filename string) error {
    return sh.Run("process", filename)
}

func ProcessAll() {
    st.Deps(
        st.F(ProcessFile, "file1.txt"),
        st.F(ProcessFile, "file2.txt"),
    )
}
```

Supported argument types:
- `string`
- `int`
- `bool`
- `float64`
- `time.Duration`

Note: Do not pass `context.Context` to `F()`. The context is automatically provided.

### Fn Interface

```go
type Fn interface {
    Name() string
    ID() string
    Run(ctx context.Context) error
}
```

Interface implemented by `F()`. You can implement custom `Fn` types for advanced use cases.

## Error Functions

### Fatal

```go
func Fatal(code int, args ...interface{}) error
```

Returns an error that causes Stave to exit with the given exit code.

```go
func Build() error {
    if !fileExists("go.mod") {
        return st.Fatal(2, "not a Go module")
    }
    return nil
}
```

### Fatalf

```go
func Fatalf(code int, format string, args ...interface{}) error
```

Like `Fatal`, but with printf-style formatting.

```go
func Deploy(env string) error {
    if env != "prod" && env != "staging" {
        return st.Fatalf(2, "invalid environment: %s", env)
    }
    return nil
}
```

### ExitStatus

```go
func ExitStatus(err error) int
```

Extracts the exit status from an error. Returns 0 for nil, 1 for errors without a status, or the error's exit status.

```go
func Build() error {
    err := sh.Run("go", "build")
    code := st.ExitStatus(err)
    if code != 0 {
        return st.Fatal(code, "build failed")
    }
    return nil
}
```

## Runtime Functions

### Verbose

```go
func Verbose() bool
```

Returns true if Stave was run with the `-v` flag.

```go
func Build() error {
    if st.Verbose() {
        fmt.Println("Starting build...")
    }
    return sh.Run("go", "build")
}
```

### Debug

```go
func Debug() bool
```

Returns true if Stave was run with the `-d` flag.

### Info

```go
func Info() bool
```

Returns true if Stave was run with the `-i` flag.

### GoCmd

```go
func GoCmd() string
```

Returns the Go command being used (default: "go").

```go
func Build() error {
    return sh.Run(st.GoCmd(), "build", "./...")
}
```

### CacheDir

```go
func CacheDir() string
```

Returns the cache directory where compiled binaries are stored.

### HashFast

```go
func HashFast() bool
```

Returns true if fast hashing mode is enabled.

### IgnoreDefault

```go
func IgnoreDefault() bool
```

Returns true if the default target should be ignored.

### EnableColor

```go
func EnableColor() bool
```

Returns true if colored output is enabled.

### TargetColor

```go
func TargetColor() string
```

Returns the ANSI escape sequence for the configured target color.

## Namespace Type

### Namespace

```go
type Namespace struct{}
```

Empty struct used to create namespaced targets. Define a type alias and attach methods.

```go
type Build st.Namespace

func (Build) Docker() error {
    return sh.Run("docker", "build", ".")
}

func (Build) Binary() error {
    return sh.Run("go", "build", "-o", "app", ".")
}
```

Usage:

```bash
stave build:docker
stave build:binary
```

## Constants

### Environment Variable Names

```go
const (
    CacheEnv         = "STAVEFILE_CACHE"
    VerboseEnv       = "STAVEFILE_VERBOSE"
    DebugEnv         = "STAVEFILE_DEBUG"
    InfoEnv          = "STAVEFILE_INFO"
    GoCmdEnv         = "STAVEFILE_GOCMD"
    IgnoreDefaultEnv = "STAVEFILE_IGNOREDEFAULT"
    HashFastEnv      = "STAVEFILE_HASHFAST"
    EnableColorEnv   = "STAVEFILE_ENABLE_COLOR"
    TargetColorEnv   = "STAVEFILE_TARGET_COLOR"
    DryRunRequestedEnv = "STAVEFILE_DRYRUN"
    DryRunPossibleEnv  = "STAVEFILE_DRYRUN_POSSIBLE"
)
```

## Color Constants

```go
type Color int

const (
    Black Color = iota
    Red
    Green
    Yellow
    Blue
    Staventa  // Magenta
    Cyan
    White
    BrightBlack
    BrightRed
    BrightGreen
    BrightYellow
    BrightBlue
    BrightStaventa
    BrightCyan
    BrightWhite
)
```

```go
const AnsiColorReset = "\033[0m"
var DefaultTargetAnsiColor = ansiColor[Cyan]
```

