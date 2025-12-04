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

Returns true if `-v` or `STAVEFILE_VERBOSE=1` is set.

```go
if st.Verbose() {
    fmt.Println("Building with verbose output...")
}
```

### Debug

```go
func Debug() bool
```

Returns true if `-d` or `STAVEFILE_DEBUG=1` is set.

### Info

```go
func Info() bool
```

Returns true if `-i` or `STAVEFILE_INFO=1` is set.

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

Returns true if `STAVEFILE_HASHFAST=1` is set.

### IgnoreDefault

```go
func IgnoreDefault() bool
```

Returns true if `STAVEFILE_IGNOREDEFAULT=1` is set.

### EnableColor

```go
func EnableColor() bool
```

Returns true if `STAVEFILE_ENABLE_COLOR=1` is set.

### TargetColor

```go
func TargetColor() string
```

Returns the ANSI color code for target names.

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

