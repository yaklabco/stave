# pkg/sh

[Home](../index.md) > [API Reference](cli.md) > pkg/sh

Package `sh` provides functions for running shell commands.

```go
import "github.com/yaklabco/stave/pkg/sh"
```

## Command Execution

### Run

```go
func Run(cmd string, args ...string) error
```

Run a command. Stdout is printed only if verbose mode is enabled.

```go
err := sh.Run("go", "build", "./...")
```

### RunV

```go
func RunV(cmd string, args ...string) error
```

Run a command, always printing stdout.

```go
err := sh.RunV("go", "test", "-v", "./...")
```

### RunWith

```go
func RunWith(env map[string]string, cmd string, args ...string) error
```

Run a command with additional environment variables.

```go
err := sh.RunWith(map[string]string{"CGO_ENABLED": "0"}, "go", "build", ".")
```

### RunWithV

```go
func RunWithV(env map[string]string, cmd string, args ...string) error
```

Run with environment, always printing stdout.

## Output Capture

### Output

```go
func Output(cmd string, args ...string) (string, error)
```

Run a command and return stdout as a string.

```go
version, err := sh.Output("go", "version")
```

### OutputWith

```go
func OutputWith(env map[string]string, cmd string, args ...string) (string, error)
```

Run with environment and return stdout.

```go
out, err := sh.OutputWith(map[string]string{"GOOS": "linux"}, "go", "env", "GOOS")
```

## Full Control

### Exec

```go
func Exec(env map[string]string, stdout, stderr io.Writer, cmd string, args ...string) (bool, error)
```

Execute a command with full control over I/O.

Parameters:

- `env`: Additional environment variables (merged with current environment)
- `stdout`: Writer for standard output (nil to discard)
- `stderr`: Writer for standard error
- `cmd`: Command to execute
- `args`: Command arguments

Returns:

- `bool`: Whether the command ran (false if not found or not executable)
- `error`: Execution error, if any

```go
ran, err := sh.Exec(
    map[string]string{"DEBUG": "1"},
    os.Stdout,
    os.Stderr,
    "my-command", "arg1",
)
```

## Command Factories

### RunCmd

```go
func RunCmd(cmd string, args ...string) func(args ...string) error
```

Create a reusable command runner.

```go
var goCmd = sh.RunCmd("go")

func Build() error {
    return goCmd("build", "./...")
}
```

### OutCmd

```go
func OutCmd(cmd string, args ...string) func(args ...string) (string, error)
```

Create a reusable output-capturing command.

```go
var gitOut = sh.OutCmd("git")

func Hash() (string, error) {
    return gitOut("rev-parse", "HEAD")
}
```

## File Helpers

### Rm

```go
func Rm(path string) error
```

Remove a file or directory recursively. Returns nil if the path does not exist.

```go
err := sh.Rm("build/")
```

### Copy

```go
func Copy(dst, src string) error
```

Copy a file, preserving permissions.

```go
err := sh.Copy("dist/app", "build/app")
```

## Error Inspection

### CmdRan

```go
func CmdRan(err error) bool
```

Returns true if the command ran (even if it exited non-zero). Returns false if the command was not found or not executable.

```go
err := sh.Run("maybe-missing")
if !sh.CmdRan(err) {
    fmt.Println("command not found")
}
```

### ExitStatus

```go
func ExitStatus(err error) int
```

Extract exit code from an error. Returns 0 for nil, 1 for unrecognized errors.

```go
err := sh.Run("false")
code := sh.ExitStatus(err)  // 1
```

## Environment Variable Expansion

All commands and arguments undergo `$VAR` expansion:

```go
sh.Run("echo", "$HOME")  // expands $HOME
```

Variables from the `env` parameter in `RunWith`/`OutputWith`/`Exec` are also available for expansion.

## Dry-Run Behavior

When `--dryrun` is active, all functions print `DRYRUN: cmd args...` instead of executing. `Rm` and `Copy` also respect dry-run mode.

---

## See Also

- [Shell Commands](../user-guide/shell-commands.md) - Usage guide
- [Advanced Topics](../user-guide/advanced.md) - Dry-run mode
- [Home](../index.md)
