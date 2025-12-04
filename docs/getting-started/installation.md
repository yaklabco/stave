# Installation

[Home](../index.md) > Getting Started > Installation

Stave requires Go 1.25.4 or later.

## Using go install

The recommended installation method:

```bash
go install github.com/yaklabco/stave@latest
```

This installs the `stave` binary to `$GOBIN` (or `$GOPATH/bin` if `GOBIN` is unset).

## Building from Source

Clone the repository and run the bootstrap script:

```bash
git clone https://github.com/yaklabco/stave.git
cd stave
go run bootstrap.go
```

The bootstrap script compiles and installs Stave with version metadata embedded in the binary.

## Verifying Installation

```bash
stave --version
```

This prints the version, commit hash, and build timestamp.

## Upgrading

To upgrade to the latest version:

```bash
go install github.com/yaklabco/stave@latest
```

To upgrade to a specific version:

```bash
go install github.com/yaklabco/stave@v0.1.0
```

---

## See Also

- [Quickstart](quickstart.md) - Create your first stavefile
- [Home](../index.md)

