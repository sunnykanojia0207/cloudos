# ADR-0009: Stable Platform Contracts

**Status:** Active  
**Date:** 2026-06-30  
**Author:** Sunny (Product Lead)

## Context

CloudOS has reached an architectural inflection point. The core execution
pipeline is complete and proven:

```
Source → Buildpack Engine → Artifact → Runtime → RunningInstance
```

Three major interfaces now define how every application is deployed:
`Buildpack`, `Runtime`, and `Workflow`. These interfaces are no longer
experimental — they are the foundation on which Docker runtimes, SSH
runtimes, Kubernetes runtimes, community buildpacks, and third-party
providers will be built.

Without explicit stability guarantees, future changes risk breaking
the ecosystem before it forms. Contributors, plugin authors, and AI
coding agents need to know which contracts are safe to depend on.

## Decision

The following interfaces are declared **Stable Platform Contracts**,
versioned and frozen from this point forward:

### Contract 1: Runtime

File: `kernel/runtime/runtime.go`

```
Runtime interface
  ├── Prepare(ctx, req)       → PreparedApplication
  ├── Start(ctx, app)         → RunningInstance
  ├── Stop(ctx, instanceID)
  ├── Restart(ctx, instanceID)
  ├── Destroy(ctx, instanceID)
  ├── Health(ctx, instanceID) → HealthReport
  ├── Logs(ctx, instanceID, opts) → LogStream
  └── Metrics(ctx, instanceID) → Metrics
```

Commitment: Every Runtime implementation must implement this exact
interface. New methods may be added to the interface only through
backward-compatible extension (e.g., optional interfaces that runtimes
may opt into).

### Contract 2: Buildpack

File: `kernel/buildpack/buildpack.go`

```
Buildpack interface
  ├── Name()       → string
  ├── Version()    → string
  ├── Detect(ctx, src)   → (bool, error)
  ├── Plan(ctx, src)     → (*BuildPlan, error)
  └── Build(ctx, plan)   → (*BuildResult, error)
```

Commitment: Every Buildpack implementation must implement this exact
interface. The Engine orchestrates detection, planning, and building.
No buildpack may call Runtime methods, start processes, or modify
infrastructure.

### Contract 3: Workflow

File: `kernel/workflow/`

```
Node interface
  ├── ID() / Name() / Type()
  ├── Status() / SetStatus()
  └── Dependencies()

Executor.Action dispatch
  ├── validate, resource.*, controller.*
  ├── source.clone, build.install, build.execute
  └── provider.deploy → uses Runtime interface
```

Commitment: The workflow engine is the only execution coordinator.
New actions may be added, but existing action semantics are stable.
Workflow nodes never call Docker, SSH, or Kubernetes directly.

### Contract 4: Controller

File: `kernel/controller/`

```
Controller interface
  ├── Name() → string
  ├── Kind() → string
  └── Reconcile(ctx, resource) → error
```

Commitment: Controllers reconcile resource state. They never manage
processes directly — that is the Runtime's responsibility.

### Contract 5: Resource

File: `kernel/resource/`

```
Resource interface
  ├── GetMetadata() → *Metadata
  ├── GetSpec()     → interface{}
  ├── GetStatus()   → interface{}
  └── SetStatus(status)

Registry CRUD
  ├── Create(ctx, resource)
  ├── Get(kind, id) → Resource
  ├── Update(ctx, resource)
  └── Delete(ctx, kind, id)
```

Commitment: Every entity in CloudOS is a Resource. The Registry is
the single source of truth. No subsystem may bypass the Registry
to access resource state directly.

### Contract 6: Provider

File: `providers/provider.go`

```
ProviderDescriptor interface
  ├── ID() / Name() / Version()
  └── Category() → ProviderCategory
```

Commitment: Providers describe what CloudOS can do. They do not
execute work — they register capabilities that the kernel discovers.

### Contract 7: Capability

File: `capabilities/capability.go`

```
CapabilityDescriptor interface
  ├── ID() / Name() / Version()
  └── Category() → CapabilityCategory
```

Commitment: Capabilities describe what the kernel can do for users.
They are discovered, not hardcoded.

## Versioning Policy

Each stable contract follows semantic versioning at the package level:

- **Major version** (e.g., `runtime/v2`): Breaking change to the interface.
  Must be accompanied by a codemod or migration path. Must be documented
  in an ADR. Must be announced at least one minor release in advance.

- **Minor version** (e.g., `runtime.LogOptions` adding a field): New
  functionality that does not break existing implementations. Optional
  interfaces may be added for runtimes to opt into.

- **Patch version**: Bug fixes, documentation, internal refactoring that
  does not change the public interface.

## Extension Mechanism

When new functionality is needed, prefer extension over modification:

```go
// Instead of adding to the Runtime interface:
type Runtime interface {
    Name() string
    // ... existing methods
    Snapshot(ctx, id) (*Snapshot, error)  // DON'T — breaks existing impls
}

// Use an optional interface:
type Snapshotter interface {
    Snapshot(ctx context.Context, id string) (*Snapshot, error)
}
```

Runtimes, Buildpacks, and other contract implementors may optionally
implement extension interfaces. Callers should type-assert to check:

```go
if s, ok := runtime.(Snapshotter); ok {
    snap, err := s.Snapshot(ctx, id)
}
```

## Architectural Boundaries (Reinforced from ADR-0008)

These rules remain inviolable:

1. **Workflows never call infrastructure directly.**
   No Docker, SSH, Kubernetes, or process management in workflow code.

2. **Controllers never manage processes.**
   Controllers create resources and submit workflows. Runtimes run code.

3. **Buildpacks never start applications.**
   Buildpacks detect, plan, and build. Runtime.Start() is called by the
   workflow executor, not by buildpacks.

4. **Runtimes never modify source code.**
   Runtimes execute prepared applications. They do not install dependencies,
   run build steps, or modify application files.

5. **Everything is a Resource.**
   All persistent state flows through the Resource Registry. No subsystem
   stores state independently.

## Consequences

**Positive:**
- Plugin authors have clear, stable targets to implement against
- AI coding agents can safely generate code against documented contracts
- Future runtimes (Docker, SSH, K8s) slot in without core changes
- Community buildpacks can be distributed independently
- Breaking changes are signaled, documented, and migrated

**Negative:**
- Interface changes require more ceremony (ADR, deprecation, migration)
- Some desirable features may require extension interfaces rather than
  direct interface changes
- Early design mistakes in the interfaces are harder to fix

**Neutral:**
- The current interfaces have been validated through two major milestones
- ADR-0008 (Runtime Contract) and ADR-0009 (this document) form a pair
  that together define the platform's execution architecture
