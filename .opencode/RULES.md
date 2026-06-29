# CloudOS Development Rules

## Core Rules

1. **No circular dependencies** between core modules
2. **Plugin isolation** — Plugins must not import other plugins directly
3. **API compatibility** — Breaking changes require major version bump
4. **Test coverage** — Minimum 80% coverage for core packages
5. **Documentation required** — Every public API must have doc comments
6. **Security first** — All inputs validated, all outputs sanitized

## Development Workflow

1. Branch from `main` following conventional commits naming
2. Run `make lint && make test` before committing
3. Update changelog for user-facing changes
4. Request review from relevant team members
5. Squash merge with conventional commit message

## Repository Structure

- `apps/*` — Independent deployable applications
- `core/*` — Internal Go packages (not imported by external projects)
- `plugins/*` — Plugin implementations
- `packages/*` — Shared TypeScript/React libraries
- `sdk/*` — Client SDKs for external consumers
- `cli/` — CLI tool (single Go module)
