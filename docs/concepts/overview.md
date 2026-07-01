# CloudOS Concepts

> **CloudOS v0.6**

This guide explains the core concepts behind CloudOS.

---

## Overview

CloudOS deploys applications from Git repositories. When you run
`cloudosctl deploy https://github.com/user/repo`, the following happens:

1. An **Application Resource** is created
2. An **Application Controller** detects the new resource and creates a **Workflow Execution**
3. The **Workflow Engine** runs the deployment through its steps
4. A **Buildpack** detects the stack and builds the application
5. A **Runtime** starts the application and allocates a port
6. A **Health Check** verifies the application is responding
7. A **Deployment Report** captures everything that happened

---

## Applications

An Application is the top-level concept. It represents a deployed application
and tracks its entire lifecycle:

- **Creation** — when you run `cloudosctl deploy`
- **Deployments** — each deploy creates a new numbered deployment
- **Status** — the current phase (Running, Deploying, Error, Stopped)
- **Health** — latest health check result (Healthy, Degraded, Unhealthy)
- **URL** — where the application is accessible

```json
{
  "id": "go-api",
  "kind": "Application",
  "metadata": { "name": "go-api" },
  "spec": {
    "source": { "type": "git", "url": "https://github.com/cloudos-examples/go-api" }
  },
  "status": {
    "phase": "Running",
    "health": "Healthy",
    "url": "http://localhost:31245",
    "deploymentCount": 1
  }
}
```

---

## Deployments

Every `cloudosctl deploy` creates a numbered deployment. Each deployment records:

| Field | Description |
| :---- | :---------- |
| Number | Sequential deployment number (#1, #2, ...) |
| Repository | Source Git URL |
| Branch | Git branch (default: main) |
| Commit SHA | Exact commit that was deployed |
| Detected Runtime | Language/framework detected (Go, Node, Python, etc.) |
| Buildpack | Buildpack used for this stack |
| Build Success | Whether the build completed successfully |
| Runtime Name | Which runtime executed the application |
| Health Status | Result of the health check |
| Endpoint | Allocated URL |
| Duration | Total deployment time |
| Errors | Any errors that occurred |
| Workflow Steps | Number of steps in the workflow |

---

## Workflows

A Workflow is a sequence of steps that runs for every deployment. The standard
deployment workflow has 6 steps:

```
1. Validate Application     — check configuration
2. Clone Source Repository  — git clone from URL
3. Build Artifact           — detect stack + build
4. Deploy Application       — run in runtime
5. Health Check              — HTTP GET /health
6. Complete Deployment      — record results
```

Each step has:
- **Status** — succeeded, failed, running, skipped, cancelled
- **Result** — human-readable description of the outcome
- **Error** — error message if the step failed

---

## Buildpacks

Buildpacks detect the application stack and determine how to build it. CloudOS
includes 7 built-in buildpacks:

| Buildpack | Detection Signal | Build Command | Artifact |
| :-------- | :--------------- | :------------ | :------- |
| **Go** | `go.mod` | `go build -o app` | Binary |
| **Node.js** | `package.json` (no react/next) | `npm install` | Node.js app |
| **React** | `package.json` + `react` | `npm run build` | Static `dist/` |
| **Next.js** | `package.json` + `next` | `next build` | Static + SSR |
| **Python** | `requirements.txt` | `pip install -r requirements.txt` | Python app |
| **Laravel** | `composer.json` | `composer install` | PHP app |
| **Static** | (fallback — no detection) | `copy` | Static files |

Buildpacks implement the `buildpack.cloudos.io/v1` interface:

```go
type Buildpack interface {
    Detect(ctx context.Context, projectDir string) (bool, error)
    Plan(ctx context.Context, projectDir string) (*BuildPlan, error)
    Build(ctx context.Context, plan *BuildPlan) (*Artifact, error)
}
```

---

## Runtimes

Runtimes execute built applications. CloudOS has two active runtime
implementations:

### LocalRuntime

The default runtime. Runs applications as local processes.

- **Isolation:** None (shares the host process space)
- **Dependencies:** None
- **Logs:** stdout/stderr captured
- **Use case:** Development, testing

### OCI Runtime (Docker)

The container runtime. Runs applications in Docker containers.

- **Isolation:** Container
- **Dependencies:** Docker daemon
- **Logs:** Docker container logs
- **Use case:** Production-like environment locally

Both runtimes implement the `runtime.cloudos.io/v1` interface:

```go
type Runtime interface {
    Prepare(ctx context.Context, app *Application) (*Environment, error)
    Start(ctx context.Context, app *Application, env *Environment) error
    Stop(ctx context.Context, app *Application) error
    Destroy(ctx context.Context, app *Application) error
    Logs(ctx context.Context, app *Application) (LogStream, error)
}
```

The OCI Runtime proved **contract substitutability** — the Runtime interface
is stable enough that a second implementation plugged in with zero changes
to Workflow, Controller, Buildpack, or Certification.

---

## Resources

CloudOS uses a resource-oriented API. Every entity is a resource:

| Resource | Kind | Description |
| :------- | :--- | :---------- |
| Application | `Application` | A deployed application |
| Capability | `Capability` | An abstract capability (compute, storage, etc.) |
| Provider | `Provider` | A concrete implementation of a capability |

Resources are accessed via:

```
GET    /api/v1/resources/{kind}
GET    /api/v1/resources/{kind}/{id}
POST   /api/v1/resources/{kind}
PUT    /api/v1/resources/{kind}/{id}
DELETE /api/v1/resources/{kind}/{id}
```

---

## Deployment Report

Every deployment produces a structured report with 19 fields:

```text
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

---

## Architecture Diagram

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
│  ┌────────────────┐ │        │  ┌──────────────────────┐ │
│  │ Go Buildpack   │ │        │  │ LocalRuntime         │ │
│  │ Node Buildpack │ │        │  │ (local processes)    │ │
│  │ Python Bp      │ │        │  └──────────────────────┘ │
│  │ React Bp       │ │        │  ┌──────────────────────┐ │
│  │ Next.js Bp     │ │        │  │ OCI Runtime (Docker) │ │
│  │ Laravel Bp     │ │        │  │ (containers)         │ │
│  │ Static Bp      │ │        │  └──────────────────────┘ │
│  └────────────────┘ │        └──────────────────────────┘
└─────────────────────┘
```

---

## Certification

CloudOS includes a certification test suite that validates the full pipeline:

1. **Detect** — the correct buildpack is selected
2. **Plan** — the build plan includes the right commands
3. **Build** — the artifact is produced without errors
4. **Runtime** — the application starts and binds to a port
5. **Health** — the application responds to HTTP requests
6. **Logs** — structured logs are emitted during deployment

7 stacks are certified (Go) or detection-verified (Node, React, Next.js,
Python, Laravel, Static).

See [COMPATIBILITY.md](../../COMPATIBILITY.md) for the full certification matrix.
