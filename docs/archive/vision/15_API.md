# API Design

> **Document:** 10_API.md  
> **Depends On:** [05_SYSTEM_ARCHITECTURE.md](./05_SYSTEM_ARCHITECTURE.md)  

---

## 1. API Philosophy

CloudOS exposes three API surfaces:

| API | Primary Consumers | Protocol | Status |
|-----|-------------------|----------|--------|
| **GraphQL** | Dashboard, Mobile, Desktop, AI | HTTP/2 + WebSocket | Primary |
| **REST** | SDKs, third-party integrations | HTTP/1.1 + HTTP/2 | Compatibility |
| **gRPC** | Internal services, plugin runtime | HTTP/2 | Internal |

### Design Principles

1. **API First** — Every feature exists as an API before any UI is built
2. **Consistent patterns** — Same resource model across all three APIs
3. **Schema-driven** — GraphQL schema is the source of truth; REST and gRPC are derived
4. **Type-safe** — All APIs produce typed schemas (GraphQL SDL, OpenAPI, Protobuf)

---

## 2. GraphQL API (Primary)

### 2.1 Schema Organization

Schema is organized by domain with federation support:

```graphql
type Query {
    # Auth
    me: User!
    organization(slug: String!): Organization
    
    # Projects & Deployments
    projects(orgId: ID!): [Project!]!
    project(id: ID!): Project
    deployments(projectId: ID!, first: Int, after: String): DeploymentConnection!
    deployment(id: ID!): Deployment

    # Resources
    databases(projectId: ID!): [Database!]!
    storage(projectId: ID!): [Bucket!]!
    instances(projectId: ID!): [Instance!]!
    
    # Monitoring
    metrics(resourceId: ID!, from: DateTime!, to: DateTime!): Metrics!
    logs(filter: LogFilter!): [LogEntry!]!
    
    # Plugins
    installedPlugins: [PluginInstance!]!
    marketplace(query: String): [PluginManifest!]!
    
    # AI
    aiProviders: [AIProvider!]!
}

type Mutation {
    # Auth
    login(input: LoginInput!): AuthPayload!
    register(input: RegisterInput!): AuthPayload!
    
    # Deployments
    deploy(projectId: ID!, branch: String, env: JSON): Deployment!
    rollback(deploymentId: ID!): Deployment!
    
    # Resources
    createDatabase(input: CreateDatabaseInput!): Database!
    createBucket(input: CreateBucketInput!): Bucket!
    
    # Plugins
    installPlugin(pluginId: String!, version: String): PluginInstance!
    
    # AI
    aiQuery(input: AIQueryInput!): AIResponse!
}

type Subscription {
    deploymentStatus(projectId: ID!): DeploymentStatus!
    logs(projectId: ID!): LogEntry!
    alerts: Alert!
}
```

### 2.2 Key Patterns

- **Pagination**: Relay-style cursor pagination on all list queries
- **Errors**: Union types for error handling (never throw on validation)
- **Real-time**: Subscriptions via WebSocket for deployments, logs, and alerts
- **Authorization**: JWT in `Authorization` header, validated per-request

---

## 3. REST API (Compatibility)

Base URL: `https://api.cloudos.io/v1`

Authentication: `Authorization: Bearer <token>` or `X-API-Key: <key>`

Standard response envelope:
```json
{
    "data": { ... },
    "meta": {
        "request_id": "req_abc",
        "timestamp": "2026-06-29T12:00:00Z",
        "version": "1.0"
    }
}
```

Error response:
```json
{
    "error": {
        "code": "resource_not_found",
        "message": "Project 'proj_abc' not found",
        "details": null
    },
    "meta": { "request_id": "req_abc", "timestamp": "..." }
}
```

### Key Endpoints

```
GET    /v1/projects
POST   /v1/projects
GET    /v1/projects/:id
DELETE /v1/projects/:id
POST   /v1/projects/:id/deploy

GET    /v1/projects/:id/databases
POST   /v1/projects/:id/databases
DELETE /v1/databases/:id

GET    /v1/projects/:id/metrics
GET    /v1/projects/:id/logs

GET    /v1/plugins
POST   /v1/plugins/:id/install
```

---

## 4. gRPC API (Internal)

Services defined in Protobuf:

```protobuf
service EventService {
    rpc Publish(Event) returns (PublishResponse);
    rpc Subscribe(SubscribeRequest) returns (stream Event);
}

service PluginService {
    rpc RegisterCapability(RegisterRequest) returns (RegisterResponse);
    rpc InvokeCapability(InvokeRequest) returns (InvokeResponse);
    rpc HealthCheck(HealthRequest) returns (HealthResponse);
}
```

All internal services use mTLS, distributed tracing, and circuit breakers.

---

## 5. Rate Limiting

| Tier | GraphQL | REST | gRPC |
|------|---------|------|------|
| Free | 100 req/min | 100 req/min | 1000 req/min |
| Pro | 1000 req/min | 1000 req/min | 10000 req/min |
| Enterprise | Custom | Custom | Custom |

---

> **Next:** [11_SECURITY.md](./11_SECURITY.md) — Security architecture
