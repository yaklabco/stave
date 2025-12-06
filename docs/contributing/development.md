# Development Setup

[Home](../index.md) > [Contributing](development.md) > Development

This guide covers setting up a development environment for contributing to Stave.

## Prerequisites

- Go 1.25.4 or later
- macOS: Homebrew (for tool installation)

## Clone Repository

```bash
git clone https://github.com/yaklabco/stave.git
cd stave
```

## Bootstrap

Build and install Stave from source:

```bash
go run bootstrap.go
```

This compiles Stave with version metadata and installs it to `$GOBIN`.

## Install Development Tools

On macOS, install dependencies from the Brewfile:

```bash
brew bundle --file=Brewfile
```

This installs:

- `golangci-lint` - Go linter
- `gotestsum` - Test runner
- `goreleaser` - Release automation
- `svu` - Semantic version utility
- `markdownlint-cli2` - Markdown linter

## Setup Git Hooks

Stave uses its own native hook system.

### Fresh Clone

For new clones, install hooks:

```bash
stave --hooks install
```

### Switching from Husky

If your local clone still has Husky configured (via `core.hooksPath`), unset it first:

```bash
git config --unset core.hooksPath
stave --hooks install
```

This installs Git hooks defined in `stave.yaml`. The `pre-push` hook runs `Test` to ensure code quality before pushing.

To view configured hooks:

```bash
stave --hooks list
```

See [Git Hooks](../user-guide/hooks.md) for more details on hook configuration.

## Common Tasks

All development tasks are defined in `stavefile.go`.

### Run Tests

```bash
stave test
```

Runs linting and Go tests with coverage. Produces `coverage.out` and `coverage.html`.

### Run Linter

```bash
stave lint
```

Runs `golangci-lint` with auto-fix enabled and `markdownlint-cli2` on Markdown files.

### Build

```bash
stave build
```

Runs goreleaser in snapshot mode to produce binaries in `dist/`.

### Install Locally

```bash
stave install
```

Builds and installs Stave to `$GOBIN` with version flags.

### Clean

```bash
stave clean
```

Removes the `dist/` directory.

### All (Default)

```bash
stave
```

Runs `Init`, `Test`, then `Build`.

## Project Layout

```text
stave/
├── main.go                 # Binary entrypoint
├── bootstrap.go            # Bootstrap installer
├── stavefile.go            # Build targets
├── cmd/stave/              # CLI command definitions
│   ├── stave.go            # Root command (cobra)
│   └── version/            # Version handling
├── pkg/
│   ├── stave/              # Core runtime (Run, Compile, etc.)
│   ├── st/                 # User API (Deps, Fatal, etc.)
│   ├── sh/                 # Shell command execution
│   ├── target/             # File modification utilities
│   └── ui/                 # Terminal styling
├── config/                 # Configuration system
├── internal/
│   ├── parse/              # Stavefile AST parsing
│   ├── dryrun/             # Dry-run mode
│   ├── env/                # Environment utilities
│   ├── hooks/              # Git hook runtime and script generation
│   ├── parallelism/        # GOMAXPROCS handling
│   └── log/                # Logging constants
└── docs/                   # Documentation
```

## Running Tests

### All Tests

```bash
go test ./...
```

Or with gotestsum:

```bash
go tool gotestsum ./...
```

### Specific Package

```bash
go test ./pkg/st/...
```

### With Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Code Style

- Follow standard Go conventions
- Run `golangci-lint run --fix` before committing
- Use `slog` for structured logging
- Error messages should be lowercase without trailing punctuation
- Wrap errors with context: `fmt.Errorf("parsing file: %w", err)`

### Boolean configuration and environment parsing

- Use the helpers in `internal/env` (`ParseBool`, `ParseBoolEnv`, `ParseBoolEnvDefaultFalse`, `ParseBoolEnvDefaultTrue`) instead of implementing custom boolean parsing.
- Boolean values should accept `true`, `yes`, and `1` as true, and `false`, `no`, and `0` as false (case-insensitive, whitespace-insensitive).
- Prefer `ParseBoolEnvDefaultFalse` for opt-in features (invalid or unset values resolve to `false`) and `ParseBoolEnvDefaultTrue` for opt-out defaults (invalid or unset values resolve to `true`).

## Commit Messages

Use conventional commit format:

```text
type(scope): description

[optional body]

[optional footer]
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `chore`: Maintenance
- `refactor`: Code restructuring
- `test`: Test additions/changes

Examples:

- `feat(config): add XDG-compliant configuration system`
- `fix(parse): handle empty function bodies`
- `docs: update installation instructions`

## Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make changes and commit
4. Push to your fork
5. Open a pull request against `main`

Ensure:

- Tests pass: `stave test`
- Linter passes: `stave lint`
- Commit messages follow conventions
- PR description explains the change

---

## See Also

- [Architecture](architecture.md) - Codebase structure
- [Home](../index.md)
