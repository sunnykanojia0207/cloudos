# ADR-0011: From Architecture to Product (CloudOS 1.0)

## Status

Accepted — Immediate

## Context

CloudOS began as an architecture-first project. Seven core interfaces were
designed, implemented, validated, and frozen. The OCI Runtime milestone
(Phase 4C) proved the central hypothesis: the Runtime contract holds under
substitution — a second implementation plugged in with zero changes to
Workflow, Controller, Buildpack, or Certification.

That milestone changes the project's character.

The architecture is complete. The three core contracts (Runtime, Buildpack,
Workflow) have been validated by two implementations each. The kernel is
frozen. Every architectural question posed by ADR-0001 through ADR-0009
has been answered.

**The project has exhausted the architecture phase.**

Continuing to build architectural subsystems would be diminishing returns.
The next phase needs a different discipline: product thinking.

## Decision

### 1. Declare Phase 5 — Product Phase

Phase 1-3 built the kernel. Phase 4 validated it was replaceable.

Phase 5 ships features that users touch.

### 2. Freeze three contracts at v1.0

No breaking changes. Only additive extensions via optional interfaces.

```
Runtime API   → runtime.cloudos.io/v1   (frozen)
Buildpack API → buildpack.cloudos.io/v1 (frozen)
Workflow API  → workflow.cloudos.io/v1  (frozen)
```

Each constant is defined in its owning package and embedded in every
result from that subsystem.

### 3. Focus on five user journeys

Every feature must answer: which journey does it improve?

| Journey | Description | First feature |
|---------|-------------|---------------|
| Deploy  | Turn a URL into a running app | Deployment Report |
| Observe | See what the app is doing | Live Logs |
| Debug   | Understand why something failed | Rich error diagnostics |
| Scale   | Grow from one instance to many | Multi-instance resources |
| Share   | Let others use the platform | Plugin SDK |

### 4. Apply the Product CTO test

Every proposed feature must pass four questions:

1. Which user problem does it solve?
2. Which journey (Deploy/Observe/Debug/Scale/Share) does it improve?
3. Can the same outcome be achieved with existing features? (If yes, don't build it.)
4. Will someone notice it in a demo? (If no, it belongs later.)

### 5. Maintain the replaceability metric

Architecture quality is no longer measured by "how many components."
It is measured by "how easily can a component be replaced without
changing the kernel."

The OCI Runtime set the standard. Community buildpacks, providers, and
plugins are expected to meet the same bar.

### 6. Establish Plugin Certification

Every plugin (runtime, buildpack, provider) declares which API version
it implements. The certification harness validates that declaration.

```go
type PluginManifest struct {
    Name             string   `json:"name"`
    RuntimeAPIVersion string  `json:"runtimeAPIVersion"`  // "runtime.cloudos.io/v1"
    BuildpackVersion  string  `json:"buildpackVersion"`   // "buildpack.cloudos.io/v1"
}
```

Certification tests verify the plugin actually works with the declared
version — exactly like the existing stack certification tests.

## Consequences

### Positive

- Product focus aligns engineering effort with user value
- Frozen contracts give plugin developers a stable target
- Certification program extends naturally from stacks to plugins
- Replaceability remains the quality north star
- Every release can be judged by a single question: "Did this make
  deploying, observing, debugging, scaling, or sharing meaningfully better?"

### Negative

- Architectural work is deprioritized — hard to ship a novel subsystem
- Plugin SDK will feel incomplete until plugin certification is live
- Old habit of "one more abstraction" must be actively resisted

### Neutral

- Existing ADRs remain valid as the architectural record
- No code is deleted — just no new subsystems are added
- Bug fixes and incremental improvements to existing subsystems continue

## Implementation

### Today — Write the deployment report

The first product feature. A structured, user-facing summary of every
deployment:

```
Deployment #42
───────────────
Repository    github.com/user/api
Detected      Go 1.24
Buildpack     Go Buildpack v1
Runtime       OCI Runtime v1
Workflow      7 steps
Started       10:21:14
Completed     10:21:22
Duration      8.2 seconds
Build         Success
Health        Healthy
Endpoint      http://localhost:31245
Warnings      None
Logs          234 lines
```

### This week — Deployment Report MVP

- Collect build metadata from Workflow Execution
- Render as structured output in CLI or API response
- Store with the Application status

### This week — Live Logs stream

- Expose Runtime.Logs() through a streaming endpoint
- Basic filtering (by app, by time range)

### Next — Plugin SDK

- `cloudos create-plugin runtime` scaffolding command
- `cloudos create-plugin buildpack` scaffolding command
- Plugin certification harness (reuses certification test pattern)

## References

- ADR-0009: Seven Core Interfaces as Stable Platform Contracts
- ADR-0010: Architecture Freeze
- Phase 4C: OCI Runtime (proved contract substitutability)
- `kernel/runtime/runtime.go` — Runtime API v1.0
- `kernel/buildpack/buildpack.go` — Buildpack API v1.0
- `kernel/workflow/executor.go` — Workflow API v1.0
- `tests/certification/` — Certification test framework
