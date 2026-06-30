# ADR-0004: Workflow Engine

## Context

CloudOS has an Intent Engine that parses natural language user goals and
produces an execution plan. The initial implementation (Sprint 7) uses a
linear Parse → Plan → Execute pipeline that runs synchronously in a single
goroutine with no retry, no rollback, no pause/resume, and no persistent
execution state.

As we add more resource types (Database, Deployment, Storage, Secrets), the
execution of each intent will involve more steps. A linear pipeline cannot
support:

- Parallel execution of independent steps
- Retry with exponential backoff on transient failures
- Rollback of completed steps when a later step fails
- Pause/resume for human approval gates
- Observability (progress, history, audit trail)
- Long-running operations that survive process restarts

We need a general-purpose execution engine that treats every user action as
a workflow — a DAG of nodes with dependency tracking, lifecycle management,
and pluggable execution.

## Decision

### Separation: Definition vs Run

We adopt the **immutable definition / mutable run** pattern used by Temporal,
AWS Step Functions, and Apache Airflow:

- **WorkflowDefinition** — an immutable blueprint. It contains the node graph,
  dependencies, and action specifications. Once created, it is never modified.
- **WorkflowRun** — a mutable execution instance. It clones the definition's
  nodes and tracks their live status through the lifecycle. Multiple runs can
  share the same definition.

This separation enables:
- Replay: re-execute a failed run from a stored definition
- Auditing: definitions are immutable records of what was intended
- Debugging: compare a failed run against a successful run of the same definition

### Node Hierarchy

We define a `Node` interface with concrete types for each node kind:

```
Node (interface)
├── TaskNode       — executes an action against the kernel (implemented)
├── ConditionNode  — branches based on previous output (future)
├── ParallelNode   — fan-out/fan-in (future)
├── DelayNode      — waits for a duration (future)
├── ApprovalNode   — blocks until human approval (future)
├── EventNode      — waits for an external event (future)
└── EndNode        — terminal marker (implemented)
```

Only `TaskNode` and `EndNode` are implemented in Sprint 8. The interface
allows future node types without changing the Engine, Scheduler, or Graph.

### DAG-based Execution Graph

Nodes form a **directed acyclic graph (DAG)** where edges represent
dependency relationships. The Scheduler uses topological ordering to
determine which nodes are eligible to run. Independent nodes can execute
in parallel (when the Executor and Queue support concurrency).

### Subsystem Responsibilities

We separate concerns into distinct components:

| Component | Responsibility |
|-----------|---------------|
| **Graph** | DAG operations: topological sort, cycle detection, transitive dependency resolution |
| **Scheduler** | Decides which nodes are eligible to run next based on dependency status |
| **Executor** | Runs a single TaskNode against kernel subsystems (resource registry, controllers, health) |
| **Queue** | In-memory channel-backed FIFO queue for async execution with pluggable interface |
| **Engine** | Top-level coordinator: submit, cancel, pause, resume, scheduler loop, event publishing |
| **RetryEvaluator** | Determines whether a failed node should be retried with exponential backoff |
| **EventPublisher** | Publishes workflow lifecycle events to the kernel event bus |
| **Builder** | Converts Intent Engine ExecutionPlan → WorkflowDefinition DAG |

### Architecture Flow

```
User Intent
     │
     ▼
  Planner (existing)
     │
     ▼
  Workflow Builder (new)
     │
     ▼
  WorkflowDefinition (immutable)
     │
     ▼
  WorkflowRun (mutable)
     │
     ▼
  Queue
     │
     ▼
  Scheduler → Executor → (repeat) → Complete
     ▲                            │
     └──── WorkflowRun state ─────┘
```

The Engine dequeues a run, asks the Scheduler which nodes are ready,
executes them via the Executor, updates the run's node statuses, and
re-enqueues until all nodes are terminal.

### Execution as a CloudOS Resource (Future)

The `WorkflowRun` will be registered as a CloudOS Resource with the Resource
Engine, giving it automatic CRUD, watch, events, namespacing, and REST API.
This is deferred to a follow-up sprint.

### Queue

The Queue is an in-memory channel-backed FIFO. It is designed to be
replaceable by a persistent implementation (Redis, RabbitMQ, etc.) without
changing the Engine or Scheduler.

### Retry Policy

Each TaskNode can have a `RetryPolicy` with:
- `MaxRetries`: maximum retry attempts (default: 3)
- `BackoffBase`: initial backoff duration (default: 100ms)
- `BackoffMax`: maximum backoff cap (default: 10s)

Backoff is calculated as: `min(base * 2^attempt, max)`

### Rollback (Deferred)

We add a `Compensatable` interface but do not implement rollback logic yet.
Rollback requires every controller to define a compensation action. The
Project Controller is the only controller today, so rollback would have no
effect. This will be implemented when multiple controllers exist and actual
resource state needs to be unwound.

## Alternatives Considered

### Temporal Workflow SDK
- **Pros**: production-grade, durable execution, built-in retry
- **Cons**: external dependency (Temporal Server), Go SDK requires gRPC
  setup, overkill for CloudOS's current scope
- **Rejected for now**; may adopt later for multi-node execution

### AWS Step Functions
- **Pros**: serverless, no infrastructure to manage
- **Cons**: vendor lock-in, Amazon States Language is JSON-only,
  no local development without AWS account
- **Rejected**: CloudOS must be self-hosted

### Keep the linear pipeline
- **Pros**: simplest possible, no new code
- **Cons**: no retry, no parallel, no observability, no pause/resume
- **Rejected**: does not scale to the planned resource types

## Consequences

### Positive
- All future resource operations (Database, Deployment, Storage, Secrets,
  Domains) inherit workflow execution for free
- DAG enables parallel execution when independent nodes exist
- Retry policy is per-node and configurable
- Cancellation and pause/resume are first-class Engine operations
- The Queue layer enables future distribution (multiple workers, priorities)
- The Definition/Run separation enables replay and audit
- Builder package provides template plans that match the Intent Engine's
  existing patterns — migration is a one-liner per intent type

### Negative
- The Engine's scheduler loop polls via re-enqueue rather than using a
  push-based event model; this is simple but not optimal for latency
- In-memory queue means executions are lost on process restart; this is
  acceptable until we introduce persistent storage
- Rollback is deferred, so partial failures may leave resources in an
  inconsistent state until compensation actions are implemented
- Node status mutations are protected by the Engine's mutex but the
  current implementation does not support distributed locking

### Execution Resource (Sprint 9 Addition)

We treat `WorkflowExecution` as a first-class CloudOS Resource, registered
with the Resource Engine at kernel boot. This gives it automatic:

- **CRUD** via existing REST endpoints (`GET/POST/PUT/DELETE /api/v1/resources/WorkflowExecution/{id}`)
- **Watch** for real-time dashboard updates
- **Events** published to the kernel event bus
- **SDK** generation (consistent with all other resource types)
- **Namespacing** (executions belong to projects)
- **Labels** for filtering (`workflow.cloudos.io/status`)
- **ResourceVersion** for optimistic concurrency

#### Resource Model

```yaml
kind: WorkflowExecution

metadata:
  id: wf_42
  name: execution-wf_42
  namespace: default
  labels:
    workflow.cloudos.io/status: running

spec:
  workflowID: create-project
  intentID: intent_3
  requestedBy: admin
  priority: 0
  parameters: {}
  timeout: ""

status:
  phase: running
  progress: 0.4
  currentNode: step-2
  completedNodes: ["step-1"]
  failedNodes: []
  totalNodes: 5
  startedAt: "2026-06-30T05:00:00Z"
  result: ""
  error: ""
  conditions:
    - type: Scheduled
      status: "False"
      lastTransitionTime: "..."
    - type: Running
      status: "True"
      lastTransitionTime: "..."
      reason: ExecutionStarted
      message: "Workflow execution is running"
```

#### Conditions

Following Kubernetes conventions, conditions enable structured lifecycle
tracking for dashboards and automation:

| Condition | Purpose |
|-----------|---------|
| `Scheduled` | Execution has been submitted and queued |
| `Running` | Execution is actively being processed |
| `Paused` | Execution has been paused by user |
| `Completed` | Execution finished successfully |
| `Failed` | Execution finished with errors |
| `Cancelled` | Execution was cancelled by user |

#### Persistence Architecture

The Workflow Engine never owns history directly. It creates and updates
`WorkflowExecution` Resources through the Resource Engine's `Create()` and
`Update()` methods. The Resource Engine owns all persistence.

```
Workflow Engine           Resource Engine
     │                          │
     ├── Submit() ───────────────► Create(WorkflowExecution)
     │                          │
     ├── processItem() ─────────► Update(WorkflowExecution)  [per cycle]
     │                          │
     ├── Cancel() ──────────────► Update(WorkflowExecution)
     │                          │
     └── Complete/Fail ─────────► Update(WorkflowExecution)  [final state]
```

If the Resource Registry is unavailable, the Engine logs a warning and
continues execution — persistence is observable but not critical for
correctness.

### Workflow Service Layer (Sprint 9 Addition)

We introduce a `Service` layer between REST and the Workflow Engine, following
the principle that the *Service coordinates*, the *Engine executes*, and the
*Resource owns state*.

```
REST (future)
    │
    ▼
Workflow Service    ← business logic, coordination
    │
    ▼
Workflow Engine     ← execution, scheduling, DAG
    │
    ▼
WorkflowExecution   ← persistence via Resource Engine
```

The Service provides:

| Method | Description |
|--------|-------------|
| `Submit` | Creates a run from a definition and enqueues it |
| `Get` | Returns the current state of a run |
| `List` | Returns all active runs |
| `GetExecution` | Reads the persistent WorkflowExecution Resource |
| `ListExecutions` | Lists all WorkflowExecution Resources |
| `Pause` | Pauses a running workflow |
| `Resume` | Resumes a paused workflow |
| `Cancel` | Cancels a running or pending workflow |
| `Retry` | Creates a new run preserving succeeded node statuses, retrying only failed/skipped nodes |
| `Replay` | Creates a brand-new execution from a completed run's definition |
| `Clone` | Same as Replay with parameter modifications |

#### Retry

Retry preserves the status of already-succeeded nodes so only failed,
skipped, or cancelled nodes are re-executed. This enables partial retry of
multi-step workflows without redoing completed work.

```
Original Run:
  Step 1 ✔
  Step 2 ✘   ← failed
  Step 3 ⏳  ← skipped (depends on 2)

Retry Run:
  Step 1 ✔   ← preserved from original
  Step 2 ⏳  ← reset to pending, re-executes
  Step 3 ⏳  ← reset to pending, executes after 2
```

#### Replay

Replay creates a fully fresh execution from a completed run's definition.
Every node runs from scratch — no preserved statuses. This is useful for
testing the same workflow multiple times.

#### Clone

Clone is Replay with parameter overrides. The user can modify inputs
(requestedBy, priority, parameters) before re-executing.

### Long-term Roadmap

```
Sprint 9  │ Workflow Service ✓
Sprint 10 │ Intent Integration
Sprint 11 │ Workflow Dashboard
Sprint 12 │ Artifacts (logs, plans, reports as Resources)
Sprint 13 │ SSE Streaming (GET /api/v1/workflow-executions/{id}/watch)
Sprint 14 │ Mission Engine (continuous objective maintenance)
          │ Auth, Storage, Database, Deployments
```

### Migration Path
1. Keep the existing Intent Engine Executor for backward compatibility
2. New intents produce WorkflowDefinitions via the Builder
3. After all intent types are migrated, remove the linear executor
4. WorkflowExecution is registered as a Resource in kernel boot
