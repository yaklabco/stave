# Arguments

[Home](../index.md) > [User Guide](stavefiles.md) > Arguments

Targets can accept typed arguments from the command line.

## Supported Types

- `string`
- `int`
- `bool`
- `float64`
- `time.Duration`

## Defining Arguments

Add parameters after the optional context:

```go
func Greet(name string, times int) error {
    for i := 0; i < times; i++ {
        fmt.Printf("Hello, %s!\n", name)
    }
    return nil
}
```

## Invoking with Arguments

Pass arguments positionally after the target name:

```bash
stave greet Alice 3
```

Output:

```
Hello, Alice!
Hello, Alice!
Hello, Alice!
```

## Type Parsing

Arguments are parsed according to their declared type:

| Type | Example Input | Parsed Value |
|------|---------------|--------------|
| `string` | `hello` | `"hello"` |
| `int` | `42` | `42` |
| `bool` | `true`, `false`, `1`, `0` | `true`, `false` |
| `float64` | `3.14` | `3.14` |
| `time.Duration` | `5m30s` | `5*time.Minute + 30*time.Second` |

## Arguments with Context

Context and arguments can be combined:

```go
func Deploy(ctx context.Context, env string, dryRun bool) error {
    if dryRun {
        fmt.Printf("Would deploy to %s\n", env)
        return nil
    }
    // actual deployment
    return nil
}
```

```bash
stave deploy production false
```

## Arguments in Dependencies

Use `st.F` to pass arguments when declaring dependencies:

```go
func Deploy(env string) error {
    fmt.Printf("Deploying to %s\n", env)
    return nil
}

func DeployAll() {
    st.Deps(
        st.F(Deploy, "staging"),
        st.F(Deploy, "production"),
    )
}
```

Each `st.F` call with distinct arguments is a separate dependency. Both run (in parallel by default), each exactly once.

## Argument Errors

Missing or malformed arguments cause Stave to exit with code 2:

```bash
stave greet Alice
# Error: not enough arguments for target "Greet", expected 2, got 1
```

```bash
stave greet Alice notanumber
# Error: can't convert argument "notanumber" to int
```

## Viewing Argument Requirements

Use `stave -i` to see a target's arguments:

```bash
stave -i greet
```

Output:

```
Usage:

    stave greet <name> <times>
```

---

## See Also

- [Targets](targets.md) - Defining targets
- [Dependencies](dependencies.md) - Using `st.F` with arguments
- [Home](../index.md)

