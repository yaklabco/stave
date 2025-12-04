# Namespaces

[Home](../index.md) > [User Guide](stavefiles.md) > Namespaces

Namespaces group related targets under a common prefix.

## Defining a Namespace

Create a type alias of `st.Namespace`:

```go
type Build st.Namespace
```

Methods on this type become namespaced targets:

```go
// Docker builds a Docker image.
func (Build) Docker() error {
    return sh.Run("docker", "build", "-t", "myapp", ".")
}

// Binary builds a native binary.
func (Build) Binary() error {
    return sh.Run("go", "build", "-o", "myapp", ".")
}
```

## Invoking Namespaced Targets

Use the format `namespace:target`:

```bash
stave build:docker
stave build:binary
```

Namespace and target names are case-insensitive.

## Listing Namespaced Targets

```bash
stave -l
```

Output:

```text
Targets:
  build:binary    builds a native binary.
  build:docker    builds a Docker image.
```

## Namespaces in Dependencies

Reference namespaced targets in `st.Deps` by passing the method value:

```go
func All() {
    st.Deps(Build.Docker, Build.Binary)
}
```

Or use `st.F` if the method requires arguments:

```go
func (Deploy) Environment(env string) error {
    // ...
}

func DeployAll() {
    st.Deps(
        st.F(Deploy.Environment, "staging"),
        st.F(Deploy.Environment, "production"),
    )
}
```

## Multiple Namespaces

Define as many namespaces as needed:

```go
type Build st.Namespace
type Test st.Namespace
type Deploy st.Namespace

func (Build) All() error { /* ... */ }
func (Test) Unit() error { /* ... */ }
func (Test) Integration() error { /* ... */ }
func (Deploy) Staging() error { /* ... */ }
func (Deploy) Production() error { /* ... */ }
```

Targets:

```text
build:all
test:unit
test:integration
deploy:staging
deploy:production
```

## Combining Namespaced and Top-Level Targets

Namespaced and top-level targets coexist:

```go
func Build() error { /* ... */ }       // stave build

type Docker st.Namespace
func (Docker) Build() error { /* ... */ }  // stave docker:build
```

---

## See Also

- [Targets](targets.md) - Defining targets
- [Dependencies](dependencies.md) - Using dependencies
- [Home](../index.md)
