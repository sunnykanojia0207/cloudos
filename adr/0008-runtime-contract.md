# ADR-0008: Runtime Contract

**Status:** Active  
**Date:** 2026-06-30  
**Author:** Sunny (Product Lead)

## Context

CloudOS deploys applications. Different applications need different execution
environments — local processes, Docker containers, remote SSH hosts, Kubernetes
clusters. Today only `LocalRuntime` exists, and it is called directly from the
workflow executor with an inline `StartConfig`.

As we add more runtimes (Docker, SSH, Kubernetes) and more build systems
(Buildpacks), we need a stable contract that separates concerns cleanly and
prevents architectural erosion.

### Problems with the current approach

1. **No separation between preparation and execution.** `Start()` accepts a
   `StartConfig` that bundles source info, build output, and runtime config
   into one call. This makes it impossible to inspect or cache a prepared
   environment before starting it.

2. **No `Destroy()` lifecycle step.** `Stop()` is treated as final, but there
   is no explicit teardown step. Resources (ports, temp directories, network
   connections) may leak.

3. **No runtime type enumeration.** Code must type-assert or string-match to
   know which runtime is in use. This prevents generic tooling from working
   across runtimes.

4. **No metrics contract.** Health is exposed but performance metrics are not.
   Every runtime implements them differently or not at all.

5. **`Logs()` returns `LogReader`.** The interface is flexible but has no
   standard filtering or streaming contract. Callers cannot specify which
   streams to follow or how many lines of history to get.

## Decision

### The Runtime Interface

We adopt the following as the **only supported execution contract**:

```go
type Runtime interface {
    // Identity
    Name() string
    Type() RuntimeType

    // Lifecycle
    Prepare(ctx context.Context, req *PrepareRequest) (*PreparedApplication, error)
    Start(ctx context.Context, app *PreparedApplication) (*RunningInstance, error)
    Stop(ctx context.Context, instanceID string) error
    Restart(ctx context.Context, instanceID string) error
    Destroy(ctx context.Context, instanceID string) error

    // Observability
    Health(ctx context.Context, instanceID string) (*HealthReport, error)
    Logs(ctx context.Context, instanceID string, opts LogOptions) (LogStream, error)
    Metrics(ctx context.Context, instanceID string) (*Metrics, error)
}
```

### Key Design Choices

#### `Prepare()` is separate from `Start()`

`Prepare()` validates the request, allocates resources (ports, directories,
network), and returns a `PreparedApplication` — an immutable descriptor of
what will run. `Start()` uses that descriptor to launch the process.

This enables:
- Inspecting what would run without starting it
- Caching prepared environments
- Parallel preparation of multiple applications
- Cleaner error messages ("port 3000 is in use" before the user confirms)

#### Everything operates on `instanceID`

Every lifecycle method takes `instanceID string`, not `Application`. This
decouples the runtime from the Application resource model. One Application
may have multiple instances (replicas, restarts). One instance always belongs
to exactly one runtime.

#### `Destroy()` is the terminal step

`Stop()` stops the process. `Destroy()` releases all resources (ports,
directories, network). This separation lets callers stop and restart without
re-preparing, while ensuring resources are freed on final teardown.

### Compatibility Promise

These rules are **inviolable**. They define the architecture's hard boundaries:

1. **The `Runtime` interface is the only supported execution contract.**
   No code outside a Runtime implementation may start, stop, or inspect
   processes directly.

2. **Workflows never call Docker, SSH, Kubernetes, or processes directly.**
   The workflow executor dispatches to `Runtime` methods by name. It has
   no knowledge of the underlying infrastructure.

3. **Controllers never manage processes directly.**
   Controllers create Application resources and submit workflow definitions.
   The workflow engine and runtime handle execution.

4. **Buildpacks never start applications.**
   Buildpacks detect, plan, and build. They produce artifacts. They do not
   call `Runtime.Start()` or manage running instances.

5. **Runtimes never modify source code.**
   Runtimes receive a prepared application directory (or artifact). They
   execute the start command and expose health/logs/metrics. They do not
   run build steps, install dependencies, or modify application files.

### Architectural Flow

```
Git Repository
     ↓
   Source
     ↓
  Buildpack           ←  detects, plans, builds
     ↓
  Artifact            ←  build output (dir, binary, image)
     ↓
  Runtime.Prepare()   ←  validates, allocates resources
     ↓
  Runtime.Start()     ←  launches process
     ↓
  RunningInstance     ←  health, logs, metrics
     ↓
  WorkflowExecution   ←  tracks state, artifacts, results
```

Notice: the Runtime never sees source code. It only sees an artifact.

Examples of artifacts by application type:

```
React       →  dist/
Go          →  binary
Laravel     →  prepared application directory
Python      →  virtual environment + source
Docker      →  container image
```

### RuntimeType Enum

```go
type RuntimeType string

const (
    RuntimeTypeLocal      RuntimeType = "local"
    RuntimeTypeDocker     RuntimeType = "docker"
    RuntimeTypeSSH        RuntimeType = "ssh"
    RuntimeTypeKubernetes RuntimeType = "kubernetes"
)
```

### New Types

```go
type PrepareRequest struct {
    AppID       string
    Name        string
    WorkDir     string
    Command     string
    Args        []string
    Port        int               // 0 = auto-assign
    EnvVars     map[string]string
    HealthCheck *HealthPolicy
    Labels      map[string]string
}

type PreparedApplication struct {
    ID        string
    AppID     string
    WorkDir   string
    Command   string
    Args      []string
    Port      int
    EnvVars   map[string]string
    Artifact  ArtifactRef         // optional, for future use
}

type LogOptions struct {
    Tail     int                  // number of recent lines to include
    Follow   bool                 // stream new lines
    Source   string               // "stdout", "stderr", or "" for both
}

type LogStream interface {
    Lines() <-chan LogEntry
    Close() error
}

type Metrics struct {
    CPUUsage    float64           // percentage (0-100)
    MemoryUsage int64             // bytes
    Uptime      time.Duration
    Timestamp   time.Time
}

type ArtifactRef struct {
    Type string
    Path string
    Hash string
}
```

### Migration Plan

1. **Expand the interface.** Add `Prepare`, `Destroy`, `Type()`, `Metrics()`,
   `LogOptions`/`LogStream` while keeping existing methods for backward compat
   during the transition.

2. **Migrate LocalRuntime.** Move port allocation, directory creation, and
   command validation into `Prepare()`. `Start()` becomes a lightweight wrapper
   that launches the process from a `PreparedApplication`.

3. **Update Workflow Executor.** Replace `runtimeManager.Start(ctx, StartConfig)`
   with `Prepare()` → `Start()` in `execProviderDeploy`.

4. **Remove old `StartConfig` and `LogReader`.** Once all callers are migrated.

### Consequences

**Positive:**
- Clean separation of concerns — each layer owns exactly one responsibility
- Runtime-agnostic workflow engine — any runtime backend works without changes
- Testable in isolation — mock Runtime interface for workflow tests
- Future-proof — Docker, SSH, K8s runtimes slot in without touching core
- Clear contributor boundaries — "where does this code go?" is always answered

**Negative:**
- One-time migration cost for LocalRuntime and workflow executor
- More methods to implement for new runtime backends
- `PrepareRequest` duplicates some fields from `Application` resource

**Neutral:**
- Old `StartConfig` type will be removed; callers must use `PrepareRequest`
- Old `LogReader` interface will remain for internal log store use; public
  API uses `LogStream`
