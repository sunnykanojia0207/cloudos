# ADR-0010: Architecture Freeze and Phase 4 — From Building to Using

**Status:** Active  
**Date:** 2026-06-30  
**Author:** Sunny (Product Lead)

## Context

The core execution architecture of CloudOS is complete:

```
User Intent
     ↓
Workflow Service → Workflow Engine → Controller Runtime → Resource Engine
     ↓
Source → Buildpack Engine → Artifact → Runtime → RunningInstance
     ↓
Health / Logs / Metrics
```

Seven contracts are frozen (ADR-0009). The Runtime and Buildpack interfaces
are implemented and wired into the workflow executor. The entire pipeline
from "deploy my app" to "running at http://localhost:PORT" is functional.

## Decision

### Architecture Freeze

The core platform is declared **architecture-frozen effective immediately**.

No new architectural subsystems will be created. No existing contracts will
be modified. All future work falls into one of three categories:

| Category | Description |
|---|---|
| **Implementations** | New runtimes (Docker, SSH, K8s), new buildpacks (Rust, Java, .NET, Ruby, PHP, Flutter) — no contract changes |
| **User-facing features** | Deployment Report, live logs UI, dashboard improvements — no foundation changes |
| **Ecosystem** | Plugin SDK, CLI, community buildpacks, documentation — no core changes |

### Phase 4: From Building to Using

Until now, every sprint has been about **building CloudOS**.
From now on, every sprint should be about **using CloudOS**.

The guiding question changes from:

> "What subsystem should we build?"

to:

> "What can CloudOS deploy?"

### Milestones

#### Milestone A — Deploy Every Stack

Deploy the following using exactly the same workflow, no architecture changes:

- React (static artifact → npx serve)
- Next.js (node artifact → npm start)
- Go (binary artifact → direct execution)
- Laravel (source artifact → php artisan serve)
- Python (source artifact → python/gunicorn)

Each proves the Buildpack Engine + Runtime pipeline works end-to-end.

#### Milestone B — Multiple Runtimes

- LocalRuntime ✅
- DockerRuntime (next)
- SSHRuntime (after Docker)
- KubernetesRuntime (after SSH)

No Workflow, Controller, or Resource changes. Only new Runtime implementations.

#### Milestone C — More Buildpacks

- Rust, Java, .NET, Ruby, PHP, Flutter
- No Engine changes — only new Buildpack implementations

### Postponed

- **Published Artifacts as Resources** — postponed until users ask for them.
  The current user story is "deploy app → get URL", not "browse artifacts".

### Build Instead: Deployment Report

The next user-facing feature should be a Deployment Report:

```
Deployment
  Application:  crm-api
  Runtime:      Go
  Buildpack:    Go 1.24
  Repository:   github.com/...
  Commit:       abc123
  Build Time:   8.1s
  Startup Time: 0.4s
  Health:       Healthy
  URL:          http://localhost:31245
```

This is immediately useful. Once Deployment Reports exist, Published
Artifacts become a natural extension.

### CloudOS 1.0 Definition

The product target for 1.0:

A developer can:
1. Connect a Git repository
2. CloudOS automatically detects the stack
3. Build the application
4. Deploy it locally
5. View live deployment progress
6. Access the running URL
7. Inspect logs, health, and metrics
8. Redeploy after changes

Everything beyond that — Docker, Kubernetes, clusters, AI missions,
distributed execution — is CloudOS 2.x.

### Plugin SDK

The next ecosystem investment should be a formal SDK:

```
cloudos-sdk/
  runtime/
  buildpack/
  provider/
  controller/
  resource/
  intent/
  dashboard/
```

Someone should be able to build a Rust buildpack, a Bun runtime, a Podman
runtime, a Fly.io provider, or an AI capability without modifying the core.

## Completion Estimates

| Layer | Completion |
|---|---|
| Kernel | 100% |
| Event System | 100% |
| Resource Engine | 100% |
| Controller Runtime | 100% |
| Workflow Engine | 100% |
| Workflow Service | 100% |
| Runtime Contract | 100% |
| Buildpack Engine | 100% |
| Intent Engine | 95% |
| Application Platform | 90% |
| Local Deployment | 95% |

**CloudOS Core Platform: ≈ 99% complete**

## Consequences

**Positive:**
- Clear product focus — deploy real applications, prove the architecture
- No more architectural risk — all foundations are stable and frozen
- Contributor-friendly — plugin SDK lowers the barrier to entry
- Measurable progress — each deployed stack is a visible milestone

**Negative:**
- Some postponed features (clustering, distributed execution) remain in backlog
- Plugin SDK requires investment before community contributions start
- Existing buildpacks need field testing with real applications

**Neutral:**
- The architecture freeze means new feature requests get pushed to the plugin
  layer rather than the core
- CloudOS 1.0 is a deliberately scoped product — it does what it does well
  rather than trying to do everything
