# CloudOS Engineering Backlog — v0.1

> **Document ID:** CLOUDOS-BACKLOG-001  
> **Status:** v1.0 — Approved  
> **Last Updated:** 2026-06-29  
> **Audience:** All Engineers — this is the build plan  
> **Depends On:** All documents 01-09

---

## Table of Contents

1. [Backlog Structure](#1-backlog-structure)
2. [Legend](#2-legend)
3. [Sprint 0: Foundation & Tooling](#3-sprint-0-foundation--tooling)
4. [Sprint 1: Kernel Core](#4-sprint-1-kernel-core)
5. [Sprint 2: API & Auth](#5-sprint-2-api--auth)
6. [Sprint 3: Dashboard](#6-sprint-3-dashboard)
7. [Sprint 4: Storage, Database & Secrets](#7-sprint-4-storage-database--secrets)
8. [Sprint 5: Deployments & Networking](#8-sprint-5-deployments--networking)
9. [Sprint 6: AI Engine](#9-sprint-6-ai-engine)
10. [Sprint 7: Marketplace & Templates](#10-sprint-7-marketplace--templates)
11. [Sprint 8: Automation & Scheduler](#11-sprint-8-automation--scheduler)
12. [Sprint 9: Monitoring & Analytics](#12-sprint-9-monitoring--analytics)
13. [Sprint 10: Production Hardening](#13-sprint-10-production-hardening)
14. [Backlog Index](#14-backlog-index)

---

## 1. Backlog Structure

```
Epic          → A major area of work (e.g., "Kernel")
  Feature     → A user-visible capability (e.g., "Boot Sequence")
    Task      → An engineer-sized unit (e.g., "Implement KPM")
      Subtask → A concrete step (e.g., "Define Subsystem interface")
```

### Sprint Overview

| Sprint | Theme | Tasks | Focus |
|--------|-------|-------|-------|
| 0 | Foundation & Tooling | 8 | Repository, Go module, CI, logging, config, testing framework |
| 1 | Kernel Core | 10 | KPM, capability registry, event bus, DI, plugin loader, health checks |
| 2 | API & Auth | 8 | REST API, JWT auth, RBAC, user management |
| 3 | Dashboard | 6 | React app, login, status, deploy pages |
| 4 | Storage, Database & Secrets | 8 | Local storage, SQLite, secrets manager |
| 5 | Deployments & Networking | 7 | Compute deployment, DNS, SSL, networking |
| 6 | AI Engine | 6 | Intent parser, planner, task engine, OpenAI integration |
| 7 | Marketplace & Templates | 6 | Plugin registry, template system, publishing |
| 8 | Automation & Scheduler | 5 | Cron scheduler, workflow engine, background jobs |
| 9 | Monitoring & Analytics | 6 | Metrics, logging pipeline, alerting, dashboards |
| 10 | Production Hardening | 8 | Security audit, performance, docs, release |
| **Total** | | **78** | |

---

## 2. Legend

### Priority

| Label | Meaning |
|-------|---------|
| P0 | 🔴 Critical — blocks everything |
| P1 | 🟡 High — important for sprint goal |
| P2 | 🟢 Medium — valuable but not blocking |
| P3 | 🔵 Low — nice to have |

### Difficulty

| Label | Meaning | Time Estimate |
|-------|---------|---------------|
| D1 | 🟢 Trivial | < 1 hour |
| D2 | 🟡 Small | 1-3 hours |
| D3 | 🟠 Medium | 3-8 hours |
| D4 | 🔴 Large | 1-3 days |
| D5 | ⚫ Unknown | Needs research |

### Task Status

| Status | Meaning |
|--------|---------|
| 📋 Backlog | Not started |
| 🏗️ In Progress | Active development |
| ✅ Done | Meets DoD |
| 🚫 Blocked | Waiting on dependency |

### Acceptance Criteria Format

```
GIVEN [context/state]
WHEN [action is performed]
THEN [expected result]
```

---

## 3. Sprint 0: Foundation & Tooling

**Theme:** Establish the development environment, project structure, and tooling that every subsequent sprint depends on.

**Goal:** `go build ./...` passes, linter is green, CI runs on every push, basic logging and config systems work.

**Duration:** ~3 days

---

### Epic: Repository Setup

#### Feature 0.1: Initialize Repository

**Business Value:** Single source of truth for all code.  
**Technical Value:** Foundation for all future work.  
**Dependencies:** None.

---

##### Task S0-T001: Initialize Go Monorepo

**Description:** Create the Go module at `core/cloudos/`, root `go.work`, and base directory structure matching the system architecture. Every directory must exist and compile (even if empty).

**Priority:** P0 🔴  
**Difficulty:** D1 🟢  
**Owner:** Platform  
**Estimated Time:** 1 hour

**Subtask S0-T001.1:** Run `go mod init github.com/cloudos/core` in `core/`  
**Subtask S0-T001.2:** Create `go.work` pointing to `core/cloudos`  
**Subtask S0-T001.3:** Create directories: `core/cloudos/cmd/cloudos/`, `core/cloudos/internal/{kernel,capability,eventbus,config,logging,metrics,auth,providers/{compute,storage,database}}`, `core/cloudos/pkg/{types,api,sdk}`  
**Subtask S0-T001.4:** Create `apps/dashboard/` with `package.json` stub  
**Subtask S0-T001.5:** Create `cli/cloudos/` with placeholder `main.go`  
**Subtask S0-T001.6:** Verify `go build ./...` and `go work sync` pass

**Acceptance Criteria:**
```
GIVEN a fresh clone
WHEN I run `go build ./...` from repository root
THEN all packages compile without errors
```

---

##### Task S0-T002: Configure Linting & Formatting

**Description:** Set up `golangci-lint`, `.editorconfig`, pre-commit hooks, and `Makefile` or `Taskfile.yml` for common commands (build, lint, test, clean).

**Priority:** P1 🟡  
**Difficulty:** D1 🟢  
**Owner:** Platform  
**Estimated Time:** 1 hour

**Subtask S0-T002.1:** Create `.golangci.yml` with Go 1.24 rules (gofmt, goimports, govet, staticcheck, errcheck, ineffassign)  
**Subtask S0-T002.2:** Create `.editorconfig` with consistent indentation settings  
**Subtask S0-T002.3:** Create `Makefile` with targets: `build`, `lint`, `test`, `clean`, `dev`  
**Subtask S0-T002.4:** Configure pre-commit hooks (gofumpt, goimports, whitespace)  

**Acceptance Criteria:**
```
GIVEN the repository is set up
WHEN I run `make lint`
THEN all Go files pass linting with zero warnings
```

---

##### Task S0-T003: Configure Git Hooks & Gitignore

**Description:** Set up `.gitignore` for Go + TypeScript monorepo, husky or lefthook for pre-commit hooks, and commit message conventions.

**Priority:** P2 🟢  
**Difficulty:** D1 🟢  
**Owner:** Platform  
**Estimated Time:** 30 minutes

**Subtask S0-T003.1:** Create `.gitignore` covering Go binaries, node_modules, IDE files, OS files  
**Subtask S0-T003.2:** Configure lefthook or pre-commit for lint + test checks  
**Subtask S0-T003.3:** Document commit message convention in `CONTRIBUTING.md`

**Acceptance Criteria:**
```
GIVEN a modified Go file
WHEN I try to commit with lint errors
THEN the commit is blocked
```

---

#### Feature 0.2: Continuous Integration

**Business Value:** Every commit is verified automatically.  
**Technical Value:** Catches regressions early, enables team workflow.

---

##### Task S0-T004: Create GitHub Actions CI Pipeline

**Description:** Create CI workflow with build, lint, test, and race-detection stages. Cache Go modules. Run on push/PR to main.

**Priority:** P0 🔴  
**Difficulty:** D2 🟡  
**Owner:** DevOps  
**Estimated Time:** 2 hours

**Subtask S0-T004.1:** Create `.github/workflows/ci.yml` with Go 1.24 setup  
**Subtask S0-T004.2:** Add `go build ./...` stage  
**Subtask S0-T004.3:** Add `golangci-lint` stage  
**Subtask S0-T004.4:** Add `go test -race ./...` stage  
**Subtask S0-T004.5:** Add Go module caching  
**Subtask S0-T004.6:** Add status badge to `README.md`

**Acceptance Criteria:**
```
GIVEN a PR is opened to main
WHEN CI runs
THEN all stages (build, lint, test) complete successfully
```

---

### Epic: Developer Experience

#### Feature 0.3: Development Tooling

**Business Value:** Fast feedback loop, consistent environment.  
**Technical Value:** Reduces onboarding time, prevents environment issues.

---

##### Task S0-T005: Set Up Development Environment

**Description:** Create `docker-compose.dev.yaml` for local development with hot-reload (air for Go, Vite for dashboard). Create `.env.example` with development defaults.

**Priority:** P1 🟡  
**Difficulty:** D2 🟡  
**Owner:** Platform  
**Estimated Time:** 2 hours

**Subtask S0-T005.1:** Create `Dockerfile.dev` for Go with `air` hot-reload  
**Subtask S0-T005.2:** Create `docker-compose.dev.yaml` with volume mounts  
**Subtask S0-T005.3:** Create `.env.example` with safe development defaults  
**Subtask S0-T005.4:** Create `scripts/dev.sh` for one-command dev startup

**Acceptance Criteria:**
```
GIVEN Docker is installed
WHEN I run `docker compose -f docker-compose.dev.yaml up`
THEN the Kernel starts with hot-reload and I can edit Go files to trigger rebuilds
```

---

##### Task S0-T006: Set Up Testing Framework

**Description:** Configure Go testing with `testing` package, set up test helpers, create test fixtures directory, configure `testify` for assertions, and create a `Makefile` test target with coverage.

**Priority:** P1 🟡  
**Difficulty:** D1 🟢  
**Owner:** QA  
**Estimated Time:** 1 hour

**Subtask S0-T006.1:** Add `testify` dependency to `go.mod`  
**Subtask S0-T006.2:** Create `core/cloudos/internal/testutil/` with common test helpers  
**Subtask S0-T006.3:** Create `tests/integration/` and `tests/e2e/` directories  
**Subtask S0-T006.4:** Add `make test` and `make test-coverage` targets  
**Subtask S0-T006.5:** Write a simple passing test to validate the framework

**Acceptance Criteria:**
```
GIVEN the testing framework is set up
WHEN I run `make test-coverage`
THEN tests run and coverage report is generated
```

---

### Epic: Core Infrastructure

#### Feature 0.4: Structured Logging

**Business Value:** Every subsystem emits structured, queryable logs from day one.  
**Technical Value:** Debugging without logs is impossible; logs must be built-in, not bolted on.

---

##### Task S0-T007: Implement Structured Logger

**Description:** Implement a structured logger using Go 1.24's `log/slog`. Support levels (debug, info, warn, error), structured fields (key-value pairs), multiple outputs (stdout JSON, file JSON with rotation), and context field propagation (trace ID, subsystem, request ID).

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Estimated Time:** 4 hours

**Subtask S0-T007.1:** Define `Logger` interface wrapping `slog.Logger`  
**Subtask S0-T007.2:** Implement JSON output handler for stdout  
**Subtask S0-T007.3:** Implement file output handler with rotation  
**Subtask S0-T007.4:** Implement context field propagation (trace ID, subsystem)  
**Subtask S0-T007.5:** Create `NewSubsystemLogger(subsystem)` factory function  
**Subtask S0-T007.6:** Verify thread-safety with `go test -race`  
**Subtask S0-T007.7:** Write unit tests for all log levels and field propagation

**Acceptance Criteria:**
```
GIVEN a subsystem logger
WHEN I call logger.Info("message", "key", "value")
THEN JSON output contains {"level":"INFO","msg":"message","key":"value","subsystem":"<name>","time":"<ISO8601>"}
```

---

#### Feature 0.5: Configuration System

**Business Value:** All subsystems read configuration from a unified source. No hardcoded values anywhere.  
**Technical Value:** Makes the system deployable to any environment without code changes.

---

##### Task S0-T008: Implement Configuration Manager

**Description:** Implement `ConfigProvider` interface, YAML file loader with environment variable interpolation (`${VAR_NAME}` with default syntax), schema validation, and file watcher for hot-reload. Publish `config.changed` events on successful reload.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Estimated Time:** 6 hours

**Subtask S0-T008.1:** Define `ConfigProvider` interface and `Config` struct  
**Subtask S0-T008.2:** Implement `YAMLConfigProvider` with file reading and parsing  
**Subtask S0-T008.3:** Implement env var interpolation (`${VAR:-default}` syntax)  
**Subtask S0-T008.4:** Implement JSON Schema validation  
**Subtask S0-T008.5:** Implement file watcher with fsnotify for hot-reload  
**Subtask S0-T008.6:** Integrate with Event Bus for `config.changed` events  
**Subtask S0-T008.7:** Create default `cloudos.yaml` config file  
**Subtask S0-T008.8:** Write unit and integration tests

**Acceptance Criteria:**
```
GIVEN a cloudos.yaml with `${PORT:-8080}`
WHEN the Kernel starts with PORT=9090
THEN the configuration resolves to port 9090

GIVEN a running Kernel
WHEN I edit cloudos.yaml
THEN the config is reloaded and a config.changed event is published
```

---

## 4. Sprint 1: Kernel Core

**Theme:** Build the minimum Kernel that can boot, initialize subsystems, register capabilities, and report health.

**Goal:** Kernel binary starts, subsystems initialize in dependency order, capabilities register, health endpoint returns 200.

**Duration:** ~5 days

---

### Epic: Kernel Runtime

#### Feature 1.1: Kernel Process Manager

**Business Value:** The Kernel is the heart of CloudOS. It must start, run, and stop reliably.  
**Technical Value:** Every other subsystem depends on KPM for lifecycle management.

---

##### Task S1-T001: Define Subsystem Interface

**Description:** Define the `Subsystem` interface: `Init(ctx, config) error`, `Start(ctx) error`, `Stop(ctx) error`, `Health(ctx) (*HealthStatus, error)`. Define `SubsystemState` enum. This is the contract every Kernel subsystem implements.

**Priority:** P0 🔴  
**Difficulty:** D2 🟡  
**Owner:** Platform  
**Dependencies:** S0-T001 (Go module exists)  
**Estimated Time:** 2 hours

**Subtask S1-T001.1:** Define `Subsystem` interface in `internal/kernel/subsystem.go`  
**Subtask S1-T001.2:** Define `SubsystemState` enum (uninitialized → initialized → running → stopping → stopped → failed)  
**Subtask S1-T001.3:** Define `HealthStatus` struct with status, message, timestamp  
**Subtask S1-T001.4:** Write Go doc comments on every method  
**Subtask S1-T001.5:** Create `MockSubsystem` for testing  
**Subtask S1-T001.6:** Write unit tests

**Acceptance Criteria:**
```
GIVEN the Subsystem interface
WHEN I implement it in a mock
THEN Init, Start, Stop, and Health all compile and return expected results
```

---

##### Task S1-T002: Implement Kernel Process Manager

**Description:** Implement `KernelProcessManager`: holds subsystem registry with dependency ordering, `Boot()` initializes in order, `Start()` starts subsystems, `Shutdown()` stops in reverse order with drain timeouts, `Crash()` handles unrecoverable failures, `Restart()` for individual subsystems. Thread-safe.

**Priority:** P0 🔴  
**Difficulty:** D4 🔴  
**Owner:** Platform  
**Dependencies:** S1-T001 (Subsystem interface)  
**Estimated Time:** 2 days

**Subtask S1-T002.1:** Define `Registration` struct with subsystem, name, dependencies, state  
**Subtask S1-T002.2:** Implement `Register(subsystem, name, deps)` with cycle detection  
**Subtask S1-T002.3:** Implement `Boot()` — topological sort, init in order, error handling  
**Subtask S1-T002.4:** Implement `Start()` — start all, timeouts, error propagation  
**Subtask S1-T002.5:** Implement `Shutdown()` — reverse order, drain timeout (30s), force kill  
**Subtask S1-T002.6:** Implement `Restart(name)` — stop + start with backoff  
**Subtask S1-T002.7:** Implement `Crash(name)` — mark failed, escalate to health manager  
**Subtask S1-T002.8:** Add thread-safety with `sync.RWMutex`  
**Subtask S1-T002.9:** Write comprehensive tests (normal boot, failure, shutdown, restart, race)

**Acceptance Criteria:**
```
GIVEN 3 subsystems where A→B→C (C depends on B depends on A)
WHEN KPM.Boot() is called
THEN A.Init() → B.Init() → C.Init() executes in order

GIVEN a running Kernel
WHEN KPM.Shutdown() is called
THEN C.Stop() → B.Stop() → A.Stop() executes with drain timeout
```

---

##### Task S1-T003: Implement Boot Sequence

**Description:** Implement `Boot()` that parses CLI flags, loads config file, initializes config manager, creates KPM, registers built-in subsystems, calls KPM.Boot(), prints banner, transitions to running state.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Dependencies:** S1-T002 (KPM), S0-T008 (Config Manager)  
**Estimated Time:** 4 hours

**Subtask S1-T003.1:** Create `cmd/cloudos/main.go` with flag parsing (--config, --port)  
**Subtask S1-T003.2:** Load config file using ConfigManager  
**Subtask S1-T003.3:** Create KPM and register all built-in subsystems  
**Subtask S1-T003.4:** Call KPM.Boot() and KPM.Start()  
**Subtask S1-T003.5:** Print CloudOS ASCII banner with version and health URL  
**Subtask S1-T003.6:** Handle SIGTERM/SIGINT for graceful shutdown  
**Subtask S1-T003.7:** Log boot duration and subsystem states

**Acceptance Criteria:**
```
GIVEN the Kernel binary
WHEN I run `./cloudos --config cloudos.yaml`
THEN:
  - Boot completes in < 1 second
  - Banner is printed with version and health URL
  - All subsystems report "running"
  - SIGTERM triggers graceful shutdown in < 5 seconds
```

---

#### Feature 1.2: Capability Registry

**Business Value:** The central catalog of everything CloudOS can do.  
**Technical Value:** Enables capability discovery, version negotiation, and provider-agnostic architecture.

---

##### Task S1-T004: Define Capability Types

**Description:** Define `Capability`, `CapabilityInfo`, `VersionConstraint` with parsing/matching. This establishes the type system that all capabilities and providers use.

**Priority:** P0 🔴  
**Difficulty:** D2 🟡  
**Owner:** Platform  
**Dependencies:** S1-T001 (Subsystem types available)  
**Estimated Time:** 3 hours

**Subtask S1-T004.1:** Define `Capability` interface in `internal/capability/`  
**Subtask S1-T004.2:** Define `CapabilityInfo` struct (name, version, features, providerID, health)  
**Subtask S1-T004.3:** Define `VersionConstraint` with parsing and matching (`>=1.0.0, <2.0.0`)  
**Subtask S1-T004.4:** Define `CapabilityFilter` for listing  
**Subtask S1-T004.5:** Write tests for version constraint parsing/matching edge cases

**Acceptance Criteria:**
```
GIVEN VersionConstraint ">=1.0.0, <2.0.0"
WHEN I match against "1.5.0"
THEN it returns true

WHEN I match against "2.0.0"
THEN it returns false

GIVEN an invalid constraint ">=abc"
WHEN I parse it
THEN I get a parse error
```

---

##### Task S1-T005: Implement CapabilityRegistry

**Description:** Thread-safe registry with Register, Unregister, Get (best version match), List (filtered), HasFeature. Publishes events on register/unregister. Supports multiple versions of the same capability.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Dependencies:** S1-T004 (Capability types), S1-T006 (Event Bus)  
**Estimated Time:** 6 hours

**Subtask S1-T005.1:** Implement `CapabilityRegistry` struct  
**Subtask S1-T005.2:** Implement `Register(info)` — validate, store, publish event  
**Subtask S1-T005.3:** Implement `Unregister(name, version)` — remove, publish event  
**Subtask S1-T005.4:** Implement `Get(name, constraint)` — find best version match  
**Subtask S1-T005.5:** Implement `List(filter)` — return matching capabilities  
**Subtask S1-T005.6:** Implement `HasFeature(name, feature)` — feature flag check  
**Subtask S1-T005.7:** Add thread-safety with `sync.RWMutex`  
**Subtask S1-T005.8:** Write unit tests with concurrent access

**Acceptance Criteria:**
```
GIVEN a registry with compute v1.0 and v2.0
WHEN I call Get("compute", ">=1.5.0")
THEN I receive compute v2.0

WHEN I call List({})
THEN both versions are returned

GIVEN 10 concurrent Register calls
WHEN all complete
THEN all 10 capabilities are registered with no race conditions
```

---

#### Feature 1.3: Event Bus

**Business Value:** All subsystems communicate through typed, asynchronous events.  
**Technical Value:** Loose coupling, audit trail, extensibility.

---

##### Task S1-T006: Implement In-Memory Event Bus

**Description:** Thread-safe in-memory pub/sub event bus. Support exact and wildcard subjects (`*`, `>`), fan-out to multiple subscribers, subscriber timeouts (5s), dead letter handling, ordered delivery per subject.

**Priority:** P0 🔴  
**Difficulty:** D4 🔴  
**Owner:** Platform  
**Dependencies:** S0-T001 (Go module)  
**Estimated Time:** 2 days

**Subtask S1-T006.1:** Define `Event` struct (id, type, source, timestamp, payload, metadata)  
**Subtask S1-T006.2:** Define `Publisher`, `Subscriber`, `EventBus` interfaces  
**Subtask S1-T006.3:** Define subject naming convention in constants  
**Subtask S1-T006.4:** Implement `InMemoryEventBus` with concurrent subscriber map  
**Subtask S1-T006.5:** Implement wildcard subject matching (`*` single, `>` multi-level)  
**Subtask S1-T006.6:** Implement subscriber timeouts and dead letter handling  
**Subtask S1-T006.7:** Implement ordered delivery per subject (FIFO per subject)  
**Subtask S1-T006.8:** Add metrics tracking (publish count, delivery latency, dropped)  
**Subtask S1-T006.9:** Write comprehensive tests (concurrency, wildcards, ordering, timeouts)

**Acceptance Criteria:**
```
GIVEN 3 subscribers on "events.kernel.*" and a publisher
WHEN publisher.Publish("events.kernel.boot.complete", payload)
THEN all 3 subscribers receive the event within 100ms

GIVEN a slow subscriber (>5s processing)
WHEN an event is published
THEN the slow subscriber is timed out and the event goes to dead letter queue

GIVEN 1000 concurrent publishes
WHEN all complete
THEN no events are lost and delivery is verified
```

---

#### Feature 1.4: Dependency Injection

**Business Value:** Clean separation of concerns, testability, hot-replaceable components.  
**Technical Value:** The DI container manages subsystem lifecycles without global state.

---

##### Task S1-T007: Implement Dependency Injection Container

**Description:** Implement a simple DI container that manages subsystem creation, dependency resolution, and lifecycle. Supports singleton and factory scopes. Used by KPM to wire up subsystems.

**Priority:** P1 🟡  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Dependencies:** S1-T001 (Subsystem interface)  
**Estimated Time:** 4 hours

**Subtask S1-T007.1:** Define `Container` interface with `Register`, `Resolve`, `Build`  
**Subtask S1-T007.2:** Implement singleton scope (same instance every resolve)  
**Subtask S1-T007.3:** Implement factory scope (new instance every resolve)  
**Subtask S1-T007.4:** Implement dependency graph validation (detect cycles)  
**Subtask S1-T007.5:** Integrate with KPM for automatic wiring  
**Subtask S1-T007.6:** Write tests with mock dependencies

**Acceptance Criteria:**
```
GIVEN a Container with A depends on B depends on C registered
WHEN I call container.Resolve(A)
THEN C is created first, then B with C injected, then A with B injected

GIVEN a circular dependency (A→B→A)
WHEN I call container.Resolve(A)
THEN a clear circular dependency error is returned
```

---

#### Feature 1.5: Lifecycle Manager

**Business Value:** The Kernel must self-heal — restart failed subsystems, drain gracefully, report health.  
**Technical Value:** Production readiness requires battle-tested lifecycle management.

---

##### Task S1-T008: Implement Health Manager

**Description:** Periodic health checks on all registered subsystems (configurable interval, default 5s). Tracks state transitions (healthy → degraded → unhealthy). Publishes health events on state change. Exposes aggregate system health.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Dependencies:** S1-T001 (Subsystem), S1-T006 (Event Bus)  
**Estimated Time:** 4 hours

**Subtask S1-T008.1:** Define `HealthManager` with `RegisterSubsystem` and `UnregisterSubsystem`  
**Subtask S1-T008.2:** Implement periodic health check goroutine (configurable interval)  
**Subtask S1-T008.3:** Implement state transition detection (healthy→degraded→unhealthy)  
**Subtask S1-T008.4:** Publish health events on state change  
**Subtask S1-T008.5:** Expose `AggregateHealth()` for API  
**Subtask S1-T008.6:** Implement health check timeout (default 5s)  
**Subtask S1-T008.7:** Write tests for all state transitions

**Acceptance Criteria:**
```
GIVEN 3 subsystems all healthy
WHEN I query aggregate health
THEN status is "healthy"

GIVEN 1 subsystem becomes unhealthy for 3 consecutive checks
WHEN the health manager detects this
THEN a "health.subsystem.unhealthy" event is published
AND aggregate health becomes "degraded"
```

---

#### Feature 1.6: Plugin Loader

**Business Value:** Providers (implementations of capabilities) must be loaded, initialized, and managed at runtime.  
**Technical Value:** The plugin loading system enables the entire provider ecosystem.

---

##### Task S1-T009: Implement Plugin Loader

**Description:** PluginLoader manages provider lifecycle: discover built-in providers, initialize, start, health check, stop. Handles crash detection with automatic restart (max 3 retries with backoff). Integrates with CapabilityRegistry. Publishes lifecycle events.

**Priority:** P0 🔴  
**Difficulty:** D4 🔴  
**Owner:** Platform  
**Dependencies:** S1-T002 (KPM), S1-T005 (CapabilityRegistry), S1-T008 (Health Manager)  
**Estimated Time:** 2 days

**Subtask S1-T009.1:** Define `PluginLoader` struct  
**Subtask S1-T009.2:** Implement built-in provider registration  
**Subtask S1-T009.3:** Implement provider lifecycle (init→start→health→stop)  
**Subtask S1-T009.4:** Implement crash detection via health check failures  
**Subtask S1-T009.5:** Implement restart with exponential backoff (1s, 5s, 15s)  
**Subtask S1-T009.6:** Enforce max retries (3) and escalate to Health Manager  
**Subtask S1-T009.7:** Publish lifecycle events (provider.registered, provider.ready, provider.crashed)  
**Subtask S1-T009.8:** Register/unregister capabilities as providers start/stop  
**Subtask S1-T009.9:** Write integration tests

**Acceptance Criteria:**
```
GIVEN a built-in provider registered with PluginLoader
WHEN Kernel boots
THEN provider.Init() → provider.Start() → health checks begin

GIVEN a provider crashes 3 times
WHEN max retries is exceeded
THEN the provider is marked "dead" and a "provider.crashed" event is published
```

---

##### Task S1-T010: Create Kernel Main Entry Point

**Description:** Wire everything together in `cmd/cloudos/main.go`: parse flags, load config, create components, register subsystems, boot, block on signals, shutdown. Target: single binary < 50MB, boot < 1s.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Dependencies:** All S1 tasks  
**Estimated Time:** 4 hours

**Subtask S1-T010.1:** Wire together KPM, Config, EventBus, Registry, PluginLoader, HealthManager  
**Subtask S1-T010.2:** Implement clean shutdown with signal handling  
**Subtask S1-T010.3:** Verify binary size (< 50MB with `go build -ldflags="-s -w"`)  
**Subtask S1-T010.4:** Verify boot time (< 1s)  
**Subtask S1-T010.5:** Write integration test (boot → health check → shutdown)

**Acceptance Criteria:**
```
GIVEN the compiled Kernel binary
WHEN I run it
THEN binary < 50MB, boot < 1s, health check returns 200, shutdown completes cleanly
```

---

## 5. Sprint 2: API & Auth

**Theme:** Expose Kernel capabilities via REST API, secure with JWT authentication and RBAC.

**Goal:** `curl localhost:8080/api/v1/health` returns 200, login returns JWT, authenticated requests work.

**Duration:** ~4 days

---

### Epic: REST API

#### Feature 2.1: HTTP Server & Middleware

**Business Value:** External systems and users interact with CloudOS through the API.  
**Technical Value:** Clean middleware architecture enables auth, logging, rate limiting, and CORS.

---

##### Task S2-T001: Implement HTTP Server

**Description:** Implement HTTP server using Go 1.24 `net/http`. Support configurable port, graceful shutdown with connection draining, middleware chain (recovery, logging, CORS, request ID, auth). Structured JSON responses.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Backend  
**Dependencies:** S1-T010 (Kernel running)  
**Estimated Time:** 6 hours

**Subtask S2-T001.1:** Create `internal/api/server.go` with `Server` struct  
**Subtask S2-T001.2:** Implement `RecoveryMiddleware` (catch panics → 500 JSON)  
**Subtask S2-T001.3:** Implement `LoggingMiddleware` (method, path, status, duration)  
**Subtask S2-T001.4:** Implement `CORS.Middleware` (configurable origins)  
**Subtask S2-T001.5:** Implement `RequestIDMiddleware` (generate if missing)  
**Subtask S2-T001.6:** Implement graceful shutdown with connection draining  
**Subtask S2-T001.7:** Create `Response` helper for consistent JSON responses  
**Subtask S2-T001.8:** Create error handler mapping (CapabilityError → HTTP status)  
**Subtask S2-T001.9:** Write integration tests

**Acceptance Criteria:**
```
GIVEN the server is running
WHEN I send a valid request
THEN I get a consistent JSON response with request_id in headers

GIVEN a handler panics
WHEN the request reaches it
THEN I get a 500 JSON response and the stack trace is logged
```

---

##### Task S2-T002: Define Route Structure

**Description:** Define all routes for v0.1 using Go 1.24 pattern matching. Group routes by prefix. Mount on server.

**Priority:** P0 🔴  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S2-T001 (HTTP Server)  
**Estimated Time:** 2 hours

**Subtask S2-T002.1:** Define route constants in `internal/api/routes.go`  
**Subtask S2-T002.2:** Implement route groups (public, authenticated, admin)  
**Subtask S2-T002.3:** Mount routes on server  
**Subtask S2-T002.4:** Write integration tests (all routes respond, unknown route = 404)

**Routes:**
```
GET    /health                          → Health check
GET    /api/v1/capabilities             → List capabilities

POST   /api/v1/auth/login               → Login
POST   /api/v1/auth/register            → Register
POST   /api/v1/auth/refresh             → Refresh token
POST   /api/v1/auth/logout              → Logout

POST   /api/v1/deploy                   → Deploy container
GET    /api/v1/containers               → List containers
GET    /api/v1/containers/{id}          → Container detail
GET    /api/v1/containers/{id}/logs     → Container logs
POST   /api/v1/containers/{id}/stop     → Stop container

POST   /api/v1/storage/buckets         → Create bucket
GET    /api/v1/storage/buckets          → List buckets
PUT    /api/v1/storage/buckets/{name}/objects/{key} → Upload object
GET    /api/v1/storage/buckets/{name}/objects/{key} → Download object

POST   /api/v1/databases               → Create database
GET    /api/v1/databases                → List databases
GET    /api/v1/databases/{id}          → Database detail
DELETE /api/v1/databases/{id}          → Delete database
```

---

##### Task S2-T003: Implement Capability Handlers

**Description:** Implement HTTP handlers that parse JSON requests, validate, call capability methods, and serialize JSON responses. Handle streaming endpoints (logs) with SSE. Map capability errors to HTTP errors.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Backend  
**Dependencies:** S2-T002 (Routes), S1-T005 (CapabilityRegistry)  
**Estimated Time:** 6 hours

**Subtask S2-T003.1:** Implement health handler (wires through to HealthManager)  
**Subtask S2-T003.2:** Implement capability list handler  
**Subtask S2-T003.3:** Implement deploy handler (validate request, call compute.RunContainer)  
**Subtask S2-T003.4:** Implement container list/get/logs handlers  
**Subtask S2-T003.5:** Implement storage bucket/object handlers  
**Subtask S2-T003.6:** Implement database CRUD handlers  
**Subtask S2-T003.7:** Implement consistent error mapping (CapabilityError → HTTP)  
**Subtask S2-T003.8:** Write integration tests for all handlers

**Acceptance Criteria:**
```
GIVEN a valid deploy request
WHEN I POST to /api/v1/deploy with {"image": "nginx:alpine", "name": "my-app"}
THEN I get a 201 response with container ID and status "creating"
```

---

### Epic: Authentication

#### Feature 2.2: JWT Authentication

**Business Value:** Users must authenticate before using CloudOS.  
**Technical Value:** JWT enables stateless auth, API keys enable service-to-service auth.

---

##### Task S2-T004: Implement Auth Engine

**Description:** Implement AuthEngine: user registration (bcrypt-hashed passwords), login (validate → issue JWT), JWT signing/validation (HMAC-SHA256 or Ed25519), token refresh, token revocation (blacklist). Store users in SQLite.

**Priority:** P0 🔴  
**Difficulty:** D4 🔴  
**Owner:** Security  
**Dependencies:** S0-T008 (Config for JWT secret), PROV-005 (SQLite)  
**Estimated Time:** 2 days

**Subtask S2-T004.1:** Define User model and SQLite storage  
**Subtask S2-T004.2:** Implement user registration with bcrypt hashing  
**Subtask S2-T004.3:** Implement JWT signing (HS256 or Ed25519)  
**Subtask S2-T004.4:** Implement JWT validation (parse, verify signature, check expiry)  
**Subtask S2-T004.5:** Implement token refresh (short-lived access + long-lived refresh)  
**Subtask S2-T004.6:** Implement token revocation blacklist (in-memory with TTL)  
**Subtask S2-T004.7:** Implement API key generation and validation  
**Subtask S2-T004.8:** Write comprehensive tests (valid, expired, revoked, malformed tokens)

**Acceptance Criteria:**
```
GIVEN a registered user with email "test@example.com" and password "securePass123!"
WHEN I POST to /api/v1/auth/login with correct credentials
THEN I receive {"access_token": "<JWT>", "refresh_token": "<token>", "expires_in": 3600}

GIVEN an expired JWT
WHEN I use it to authenticate
THEN I receive a 401 response with "token_expired"

GIVEN a revoked JWT
WHEN I use it to authenticate
THEN I receive a 401 response with "token_revoked"
```

---

##### Task S2-T005: Implement Auth Middleware

**Description:** Implement auth middleware that validates JWT from Authorization header (Bearer), extracts user identity and roles, injects into request context. Support API key authentication as alternative. Return 401 for invalid/missing tokens.

**Priority:** P0 🔴  
**Difficulty:** D2 🟡  
**Owner:** Security  
**Dependencies:** S2-T004 (Auth Engine), S2-T001 (HTTP Server)  
**Estimated Time:** 3 hours

**Subtask S2-T005.1:** Implement `AuthMiddleware` — extract Bearer token, validate via AuthEngine  
**Subtask S2-T005.2:** Implement `APIKeyMiddleware` — extract X-API-Key header, validate  
**Subtask S2-T005.3:** Inject user identity into request context  
**Subtask S2-T005.4:** Apply middleware to authenticated route group  
**Subtask S2-T005.5:** Write integration tests (no token, invalid token, valid token, API key)

**Acceptance Criteria:**
```
GIVEN no Authorization header
WHEN I call an authenticated endpoint
THEN I receive 401

GIVEN a valid JWT
WHEN I call an authenticated endpoint
THEN the handler can read user identity from context
```

---

#### Feature 2.3: RBAC Authorization

**Business Value:** Different users have different permissions. Admins manage, users deploy, readonly users view.  
**Technical Value:** Permission checks at the middleware level prevent unauthorized operations.

---

##### Task S2-T006: Implement RBAC

**Description:** Define roles (admin, user, readonly) with permission sets. Implement AuthorizationEngine with CheckPermission(user, resource, action). Implement permission middleware. Default assignment: admin=all, user=deploy+storage, readonly=view only.

**Priority:** P1 🟡  
**Difficulty:** D3 🟠  
**Owner:** Security  
**Dependencies:** S2-T005 (Auth middleware)  
**Estimated Time:** 4 hours

**Subtask S2-T006.1:** Define Role, Permission types and role→permission mapping  
**Subtask S2-T006.2:** Implement `AuthorizationEngine` with `CheckPermission(user, resource, action)`  
**Subtask S2-T006.3:** Implement `RequirePermission(resource, action)` middleware factory  
**Subtask S2-T006.4:** Apply permission middleware to all protected routes  
**Subtask S2-T006.5:** Write tests (admin allowed, user allowed, readonly denied, unauthenticated denied)

**Acceptance Criteria:**
```
GIVEN a user with role "readonly"
WHEN they call POST /api/v1/deploy
THEN they receive 403

GIVEN a user with role "admin"
WHEN they call the same endpoint
THEN the request succeeds
```

---

#### Feature 2.4: User Management

**Business Value:** Users need to manage their accounts, API keys, and profile.  
**Technical Value:** Foundation for multi-tenant organization features.

---

##### Task S2-T007: Implement User Management Endpoints

**Description:** Implement profile management: GET /profile (current user), PATCH /profile (update name), POST /auth/keys (create API key), GET /auth/keys (list), DELETE /auth/keys/{id} (revoke). Admin-only: GET /users (list all), PATCH /users/{id}/role (change role).

**Priority:** P2 🟢  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S2-T004 (Auth Engine), S2-T006 (RBAC)  
**Estimated Time:** 4 hours

**Subtask S2-T007.1:** Implement GET /api/v1/auth/profile  
**Subtask S2-T007.2:** Implement PATCH /api/v1/auth/profile  
**Subtask S2-T007.3:** Implement API key CRUD endpoints  
**Subtask S2-T007.4:** Implement admin user list and role management  
**Subtask S2-T007.5:** Write integration tests

**Acceptance Criteria:**
```
GIVEN a logged-in user
WHEN they call GET /api/v1/auth/profile
THEN they receive their user info

GIVEN an admin user
WHEN they change another user's role
THEN the role is updated
```

---

##### Task S2-T008: Create User Registration Flow

**Description:** Registration endpoint with email validation, password strength requirements (min 8 chars, mixed case), rate limiting (3 attempts per minute), email verification stub.

**Priority:** P1 🟡  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S2-T007 (User Management)  
**Estimated Time:** 3 hours

**Subtask S2-T008.1:** Implement register endpoint with input validation  
**Subtask S2-T008.2:** Add password strength validation  
**Subtask S2-T008.3:** Add rate limiting on register endpoint (3/minute/IP)  
**Subtask S2-T008.4:** Create email verification token (stub — no email sending yet)  
**Subtask S2-T008.5:** Write integration tests

**Acceptance Criteria:**
```
GIVEN valid registration data
WHEN I POST to /api/v1/auth/register
THEN a user is created and a verification token is returned

GIVEN a password "short"
WHEN I try to register
THEN I get a 400 error explaining password requirements
```

---

## 6. Sprint 3: Dashboard

**Theme:** Build the React web dashboard as the primary user interface.

**Goal:** User can log in, see system status, and deploy containers from the browser.

**Duration:** ~5 days

---

### Epic: Web Dashboard

#### Feature 3.1: Application Scaffold

**Business Value:** The dashboard is how most users interact with CloudOS.  
**Technical Value:** Well-structured React codebase enables rapid feature development.

---

##### Task S3-T001: Scaffold React Application

**Description:** Initialize React 19 + TypeScript application with Vite. Configure Tailwind CSS v4, React Router v7, and API client. Create base layout with sidebar navigation. Set up build pipeline integrated with Go binary (embedded dist).

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Frontend  
**Dependencies:** S0-T001 (Repository exists)  
**Estimated Time:** 4 hours

**Subtask S3-T001.1:** Run `npm create vite@latest apps/dashboard -- --template react-ts`  
**Subtask S3-T001.2:** Configure Tailwind CSS v4  
**Subtask S3-T001.3:** Set up React Router v7 with route definitions  
**Subtask S3-T001.4:** Create API client module (fetch wrapper with auth header injection)  
**Subtask S3-T001.5:** Create base layout component (sidebar + content area)  
**Subtask S3-T001.6:** Configure build to output to `core/cloudos/internal/dashboard/` for Go embedding  
**Subtask S3-T001.7:** Set up ESLint and Prettier for TypeScript/React

**Acceptance Criteria:**
```
GIVEN the dev server is running
WHEN I navigate to http://localhost:5173
THEN the base layout renders with sidebar navigation

GIVEN `npm run build` completes
WHEN I check the output directory
THEN static files are ready for Go embedding
```

---

#### Feature 3.2: Authentication UI

**Business Value:** Users must log in before using the dashboard.  
**Technical Value:** Token management, protected routes, and auth state must be handled correctly.

---

##### Task S3-T002: Implement Login Page

**Description:** Login form with email/password fields. Submit to `/api/v1/auth/login`. Store JWT in httpOnly cookie (set by server proxy) + in-memory variable. Display error on failure. Redirect to dashboard on success. Include link to registration.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Frontend  
**Dependencies:** S3-T001 (React scaffold), S2-T004 (Auth Engine)  
**Estimated Time:** 4 hours

**Subtask S3-T002.1:** Create LoginPage component with form  
**Subtask S3-T002.2:** Implement form validation (email format, password required)  
**Subtask S3-T002.3:** Implement API call to login endpoint  
**Subtask S3-T002.4:** Implement JWT storage (memory, not localStorage)  
**Subtask S3-T002.5:** Implement auth context/state management  
**Subtask S3-T002.6:** Implement protected route wrapper (redirect to login if unauthenticated)  
**Subtask S3-T002.7:** Create RegisterPage with registration form  
**Subtask S3-T002.8:** Style with Tailwind — clean, professional, accessible

**Acceptance Criteria:**
```
GIVEN I am on the login page
WHEN I enter valid credentials and submit
THEN I am redirected to the dashboard

GIVEN I am on the login page
WHEN I enter invalid credentials
THEN an error message is displayed

GIVEN I am not logged in
WHEN I navigate to /dashboard
THEN I am redirected to /login
```

---

#### Feature 3.3: System Status

**Business Value:** Users need to know if CloudOS is healthy.  
**Technical Value:** Demonstrates real-time API integration.

---

##### Task S3-T003: Implement Status Page

**Description:** System status dashboard: overall health indicator (green/yellow/red), subsystem health cards (name, status, uptime), capability list (name, version, provider). Auto-refresh every 10 seconds.

**Priority:** P1 🟡  
**Difficulty:** D2 🟡  
**Owner:** Frontend  
**Dependencies:** S3-T001 (React scaffold), S2-T003 (Health API)  
**Estimated Time:** 3 hours

**Subtask S3-T003.1:** Create StatusPage component  
**Subtask S3-T003.2:** Implement health API call with auto-refresh  
**Subtask S3-T003.3:** Create health indicator component (green/yellow/red circle)  
**Subtask S3-T003.4:** Create subsystem health card component  
**Subtask S3-T003.5:** Create capability list component  
**Subtask S3-T003.6:** Implement loading state (skeleton) and error state  
**Subtask S3-T003.7:** Style with Tailwind

**Acceptance Criteria:**
```
GIVEN the dashboard is loaded
WHEN I navigate to /status
THEN I see system health, subsystem cards, and capability list

GIVEN a subsystem becomes unhealthy
WHEN the next auto-refresh completes
THEN the health indicator updates
```

---

#### Feature 3.4: Deploy UI

**Business Value:** Users deploy applications from the dashboard.  
**Technical Value:** Demonstrates capability-to-UI mapping end-to-end.

---

##### Task S3-T004: Implement Deploy Page

**Description:** Deploy form: container image input, optional env vars (key-value editor), port mapping, resource limits (add/remove). Submit to POST /api/v1/deploy. Show deployment progress (creating → running → failed). List running containers with status, uptime, resource usage. Container detail view with log streaming.

**Priority:** P1 🟡  
**Difficulty:** D4 🔴  
**Owner:** Frontend  
**Dependencies:** S3-T001 (React scaffold), S2-T003 (Deploy API)  
**Estimated Time:** 2 days

**Subtask S3-T004.1:** Create DeployForm component with validation  
**Subtask S3-T004.2:** Create environment variable key-value editor  
**Subtask S3-T004.3:** Create port mapping input  
**Subtask S3-T004.4:** Implement deploy API call with progress display  
**Subtask S3-T004.5:** Create ContainerList component  
**Subtask S3-T004.6:** Create ContainerDetail component with log streaming  
**Subtask S3-T004.7:** Implement log streaming via SSE  
**Subtask S3-T004.8:** Style with Tailwind

**Acceptance Criteria:**
```
GIVEN the deploy page
WHEN I fill in "nginx:alpine" as image and click Deploy
THEN I see progress: "creating" → "running"
AND the container appears in the container list

GIVEN a running container
WHEN I click on it
THEN I see container details and streaming logs
```

---

##### Task S3-T005: Implement Storage Page

**Description:** Storage browser: list buckets, click → list objects, upload file (drag-and-drop + file picker), download file, delete file. Show storage usage.

**Priority:** P2 🟢  
**Difficulty:** D3 🟠  
**Owner:** Frontend  
**Dependencies:** S3-T001 (React scaffold), S2-T003 (Storage API)  
**Estimated Time:** 4 hours

**Subtask S3-T005.1:** Create BucketList component  
**Subtask S3-T005.2:** Create ObjectList component  
**Subtask S3-T005.3:** Implement file upload with drag-and-drop  
**Subtask S3-T005.4:** Implement file download  
**Subtask S3-T005.5:** Implement delete with confirmation dialog  
**Subtask S3-T005.6:** Display storage usage metrics  
**Subtask S3-T005.7:** Style with Tailwind

**Acceptance Criteria:**
```
GIVEN the storage page
WHEN I create a bucket and upload a file via drag-and-drop
THEN the file appears in the object list and can be downloaded
```

---

##### Task S3-T006: Implement Database Page

**Description:** Database management: create database (name, engine selector), list databases, view database detail (connection string with copy button), delete database (with confirmation).

**Priority:** P2 🟢  
**Difficulty:** D2 🟡  
**Owner:** Frontend  
**Dependencies:** S3-T001 (React scaffold), S2-T003 (Database API)  
**Estimated Time:** 3 hours

**Subtask S3-T006.1:** Create DatabaseList component  
**Subtask S3-T006.2:** Create CreateDatabase form  
**Subtask S3-T006.3:** Create DatabaseDetail component with connection string  
**Subtask S3-T006.4:** Implement clipboard copy for connection string  
**Subtask S3-T006.5:** Implement delete with confirmation  
**Subtask S3-T006.6:** Style with Tailwind

**Acceptance Criteria:**
```
GIVEN the database page
WHEN I create a new SQLite database
THEN it appears in the list with a copyable connection string
```

---

## 7. Sprint 4: Storage, Database & Secrets

**Theme:** Implement the built-in storage, database, and secrets providers.

**Goal:** Files can be uploaded and stored, databases can be created and connected, secrets are encrypted at rest.

**Duration:** ~4 days

---

### Epic: Storage

#### Feature 4.1: Local Storage Provider

**Business Value:** Users need to store files (deployment artifacts, backups, assets).  
**Technical Value:** Proves the storage capability interface with a simple filesystem implementation.

---

##### Task S4-T001: Implement storage.local Provider

**Description:** Built-in storage provider storing objects on local filesystem at `~/.cloudos/data/storage/`. Supports CreateBucket, PutObject, GetObject, DeleteObject, ListObjects. Metadata stored as JSON sidecar. Path traversal prevention.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Backend  
**Dependencies:** S1-T009 (Plugin Loader)  
**Estimated Time:** 6 hours

**Subtask S4-T001.1:** Create `internal/providers/storage/local/provider.go`  
**Subtask S4-T001.2:** Implement Provider interface (Init, Start, Stop, Health, GetCapabilities)  
**Subtask S4-T001.3:** Implement StorageCapability interface methods  
**Subtask S4-T001.4:** Implement file-based storage (bucket = directory, object = file)  
**Subtask S4-T001.5:** Implement metadata JSON sidecar files  
**Subtask S4-T001.6:** Implement path traversal prevention (validate all paths)  
**Subtask S4-T001.7:** Register as built-in provider in PluginLoader  
**Subtask S4-T001.8:** Write unit and integration tests

**Acceptance Criteria:**
```
GIVEN a local storage provider
WHEN I create a bucket "my-app" and put object "index.html"
THEN a file exists at ~/.cloudos/data/storage/my-app/index.html
AND a sidecar at ~/.cloudos/data/storage/my-app/.meta.index.html.json

GIVEN a path traversal attempt in the key "../../etc/passwd"
WHEN I call PutObject
THEN I get an ErrAccessDenied error
```

---

##### Task S4-T002: Implement Storage API Handlers

**Description:** API endpoints for storage operations. File upload via multipart/form-data. File download with correct Content-Type from metadata. List buckets/objects.

**Priority:** P1 🟡  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S2-T003 (API Handlers), S4-T001 (storage.local)  
**Estimated Time:** 3 hours

**Subtask S4-T002.1:** Implement CreateBucket handler  
**Subtask S4-T002.2:** Implement ListBuckets handler  
**Subtask S4-T002.3:** Implement PutObject handler (multipart upload)  
**Subtask S4-T002.4:** Implement GetObject handler (stream with Content-Type)  
**Subtask S4-T002.5:** Implement ListObjects handler  
**Subtask S4-T002.6:** Implement DeleteObject handler  
**Subtask S4-T002.7:** Write integration tests

**Acceptance Criteria:**
```
GIVEN the API server
WHEN I PUT a file to /api/v1/storage/buckets/my-app/objects/test.txt
THEN the file is stored and GET returns it with correct Content-Type
```

---

#### Feature 4.2: Storage Quotas

**Business Value:** Prevent a single user from consuming all disk space.  
**Technical Value:** Quota enforcement is a cross-cutting concern demonstrated here.

---

##### Task S4-T003: Implement Storage Quotas

**Description:** Per-user storage quotas (configurable default: 1GB). Quota check on upload. Quota exceeded error. Usage tracking. Display in dashboard.

**Priority:** P2 🟢  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S4-T001 (storage.local)  
**Estimated Time:** 3 hours

**Subtask S4-T003.1:** Add quota configuration to provider config  
**Subtask S4-T003.2:** Implement usage tracking (sum file sizes per bucket per user)  
**Subtask S4-T003.3:** Implement quota check before PutObject  
**Subtask S4-T003.4:** Return QuotaExceeded error with current/max display  
**Subtask S4-T003.5:** Write tests (under quota, at quota, over quota)

**Acceptance Criteria:**
```
GIVEN a user with 1GB quota and 900MB used
WHEN they upload a 200MB file
THEN they get a quota exceeded error

GIVEN the same user
WHEN they upload a 50MB file
THEN the upload succeeds
```

---

### Epic: Database

#### Feature 4.3: SQLite Database Provider

**Business Value:** Users need databases for their applications.  
**Technical Value:** Proves the database capability interface.

---

##### Task S4-T004: Implement database.sqlite Provider

**Description:** Built-in SQLite provider using `modernc.org/sqlite` (pure Go, no CGO). Supports Create (database file), Get (info), Delete (remove file), List (all databases), GetConnectionString, CreateBackup (file copy), RestoreBackup.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Backend  
**Dependencies:** S1-T009 (Plugin Loader)  
**Estimated Time:** 6 hours

**Subtask S4-T004.1:** Create `internal/providers/database/sqlite/provider.go`  
**Subtask S4-T004.2:** Add `modernc.org/sqlite` dependency to go.mod  
**Subtask S4-T004.3:** Implement Provider interface  
**Subtask S4-T004.4:** Implement DatabaseCapability interface  
**Subtask S4-T004.5:** Implement backup as atomic file copy  
**Subtask S4-T004.6:** Implement restore from backup file  
**Subtask S4-T004.7:** Register as built-in provider in PluginLoader  
**Subtask S4-T004.8:** Write unit and integration tests

**Acceptance Criteria:**
```
GIVEN a SQLite provider
WHEN I create a database "my-app-db"
THEN a SQLite file exists at ~/.cloudos/data/databases/my-app-db.db
AND GetConnectionString returns a valid sqlite:// path

GIVEN a database with data
WHEN I create a backup and restore it to a new database
THEN the new database contains all the data from the backup
```

---

##### Task S4-T005: Implement Database API Handlers

**Description:** API endpoints for database operations. Create, list, get detail (with connection string), delete. Connection string has copy-to-clipboard support in dashboard.

**Priority:** P1 🟡  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S2-T003 (API Handlers), S4-T004 (database.sqlite)  
**Estimated Time:** 3 hours

**Subtask S4-T005.1:** Implement CreateDatabase handler  
**Subtask S4-T005.2:** Implement ListDatabases handler  
**Subtask S4-T005.3:** Implement GetDatabase handler (with connection string)  
**Subtask S4-T005.4:** Implement DeleteDatabase handler  
**Subtask S4-T005.5:** Implement CreateBackup / ListBackups / RestoreBackup handlers  
**Subtask S4-T005.6:** Write integration tests

**Acceptance Criteria:**
```
GIVEN the API server
WHEN I POST to /api/v1/databases with {"name": "my-db", "engine": "sqlite"}
THEN I receive a database info object with a connection string

GIVEN an existing database
WHEN I GET /api/v1/databases/{id}
THEN I receive the database details including the connection string
```

---

### Epic: Secrets

#### Feature 4.4: Secrets Manager

**Business Value:** API keys, passwords, and tokens must be stored securely.  
**Technical Value:** Providers need runtime access to secrets without exposing them in config files.

---

##### Task S4-T006: Implement Secrets Manager

**Description:** Encrypted secrets storage using AES-256-GCM. Secrets are stored in SQLite with encryption at rest. Secrets API: Set, Get, Delete, List, Rotate. Master key from environment variable (`CLOUDOS_MASTER_KEY`) or auto-generated file (`~/.cloudos/master.key`). Access control per provider.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Security  
**Dependencies:** S0-T008 (Config), S4-T004 (SQLite)  
**Estimated Time:** 6 hours

**Subtask S4-T006.1:** Define `SecretsManager` interface (Get, Set, Delete, List)  
**Subtask S4-T006.2:** Implement AES-256-GCM encryption/decryption  
**Subtask S4-T006.3:** Implement master key loading (env var, file, generate)  
**Subtask S4-T006.4:** Implement SQLite-backed secrets storage  
**Subtask S4-T006.5:** Implement access control per secret path  
**Subtask S4-T006.6:** Implement `SecretReader` interface for providers  
**Subtask S4-T006.7:** Write tests (encrypt/decrypt, persistence, access control)

**Acceptance Criteria:**
```
GIVEN a SecretsManager with master key
WHEN I store a secret at path "providers/ai.openai/api_key"
THEN the secret is encrypted at rest in SQLite
AND reading it returns the plaintext value

GIVEN a provider with access to "providers/ai.openai/*"
WHEN it tries to read "providers/ai.openai/api_key"
THEN it succeeds

GIVEN a provider WITHOUT access to "providers/ai.openai/*"
WHEN it tries to read "providers/ai.openai/api_key"
THEN it gets an access denied error

GIVEN the SQLite database is read directly
WHEN I examine the secrets table
THEN the values are encrypted and unreadable
```

---

##### Task S4-T007: Integrate Secrets with Config

**Description:** Config system recognizes `secret://` prefix in config values and resolves them through SecretsManager. Providers receive `SecretReader` in their `ProviderConfig` at initialization.

**Priority:** P1 🟡  
**Difficulty:** D2 🟡  
**Owner:** Platform  
**Dependencies:** S4-T006 (Secrets Manager), S0-T008 (Config)  
**Estimated Time:** 3 hours

**Subtask S4-T007.1:** Add `secret://` URL parser to config system  
**Subtask S4-T007.2:** Implement secret resolution during config loading  
**Subtask S4-T007.3:** Inject `SecretReader` into `ProviderConfig`  
**Subtask S4-T007.4:** Ensure secrets are never logged (mask in all output)  
**Subtask S4-T007.5:** Write integration tests

**Acceptance Criteria:**
```
GIVEN a config value "secret://providers/storage.s3/access_key"
WHEN the config is loaded
THEN the value is resolved through SecretsManager
AND the plaintext is NEVER written to logs
```

---

##### Task S4-T008: Implement Secrets API Endpoints

**Description:** REST endpoints for secrets management: `POST /api/v1/secrets` (create), `GET /api/v1/secrets` (list paths), `GET /api/v1/secrets/{path}` (read), `DELETE /api/v1/secrets/{path}` (delete), `POST /api/v1/secrets/rotate` (rotate master key). Admin-only access.

**Priority:** P2 🟢  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S4-T006 (Secrets Manager), S2-T006 (RBAC)  
**Estimated Time:** 3 hours

**Subtask S4-T008.1:** Implement CreateSecret handler  
**Subtask S4-T008.2:** Implement ListSecrets handler (returns paths only, never values)  
**Subtask S4-T008.3:** Implement GetSecret handler  
**Subtask S4-T008.4:** Implement DeleteSecret handler  
**Subtask S4-T008.5:** Implement RotateMasterKey handler  
**Subtask S4-T008.6:** Apply admin-only RBAC to all secrets endpoints  
**Subtask S4-T008.7:** Write integration tests

**Acceptance Criteria:**
```
GIVEN an admin user
WHEN they POST to /api/v1/secrets with {"path": "my-key", "value": "my-value"}
THEN the secret is stored

GIVEN a non-admin user
WHEN they call any secrets endpoint
THEN they receive 403
```

---

## 8. Sprint 5: Deployments & Networking

**Theme:** Implement the compute deployment flow and basic networking (DNS, SSL).

**Goal:** User can deploy a container from an image, access it via a subdomain with automatic SSL.

**Duration:** ~4 days

---

### Epic: Compute

#### Feature 5.1: Local Compute Provider

**Business Value:** Users deploy and run applications on CloudOS.  
**Technical Value:** Proves the compute capability interface with process execution.

---

##### Task S5-T001: Implement compute.local Provider

**Description:** Built-in compute provider that runs containers as local OS processes (not Docker — direct process execution). Supports RunContainer (exec image or command), StopContainer (signal), GetContainer (process info), ListContainers, GetContainerLogs (stdout/stderr), GetContainerMetrics (CPU/memory from /proc). Resource limits via OS primitives.

**Priority:** P0 🔴  
**Difficulty:** D4 🔴  
**Owner:** Backend  
**Dependencies:** S1-T009 (Plugin Loader), S4-T001 (Storage — for image pull)  
**Estimated Time:** 2 days

**Subtask S5-T001.1:** Create `internal/providers/compute/local/provider.go`  
**Subtask S5-T001.2:** Implement Provider interface  
**Subtask S5-T001.3:** Implement RunContainer (create process with resource limits)  
**Subtask S5-T001.4:** Implement StopContainer (SIGTERM → SIGKILL after timeout)  
**Subtask S5-T001.5:** Implement GetContainer / ListContainers  
**Subtask S5-T001.6:** Implement GetContainerLogs (capture stdout/stderr)  
**Subtask S5-T001.7:** Implement GetContainerMetrics (parse /proc/[pid]/stat, /proc/[pid]/status)  
**Subtask S5-T001.8:** Implement resource limits (RLIMIT_CPU, RLIMIT_AS, RLIMIT_NOFILE)  
**Subtask S5-T001.9:** Register as built-in provider in PluginLoader  
**Subtask S5-T001.10:** Write tests (run, stop, logs, metrics, resource limits, concurrent)

**Acceptance Criteria:**
```
GIVEN a local compute provider
WHEN I RunContainer with image "nginx:alpine" and command ["nginx", "-g", "daemon off;"]
THEN a process is started, container ID is returned, status is "running"

GIVEN a running container
WHEN I GetContainerLogs with follow=true
THEN I receive log output

GIVEN a container exceeding its memory limit
WHEN it allocates too much memory
THEN the process is killed and container status becomes "failed"
```

---

#### Feature 5.2: Deployment Flow

**Business Value:** End-to-end deployment from user request to running container.  
**Technical Value:** Orchestrates multiple capabilities (storage + compute) for a single user intent.

---

##### Task S5-T002: Implement Deploy Endpoint

**Description:** The deploy endpoint orchestrates the full deploy flow: validate request → pull/expand image → create container → return access URL. Uses storage capability for image assets and compute capability for running.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Backend  
**Dependencies:** S2-T003 (API), S5-T001 (compute.local), S4-T001 (storage.local)  
**Estimated Time:** 6 hours

**Subtask S5-T002.1:** Implement deploy handler that orchestrates storage + compute  
**Subtask S5-T002.2:** Implement image name parsing (registry, name, tag)  
**Subtask S5-T002.3:** Implement basic image pulling (Docker Hub API)  
**Subtask S5-T002.4:** Assign port mapping and return access URL  
**Subtask S5-T002.5:** Write integration tests (deploy → verify running → stop)

**Acceptance Criteria:**
```
GIVEN a valid deploy request with image "nginx:alpine"
WHEN I POST to /api/v1/deploy
THEN:
  1. Container is created and running
  2. Response includes container ID, status "running", and access URL
  3. GET /api/v1/containers/{id} returns the container info

GIVEN a running container
WHEN I POST /api/v1/containers/{id}/stop
THEN the container is stopped and status becomes "stopped"
```

---

#### Feature 5.3: Container Management

**Business Value:** Users need to see what's running and manage containers.  
**Technical Value:** Demonstrates the full compute lifecycle API.

---

##### Task S5-T003: Implement Container Management Endpoints

**Description:** List all containers (with pagination), get container detail (config, status, resources, timestamps), stream container logs (SSE), stop container (graceful + force), remove container.

**Priority:** P1 🟡  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S2-T003 (API), S5-T001 (compute.local)  
**Estimated Time:** 4 hours

**Subtask S5-T003.1:** Implement ListContainers handler (with status filter, pagination)  
**Subtask S5-T003.2:** Implement GetContainer handler (full detail)  
**Subtask S5-T003.3:** Implement streaming logs via SSE  
**Subtask S5-T003.4:** Implement StopContainer handler (graceful=30s, force=SIGKILL)  
**Subtask S5-T003.5:** Implement RemoveContainer handler  
**Subtask S5-T003.6:** Write integration tests

**Acceptance Criteria:**
```
GIVEN 3 containers in different states
WHEN I GET /api/v1/containers
THEN I get a paginated list with status, uptime, and resource usage

GIVEN a running container
WHEN I stream GET /api/v1/containers/{id}/logs
THEN I receive SSE events with log lines
```

---

### Epic: Networking

#### Feature 5.4: DNS & SSL

**Business Value:** Deployed applications need accessible URLs with HTTPS.  
**Technical Value:** Auto-provisioning subdomains and SSL certificates enables zero-config deployment.

---

##### Task S5-T004: Implement Built-in DNS Provider

**Description:** Simple built-in DNS server for `.cloudos.local` domains. Maps `<container-name>.cloudos.local` to assigned port. Uses hosts file or embedded DNS server.

**Priority:** P1 🟡  
**Difficulty:** D3 🟠  
**Owner:** Backend  
**Dependencies:** S5-T001 (compute.local)  
**Estimated Time:** 4 hours

**Subtask S5-T004.1:** Create `internal/providers/dns/builtin/provider.go`  
**Subtask S5-T004.2:** Implement Provider interface  
**Subtask S5-T004.3:** Implement DNSCapability interface (CreateRecord, ListRecords, DeleteRecord)  
**Subtask S5-T004.4:** Implement DNS record → port mapping  
**Subtask S5-T004.5:** Register as built-in provider  
**Subtask S5-T004.6:** Write integration tests

**Acceptance Criteria:**
```
GIVEN a container "my-app" running on port 45678
WHEN I query DNS for "my-app.cloudos.local"
THEN it resolves to the host's IP address with port 45678
```

---

##### Task S5-T005: Implement SSL Certificate Provisioning

**Description:** Auto-provision SSL certificates using Let's Encrypt (via ACME client). Generate self-signed certificates for development. Store certificates in Secrets Manager. Auto-renew before expiry.

**Priority:** P2 🟢  
**Difficulty:** D3 🟠  
**Owner:** Security  
**Dependencies:** S5-T004 (DNS — for domain verification)  
**Estimated Time:** 6 hours

**Subtask S5-T005.1:** Implement ACME client for Let's Encrypt  
**Subtask S5-T005.2:** Implement HTTP-01 challenge responder  
**Subtask S5-T005.3:** Store certificates in Secrets Manager  
**Subtask S5-T005.4:** Implement auto-renewal check (daily)  
**Subtask S5-T005.5:** Generate self-signed cert for development  
**Subtask S5-T005.6:** Write tests (cert generation, renewal trigger)

**Acceptance Criteria:**
```
GIVEN a domain "my-app.cloudos.local"
WHEN I request SSL provisioning
THEN a Let's Encrypt certificate is obtained and stored in Secrets Manager

GIVEN a certificate expiring in < 30 days
WHEN the auto-renewal check runs
THEN a new certificate is requested and stored
```

---

##### Task S5-T006: Implement Port Management

**Description:** Port allocation, conflict detection, and mapping. Assign available host ports when deploying containers. Release ports on container stop. Configuration for port range.

**Priority:** P1 🟡  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S5-T001 (compute.local)  
**Estimated Time:** 3 hours

**Subtask S5-T006.1:** Implement port allocator (find free port in range 40000-50000)  
**Subtask S5-T006.2:** Implement port release on container stop  
**Subtask S5-T006.3:** Implement port conflict detection  
**Subtask S5-T006.4:** Return assigned ports in deploy response  
**Subtask S5-T006.5:** Write tests (allocation, conflict, release, concurrent)

**Acceptance Criteria:**
```
GIVEN 1000 concurrent deploy requests
WHEN all complete
THEN no two containers have the same host port assignment
```

---

##### Task S5-T007: Implement Networking Capability Stub

**Description:** Define NetworkingCapability interface and create a built-in networking provider stub. Wire into CapabilityRegistry. No advanced features yet (firewall rules come later).

**Priority:** P3 🔵  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S1-T005 (CapabilityRegistry)  
**Estimated Time:** 2 hours

**Subtask S5-T007.1:** Define NetworkingCapability interface with basic methods  
**Subtask S5-T007.2:** Create built-in networking provider stub  
**Subtask S5-T007.3:** Register in CapabilityRegistry  
**Subtask S5-T007.4:** Write minimal tests

**Acceptance Criteria:**
```
GIVEN the networking capability is registered
WHEN I list capabilities
THEN capability.networking appears with version 0.1.0
```

---

## 9. Sprint 6: AI Engine

**Theme:** Integrate AI capabilities — chat completion, intent parsing, and basic task planning.

**Goal:** User can send a chat message and get an AI response. AI can parse simple intents like "deploy my app".

**Duration:** ~4 days

---

### Epic: AI Integration

#### Feature 6.1: AI Capability Interface

**Business Value:** AI is the primary interface for CloudOS — natural language operations.  
**Technical Value:** The AI capability follows the same Capability-Provider pattern as everything else.

---

##### Task S6-T001: Define AI Capability Interface

**Description:** Define `AICapability` interface: ChatCompletion, ChatCompletionStream, GenerateEmbedding, ListModels. Define request/response types. Register capability.

**Priority:** P1 🟡  
**Difficulty:** D2 🟡  
**Owner:** AI  
**Dependencies:** S1-T005 (CapabilityRegistry)  
**Estimated Time:** 3 hours

**Subtask S6-T001.1:** Define AICapability interface in `internal/capability/ai.go`  
**Subtask S6-T001.2:** Define ChatCompletionRequest/Response, Streaming types  
**Subtask S6-T001.3:** Define ModelInfo and ListModelsResponse  
**Subtask S6-T001.4:** Register in CapabilityRegistry  
**Subtask S6-T001.5:** Write interface tests with mock provider

**Acceptance Criteria:**
```
GIVEN the AI capability interface
WHEN implemented by a mock
THEN ChatCompletion returns a valid response
AND ListModels returns available models
```

---

##### Task S6-T002: Implement OpenAI Provider

**Description:** Implement `ai.openai` provider: HTTP client to OpenAI API, ChatCompletion (POST /v1/chat/completions), ChatCompletionStream (SSE), ListModels (GET /v1/models). API key from Secrets Manager. Configurable model, temperature, max tokens. Error handling for rate limits, timeouts, auth failures.

**Priority:** P1 🟡  
**Difficulty:** D3 🟠  
**Owner:** AI  
**Dependencies:** S6-T001 (AI interface), S4-T006 (Secrets Manager)  
**Estimated Time:** 6 hours

**Subtask S6-T002.1:** Create `internal/providers/ai/openai/provider.go`  
**Subtask S6-T002.2:** Implement OpenAI HTTP client  
**Subtask S6-T002.3:** Implement ChatCompletion (non-streaming)  
**Subtask S6-T002.4:** Implement ChatCompletionStream (SSE parsing)  
**Subtask S6-T002.5:** Implement ListModels  
**Subtask S6-T002.6:** Read API key from Secrets Manager  
**Subtask S6-T002.7:** Implement circuit breaker for rate limits  
**Subtask S6-T002.8:** Write tests with mock HTTP server

**Acceptance Criteria:**
```
GIVEN a valid OpenAI API key in Secrets Manager
WHEN I call ChatCompletion with "Hello"
THEN I get a response with content and usage statistics

GIVEN a streaming request
WHEN I call ChatCompletionStream
THEN I receive chunks that assemble into the full response
```

---

#### Feature 6.2: Chat API

**Business Value:** Expose AI capabilities through the REST API for dashboard and CLI.  
**Technical Value:** Streaming responses require SSE support.

---

##### Task S6-T003: Implement Chat API Endpoints

**Description:** `POST /api/v1/ai/chat` (non-streaming completion), `POST /api/v1/ai/chat/stream` (SSE streaming), `GET /api/v1/ai/models` (list models). Auth required. Input validation. Error handling.

**Priority:** P1 🟡  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S2-T003 (API), S6-T002 (OpenAI provider)  
**Estimated Time:** 4 hours

**Subtask S6-T003.1:** Implement ChatCompletion handler (non-streaming)  
**Subtask S6-T003.2:** Implement ChatStream handler (SSE)  
**Subtask S6-T003.3:** Implement ListModels handler  
**Subtask S6-T003.4:** Add request validation (message required, model validation)  
**Subtask S6-T003.5:** Add auth middleware to AI endpoints  
**Subtask S6-T003.6:** Write integration tests

**Acceptance Criteria:**
```
GIVEN a valid chat request
WHEN I POST to /api/v1/ai/chat with {"message": "Hello", "model": "gpt-4o-mini"}
THEN I receive a JSON response with the AI's reply

GIVEN a streaming request
WHEN I POST to /api/v1/ai/chat/stream
THEN I receive SSE events with content chunks
```

---

#### Feature 6.3: Intent Parser

**Business Value:** Natural language is the primary interface for CloudOS. Users say what they want, not which API to call.  
**Technical Value:** Intent parsing bridges human language and capability calls.

---

##### Task S6-T004: Implement Basic Intent Parser

**Description:** Parse natural language intents using template matching + LLM: "deploy nginx" → intent=deploy, image=nginx. "show my containers" → intent=list_containers. "create a database" → intent=create_database. Use the AI provider for complex parsing. Maintain conversation context.

**Priority:** P2 🟢  
**Difficulty:** D4 🔴  
**Owner:** AI  
**Dependencies:** S6-T002 (OpenAI provider)  
**Estimated Time:** 2 days

**Subtask S6-T004.1:** Define Intent types (Deploy, ListContainers, GetStatus, CreateDatabase, etc.)  
**Subtask S6-T004.2:** Implement keyword-based intent matcher for simple commands  
**Subtask S6-T004.3:** Implement LLM-based intent parser for complex commands  
**Subtask S6-T004.4:** Extract parameters from parsed intent (image name, database name, etc.)  
**Subtask S6-T004.5:** Implement conversation context management  
**Subtask S6-T004.6:** Write tests (known intents, unknown intents, parameter extraction)

**Acceptance Criteria:**
```
GIVEN the phrase "deploy nginx:alpine"
WHEN parsed by the intent parser
THEN intent=deploy, image="nginx:alpine"

GIVEN the phrase "show me my containers"
WHEN parsed
THEN intent=list_containers

GIVEN the phrase "what's the weather"
WHEN parsed
THEN intent=unknown (gracefully handled)
```

---

#### Feature 6.4: Task Engine

**Business Value:** The AI should be able to execute multi-step plans, not just respond to chat.  
**Technical Value:** Task engine bridges intent parsing and capability execution.

---

##### Task S6-T005: Implement Task Engine

**Description:** Task engine receives parsed intents, generates execution plans (sequences of capability calls), executes them, and reports results. Supports: single-step tasks (deploy container), multi-step tasks (deploy with database).

**Priority:** P2 🟢  
**Difficulty:** D4 🔴  
**Owner:** AI  
**Dependencies:** S6-T004 (Intent Parser), S1-T005 (CapabilityRegistry)  
**Estimated Time:** 2 days

**Subtask S6-T005.1:** Define Task, Plan, Step types  
**Subtask S6-T005.2:** Implement intent→plan mapping (which capabilities to call)  
**Subtask S6-T005.3:** Implement plan executor (step-by-step with error handling)  
**Subtask S6-T005.4:** Implement result reporter (success/failure with details)  
**Subtask S6-T005.5:** Write tests (single step, multi step, error recovery)

**Acceptance Criteria:**
```
GIVEN an intent to "deploy nginx"
WHEN the task engine executes
THEN it calls compute.RunContainer with "nginx:alpine"
AND returns success with container ID

GIVEN an intent to deploy with a database
WHEN the task engine executes
THEN it creates a database first, then deploys the container with the connection string
```

---

##### Task S6-T006: Implement Dashboard Chat UI

**Description:** Chat interface with message history, streaming response display (markdown rendering), model selector, and conversation persistence (stored in SQLite).

**Priority:** P2 🟢  
**Difficulty:** D3 🟠  
**Owner:** Frontend  
**Dependencies:** S3-T001 (React scaffold), S6-T003 (Chat API)  
**Estimated Time:** 6 hours

**Subtask S6-T006.1:** Create ChatPage component with message list and input  
**Subtask S6-T006.2:** Implement SSE streaming display (markdown rendered)  
**Subtask S6-T006.3:** Add model selector dropdown  
**Subtask S6-T006.4:** Persist conversation history via API  
**Subtask S6-T006.5:** Style with Tailwind

**Acceptance Criteria:**
```
GIVEN the chat page
WHEN I type "hello" and send
THEN the AI response appears in the chat with markdown rendered

GIVEN a conversation
WHEN I refresh the page
THEN the conversation history is restored
```

---

## 10. Sprint 7: Marketplace & Templates

**Theme:** Build the plugin marketplace and template system.

**Goal:** Users can browse and install plugins from a registry, deploy from templates.

**Duration:** ~4 days

---

### Epic: Marketplace

#### Feature 7.1: Plugin Registry (Stub)

**Business Value:** The plugin ecosystem is what makes CloudOS extensible.  
**Technical Value:** Registry service enables discovery and distribution of community providers.

---

##### Task S7-T001: Implement Plugin Registry API

**Description:** Local plugin registry for discovering and managing plugins. List available plugins (from local store), view plugin details, get installation status. Marketplace backend (remote registry) comes later.

**Priority:** P2 🟢  
**Difficulty:** D3 🟠  
**Owner:** Backend  
**Dependencies:** S1-T009 (Plugin Loader)  
**Estimated Time:** 6 hours

**Subtask S7-T001.1:** Define Plugin model and local storage  
**Subtask S7-T001.2:** Implement PluginRegistry: List, Get, Search, Install, Uninstall  
**Subtask S7-T001.3:** Implement local plugin store (file-based, ~/.cloudos/plugins/)  
**Subtask S7-T001.4:** Define PluginManifest struct matching `.cosp` manifest format  
**Subtask S7-T001.5:** Implement InstallPlugin flow (download → verify → extract → register)  
**Subtask S7-T001.6:** Implement UninstallPlugin flow (deregister → remove files)  
**Subtask S7-T001.7:** Write integration tests

**Acceptance Criteria:**
```
GIVEN a .cosp file in the local store
WHEN I call ListPlugins
THEN it appears in the list with name, version, author, and capabilities

GIVEN a valid .cosp file
WHEN I call InstallPlugin
THEN the plugin is extracted, verified, and registered in PluginLoader
```

---

##### Task S7-T002: Implement Marketplace API Endpoints

**Description:** REST endpoints for plugin management: `GET /api/v1/marketplace/plugins` (list), `GET /api/v1/marketplace/plugins/{name}` (detail), `POST /api/v1/marketplace/plugins/{name}/install` (install), `POST /api/v1/marketplace/plugins/{name}/uninstall` (uninstall), `POST /api/v1/marketplace/plugins/{name}/update` (update).

**Priority:** P2 🟢  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S7-T001 (Plugin Registry), S2-T003 (API)  
**Estimated Time:** 3 hours

**Subtask S7-T002.1:** Implement ListPlugins handler  
**Subtask S7-T002.2:** Implement GetPlugin handler  
**Subtask S7-T002.3:** Implement InstallPlugin handler  
**Subtask S7-T002.4:** Implement UninstallPlugin handler  
**Subtask S7-T002.5:** Implement UpdatePlugin handler  
**Subtask S7-T002.6:** Write integration tests

**Acceptance Criteria:**
```
GIVEN available plugins in the registry
WHEN I GET /api/v1/marketplace/plugins
THEN I receive a paginated list with name, version, author, description
```

---

#### Feature 7.2: Template System

**Business Value:** Templates enable one-click deployment of common application stacks.  
**Technical Value:** Template definitions are a form of infrastructure-as-code.

---

##### Task S7-T003: Implement Template Engine

**Description:** Template engine that defines deployable application stacks. Template = YAML file describing containers, databases, environment variables, and dependencies. Built-in templates: "Static Site", "Node.js App", "Python API", "WordPress", "Laravel".

**Priority:** P2 🟢  
**Difficulty:** D3 🟠  
**Owner:** Backend  
**Dependencies:** S5-T002 (Deploy)  
**Estimated Time:** 6 hours

**Subtask S7-T003.1:** Define Template YAML schema (containers, databases, env, ports)  
**Subtask S7-T003.2:** Implement template parser and validator  
**Subtask S7-T003.3:** Implement template executor (create resources per template)  
**Subtask S7-T003.4:** Create built-in templates: "static-site", "node-app", "python-api", "wordpress"  
**Subtask S7-T003.5:** Implement template variable substitution (user-provided values)  
**Subtask S7-T003.6:** Write integration tests

**Acceptance Criteria:**
```
GIVEN a "static-site" template
WHEN applied with variable "SITE_NAME=my-site"
THEN a storage bucket is created and a compute container serves the site

GIVEN a "wordpress" template
WHEN applied
THEN a database is created and a container runs WordPress connected to it
```

---

##### Task S7-T004: Implement Template API Endpoints

**Description:** `GET /api/v1/templates` (list templates), `GET /api/v1/templates/{name}` (detail), `POST /api/v1/templates/{name}/deploy` (deploy from template with variable values).

**Priority:** P2 🟢  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S7-T003 (Template Engine), S2-T003 (API)  
**Estimated Time:** 3 hours

**Subtask S7-T004.1:** Implement ListTemplates handler  
**Subtask S7-T004.2:** Implement GetTemplate handler (with variables, description)  
**Subtask S7-T004.3:** Implement DeployFromTemplate handler  
**Subtask S7-T004.4:** Write integration tests

**Acceptance Criteria:**
```
GIVEN available templates
WHEN I GET /api/v1/templates
THEN I receive a list with name, description, and required variables

GIVEN a template with variables
WHEN I POST /api/v1/templates/static-site/deploy with variable values
THEN the application is deployed with the provided values
```

---

##### Task S7-T005: Implement Dashboard Template UI

**Description:** Template browser in dashboard: list available templates with icons, template detail view with variable inputs, deploy from template with progress display.

**Priority:** P3 🔵  
**Difficulty:** D2 🟡  
**Owner:** Frontend  
**Dependencies:** S3-T001 (React scaffold), S7-T004 (Template API)  
**Estimated Time:** 4 hours

**Subtask S7-T005.1:** Create TemplateList component  
**Subtask S7-T005.2:** Create TemplateDetail component (description, variables form)  
**Subtask S7-T005.3:** Implement deploy-from-template flow  
**Subtask S7-T005.4:** Style with Tailwind

**Acceptance Criteria:**
```
GIVEN the templates page
WHEN I select "static-site" and fill in required variables
THEN the application deploys and I see progress
```

---

##### Task S7-T006: Implement Dashboard Plugin Manager

**Description:** Plugin management page: list installed plugins, browse available plugins, install/uninstall buttons, plugin detail (version, author, capabilities, permissions).

**Priority:** P3 🔵  
**Difficulty:** D3 🟠  
**Owner:** Frontend  
**Dependencies:** S3-T001 (React scaffold), S7-T002 (Marketplace API)  
**Estimated Time:** 4 hours

**Subtask S7-T006.1:** Create PluginList components (installed + available)  
**Subtask S7-T006.2:** Create PluginDetail component  
**Subtask S7-T006.3:** Implement install/uninstall actions  
**Subtask S7-T006.4:** Show permission approval dialog on install  
**Subtask S7-T006.5:** Style with Tailwind

**Acceptance Criteria:**
```
GIVEN the plugins page
WHEN I click "Install" on an available plugin
THEN a permission dialog appears, and on approval the plugin is installed
```

---

## 11. Sprint 8: Automation & Scheduler

**Theme:** Implement scheduled tasks, background jobs, and workflow automation.

**Goal:** Users can schedule recurring operations, run background jobs, and define workflows.

**Duration:** ~4 days

---

### Epic: Automation

#### Feature 8.1: Scheduler

**Business Value:** Cron-like scheduling for recurring operations (backups, cleanup, health checks).  
**Technical Value:** Scheduler is a Kernel primitive used by all subsystems.

---

##### Task S8-T001: Implement Scheduler

**Description:** In-memory scheduler supporting cron expressions and interval-based scheduling. Jobs are registered with a handler function and executed on schedule. Job history (success/failure/duration) tracked. Error handling with retry.

**Priority:** P1 🟡  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Dependencies:** S1-T001 (Subsystem interface)  
**Estimated Time:** 6 hours

**Subtask S8-T001.1:** Define Job, Schedule, CronExpression types  
**Subtask S8-T001.2:** Implement cron expression parser (standard 5-field cron)  
**Subtask S8-T001.3:** Implement in-memory scheduler with time.Ticker  
**Subtask S8-T001.4:** Implement job registration: Schedule(name, cron, handler)  
**Subtask S8-T001.5:** Implement job history tracking (last run, last result, duration)  
**Subtask S8-T001.6:** Implement job error handling (retry up to 3 times)  
**Subtask S8-T001.7:** Register scheduler as Kernel subsystem  
**Subtask S8-T001.8:** Write tests (cron parsing, scheduling, execution, error handling)

**Acceptance Criteria:**
```
GIVEN a job scheduled with cron "0 */1 * * *" (every hour)
WHEN the scheduler runs
THEN the handler is called every hour

GIVEN a job that fails
WHEN the scheduler detects the failure
THEN it retries up to 3 times with backoff
AND the failure is logged
```

---

##### Task S8-T002: Implement Scheduler API Endpoints

**Description:** REST endpoints for job management: `POST /api/v1/scheduler/jobs` (create scheduled job), `GET /api/v1/scheduler/jobs` (list), `GET /api/v1/scheduler/jobs/{id}` (detail with history), `DELETE /api/v1/scheduler/jobs/{id}` (delete), `POST /api/v1/scheduler/jobs/{id}/trigger` (manual trigger).

**Priority:** P2 🟢  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S8-T001 (Scheduler), S2-T003 (API)  
**Estimated Time:** 3 hours

**Subtask S8-T002.1:** Implement CreateJob handler  
**Subtask S8-T002.2:** Implement ListJobs handler  
**Subtask S8-T002.3:** Implement GetJob handler (with history)  
**Subtask S8-T002.4:** Implement DeleteJob handler  
**Subtask S8-T002.5:** Implement TriggerJob handler  
**Subtask S8-T002.6:** Write integration tests

**Acceptance Criteria:**
```
GIVEN a scheduler API
WHEN I POST /api/v1/scheduler/jobs with {"name": "backup", "cron": "0 3 * * *", "action": "database.backup"}
THEN the job is registered and appears in the job list
```

---

#### Feature 8.2: Workflow Engine

**Business Value:** Multi-step workflows (deploy-with-database, backup-to-storage) need coordination.  
**Technical Value:** Workflow engine defines state machines for complex operations.

---

##### Task S8-T003: Implement Workflow Engine

**Description:** Simple workflow engine for multi-step operations. Workflow = DAG of steps. Each step = capability call. Steps can run sequentially or in parallel. State machine tracks workflow execution (pending → running → completed/failed). Rollback on failure.

**Priority:** P2 🟢  
**Difficulty:** D4 🔴  
**Owner:** Platform  
**Dependencies:** S1-T005 (CapabilityRegistry), S8-T001 (Scheduler)  
**Estimated Time:** 2 days

**Subtask S8-T003.1:** Define Workflow, Step, WorkflowState types  
**Subtask S8-T003.2:** Implement workflow definition parser (YAML-based steps with dependencies)  
**Subtask S8-T003.3:** Implement step executor (sequential + parallel)  
**Subtask S8-T003.4:** Implement state machine (pending → running → success/failed)  
**Subtask S8-T003.5:** Implement rollback on failure (reverse steps)  
**Subtask S8-T003.6:** Implement workflow persistence (save/load from SQLite)  
**Subtask S8-T003.7:** Write tests (sequential, parallel, failure, rollback)

**Acceptance Criteria:**
```
GIVEN a workflow with steps [create_db → deploy_app → configure_dns]
WHEN executed
THEN steps run in order, each receiving output from the previous

GIVEN step 2 fails
WHEN rollback is triggered
THEN step 1 is rolled back (database deleted)
AND the workflow status is "failed"
```

---

##### Task S8-T004: Implement Background Job Queue

**Description:** Simple in-memory job queue for async processing. Jobs are queued, picked up by worker goroutines, executed, and results stored. Used for long-running operations (image pulls, backups, large uploads). Job status can be polled.

**Priority:** P2 🟢  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Dependencies:** S1-T001 (Subsystem interface)  
**Estimated Time:** 6 hours

**Subtask S8-T004.1:** Define Job, JobStatus, JobResult types  
**Subtask S8-T004.2:** Implement in-memory job queue with configurable worker pool  
**Subtask S8-T004.3:** Implement job lifecycle (queued → running → completed/failed)  
**Subtask S8-T004.4:** Implement job progress tracking (percent + message)  
**Subtask S8-T004.5:** Implement job cancellation  
**Subtask S8-T004.6:** Implement job result storage (for polling)  
**Subtask S8-T004.7:** Write tests (queue, execute, progress, cancel, concurrent)

**Acceptance Criteria:**
```
GIVEN a job is queued
WHEN a worker picks it up
THEN status changes to "running" with progress 0%
AND when complete, status is "completed" with result data

GIVEN a long-running job
WHEN I cancel it
THEN status changes to "cancelled" and the worker stops processing
```

---

##### Task S8-T005: Implement Dashboard Automation UI

**Description:** Jobs and workflows management in dashboard: job list with schedule and last run, workflow list with status and steps, create scheduled job form, workflow execution visualization.

**Priority:** P3 🔵  
**Difficulty:** D3 🟠  
**Owner:** Frontend  
**Dependencies:** S3-T001 (React scaffold), S8-T002 (Scheduler API), S8-T003 (Workflow Engine)  
**Estimated Time:** 6 hours

**Subtask S8-T005.1:** Create Jobs list page with schedule, last run, next run  
**Subtask S8-T005.2:** Create CreateJob form with cron expression builder  
**Subtask S8-T005.3:** Create Workflow list page with status and step visualization  
**Subtask S8-T005.4:** Create Workflow detail page with step-by-step progress  
**Subtask S8-T005.5:** Style with Tailwind

**Acceptance Criteria:**
```
GIVEN the automation page
WHEN I create a scheduled backup job
THEN it appears in the job list with the cron schedule

GIVEN a running workflow
WHEN I view its detail page
THEN I see step-by-step progress with status for each step
```

---

## 12. Sprint 9: Monitoring & Analytics

**Theme:** Implement observability — metrics collection, logging pipeline, alerting, and dashboards.

**Goal:** System metrics are collected, logs are searchable, alerts can be configured, dashboards display real-time data.

**Duration:** ~4 days

---

### Epic: Observability

#### Feature 9.1: Metrics Pipeline

**Business Value:** Without metrics, you can't know if the system is healthy.  
**Technical Value:** Prometheus-compatible metrics enable integration with standard tools.

---

##### Task S9-T001: Implement Metrics Pipeline

**Description:** Extend the metrics system from Sprint 0. Implement Prometheus exporter endpoint (`/metrics`). Collect standard metrics: HTTP request count/latency, capability operation count/latency, subsystem health, memory/CPU/goroutine per subsystem. Configurable retention.

**Priority:** P1 🟡  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Dependencies:** S0-T007 (Logger), S2-T001 (HTTP Server)  
**Estimated Time:** 6 hours

**Subtask S9-T001.1:** Implement Prometheus metrics exporter on `/metrics`  
**Subtask S9-T001.2:** Add HTTP metrics (request count, latency histogram by path/method/status)  
**Subtask S9-T001.3:** Add capability metrics (operation count, latency by capability)  
**Subtask S9-T001.4:** Add system metrics (memory, goroutines, CPU per subsystem)  
**Subtask S9-T001.5:** Add health metrics (subsystem state counts)  
**Subtask S9-T001.6:** Implement configurable retention with periodic cleanup  
**Subtask S9-T001.7:** Write tests (metric recording, Prometheus format, retention)

**Acceptance Criteria:**
```
GIVEN the metrics system
WHEN I curl /metrics
THEN I receive Prometheus-formatted text with:
  - cloudos_http_requests_total{method,path,status}
  - cloudos_http_request_duration_seconds{method,path}
  - cloudos_capability_operations_total{capability,operation}
  - cloudos_subsystem_state{subsystem}
  - cloudos_memory_bytes, cloudos_goroutines
```

---

#### Feature 9.2: Logging Pipeline

**Business Value:** Structured, searchable logs are essential for debugging and audit.  
**Technical Value:** Log aggregation enables centralized monitoring.

---

##### Task S9-T002: Implement Log Aggregation

**Description:** Extend logging from Sprint 0. Add log levels, structured fields, context propagation. Implement log query API (`GET /api/v1/logs?level=error&service=kernel&since=1h`). Configurable log retention and rotation.

**Priority:** P1 🟡  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Dependencies:** S0-T007 (Logger), S2-T003 (API)  
**Estimated Time:** 6 hours

**Subtask S9-T002.1:** Add log level filtering (debug, info, warn, error)  
**Subtask S9-T002.2:** Add structured field propagation (trace_id, subsystem, request_id)  
**Subtask S9-T002.3:** Implement in-memory log buffer with ring buffer for recent logs  
**Subtask S9-T002.4:** Implement log query API with filtering (level, service, time range, text search)  
**Subtask S9-T002.5:** Implement log file rotation (size-based, configurable max files)  
**Subtask S9-T002.6:** Add log streaming endpoint (SSE for tail -f)  
**Subtask S9-T002.7:** Write tests (query, filtering, rotation, streaming)

**Acceptance Criteria:**
```
GIVEN logs are being generated
WHEN I GET /api/v1/logs?level=error&since=1h
THEN I receive matching log entries with timestamp, level, message, and structured fields

GIVEN the log streaming endpoint
WHEN I connect to GET /api/v1/logs/stream
THEN I receive SSE events for new log entries in real-time
```

---

#### Feature 9.3: Alerting

**Business Value:** Operators need to know when things go wrong without watching dashboards.  
**Technical Value:** Alert rules enable automated incident response.

---

##### Task S9-T003: Implement Alert Engine

**Description:** Rule-based alert engine. Alert rules defined as YAML: condition (metric > threshold for duration), severity (critical/warning/info), actions (log, webhook, email stub). Alert state machine: OK → Pending → Firing → Resolved.

**Priority:** P2 🟢  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Dependencies:** S9-T001 (Metrics), S8-T001 (Scheduler)  
**Estimated Time:** 6 hours

**Subtask S9-T003.1:** Define AlertRule, AlertState, AlertEvent types  
**Subtask S9-T003.2:** Implement alert rule parser (YAML conditions)  
**Subtask S9-T003.3:** Implement alert evaluator (check metrics vs rules on schedule)  
**Subtask S9-T003.4:** Implement alert state machine (OK → Pending → Firing → Resolved)  
**Subtask S9-T003.5:** Implement alert actions (log entry, webhook POST)  
**Subtask S9-T003.6:** Implement alert history storage  
**Subtask S9-T003.7:** Write tests (rule evaluation, state transitions, actions)

**Default alert rules:**
```yaml
rules:
  - name: "HighErrorRate"
    condition: "cloudos_http_requests_total{status=~'5..'} / cloudos_http_requests_total > 0.05"
    duration: "5m"
    severity: critical
    actions: ["log", "webhook"]

  - name: "HighMemoryUsage"
    condition: "cloudos_memory_bytes > 1073741824"  # > 1GB
    duration: "5m"
    severity: warning
    actions: ["log"]

  - name: "SubsystemDown"
    condition: "cloudos_subsystem_state{state='failed'} > 0"
    duration: "1m"
    severity: critical
    actions: ["log", "webhook"]
```

**Acceptance Criteria:**
```
GIVEN an alert rule with condition "cloudos_memory_bytes > 104857600" for 1m
WHEN memory exceeds 100MB for 1 minute
THEN the alert state transitions: OK → Pending → Firing
AND a log action is triggered
WHEN memory drops below threshold
THEN the alert transitions: Firing → Resolved
```

---

##### Task S9-T004: Implement Alert API Endpoints

**Description:** `GET /api/v1/alerts` (list alert rules), `POST /api/v1/alerts` (create rule), `PUT /api/v1/alerts/{id}` (update), `DELETE /api/v1/alerts/{id}` (delete), `GET /api/v1/alerts/events` (alert history).

**Priority:** P2 🟢  
**Difficulty:** D2 🟡  
**Owner:** Backend  
**Dependencies:** S9-T003 (Alert Engine), S2-T003 (API)  
**Estimated Time:** 3 hours

**Subtask S9-T004.1:** Implement ListAlertRules handler  
**Subtask S9-T004.2:** Implement CreateAlertRule handler  
**Subtask S9-T004.3:** Implement UpdateAlertRule handler  
**Subtask S9-T004.4:** Implement DeleteAlertRule handler  
**Subtask S9-T004.5:** Implement AlertEvents handler (history)  
**Subtask S9-T004.6:** Write integration tests

**Acceptance Criteria:**
```
GIVEN an alerts API
WHEN I POST /api/v1/alerts with a valid rule
THEN the rule is created and appears in the list

GIVEN firing alerts
WHEN I GET /api/v1/alerts/events
THEN I see alert events with start time, severity, and value
```

---

#### Feature 9.4: Monitoring Dashboard

**Business Value:** Visual dashboards provide at-a-glance system health.  
**Technical Value:** Real-time data visualization with auto-refresh.

---

##### Task S9-T005: Implement Dashboard Monitoring UI

**Description:** Monitoring dashboard with: system health overview, metrics charts (CPU, memory, request rate, error rate), alert list with status, log viewer with search/filter, "since" time range selector.

**Priority:** P2 🟢  
**Difficulty:** D3 🟠  
**Owner:** Frontend  
**Dependencies:** S3-T001 (React scaffold), S9-T001 (Metrics), S9-T002 (Log API), S9-T004 (Alert API)  
**Estimated Time:** 8 hours

**Subtask S9-T005.1:** Create MonitoringPage with tabs (Overview, Metrics, Logs, Alerts)  
**Subtask S9-T005.2:** Create health overview cards (green/yellow/red per subsystem)  
**Subtask S9-T005.3:** Create metrics chart component (time series, auto-refresh)  
**Subtask S9-T005.4:** Create alert list with status indicators  
**Subtask S9-T005.5:** Create log viewer with search and level filter  
**Subtask S9-T005.6:** Implement time range selector (1h, 6h, 24h, 7d)  
**Subtask S9-T005.7:** Style with Tailwind

**Acceptance Criteria:**
```
GIVEN the monitoring page
WHEN I select the Metrics tab
THEN I see CPU and memory charts that auto-refresh every 10 seconds

GIVEN the monitoring page
WHEN I select the Logs tab and filter by level=error
THEN only error-level logs are displayed
```

---

##### Task S9-T006: Implement Health Check Endpoint Improvements

**Description:** Enhance `/health` endpoint: return per-subsystem health, dependency health, uptime, version, metrics summary. Add readiness and liveness probes for container orchestration.

**Priority:** P2 🟢  
**Difficulty:** D2 🟡  
**Owner:** Platform  
**Dependencies:** S1-T008 (Health Manager), S2-T003 (API)  
**Estimated Time:** 3 hours

**Subtask S9-T006.1:** Enhance health response with per-subsystem status  
**Subtask S9-T006.2:** Add dependency health (subsystem dependencies)  
**Subtask S9-T006.3:** Add system info (version, uptime, build time)  
**Subtask S9-T006.4:** Add `/readyz` (readiness) and `/livez` (liveness) endpoints  
**Subtask S9-T006.5:** Write tests

**Acceptance Criteria:**
```
GIVEN a running Kernel
WHEN I GET /health
THEN I receive: {"status":"healthy","version":"0.1.0","uptime":"2h30m","subsystems":[{"name":"kernel","status":"healthy"},...]}
```

---

## 13. Sprint 10: Production Hardening

**Theme:** Security audit, performance optimization, comprehensive testing, documentation, and release.

**Goal:** CloudOS v0.1 is ready for alpha users. Documented, tested, packaged.

**Duration:** ~5 days

---

### Epic: Security

#### Feature 10.1: Security Hardening

**Business Value:** Users trust CloudOS with their applications and data.  
**Technical Value:** Security is built in, not bolted on.

---

##### Task S10-T001: Security Audit

**Description:** Systematic security review: dependency vulnerabilities (govulncheck, npm audit), JWT implementation review (signing algorithm, expiry, revocation), API authentication coverage (every endpoint has auth or is explicitly public), secrets handling (never logged, encrypted at rest), SQL injection review, path traversal review, permission boundary review.

**Priority:** P0 🔴  
**Difficulty:** D3 🟠  
**Owner:** Security  
**Dependencies:** All Sprint 0-9 tasks  
**Estimated Time:** 8 hours

**Subtask S10-T001.1:** Run `govulncheck ./...` and fix vulnerabilities  
**Subtask S10-T001.2:** Run `npm audit` in dashboard and fix vulnerabilities  
**Subtask S10-T001.3:** Review all API endpoints for auth coverage  
**Subtask S10-T001.4:** Review secrets handling in config and logs  
**Subtask S10-T001.5:** Review file path handling for traversal attacks  
**Subtask S10-T001.6:** Review permission enforcement (no escalation paths)  
**Subtask S10-T001.7:** Document findings in ADR

**Acceptance Criteria:**
```
GIVEN a completed security audit
WHEN I review the audit report
THEN all critical and high findings are fixed
AND all findings are documented in /decisions/ with remediation status
```

---

##### Task S10-T002: Add Rate Limiting

**Description:** Implement per-IP and per-user rate limiting on API endpoints. Configurable limits per endpoint group (auth: 5/min, deploy: 10/min, read: 100/min). Return 429 with Retry-After header.

**Priority:** P1 🟡  
**Difficulty:** D2 🟡  
**Owner:** Platform  
**Dependencies:** S2-T001 (HTTP Server)  
**Estimated Time:** 4 hours

**Subtask S10-T002.1:** Implement token bucket rate limiter  
**Subtask S10-T002.2:** Add per-IP rate limiting middleware  
**Subtask S10-T002.3:** Add per-user rate limiting middleware  
**Subtask S10-T002.4:** Add configurable limits per endpoint group  
**Subtask S10-T002.5:** Return 429 with Retry-After header  
**Subtask S10-T002.6:** Write tests (rate limit hit, rate limit reset, configurable)

**Acceptance Criteria:**
```
GIVEN a rate limit of 5 requests per minute on auth endpoints
WHEN I send 6 requests in 1 minute to POST /api/v1/auth/login
THEN the 6th request gets a 429 response with Retry-After header
```

---

#### Feature 10.2: Input Validation

**Business Value:** Invalid input should never crash the system or corrupt data.  
**Technical Value:** Defense in depth — validate at API layer and capability layer.

---

##### Task S10-T003: Implement Comprehensive Input Validation

**Description:** Validate all API inputs: request body JSON schema validation, string length limits, numeric range limits, enum validation, Content-Type checking, maximum body size enforcement. Return consistent 400 errors with field-level details.

**Priority:** P1 🟡  
**Difficulty:** D3 🟠  
**Owner:** Backend  
**Dependencies:** S2-T003 (API)  
**Estimated Time:** 6 hours

**Subtask S10-T003.1:** Implement JSON schema validation middleware  
**Subtask S10-T003.2:** Define validation schemas for all endpoints  
**Subtask S10-T003.3:** Implement string length validation (min/max)  
**Subtask S10-T003.4:** Implement numeric range validation  
**Subtask S10-T003.5:** Implement enum validation  
**Subtask S10-T003.6:** Implement maximum body size (default 10MB)  
**Subtask S10-T003.7:** Return 400 with field-level error details  
**Subtask S10-T003.8:** Write tests (invalid JSON, wrong types, out of range, too large)

**Acceptance Criteria:**
```
GIVEN a deploy request with image="" (empty string)
WHEN submitted
THEN I get 400 with {"error":{"code":"VALIDATION_ERROR","fields":[{"field":"image","message":"image is required"}]}}

GIVEN a body exceeding 10MB
WHEN submitted
THEN I get 413 with {"error":{"code":"REQUEST_TOO_LARGE","message":"Request body exceeds 10MB limit"}}
```

---

### Epic: Performance

#### Feature 10.3: Performance Optimization

**Business Value:** Fast response times and low resource usage.  
**Technical Value:** Performance targets must be verified before release.

---

##### Task S10-T004: Performance Profiling

**Description:** Profile the Kernel: CPU profile, memory profile, goroutine profile, mutex profile. Identify hot spots and optimize. Benchmark key operations: capability call latency, event bus throughput, config load time, API request latency.

**Priority:** P2 🟢  
**Difficulty:** D3 🟠  
**Owner:** Platform  
**Dependencies:** All Sprints  
**Estimated Time:** 8 hours

**Subtask S10-T004.1:** Run `pprof` for CPU profile (boot, deploy, API request)  
**Subtask S10-T004.2:** Run `pprof` for memory profile (steady state, under load)  
**Subtask S10-T004.3:** Run race detector for all integration tests  
**Subtask S10-T004.4:** Benchmark capability call latency (target: < 1ms local)  
**Subtask S10-T004.5:** Benchmark event bus throughput (target: > 100,000 events/sec)  
**Subtask S10-T004.6:** Optimize top 3 hot spots found in profiling  
**Subtask S10-T004.7:** Write benchmark tests

**Acceptance Criteria:**
```
GIVEN a profiled Kernel
WHEN I review the performance report
THEN capability call latency < 1ms (local provider)
AND event bus throughput > 100,000 events/sec
AND memory usage < 100MB at idle
AND no goroutine leaks in any integration test
```

---

##### Task S10-T005: Binary Optimization

**Description:** Optimize Go binary size and startup time. Use `-ldflags="-s -w"` for stripping, `-trimpath` for smaller binaries, and build caching. Target: < 30MB binary, < 500ms cold start.

**Priority:** P2 🟢  
**Difficulty:** D1 🟢  
**Owner:** Platform  
**Dependencies:** S10-T004 (Profiling)  
**Estimated Time:** 2 hours

**Subtask S10-T005.1:** Add ldflags for size reduction  
**Subtask S10-T005.2:** Add trimpath for smaller binaries  
**Subtask S10-T005.3:** Verify binary size < 30MB  
**Subtask S10-T005.4:** Measure cold start time (target: < 500ms)

**Acceptance Criteria:**
```
GIVEN the optimized build command
WHEN I run `go build -ldflags="-s -w" -trimpath -o cloudos ./cmd/cloudos`
THEN binary size < 30MB
AND cold start to health check response < 500ms
```

---

### Epic: Documentation

#### Feature 10.4: User & Developer Documentation

**Business Value:** Users and contributors need documentation to use and extend CloudOS.  
**Technical Value:** Well-documented projects attract contributors and users.

---

##### Task S10-T006: Write User Documentation

**Description:** Quickstart guide (5 minutes to first deploy), CLI reference, API reference (auto-generated from routes), configuration reference, troubleshooting guide, architecture overview for users.

**Priority:** P1 🟡  
**Difficulty:** D3 🟠  
**Owner:** Documentation  
**Dependencies:** All Sprints  
**Estimated Time:** 8 hours

**Subtask S10-T006.1:** Write 5-minute quickstart (Docker, login, deploy, dashboard)  
**Subtask S10-T006.2:** Write CLI reference (all commands with examples)  
**Subtask S10-T006.3:** Generate API reference from route definitions with OpenAPI  
**Subtask S10-T006.4:** Write configuration reference (all config options)  
**Subtask S10-T006.5:** Write troubleshooting guide (common issues and solutions)  
**Subtask S10-T006.6:** Verify all documented commands actually work

**Acceptance Criteria:**
```
GIVEN the user documentation
WHEN a new user follows the quickstart
THEN they can deploy their first container in < 5 minutes
AND all commands in the CLI reference produce documented output
```

---

##### Task S10-T007: Write Developer Documentation

**Description:** Architecture overview for contributors, how to add a provider (SDK guide), how to add a capability, how to build the dashboard, how to run tests, how to release.

**Priority:** P2 🟢  
**Difficulty:** D2 🟡  
**Owner:** Documentation  
**Dependencies:** All Sprints  
**Estimated Time:** 6 hours

**Subtask S10-T007.1:** Write contributor onboarding guide  
**Subtask S10-T007.2:** Write "How to add a provider" SDK guide  
**Subtask S10-T007.3:** Write "How to add a capability" guide  
**Subtask S10-T007.4:** Write dashboard development guide  
**Subtask S10-T007.5:** Write testing guide  
**Subtask S10-T007.6:** Write release process guide

**Acceptance Criteria:**
```
GIVEN the developer documentation
WHEN a new contributor reads the SDK guide
THEN they can write and register a simple provider in < 1 hour
```

---

### Epic: Release

#### Feature 10.5: Release Process

**Business Value:** Users need a stable, versioned release they can depend on.  
**Technical Value:** Automated release pipeline ensures consistency.

---

##### Task S10-T008: v0.1 Release

**Description:** Finalize v0.1: update CHANGELOG.md, tag v0.1.0, build Docker image, push to ghcr.io, build CLI binaries for all platforms, create GitHub release with release notes, verify quickstart end-to-end.

**Priority:** P0 🔴  
**Difficulty:** D2 🟡  
**Owner:** DevOps  
**Dependencies:** All tasks  
**Estimated Time:** 6 hours

**Subtask S10-T008.1:** Update CHANGELOG.md with all changes since v0.0.0  
**Subtask S10-T008.2:** Tag v0.1.0 in git  
**Subtask S10-T008.3:** Build and push Docker image to ghcr.io/cloudos/cloudos:v0.1.0  
**Subtask S10-T008.4:** Build CLI binaries (linux-amd64, darwin-amd64, windows-amd64)  
**Subtask S10-T008.5:** Create GitHub release with release notes and binary attachments  
**Subtask S10-T008.6:** Verify `docker compose up` quickstart works end-to-end  
**Subtask S10-T008.7:** Verify CLI commands (login, deploy, ps, status)  
**Subtask S10-T008.8:** Run E2E test suite (login → deploy → verify → stop)

**Acceptance Criteria:**
```
GIVEN the v0.1.0 release
WHEN I run `docker compose up`
THEN CloudOS starts in < 5 seconds

GIVEN the running release
WHEN I run `cloudos deploy nginx:alpine`
THEN a container is deployed and accessible

GIVEN the CLI binary for my platform
WHEN I run `cloudos --help`
THEN all documented commands are available

GIVEN the release artifacts
WHEN I inspect them
THEN Docker image is on ghcr.io, CLI binaries are on GitHub Releases, CHANGELOG is updated
```

---

## 14. Backlog Index

### Task Count by Sprint

| Sprint | Theme | Tasks | Total Subtasks | Est. Effort |
|--------|-------|-------|----------------|-------------|
| 0 | Foundation & Tooling | 8 | 42 | ~3 days |
| 1 | Kernel Core | 10 | 68 | ~5 days |
| 2 | API & Auth | 8 | 52 | ~4 days |
| 3 | Dashboard | 6 | 39 | ~5 days |
| 4 | Storage, Database & Secrets | 8 | 55 | ~4 days |
| 5 | Deployments & Networking | 7 | 41 | ~4 days |
| 6 | AI Engine | 6 | 35 | ~4 days |
| 7 | Marketplace & Templates | 6 | 34 | ~4 days |
| 8 | Automation & Scheduler | 5 | 34 | ~4 days |
| 9 | Monitoring & Analytics | 6 | 42 | ~4 days |
| 10 | Production Hardening | 8 | 52 | ~5 days |
| **Total** | | **78** | **494** | **~46 days** |

### Epic Index

| Epic | Sprint | Tasks |
|------|--------|-------|
| Repository Setup | 0 | S0-T001 to S0-T003 |
| Continuous Integration | 0 | S0-T004 |
| Developer Experience | 0 | S0-T005 to S0-T006 |
| Core Infrastructure | 0 | S0-T007 to S0-T008 |
| Kernel Runtime | 1 | S1-T001 to S1-T003 |
| Capability Registry | 1 | S1-T004 to S1-T005 |
| Event Bus | 1 | S1-T006 |
| Dependency Injection | 1 | S1-T007 |
| Lifecycle Manager | 1 | S1-T008 |
| Plugin Loader | 1 | S1-T009 to S1-T010 |
| REST API | 2 | S2-T001 to S2-T003 |
| Authentication | 2 | S2-T004 to S2-T005 |
| RBAC Authorization | 2 | S2-T006 |
| User Management | 2 | S2-T007 to S2-T008 |
| Web Dashboard | 3 | S3-T001 to S3-T006 |
| Storage | 4 | S4-T001 to S4-T003 |
| Database | 4 | S4-T004 to S4-T005 |
| Secrets | 4 | S4-T006 to S4-T008 |
| Compute | 5 | S5-T001 to S5-T003 |
| Networking | 5 | S5-T004 to S5-T007 |
| AI Integration | 6 | S6-T001 to S6-T006 |
| Marketplace | 7 | S7-T001 to S7-T006 |
| Automation | 8 | S8-T001 to S8-T005 |
| Observability | 9 | S9-T001 to S9-T006 |
| Security | 10 | S10-T001 to S10-T003 |
| Performance | 10 | S10-T004 to S10-T005 |
| Documentation | 10 | S10-T006 to S10-T007 |
| Release | 10 | S10-T008 |

### Priority Distribution

| Priority | Count |
|----------|-------|
| P0 — Critical | 20 |
| P1 — High | 26 |
| P2 — Medium | 24 |
| P3 — Low | 8 |

---

> **End of Document — CloudOS Engineering Backlog v0.1**
