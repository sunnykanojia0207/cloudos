# CloudOS Documentation Index

> **Purpose:** Navigation map for the CloudOS documentation pyramid.
> Each document builds on the previous one. Read in order for full context.

---

## The Documentation Pyramid

```
00_INDEX.md              ← You are here — navigation map
     ↓
01_MASTER_SPEC.md        ← Single source of truth — vision, goals, requirements
     ↓
02_PRODUCT_VISION.md     ← Market analysis, competitive landscape, success metrics
     ↓
03_DESIGN_PHILOSOPHY.md  ← Philosophical foundations, principles, trade-offs
     ↓
04_FEATURES.md           ← Complete feature catalog across all categories
     ↓
05_SYSTEM_ARCHITECTURE.md ← System design, modules, interfaces, deployment
     ↓
06_KERNEL_AND_PLUGIN_ARCHITECTURE.md ← Kernel design, plugin SDK, capability interfaces, provider model, marketplace
     ↓
07_AI_OPERATING_SYSTEM.md ← AI-native OS architecture, multi-agent system, intent engine, memory, safety
     ↓
08_KERNEL.md             ← Kernel internals, 26 subsystems, lifecycle, communication, security
     ↓
09_IMPLEMENTATION_BLUEPRINT.md ← Engineering roadmap — phases, tasks, milestones, team roles, delivery plan
     ↓
10_CAPABILITIES.md       ← Complete Go interface definitions for all 12 capabilities, versioning, error contracts, wrappers
     ↓
11_PROVIDER_ARCHITECTURE.md ← Provider SDK, 4 runtimes, selection & chaining, packaging (.cosp), sandboxing
     ↓
12_UI_SYSTEM.md          ← Design language, components, mobile/desktop UX
     ↓
13_PLUGIN_SYSTEM.md      ← Plugin architecture, lifecycle, SDK, marketplace
     ↓
14_DATABASE.md           ← Database strategy, schema, providers
     ↓
15_API.md                ← GraphQL, REST, gRPC API design
     ↓
16_SECURITY.md           ← Authentication, authorization, encryption, audit
     ↓
17_DEPLOYMENT.md         ← Deployment tiers, configuration, targets
     ↓
18_DEVELOPER_GUIDE.md    ← Engineering workflow, processes, standards
     ↓
19_CODING_STANDARD.md    ← Language-specific coding conventions
     ↓
20_ROADMAP.md            ← Phased delivery plan with milestones
```

---

## Quick Reference

| Document | What It Covers | Must Read For |
|----------|---------------|---------------|
| 01_MASTER_SPEC.md | Product vision, goals, requirements, strategy | Everyone — the single source of truth |
| 02_PRODUCT_VISION.md | Market analysis, personas, competitive landscape | Product decisions, prioritization |
| 03_DESIGN_PHILOSOPHY.md | Core principles, trade-offs, design decisions | Architecture reviews, design discussions |
| 04_FEATURES.md | Complete feature catalog (~600 features) | Feature planning, scope decisions |
| 05_SYSTEM_ARCHITECTURE.md | Modules, interfaces, data flow, deployment topologies | Engineering implementation |
| 06_KERNEL_AND_PLUGIN_ARCHITECTURE.md | Kernel design, plugin system, capability interfaces, provider model, marketplace | Core platform engineering, plugin development |
| 07_AI_OPERATING_SYSTEM.md | AI-native OS, intent engine, multi-agent system, memory, safety, providers | AI architecture, LLM integration, agent development |
| 08_KERNEL.md | Kernel internals, lifecycle, 26 subsystems, IPC, security, recovery | Kernel development, platform engineering |
| **09_IMPLEMENTATION_BLUEPRINT.md** | **Engineering roadmap — 15 phases, milestones, 100+ tasks, folder ownership, team roles, DoD, release plan** | **All engineers — this is the build plan** |
| 10_CAPABILITIES.md | Complete Go interface definitions for all 12 capabilities, versioning contracts, error semantics, cross-cutting wrappers | Capability design, interface development, provider implementation |
| 11_PROVIDER_ARCHITECTURE.md | Provider SDK, registration & discovery, selection & chaining, packaging (.cosp), 4 runtimes, sandboxing | Provider development, infrastructure integration |
| 12_UI_SYSTEM.md | Design tokens, components, navigation, mobile UX | Frontend development |
| 13_PLUGIN_SYSTEM.md | Plugin lifecycle, SDK, marketplace | Plugin development |
| 14_DATABASE.md | Database strategy, schema, providers | Data layer work |
| 15_API.md | API design, GraphQL schema, REST endpoints | API development |
| 16_SECURITY.md | Auth, encryption, compliance | Security reviews |
| 17_DEPLOYMENT.md | Deployment tiers, configuration | DevOps, deployment |
| 18_DEVELOPER_GUIDE.md | Engineering workflow, CI/CD, code review | All contributors |
| 19_CODING_STANDARD.md | Go/TS/Python conventions | Code contributions |
| 20_ROADMAP.md | Phased delivery plan | Project planning |

---

## Related Directories

| Directory | Purpose |
|-----------|---------|
| `/epics/*` | Epic-level requirements and architecture per feature area |
| `/tasks/*` | Granular task breakdowns for implementation |
| `/specs/*` | API contracts and technical specifications |
| `/decisions/*` | Architecture Decision Records (ADRs) |
| `/prompts/*` | Reusable AI prompts for code generation |
| `/tests/*` | Integration, performance, security, unit tests |
