# CloudOS Style Guide

## Go

- Use `gofmt` for formatting (no exceptions)
- Follow [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Interface names end in `-er` where possible (e.g., `Storer`, `Provider`)
- Errors are suffixed with `Error` (e.g., `ErrNotFound`)
- Use `require`/`assert` from `testify` in tests

## TypeScript / React

- Use Prettier with the project config
- Use TypeScript strict mode
- Prefer interfaces over type aliases for object types
- Use named function components (no `export default`)
- Test with Vitest and React Testing Library
- State management: Zustand for global, TanStack Query for server

## Documentation

- Markdown with GitHub-flavored syntax
- Code blocks with language tags
- API docs follow OpenAPI 3.1 specification
- Architecture diagrams in Mermaid where appropriate
