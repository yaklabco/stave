# Versioning from Conventional Commits

[Home](../index.md) > [User Guide](stavefiles.md) > Versioning

Stave can compute the next semantic version and tag directly from your Conventional Commits using built-in helpers that wrap the svu library. This lets you automate release versions and verify your CHANGELOG aligns with what will be released.

## Quick Start

```go
//go:build stave

package main

import (
    "fmt"
    "github.com/yaklabco/stave/pkg/changelog"
)

// PrintNext shows the next version and tag derived from commit history.
func PrintNext() error {
    v, err := changelog.NextVersion() // e.g., "0.4.0"
    if err != nil {
        return err
    }
    tag, err := changelog.NextTag() // e.g., "v0.4.0"
    if err != nil {
        return err
    }
    fmt.Printf("next version: %s (tag %s)\n", v, tag)
    return nil
}
```

## Typical Workflows

- Generate a release tag after CI passes:

```go
func TagRelease() error {
    tag, err := changelog.NextTag()
    if err != nil {
        return err
    }
    return sh.Run("git", "tag", "-a", tag, "-m", "release "+tag)
}
```

- Verify your CHANGELOG has a heading and link for the upcoming version:

```go
func CheckChangelog() error {
    return changelog.VerifyNextVersion("CHANGELOG.md")
}
```

## How It Works

- `NextVersion` and `NextTag` use svu to infer the next semantic version from your commit messages following Conventional Commits.
- `NextVersion` returns the version without a leading `v` to match CHANGELOG headings; `NextTag` includes the `v` prefix for Git tags.

## See Also

- [Changelog Utilities](../api-reference/changelog.md)
- [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
- [CLI Reference](../api-reference/cli.md)
- [Home](../index.md)
