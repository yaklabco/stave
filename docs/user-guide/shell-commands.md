# Shell Commands

[Home](../index.md) > [User Guide](stavefiles.md) > Shell Commands

The `pkg/sh` package provides functions for running external commands.

## Basic Execution

### sh.Run

Run a command, printing stdout only if verbose mode is enabled:

```go
err := sh.Run("go", "build", "./...")
```

### sh.RunV

Run a command, always printing stdout:

```go
err := sh.RunV("go", "test", "-v", "./...")
```

## Capturing Output

### sh.Output

Capture stdout as a string:

```go
version, err := sh.Output("go", "version")
if err != nil {
    return err
}
fmt.Println(version)
```

### sh.OutputWith

Capture output with custom environment:

```go
out, err := sh.OutputWith(map[string]string{"GOOS": "linux"}, "go", "env", "GOOS")
```

## Environment Variables

### sh.RunWith

Run with custom environment variables:

```go
err := sh.RunWith(map[string]string{
    "CGO_ENABLED": "0",
    "GOOS":        "linux",
}, "go", "build", "-o", "app-linux", ".")
```

### sh.RunWithV

Same as `RunWith`, always printing stdout:

```go
err := sh.RunWithV(env, "go", "build", "./...")
```

## Variable Expansion

Commands and arguments undergo `$VAR` expansion using the environment:

```go
// $HOME expands to the user's home directory
err := sh.Run("ls", "$HOME")
```

Custom environment variables from `RunWith`/`OutputWith` are also expanded.

## Full Control

### sh.Exec

For complete control over stdin, stdout, and stderr:

```go
ran, err := sh.Exec(
    map[string]string{"DEBUG": "1"},  // environment
    os.Stdout,                         // stdout writer
    os.Stderr,                         // stderr writer
    "my-command",
    "arg1", "arg2",
)
```

Returns whether the command ran (vs. not found) and any error.

## Command Factories

Create reusable command runners:

### sh.RunCmd

```go
var goCmd = sh.RunCmd("go")

func Build() error {
    return goCmd("build", "./...")
}

func Test() error {
    return goCmd("test", "./...")
}
```

### sh.OutCmd

```go
var gitOut = sh.OutCmd("git")

func CurrentBranch() (string, error) {
    return gitOut("rev-parse", "--abbrev-ref", "HEAD")
}
```

## File Helpers

### sh.Rm

Remove a file or directory (like `rm -rf`):

```go
err := sh.Rm("build/")
```

Returns nil if the path does not exist.

### sh.Copy

Copy a file:

```go
err := sh.Copy("dist/app", "build/app")
```

Preserves file permissions.

## Error Handling

### sh.CmdRan

Check if a command ran (vs. not found):

```go
err := sh.Run("nonexistent-command")
if !sh.CmdRan(err) {
    fmt.Println("command not found")
}
```

### sh.ExitStatus

Get the exit code from an error:

```go
err := sh.Run("false")  // exits with 1
code := sh.ExitStatus(err)  // returns 1
```

## Dry-Run Mode

When `--dryrun` is passed to Stave, all `sh.Run*` functions print commands instead of executing them:

```bash
stave --dryrun build
```

Output:

```text
DRYRUN: go build ./...
```

The `sh.Rm` and `sh.Copy` helpers also respect dry-run mode.

---

## See Also

- [pkg/sh API](../api-reference/sh.md) - Function reference
- [Advanced Topics](advanced.md) - Dry-run mode details
- [Home](../index.md)
