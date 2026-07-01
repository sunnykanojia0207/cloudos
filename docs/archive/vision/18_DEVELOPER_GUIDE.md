# Developer Guide — Engineering Workflow

> **Document:** 13_DEVELOPER_GUIDE.md  
> **Depends On:** [14_CODING_STANDARD.md](./14_CODING_STANDARD.md)  

---

## 1. Development Pipeline

CloudOS follows a **spec-first, test-driven, review-gated** development pipeline:

```
IDEA
  │
  ▼
MASTER SPEC ← Always reference 01_MASTER_SPEC.md
  │
  ▼
EPIC ← Defined in /epics/{feature}/README.md
  │
  ▼
TECHNICAL DESIGN ← Architecture Decision Record in /decisions/
  │
  ▼
TASKS ← Breakdown in /tasks/{feature}/{NNN}_description.md
  │
  ▼
AI PROMPTS ← Reusable prompts in /prompts/{domain}/
  │
  ▼
IMPLEMENTATION ← Code + Tests
  │
  ▼
CODE REVIEW ← @reviewer agent + human review
  │
  ▼
QA / STAGING ← Integration tests + staging deploy
  │
  ▼
RELEASE ← SemVer + changelog
```

---

## 2. Repository Workflow

### 2.1 Branch Strategy

```
main           ← Production-ready, protected
  │
  ├── develop  ← Integration branch
  │     │
  │     ├── feat/feature-name    ← Feature branches
  │     ├── fix/bug-description  ← Bug fixes
  │     ├── docs/topic           ← Documentation
  │     └── refactor/area        ← Refactoring
  │
  └── release/v*.*.*  ← Release branches
```

- `main` is protected — no direct commits
- `develop` is the default branch for PRs
- Feature branches branch from `develop`, merge back to `develop`
- Release branches branch from `develop` when ready to release

### 2.2 Conventional Commits

```
type(scope): description

types: feat, fix, docs, style, refactor, perf, test, chore, ci
scope: core, cli, dashboard, mobile, desktop, plugins, sdk, docs

Examples:
  feat(core): add PostgreSQL database provider
  fix(cli): correct help text for deploy command
  docs(api): update GraphQL schema documentation
  perf(core): optimize plugin startup time by 40%
```

---

## 3. Specification Workflow

### 3.1 Before Writing Code

Every feature must be preceded by the appropriate specification level:

| Feature Scope | Required Specification | Location |
|---------------|----------------------|----------|
| New capability | Epic (requirements + architecture) | `/epics/{feature}/` |
| Cross-cutting change | Architecture Decision Record | `/decisions/` |
| New API endpoint | API schema update | Review existing spec |
| New plugin | Plugin requirements | `/epics/{plugin}/` |
| UI change | UI spec (if complex) | Refer to `06_UI_SYSTEM.md` |

### 3.2 Specification Review Process

```
1. Author writes spec in /epics/{feature}/
2. @architect reviews for consistency with MASTER_SPEC
3. @security reviews for security implications
4. Spec approved → tasks created in /tasks/
5. Implementation begins
```

---

## 4. AI Workflow

### 4.1 Using OpenCode Agents

CloudOS leverages AI agents for all development phases:

| Phase | Agent | Prompt Location |
|-------|-------|-----------------|
| Architecture | `@architect` | `/prompts/architecture/` |
| Backend | `@backend` | `/prompts/backend/` |
| Frontend | `@frontend` | `/prompts/frontend/` |
| Database | `@database` | `/prompts/database/` |
| Plugins | `@plugin` | `/prompts/plugins/` |
| AI/ML | `@ai-ml` | `/prompts/ai/` |
| Review | `@reviewer` | `/prompts/review/` |
| Security | `@security` | (uses security checklist) |
| Testing | `@testing` | `/prompts/testing/` |

### 4.2 Agent Instructions

Every AI prompt should include:

```
Context:
- Reference to relevant spec/feature/task
- Current state of the codebase
- Specific files to modify or create

Requirements:
- What needs to be built
- Acceptance criteria
- Non-functional requirements

Constraints:
- Coding standards to follow
- Dependencies to respect
- Patterns to use or avoid

Verification:
- How to verify the work is correct
- Tests to run
- Edge cases to consider
```

---

## 5. Code Review Process

### 5.1 Review Requirements

| Change Type | Required Reviewers | Automated Checks |
|-------------|-------------------|-----------------|
| Core change | 2 engineers + @security | Lint, test, coverage, vuln scan |
| New plugin | 1 engineer + @reviewer | Lint, test, WASM verification |
| UI change | 1 engineer + visual review | Lint, test, a11y check |
| Documentation | 1 engineer | Markdown lint |
| Configuration | 1 engineer | Schema validation |
| Emergency fix | 1 engineer + post-hoc review | All checks pass |

### 5.2 Review Checklist

```
□ Code follows project coding standards
□ Tests pass and coverage meets threshold
□ No security vulnerabilities introduced
□ Documentation updated (if applicable)
□ No breaking changes (or migration path documented)
□ Error handling is comprehensive
□ Logging is appropriate (not too much, not too little)
□ Performance considerations addressed
□ Backward compatibility maintained
□ No secrets or credentials exposed
```

---

## 6. Testing Strategy

### 6.1 Test Pyramid

```
         ╱╲
        ╱ E2E ╲           ← Critical paths only (Playwright)
       ╱────────╲
      ╱Integration╲        ← Service boundaries (Testcontainers)
     ╱──────────────╲
    ╱   Unit Tests    ╲    ← All business logic (Vitest, Go test)
   ╱────────────────────╲
  ╱   Static Analysis    ╲  ← Format, lint, type-check (CI)
 ╱──────────────────────────╲
```

### 6.2 Coverage Targets

| Layer | Coverage | Framework |
|-------|----------|-----------|
| Core (Go) | 85%+ | Go testing + Testify |
| API (Go) | 90%+ | Go testing + httptest |
| Dashboard (TS) | 80%+ | Vitest + RTL |
| Plugins | 70%+ | Plugin SDK test harness |
| CLI | 80%+ | Go testing |
| SDK (TS) | 85%+ | Vitest |
| E2E (critical paths) | Covered | Playwright |

### 6.3 Test Requirements

- Every feature requires tests before merge
- Bug fixes require a regression test
- Performance changes require benchmark comparisons
- Security changes require vulnerability verification

---

## 7. Definition of Ready

A task is **ready** for implementation when:

```
□ Spec exists (epic or task description)
□ Dependencies are identified and available
□ Acceptance criteria are defined
□ Expected effort is estimated
□ No blocking questions remain
□ Relevant background context is documented
```

---

## 8. Definition of Done

A feature is **done** when:

```
□ Code is implemented
□ Tests pass (unit + integration + E2E)
□ Code review is approved
□ Documentation is updated
□ Changelog entry exists
□ No security vulnerabilities
□ Performance is acceptable (or benchmarked)
□ Feature flag is in place (if applicable)
□ Migration path exists (if breaking change)
□ Monitoring and alerting are configured
```

---

## 9. CI/CD Pipeline

```
Commit → Lint → Type Check → Unit Tests → Build → Integration Tests → Security Scan → Publish Artifact
  │        │         │            │          │            │                │               │
  │     gofmt     tsc --strict  go test    go build    testcontainers   gosec/trivy    docker push
  │     prettier  go vet                    pnpm build                                  npm publish
  │     ruff                                                                             
```

### 9.1 Pipeline Stages

| Stage | Timing | Failure Action |
|-------|--------|---------------|
| Lint | Every commit | Block merge |
| Type check | Every commit | Block merge |
| Unit tests | Every commit | Block merge |
| Build | Every commit | Block merge |
| Integration tests | Every PR | Block merge |
| Security scan | Every PR + nightly | Block merge (critical) |
| Performance | Nightly | Alert |
| E2E | Every deploy to staging | Block promotion |

---

## 10. Release Strategy

### 10.1 Versioning

CloudOS follows **Semantic Versioning** (MAJOR.MINOR.PATCH):

- **MAJOR** — Breaking API changes, breaking plugin interface changes
- **MINOR** — New features, non-breaking API additions
- **PATCH** — Bug fixes, performance improvements, security patches

### 10.2 Release Cadence

| Release Type | Frequency | Process |
|-------------|-----------|---------|
| **Nightly** | Daily | Automated build from `develop` |
| **Beta** | Weekly | Tagged from `develop` after QA |
| **Stable** | Monthly | Tagged from `release/v*.*.*` |
| **Security** | As needed | Emergency patch on `main` |

### 10.3 Release Process

```
1. Create release branch: release/vX.Y.Z
2. Run full test suite
3. Run security audit
4. Update CHANGELOG.md
5. Create GitHub Release
6. Tag with semver
7. Deploy to staging (24h soak)
8. Deploy to production
9. Merge back to main + develop
```

---

## 11. Engineering Principles

| # | Principle | Explanation |
|---|-----------|-------------|
| 1 | **Spec first, code second** | No implementation without specification. No exception. |
| 2 | **Test before merge** | Every PR must pass all tests. No exceptions. |
| 3 | **Review before merge** | Every PR must be reviewed. No self-merges. |
| 4 | **Document as you build** | Docs are part of the definition of done. |
| 5 | **Fail fast, fail clearly** | Errors should be immediate and descriptive, not silent or cryptic. |
| 6 | **Own your dependencies** | Every dependency is a risk. Evaluate before adding. |
| 7 | **If it hurts, automate it** | Repetitive manual processes must be automated. |
| 8 | **Leave it better than you found it** | Always leave the codebase slightly better. |
| 9 | **Measure everything** | If it isn't measured, it isn't managed. |
| 10 | **Security is everyone's job** | Every engineer is responsible for security. Not just the security team. |

---

> **Next:** [14_CODING_STANDARD.md](./14_CODING_STANDARD.md) — Language-specific coding conventions
