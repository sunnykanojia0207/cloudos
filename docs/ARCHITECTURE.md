# CloudOS Architecture

## Overview

CloudOS is an AI-first, plugin-based cloud operating system. The architecture follows Clean Architecture and Hexagonal Architecture principles:

```
┌─────────────────────────────────────────────┐
│                  Apps                        │
│  (dashboard, desktop, mobile, CLI, SDK)      │
├─────────────────────────────────────────────┤
│                 Kernel                       │
│  lifecycle | events | scheduler | health     │
│  security | registry | plugin | di           │
├─────────────────────────────────────────────┤
│              Capabilities                    │
│  compute | storage | database | ai | network │
├─────────────────────────────────────────────┤
│                Providers                     │
│  compute.local | storage.local | db.sqlite   │
└─────────────────────────────────────────────┘
```

## Layer Rules

1. **Kernel** must never import providers. It discovers them through the registry.
2. **Capabilities** define interfaces only. No implementation.
3. **Providers** implement capabilities. They import capabilities but not kernel.
4. **Apps** consume the kernel API. They never import internal kernel packages.
5. **Packages** are shared libraries. All layers may import packages.

## Directory Structure

```
CloudOS/
├── kernel/         # Operating system core
│   ├── kernel.go   # Kernel orchestrator
│   ├── lifecycle/  # Component lifecycle management
│   ├── events/     # In-memory event bus
│   ├── scheduler/  # Periodic and one-shot task scheduling
│   ├── health/     # Health check aggregation
│   ├── security/   # Authentication and authorisation
│   ├── registry/   # Generic name-based registry
│   ├── plugin/     # Plugin interface and registry
│   └── di/         # Dependency injection container
│
├── capabilities/   # Abstract capability interfaces
│   ├── capability.go  # Base types and interfaces
│   ├── compute.go     # Compute capability
│   ├── storage.go     # Storage capability
│   ├── database.go    # Database capability
│   ├── ai.go          # AI capability
│   └── network.go     # Network capability
│
├── providers/      # Concrete provider implementations
│   ├── provider.go     # Base Provider interface
│   ├── compute/local   # Process-based compute
│   ├── storage/local   # Filesystem-based storage
│   └── database/sqlite # SQLite database
│
├── packages/       # Shared foundational libraries
│   ├── config/     # YAML config with env interpolation
│   ├── logging/    # Structured logging (slog wrapper)
│   ├── errors/     # Typed error framework
│   ├── types/      # Shared domain types
│   ├── version/    # Version information
│   ├── build/      # Build metadata
│   └── sdk-go/     # Go API client SDK
│
├── apps/           # Frontend applications
│   ├── dashboard/  # React dashboard (UI)
│   └── desktop/    # Tauri desktop app
│
├── tools/          # CLI tools
│   └── cloudos/    # Kernel binary entry point
│
├── docs/           # Documentation
└── scripts/        # Developer scripts
```

## Key Design Decisions

1. **Single Go module** — All Go code lives under one `go.mod` for simple dependency management.
2. **No Kubernetes dependency** — Single binary deployment for v0.1.
3. **SQLite for state** — Zero-configuration database for single-node operation.
4. **In-memory event bus** — No NATS/Redis dependency. Sufficient for single-node.
5. **Clean Architecture** — Inner layers (capabilities) define contracts. Outer layers (providers) implement them.
