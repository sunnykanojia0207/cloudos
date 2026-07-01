# ADR-0006: Natural Infrastructure

## Context

CloudOS has proven it can deploy software end-to-end: Git Repository → Running Application URL. The execution loop is verified via a golden integration test.

But the user interface is still a **programming interface**, not a **natural one**. To deploy an application today, you must:

1. Create an Application resource programmatically
2. Wait for the controller to reconcile it
3. Poll for the URL

This is no better than a shell script. The AI-native promise of CloudOS — "Deploy my React app" → Running URL — is not yet fulfilled.

We need a **Natural Infrastructure** layer: a thin, deterministic bridge between natural language and resource creation.

### Constraints

- **The AI never deploys.** The Intent Engine only creates resources. The existing controller → workflow → runtime chain handles execution deterministically. This separation of concerns is inviolable.
- **No new infrastructure primitives.** No Docker, no Kubernetes, no SSH, no remote nodes until the AI-native loop is complete.
- **Trust is the product.** Users must see what CloudOS plans to do before it does it (Plan Preview / Explain Mode).
- **Minimum new code.** Reuse the existing Intent Engine, Application Controller, and Workflow Engine. Add only what's missing.

## Decision

### Architecture

```
User: "Deploy my React app from https://github.com/user/app"

↓

[1] Intent Parser (extended)
    ↓ "Deploy from URL" → {sourceURL, appName?, runtime?}

↓

[2] Plan Preview
    ↓ "I will: Create Application, Clone Repository, Install, Build, Start, Health Check"
    ↓ "Estimated: 23 seconds. Continue?"
    ↓ User confirms

↓

[3] Intent Executor (extended)
    ↓ Creates Application resource with spec derived from intent

↓

[4] Application Controller (existing)
    ↓ Validates, creates WorkflowDefinition, submits workflow

↓

[5] Workflow Engine (existing)
    ↓ Executes deployment DAG

↓

[6] URL
```

### 1. Intent Parser — Deploy Intents

Extend the existing parser to recognize patterns like:

```
deploy <my|a> <runtime?> app from <url>
deploy from <url>
create application <name> from <url>
deploy <url>
```

The parser extracts:
- `sourceURL` (required) — GitHub or other git URL
- `appName` (optional, derived from URL if omitted)
- `runtime` (optional, auto-detected by buildpacks if omitted)

### 2. Plan Preview

The planner generates a `PlanPreview` alongside the `Plan`:

```go
type PlanPreview struct {
    Title       string          // "Deploy React App"
    Description string          // "I will deploy your application to CloudOS"
    Steps       []PreviewStep   // ordered steps with icons
    Estimated   string          // "~23 seconds"
    Resources   []ResourceSummary // resources to be created
}

type PreviewStep struct {
    Icon        string // "✓" "→" "⚡"
    Description string // "Clone repository"
    Detail      string // "https://github.com/user/app → main"
}

type ResourceSummary struct {
    Kind string // "Application"
    Name string // "react-app"
    Spec string // "Runtime: Node → Port: 3000"
}
```

The Plan Preview is returned BEFORE the plan is executed. The API returns:
1. `POST /api/v1/intents` → `{id, preview}` (no execution yet)
2. `POST /api/v1/intents/{id}/confirm` → starts execution

This keeps the AI transparent and trustworthy.

### 3. Intent Executor — Deploy Action

The executor gains a new action handler for deploy intents. It:

1. Validates the source URL and extracts metadata
2. Creates an Application resource in the registry
3. Returns the Application ID

The existing Application Controller handles everything else (workflow → deploy → URL).

### 4. Knowledge Engine (later priority)

Store deployment history as structured knowledge:

```go
type DeploymentKnowledge struct {
    AppID       string
    Runtime     string
    BuildTime   time.Duration
    DeployTime  time.Duration
    HealthTime  time.Duration
    Logs        []string
    Success     bool
    Error       string
    CreatedAt   time.Time
}
```

This enables AI queries like "Why was this deployment faster?" or "What failed yesterday?" without searching raw logs.

## Alternatives Considered

### AI agent directly creates Application resource

The LLM generates an Application resource YAML and posts it to the API.

Rejected because:
- Couples AI behavior to resource schema — changing the schema breaks prompts
- No structured preview — user can't see what's planned
- Harder to make deterministic
- Hallucination risk (invented fields, invalid values)

### AI agent directly runs deployment workflow

The LLM calls the Workflow Service directly with a deployment workflow.

Rejected because:
- Violates the "AI never deploys" principle
- Bypasses Application controller validation
- No uniform resource model

### Skip Plan Preview, just execute

Rejected because:
- Trust is the product
- Without preview, the AI is a black box
- Users need to see what's happening before committing

## Consequences

### Positive

- **AI-native UX**: "Deploy my React app from URL" → Running URL in one flow
- **Trustworthy**: Plan Preview shows everything before execution
- **Deterministic**: The AI only creates resources; the execution chain is unchanged
- **Minimum new code**: Reuses Intent Engine, Application Controller, Workflow Engine
- **No new primitives**: Extends existing types, adds no new infrastructure
- **Extensible**: New intent types (scale, stop, update) follow the same pattern

### Negative

- **Parser complexity**: Natural language patterns are fuzzy; misparses lead to confusion
- **Plan Preview is a promise**: If the plan changes during execution, the preview was wrong
- **Knowledge Engine deferred**: Deployment history queries won't work until Phase 2

### Migration

1. Extend Intent Parser — add deploy patterns
2. Add PlanPreview to Planner
3. Extend Intent Executor — add deploy action handler
4. Add confirm endpoint for plan preview
5. Wire up integration tests
6. Knowledge Engine (next sprint)

## AI-Native but Not AI-Brittle

The critical design principle:

- The **Intent Engine** understands natural language
- The **Resource Engine** manages state
- The **Controller Runtime** reconciles state
- The **Workflow Engine** executes tasks
- The **Providers** run code

Each layer is independently testable, debuggable, and replaceable. The AI is just the top layer — it parses input and creates resources. Everything else is deterministic Go code.

If the AI is unavailable, CloudOS still works. Users can create Application resources directly via API or dashboard. Natural language is the default UX, not the only UX.
