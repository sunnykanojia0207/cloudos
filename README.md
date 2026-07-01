# CloudOS

**Deploy applications from Git. See everything that happens.**

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Release](https://img.shields.io/badge/release-v0.6--rc-blue)](https://github.com/cloudos/cloudos/releases)

CloudOS turns any Git repository into a running, observable application. It detects your stack, builds it, deploys it, and shows you the entire journey — from `git clone` to a health-checked URL. No cloud account required. No YAML pipelines to write. Just one command.

```bash
cloudosctl deploy https://github.com/cloudos-examples/go-api
```

---

## Why CloudOS?

Most deployment tools are black boxes. You push code and hope for the best.

CloudOS is the opposite. Every deployment is transparent:

| You get | Instead of |
|---------|-----------|
| A structured timeline of every deployment step | "Deploying..." for 30 seconds with no feedback |
| Live streaming logs — build, deploy, and runtime | Logs hidden in a file you have to find |
| Side-by-side deployment comparison | Scrolling through CI output |
| Application status at a glance | Checking five different dashboards |
| Works on your machine, no cloud account | Creating an AWS account + IAM roles + VPC |

CloudOS runs **locally** by default. Docker for isolation is optional, not required. Your data stays on your machine.

---

## Core Features

### Deploy from Git

```bash
cloudosctl deploy https://github.com/user/repo
```

CloudOS detects the stack automatically — Go, Node.js, Python, React, Next.js, Laravel, or static files. It builds the artifact, deploys it to a runtime, health-checks it, and gives you a URL.

### Watch Live Logs

```bash
cloudosctl logs my-app -f
```

Stream every log line as it happens — from `Cloning repository` to `Server listening on :31245`. Use `-n` for historical tail, `-d` to download as a file.

### View the Timeline

```bash
cloudosctl timeline my-app
```

Every deployment step is recorded: validate, clone, detect, build, deploy, health check, complete. See exactly what happened, how long each step took, and where it failed if something went wrong.

### Compare Deployments

```bash
cloudosctl compare my-app 41 42
```

Side-by-side diff of any two deployments: commit SHA, duration, health status, build result, and per-step changes. Know what changed between "works" and "doesn't work."

### Application Dashboard

```bash
cloudosctl status my-app
```

Everything about an application in one view: identity, status, health, URL, repository, buildpack, deployment summary, available commands. Supports `--watch` (live refresh) and `--json`.

### Browser Dashboard

The CloudOS web dashboard (available at `http://localhost:8080`) provides:
- Application grid with status, health, and quick actions
- Detail views with Overview, Deployments, Timeline, Logs, and Settings tabs
- Real-time log streaming with pause, search, and download
- Dark-first design, responsive layout

### Verify Your Environment

```bash
cloudosctl doctor
```

16 read-only checks across all toolchains (Git, Docker, Go, Node.js, Python, PHP, ports, runtimes, buildpacks). Every failure includes what failed, why it matters, and exactly how to fix it.

---

## Quick Start

**Prerequisites:** [Git](https://git-scm.com/) and [Docker](https://docker.com/) (optional for LocalRuntime).

```bash
# 1. Install CloudOS
curl -fsSL https://cloudos.io/install.sh | sh

# Build from source (alternative):
# git clone https://github.com/cloudos/cloudos.git
# cd cloudos && go build ./tools/cloudosctl/
# mv cloudosctl /usr/local/bin/

# 2. Verify installation
cloudosctl version

# 3. Check environment
cloudosctl doctor

# 4. Start the kernel (in another terminal)
go run ./tools/cloudos

# 5. Deploy an application
cloudosctl deploy https://github.com/cloudos-examples/go-api

# 6. Watch it build & deploy
cloudosctl logs go-api -f

# 7. Open in browser
cloudosctl open go-api

# 8. See everything about the deployment
cloudosctl status go-api

# 9. View the deployment timeline
cloudosctl timeline go-api
```

**Total time from zero to running application: ~3 minutes.**

See the [full Quick Start](docs/getting-started/quick-start.md) for detailed walkthrough.

---

## Architecture

CloudOS has a clean, layered architecture:

```
┌──────────────────────────────────────────────┐
│                    CLI                        │
│  cloudosctl — deploy, logs, status, etc.     │
├──────────────────────────────────────────────┤
│              Control Plane API                │
│  REST API — applications, logs, timelines     │
├──────────────────────────────────────────────┤
│            Application Controller             │
│  Creates workflows, tracks deployments        │
├──────────────────────────────────────────────┤
│             Workflow Engine                   │
│  Validates → Clones → Detects → Builds →     │
│  Deploys → Health-checks → Reports            │
├─────────────────────┬────────────────────────┤
│   Buildpack Engine  │     Runtime Layer       │
│   Go Buildpack      │   LocalRuntime          │
│   Node Buildpack    │   OCI Runtime (Docker)   │
│   Python Buildpack  │    (SSH, K8s planned)   │
│   React Buildpack   │                         │
│   Next.js Buildpack │                         │
│   Laravel Buildpack │                         │
│   Static Buildpack  │                         │
└─────────────────────┴────────────────────────┘
```

Three core contracts are frozen at v1.0:
- **Runtime API** `runtime.cloudos.io/v1` — how applications run
- **Buildpack API** `buildpack.cloudos.io/v1` — how stacks are detected and built
- **Workflow API** `workflow.cloudos.io/v1` — how deployments execute

---

## Certified Stacks

| Stack | Detection | Build | Run | Health | Logs | Status |
| :---- | :-------: | :---: | :-: | :----: | :--: | :----: |
| **Go** | `go.mod` | `go build` | native binary | `/health` | ✅ | **Certified** |
| **Node.js** | `package.json` | `npm install` | `node` | `/health` | ✅ | Detection Verified |
| **React (Vite)** | `package.json` (react) | `npm run build` | static hosting | N/A | — | Detection Verified |
| **Next.js** | `package.json` (next) | `next build` | SSR server | `/api/health` | ✅ | Detection Verified |
| **Python (Flask)** | `requirements.txt` | `pip install` | `python` | `/health` | ✅ | Detection Verified |
| **Laravel / PHP** | `composer.json` | `composer install` | `artisan serve` | `/health` | ✅ | Detection Verified |
| **Static** | (fallback) | none | `npx serve` | N/A | — | Detection Verified |

---

## CLI Reference

```
cloudosctl doctor                        Check environment readiness
cloudosctl logs <app>                    Show recent logs (tail 50)
cloudosctl logs <app> -f                 Stream logs in real-time
cloudosctl logs <app> -n 100            Show last 100 lines
cloudosctl logs <app> -d                 Download all logs as text
cloudosctl deploy <git-url>              Deploy from Git repository
cloudosctl ps                            List all running applications
cloudosctl open <app>                    Open application in browser
cloudosctl status <app>                  Show application dashboard
cloudosctl status <app> --json           JSON output
cloudosctl status <app> --watch          Live-updating dashboard
cloudosctl timeline <app>                Show latest deployment timeline
cloudosctl timeline <app> -n 2           Show timeline for deployment #2
cloudosctl compare <app> 41 42           Compare two deployments
```

Full reference: [CLI Reference](docs/cli/reference.md)

---

## Project Status

| Area | Status |
|------|--------|
| CLI (`cloudosctl`) | ✅ 9 commands, production-ready |
| Dashboard (React) | ✅ Application-centric, dark-first |
| Buildpack Engine | ✅ 7 certified stacks |
| Runtime — Local | ✅ Active |
| Runtime — OCI (Docker) | ✅ Active (proves contract substitutability) |
| Deployment Report | ✅ Structured per-deployment metadata |
| Live Logs — SSE stream | ✅ Real-time, tail, download |
| Deployment Timeline | ✅ Step-by-step with status per node |
| Deployment Comparison | ✅ Side-by-side diff |
| Doctor | ✅ 16 read-only checks with remediation |
| Certification Tests | ✅ 7 stacks + OCI, harness framework |
| Plugin SDK | 📋 Planned (v0.7) |
| Auth (JWT, RBAC) | 📋 Planned |
| AI Engine | 📋 Planned |

Overall maturity: **~85%** — feature complete for v0.6 Release Candidate.

---

## Roadmap

| Version | Focus | Status |
| :------ | :---- | :----- |
| **v0.6 RC** | Documentation, polish, release readiness | 🔜 In progress |
| **v0.6** | First public Release Candidate | 📋 Next |
| **v0.7** | Plugin SDK, community buildpacks | 📋 Planned |
| **v0.8** | Authentication, organizations | 📋 Planned |
| **v0.9** | Production runtimes (SSH, K8s) | 📋 Planned |
| **v1.0** | Stable release, plugin marketplace | 📋 Planned |

---

## Contributing

CloudOS is open source (MIT). Contributions welcome!

- [Development Guide](docs/DEVELOPMENT.md) — build, test, and code standards
- [Certification Tests](docs/COMPATIBILITY.md) — adding new stack certifications
- [ADR Index](adr/) — Architecture Decision Records
- [GitHub Issues](https://github.com/cloudos/cloudos/issues)

---

## License

MIT — see [LICENSE](LICENSE).
