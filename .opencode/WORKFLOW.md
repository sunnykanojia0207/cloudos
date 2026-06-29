# CloudOS Workflows

## Standard Development Workflow

1. **Plan** — Use `@architect` to design system changes
2. **Spec** — Document API contracts in `specs/`
3. **Implement** — Build with TDD (test-first)
4. **Review** — Use `@reviewer` for code review
5. **Test** — Run full test suite
6. **Document** — Update relevant docs
7. **Deploy** — Use `@devops` for deployment

## Feature Workflow

```
Request → Architect Review → Spec → Backend → Database → Frontend → Review → Test → Deploy
```

For plugin features:
```
Request → Plugin Spec → Plugin SDK → Plugin Impl → Marketplace → Review → Test → Publish
```

## Emergency Fix Workflow

```
Bug Report → Debug → Fix → Test → Security Review → Hotfix Deploy → Postmortem
```
