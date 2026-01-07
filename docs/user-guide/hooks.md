# Git Hooks

[Home](../index.md) > [User Guide](stavefiles.md) > Git Hooks

Stave provides native Git hook management, allowing projects to run Stave targets automatically when Git events occur (commit, push, etc.).

## Overview

Git hooks are scripts that Git executes before or after events such as `commit`, `push`, and `merge`. Stave can manage these hooks, replacing external tools like Husky or pre-commit.

Benefits:

- No external dependencies (Node.js, Python)
- Hook behavior defined as Stave targets
- Declarative configuration in `stave.yaml`
- Portable POSIX-compatible scripts

## Quick Setup

1. Add hooks configuration to `stave.yaml`:

```yaml
hooks:
  pre-commit:
    - target: Fmt
    - target: Lint
  pre-push:
    - target: Test
```

1. Install the hooks:

```bash
stave --hooks install
```

1. Commit as normal. The configured targets run automatically.

## Configuration

### Project Configuration

Define hooks in your project's `stave.yaml`:

```yaml
hooks:
  pre-commit:
    - target: fmt
    - target: lint
      args: ["--fast"]
  pre-push:
    - target: test
      args: ["./..."]
  commit-msg:
    - target: validate-commit-message
      passStdin: true
```

### Hook Target Options

Each hook entry supports the following options:

| Option      | Type     | Description                                       |
| ----------- | -------- | ------------------------------------------------- |
| `target`    | string   | Name of the Stave target to run (required)        |
| `args`      | []string | Additional arguments passed to the target         |
| `workdir`   | string   | Working directory for the target invocation       |
| `passStdin` | bool     | Forward stdin from Git to the target (see below)  |

### Working Directory

The `workdir` option allows you to specify the directory in which the Stave target should be executed. This is useful if your project structure requires certain targets to run from a specific subdirectory.

If `workdir` is a relative path, it is resolved relative to the directory containing the `stave.yaml` (or the configuration file being used). If no configuration file is used, it is resolved relative to the current working directory.

Example:

```yaml
hooks:
  pre-commit:
    - target: Lint
      workdir: ./frontend
```

### Supported Git Hooks

Stave supports all standard Git hooks:

| Hook                    | When It Runs                              |
| ----------------------- | ----------------------------------------- |
| `pre-commit`            | Before commit message editor opens        |
| `prepare-commit-msg`    | After default message, before editor      |
| `commit-msg`            | After commit message is entered           |
| `post-commit`           | After commit completes                    |
| `pre-push`              | Before push to remote                     |
| `pre-rebase`            | Before rebase starts                      |
| `post-checkout`         | After checkout completes                  |
| `post-merge`            | After merge completes                     |
| `pre-receive`           | Server-side, before refs are updated      |
| `post-receive`          | Server-side, after refs are updated       |

Unrecognized hook names generate a warning but are still installed.

## CLI Commands

### stave --hooks

List configured hooks (default when no subcommand given):

```bash
stave --hooks
```

Output:

```text
Configured Git hooks:

  pre-commit:
    - fmt
    - lint --fast
  pre-push:
    - test ./...

All 2 hook(s) installed.
```

### stave --hooks init

Display setup instructions for new projects:

```bash
stave --hooks init
```

### stave --hooks install

Install hook scripts to the Git repository:

```bash
stave --hooks install
```

This writes executable scripts to `.git/hooks/` (or the directory configured via `core.hooksPath`).

Flags:

| Flag      | Description                           |
| --------- | ------------------------------------- |
| `--force` | Overwrite existing non-Stave hooks    |

If an existing hook was not installed by Stave, the command fails unless `--force` is specified.

### stave --hooks uninstall

Remove Stave-managed hooks:

```bash
stave --hooks uninstall
```

Only removes hooks that were installed by Stave (identified by a marker comment).

Flags:

| Flag    | Description                                            |
| ------- | ------------------------------------------------------ |
| `--all` | Remove all Stave-managed hooks (not just configured)   |

### stave --hooks list

Alias for `stave --hooks` (no subcommand). Lists configured hooks and their installation status.

### stave --hooks run

Execute targets for a specific hook. This is called by the generated hook scripts:

```bash
stave --hooks run pre-commit -- "$@"
```

You can run this manually for debugging:

```bash
stave --hooks run pre-commit
```

## Environment Variables

Control hook behavior through environment variables:

| Variable            | Effect                                           |
| ------------------- | ------------------------------------------------ |
| `STAVE_HOOKS=0`     | Disable all hooks (exit silently with 0)         |
| `STAVE_HOOKS=debug` | Enable shell tracing (`set -x`) in hook scripts  |
| `STAVE_QUIET=1`     | Suppress decorative output (auto-detected in CI) |

### Disabling Hooks

To temporarily skip hooks:

```bash
STAVE_HOOKS=0 git commit -m "WIP"
```

Or set globally in your shell profile to disable hooks on a specific machine.

### Debugging Hooks

Enable verbose output:

```bash
STAVE_HOOKS=debug git commit -m "test"
```

### Quiet Mode

Decorative output (hook run messages, test headers) is automatically suppressed when running in CI environments. Stave detects CI via these environment variables:

- `CI`
- `GITHUB_ACTIONS`
- `GITLAB_CI`
- `JENKINS_URL`
- `CIRCLECI`
- `BUILDKITE`

To force quiet mode outside CI:

```bash
STAVE_QUIET=1 stave test
```

## User Init Script

Hook scripts source an optional user init script before running Stave. This is useful for:

- Initializing version managers (nvm, pyenv, etc.)
- Adjusting `PATH` for GUI Git clients
- Machine-specific environment setup

Location:

```text
${XDG_CONFIG_HOME:-$HOME/.config}/stave/hooks/init.sh
```

Example `init.sh`:

```sh
# Load nvm for Node.js version management
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"

# Ensure Go is on PATH
export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"
```

## Generated Hook Scripts

Stave generates POSIX-compatible shell scripts. Example `pre-commit` hook:

```sh
#!/bin/sh
# Installed by Stave: DO NOT EDIT BY HAND

# Optional user-level initialization (PATH, version managers, etc.)
init_script="${XDG_CONFIG_HOME:-$HOME/.config}/stave/hooks/init.sh"
[ -f "$init_script" ] && . "$init_script"

# Global toggle and debug controls
if [ "${STAVE_HOOKS-}" = "0" ]; then
  exit 0
fi
[ "${STAVE_HOOKS-}" = "debug" ] && set -x

if command -v stave >/dev/null 2>&1; then
  exec stave --hooks run pre-commit -- "$@"
else
  echo "stave: 'stave' binary not found on PATH; skipping pre-commit hook." >&2
  exit 0
fi
```

Key properties:

- POSIX `sh` compatible (no bash-specific features)
- Graceful fallback if `stave` is not on PATH
- Respects `STAVE_HOOKS` environment variable
- Clearly marked as Stave-managed

## Custom Hooks Path

If your repository uses `core.hooksPath`:

```bash
git config core.hooksPath .githooks
```

Stave respects this setting and installs hooks to the configured directory.

## Stdin Handling

Some hooks receive data via stdin:

| Hook                 | Stdin Contains                          |
| -------------------- | --------------------------------------- |
| `commit-msg`         | Path to commit message file             |
| `prepare-commit-msg` | Path to commit message file             |
| `pre-receive`        | List of refs being updated              |
| `post-receive`       | List of refs that were updated          |

Use `passStdin: true` to forward stdin to your target:

```yaml
hooks:
  commit-msg:
    - target: validate-commit-message
      passStdin: true
```

## Execution Behavior

### Sequential Execution

Targets for a hook run sequentially in the order defined:

```yaml
hooks:
  pre-commit:
    - target: fmt      # Runs first
    - target: lint     # Runs second
    - target: typecheck # Runs third
```

### Fail-Fast

Execution stops on the first failure. If `lint` fails, `typecheck` does not run:

```text
stave: hook pre-commit failed at target lint (exit 1)
```

### Exit Codes

- Exit `0`: Hook passes, Git operation proceeds
- Non-zero exit: Hook fails, Git operation is blocked

## Example: Complete Setup

### stavefile.go

```go
//go:build stave

package main

import (
    "github.com/yaklabco/stave/pkg/sh"
)

// Fmt formats Go code.
func Fmt() error {
    return sh.RunV("go", "fmt", "./...")
}

// Lint runs the linter.
func Lint() error {
    return sh.RunV("golangci-lint", "run")
}

// Test runs the test suite.
func Test() error {
    return sh.RunV("go", "test", "./...")
}
```

### stave.yaml

```yaml
hooks:
  pre-commit:
    - target: Fmt
    - target: Lint
  pre-push:
    - target: Test
```

### Installation

```bash
stave --hooks install
```

Now every commit runs `Fmt` and `Lint`, and every push runs `Test`.

## Switching Hook Systems

If your repository uses Husky or another hook manager that sets `core.hooksPath`, you must unset it before installing Stave hooks.

### Switch to Native Stave Hooks

```bash
# Unset any custom hooks path (e.g., from Husky)
git config --unset core.hooksPath

# Install Stave-managed hooks
stave --hooks install
```

### Switch Back to Husky

```bash
# Uninstall Stave hooks
stave --hooks uninstall

# Restore Husky's hooks path
git config core.hooksPath .husky
```

### Optional: Add a Convenience Target

For projects that need to switch between hook systems frequently, you can add a `Hooks` target to your `stavefile.go`:

```go
// SwitchHooks configures git hooks to use either "husky" or "stave" (native).
func SwitchHooks(system string) error {
    switch strings.ToLower(system) {
    case "husky":
        _ = sh.Run("stave", "--hooks", "uninstall")
        return sh.Run("git", "config", "core.hooksPath", ".husky")
    case "stave":
        _ = sh.Run("git", "config", "--unset", "core.hooksPath")
        return sh.Run("stave", "--hooks", "install")
    default:
        return fmt.Errorf("unknown hooks system %q: use 'husky' or 'stave'", system)
    }
}
```

This enables:

```bash
stave SwitchHooks stave   # Switch to native Stave hooks
stave SwitchHooks husky   # Switch to Husky
```

## Migration from Husky

To permanently migrate from Husky to Stave hooks:

1. Add `hooks` configuration to `stave.yaml` based on your `.husky/*` scripts
2. Unset the Husky hooks path: `git config --unset core.hooksPath`
3. Install Stave hooks: `stave --hooks install`
4. Remove Husky: `npm uninstall husky`
5. Delete the `.husky/` directory
6. Remove the `prepare` script from `package.json`

---

## See Also

- [Configuration](configuration.md) - Full configuration reference
- [CLI Reference](../api-reference/cli.md) - Command-line flags
- [Targets](targets.md) - Defining target functions
- [Home](../index.md)
