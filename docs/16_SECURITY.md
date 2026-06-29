# Security Architecture

> **Document:** 10_SECURITY.md
> **Status:** Draft v0.1
> **Depends On:** [04_SYSTEM_DESIGN.md](./04_SYSTEM_DESIGN.md)

---

## 1. Security Principles

1. **Zero Trust** — Authenticate and authorize every request, every time
2. **Least Privilege** — Every component gets minimum required permissions
3. **Defense in Depth** — Multiple overlapping security layers
4. **Secure by Default** — Secure configuration is the default
5. **Fail Secure** — On error, deny access by default

---

## 2. Authentication

### 2.1 Authentication Methods

| Method | Security Level | Use Case |
|--------|---------------|----------|
| **Email + Password** | Standard | Web app login |
| **JWT (Access + Refresh)** | Standard | API authentication |
| **API Keys** | High | Service-to-service |
| **OAuth 2.0 / OIDC** | Standard | SSO, third-party |
| **WebAuthn / Passkeys** | High | Passwordless |
| **TOTP MFA** | High | Second factor |
| **SMS MFA** | Medium | Backup MFA |
| **Magic Links** | Medium | Passwordless email |

### 2.2 Token Policy

```
Access Token:
  Type:     JWT (RS256)
  Lifetime: 15 minutes
  Storage:  Memory (web), Secure storage (mobile)
  Rotation: On refresh

Refresh Token:
  Type:     Opaque (64 bytes random)
  Lifetime: 7 days (renewable up to 30 days)
  Storage:  HTTP-only cookie (web), Secure enclave (mobile)
  Rotation: On use (token rotation)

API Key:
  Format:   cos_XXXX_XXXXXXXXXXXXXXXXXXXXXXXX (prefix + org + key)
  Lifetime: 90 days max (auto-rotatable)
  Scope:    Explicit permission set at creation
```

---

## 3. Authorization

### 3.1 RBAC Model

```
SuperAdmin:     Full system access
OrgAdmin:       Full org access, manage members
Developer:      Deploy, manage resources
Viewer:         Read-only access
Custom:         User-defined permission sets
```

### 3.2 Permission Resolution

```go
func CheckPermission(ctx, userID, orgID, resource, action string) bool {
    // 1. SuperAdmin bypass
    if user.HasRole("superadmin") {
        return true
    }

    // 2. Check explicit denies first
    if hasDeny(userID, orgID, resource, action) {
        return false
    }

    // 3. Check role-based permissions
    if roleHasPermission(user.Role, resource, action) {
        return true
    }

    // 4. Check ABAC policies
    if evaluatePolicies(user, resource, action) {
        return true
    }

    // 5. Default: deny
    return false
}
```

---

## 4. Encryption

### 4.1 At Rest

| Data Type | Encryption | Key Management |
|-----------|------------|----------------|
| Database | AES-256-GCM (tablespace) | CloudOS Key Service |
| Secrets | AES-256-GCM (per secret) | Master key (OS keychain/TPM) |
| Storage Objects | SSE-S3, SSE-C, SSE-KMS | Provider-specific |
| Audit Logs | Append-only, signed | Cryptographic chaining |

### 4.2 In Transit

- TLS 1.3 minimum for all external connections
- mTLS for internal service mesh
- WireGuard for VPN connections
- HSTS with preload

---

## 5. Audit Logging

```
┌─────────────────────────────────────────────┐
│            Audit Log Entry                    │
├─────────────────────────────────────────────┤
│ id:          "aud_20260629_abc123"          │
│ timestamp:   "2026-06-29T12:00:00Z"         │
│ actor:       { id, email, ip, user_agent }  │
│ action:      "deployment.create"            │
│ resource:    { type, id, name }             │
│ context:     { org_id, project_id }         │
│ changes:     { before, after }              │
│ status:      "success" | "failure"          │
│ prev_hash:   "sha256:..."                   │
│ signature:   "..."                          │
└─────────────────────────────────────────────┘
```

---

## 6. Vulnerability Management

- Automated scanning on every dependency update (Dependabot, Renovate)
- Weekly full vulnerability scans (Trivy, Grype)
- Quarterly penetration testing
- Responsible disclosure program (bug bounty)
- 90-day vulnerability fix SLA (critical: 7 days)

---

> **Next:** [11_DEPLOYMENT.md](./11_DEPLOYMENT.md) — Deployment guide
