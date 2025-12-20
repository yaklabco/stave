# Dependencies

[Home](../index.md) > [User Guide](stavefiles.md) > Dependencies

Dependencies declare that a target requires other targets to run first.

## st.Deps

`st.Deps` runs dependencies in parallel:

```go
func Build() error {
    st.Deps(Generate, Compile)
    return link()
}
```

`Generate` and `Compile` run concurrently. `Build` continues after both complete.

## st.SerialDeps

`st.SerialDeps` runs dependencies sequentially:

```go
func Deploy() error {
    st.SerialDeps(Build, Test, Push)
    return nil
}
```

`Build` runs first, then `Test`, then `Push`.

## Context Variants

Pass a context to dependencies:

```go
func Build(ctx context.Context) error {
    st.CtxDeps(ctx, Generate, Compile)
    return link()
}

func Deploy(ctx context.Context) error {
    st.SerialCtxDeps(ctx, Build, Test, Push)
    return nil
}
```

The context is passed to dependencies that accept it.

## Once Semantics

Each dependency runs exactly once per Stave invocation, regardless of how many times it appears in the dependency graph:

```go
func A() { st.Deps(C) }
func B() { st.Deps(C) }
func All() { st.Deps(A, B) }  // C runs once
```

## Error Handling

If a dependency fails, execution stops and the error propagates:

```go
func Build() error {
    st.Deps(Generate)  // if Generate fails, Build does not continue
    return compile()
}
```

Multiple parallel dependencies that fail report all errors.

## Circular Dependency Detection

Stave automatically detects circular dependencies between targets and fails fast with a clear error message. For example, if `A` depends on `B` and `B` depends on `A`, Stave will stop before executing any target and report the cycle:

```text
stave: detected circular dependency: A -> B -> A
```

Typical causes and fixes:

- A target indirectly depends on itself via a long chain. Inspect the reported cycle and break it by removing the unnecessary dependency.
- Use `st.SerialDeps` to clarify order when sequencing is required, rather than introducing back-edges in the graph.

Notes:

- Cycles are detected across namespaces as well (e.g., `build:All -> test:Prepare -> build:All`).
- Each distinct set of target arguments passed via `st.F` participates in cycle detection independently, so a cycle can include a specific argumentized invocation.

## Dependencies with Arguments

Use `st.F` to wrap a target with arguments:

```go
func Deploy(env string) error {
    // ...
}

func DeployAll() {
    st.Deps(
        st.F(Deploy, "staging"),
        st.F(Deploy, "production"),
    )
}
```

Each `st.F` call with different arguments is treated as a distinct dependency (each runs once).

## Watch Mode Dependencies

When using [Watch Mode](watch.md), use `watch.Deps` instead of `st.Deps`. This ensures that the dependencies are aware of the cancellable context and will be terminated if a file change triggers a re-run.

```go
import "github.com/yaklabco/stave/pkg/watch"

func All() {
    watch.Deps(Build, Test)
}
```

## Dependency Trees

Complex builds compose naturally:

```go
func All() {
    st.Deps(Build, Test, Lint)
}

func Build() error {
    st.Deps(Generate)
    return sh.Run("go", "build", "./...")
}

func Test() error {
    st.Deps(Build)
    return sh.Run("go", "test", "./...")
}

func Lint() error {
    return sh.Run("golangci-lint", "run")
}

func Generate() error {
    return sh.Run("go", "generate", "./...")
}
```

Running `stave all`:

1. `Generate` runs (required by `Build`)
2. `Build`, `Test`, and `Lint` become ready
3. `Build` completes (after `Generate`)
4. `Test` runs (after `Build`)
5. `Lint` runs in parallel with the above (no dependencies)

---

## See Also

- [Targets](targets.md) - Defining targets
- [Arguments](arguments.md) - Targets with arguments
- [pkg/st API](../api-reference/st.md) - Function reference
- [Home](../index.md)
