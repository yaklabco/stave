# Watch Mode

[Home](../index.md) > [User Guide](stavefiles.md) > Watch Mode

Watch mode allows Stave to monitor your project's files and automatically re-run targets whenever changes are detected. This is particularly useful for development workflows like auto-testing or auto-rebuilding.

## Basic Usage

To enable watch mode for a target, use the `watch.Watch` function in your `stavefile.go`:

```go
import (
    "github.com/yaklabco/stave/pkg/watch"
)

// WatchTests watches for Go file changes and runs tests
func WatchTests() error {
    watch.Watch("**/*.go")
    
    return watch.RunV("go", "test", "./...")
}
```

When you run `stave WatchTests`, Stave will:

1. Run the target once.
2. Monitor all `.go` files in the project.
3. If a `.go` file changes, it will cancel the current execution (if it's still running) and start it again.

## Glob Patterns

`watch.Watch` accepts one or more glob patterns. See [pkg/watch API Reference](../api-reference/watch.md) for supported wildcard syntax.

```go
watch.Watch("src/**/*.go", "templates/*.tmpl", "config.yaml")
```

## Cancellable Contexts

Stave's watch mode works by using Go's `context.Context`. When a file change is detected:

1. The context associated with the current target is cancelled.
2. Stave waits for the current execution to finish (handling the cancellation).
3. Stave re-runs the target with a new context.

### Watch-Aware Shell Helpers

To properly support cancellation, you should use the shell helpers provided by `pkg/watch` instead of `pkg/sh`. These helpers automatically listen for context cancellation and terminate the underlying processes.

```go
// GOOD: This will be terminated on re-run
func Build() error {
    watch.Watch("main.go")
    return watch.Run("go", "build", "-o", "myapp", "main.go")
}

// BAD: This might keep running even after a re-run is triggered
func Build() error {
    watch.Watch("main.go")
    return sh.Run("go", "build", "-o", "myapp", "main.go")
}
```

## Watch-Aware Dependencies

If your target has dependencies, use `watch.Deps` instead of `st.Deps`. This ensures that dependencies are also aware of the watch mode's cancellable context.

```go
func All() {
    watch.Deps(Build, Test)
}
```

## Example: Complex Watch Workflow

```go
func Dev() {
    // Watch all files in the current directory
    watch.Watch("**/*")
    
    // Run multiple commands in sequence
    // If any file changes while these are running, the sequence
    // will be cancelled and restarted from the beginning.
    watch.Deps(Generate)
    
    if err := watch.Run("go", "build", "."); err != nil {
        return err
    }
    
    return watch.Run("./myapp")
}
```

---

## See Also

- [pkg/watch API Reference](../api-reference/watch.md)
- [Shell Commands](shell-commands.md)
- [Home](../index.md)
