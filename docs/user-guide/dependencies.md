# Dependencies

Stave provides powerful dependency management through the `st` package. Dependencies are executed exactly once, even if multiple targets depend on them.

## Basic Dependencies with st.Deps

Use `st.Deps()` to declare dependencies that run in parallel:

```go
//go:build stave

package main

import (
    "fmt"
    "github.com/yaklabco/stave/pkg/st"
)

func Lint() error {
    fmt.Println("Linting...")
    return nil
}

func Test() error {
    fmt.Println("Testing...")
    return nil
}

func Build() error {
    // Lint and Test run in parallel
    st.Deps(Lint, Test)
    fmt.Println("Building...")
    return nil
}
```

When you run `stave build`:
1. `Lint` and `Test` start simultaneously (parallel)
2. `Build` waits for both to complete
3. `Build` runs after dependencies finish

## Serial Dependencies with st.SerialDeps

Use `st.SerialDeps()` when dependencies must run sequentially:

```go
func Deploy() error {
    // These run one after another, in order
    st.SerialDeps(Build, Test, Push)
    return nil
}
```

Use serial dependencies when:
- Order matters (e.g., build before test)
- Resources are limited (e.g., single database connection)
- Operations would conflict if run in parallel

## Context-Aware Dependencies

For long-running operations or when you need cancellation support, use the context variants:

```go
import "context"

func Build(ctx context.Context) error {
    // Pass context to dependencies
    st.CtxDeps(ctx, Lint, Test)
    
    // Use context for your own operations
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue building
    }
    return nil
}
```

Available context functions:
- `st.CtxDeps(ctx, fns...)` - Parallel with context
- `st.SerialCtxDeps(ctx, fns...)` - Serial with context

## Once Semantics

Each dependency runs exactly once per Stave invocation, regardless of how many targets depend on it:

```go
func Common() {
    fmt.Println("Common runs once")
}

func A() {
    st.Deps(Common)
    fmt.Println("A")
}

func B() {
    st.Deps(Common)
    fmt.Println("B")
}

func All() {
    st.Deps(A, B)  // Common runs only once
}
```

Running `stave all` outputs:

```
Common runs once
A
B
```

## Dependencies with Arguments

Use `st.F()` to wrap functions that take arguments:

```go
func ProcessFile(filename string) error {
    fmt.Printf("Processing %s\n", filename)
    return nil
}

func ProcessAll() {
    st.Deps(
        st.F(ProcessFile, "file1.txt"),
        st.F(ProcessFile, "file2.txt"),
        st.F(ProcessFile, "file3.txt"),
    )
}
```

Supported argument types:
- `string`
- `int`
- `bool`
- `float64`
- `time.Duration`

Note: Do not pass `context.Context` to `st.F()`. The context is automatically provided if your function accepts one.

## Error Handling

If any dependency fails, execution stops and the error propagates:

```go
func MayFail() error {
    return errors.New("something went wrong")
}

func Build() error {
    st.Deps(MayFail, AlsoRuns)  // AlsoRuns may or may not complete
    // This line won't execute if MayFail returns an error
    return nil
}
```

When a dependency fails:
1. Other parallel dependencies may continue briefly
2. Once all running deps complete or fail, execution stops
3. The first error is reported

For custom exit codes, use `st.Fatal()`:

```go
func Critical() error {
    if somethingBad {
        return st.Fatal(2, "critical failure")
    }
    return nil
}
```

## Dependency Trees

Dependencies can have their own dependencies, forming a tree:

```go
func Deps() error {
    return sh.Run("go", "mod", "download")
}

func Generate() error {
    st.Deps(Deps)
    return sh.Run("go", "generate", "./...")
}

func Lint() error {
    st.Deps(Generate)
    return sh.Run("golangci-lint", "run")
}

func Test() error {
    st.Deps(Generate)
    return sh.Run("go", "test", "./...")
}

func Build() error {
    st.Deps(Lint, Test)  // Both depend on Generate, which depends on Deps
    return sh.Run("go", "build", "./...")
}
```

The dependency tree for `Build`:

```
Build
  Lint
    Generate
      Deps
  Test
    Generate (already ran, skipped)
      Deps (already ran, skipped)
```

## Best Practices

1. **Use parallel deps by default** - `st.Deps()` is faster unless order matters
2. **Keep deps pure** - Dependencies should be idempotent
3. **Handle errors** - Return errors rather than calling `os.Exit()`
4. **Use context** - For long operations that may need cancellation
5. **Document dependencies** - Make dependency relationships clear in comments

## Common Patterns

### Build Pipeline

```go
var Default = All

func All() {
    st.SerialDeps(Clean, Deps, Generate)
    st.Deps(Lint, Test)
    st.Deps(Build)
}
```

### Conditional Dependencies

```go
func Build() error {
    if os.Getenv("SKIP_LINT") == "" {
        st.Deps(Lint)
    }
    return sh.Run("go", "build", "./...")
}
```

### Parallel Build Matrix

```go
func BuildAll() {
    st.Deps(
        st.F(BuildFor, "linux", "amd64"),
        st.F(BuildFor, "linux", "arm64"),
        st.F(BuildFor, "darwin", "amd64"),
        st.F(BuildFor, "darwin", "arm64"),
    )
}

func BuildFor(goos, goarch string) error {
    env := map[string]string{"GOOS": goos, "GOARCH": goarch}
    output := fmt.Sprintf("bin/myapp-%s-%s", goos, goarch)
    return sh.RunWith(env, "go", "build", "-o", output, "./...")
}
```

