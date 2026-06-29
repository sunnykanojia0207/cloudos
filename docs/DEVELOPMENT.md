# CloudOS Development Guide

## Prerequisites

- Go 1.24+
- Make (optional, GNU Make 4.x)
- Git

## Quick Start

```bash
# Clone the repository
git clone https://github.com/cloudos/cloudos.git
cd cloudos

# Build everything
go build ./...

# Run all tests
go test ./...

# Run with race detector
go test -race ./...
```

## Project Layout

```
kernel/          → Go package: core operating system
capabilities/    → Go package: abstract interfaces
providers/       → Go package: concrete implementations
packages/        → Go package: shared libraries
tools/cloudos/   → Go package: main binary entry point
apps/            → Frontend applications (React/TypeScript)
docs/            → Documentation
scripts/         → Development scripts
```

## Common Commands

### Using Make

```bash
make build        # Build all Go packages
make test         # Run all tests
make test-race    # Run tests with race detector
make test-coverage # Run tests with coverage report
make lint         # Run golangci-lint
make fmt          # Format code and tidy modules
make clean        # Clean build artifacts
make dev          # Run the kernel
```

### Using Go Directly

```bash
go build ./...                        # Build everything
go test ./...                         # Test everything
go vet ./...                          # Vet code
go run ./tools/cloudos                # Run the kernel
go run ./tools/cloudos -version       # Show version
go run ./tools/cloudos -config my.yaml # Custom config
```

## Configuration

Configuration is read from `cloudos.yaml` in the working directory by default.
Environment variables are interpolated using `${VAR:-default}` syntax.

Example:

```bash
CLOUDOS_LOG_LEVEL=debug CLOUDOS_PORT=9090 go run ./tools/cloudos
```

## Adding a New Capability

1. Define the interface in `capabilities/capability.go` or a new file.
2. Create request/response types in the same package.
3. Implement the interface in a provider under `providers/`.
4. Register the provider with the kernel.

## Adding a New Provider

1. Create a directory under `providers/` (e.g., `providers/storage/s3/`).
2. Implement the `providers.Provider` interface.
3. Implement the corresponding capability interface.
4. Register the provider with the kernel's provider registry.

## Code Quality

- All code must pass `go vet ./...` and `golangci-lint run ./...`.
- All exported symbols must have doc comments.
- Tests are required for all packages. Use `testify` for assertions.
- Commit messages must follow conventional commits:
  `type(scope): description` (e.g., `feat(kernel): add health manager`).

## Dependency Injection Strategy

CloudOS uses a simple string-keyed DI container instead of a reflection-based
framework. Dependencies are registered by name and retrieved by name:

```go
container.Register("config", cfg)
container.Register("logger", log)

cfg := container.MustGet("config").(*config.Config)
```

This keeps the dependency graph explicit and avoids runtime reflection.
