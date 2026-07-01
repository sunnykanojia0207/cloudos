# CloudOS Provider Architecture

> **Document ID:** CLOUDOS-PROV-001  
> **Status:** v1.0 — Approved  
> **Classification:** Public — Open Source  
> **Last Updated:** 2026-06-29  
> **Audience:** Provider Developers, SDK Authors, Platform Engineers, DevOps, Plugin Developers  
> **Depends On:** [01_MASTER_SPEC.md](./01_MASTER_SPEC.md), [06_KERNEL_AND_PLUGIN_ARCHITECTURE.md](./06_KERNEL_AND_PLUGIN_ARCHITECTURE.md), [08_KERNEL.md](./08_KERNEL.md), [10_CAPABILITIES.md](./10_CAPABILITIES.md)

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [What is a Provider?](#2-what-is-a-provider)
3. [Provider Interface & SDK](#3-provider-interface--sdk)
   - [3.1 Core Provider Interface](#31-core-provider-interface)
   - [3.2 Provider Plugin SDK](#32-provider-plugin-sdk)
   - [3.3 Provider Lifecycle Handlers](#33-provider-lifecycle-handlers)
   - [3.4 Provider Context](#34-provider-context)
4. [Provider Registration & Discovery](#4-provider-registration--discovery)
   - [4.1 Registration Flow](#41-registration-flow)
   - [4.2 Capability Advertisement](#42-capability-advertisement)
   - [4.3 Provider Query & Discovery](#43-provider-query--discovery)
   - [4.4 Built-in vs. External Providers](#44-built-in-vs-external-providers)
5. [Provider Selection & Routing](#5-provider-selection--routing)
   - [5.1 Selection Strategies](#51-selection-strategies)
   - [5.2 Selection Algorithm](#52-selection-algorithm)
   - [5.3 Provider Chaining](#53-provider-chaining)
   - [5.4 Provider Weighting](#54-provider-weighting)
   - [5.5 Fallback & Failover](#55-fallback--failover)
6. [Provider Lifecycle](#6-provider-lifecycle)
   - [6.1 Lifecycle State Machine](#61-lifecycle-state-machine)
   - [6.2 Startup Sequence](#62-startup-sequence)
   - [6.3 Health Monitoring](#63-health-monitoring)
   - [6.4 Graceful Shutdown](#64-graceful-shutdown)
   - [6.5 Upgrade & Migration](#65-upgrade--migration)
7. [Provider Packaging (.cosp)](#7-provider-packaging-cosp)
   - [7.1 Archive Format](#71-archive-format)
   - [7.2 Manifest Specification](#72-manifest-specification)
   - [7.3 Configuration Schema](#73-configuration-schema)
   - [7.4 Signing & Verification](#74-signing--verification)
8. [Provider Types & Runtimes](#8-provider-types--runtimes)
   - [8.1 Native (Go) Runtime](#81-native-go-runtime)
   - [8.2 WASM Runtime](#82-wasm-runtime)
   - [8.3 HTTP Runtime](#83-http-runtime)
   - [8.4 Hosted Service Runtime](#84-hosted-service-runtime)
   - [8.5 Runtime Comparison](#85-runtime-comparison)
9. [Provider Security & Sandboxing](#9-provider-security--sandboxing)
   - [9.1 Permission Model](#91-permission-model)
   - [9.2 Sandbox Isolation](#92-sandbox-isolation)
   - [9.3 Resource Limits](#93-resource-limits)
   - [9.4 Secure Communication](#94-secure-communication)
   - [9.5 Secrets Injection](#95-secrets-injection)
10. [Provider Configuration](#10-provider-configuration)
    - [10.1 Configuration Sources](#101-configuration-sources)
    - [10.2 Configuration Schema](#102-configuration-schema)
    - [10.3 Hot-Reload](#103-hot-reload)
    - [10.4 Secrets & Sensitive Fields](#104-secrets--sensitive-fields)
11. [Provider Health](#11-provider-health)
    - [11.1 Health Check Contract](#111-health-check-contract)
    - [11.2 Health Metrics](#112-health-metrics)
    - [11.3 Degraded Modes](#113-degraded-modes)
    - [11.4 Self-Healing](#114-self-healing)
12. [Provider Development Guide](#12-provider-development-guide)
    - [12.1 Writing a Provider](#121-writing-a-provider)
    - [12.2 Testing a Provider](#122-testing-a-provider)
    - [12.3 Publishing a Provider](#123-publishing-a-provider)
13. [Built-in Provider Catalog](#13-built-in-provider-catalog)
14. [Connection to Other Documents](#14-connection-to-other-documents)

---

## 1. Executive Summary

Providers are **concrete implementations** of CloudOS capability interfaces. They are the execution layer — the actual code that runs containers, stores objects, queries databases, calls AI models, manages DNS, processes payments, and everything else CloudOS can do.

**The key architectural rule:**

> The Kernel defines capability interfaces. Providers implement them. The Kernel NEVER imports, references, or depends on any specific provider. Providers are **strangers** that implement Kernel-defined contracts.

This document defines:

1. **Provider Interface & SDK** — how a provider implements and exposes a capability
2. **Registration & Discovery** — how providers announce themselves to the Kernel
3. **Selection & Routing** — how the Kernel chooses which provider to call
4. **Lifecycle** — how providers start, run, heal, and stop
5. **Packaging** — the `.cosp` format for distributing providers
6. **Runtimes** — Native, WASM, HTTP, and Hosted provider execution models
7. **Security** — sandboxing, permissions, resource limits
8. **Development** — how to write, test, and publish a provider

---

## 2. What is a Provider?

### 2.1 Definition

A **Provider** is a process (or WASM module) that:

1. **Implements one or more capability interfaces** from [10_CAPABILITIES.md](./10_CAPABILITIES.md)
2. **Registers itself** with the Kernel's Capability Registry
3. **Accepts configuration** from the Kernel's Configuration Manager
4. **Reports health** to the Kernel's Health Manager
5. **Respects resource limits** set by the Kernel's Resource Manager
6. **Communicates via gRPC** bidirectional streams

### 2.2 What a Provider is NOT

| Is NOT | Why |
|--------|-----|
| A Kernel subsystem | Providers are external processes, not part of the Kernel |
| An application | Applications consume capabilities; providers implement them |
| A user interface | UI is in the Application Layer |
| Business logic | Business logic lives in the AI Orchestrator or applications |
| Directly accessible by users | Users go through AI → Capability → Provider |

### 2.3 Provider Identity

Every provider has a unique reverse-DNS identifier:

```
<type>.<domain>.<name>
```

| Provider | Identifier | Type |
|----------|------------|------|
| Docker Compute | `compute.docker` | Official |
| S3 Storage | `storage.s3` | Official |
| MinIO Storage | `storage.minio` | Official |
| PostgreSQL Database | `database.postgresql` | Official |
| OpenAI AI | `ai.openai` | Official |
| Stripe Billing | `billing.stripe` | Official |
| Cloudflare DNS | `dns.cloudflare` | Official |

**Provider types:**

| Type | Identifier Prefix | Source | Review Level |
|------|------------------|--------|-------------|
| **Built-in** | `*.builtin` | Ships with CloudOS Binary | Full Kernel review |
| **Official** | `capability.providername` | CloudOS-maintained | Full platform review |
| **Community** | `community.author.providername` | Community-contributed | Automated scanning |
| **Enterprise** | `enterprise.company.providername` | Enterprise contract | Enterprise compliance |
| **Custom** | `custom.orgname.providername` | Self-developed | User assumes risk |

### 2.4 Provider vs. Plugin vs. Capability

```
Capability  →  Abstract interface          (What)
Provider    →  Concrete implementation     (How)
Plugin      →  Distribution package        (.cosp)
```

A single `.cosp` plugin can contain multiple providers:

```
storage-minio.cosp
├── provider: storage.minio       → implements Storage capability
├── provider: monitoring.minio    → implements Monitoring capability
└── manifest.yaml
```

---

## 3. Provider Interface & SDK

### 3.1 Core Provider Interface

Every provider, regardless of runtime, implements the following core interface:

```go
// Provider is the base interface all providers must implement.
// This is the contract between the Kernel and any provider.
type Provider interface {
    // Init initializes the provider with its configuration.
    // Called once at startup, before any capability methods.
    Init(ctx context.Context, config *ProviderConfig) error

    // HealthCheck returns the current health status of the provider.
    // Called periodically by the Health Manager.
    HealthCheck(ctx context.Context) (*HealthStatus, error)

    // GetCapabilities returns the list of capabilities this provider implements.
    GetCapabilities(ctx context.Context) ([]*CapabilityRegistration, error)

    // GetProviderInfo returns metadata about this provider.
    GetProviderInfo(ctx context.Context) (*ProviderInfo, error)
}

type ProviderConfig struct {
    ID             string                 // Unique instance ID
    Name           string                 // Provider name (e.g., "compute.docker")
    Version        string                 // Provider version
    Settings       map[string]interface{} // Configuration values
    Secrets        SecretReader           // Interface to read secrets
    Logger         Logger                 // Structured logger
    Metrics        MetricsRecorder        // Metrics recording interface
    ResourceLimits *ResourceLimits        // CPU, memory, disk limits
}

type CapabilityRegistration struct {
    Name     string   // e.g., "capability.compute"
    Version  string   // e.g., "1.0.0"
    Features []string // Supported optional features
}

type ProviderInfo struct {
    Name        string
    Version     string
    Description string
    Author      string
    License     string
    Homepage    string
    Runtime     string   // "native", "wasm", "http", "hosted"
    Capabilities []string
    Health      string   // "healthy", "degraded", "unhealthy"
    Uptime      time.Duration
    MemoryUsage int64
}

type HealthStatus struct {
    Status      string            // "healthy", "degraded", "unhealthy"
    Message     string            // Human-readable status description
    Metrics     map[string]float64 // Custom health metrics
    LastChecked time.Time
}
```

### 3.2 Provider Plugin SDK

The CloudOS Provider SDK provides everything needed to build a provider:

```go
import "github.com/cloudos/sdk/provider"

// 1. Define your provider struct
type DockerComputeProvider struct {
    provider.BaseProvider         // Embeds default implementations
    client     *docker.Client
    config     *DockerConfig
}

// 2. Implement the Provider interface
func (p *DockerComputeProvider) Init(ctx context.Context, cfg *provider.ProviderConfig) error {
    p.config = parseConfig(cfg.Settings)
    client, err := docker.NewClient(p.config.Endpoint)
    if err != nil {
        return fmt.Errorf("docker connection failed: %w", err)
    }
    p.client = client
    return nil
}

func (p *DockerComputeProvider) HealthCheck(ctx context.Context) (*provider.HealthStatus, error) {
    _, err := p.client.Ping(ctx)
    if err != nil {
        return &provider.HealthStatus{
            Status:  "unhealthy",
            Message: fmt.Sprintf("Docker daemon unreachable: %v", err),
        }, nil
    }
    return &provider.HealthStatus{Status: "healthy"}, nil
}

func (p *DockerComputeProvider) GetCapabilities(ctx context.Context) ([]*provider.CapabilityRegistration, error) {
    return []*provider.CapabilityRegistration{
        {
            Name:     "capability.compute",
            Version:  "1.0.0",
            Features: []string{"containers", "logs", "metrics", "exec", "volumes"},
        },
    }, nil
}

// 3. Implement the capability interface(s)
func (p *DockerComputeProvider) RunContainer(ctx context.Context, req *capability.RunContainerRequest) (*capability.RunContainerResponse, error) {
    // ... Docker-specific implementation
}

// 4. Main entry point
func main() {
    provider.Serve(&DockerComputeProvider{}, provider.ServeOptions{
        Name:    "compute.docker",
        Version: "1.0.0",
    })
}
```

### 3.3 Provider Lifecycle Handlers

Providers can optionally implement lifecycle hooks:

```go
// LifecycleHooks are optional interfaces providers can implement.
type LifecycleHooks interface {
    // PreStop is called before the provider is stopped.
    // Use for cleanup: close connections, flush buffers, etc.
    PreStop(ctx context.Context) error

    // ConfigChanged is called when the provider's configuration changes at runtime.
    ConfigChanged(ctx context.Context, oldConfig, newConfig map[string]interface{}) error

    // ResourceLimitWarning is called when the provider approaches its resource limits.
    ResourceLimitWarning(ctx context.Context, limitType string, current, max int64)
}
```

### 3.4 Provider Context

The SDK provides a rich context for providers:

```go
type ProviderContext struct {
    context.Context

    // Provider identity
    ID      string
    Name    string
    Version string

    // Capabilities this provider has registered
    Capabilities []*CapabilityRegistration

    // Kernel services accessible to the provider
    Logger         Logger
    Metrics        MetricsRecorder
    Secrets        SecretReader
    EventBus       EventPublisher   // Publish events to Kernel Event Bus
    ConfigWatcher  ConfigWatcher    // Watch config changes

    // Resource constraints
    ResourceLimits *ResourceLimits

    // Distributed tracing
    Tracer Tracer

    // Health state
    Health *HealthStatus
}
```

---

## 4. Provider Registration & Discovery

### 4.1 Registration Flow

```mermaid
sequenceDiagram
    participant PL as Plugin Runtime
    participant P as Provider Process
    participant CR as Capability Registry
    participant EB as Event Bus
    participant HM as Health Manager

    PL->>P: Spawn provider process
    P->>PL: gRPC connection established
    PL->>P: Send configuration (Init)
    P-->>PL: Init OK

    P->>PL: GetCapabilities()
    PL->>CR: Register(capability, version, providerID)

    par Validate
        CR->>CR: Check version compatibility
        CR->>CR: Detect conflicts
        CR->>CR: Store registration
    end

    CR->>EB: Publish("capability.registered", registration)
    EB-->>HM: Notify: new provider available
    HM->>P: Start health checks (every 5s)

    PL-->>P: Registration confirmed
    P->>EB: Publish("provider.ready", providerInfo)
```

### 4.2 Capability Advertisement

When a provider registers, it advertises which capabilities it implements, at which versions, and with which features:

```json
{
    "provider": "compute.docker",
    "version": "1.2.0",
    "capabilities": [
        {
            "name": "capability.compute",
            "version": "1.0.0",
            "features": ["containers", "logs", "metrics", "exec", "volumes", "networks"],
            "methods": ["RunContainer", "StopContainer", "GetContainer", "ListContainers", "GetContainerLogs", "GetContainerMetrics", "Scale", "Exec"]
        }
    ],
    "config": {
        "endpoint": {
            "type": "string",
            "required": true,
            "secret": false
        },
        "tls_verify": {
            "type": "boolean",
            "default": true
        }
    }
}
```

### 4.3 Provider Query & Discovery

Consumers (AI Orchestrator, Applications, other capabilities) discover providers through the Capability Registry:

```go
// ProviderRegistry (accessed by consumers)
type ProviderRegistry interface {
    // GetProvider returns the active provider for a capability.
    GetProvider(ctx context.Context, capability string, opts *ProviderOptions) (Provider, error)

    // ListProviders returns all providers implementing a capability.
    ListProviders(ctx context.Context, capability string) ([]*ProviderInfo, error)

    // GetProviderHealth returns the health of a specific provider.
    GetProviderHealth(ctx context.Context, providerID string) (*HealthStatus, error)

    // WatchProviderChanges returns a channel for provider registration changes.
    WatchProviderChanges(ctx context.Context) (<-chan *ProviderChangeEvent, error)
}

type ProviderOptions struct {
    Version         string    // Required capability version
    Feature         string    // Required optional feature
    SelectionStrategy string  // "primary_only", "failover", "cost_optimized", "latency_optimized"
}

type ProviderChangeEvent struct {
    Type       string    // "registered", "unregistered", "health_changed"
    ProviderID string
    Capability string
    Health     string
}
```

### 4.4 Built-in vs. External Providers

| Aspect | Built-in Provider | External Provider |
|--------|------------------|-------------------|
| **Process** | Runs inside Kernel process | Runs as separate OS process |
| **Communication** | Direct Go function call | gRPC bidirectional stream |
| **Isolation** | Process-level (same process, separate goroutines) | OS-level (separate process) |
| **Latency** | Microseconds | Milliseconds (gRPC overhead) |
| **Packaging** | Compiled into Kernel binary | `.cosp` file loaded at runtime |
| **Health impact** | Can affect Kernel if panics | Crash-isolated from Kernel |
| **Startup** | Starts with Kernel | Starts on-demand or at boot |
| **Resource limits** | Go runtime limits | cgroups / OS-level limits |

**Built-in providers** are for minimal, always-required functionality:

- `storage.local` — Local filesystem storage
- `compute.local` — Local process execution
- `database.sqlite` — Embedded SQLite
- `dns.builtin` — Built-in DNS
- `auth.builtin` — Local auth backed by Kernel's State Store

**External providers** are everything else — Docker, S3, PostgreSQL, OpenAI, etc.

---

## 5. Provider Selection & Routing

### 5.1 Selection Strategies

When multiple providers implement the same capability, the Kernel must decide which one to use. This is controlled by the **selection strategy**:

```yaml
# config.yaml
capabilities:
  compute:
    providers:
      - name: compute.docker
        priority: 10
      - name: compute.firecracker
        priority: 20
      - name: compute.k8s
        priority: 30
    strategy: failover          # How to select among configured providers
    selection_criteria:         # Additional criteria for cost/latency strategies
      cost_weight: 0.6
      latency_weight: 0.4
```

| Strategy | Description | Use Case |
|----------|-------------|----------|
| `primary_only` | Always use the highest-priority provider | Development, single-provider deployments |
| `failover` | Use primary. On failure, try fallbacks in priority order | Production high-availability |
| `cost_optimized` | Select the provider with lowest cost per operation | Cost-sensitive workloads |
| `latency_optimized` | Select the provider with lowest p50/p99 latency | Performance-sensitive workloads |
| `round_robin` | Distribute evenly across all providers | Load balancing, testing |
| `weighted_random` | Random selection weighted by priority | Gradual traffic shifting |
| `ai_optimized` | AI Orchestrator decides based on intent context | Intelligent routing (phase 2+) |

### 5.2 Selection Algorithm

```go
func (s *ProviderSelector) Select(ctx context.Context, capability string, req interface{}) (Provider, error) {
    providers := s.getProvidersForCapability(capability)
    if len(providers) == 0 {
        return nil, &CapabilityError{Code: "NO_PROVIDER_AVAILABLE", HTTPStatus: 503}
    }

    // Filter by health — only consider healthy providers
    healthy := s.filterHealthy(providers)

    if len(healthy) == 0 {
        // All providers unhealthy — try degraded mode
        healthy = s.filterDegraded(providers)
        if len(healthy) == 0 {
            return nil, &CapabilityError{Code: "ALL_PROVIDERS_UNHEALTHY", HTTPStatus: 503}
        }
    }

    // Filter by version compatibility
    compatible := s.filterVersion(healthy, req)

    switch s.strategy {
    case "primary_only":
        return compatible[0], nil

    case "failover":
        return s.selectFailover(compatible)

    case "cost_optimized":
        return s.selectByCost(compatible, req)

    case "latency_optimized":
        return s.selectByLatency(compatible)

    case "round_robin":
        return s.selectRoundRobin(compatible)

    case "weighted_random":
        return s.selectWeightedRandom(compatible)

    case "ai_optimized":
        return s.selectByAI(ctx, compatible, req)

    default:
        return compatible[0], nil
    }
}
```

### 5.3 Provider Chaining

Provider chaining allows operations to flow through multiple providers. This enables patterns like:

1. **Cache → Primary** — Check local cache, then fall back to cloud
2. **Primary → Replica** — Write to primary, read from replicas
3. **Local → Cloud** — Use local storage, replicate to cloud for backup

```yaml
capabilities:
  storage:
    chain:
      - provider: storage.local
        role: cache
        read_priority: 10
        write_policy: write_through
      - provider: storage.s3
        role: primary
        read_priority: 20
        write_policy: primary
    chain_strategy: read_local_write_through
```

**Chain strategies:**

| Strategy | Read Behavior | Write Behavior |
|----------|--------------|----------------|
| `read_local_write_through` | Read from local cache (miss → primary) | Write to primary + cache |
| `read_primary_write_all` | Read from primary | Write to all providers |
| `read_any_write_primary` | Read from any healthy provider | Write only to primary |
| `active_active` | Read from nearest provider | Write to all, conflict resolution |

**Chain implementation:**

```go
type ProviderChain struct {
    providers []*ChainLink
    strategy  ChainStrategy
}

func (c *ProviderChain) Read(ctx context.Context, req interface{}) (interface{}, error) {
    for _, link := range c.sortByReadPriority() {
        if link.CanHandleRead(req) {
            resp, err := link.Provider.Execute(ctx, req)
            if err == nil {
                return resp, nil
            }
            // Log failure, try next in chain
            link.RecordFailure(err)
        }
    }
    return nil, &CapabilityError{Code: "CHAIN_EXHAUSTED", HTTPStatus: 503}
}
```

### 5.4 Provider Weighting

For weighted strategies, each provider is assigned a weight:

```yaml
capabilities:
  ai:
    providers:
      - name: ai.openai
        weight: 60    # 60% of requests
        cost_per_1k: 0.01
      - name: ai.anthropic
        weight: 30    # 30% of requests
        cost_per_1k: 0.015
      - name: ai.ollama
        weight: 10    # 10% of requests (local)
        cost_per_1k: 0.0
    strategy: weighted_random
```

Weights are dynamic — a provider's effective weight decreases as its latency or error rate increases:

```go
func (p *ProviderMetrics) EffectiveWeight() float64 {
    baseWeight := p.Config.Weight
    multiplier := 1.0

    // Reduce weight based on error rate
    if p.ErrorRate > 0.05 {
        multiplier *= (1 - p.ErrorRate)
    }

    // Reduce weight based on latency (if above threshold)
    if p.P99Latency > p.LatencyThreshold {
        multiplier *= p.LatencyThreshold / p.P99Latency
    }

    // Reduce weight if circuit breaker is open
    if p.CircuitBreaker.IsOpen() {
        multiplier *= 0.1
    }

    return baseWeight * multiplier
}
```

### 5.5 Fallback & Failover

When the primary provider fails, the Kernel automatically falls back:

```yaml
capabilities:
  ai:
    primary: ai.openai
    fallback:
      - provider: ai.anthropic
        delay: "1s"             # Wait 1s before trying fallback
        degrade_after: "3s"     # Report degraded after 3s
      - provider: ai.gemini
        delay: "5s"
        degrade_after: "10s"
      - provider: ai.ollama
        delay: "15s"            # Last resort — local model
        degrade_after: "30s"
    circuit_breaker:
      error_threshold: 5        # Consecutive errors before opening
      recovery_timeout: "30s"   # Time before half-open retry
      half_open_max: 3          # Max requests in half-open state
```

**Fallback flow:**

```mermaid
graph TD
    A[Request] --> B{Primary Healthy?}
    B -->|Yes| C[Execute on Primary]
    B -->|No| D{Circuit Breaker Open?}
    D -->|No| C
    D -->|Yes| E[Try Fallback 1]
    E -->|Success| F[Return Response]
    E -->|Fail| G[Try Fallback 2]
    G -->|Success| F
    G -->|Fail| H[Try Fallback N]
    H -->|Success| F
    H -->|All Fail| I[Return Error]
    C -->|Error| J{Retryable?}
    J -->|Yes & Retries Left| K[Retry with Backoff]
    K --> C
    J -->|No| D
```

---

## 6. Provider Lifecycle

### 6.1 Lifecycle State Machine

```mermaid
stateDiagram-v2
    [*] --> DISCOVERED: Plugin file found

    DISCOVERED --> DOWNLOADING: Install triggered
    DOWNLOADING --> VERIFYING: Archive retrieved
    VERIFYING --> REJECTED: Invalid signature
    VERIFYING --> EXTRACTING: Verified

    REJECTED --> [*]

    EXTRACTING --> CONFIGURING: Files installed
    CONFIGURING --> INITIALIZING: Config validated
    INITIALIZING --> ACTIVATING: Init() success

    ACTIVATING --> REGISTERING: Capabilities ready
    REGISTERING --> ACTIVE: Registration complete

    ACTIVE --> HEALTH_CHECK: Every 5 seconds
    HEALTH_CHECK --> ACTIVE: Healthy
    HEALTH_CHECK --> DEGRADED: Latency/error threshold
    DEGRADED --> ACTIVE: Recovery
    DEGRADED --> UNHEALTHY: Persistent failure

    UNHEALTHY --> RESTARTING: Auto-restart
    RESTARTING --> INITIALIZING: Restart success
    RESTARTING --> DEAD: Max retries exceeded

    DEAD --> [*]

    ACTIVE --> UPGRADING: Version update
    UPGRADING --> CONFIGURING: New config applied
    UPGRADING --> ACTIVATING: No config change

    ACTIVE --> DEACTIVATING: Shutdown / stop
    DEACTIVATING --> UNINSTALLING: PreStop() completed
    UNINSTALLING --> UNINSTALLED: Files cleaned
    UNINSTALLED --> [*]
```

### 6.2 Startup Sequence

```mermaid
sequenceDiagram
    participant PL as Plugin Runtime
    participant P as Provider
    participant CR as Capability Registry
    participant HM as Health Manager
    participant EB as Event Bus

    PL->>PL: Resolve provider binary/WASM
    PL->>P: Spawn process / load WASM
    P->>PL: gRPC connect (handshake)

    PL->>P: Init(config)
    Note over P: Parse config, connect to external services
    P-->>PL: Init OK

    PL->>P: GetCapabilities()
    P-->>PL: [capability.compute v1.0.0, features]

    loop For each capability
        PL->>CR: Register(name, version, providerID, features)
        CR-->>PL: Registration ID
    end

    PL->>P: Start()  // Start serving
    P->>EB: Publish("provider.started", info)
    P-->>PL: Start OK

    HM->>P: HealthCheck()
    P-->>HM: Status: healthy
    HM->>EB: Publish("provider.healthy", providerID)

    PL->>EB: Publish("provider.ready", providerID)
```

### 6.3 Health Monitoring

Health checks run on a configurable interval (default 5 seconds):

```go
type HealthCheckRequest struct {
    // Empty — health check takes no parameters
}

type HealthCheckResponse struct {
    Status      string            // "healthy", "degraded", "unhealthy"
    Message     string            // Human-readable message
    Uptime      time.Duration
    Version     string
    Metrics     map[string]float64 // Custom metrics
    LastChecked time.Time
}
```

**Health state transitions:**

| Current State | Health Check Result | Next State |
|--------------|-------------------|------------|
| Active | Healthy (response < 1s) | Active |
| Active | Degraded (response 1-5s OR error rate > 5%) | Degraded |
| Active | Unhealthy (response > 5s OR error) | Unhealthy |
| Degraded | Healthy (< 1s, no errors) | Active |
| Degraded | Degraded again | Degraded |
| Degraded | Unhealthy (3 consecutive) | Unhealthy |
| Unhealthy | Any | Restarting |

**Escalation chain:**

```
Provider unhealthy (3 consecutive)
  → Restart provider (max 3 retries)
    → Still unhealthy
      → Escalate to Kernel Health Manager
        → Escalate to AI Orchestrator
          → AI decides: notify admin, switch provider, or degrade
```

### 6.4 Graceful Shutdown

```mermaid
sequenceDiagram
    participant USER as User / System
    participant PL as Plugin Runtime
    participant P as Provider
    participant CR as Capability Registry
    participant EB as Event Bus

    USER->>PL: Stop provider

    PL->>PL: Block new requests
    PL->>EB: Publish("provider.draining", providerID)

    PL->>P: PreStop()
    Note over P: Close connections, flush buffers
    P-->>PL: PreStop OK

    PL->>P: Stop()
    Note over P: Stop accepting requests
    P-->>PL: Stop OK

    PL->>CR: Unregister(capabilities)
    PL->>EB: Publish("provider.stopped", providerID)

    PL->>PL: Wait for drain (max 30s)
    PL->>P: SIGTERM (or WASM teardown)
    PL->>EB: Publish("provider.gone", providerID)
```

**Drain phases:**

| Phase | Duration | Action |
|-------|----------|--------|
| 1. Drain | 15 seconds | Stop new requests, finish in-flight |
| 2. PreStop | 10 seconds | Call PreStop() lifecycle hook |
| 3. Stop | 5 seconds | Call Stop() to terminate |
| 4. Force | 5 seconds | SIGKILL / force terminate |

### 6.5 Upgrade & Migration

```yaml
# Upgrade policy in manifest
spec:
  upgrade:
    strategy: rolling        # rolling, blue_green, recreate
    max_unavailable: 1       # Max providers unavailable during upgrade
    min_ready: 1             # Min providers must be healthy
    health_check_after: 5s   # Wait before health check after upgrade
```

**Upgrade flow:**

```mermaid
graph TD
    A[Start Upgrade] --> B{Strategy?}
    B -->|Rolling| C[Start new provider v2]
    B -->|Blue-Green| D[Start full v2 stack]
    B -->|Recreate| E[Stop v1, Start v2]

    C --> F[Health check v2]
    F -->|Healthy| G[Route traffic to v2]
    F -->|Unhealthy| H[Rollback to v1]

    D --> I[Health check v2]
    I -->|Healthy| J[Switch traffic to v2]
    I -->|Unhealthy| K[Keep v1, discard v2]

    G --> L[Drain & stop v1]
    J --> L
    E --> L

    L --> M[Complete]
    H --> M
    K --> M
```

---

## 7. Provider Packaging (.cosp)

### 7.1 Archive Format

Providers are distributed as `.cosp` (CloudOS Plugin) files — a standard `.tar.gz` archive:

```
provider-name-1.2.0.cosp
├── manifest.yaml              # Required: provider metadata, capabilities, permissions
├── provider.wasm              # WASM binary (or provider binary for native)
│   OR
├── provider-linux-amd64       # Native Linux binary
├── provider-darwin-amd64      # macOS binary (optional)
├── provider-windows-amd64.exe # Windows binary (optional)
├── config.schema.json         # JSON Schema for provider configuration
├── permissions.yaml           # Required: declared permissions
├── signature.sig              # GPG signature (required for marketplace)
├── checksums.txt              # SHA-256 checksums
├── ui/                        # Optional: dashboard UI panels
│   ├── panel.js
│   └── panel.css
└── assets/                    # Icons, screenshots, documentation
    ├── icon.svg
    ├── screenshot-1.png
    └── README.md
```

### 7.2 Manifest Specification

```yaml
# manifest.yaml
apiVersion: cloudos.io/v1
kind: Provider

metadata:
  # Identity
  name: compute.docker
  displayName: Docker Compute
  version: 1.2.0
  author:
    name: CloudOS Team
    email: team@cloudos.io
    url: https://cloudos.io/authors/cloudos-team
  license: MIT
  description: >
    Docker provider for CloudOS Compute capability.
    Supports containers, logs, metrics, exec, volumes, and networks.
  tags:
    - compute
    - docker
    - containers
  icon: assets/icon.svg
  screenshots:
    - assets/screenshot-1.png
  homepage: https://github.com/cloudos/plugins/compute-docker
  repository: https://github.com/cloudos/plugins/compute-docker
  documentation: https://docs.cloudos.io/plugins/compute-docker

spec:
  # Runtime: native, wasm, http, hosted
  runtime: native

  # Provider type: builtin, official, community, enterprise, custom
  providerType: official

  # Entry point
  entrypoint: provider-linux-amd64

  # Capabilities this provider implements
  capabilities:
    - name: capability.compute
      version: ">=1.0.0, <2.0.0"
      features:
        - containers
        - logs
        - metrics
        - exec
        - volumes
        - networks

  # Configuration schema reference
  config:
    schema: config.schema.json
    defaults:
      endpoint: "unix:///var/run/docker.sock"
      tls_verify: false

  # Declared permissions (user must approve at install time)
  permissions:
    - network:outbound: ["*:443", "*:80"]
    - network:inbound: []
    - fs:read: ["/var/run/docker.sock", "/etc/docker"]
    - fs:write: []
    - process:exec: ["/usr/bin/docker"]
    - capability:register: ["capability.compute"]
    - event:publish: ["events.compute.*"]
    - event:subscribe: ["events.config.*", "events.health.*"]
    - secrets:read: ["compute/docker/*"]

  # Resource limits
  resources:
    cpu: "0.5"
    memory: "256Mi"
    disk: "512Mi"

  # Lifecycle configuration
  lifecycle:
    startup: 30s
    healthInterval: 5s
    healthTimeout: 5s
    shutdown: 15s

  # Dependencies
  dependencies:
    kernel: ">=0.1.0, <1.0.0"
    capabilities:
      - capability.networking: ">=1.0.0"
    plugins: []

  # Upgrade strategy
  upgrade:
    strategy: rolling
    max_unavailable: 1
    min_ready: 0
```

### 7.3 Configuration Schema

Configuration schemas use JSON Schema (Draft 2020-12):

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Docker Compute Provider Configuration",
  "type": "object",
  "required": ["endpoint"],
  "properties": {
    "endpoint": {
      "type": "string",
      "description": "Docker daemon endpoint",
      "default": "unix:///var/run/docker.sock",
      "examples": ["unix:///var/run/docker.sock", "tcp://localhost:2375"]
    },
    "tls_verify": {
      "type": "boolean",
      "description": "Enable TLS verification for TCP connections",
      "default": false
    },
    "tls_cert_path": {
      "type": "string",
      "description": "Path to TLS certificate",
      "secret": true
    },
    "tls_key_path": {
      "type": "string",
      "description": "Path to TLS key",
      "secret": true
    },
    "default_resource_limits": {
      "type": "object",
      "properties": {
        "cpu": { "type": "string", "default": "1" },
        "memory": { "type": "string", "default": "512Mi" }
      }
    },
    "registry_mirrors": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Container registry mirrors"
    }
  }
}
```

Fields marked `"secret": true` are automatically routed to the Secrets Manager — they are never stored in plaintext configuration files.

### 7.4 Signing & Verification

All providers in the marketplace MUST be signed:

```bash
# Sign a provider package
cloudos plugin sign --key ~/.cloudos/signing.key storage.minio.cosp

# Verify a provider package
cloudos plugin verify storage.minio.cosp
# Output: Signature valid (signed by CloudOS Team <team@cloudos.io>)
```

**Signing process:**

```mermaid
graph LR
    A[Provider .cosp] --> B[SHA-256 all files]
    B --> C[checksums.txt]
    A --> D[GPG sign checksums.txt]
    D --> E[signature.sig]

    F[Install] --> G[Verify checksums.txt]
    G --> H[Verify signature.sig]
    H --> I{Valid?}
    I -->|Yes| J[Allow install]
    I -->|No| K[Reject]
```

**Verification levels:**

| Level | What is Verified | Used For |
|-------|-----------------|----------|
| **None** | Nothing | Development builds |
| **Checksum** | SHA-256 integrity | Local installs, dev testing |
| **Signature** | GPG signature | Official marketplace |
| **Attestation** | Build provenance | Enterprise / compliance |

---

## 8. Provider Types & Runtimes

### 8.1 Native (Go) Runtime

**Best for:** Built-in providers, official providers, performance-critical providers

Native providers are compiled Go programs that run as separate OS processes and communicate with the Kernel via gRPC.

**Communication:**

```
Kernel ←→ gRPC bidirectional stream ←→ Native Provider Process
```

**Characteristics:**

| Property | Value |
|----------|-------|
| **Language** | Go (primary), Rust, C (via FFI) |
| **Startup** | < 100ms (Go binary) |
| **Latency** | < 1ms (gRPC local) |
| **Isolation** | Process-level (separate PID) |
| **Resource limits** | OS cgroups / RLIMIT |
| **Communication** | gRPC over Unix socket (local) or TCP |
| **Memory** | ~10-50 MB per provider |
| **Security** | OS-level process isolation |

**Native provider SDK example:**

```go
package main

import (
    "github.com/cloudos/sdk/provider"
    "github.com/cloudos/capability"
)

type MyProvider struct {
    provider.BaseProvider
    // Provider-specific state
}

func (p *MyProvider) Init(ctx context.Context, cfg *provider.ProviderConfig) error {
    // Initialize connections, parse config, set up state
    return nil
}

func (p *MyProvider) HealthCheck(ctx context.Context) (*provider.HealthStatus, error) {
    return &provider.HealthStatus{Status: "healthy"}, nil
}

func (p *MyProvider) GetCapabilities(ctx context.Context) ([]*provider.CapabilityRegistration, error) {
    return []*provider.CapabilityRegistration{
        {Name: "capability.storage", Version: "1.0.0", Features: []string{"buckets", "objects"}},
    }, nil
}

// Implement capability interfaces
func (p *MyProvider) CreateBucket(ctx context.Context, req *capability.CreateBucketRequest) (*capability.BucketInfo, error) {
    // Implementation
    return &capability.BucketInfo{Name: req.Name}, nil
}

func main() {
    provider.Serve(&MyProvider{}, provider.ServeOptions{
        Name:    "storage.custom",
        Version: "1.0.0",
    })
}
```

### 8.2 WASM Runtime

**Best for:** Community providers, sandboxed third-party code, multi-language support

WASM providers run inside the Kernel's WASM runtime (wasmer/wasmtime). They are fully sandboxed and have restricted access to system resources.

**Communication:**

```
Kernel WASM Runtime ←→ WASM Module (memory-sandboxed)
```

**Characteristics:**

| Property | Value |
|----------|-------|
| **Languages** | Rust, Go (tinygo), AssemblyScript, C, C++, Python |
| **Startup** | < 5ms (cold), < 1ms (hot) |
| **Latency** | < 100μs (in-process) |
| **Isolation** | Memory-sandboxed, no syscalls |
| **Resource limits** | WASM runtime enforced |
| **Communication** | Direct function calls via WASM exports |
| **Memory** | ~1-10 MB per provider |
| **Security** | Highest — no filesystem, no network by default |

**WASM provider example (Rust):**

```rust
use cloudos_sdk::*;

#[derive(Default)]
struct HelloProvider;

impl Provider for HelloProvider {
    fn init(&mut self, config: &Config) -> Result<()> {
        Ok(())
    }

    fn health_check(&self) -> Result<HealthStatus> {
        Ok(HealthStatus::healthy())
    }

    fn get_capabilities(&self) -> Result<Vec<CapabilityRegistration>> {
        Ok(vec![CapabilityRegistration::new("capability.compute", "1.0.0")])
    }
}

impl ComputeCapability for HelloProvider {
    fn run_container(&self, req: &RunContainerRequest) -> Result<RunContainerResponse> {
        Ok(RunContainerResponse {
            id: uuid::Uuid::new_v4().to_string(),
            name: req.name.clone(),
            status: "running".into(),
            ..Default::default()
        })
    }
}

register_provider!(HelloProvider, "compute.hello", "1.0.0");
```

### 8.3 HTTP Runtime

**Best for:** Enterprise integrations, legacy systems, external APIs

HTTP providers are remote services that expose a REST/gRPC API. The Kernel communicates with them over the network.

**Communication:**

```
Kernel ←→ HTTP/gRPC ←→ Remote HTTP Service
```

**Characteristics:**

| Property | Value |
|----------|-------|
| **Language** | Any (HTTP server) |
| **Startup** | N/A (external) |
| **Latency** | Network-dependent (1-100ms) |
| **Isolation** | Network-level |
| **Resource limits** | None (external) |
| **Communication** | HTTP/gRPC over TCP |
| **Memory** | N/A |
| **Security** | TLS + API keys |

**HTTP provider configuration:**

```yaml
providers:
  - name: billing.enterprise
    type: http
    config:
      base_url: "https://billing.internal.example.com"
      timeout: 30s
      headers:
        Authorization: "Bearer ${BILLING_API_KEY}"
      tls:
        ca_cert_path: "/etc/cloudos/certs/internal-ca.pem"
```

### 8.4 Hosted Service Runtime

**Best for:** Cloud-managed providers, SaaS integrations

Hosted service providers are managed by the CloudOS team and run on CloudOS infrastructure. The Kernel connects to them via a lightweight agent.

**Communication:**

```
Kernel ←→ gRPC ←→ CloudOS Agent ←→ CloudOS Hosted Service
```

### 8.5 Runtime Comparison

| Criterion | Native | WASM | HTTP | Hosted |
|-----------|--------|------|------|--------|
| **Performance** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |
| **Security** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Startup speed** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Resource efficiency** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Language flexibility** | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Isolation** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Ease of development** | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Debugging** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Best for** | Official/built-in providers | Community/third-party | Enterprise/legacy | Cloud-managed |

---

## 9. Provider Security & Sandboxing

### 9.1 Permission Model

Every provider declares the permissions it requires in its manifest. Users approve these at install time.

```yaml
# permissions.yaml
permissions:
  # Network: which hosts and ports the provider can connect to
  network:
    outbound:
      - host: "*"
        port: 443
        description: "HTTPS API calls"
      - host: "api.docker.com"
        port: 443
    inbound:
      - port: 8080
        description: "Health check endpoint"

  # Filesystem: which paths the provider can access
  filesystem:
    read:
      - path: "/var/run/docker.sock"
        description: "Docker daemon socket"
      - path: "/etc/docker"
        description: "Docker configuration"
    write: []

  # Process: which executables the provider can run
  process:
    exec:
      - path: "/usr/bin/docker"
        args: ["*"]

  # Capabilities: which capabilities this provider registers
  capabilities:
    register:
      - "capability.compute"

  # Events: which event bus subjects the provider accesses
  events:
    publish:
      - "events.compute.**"
    subscribe:
      - "events.config.*"
      - "events.health.*"

  # Secrets: which secret paths the provider can read
  secrets:
    read:
      - path: "compute/docker/*"
```

**Permission levels:**

| Level | Icon | Description | Example Providers |
|-------|------|-------------|------------------|
| **None** | 🔒 | No external access | `time.builtin`, `math.builtin` |
| **Restricted** | 🔐 | Declared specific access | `storage.minio`, `database.sqlite` |
| **Full** | 🛡️ | Broad access within domain | `compute.docker`, `storage.s3` |
| **System** | ⚡ | Kernel-level access | Built-in providers only |

### 9.2 Sandbox Isolation

**WASM sandbox:**

```
┌──────────────────────────────────────────────┐
│               Kernel Process                  │
│  ┌────────────────────────────────────────┐   │
│  │         WASM Runtime (wasmer/wasmtime)   │   │
│  │  ┌────────────────────────────────┐     │   │
│  │  │  Provider WASM Module          │     │   │
│  │  │  - Memory sandboxed            │     │   │
│  │  │  - No syscalls                 │     │   │
│  │  │  - No filesystem (by default)  │     │   │
│  │  │  - No network (by default)     │     │   │
│  │  │  - CPU limits enforced         │     │   │
│  │  │  - Memory limits enforced      │     │   │
│  │  │  - Wall clock = provider clock │     │   │
│  │  └────────────────────────────────┘     │   │
│  └────────────────────────────────────────┘   │
│                                                │
│  ┌────────────────────────────────────────┐   │
│  │         Native Provider Process         │   │
│  │  - Separate OS process (PID)           │   │
│  │  - cgroups CPU/memory limits           │   │
│  │  - seccomp filter (restricted syscalls)│   │
│  │  - AppArmor/SELinux profile            │   │
│  │  - Capability dropping                 │   │
│  └────────────────────────────────────────┘   │
└──────────────────────────────────────────────┘
```

**Sandbox restrictions by runtime:**

| Restriction | Native | WASM | HTTP |
|------------|--------|------|------|
| No syscalls | seccomp filter | ✅ Inherent | N/A |
| No arbitrary file access | ✅ AppArmor | ✅ Inherent | ✅ |
| No arbitrary network | ✅ Network policy | ✅ Opt-in | ✅ N/A |
| No process spawning | ✅ seccomp | ✅ Inherent | ✅ |
| Memory limit | ✅ cgroups | ✅ WASM runtime | N/A |
| CPU limit | ✅ cgroups | ✅ WASM runtime | N/A |
| No kernel module loading | ✅ seccomp | ✅ Inherent | ✅ |

### 9.3 Resource Limits

```go
type ResourceLimits struct {
    CPU          string  // CPU cores: "0.5", "1", "2"
    Memory       string  // Memory: "128Mi", "512Mi", "1Gi"
    Disk         string  // Temporary disk: "256Mi", "1Gi"
    Network      string  // Bandwidth: "10mbps", "100mbps"
    MaxProcesses int32   // Max child processes (0 = default)
    MaxOpenFiles int32   // Max open file descriptors
    MaxConcurrency int32 // Max concurrent operations
}
```

**Enforcement by runtime:**

| Limit | Native | WASM | HTTP |
|-------|--------|------|------|
| CPU | cgroups CPU quota | WASM fuel metering | N/A |
| Memory | cgroups memory limit | WASM linear memory limit | N/A |
| Disk | Disk quota / tmpfs size | WASI filesystem limits | N/A |
| Network | Traffic shaping | WASI network limits | N/A |
| Processes | cgroups pids limit | ✅ Inherent (no exec) | N/A |
| Open files | rlimit nofile | WASI fd limits | N/A |
| Concurrency | Semaphore in SDK | Synchronous WASM calls | Connection pool |

### 9.4 Secure Communication

**Kernel ↔ Provider communication security:**

```
┌──────────┐         mTLS + gRPC          ┌──────────┐
│  Kernel  │ ◄──────────────────────────► │ Provider │
│          │     (with JWT auth token)    │          │
└──────────┘                               └──────────┘
```

1. **Transport:** All gRPC communications use TLS (mTLS for mutual verification)
2. **Authentication:** Each provider receives a JWT at startup, signed by the Kernel
3. **Authorization:** Every capability call carries the caller's identity and permissions
4. **Audit:** All provider operations are logged to the immutable audit chain
5. **Secrets:** Secrets are never sent in configuration — providers read them via the Secrets Manager API

### 9.5 Secrets Injection

Providers never receive secrets in configuration. Instead, they request them at runtime:

```go
func (p *MyProvider) Init(ctx context.Context, cfg *provider.ProviderConfig) error {
    // Read secret at runtime — NOT from config
    apiKey, err := cfg.Secrets.Read(ctx, "compute/docker/api_key")
    if err != nil {
        return fmt.Errorf("failed to read API key: %w", err)
    }
    p.apiKey = apiKey
    return nil
}
```

**Secrets Manager API for providers:**

```go
type SecretReader interface {
    Read(ctx context.Context, path string) (string, error)
    ReadJSON(ctx context.Context, path string, v interface{}) error
    Exists(ctx context.Context, path string) (bool, error)
}
```

**Secret paths follow provider identity:**

```
providers/<provider-name>/<secret-key>
providers/compute.docker/tls_cert
providers/compute.docker/tls_key
providers/storage.s3/access_key
providers/storage.s3/secret_key
```

---

## 10. Provider Configuration

### 10.1 Configuration Sources

Providers receive configuration from the Kernel's Configuration Manager. The config hierarchy:

1. **Default values** — From `manifest.yaml` `spec.config.defaults`
2. **Global config** — `config.yaml` at the system level
3. **Project config** — Per-project configuration overrides
4. **Environment variables** — `${VAR_NAME}` interpolation
5. **Secrets** — Automatically resolved via Secrets Manager

```yaml
# Global config.yaml
capabilities:
  storage:
    provider: storage.s3
    config:
      region: us-east-1
      bucket_prefix: cloudos-prod
      access_key: ${AWS_ACCESS_KEY_ID}     # Environment variable
      secret_key: ${AWS_SECRET_ACCESS_KEY}  # Environment variable
```

### 10.2 Configuration Schema

As defined in section 7.3, each provider ships a `config.schema.json`. The Kernel validates all configuration changes against this schema before applying them:

```go
type ConfigValidationResult struct {
    Valid   bool
    Errors  []string
    Warnings []string
}
```

**Validation is performed:**

1. At install time — Before the provider starts
2. At config change — Before hot-reload is applied
3. At startup — As part of the Init() call

### 10.3 Hot-Reload

Providers can receive configuration changes at runtime without restarting:

```go
// Provider implements ConfigWatcher interface
func (p *MyProvider) ConfigChanged(ctx context.Context, old, new map[string]interface{}) error {
    // Hot-reload logic
    newRegion := new["region"].(string)
    if newRegion != p.config.Region {
        // Close existing connection, open new one
        p.client.Close()
        client, err := NewClient(newRegion, p.config.SecretKey)
        if err != nil {
            return err
        }
        p.client = client
    }
    return nil
}
```

**Config change flow:**

```mermaid
sequenceDiagram
    participant USER as User / AI
    participant CM as Config Manager
    participant CR as Capability Registry
    participant P as Provider

    USER->>CM: Update config
    CM->>CM: Validate against schema
    CM->>CR: Get provider for capability
    CM->>P: ConfigChanged(old, new)
    P-->>CM: Apply or Error
    alt Applied
        CM->>CM: Store new config
        CM->>P: Emit config.apply event
    else Error
        CM->>USER: Config rejected (reason)
        CM->>CM: Keep old config
    end
```

### 10.4 Secrets & Sensitive Fields

Fields marked `"secret": true` in the schema are treated specially:

1. **Stored in Secrets Manager** — encrypted at rest
2. **Never logged** — masked in all output
3. **Never in config files** — replaced with `secret://path` references
4. **Access controlled** — only authorized providers can read
5. **Rotated** — support for automatic rotation

```yaml
# config.yaml (what user sees)
config:
  region: us-east-1
  access_key: secret://providers/storage.s3/access_key
  secret_key: secret://providers/storage.s3/secret_key

# After resolution (what provider receives)
config:
  region: us-east-1
  access_key: AKIAIOSFODNN7EXAMPLE
  secret_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

---

## 11. Provider Health

### 11.1 Health Check Contract

```go
type HealthChecker interface {
    HealthCheck(ctx context.Context) (*HealthStatus, error)
}

type HealthStatus struct {
    Status      string            // "healthy", "degraded", "unhealthy"
    Message     string            // Human-readable description
    Version     string            // Provider version
    Uptime      time.Duration     // How long the provider has been running
    Metrics     map[string]float64 // Custom metrics for monitoring
    Dependencies []*DependencyHealth // Health of external dependencies
    LastChecked time.Time
}

type DependencyHealth struct {
    Name   string
    Status string
    Latency time.Duration
    Error  string
}
```

**Health check response examples:**

```json
// Healthy
{ "status": "healthy", "message": "All systems operational", "uptime": "72h30m" }

// Degraded
{ "status": "degraded", "message": "Docker daemon latency elevated (2.3s)", "uptime": "72h30m" }

// Unhealthy
{ "status": "unhealthy", "message": "Docker daemon unreachable: connection refused", "uptime": "72h30m" }
```

### 11.2 Health Metrics

Providers expose metrics that the Kernel's Monitoring system collects:

```go
// Provider metrics automatically collected by the SDK
type ProviderMetrics struct {
    // Request metrics
    RequestCount    int64            // Total requests handled
    ErrorCount      int64            // Total errors
    RequestLatency  map[string]time.Duration // Per-operation latency
    Concurrency     int32            // Current concurrent requests

    // Resource metrics
    MemoryUsage     int64            // Current memory usage in bytes
    CPUUsage        float64          // Current CPU usage (0.0-1.0)
    GoroutineCount  int32            // Number of goroutines (Go only)
    OpenConnections int32            // Current open connections

    // Business metrics (provider-specific)
    CustomMetrics   map[string]float64
}
```

### 11.3 Degraded Modes

Providers can operate in degraded mode when dependencies are partially unavailable:

```go
type DegradedMode string

const (
    DegradedModeNone         DegradedMode = "none"          // Full functionality
    DegradedModeReadOnly     DegradedMode = "read_only"     // Reads only, no writes
    DegradedModeLocalOnly    DegradedMode = "local_only"    // Only cached/local data
    DegradedModeFallback     DegradedMode = "fallback"      // Using fallback provider
    DegradedModeMinimal      DegradedMode = "minimal"       // Core operations only
    DegradedModeUnavailable  DegradedMode = "unavailable"   // No operations
)
```

**Degradation triggers:**

| Trigger | New Mode | Recovery |
|---------|----------|----------|
| Primary database unreachable | `read_only` | Database restored |
| External API rate limited | `local_only` | Rate limit reset |
| Cloud provider outage | `fallback` | Primary restored |
| Resource limit reached | `minimal` | Resources freed |
| Critical dependency down | `unavailable` | Dependency restored |

### 11.4 Self-Healing

The Kernel automatically attempts to heal unhealthy providers:

| Failure Pattern | Detection | Healing Action | Max Retries |
|----------------|-----------|---------------|-------------|
| Process crash | Process exit signal | Restart process | 3 |
| Health check timeout | 3 consecutive timeouts | Restart provider | 3 |
| Memory leak | Memory > 90% limit | Restart provider | 2 |
| Deadlock | Requests stalled > 30s | Force restart | 1 |
| External dependency down | Health check reports unhealthy | Wait and retry | Infinite |
| Config error | Init() returns error | Revert to last good config | 3 |

**Self-healing escalation:**

```
Failure detected
  → Attempt 1: Restart (wait 1s)
  → Attempt 2: Restart (wait 5s)
  → Attempt 3: Restart (wait 15s)
  → All attempts failed:
    → Degrade capability (switch to fallback provider)
    → Notify AI Orchestrator
    → AI decides: notify admin, scale replacement, or accept degraded state
```

---

## 12. Provider Development Guide

### 12.1 Writing a Provider

**Step 1: Set up your project**

```bash
# Using the CloudOS CLI
cloudos provider init my-provider
cd my-provider

# Structure:
my-provider/
├── main.go              # Provider entry point
├── provider.go          # Provider implementation
├── config.go            # Configuration parsing
├── capability.go        # Capability implementation
├── config.schema.json   # Configuration schema
├── permissions.yaml     # Permission declarations
└── manifest.yaml        # Provider manifest
```

**Step 2: Implement the provider**

```go
// main.go
package main

import (
    "github.com/cloudos/sdk/provider"
)

func main() {
    provider.Serve(&MyStorageProvider{}, provider.ServeOptions{
        Name:    "storage.custom",
        Version: "1.0.0",
    })
}
```

**Step 3: Implement capability interfaces**

```go
// capability.go
func (p *MyStorageProvider) CreateBucket(ctx context.Context, req *capability.CreateBucketRequest) (*capability.BucketInfo, error) {
    // Validate request
    if !isValidBucketName(req.Name) {
        return nil, &capability.CapabilityError{
            Code:       "INVALID_BUCKET_NAME",
            Message:    "Bucket name must be 3-63 characters",
            HTTPStatus: 400,
        }
    }

    // Business logic
    bucket, err := p.client.CreateBucket(ctx, req.Name)
    if err != nil {
        return nil, fmt.Errorf("create bucket: %w", err)
    }

    // Return response
    return &capability.BucketInfo{
        Name:      bucket.Name,
        CreatedAt: bucket.CreatedAt,
    }, nil
}
```

**Step 4: Test**

```go
// provider_test.go
func TestCreateBucket(t *testing.T) {
    p := &MyStorageProvider{}
    err := p.Init(context.Background(), &provider.ProviderConfig{
        Settings: map[string]interface{}{"endpoint": "http://localhost:9000"},
    })
    require.NoError(t, err)

    bucket, err := p.CreateBucket(context.Background(), &capability.CreateBucketRequest{
        Name: "test-bucket",
    })
    require.NoError(t, err)
    assert.Equal(t, "test-bucket", bucket.Name)
}
```

**Step 5: Build and package**

```bash
# Build
cloudos provider build

# Package
cloudos provider package --output my-provider.cosp

# Sign (for marketplace)
cloudos provider sign --key ~/.cloudos/signing.key my-provider.cosp

# Test locally
cloudos provider install my-provider.cosp
cloudos provider test my-provider --capability storage --operation CreateBucket
```

### 12.2 Testing a Provider

**Unit tests:**

```go
func TestProvider_Init(t *testing.T) {
    tests := []struct {
        name    string
        config  map[string]interface{}
        wantErr bool
    }{
        {"valid config", map[string]interface{}{"endpoint": "http://localhost:9000"}, false},
        {"missing endpoint", map[string]interface{}{}, true},
        {"invalid endpoint", map[string]interface{}{"endpoint": "not-a-url"}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := &MyProvider{}
            err := p.Init(context.Background(), &provider.ProviderConfig{Settings: tt.config})
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

**Integration tests with test harness:**

```go
func TestProvider_Integration(t *testing.T) {
    // Start a test instance of the provider
    harness := provider.NewTestHarness(t)
    defer harness.Close()

    p := &MyProvider{}
    harness.RegisterProvider(p, provider.ServeOptions{Name: "storage.test", Version: "1.0.0"})

    // Test capability operations
    t.Run("CreateAndGetBucket", func(t *testing.T) {
        bucket, err := p.CreateBucket(ctx, &capability.CreateBucketRequest{Name: "test"})
        require.NoError(t, err)

        got, err := p.GetBucket(ctx, &capability.GetBucketRequest{Name: bucket.Name})
        require.NoError(t, err)
        assert.Equal(t, bucket.Name, got.Name)
    })

    t.Run("HealthCheck", func(t *testing.T) {
        status, err := p.HealthCheck(ctx)
        require.NoError(t, err)
        assert.Equal(t, "healthy", status.Status)
    })
}
```

**Conformance tests (mandatory for marketplace):**

```bash
# Run conformance test suite for storage capability
cloudos provider conformance --capability storage my-provider.cosp

# Expected output:
# ✅ CreateBucket — PASS
# ✅ PutObject — PASS
# ✅ GetObject — PASS
# ✅ DeleteObject — PASS
# ✅ ListObjects — PASS
# ❌ PresignURL — NOT IMPLEMENTED (optional, skipped)
# ✅ Conformance score: 100% (5/5 required tests passed)
```

### 12.3 Publishing a Provider

```bash
# Build, package, sign, and publish
cloudos provider publish my-provider.cosp

# This will:
# 1. Run conformance tests
# 2. Scan for vulnerabilities
# 3. Submit to marketplace registry
# 4. Notify subscribers of new version
```

**Publishing checklist:**

- [ ] All conformance tests pass
- [ ] Configuration schema is valid
- [ ] Permissions are declared
- [ ] Provider is signed
- [ ] Icon and screenshots included
- [ ] README documentation complete
- [ ] License file included
- [ ] Changelog updated

---

## 13. Built-in Provider Catalog

| Provider | Identifier | Capability | Runtime | Default? | Description |
|----------|------------|------------|---------|----------|-------------|
| Local Storage | `storage.local` | Storage | Native | ✅ | Local filesystem storage in `.cloudos/data` |
| Local Compute | `compute.local` | Compute | Native | ✅ | Local process execution with resource limits |
| SQLite | `database.sqlite` | Database | Native | ✅ | Embedded SQLite for single-node deployments |
| Built-in DNS | `dns.builtin` | DNS | Native | ✅ | Built-in DNS server for `.cloudos` domains |
| Local SSL | `ssl.letsencrypt` | Networking | Native | ✅ | Let's Encrypt automatic SSL provisioning |
| Built-in Monitoring | `monitoring.builtin` | Monitoring | Native | ✅ | Embedded Prometheus + Loki for single-node |
| Built-in Auth | `auth.builtin` | Identity | Native | ✅ | PostgreSQL-backed JWT auth |
| Built-in Queue | `messaging.builtin` | Messaging | Native | ✅ | In-process NATS JetStream for single-node |

### Marketplace Provider Catalog (Phase 2)

| Provider | Capability | Type | Runtime |
|----------|------------|------|---------|
| `storage.s3` | Storage | Official | Native |
| `storage.minio` | Storage | Official | Native |
| `storage.gcs` | Storage | Official | Native |
| `storage.r2` | Storage | Official | WASM |
| `compute.docker` | Compute | Official | Native |
| `compute.firecracker` | Compute | Official | Native |
| `compute.k8s` | Compute | Official | Native |
| `compute.fly` | Compute | Official | HTTP |
| `database.postgresql` | Database | Official | Native |
| `database.mysql` | Database | Official | Native |
| `database.mongodb` | Database | Community | WASM |
| `database.turso` | Database | Community | WASM |
| `database.planetscale` | Database | Official | HTTP |
| `ai.openai` | AI | Official | Native |
| `ai.anthropic` | AI | Official | Native |
| `ai.gemini` | AI | Official | Native |
| `ai.ollama` | AI | Official | Native |
| `ai.deepseek` | AI | Official | WASM |
| `identity.oauth` | Identity | Official | Native |
| `identity.saml` | Identity | Official | Native |
| `identity.webauthn` | Identity | Official | Native |
| `identity.ldap` | Identity | Community | Native |
| `dns.cloudflare` | DNS | Official | Native |
| `dns.route53` | DNS | Official | Native |
| `email.sendgrid` | Email | Official | HTTP |
| `email.resend` | Email | Official | HTTP |
| `email.smtp` | Email | Community | Native |
| `billing.stripe` | Billing | Official | HTTP |
| `billing.lemon` | Billing | Official | HTTP |
| `monitoring.datadog` | Monitoring | Official | HTTP |
| `monitoring.grafana` | Monitoring | Community | HTTP |
| `search.algolia` | Search | Official | HTTP |
| `search.typesense` | Search | Community | Native |
| `search.meilisearch` | Search | Community | Native |
| `messaging.nats` | Messaging | Official | Native |
| `messaging.redis` | Messaging | Community | Native |
| `messaging.kafka` | Messaging | Community | Native |

---

## 14. Connection to Other Documents

| Document | Relationship |
|----------|-------------|
| [01_MASTER_SPEC.md](./01_MASTER_SPEC.md) | Master spec defines the provider-agnostic philosophy; this document implements it |
| [06_KERNEL_AND_PLUGIN_ARCHITECTURE.md](./06_KERNEL_AND_PLUGIN_ARCHITECTURE.md) | This document expands Section 6 (Provider Layer) with full interface and lifecycle detail |
| [08_KERNEL.md](./08_KERNEL.md) | Kernel's Plugin Loader (5.12) manages the provider lifecycle defined here |
| [10_CAPABILITIES.md](./10_CAPABILITIES.md) | Providers implement the capability interfaces defined here |
| [13_PLUGIN_SYSTEM.md](./13_PLUGIN_SYSTEM.md) | Plugin system covers marketplace distribution of `.cosp` packages |
| [16_SECURITY.md](./16_SECURITY.md) | Security model covers provider isolation, sandboxing, and permission enforcement |
| [17_DEPLOYMENT.md](./17_DEPLOYMENT.md) | Deployment configurations reference provider selection and chaining |
| [18_DEVELOPER_GUIDE.md](./18_DEVELOPER_GUIDE.md) | Developer guide references provider SDK and testing patterns |

---

> **End of Document — CloudOS Provider Architecture**
