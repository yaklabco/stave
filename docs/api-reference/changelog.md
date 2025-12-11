# Changelog Utilities (pkg/changelog)

[Home](../index.md) > [API Reference](changelog.md) > Changelog Utilities

Stave provides utilities for working with project changelogs and versions. These helpers are designed to enforce keep-a-changelog formatting, ensure your CHANGELOG is updated on every push (if you choose), and to compute the next version/tag from Conventional Commits using svu.

Package import path:

```go
import "github.com/yaklabco/stave/pkg/changelog"
```

## Features

- Validate CHANGELOG.md formatting according to keep-a-changelog
- Enforce that a push includes CHANGELOG updates (configurable)
- Compute next version (no leading v) and next tag (with v) from Conventional Commits using svu
- Utilities to integrate these checks into Git hooks

## Validation and Pre-Push Checks

Use `PrePushCheck` to validate that your CHANGELOG.md is well-formed and (optionally) that the pushed changes include a modification to the CHANGELOG. This is intended to be run from a `pre-push` hook target.

```go
//go:build stave

package main

import (
    "fmt"
    "os"

    "github.com/yaklabco/stave/pkg/changelog"
)

// CheckPrePush validates CHANGELOG and ensures it was updated in this push.
func CheckPrePush() error {
    // Read refs from stdin as provided by Git's pre-push hook
    refs, err := changelog.ReadPushRefs(os.Stdin)
    if err != nil {
        return fmt.Errorf("read push refs: %w", err)
    }

    res, err := changelog.PrePushCheck(changelog.PrePushCheckOptions{
        RemoteName:    os.Getenv("GIT_REMOTE"), // optional; can be empty
        ChangelogPath: "CHANGELOG.md",          // default if empty
        Refs:          refs,
        // SkipNextVersionCheck: true, // uncomment to skip verifying next version exists in CHANGELOG
    })
    if err != nil {
        return err
    }
    if res.HasErrors() {
        return res.Error()
    }
    return nil
}
```

Notes:

- `PrePushCheck` parses your CHANGELOG, validates formatting, and verifies per-ref whether the diff includes `CHANGELOG.md`.
- You can point `ChangelogPath` at a different file if needed.
- Set `SkipNextVersionCheck: true` to skip verifying that the next version exists in the CHANGELOG.
- Set environment variable `BYPASS_CHANGELOG_CHECK=1` to bypass all checks (useful for emergency pushes).
- Set `STAVEFILE_SKIP_NEXTVER_CHANGELOG_CHECK=1` to bypass only the next-version presence check (e.g., during a release when a tag-only push occurs).

### Simple File Validation

Validate formatting offline (useful in CI):

```go
if err := changelog.ValidateFile("CHANGELOG.md"); err != nil {
    return err
}
```

### Verify Next Version Exists in Unreleased Section

Ensure your `Unreleased` section contains an entry for the next version number:

```go
if err := changelog.VerifyNextVersion("CHANGELOG.md"); err != nil {
    return err
}
```

## Auto Version/Tag from Conventional Commits (svu)

Stave bundles the `svu` library and exposes helpers to compute the next version/tag based on Conventional Commits in your repo.

```go
v, err := changelog.NextVersion() // e.g., "0.4.0" (no leading 'v')
tag, err := changelog.NextTag()   // e.g., "v0.4.0"
```

These functions:

- Inspect commits per `svu.Next(svu.Always())`
- Return trimmed values (no newline), and `NextVersion` strips the leading `v` for CHANGELOG headings

Typical uses:

- Generate release notes for the upcoming version
- Create annotated tags
- Verify your CHANGELOG’s next version matches the computed value

## Wiring into Git Hooks

Combine with Stave’s Git hooks to require CHANGELOG updates on every push:

```yaml
# stave.yaml
hooks:
  pre-push:
    - target: CheckPrePush
      passStdin: true
```

Then implement `CheckPrePush` as shown above. The `passStdin: true` ensures the Git hook’s stdin (refs) is available to your target.

## See Also

- [Git Hooks](../user-guide/hooks.md)
- [Versioning](../user-guide/versioning.md)
- [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
