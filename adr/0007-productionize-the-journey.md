# ADR-0007: Productionize the Journey

**Status:** Active  
**Date:** 2026-06-30  
**Author:** Sunny (Product Lead)

## Context

CloudOS v0.2.0-alpha ("The First Deployment") completed the core product loop:

```
User Intent → Plan Preview → Approval → Execution → Running Application
```

This is no longer a prototype. It is a complete product experience. The next milestone must make this journey **excellent** rather than adding breadth.

## North Star

> **A user with zero cloud experience can deploy a GitHub repository in under one minute.**

Every feature, every change, every decision must improve that journey or it belongs in the backlog.

## Priorities (1–5 scale)

| Priority | Item | Track |
|---|---|---|
| ⭐⭐⭐⭐⭐ | Deployment Timeline | Product Layer |
| ⭐⭐⭐⭐⭐ | Buildpacks | Product Layer |
| ⭐⭐⭐⭐⭐ | Runtime Interface | Core Platform |
| ⭐⭐⭐⭐ | Live Logs | Product Layer |
| ⭐⭐⭐⭐ | Failure Analysis | Product Layer |
| ⭐⭐⭐⭐ | GitHub Integration | Ecosystem |
| ⭐⭐⭐⭐ | Docker Runtime | Ecosystem |
| ⭐⭐⭐ | Remote Runtime | Ecosystem |

## Execution Artifacts — Architectural Addition

Every workflow execution produces artifacts that should become first-class resources:

```
WorkflowExecution
  ├── Execution Plan
  ├── Logs
  ├── Build Output
  ├── Runtime Metadata
  ├── Health Reports
  └── Diagnostics
```

This unlocks downloads, replay, AI analysis, audit, and export for free.

## Postponed Indefinitely

These features will **not** be built until deployment reliability is proven:

- Authentication / authorization
- Teams and organizations
- Billing / subscriptions
- Marketplace
- Chatbots / conversational AI
- Multi-agent orchestration
- RAG / knowledge graph / chat history

A product with one amazing workflow beats a platform with fifty unfinished features.

## Repository Tracks

The codebase should evolve along three parallel tracks:

### 1. Core Platform
- Kernel
- Resource Engine
- Workflow Engine
- Controller Runtime

Changes slowly. Stays stable. This is the foundation.

### 2. Product Layer
- Applications
- Deployments
- Dashboard
- Intent UX
- Buildpacks
- Runtimes

Evolves quickly based on user feedback. This is where the product lives.

### 3. Ecosystem
- Docker runtime
- Kubernetes runtime
- GitHub integration
- VS Code extension
- CLI
- Mobile app
- Community buildpacks

Innovates freely without destabilizing the platform.

## AI Usage Policy

This is the critical boundary:

- ✅ **Allowed:** AI analyzes *past* execution artifacts (logs, workflow state, diagnostics) to explain failures.
- ❌ **Banned:** AI controls infrastructure, creates resources autonomously, or acts as a chatbot.

The only AI feature in scope for Milestone 2 is **Failure Analysis** — an execution explainer, not a chatbot.

```
Deployment failed.
Reason: npm install failed because package-lock.json
requires Node 22 while the runtime uses Node 20.
Suggested fix: Upgrade runtime to Node 22.
```

## Flagship Identity

CloudOS's identity is defined by one demo:

> **Deploy a GitHub app in under 60 seconds.**

Every new feature either makes that journey faster, simpler, more reliable, or more extensible. If it doesn't, it belongs in the backlog.

## Open-Source Platform Mindset

From this point forward:

- Publish a proper roadmap.
- Keep writing ADRs.
- Create a contributor guide.
- Add architecture diagrams.
- Record short demo videos for each milestone.

## Roadmap

```
v0.2.0-alpha
  ✓ First Deployment
  ↓
  Deployment Timeline    ←  CURRENT (SSE landing)
  ↓
  Buildpacks
  ↓
  Runtime Interface
  ↓
  Live Logs
  ↓
  Failure Analysis
  ↓
  GitHub Integration
  ↓
  Docker Runtime
  ↓
  Remote Runtime
```
