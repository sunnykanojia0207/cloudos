# Deployment Guide

> **Document:** 11_DEPLOYMENT.md
> **Status:** Draft v0.1
> **Depends On:** [03_ARCHITECTURE.md](./03_ARCHITECTURE.md)

---

## 1. Deployment Philosophy

CloudOS deploys everywhere — from a Raspberry Pi to a multi-region Kubernetes cluster. The deployment model is tiered, with each tier sharing the same core but adapting to the platform's constraints.

---

## 2. Deployment Tiers

| Tier | Platforms | Complexity | Best For |
|------|-----------|------------|----------|
| **T1: Single Binary** | Linux, macOS, Windows, RPi | Zero | Dev, homelab |
| **T2: Docker** | Docker, Podman | Low | Production (single node) |
| **T3: Docker Compose** | Docker Compose | Low | Multi-service production |
| **T4: Kubernetes** | K8s, K3s, MicroK8s | Medium | High-availability |
| **T5: Android** | Termux | Medium | Mobile operations |
| **T6: Managed** | CloudOS Cloud | Zero | Fully managed |

---

## 3. Quick Deploy (T1 — Single Binary)

```bash
# Linux / macOS
curl -fsSL https://get.cloudos.io | bash
cloudos init
cloudos start

# Docker
docker run -d \
  --name cloudos \
  -p 8080:8080 \
  -v cloudos-data:/data \
  cloudos/cloudos:latest

# Raspberry Pi
curl -fsSL https://get.cloudos.io | bash -s -- --platform rpi
```

---

## 4. Production Deployment (T3 — Docker Compose)

```yaml
# docker-compose.yml
version: "3.9"
services:
  cloudos:
    image: cloudos/cloudos:latest
    ports:
      - "8080:8080"
    volumes:
      - cloudos-data:/data
      - ./config:/etc/cloudos
    environment:
      - CLOUDOS_DB_URL=postgres://cloudos:password@postgres:5432/cloudos
      - CLOUDOS_REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  postgres:
    image: postgres:16
    volumes:
      - pg-data:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: cloudos
      POSTGRES_USER: cloudos
      POSTGRES_PASSWORD: password
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    volumes:
      - redis-data:/data
    restart: unless-stopped

volumes:
  cloudos-data:
  pg-data:
  redis-data:
```

---

## 5. Configuration

### 5.1 Configuration Sources (priority order)

1. Environment variables (highest)
2. Config file (`/etc/cloudos/cloudos.yaml`)
3. CLI flags (lowest)

### 5.2 Configuration Reference

```yaml
# /etc/cloudos/cloudos.yaml
core:
  host: "0.0.0.0"
  port: 8080
  log_level: "info"  # debug, info, warn, error

database:
  url: "postgres://user:pass@localhost:5432/cloudos"
  max_connections: 50
  migrations: true

redis:
  address: "localhost:6379"
  db: 0

auth:
  jwt_secret: ""  # Auto-generated if empty
  session_ttl: "15m"
  refresh_ttl: "7d"
  mfa_required: false

plugins:
  dir: "/var/lib/cloudos/plugins"
  auto_update: true
  allow_unsigned: false

monitoring:
  metrics_port: 9090
  traces_sample_rate: 0.1
```

---

## 6. Raspberry Pi Deployment

```bash
# 1. Install (optimized build for ARM64)
curl -fsSL https://get.cloudos.io | bash -s -- --platform rpi

# 2. Configure for low-power
cloudos config set core.log_level warn
cloudos config set plugins.auto_update false

# 3. Run with limited resources
cloudos start --memory-limit 512MB --cpu-limit 2
```

---

## 7. Android (Termux) Deployment

```bash
# 1. Install Termux from F-Droid
# 2. Install CloudOS
pkg install cloudos

# 3. Initialize
cloudos init --mobile

# 4. Run in background
cloudos start --daemon

# 5. Access via local browser
cloudos open  # Opens http://localhost:8080
```

---

## 8. Production Checklist

- [ ] PostgreSQL configured with replication
- [ ] Redis configured with persistence
- [ ] TLS certificates provisioned (auto via Let's Encrypt)
- [ ] Secrets management configured
- [ ] Backup schedule configured
- [ ] Monitoring and alerting set up
- [ ] Resource limits configured per plugin
- [ ] Firewall rules reviewed (default-deny)
- [ ] Audit logging enabled
- [ ] Regular backup testing scheduled

---

> **Next:** [12_DEVELOPER_GUIDE.md](./12_DEVELOPER_GUIDE.md) — Developer guide
