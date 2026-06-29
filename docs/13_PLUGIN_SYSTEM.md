# Plugin System

> **Document:** 05_PLUGIN_SYSTEM.md
> **Status:** Draft v0.1
> **Depends On:** [04_SYSTEM_DESIGN.md](./04_SYSTEM_DESIGN.md)

---

## 1. Plugin System Philosophy

The plugin system is the heart of CloudOS. Every capability in the platform — storage, compute, database, AI, networking, email, SMS, DNS — is implemented as a plugin. The core provides the runtime, lifecycle management, and interface contracts.

**Core Tenets:**
1. **The core never imports a plugin** — plugins implement core-defined interfaces
2. **Plugins are isolated** — a crash in one plugin never affects another
3. **Plugins are portable** — packaged as WASM or native binaries
4. **Plugins are versioned** — semantic versioning with dependency resolution

---

## 2. Plugin Types

### 2.1 By Runtime

| Type | Runtime | Isolation | Performance | Best For |
|------|---------|-----------|-------------|----------|
| **WASM** | wazero | Memory-safe sandbox | Good | Community plugins, untrusted code |
| **Native** | OS process | Process-level | Best | System plugins, performance-critical |
| **HTTP** | External service | Network boundary | Variable | Enterprise integrations |
| **Built-in** | Core process | None | Native | Essential system plugins |

### 2.2 By Role

| Role | Description | Examples |
|------|-------------|----------|
| **Provider** | Implements a capability | storage.s3, compute.docker |
| **Hook** | Listens to events, performs actions | logging.loki, monitoring.prometheus |
| **Extension** | Adds non-capability features | analytics.google, billing.stripe |
| **UI** | Adds dashboard panels | custom-dashboard, analytics-charts |

---

## 3. Plugin Lifecycle

```
                ┌──────────────┐
                │   Discover   │
                │  (Registry)  │
                └──────┬───────┘
                       │
                ┌──────▼───────┐
                │   Download   │
                │  (Package)   │
                └──────┬───────┘
                       │
                ┌──────▼───────┐
         ┌─────│   Verify     │
         │     │ (Signature)  │───── Invalid → Reject
         │     └──────┬───────┘
         │            │ Valid
         │     ┌──────▼───────┐
         │     │   Install    │
         │     │ (Extract)    │
         │     └──────┬───────┘
         │            │
         │     ┌──────▼───────┐
         │     │  Initialize  │
         │     │  (Setup)     │───── Error → Deactivate
         │     └──────┬───────┘
         │            │
         │     ┌──────▼───────┐
         │     │   Activate   │
         │     │  (Running)   │
         │     └──────┬───────┘
         │         │         │
         │    ┌─────┘         └─────┐
         │    ▼                     ▼
         │  ┌────────┐       ┌──────────┐
         │  │Health  │       │Deactivate │
         │  │Check   │       │(Stop)    │
         │  └────────┘       └────┬─────┘
         │                        │
         │                 ┌──────▼──────┐
         │                 │  Uninstall  │
         │                 │ (Cleanup)   │
         │                 └─────────────┘
         │
         └──→ On error at any stage → Error state → User intervention
```

---

## 4. Plugin Package Format (COSP)

CloudOS plugins are distributed as `.cosp` files, a tar.gz archive with a manifest.

```
plugin.cosp
│
├── manifest.yaml         # Plugin metadata (required)
├── capability.wasm       # WASM binary (or native binary)
├── ui/                   # UI extensions (optional)
│   ├── panel.tsx         # Dashboard panel
│   └── settings.tsx      # Settings form
├── schema.sql            # Database migrations (optional)
├── config.schema.json    # Configuration schema (optional)
├── assets/               # Icons, screenshots (optional)
│   ├── icon.svg
│   └── screenshot.png
└── signature.sig         # GPG signature (required for verified)
```

### 4.1 Manifest Specification

```yaml
# manifest.yaml
apiVersion: cloudos.io/v1
kind: Plugin
metadata:
  name: storage.minio
  displayName: MinIO Storage
  version: 1.0.0
  author:
    name: CloudOS Team
    email: plugins@cloudos.io
  license: MIT
  description: MinIO-compatible S3 storage provider
  tags:
    - storage
    - s3-compatible
    - self-hosted

spec:
  runtime: wasm
  capabilities:
    - storage:
        features: [buckets, presigned-urls, multipart]
  permissions:
    - network:outbound: ["*:9000"]
    - fs:read: ["/tmp"]
    - fs:write: ["/tmp"]
  resources:
    cpu: "0.5"
    memory: "128Mi"
    disk: "1Gi"
  dependencies:
    core: ">=0.1.0"
  config:
    schema: config.schema.json
    defaults:
      endpoint: "http://localhost:9000"
      region: "us-east-1"
      secure: false
  lifecycle:
    startup: 30s
    healthInterval: 15s
    shutdown: 10s
```

---

## 5. Plugin SDK

The Plugin SDK provides everything needed to build a CloudOS plugin.

### 5.1 SDK Languages

| Language | Status | Notes |
|----------|--------|-------|
| **Go** | ✅ Primary | Best WASM support, native plugin runtime |
| **Rust** | 🚧 Beta | For performance-critical plugins |
| **TypeScript** | 🚧 Beta | Via WASM, for community plugins |
| **Python** | 🔮 Planned | For AI/ML data processing plugins |

### 5.2 SDK API

```go
package cloudos

// Plugin is the main interface every plugin must implement.
type Plugin interface {
    // Metadata returns plugin identity information.
    Metadata() PluginMetadata

    // Initialize is called after installation, before activation.
    Initialize(ctx Context) error

    // Activate is called to start the plugin.
    Activate(ctx Context) error

    // Deactivate is called to stop the plugin.
    Deactivate(ctx Context) error

    // Health returns the current health status.
    Health() HealthStatus
}

// Context provides access to CloudOS core services.
type Context interface {
    // Config returns the plugin's configuration.
    Config() Config

    // Logger returns a structured logger.
    Logger() Logger

    // Events returns an event publisher.
    Events() EventPublisher

    // Store returns a KV store for plugin state.
    Store() KVStore

    // HTTPClient returns a pre-configured HTTP client.
    HTTPClient() *http.Client
}

// CapabilityRegistrar allows plugins to register capabilities.
type CapabilityRegistrar interface {
    RegisterCapability(name string, provider interface{})
}
```

### 5.3 Example: Storage Plugin

```go
package main

import (
    "context"
    "io"
    
    "cloudos.io/sdk/go/cloudos"
    "cloudos.io/sdk/go/capability"
)

type MinIOProvider struct {
    client *minio.Client
}

func (p *MinIOProvider) Initialize(ctx cloudos.Context) error {
    cfg := struct {
        Endpoint string `json:"endpoint"`
        AccessKey string `json:"access_key"`
        SecretKey string `json:"secret_key"`
        Region    string `json:"region"`
    }{}
    ctx.Config().Unmarshal(&cfg)
    
    client, err := minio.New(cfg.Endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
        Region: cfg.Region,
        Secure: true,
    })
    if err != nil {
        return err
    }
    p.client = client
    return nil
}

func (p *MinIOProvider) CreateBucket(ctx context.Context, name string, opts ...capability.BucketOption) (*capability.Bucket, error) {
    // Implementation
}

func (p *MinIOProvider) PutObject(ctx context.Context, bucket, key string, data io.Reader, opts ...capability.PutOption) (*capability.Object, error) {
    // Implementation
}

// ... other StorageProvider methods
```

---

## 6. Plugin Registry

The plugin registry is the source of truth for all available plugins.

### 6.1 Registry API

```
# Search plugins
GET /v1/plugins?q=storage&capability=storage

# Get plugin details
GET /v1/plugins/storage.minio

# Get plugin versions
GET /v1/plugins/storage.minio/versions

# Download plugin package
GET /v1/plugins/storage.minio/versions/1.0.0/download

# Publish plugin (authenticated)
POST /v1/plugins
Content-Type: multipart/form-data
```

### 6.2 Registry Index

```json
{
  "plugin": "storage.minio",
  "versions": [
    {
      "version": "1.0.0",
      "published": "2026-06-01T00:00:00Z",
      "checksum": "sha256:abc123...",
      "signature": "...",
      "compatibility": {
        "core": ">=0.1.0 <0.2.0",
        "platforms": ["linux/amd64", "linux/arm64"]
      }
    }
  ],
  "stats": {
    "downloads": 15000,
    "rating": 4.5,
    "reviews": 42
  }
}
```

---

## 7. Plugin Security

### 7.1 Permission Model

Plugins declare required permissions in their manifest. The user approves permissions at install time.

```
Permissions:
  network:outbound:[hosts]    → Can make outbound HTTP connections
  network:inbound:[ports]     → Can listen on ports
  fs:read:[paths]             → Can read files
  fs:write:[paths]            → Can write files
  process:exec                → Can execute subprocesses
  capability:register         → Can register capabilities
  event:publish:[types]       → Can publish events
  event:subscribe:[types]     → Can subscribe to events
  secrets:read:[keys]         → Can read secrets
```

### 7.2 Sandboxing

| Runtime | Sandbox | Resource Limits |
|---------|---------|-----------------|
| WASM | Memory-safe, no syscalls | Linear memory, call depth |
| Native (Linux) | seccomp, landlock, cgroups | CPU, memory, disk, network |
| Native (macOS) | sandbox-exec | CPU, memory |
| Native (Windows) | AppContainer | CPU, memory, network |
| HTTP | Network-only | Rate limiting |

### 7.3 Plugin Signing

- All plugins in the marketplace must be signed
- Signature verification at install time
- Key management via CloudOS trust store
- Developer keys registered with the registry

---

## 8. Built-in Plugins

CloudOS ships with a set of essential built-in plugins:

| Plugin | Capability | Default Provider |
|--------|------------|------------------|
| `storage.local` | Storage | Local filesystem |
| `compute.local` | Compute | Local process execution |
| `database.sqlite` | Database | SQLite |
| `dns.builtin` | DNS | Built-in DNS server |
| `ssl.letsencrypt` | SSL | Let's Encrypt |
| `monitoring.builtin` | Monitoring | Embedded Prometheus |
| `logging.builtin` | Logging | Embedded Loki |
| `auth.builtin` | Auth | Local PostgreSQL |
| `queue.redis` | Queue | Embedded Redis |

---

> **Next:** [06_DATABASE.md](./06_DATABASE.md) — Database architecture
