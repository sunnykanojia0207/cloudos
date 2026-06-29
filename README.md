# CloudOS

**AI-First, Plugin-Based Cloud Operating System**

[![CI](https://github.com/cloudos/cloudos/actions/workflows/ci.yml/badge.svg)](https://github.com/cloudos/cloudos/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

CloudOS is a next-generation cloud operating system — reimagined as a unified, extensible, AI-first platform. It brings together the best ideas from AWS, Google Cloud, Firebase, Vercel, and Supabase into a single, self-hostable operating system for the cloud.

## Architecture

```
CloudOS/
├── kernel/         → Go: operating system core
├── capabilities/   → Go: abstract interface contracts
├── providers/      → Go: concrete implementations
├── packages/       → Go: shared libraries
├── apps/           → React/TypeScript: dashboard, desktop
├── tools/          → Go: CLI entry points
└── docs/           → Documentation
```

## Quick Start

```bash
# Prerequisites: Go 1.24+
git clone https://github.com/cloudos/cloudos.git
cd cloudos

# Build
go build ./...

# Run (starts kernel + Control Plane API on :8080)
go run ./tools/cloudos

# Test the API (in another terminal):
curl -s http://localhost:8080/api/v1/health | jq .
curl -s http://localhost:8080/api/v1/version | jq .
curl -s http://localhost:8080/api/v1/kernel | jq .
curl -s http://localhost:8080/api/v1/system | jq .
curl -s http://localhost:8080/api/v1/capabilities | jq '.items[].id'
```

## API Reference

All endpoints are versioned under `/api/v1/`.

### System
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Aggregated + per-component health status |
| GET | `/api/v1/ready` | Kernel readiness probe (200 when running) |
| GET | `/api/v1/live` | Basic liveness probe (always 200) |
| GET | `/api/v1/version` | CloudOS version, commit, build metadata |
| GET | `/api/v1/kernel` | Kernel state, uptime, started-at timestamp |
| GET | `/api/v1/system` | Runtime info (OS, arch, Go version, CPU, goroutines) |

### Capability Discovery
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/capabilities` | List all registered capabilities (ResourceObject list) |
| GET | `/api/v1/capabilities/{id}` | Get a single capability descriptor |

Every capability response uses the **ResourceObject** envelope:

```json
{
  "id": "compute",
  "kind": "capability",
  "apiVersion": "v1",
  "metadata": {
    "labels": { "category": "compute", "status": "stable" }
  },
  "spec": {
    "name": "Compute",
    "displayName": "Compute Engine",
    "description": "Deploy, scale, and manage applications...",
    "version": { "major": 1, "minor": 0, "patch": 0 },
    "category": "compute",
    "tags": ["compute", "deployment", "containers"],
    "operations": [
      { "name": "deploy", "description": "Create a new deployment", "httpMethod": "POST", "path": "/deployments" }
    ]
  },
  "status": {
    "status": "stable",
    "available": false,
    "providerCount": 0
  }
}
```

## Project Status

CloudOS is in **active pre-alpha development**.

| Area | Status |
|------|--------|
| Go module & tooling | ✅ Complete |
| Kernel (lifecycle, events, scheduler, health, security, registry, plugin, DI) | ✅ Complete |
| Capability interfaces (compute, storage, database, ai, network) | ✅ Complete |
| Built-in providers (compute.local, storage.local, database.sqlite) | ✅ Complete |
| Configuration (YAML + env vars + hot-reload) | ✅ Complete |
| Structured logging (JSON, file rotation, context) | ✅ Complete |
| Error framework | ✅ Complete |
| Control Plane API (v1 — health, version, kernel, system, capabilities) | ✅ Complete |
| Capability Discovery API (list + get by ID, ResourceObject envelope) | ✅ Complete |
| Auth (JWT, RBAC) | 📋 Planned |
| Dashboard (React) | 📋 Planned |
| AI engine | 📋 Planned |

## License

MIT
