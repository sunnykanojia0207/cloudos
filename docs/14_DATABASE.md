# Database Architecture

> **Document:** 06_DATABASE.md
> **Status:** Draft v0.1
> **Depends On:** [04_SYSTEM_DESIGN.md](./04_SYSTEM_DESIGN.md)

---

## 1. Database Strategy

CloudOS uses a **multi-database** strategy where different databases serve different purposes, each selected for its strengths.

| Database | Purpose | Deployment |
|----------|---------|------------|
| **PostgreSQL** | Primary state store, user data, project data | Embedded / External |
| **Redis** | Cache, session store, pub/sub, rate limiting | Embedded / External |
| **SQLite** | Local state, offline cache, edge nodes | Embedded only |
| **ClickHouse** | Analytics, time-series, usage metrics | Optional plugin |
| **Elasticsearch** | Full-text search, log aggregation | Optional plugin |

---

## 2. Primary Database (PostgreSQL)

### 2.1 Why PostgreSQL

- Mature, reliable, excellent ecosystem
- JSONB for flexible document storage
- Full-text search for basic search needs
- Extensions (pgvector, PostGIS, pg_cron)
- Logical replication for high availability
- Row-level security for multi-tenancy

### 2.2 Schema Organization

```
cloudos_state/
├── migrations/          # Versioned schema migrations
│   ├── 001_initial.sql
│   ├── 002_users.sql
│   └── 003_projects.sql
├── functions/           # Stored procedures
├── triggers/            # Database triggers
└── extensions/          # Required PG extensions
```

### 2.3 Core Tables

```sql
-- Organizations
CREATE TABLE organizations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    slug        TEXT UNIQUE NOT NULL,
    plan        TEXT NOT NULL DEFAULT 'free',
    settings    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Users
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    avatar_url      TEXT,
    password_hash   TEXT,
    mfa_secret      TEXT,
    mfa_enabled     BOOLEAN DEFAULT false,
    status          TEXT NOT NULL DEFAULT 'active',
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Organization membership
CREATE TABLE organization_members (
    org_id          UUID REFERENCES organizations(id),
    user_id         UUID REFERENCES users(id),
    role            TEXT NOT NULL DEFAULT 'member',
    permissions     TEXT[] DEFAULT '{}',
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (org_id, user_id)
);

-- Projects
CREATE TABLE projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    description     TEXT,
    git_repo        TEXT,
    framework       TEXT,
    build_command   TEXT,
    output_dir      TEXT,
    environment     TEXT DEFAULT 'production',
    status          TEXT DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Deployments
CREATE TABLE deployments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id),
    environment     TEXT NOT NULL DEFAULT 'production',
    status          TEXT NOT NULL DEFAULT 'pending',
    branch          TEXT NOT NULL DEFAULT 'main',
    commit_sha      TEXT,
    commit_message  TEXT,
    build_logs      JSONB,
    url             TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ
);

-- Plugin instances
CREATE TABLE plugin_instances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    plugin_id       TEXT NOT NULL,
    version         TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'inactive',
    config          JSONB DEFAULT '{}',
    enabled         BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Audit log
CREATE TABLE audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID REFERENCES organizations(id),
    user_id         UUID REFERENCES users(id),
    action          TEXT NOT NULL,
    resource_type   TEXT NOT NULL,
    resource_id     TEXT NOT NULL,
    details         JSONB DEFAULT '{}',
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_deployments_project ON deployments(project_id, created_at DESC);
CREATE INDEX idx_audit_org ON audit_log(org_id, created_at DESC);
CREATE INDEX idx_audit_user ON audit_log(user_id, created_at DESC);
CREATE INDEX idx_org_members_user ON organization_members(user_id);
CREATE INDEX idx_plugin_org ON plugin_instances(org_id, plugin_id);
```

### 2.4 Migration Strategy

- Migrations are versioned SQL files in `core/database/migrations/`
- Applied automatically on startup (or via `cloudos db migrate`)
- Rollbacks supported for the last 3 migrations
- Plugins can register their own migrations via the plugin manifest

---

## 3. Cache Layer (Redis)

### 3.1 Use Cases

| Use | Key Pattern | TTL | Eviction |
|-----|-------------|-----|----------|
| Session data | `session:{token}` | 24h | LRU |
| Rate limiting | `ratelimit:{ip}:{endpoint}` | 1m | None |
| API cache | `cache:{method}:{path}:{params}` | 5m | LRU |
| Job queue | `queue:{name}` | - | None |
| Pub/Sub | Events for real-time updates | - | - |

### 3.2 Redis Configuration

```yaml
redis:
  mode: standalone  # standalone, sentinel, cluster
  address: localhost:6379
  password: ""
  db: 0
  pool_size: 100
  min_idle: 10
  max_retries: 3
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
```

---

## 4. Local Database (SQLite)

### 4.1 Use Cases

- Offline-first operation on mobile and edge devices
- Local development without PostgreSQL dependency
- Single-node deployments where simplicity > scale

### 4.2 Sync Strategy

```
┌──────────────┐         ┌──────────────┐
│   SQLite      │         │  PostgreSQL   │
│   (Local)     │◄─────► │  (Server)     │
│              │  Sync   │              │
│  operations  │  ─────► │  operations  │
│  queue       │  ◄───── │  updates     │
└──────────────┘         └──────────────┘

Sync algorithm:
1. Local: queue all mutations with monotonic IDs
2. Online: push queued mutations to server
3. Server: apply mutations, return server state
4. Local: resolve conflicts (last-write-wins or custom)
5. Local: update local state
```

---

## 5. Database Plugin Providers

### 5.1 Managed Database Interface

```go
type DatabaseProvider interface {
    Name() string
    
    // Lifecycle
    Provision(ctx, spec) (*Database, error)
    Deprovision(ctx, id) error
    
    // Operations
    Get(ctx, id) (*Database, error)
    List(ctx, projectID) ([]*Database, error)
    
    // Connection
    ConnectionString(ctx, id) (string, error)
    Pool(ctx, id, maxConns) (*Pool, error)
    
    // Maintenance
    Backup(ctx, id) (*Backup, error)
    Restore(ctx, backupID) error
    Scale(ctx, id, spec) error
    Patch(ctx, id, version) error
}
```

### 5.2 Provider Matrix

| Provider | Engines | Best For | Status |
|----------|---------|----------|--------|
| **SQLite** | SQLite | Dev, single-node, edge | ✅ Built-in |
| **PostgreSQL** | PG | Primary, production | ✅ Plugin |
| **MySQL** | MySQL, MariaDB | LAMP, WordPress | ✅ Plugin |
| **MongoDB** | MongoDB | Document, flexible | ✅ Plugin |
| **Redis** | Redis | Cache, session, queue | ✅ Plugin |
| **Turso** | SQLite (distributed) | Edge, distributed | 🚧 Plugin |
| **Neon** | PostgreSQL (serverless) | Auto-scaling | 🚧 Plugin |
| **ClickHouse** | ClickHouse | Analytics | 🚧 Plugin |
| **PlanetScale** | MySQL (serverless) | Serverless | 🔮 Planned |
| **CockroachDB** | PostgreSQL (distributed) | Multi-region | 🔮 Planned |

### 5.3 Database Features per Provider

| Feature | SQLite | PostgreSQL | MySQL | MongoDB |
|---------|--------|------------|-------|---------|
| Backups | File copy | pg_dump/WAL | mysqldump | mongodump |
| Point-in-time | ❌ | ✅ | ✅ | ✅ |
| Read replicas | ❌ | ✅ | ✅ | ✅ |
| Sharding | ❌ | ✅ (Citus) | ✅ | ✅ |
| Full-text search | ✅ | ✅ | ✅ | ✅ (Atlas) |
| Vector search | ❌ | ✅ (pgvector) | ❌ | ✅ (Atlas) |
| Geo queries | ❌ | ✅ (PostGIS) | ✅ (SPATIAL) | ✅ (2dsphere) |
| Connection pooling | Built-in | PgBouncer | ProxySQL | Built-in |

---

## 6. Backup Strategy

### 6.1 Backup Types

| Type | Frequency | Retention | Storage |
|------|-----------|-----------|---------|
| Continuous WAL | Every 5 min | 24 hours | Local + provider |
| Daily snapshot | Daily | 30 days | Cloud storage |
| Weekly snapshot | Weekly | 90 days | Cloud storage |
| Monthly snapshot | Monthly | 1 year | Archival storage |

### 6.2 Backup Verification

- Every backup is automatically restored to a test environment
- Integrity checksum verified
- Test query execution to validate data
- Email report on backup status

---

> **Next:** [07_API.md](./07_API.md) — API design
