# Development Setup

This guide helps you set up a development environment for contributing to Stave.

## Prerequisites

- **Go 1.21+** - Stave uses modern Go features
- **Git** - For version control
- **Node.js** - For git hooks (husky)
- **Homebrew** (macOS/Linux) - For installing tools

## Clone the Repository

```bash
git clone https://github.com/yaklabco/stave.git
cd stave
```

## Bootstrap Stave

If you don't have Stave installed yet:

```bash
go run bootstrap.go
```

This compiles and installs Stave to your `$GOBIN`.

## Install Development Dependencies

Stave uses its own stavefile for development tasks. Run the init target:

```bash
stave init
```

This runs:
1. `brew bundle --file=Brewfile` - Installs tools (golangci-lint, goreleaser, etc.)
2. `npm ci` - Installs Node dependencies (husky for git hooks)
3. `git config core.hooksPath .husky` - Sets up git hooks
4. `go mod tidy` - Tidies Go modules
5. `go generate ./...` - Generates code (stringer, etc.)

### Manual Setup

If you prefer manual setup:

```bash
# Install Go tools
brew bundle --file=Brewfile

# Or install directly
brew install golangci-lint goreleaser svu

# Install Node dependencies (for git hooks)
npm ci

# Set up git hooks
git config core.hooksPath .husky
chmod +x .husky/pre-push

# Tidy modules
go mod tidy

# Generate code
go generate ./...
```

## Project Structure

```
stave/
  main.go              # CLI entry point
  bootstrap.go         # Self-bootstrap installer
  stavefile.go         # Build script (dogfooding!)
  
  cmd/stave/           # Cobra CLI
  config/              # Configuration management
  internal/            # Internal packages
  pkg/                 # Public API packages
    st/                # Core API for stavefiles
    sh/                # Shell command helpers  
    target/            # File target utilities
    stave/             # Runtime engine
  
  docs/                # Documentation
  rfcs/                # Design documents
  
  .husky/              # Git hooks
  Brewfile             # Homebrew dependencies
  .goreleaser.yaml     # Release configuration
```

## Running Tests

Run all tests:

```bash
stave test
```

Or run Go tests directly:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run a specific package's tests:

```bash
go test -v ./pkg/st/...
```

## Linting

Run linters:

```bash
stave lint
```

Or run golangci-lint directly:

```bash
golangci-lint run --fix
```

The linter configuration is in `.golangci.yml`.

## Building

Build artifacts with goreleaser:

```bash
stave build
```

Or build a simple binary:

```bash
go build -o stave .
```

## Code Generation

Some code is generated. After modifying certain files, run:

```bash
go generate ./...
```

Generated files:
- `pkg/st/color_string.go` - Stringer for Color type

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/my-feature
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation
- `refactor/` - Code refactoring

### 2. Make Changes

Edit the code. Key files:
- User-facing API: `pkg/st/`, `pkg/sh/`, `pkg/target/`
- Runtime: `pkg/stave/`
- CLI: `cmd/stave/`
- Parsing: `internal/parse/`
- Configuration: `config/`

### 3. Test Your Changes

```bash
# Run tests
stave test

# Or just Go tests without lint
go test ./...
```

### 4. Lint Your Code

```bash
stave lint
```

Fix any issues before committing.

### 5. Commit

Commits should follow conventional commit format:

```bash
git commit -m "feat: add new feature X"
git commit -m "fix: resolve issue with Y"
git commit -m "docs: update README"
```

The pre-push hook will run tests automatically.

### 6. Push and Create PR

```bash
git push -u origin feature/my-feature
```

Then create a pull request on GitHub.

## IDE Setup

### VS Code

Recommended extensions:
- Go (official)
- golangci-lint

Settings (`.vscode/settings.json`):
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.buildTags": "stave"
}
```

### GoLand / IntelliJ

1. Go to Preferences > Go > Build Tags
2. Add `stave` to Custom tags

## Debugging

### Debug Stave Itself

```bash
# Run with debug output
stave -d <target>

# Keep generated mainfile for inspection
stave --keep <target>
```

### Debug Tests

```bash
# Run single test with verbose output
go test -v -run TestMyFunction ./pkg/st/
```

### Inspect Generated Code

The `--keep` flag preserves the generated mainfile:

```bash
stave --keep build
cat stave_output_file.go
```

## Common Tasks

### Add a New CLI Flag

1. Add to `RunParams` struct in `pkg/stave/main.go`
2. Register with Cobra in `cmd/stave/stave.go`
3. Handle in the `Run()` function
4. Add tests

### Add a New st.* Function

1. Add function to appropriate file in `pkg/st/`
2. Add tests in `pkg/st/*_test.go`
3. Update documentation

### Modify the Mainfile Template

1. Edit `pkg/stave/templates/mainfile_tmpl.go`
2. Test with `--keep` to verify output
3. Run full test suite

## Troubleshooting

### "stave: command not found"

Ensure `$GOBIN` or `$GOPATH/bin` is in your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Tests Fail with Stale Binary

Force rebuild:

```bash
stave --clean
stave -f test
```

### Lint Errors After Pull

Regenerate code:

```bash
go generate ./...
go mod tidy
```

## Next Steps

- Read the [Architecture](architecture.md) document
- Review [Code Style](code-style.md) guidelines
- Check [Testing](testing.md) for test conventions
- See [Pull Requests](pull-requests.md) for contribution workflow

