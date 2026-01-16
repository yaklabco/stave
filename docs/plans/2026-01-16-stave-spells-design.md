# Stave Spells Design

A scaffolding and code generation system for stave.

## Overview & Core Concepts

**Stave Spells** is a scaffolding and code generation system for stave. It lets developers cast "spells" - pre-built templates that generate CI pipelines, linting configs, release workflows, and other project infrastructure with battle-tested defaults.

### Core Metaphor

- **Spell** - A reusable template that generates files with customizable options
- **Spellbook** - A collection of related spells, distributed as a git repository
- **Cast** - The act of executing a spell to generate files in your project
- **Overlay** - An optional layer that adds or modifies spell output (e.g., `--with linting`)

### Design Principles

1. **Explicit over magic** - Users specify exactly which spell to cast (`ci-go` not auto-detected)
2. **Safe by default** - Dry-run preview before any changes, transactional application
3. **Upgradeable** - Track what was cast, detect manual edits, three-way merge on updates
4. **Composable** - Spells have overlays, users can customize via flags/prompts/config
5. **Lean core** - Stave ships with setup wizard, spellbooks are imported on demand

### CLI Entry Points

```
stave cast <spell>              # Execute a spell
stave spells <verb>             # Manage spells and spellbooks
```

The two-command structure keeps `cast` as the quick action while `spells` provides the full management interface.

---

## CLI Interface

### `stave cast`

The primary action command. Executes a spell with optional customization.

```bash
stave cast <spell>                      # Interactive prompts for options
stave cast <spell> --with a,b           # Include overlays
stave cast <spell> --without c          # Exclude default overlays
stave cast <spell> --yes                # Skip confirmation, use defaults
stave cast <spell> --dry-run            # Preview only, don't prompt
```

When cast is invoked:

1. Show what would be created/modified
2. Prompt for confirmation: `[y/N/diff]`
3. Stage all changes in memory
4. Apply atomically if staging succeeds

### `stave spells`

Management interface with these verbs:

| Verb | Purpose |
|------|---------|
| `find <query>` | Search registry for spells matching query |
| `show <spell>` | Display spell details, variables, overlays |
| `list` | Show installed spellbooks and available spells |
| `import <source>` | Add a spellbook from git URL |
| `history` | Show spells cast in current project |
| `upgrade [spell]` | Upgrade cast spell(s) to newer versions |
| `compose` | Interactive wizard to author a new spell |
| `setup` | First-run onboarding, import recommended spellbooks |

### Spell Resolution

Short names work when unambiguous. On conflict:

```
$ stave cast ci-go
Multiple spells match 'ci-go':
  go-spells/ci-go
  internal/ci-go
Use fully qualified name or set an alias.
```

Aliasing: `stave spells import go-spells --alias go` enables `stave cast go:ci-go`

---

## Spell Definition Format

A spell is a directory containing a manifest and templates.

### Directory Structure

```
ci-go/
  spell.yml              # Manifest: metadata, variables, templates, hooks
  templates/
    ci.yml.tmpl          # Go text/template files
  overlays/
    linting/
      golangci.yml.tmpl
      ci.yml.patch       # Patch to modify base template
    coverage/
      ci.yml.patch
  hooks/
    hooks.go             # Optional Go code for validation/logic
```

### spell.yml

```yaml
name: ci-go
version: 1.2.0
description: GitHub Actions CI for Go projects
author: yaklabco
license: Apache-2.0

variables:
  go_version:
    default: "1.24"
    prompt: "Go version?"
  module_name:
    source: go.mod        # Auto-extract from project

templates:
  - src: templates/ci.yml.tmpl
    dest: .github/workflows/ci.yml

overlays:
  linting:
    default: true         # Included unless --without
    templates:
      - src: overlays/linting/golangci.yml.tmpl
        dest: .golangci.yml
    patches:
      - overlays/linting/ci.yml.patch

hooks:
  validate: Validate      # Go function, called before cast
  post_cast: PostCast     # Go function, called after success
```

### Go Hooks

Optional hooks for complex logic. Standard Go file with build tag:

```go
//go:build stave_spell

package hooks

func Validate(ctx SpellContext) error {
    // Check prerequisites, return error to abort
}

func PostCast(ctx SpellContext) error {
    // Run post-generation tasks
}
```

---

## Spellbook Structure & Distribution

### Spellbook Layout

A spellbook is a git repository containing multiple spells:

```
go-spells/
  spellbook.yml          # Spellbook manifest
  ci-go/
    spell.yml
    templates/
  linting-go/
    spell.yml
    templates/
  testing-go/
    spell.yml
    templates/
    hooks/
  release-go/
    spell.yml
    templates/
    hooks/
```

### spellbook.yml

```yaml
name: go-spells
author: yaklabco
license: Apache-2.0
description: Production-ready spells for Go projects
repository: github.com/yaklabco/go-spells

spells:
  ci-go: GitHub Actions CI for Go projects
  linting-go: golangci-lint with opinionated defaults
  testing-go: Test infrastructure with coverage reporting
  release-go: GoReleaser with changelog automation
```

### Importing Spellbooks

```bash
stave spells import github.com/yaklabco/go-spells
stave spells import github.com/mycompany/internal-spells --alias internal
stave spells import ./local-spells                      # Local path
```

Spellbooks are cloned/cached locally. Stave tracks:

- Source URL and current commit
- Alias (if set)
- Last fetched timestamp

Multiple spellbooks coexist. Each project can import different sets. User-level imports in `~/.stave/spellbooks/` are available globally.

---

## Registry & Discovery

### Architecture

A hybrid approach:

- **Central index** - Git repository containing searchable catalog of known spellbooks
- **Source repos** - Spells fetched directly from their home repositories

The index doesn't host spells, just metadata for discovery.

### Index Structure

```
stave-registry/
  index.yml
  official/
    go-spells.yml
    node-spells.yml
  community/
    rust-spells.yml
    gitlab-spells.yml
```

Each entry:

```yaml
# official/go-spells.yml
name: go-spells
repository: github.com/yaklabco/go-spells
description: Production-ready spells for Go projects
keywords: [go, golang, ci, github-actions, linting]
tier: official
spells:
  - ci-go
  - linting-go
  - testing-go
  - release-go
```

### Tiered Display

```
$ stave spells find ci

✓ ci-go          GitHub Actions CI for Go          go-spells
✓ ci-node        GitHub Actions CI for Node.js     node-spells
☆ ci-rust        GitHub Actions CI for Rust        community/rust-spells
☆ ci-go-gitlab   GitLab CI for Go                  community/gitlab-spells

✓ = official    ☆ = community
```

### Adding to Registry

- **Official tier** - Curated by yaklabco, quality reviewed
- **Community tier** - Submit PR to registry repo, basic validation (valid spellbook.yml, spells parse correctly)

---

## Local State & Storage

### State Directory

Stave maintains spell state in `.stave/` within each project:

```
.stave/
  spells.db              # SQLite database
  preferences.yml        # User's default choices for this project
  cache/                 # Cached spellbook data
```

### Database Schema

The SQLite database tracks:

**cast_history** - Record of every cast

```
id, spell_name, spellbook, version, cast_at, options_json
```

**cast_files** - Files generated by each cast

```
cast_id, file_path, content_hash, is_modified
```

**spellbooks** - Imported spellbooks

```
name, alias, source_url, commit_sha, fetched_at
```

### Content Hash Tracking

When a spell is cast, stave records SHA-256 hashes of generated content. On upgrade:

1. Compute current file hash
2. Compare to recorded hash
3. If different -> user made manual edits -> trigger merge flow

### preferences.yml

Remembers choices for repeat casts:

```yaml
ci-go:
  go_version: "1.24"
  overlays:
    - linting
    - coverage
```

Next time: `stave cast ci-go` uses these defaults, prompts only for new variables.

---

## Upgrades & Merge Flow

### Detecting Available Upgrades

```bash
$ stave spells upgrade --check

Upgrades available:
  ci-go      1.2.0 → 1.3.0   (go-spells)
  linting-go 2.0.0 → 2.1.0   (go-spells)
```

Stave compares the recorded cast version against the current spellbook version.

### Upgrade Commands

```bash
stave spells upgrade ci-go        # Upgrade specific spell
stave spells upgrade --all        # Upgrade all cast spells
stave cast ci-go                  # Also detects and offers upgrade
```

### Three-Way Merge

When upgrading files with manual edits:

```
Original (v1.2.0)  →  User's version  →  New (v1.3.0)
        ↓                   ↓                  ↓
        └───────── Three-way merge ───────────┘
```

Stave uses the original generated content (reconstructed from spell + recorded options) as the merge base.

### Conflict Handling

```
$ stave spells upgrade ci-go

Upgrading ci-go 1.2.0 → 1.3.0...

.github/workflows/ci.yml: merged (no conflicts)
.golangci.yml: CONFLICT

<<<<<<< yours
  timeout: 10m
=======
  timeout: 5m
>>>>>>> spell v1.3.0

Resolve conflicts and run 'stave spells upgrade --continue'
```

Standard git-style conflict markers. User resolves, then continues.

All output is colorized using stave's charmbracelet-based styling.

---

## Transactional Casting & Error Handling

### Staging Phase

When `stave cast ci-go` runs:

1. **Resolve** - Load spell, resolve variables (from prompts, config, or flags)
2. **Render** - Process all templates and patches in memory
3. **Validate** - Run `validate` hook if defined
4. **Stage** - Build complete changeset without touching filesystem
5. **Preview** - Display staged changes with colors:
   - Green: `+ new file`
   - Yellow: `~ modified file`
   - Cyan: `→ would add to .gitignore`

### Atomic Apply

Only after user confirms:

1. Create backup of any files being modified
2. Write all new/modified files
3. Run `post_cast` hook
4. Record cast in database
5. Clean up backups

If any step fails -> full rollback to original state.

### Error Output

```
$ stave cast ci-go

🔮 Spell: ci-go v1.2.0
   GitHub Actions CI for Go projects

Would create:
  + .github/workflows/ci.yml
  + .golangci.yml

Would modify:
  ~ .gitignore

Proceed? [y/N/diff] y

✗ Error: post_cast hook failed
  → golangci-lint not found in PATH

Rolled back all changes.
Hint: Install golangci-lint and try again.
```

All output uses stave's existing color/styling via charmbracelet libraries.

---

## Spell Authoring with Compose

### The Compose Wizard

`stave spells compose` guides users through creating a new spell interactively.

```
$ stave spells compose

🔮 Spell Composer

Spell name: ci-python
Description: GitHub Actions CI for Python projects

Variables (enter blank line when done):
  Name: python_version
  Default: 3.12
  Prompt: Python version?

  Name:

Templates to generate:
  Source path: templates/ci.yml.tmpl
  Destination: .github/workflows/ci.yml

  Source path:

Overlays (optional):
  Overlay name: linting
  Include by default? [Y/n] y

  Overlay name:

Add Go hooks? [y/N] n

Save to:
  (1) This project (.stave/spellbook/)
  (2) User spellbook (~/.stave/spellbooks/local/)
  (3) Custom path

Choice: 3
Path: ~/projects/python-spells/ci-python

✓ Created spell at ~/projects/python-spells/ci-python

  spell.yml
  templates/ci.yml.tmpl

Next: Edit templates, then 'stave cast ci-python' to test.
```

### Template Scaffolding

Compose generates starter templates with variable placeholders:

```yaml
# templates/ci.yml.tmpl (generated)
name: CI
on: [push, pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '{{ .python_version }}'
      # TODO: Add your build steps
```

---

## Onboarding Experience

### First Run

Stave ships with no built-in spells. The `setup` command guides new users:

```
$ stave spells list

No spellbooks installed.
Run 'stave spells setup' to get started.

$ stave spells setup

🔮 Welcome to Stave Spells!

What languages/platforms do you work with?
  [x] Go
  [ ] Node.js / TypeScript
  [ ] Python
  [ ] Rust
  [ ] Other

What CI platform do you use?
  [x] GitHub Actions
  [ ] GitLab CI
  [ ] CircleCI
  [ ] Other

Recommended spellbooks:

  ✓ yaklabco/go-spells
    CI, linting, testing, and release automation for Go

Import now? [Y/n] y

Importing yaklabco/go-spells... done

✓ Setup complete!

Available spells:
  ci-go        GitHub Actions CI for Go projects
  linting-go   golangci-lint with opinionated defaults
  testing-go   Test infrastructure with coverage
  release-go   GoReleaser with changelog automation

Try: stave cast ci-go
```

### Subsequent Projects

Spellbooks imported at user level (`~/.stave/spellbooks/`) are available everywhere. Setup only needed once per machine, or when adding new spellbooks.

```
$ cd new-project
$ stave cast ci-go    # Works immediately
```

---

## Implementation Considerations

### Package Structure

```
pkg/spells/
  cast.go           # Cast execution, staging, apply
  spellbook.go      # Spellbook loading, management
  registry.go       # Registry fetching, search
  state.go          # SQLite state management
  merge.go          # Three-way merge logic
  compose.go        # Authoring wizard

cmd/stave/
  spells_cmd.go     # CLI commands for 'stave spells'
  cast_cmd.go       # CLI command for 'stave cast'
```

### Dependencies

- **SQLite** - `modernc.org/sqlite` (pure Go, no CGO)
- **Git operations** - `go-git` or shell out to git
- **Templates** - Go's `text/template` with custom functions
- **Diffing/merging** - `github.com/hexops/gotextdiff` or similar
- **UI** - Existing charmbracelet stack (lipgloss, huh, etc.)

### Testing Strategy

- Unit tests for template rendering, merge logic, state management
- Integration tests with temporary git repos and spellbooks
- Testdata spellbooks with known inputs/outputs
- Golden file tests for generated content

### Rollout Phases

1. **Core** - `cast`, `spells list/show/import`, local state tracking
2. **Upgrades** - Version detection, three-way merge, conflict resolution
3. **Registry** - Central index, `find`, tiered discovery
4. **Compose** - Authoring wizard
5. **Official spellbooks** - `go-spells` with ci-go, linting-go, etc.
