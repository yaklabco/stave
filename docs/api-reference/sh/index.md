# pkg/sh

The `sh` package provides utilities for running shell commands.

```go
import "github.com/yaklabco/stave/pkg/sh"
```

## Overview

The `sh` package provides functions to:
- Run external commands
- Capture command output
- Manage command environments
- Handle file operations (copy, remove)

All command functions respect dry-run mode and verbose settings.

## Running Commands

### Run

```go
func Run(cmd string, args ...string) error
```

Runs a command, inheriting the current environment. Stdout goes to stdout only if verbose mode is enabled. Stderr always goes to stderr.

```go
func Build() error {
    return sh.Run("go", "build", "-o", "app", ".")
}
```

### RunV

```go
func RunV(cmd string, args ...string) error
```

Like `Run`, but always sends stdout to the terminal regardless of verbose mode.

```go
func Test() error {
    return sh.RunV("go", "test", "-v", "./...")
}
```

### RunWith

```go
func RunWith(env map[string]string, cmd string, args ...string) error
```

Like `Run`, but adds environment variables for the command.

```go
func BuildLinux() error {
    env := map[string]string{
        "GOOS":   "linux",
        "GOARCH": "amd64",
    }
    return sh.RunWith(env, "go", "build", "-o", "app-linux", ".")
}
```

### RunWithV

```go
func RunWithV(env map[string]string, cmd string, args ...string) error
```

Like `RunWith`, but always sends stdout to the terminal.

```go
func Test() error {
    env := map[string]string{"VERBOSE": "1"}
    return sh.RunWithV(env, "go", "test", "./...")
}
```

## Capturing Output

### Output

```go
func Output(cmd string, args ...string) (string, error)
```

Runs a command and returns its stdout as a string. Trailing newlines are trimmed.

```go
func GetVersion() (string, error) {
    return sh.Output("git", "describe", "--tags")
}
```

### OutputWith

```go
func OutputWith(env map[string]string, cmd string, args ...string) (string, error)
```

Like `Output`, but adds environment variables.

```go
func GetGoVersion() (string, error) {
    env := map[string]string{"GOPROXY": "direct"}
    return sh.OutputWith(env, "go", "version")
}
```

## Low-Level Execution

### Exec

```go
func Exec(env map[string]string, stdout, stderr io.Writer, cmd string, args ...string) (bool, error)
```

Full control over command execution. Returns whether the command ran (vs. not found).

Parameters:
- `env` - Additional environment variables (merged with current env)
- `stdout` - Writer for stdout (nil to discard)
- `stderr` - Writer for stderr (nil to discard)
- `cmd` - Command to run
- `args` - Command arguments

```go
func CustomBuild() error {
    var stdout, stderr bytes.Buffer
    ran, err := sh.Exec(nil, &stdout, &stderr, "go", "build", ".")
    if !ran {
        return fmt.Errorf("go not found")
    }
    if err != nil {
        return fmt.Errorf("build failed: %s", stderr.String())
    }
    return nil
}
```

### CmdRan

```go
func CmdRan(err error) bool
```

Reports whether a command actually ran (vs. command not found or not executable).

```go
func OptionalTool() error {
    err := sh.Run("optional-tool", "check")
    if !sh.CmdRan(err) {
        fmt.Println("optional-tool not installed, skipping")
        return nil
    }
    return err
}
```

### ExitStatus

```go
func ExitStatus(err error) int
```

Returns the exit code from a command error. Returns 0 for nil, 1 for unknown errors.

```go
func Build() error {
    err := sh.Run("go", "build", ".")
    if code := sh.ExitStatus(err); code != 0 {
        return st.Fatal(code, "build failed")
    }
    return nil
}
```

## Command Factories

### RunCmd

```go
func RunCmd(cmd string, args ...string) func(args ...string) error
```

Returns a function that runs the given command. Useful for creating command aliases.

```go
var go_ = sh.RunCmd("go")
var docker = sh.RunCmd("docker")

func Build() error {
    if err := go_("build", "-o", "app", "."); err != nil {
        return err
    }
    return docker("build", "-t", "myapp", ".")
}
```

Pre-baked arguments are passed first:

```go
var goInstall = sh.RunCmd("go", "install")

func Install() error {
    return goInstall("./...")  // Runs: go install ./...
}
```

### OutCmd

```go
func OutCmd(cmd string, args ...string) func(args ...string) (string, error)
```

Like `RunCmd`, but returns a function that captures output.

```go
var gitDescribe = sh.OutCmd("git", "describe")

func Version() (string, error) {
    return gitDescribe("--tags")  // Runs: git describe --tags
}
```

## File Operations

### Rm

```go
func Rm(path string) error
```

Removes a file or directory (including non-empty directories). Does not error if the path doesn't exist.

```go
func Clean() error {
    if err := sh.Rm("dist"); err != nil {
        return err
    }
    return sh.Rm("coverage.out")
}
```

### Copy

```go
func Copy(dst, src string) error
```

Copies a file from src to dst, preserving file mode. Overwrites dst if it exists.

```go
func Install() error {
    return sh.Copy("/usr/local/bin/myapp", "./myapp")
}
```

## Environment Variable Expansion

Command and arguments support environment variable expansion using `$FOO` or `${FOO}` syntax:

```go
func Build() error {
    // $GOPATH will be expanded from env or RunWith map
    return sh.Run("go", "build", "-o", "$GOPATH/bin/myapp", ".")
}

func BuildWithEnv() error {
    env := map[string]string{"OUTPUT": "./dist/app"}
    return sh.RunWith(env, "go", "build", "-o", "$OUTPUT", ".")
}
```

## Dry-Run Mode

When Stave is run with `--dryrun`, all `sh.Run*` commands print what would be executed instead of running:

```bash
$ stave --dryrun build
DRYRUN: go build -o app .
DRYRUN: docker build -t myapp .
```

The `sh.Rm` and `sh.Copy` functions also respect dry-run mode.

## Error Handling

Command errors include the exit code, which can be used with `st.Fatal`:

```go
func Build() error {
    err := sh.Run("go", "build", ".")
    if err != nil {
        // Error already contains exit code info
        // Returning it will exit with the same code
        return err
    }
    return nil
}
```

For custom error handling:

```go
func Build() error {
    err := sh.Run("make")
    if err != nil {
        code := sh.ExitStatus(err)
        switch code {
        case 2:
            return st.Fatal(code, "missing dependencies")
        default:
            return st.Fatalf(code, "make failed with code %d", code)
        }
    }
    return nil
}
```

