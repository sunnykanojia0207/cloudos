# Coding Standards

> **Document:** 13_CODING_STANDARD.md
> **Status:** Draft v0.1
> **Depends On:** [01_MASTER_SPEC.md](./01_MASTER_SPEC.md)

---

## 1. Language Standards

### 1.1 Go

| Rule | Standard | Enforced By |
|------|----------|-------------|
| Formatting | `gofmt` (no exceptions) | CI (gofmt -s) |
| Imports | Standard library first, then external, then internal | CI (goimports) |
| Naming | PascalCase (exported), camelCase (unexported) | Code review |
| Errors | Wrapped with `fmt.Errorf("%w")` | CI (go vet) |
| Interfaces | Small interfaces (1-3 methods preferred) | Code review |
| Tests | Testify `assert`/`require`, table-driven tests | CI |
| Concurrency | Use `errgroup` for goroutine groups | Code review |
| Context | First parameter of every function that may block | CI (lint) |

### 1.2 TypeScript / React

| Rule | Standard | Enforced By |
|------|----------|-------------|
| Formatting | Prettier (2 spaces, 100 width) | CI |
| Type Safety | `strict: true`, no `any` | CI (tsc) |
| Components | Functional + hooks, named exports | Code review |
| Props | Interface/type per component, PascalCase | Code review |
| State | Zustand (global), TanStack Query (server) | Code review |
| Styling | Tailwind CSS v4 with design tokens | Code review |
| Tests | Vitest + React Testing Library | CI |
| Imports | ESLint `sort-imports` | CI |

### 1.3 Python

| Rule | Standard | Enforced By |
|------|----------|-------------|
| Formatting | Ruff | CI |
| Types | Type hints required on all public APIs | CI (mypy) |
| Conventions | PEP 8 + PEP 257 docstrings | CI |

---

## 2. General Standards

### 2.1 Documentation

- All public APIs/exported functions must have doc comments
- Architecture decisions documented in `decisions/`
- Code examples in documentation must be tested

### 2.2 Testing

| Layer | Coverage Target | Framework |
|-------|----------------|-----------|
| Unit (Go) | 80%+ | Go testing + Testify |
| Unit (TS) | 80%+ | Vitest |
| Integration | 60%+ | Testcontainers |
| E2E | Critical paths | Playwright |
| Performance | Key scenarios | k6 |

### 2.3 Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("create bucket %q: %w", name, err)
}

// Bad
if err != nil {
    return err
}
```

### 2.4 Commit Messages

```
type(scope): description

[optional body]

[optional footer]
```

---

> **Next:** [14_ROADMAP.md](./14_ROADMAP.md) — Project roadmap
