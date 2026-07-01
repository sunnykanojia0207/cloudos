# CloudOS Architecture

## Overview

CloudOS is a developer tool that deploys applications from Git repositories
with full deployment observability. The architecture centers on three core
contracts (Runtime, Buildpack, Workflow) frozen at v1.0.

```
┌─────────────────────────────────────────────────────────┐
│                      cloudosctl CLI                      │
│  doctor │ deploy │ logs │ status │ ps │ open │ timeline  │
└──────────────────────┬──────────────────────────────────┘
                       │ HTTP REST API
┌──────────────────────▼──────────────────────────────────┐
│              Control Plane API (:8080)                    │
│  /api/v1/resources/Application                            │
│  /api/v1/applications/{id}/logs                           │
│  /api/v1/applications/{id}/deployments/{n}/timeline       │
│  /api/v1/applications/{id}/deployments/compare            │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│              Application Controller                       │
│  Watches Application resources → creates Workflows       │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│              Workflow Engine                              │
│  Validate → Clone → Detect → Build → Deploy → Check      │
└──────────┬──────────────────────────────────┬───────────┘
           │                                  │
┌──────────▼──────────┐        ┌──────────────▼────────────┐
│  Buildpack Engine   │        │      Runtime Layer         │
│  7 built-in packs   │        │  LocalRuntime (process)    │
│  Go, Node, Python   │        │  OCI Runtime (Docker)      │
│  React, Next.js     │        │  (SSH, K8s planned)       │
│  Laravel, Static    │        │                             │
└─────────────────────┘        └───────────────────────────┘
```

## Core Contracts (v1.0 Frozen)

| Contract | Package | File |
| :------- | :------ | :--- |
| `runtime.cloudos.io/v1` | `kernel/runtime` | `runtime.go` |
| `buildpack.cloudos.io/v1` | `kernel/buildpack` | `buildpack.go` |
| `workflow.cloudos.io/v1` | `kernel/workflow` | `types.go` |

These contracts are frozen at v1.0. No breaking changes. Only additive
extensions via optional interfaces. The OCI Runtime proved contract
substitutability — a second implementation plugged in with zero changes
to Workflow, Controller, Buildpack, or Certification.

## Directory Structure

```
CloudOS/
├── kernel/              # Operating system core
│   ├── kernel.go        # Kernel orchestrator
│   ├── lifecycle/       # Component lifecycle management
│   ├── events/          # In-memory event bus
│   ├── scheduler/       # Task scheduling
│   ├── health/          # Health check aggregation
│   ├── security/        # Authentication and authorization
│   ├── registry/        # Generic name-based registry
│   ├── plugin/          # Plugin interface and registry
│   ├── di/              # Dependency injection container
│   ├── runtime/         # Runtime interface (frozen v1)
│   │   ├── local/       # LocalRuntime implementation
│   │   └── oci/         # OCI Runtime implementation (Docker)
│   ├── buildpack/       # Buildpack interface (frozen v1)
│   ├── workflow/        # Workflow engine (frozen v1)
│   ├── application/     # Application controller + DeploymentReport
│   ├── source/          # Git source handling
│   ├── intent/          # Intent processing
│   ├── project/         # Project management
│   └── resource/        # Resource management
│
├── packages/            # Shared libraries
│   ├── config/          # YAML config with env interpolation
│   ├── logging/         # Structured logging (slog wrapper)
│   ├── errors/          # Typed error framework
│   ├── types/           # Shared domain types
│   ├── version/         # Version information
│   ├── build/           # Build metadata
│   └── sdk-go/          # Go API client SDK
│
├── apps/                # Frontend applications
│   └── dashboard/       # React dashboard (dark-first, application-centric)
│
├── tools/               # CLI tools
│   ├── cloudos/         # Kernel binary entry point
│   └── cloudosctl/      # CLI (9 commands)
│
├── cloudos-examples/    # 7 example applications (Go, Node, React, etc.)
├── tests/               # Certification tests + integration tests
│   └── certification/   # 7 stack certifications + harness
│
├── docs/                # Documentation
│   ├── cli/             # CLI reference
│   ├── concepts/        # Architecture and concepts
│   ├── getting-started/ # Installation, quick-start, prerequisites
│   └── releases/        # Release notes and checklists
│
└── adr/                 # Architecture Decision Records
```

## Key Design Decisions

1. **Single Go module** — All Go code lives under one `go.mod` for simple
   dependency management.
2. **No Kubernetes dependency** — Single binary deployment.
3. **SQlite for state** — Zero-configuration database for single-node operation.
4. **In-memory event bus** — No NATS/Redis dependency. Sufficient for single-node.
5. **Clean Architecture** — Inner layers (contracts) define interfaces.
   Outer layers (implementations) implement them.
6. **Contracts frozen at v1.0** — Runtime, Buildpack, Workflow APIs are
   stable. No breaking changes. Only additive extensions.
