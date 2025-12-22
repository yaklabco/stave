# Targets

[Home](../index.md) > [User Guide](stavefiles.md) > Targets

A target is an exported function in a stavefile that Stave can invoke from the command line.

## Function Signatures

Targets must have one of these signatures:

```go
func Target()
func Target() error
func Target(ctx context.Context)
func Target(ctx context.Context) error
```

Targets may also accept typed arguments after the optional context. See [Arguments](arguments.md).

## Naming and Invocation

Target names are case-insensitive. A function named `Build` can be invoked as:

```bash
stave build
stave Build
stave BUILD
```

## Documentation

The first sentence of a function's doc comment becomes its synopsis in `stave -l`:

```go
// Build compiles the application. It produces a binary in ./bin.
func Build() error {
    // ...
}
```

Output of `stave -l`:

```text
Targets:
  build    compiles the application.
```

Use `stave -i build` to see the full doc comment.

## Default Target

Set a default target to run when no target is specified:

```go
var Default = Build
```

Now `stave` with no arguments runs `Build`.

To ignore the default and list targets instead, set `STAVEFILE_IGNOREDEFAULT=1`.

## Aliases

Define alternative names for targets:

```go
var Aliases = map[string]any{
    "b": Build,
    "t": Test,
}
```

Now `stave b` runs `Build` and `stave t` runs `Test`.

## Importing Targets

Import targets from other packages using the `stave:import` directive:

```go
import (
    // stave:import
    "github.com/yourorg/shared/buildtasks"

    // stave:import deploy
    "github.com/yourorg/shared/deploytasks"
)
```

- Root import: Targets are available by their original names.
- Aliased import: Targets are prefixed with the alias (e.g., `deploy:production`).

The imported package must contain valid target functions.

## Exit Codes

Return an error to indicate failure:

```go
func Build() error {
    if err := sh.Run("go", "build"); err != nil {
        return err
    }
    return nil
}
```

Use `st.Fatal` or `st.Fatalf` to exit with a specific code:

```go
func Deploy() error {
    if os.Getenv("ENV") == "" {
        return st.Fatal(2, "ENV must be set")
    }
    // ...
}
```

---

## See Also

- [Stavefiles](stavefiles.md) - File conventions
- [Dependencies](dependencies.md) - Target dependencies
- [Namespaces](namespaces.md) - Grouping targets
- [Arguments](arguments.md) - Typed arguments
- [Git Hooks](hooks.md) - Run targets automatically on Git events
- [Home](../index.md)
