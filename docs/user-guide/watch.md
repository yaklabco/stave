# Watch Mode

[Home](../index.md) > [User Guide](stavefiles.md) > Watch Mode

Watch mode allows Stave to monitor your project's files and automatically re-run targets whenever changes are detected. This is particularly useful for development workflows like auto-testing or auto-rebuilding.

Stave supports watching multiple targets simultaneously. If you specify multiple targets on the command line, each one will be monitored and re-run independently when its watched files change.

In the target list (`stave -l`), watch targets are identified by a `[W]` suffix.

## Basic Usage

To enable watch mode for a target, use the `watch.Watch` function in your `stavefile.go`:

```go
import (
    "github.com/yaklabco/stave/pkg/sh"
    "github.com/yaklabco/stave/pkg/watch"
)

// WatchTests watches for Go file changes and runs tests
func WatchTests() error {
    watch.Watch("**/*.go")
    
    return sh.RunV("go", "test", "./...")
}
```

When you run `stave WatchTests`, Stave will:

1. Run the target once.
2. Monitor all `.go` files in the project.
3. If a `.go` file changes, it will cancel the current execution (if it's still running) and start it again.

If you run multiple targets, e.g., `stave WatchTests WatchBuild`, both will be watched and re-run as needed.

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

### Context-Aware Shell Helpers

To properly support cancellation, you should use shell helpers that are aware of the target's context.

Historically, this required using helpers from `pkg/watch`. However, in the current version of Stave, the standard helpers in `pkg/sh` are now automatically context-aware and will use the active target's context. This means they will be automatically terminated when a re-run is triggered.

```go
import "github.com/yaklabco/stave/pkg/sh"

func Build() error {
    watch.Watch("main.go")
    // This will now be automatically terminated on re-run
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
    
    if err := sh.Run("go", "build", "."); err != nil {
        return err
    }
    
    return sh.Run("./myapp")
}
```

---

## See Also

- [pkg/watch API Reference](../api-reference/watch.md)
- [Shell Commands](shell-commands.md)
- [Home](../index.md)
