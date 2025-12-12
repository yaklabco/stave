# Stave Features Beyond Mage

This document tracks **Stave functionality that does not exist in upstream Mage**, focusing on substantive behavioral differences rather than implementation details or UI polish.

The baseline for comparison is the Mage history up to the fork point:

- **Fork commit**: `6189be7 Fork mage as staff`

All features listed below are introduced in commits after that fork in `yaklabco/stave`.

---

## Stave-Only Features

### Dry-run support

- **Summary**: Simulate command execution while still compiling and running the stavefile.
- **Commits**: `2bae053 feat: add in dryrun functionality` (+ follow-on wiring such as `86c2262 chore: add context arg to dryrun.Wrap(...)`).
- **Behavior**:
  - CLI flag: `--dryrun`.
  - Environment variables:
    - `STAVEFILE_DRYRUN_POSSIBLE` – set by the outer Stave invocation when dry-run is supported in the current context.
    - `STAVEFILE_DRYRUN` – set when dry-run mode is requested (via CLI or env).
  - `internal/dryrun` exposes:
    - `IsPossible()` – checks `STAVEFILE_DRYRUN_POSSIBLE`.
    - `IsRequested()` / `SetRequested(...)` – track a requested dry-run.
    - `IsDryRun()` – true only when both _possible_ and _requested_.
    - `Wrap(ctx, cmd, args...)` – returns either a real `exec.Cmd` or an `echo "DRYRUN: ..."` command.
  - All `sh.Run*` helpers honor `dryrun.IsDryRun()` and print `DRYRUN: ...` lines instead of executing commands when in dry-run mode.
- **Why this is new vs Mage**:
  - Mage currently has **no first-class dry-run mode** in core.
  - Stave’s dry-run is an end-to-end feature:
    - Opt-in via flag / env.
    - Plumbed through the compiled stavefile and `pkg/sh`.
    - Designed to be safe for nested Stave invocations (top-level compilation still runs even when `--dryrun` is set).

---

### Additional Shell Helpers: sh.Piper and sh.PiperWith

- Summary: Stream-oriented variants of sh.Run/sh.RunWith that make it easy to wire stdin/stdout/stderr for subprocesses.
- Commits: Introduced as part of the post-fork shell utilities work in pkg/sh (see pkg/sh/cmd.go).
- Behavior:
  - sh.Piper(stdin, stdout, stderr, cmd, args...) error
    - Runs cmd with provided readers/writers, no env injection.
  - sh.PiperWith(env, stdin, stdout, stderr, cmd, args...) error
    - Like Piper, but merges an additional env map into the child process environment.
  - Both helpers honor dry-run mode via internal/dryrun, printing the simulated command.
- Why this is new vs Mage:
  - Mage provides sh.Run/sh.RunWith but lacks first-class helpers focused on explicit piping of stdio streams.

---

### XDG-Compliant Configuration System and `stave config` Subcommands

- **Summary**: A structured, XDG-aware configuration system with a dedicated `stave config` command surface.
- **Commits**: `5222738 feat(config): add XDG-compliant configuration system` (plus supporting work in `config_cmd.go` and docs).
- **Behavior**:
  - **Configuration sources (in precedence order)**:
    1. Built-in defaults.
    2. **User config**: `~/.config/stave/config.yaml` (or platform/XDG equivalent).
    3. **Project config**: `./stave.yaml` in the current project.
    4. **Environment overrides**: `STAVEFILE_*` variables.
  - **Key options** (see `config.Config` and docs for full list):
    - `cache_dir` → where to cache compiled binaries (default is XDG cache directory).
    - `go_cmd` → which `go` binary to use.
    - `verbose`, `debug`, `hash_fast`, `ignore_default`, `enable_color`, `target_color`, etc.
  - **CLI subcommands** (`stave config`):
    - `stave config` / `stave config show`:
      - Prints the _effective_ configuration, including the file it was loaded from (if any).
    - `stave config init`:
      - Writes a default `~/.config/stave/config.yaml`.
    - `stave config path`:
      - Shows resolved config, cache, and data directories, plus which config file is currently active (if any).
- **Why this is new vs Mage**:
  - Mage does **not** provide:
    - A config-file abstraction.
    - XDG-aware config/cache/data resolution.
    - A `mage config` command surface.
  - Mage behavior is driven purely by flags and environment variables, while Stave adds a richer, layered configuration model.

---

### Automatic Detection of Circular Dependencies in Targets

- Summary: Detect and stop execution when a cycle is found in target dependencies.
- Commits/Code: pkg/toposort/toposort.go (ErrCircularDependency), exercised by pkg/stave/cyclic_dependencies_test.go and testdata/cyclic_dependencies.
- Behavior:
  - Dependency resolution builds a graph of target prerequisites.
  - If a cycle exists, Stave fails fast with the error message "circular dependency detected" and does not run any targets in the cycle.
- Why this is new vs Mage:
  - Mage does not surface a standardized, user-visible cycle detection error in the same manner; Stave bakes this into its target resolution to prevent confusing partial runs.

---

### First-Class Parallelism Control via `STAVE_NUM_PROCESSORS`

- **Summary**: Centralized, tool-level control over parallelism that informs both `GOMAXPROCS` and downstream tools.
- **Commits**: Implemented post-fork as part of parallelism work (see `internal/parallelism/parallelism.go` and `TODO.md` entry “Add STAVE_NUM_PROCESSORS / GOMAXPROCS setter”).
- **Behavior**:
  - **Environment variable**:
    - `STAVE_NUM_PROCESSORS`:
      - If set, Stave uses this value as the canonical “number of processors”.
      - If unset, defaults to `runtime.NumCPU()`.
  - **Runtime effects**:
    - `internal/parallelism.Apply`:
      - Sets `runtime.GOMAXPROCS(numProcessors)`.
      - Writes both `STAVE_NUM_PROCESSORS` and `GOMAXPROCS` into the environment map passed to the compiled stavefile.
    - The repo’s own `stavefile.go` uses `STAVE_NUM_PROCESSORS` to set:
      - `-p` / `-parallel` for `go test` via `gotestsum`.
      - `--parallelism` for `goreleaser`.
  - This produces a single knob that controls:
    - CPU-level parallelism (`GOMAXPROCS`).
    - Test runner and release tooling concurrency.
- **Why this is new vs Mage**:
  - Mage does not expose a dedicated, tool-level “number of processors” setting; parallelism is determined by defaults or hand-written project code.
  - Stave standardizes this via `STAVE_NUM_PROCESSORS` and the `parallelism.Apply` helper.

---

### XDG-Aware Cache and Data Directory Handling

- **Summary**: Cross-platform, XDG-aligned resolution of config, cache, and data directories, integrated into Stave’s runtime.
- **Commits**: Part of the same configuration work as `5222738 feat(config): add XDG-compliant configuration system`.
- **Behavior**:
  - `config.XDGPaths` resolves:
    - `ConfigHome`, `CacheHome`, `DataHome` in an XDG-compliant, OS-sensitive way:
      - Honors `XDG_CONFIG_HOME`, `XDG_CACHE_HOME`, and `XDG_DATA_HOME` when set.
      - Falls back to platform defaults:
        - Linux: `~/.config`, `~/.cache`, `~/.local/share`.
        - macOS: `~/.config`, `~/Library/Caches`, `~/Library/Application Support`.
        - Windows: respects `APPDATA` / `LOCALAPPDATA` where appropriate.
  - Application-specific paths are then derived as:
    - `~/.config/stave`, `~/.cache/stave`, and the equivalent for data.
  - The cache directory used for compiled binaries:
    - Defaults to the XDG cache path.
    - Can be overridden via config (`cache_dir`) or `STAVEFILE_CACHE`.
- **Why this is new vs Mage**:
  - Mage’s cache behavior is comparatively ad hoc and not XDG-aware.
  - Stave explicitly aligns with contemporary CLI expectations around config/cache/data locations.

---

### Native Git Hooks Management

- Summary: Built-in management of Git hooks that can run Stave targets without external tools (no husky/pre-commit required).
- Commits/Code: pkg/stave/hooks_cmd.go, internal/hooks/*, config/hooks.go; end-user docs at docs/user-guide/hooks.md.
- Behavior:
  - Declarative configuration in stave.yaml under hooks: mapping hook name → list of targets/args.
  - CLI surface: `stave --hooks` (list), `stave --hooks install`, `stave --hooks uninstall`, `stave --hooks doctor`.
  - POSIX-compatible wrapper scripts installed into .git/hooks calling back into Stave.
  - Honors CI detection and can be disabled via env.
- Why this is new vs Mage:
  - Mage does not ship built-in Git hooks management; typical workflows require third-party tooling. Stave integrates this capability natively.

---

### Changelog Toolkit (Keep a Changelog + Conventional Commits Versioning)

- Summary: Public Go helpers for enforcing CHANGELOG format and for computing next versions/tags from commit history.
- Commits/Code: pkg/changelog/* (check.go, validate.go, next.go, git.go, etc.).
- Behavior:
  - Validation helpers enforce Keep a Changelog semantics and can be used in CI targets to require a CHANGELOG update.
  - Versioning helpers compute the next semantic version and build tag from Conventional Commits, leveraging svu (bundled as a module dependency).
  - The two aspects (format enforcement and version/tag generation) can be adopted independently.
- Why this is new vs Mage:
  - Mage does not include a standard changelog toolkit; Stave packages reusable, tested helpers to encourage consistent release practices.

---

## Implementation-Focused Changes (Not Counted as “New Functionality”)

The following changes are important for maintainability and UX but are **not** counted as “new functionality” in the sense of capabilities that Mage did not have:

- Modernized Go patterns (Go 1.21+):
  - Effect: adoption of modern stdlib features and idioms; improves maintainability and performance without introducing new end-user capabilities relative to Mage.

- Enhanced CLI experience / Cobra-style CLI surface:

  - **Commits**: `ce22920 feat: cobra-ify!`, `804ad4d chore: mimic spf13/cobra in mainfile_tmpl.go without adding 3rd-party dependency`, and related.
  - Effect: more structured flag parsing and help output, but conceptually similar commands and options to Mage.

- Logging revamp with `slog` and charmbracelet-style logging:
  - **Commit**: `0700a0f feat: logging revamp`.
  - Effect: higher-quality structured logs and nicer terminal presentation, without introducing fundamentally new user capabilities vs Mage.

These are kept separate here to match the “bona fide new features, not just nicer implementation” criterion.

---

## Future Candidates to Add Here

The following items have been discussed as differentiators but are **not yet** fully implemented or surfaced in the current tree. Once they land, they should be added to the “Confirmed Stave-Only Features” section above with commit references:

- (none at the moment)

This section is intended as a staging area for upcoming work so that the Stave-vs-Mage feature delta remains easy to track over time.
