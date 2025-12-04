# Quickstart

Get up and running with Stave in under 5 minutes.

## Create Your First Stavefile

Create a file named `stavefile.go` in your project root:

```go
//go:build stave

package main

import "fmt"

// Hello prints a greeting.
func Hello() {
    fmt.Println("Hello from Stave!")
}

// Build compiles the project.
func Build() error {
    fmt.Println("Building...")
    return nil
}
```

Key points:
- The `//go:build stave` directive marks this as a stavefile
- The package must be `main`
- Exported functions become targets
- Function comments become target descriptions

## Run Your First Target

List available targets:

```bash
stave -l
```

Output:

```
Targets:
  build    compiles the project.
  hello    prints a greeting.
```

Run a target:

```bash
stave hello
```

Output:

```
Hello from Stave!
```

## Add Dependencies

Modify your stavefile to use dependencies:

```go
//go:build stave

package main

import (
    "fmt"

    "github.com/yaklabco/stave/pkg/st"
)

// Lint runs the linter.
func Lint() error {
    fmt.Println("Running linter...")
    return nil
}

// Test runs tests.
func Test() error {
    fmt.Println("Running tests...")
    return nil
}

// Build compiles the project after linting and testing.
func Build() error {
    st.Deps(Lint, Test)  // Run Lint and Test in parallel
    fmt.Println("Building...")
    return nil
}
```

Run build:

```bash
stave build
```

Output (order of Lint/Test may vary due to parallel execution):

```
Running linter...
Running tests...
Building...
```

## Set a Default Target

Add a `Default` variable to specify which target runs when no target is given:

```go
//go:build stave

package main

import (
    "fmt"

    "github.com/yaklabco/stave/pkg/st"
)

// Default is the target to run when none is specified.
var Default = Build

func Lint() error {
    fmt.Println("Running linter...")
    return nil
}

func Test() error {
    fmt.Println("Running tests...")
    return nil
}

// Build is the default target.
func Build() error {
    st.Deps(Lint, Test)
    fmt.Println("Building...")
    return nil
}
```

Now simply run:

```bash
stave
```

This runs the `Build` target.

## Run Shell Commands

Use the `sh` package to run external commands:

```go
//go:build stave

package main

import (
    "github.com/yaklabco/stave/pkg/sh"
    "github.com/yaklabco/stave/pkg/st"
)

func Lint() error {
    return sh.Run("golangci-lint", "run", "./...")
}

func Test() error {
    return sh.Run("go", "test", "-v", "./...")
}

func Build() error {
    st.Deps(Lint, Test)
    return sh.Run("go", "build", "-o", "myapp", ".")
}
```

## Using Verbose Mode

Run with `-v` to see more output:

```bash
stave -v build
```

This shows which dependencies are running and command execution details.

## Project Structure

A typical project with Stave:

```
myproject/
  stavefile.go      # Build targets
  main.go           # Application code
  ...
```

For larger projects, you can use a `stavefiles/` directory:

```
myproject/
  stavefiles/
    stavefile.go    # Main targets
    helpers.go      # Helper functions (no build tag needed)
  main.go
  ...
```

## Next Steps

- [Stavefiles](../user-guide/stavefiles.md) - Learn more about stavefile structure
- [Dependencies](../user-guide/dependencies.md) - Advanced dependency management
- [Shell Commands](../user-guide/shell-commands.md) - Running external commands
- [Configuration](../user-guide/configuration.md) - Customizing Stave behavior

