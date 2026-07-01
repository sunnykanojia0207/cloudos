# CloudOS CLI Reference

> **CloudOS v0.6**

The `cloudosctl` CLI is the primary interface for interacting with CloudOS.
All commands communicate with the CloudOS API at `http://localhost:8080` by
default (configurable via the `CLOUDOS_API` environment variable).

---

## Usage

```text
cloudosctl <command> [flags]
```

### Environment Variables

| Variable | Default | Description |
| :------- | :------ | :---------- |
| `CLOUDOS_API` | `http://localhost:8080` | CloudOS API server address |

---

## Commands

### cloudosctl doctor

Check environment readiness for CloudOS deployment.

```bash
cloudosctl doctor
```

Performs 14 read-only checks:

| Check | What It Verifies | Required For |
| :---- | :--------------- | :----------- |
| CloudOS CLI | CLI version | All commands |
| Git | `git --version` | All deployments |
| Docker | `docker --version` | OCI Runtime (container isolation) |
| Docker Daemon | `docker info` | OCI Runtime health |
| Go | `go version` | Go application builds |
| Node.js | `node --version` | JavaScript/TypeScript builds |
| npm | `npm --version` | Dependency management |
| Python | `python3 --version` | Python application builds |
| PHP | `php --version` | Laravel application builds |
| Composer | `composer --version` | PHP dependency management |
| Ports | Port binding checks | API server + runtime allocation |
| Runtime | Docker or LocalRuntime detection | Application execution |
| Buildpacks | Toolchain availability | Stack detection |
| Working Directory | Write permissions | CloudOS data storage |

**Output (all checks pass):**

```
Checking CloudOS Environment...

  ✓ CloudOS CLI
    Version 0.6.0-rc1

  ✓ Git
    git version 2.40.0

  ✓ Docker
    Docker version 25.0.0, build abcdef1

  ✓ Docker Daemon
    Running

  ✓ Go
    go version go1.24.0 linux/amd64

  ✓ Node.js
    v22.0.0

  ✓ npm
    10.0.0

  ✓ Python
    Python 3.12.0

  ✓ PHP
    PHP 8.3.0 (cli) (built: ...)

  ✓ Composer
    Composer version 2.7.0 ...

  ✓ Ports
    Required ports are available

  ✓ Runtime
    OCI Runtime available (Docker)

  ✓ Buildpacks
    7 buildpacks registered

  ✓ Working Directory
    /home/user/projects

Environment Status

  ✓ Ready to deploy applications
```

**Output (with failures):**

```
Checking CloudOS Environment...

  ✓ CloudOS CLI
    Version 0.6.0-rc1

  ✗ Docker
    Docker is not installed or not found in PATH.
    → Docker provides container isolation for deployed applications.
       Without Docker, applications run without isolation.
    Fix: Install Docker Desktop:
         Download from: https://docs.docker.com/desktop/
           Then run:
             cloudosctl doctor

  ✓ Git
    git version 2.40.0
  ...
```

Every failure includes:
- **What failed** — the specific check that failed
- **Why it matters** — which deployments will be affected
- **How to fix** — platform-specific installation instructions

Every failure includes:
- **What failed** — the specific check that failed
- **Why it matters** — which deployments will be affected
- **How to fix** — platform-specific installation instructions

---

### cloudosctl logs

View application logs — snapshot, stream, or download.

```bash
# Show the most recent 50 log lines (default)
cloudosctl logs <app-id>

# Show the last N lines
cloudosctl logs <app-id> -n 100

# Stream logs in real-time (SSE)
cloudosctl logs <app-id> -f

# Download all logs as a text file
cloudosctl logs <app-id> -d
```

**Flags:**

| Flag | Short | Default | Description |
| :--- | :---- | :------ | :---------- |
| `--tail` | `-n` | `50` | Number of recent log lines to show |
| `--follow` | `-f` | `false` | Stream new log events in real-time |
| `--download` | `-d` | `false` | Download logs as `<app-id>.log` |

**Examples:**

```bash
# Stream logs during deployment
cloudosctl logs go-api -f

# See what happened 30 seconds ago
cloudosctl logs go-api -n 200

# Save logs for debugging
cloudosctl logs go-api -d
```

**Output format (streaming):**

```
10:21:14 • Workflow [init] Application "go-api" created
10:21:15 • Workflow [clone] Cloning repository...
10:21:16 • Workflow [detect] Detecting runtime... Go 1.24
10:21:18 • Workflow [build] Building binary...
10:21:22 • Workflow [deploy] Deploying application...
10:21:23 • Workflow [health] Health check... HTTP 200
10:21:23 • Workflow [complete] Deployment #1 complete
10:21:24 • App              Server listening on :31245
```

Each line shows: `HH:MM:SS • Source [Step] Message`

---

### cloudosctl deploy

Deploy an application from a Git repository.

```bash
cloudosctl deploy <git-url>
```

**Arguments:**

| Argument | Description |
| :------- | :---------- |
| `git-url` | URL to a public Git repository (HTTPS) |

**Process:**

1. CloudOS validates the URL and creates an Application resource
2. The Application Controller creates a Workflow Execution
3. The workflow runs through: Validate → Clone → Detect → Build → Deploy → Health Check → Complete
4. CloudOS polls until deployment completes (up to 90 seconds)
5. On success, it prompts: "Open in browser? [y/N]"

**Examples:**

```bash
# Deploy the Go API example
cloudosctl deploy https://github.com/cloudos-examples/go-api

# Deploy a React application
cloudosctl deploy https://github.com/cloudos-examples/react-app

# Deploy your own application
cloudosctl deploy https://github.com/your-org/your-app
```

**Output:**

```
Deploying go-api from https://github.com/cloudos-examples/go-api...
Watch logs: cloudosctl logs go-api -f
✓ Deployment #1 completed (8.2s)
Open in browser? [y/N]
```

**Note:** The `cloudosctl` CLI must be able to reach the CloudOS API
(`http://localhost:8080`). Start the kernel first with `go run ./tools/cloudos`.

---

### cloudosctl ps

List all running applications.

```bash
cloudosctl ps
```

**Output:**

```
ID         PHASE      HEALTH     URL
go-api     Running    Healthy    http://localhost:31245
my-app     Running    Degraded   http://localhost:31246
static     Running    Healthy    http://localhost:31247
```

**Columns:**

| Column | Description |
| :----- | :---------- |
| `ID` | Application identifier |
| `PHASE` | Current lifecycle phase (Running, Deploying, Error, Stopped) |
| `HEALTH` | Latest health check result (Healthy, Degraded, Unhealthy) |
| `URL` | Application endpoint |

---

### cloudosctl open

Open an application in your default browser.

```bash
cloudosctl open <app-id>
```

Resolves the application URL from the CloudOS API and opens it in your
system's default browser. Prints the URL if browser launch fails.

**Platform support:**

| Platform | Browser Launch |
| :------- | :------------- |
| macOS | `open <url>` |
| Linux | `xdg-open <url>` |
| Windows | `cmd /c start <url>` |

**Examples:**

```bash
cloudosctl open go-api
# Opens http://localhost:31245 in your browser
```

---

### cloudosctl status

Show a complete application dashboard in the terminal.

```bash
cloudosctl status <app-id>
cloudosctl status <app-id> --json
cloudosctl status <app-id> --watch
```

**Flags:**

| Flag | Default | Description |
| :--- | :------ | :---------- |
| `--json` | `false` | Output as structured JSON |
| `--watch` | `false` | Live-refresh every 2 seconds (ANSI clear) |

**Output (default):**

```
  ────────────────────────────────────────────────────────────────────────────

  Application
  go-api

  ────────────────────────────────────────────────────────────────────────────

  Status
  ✓ Running

  Health
  ✓ Healthy

  URL
  http://localhost:31245

  ────────────────────────────────────────────────────────────────────────────

  Repository
  https://github.com/cloudos-examples/go-api

  Branch
  main

  Commit
  7f41ab2

  Detected Runtime
  Go

  Buildpack
  Go Buildpack

  Deployment
  #1

  Duration
  8.2 seconds

  Started
  2026-07-01 09:15:12

  Completed
  2026-07-01 09:15:20

  ────────────────────────────────────────────────────────────────────────────

  Deployment Summary

    Latest deployment: ✓ Success
    Health:           ✓ Healthy
    Warnings:         0
    Errors:           0

  ────────────────────────────────────────────────────────────────────────────

  Available Commands

    cloudosctl logs go-api -f
    cloudosctl timeline go-api
    cloudosctl open go-api
    cloudosctl compare go-api 1 2

  ────────────────────────────────────────────────────────────────────────────
```

**JSON output:**

```bash
cloudosctl status go-api --json
```

```json
{
  "name": "go-api",
  "id": "go-api",
  "phase": "Running",
  "health": "Healthy",
  "url": "http://localhost:31245",
  "repository": "https://github.com/cloudos-examples/go-api",
  "branch": "main",
  "commitSha": "7f41ab2",
  "detectedRuntime": "Go",
  "buildpack": "Go Buildpack",
  "deploymentNumber": 1,
  "duration": "8.2s",
  "startedAt": "2026-07-01T09:15:12Z",
  "completedAt": "2026-07-01T09:15:20Z",
  "warnings": 0,
  "errors": 0,
  "latestDeployment": "Success"
}
```

---

### cloudosctl timeline

Show the step-by-step timeline of a deployment.

```bash
cloudosctl timeline <app-id>
cloudosctl timeline <app-id> -n <deployment-number>
```

**Flags:**

| Flag | Short | Default | Description |
| :--- | :---- | :------ | :---------- |
| `--number` | `-n` | latest | Deployment number to view |

**Examples:**

```bash
# View the latest deployment
cloudosctl timeline go-api

# View a specific deployment
cloudosctl timeline go-api -n 3
```

**Output:**

```
Timeline: Deployment #1
Status:   Completed
Duration: 8.2 seconds
Started:  2026-07-01 09:15:12
Ended:    2026-07-01 09:15:20
Workflow: exec-abc123

Steps:
  ✓ Validate Application
    Configuration valid
  ✓ Clone Source Repository
    Cloned 42 commits
  ✓ Build Artifact
    Build completed, binary=app
  ✓ Deploy Application
    Deployed to runtime
  ✓ Health Check
    HTTP 200 OK
  ✓ Complete Deployment
    Deployment #1 complete
```

**Step status icons:**

| Icon | Meaning |
| :--: | :------ |
| ✓ | Succeeded |
| ✗ | Failed |
| ◌ | In progress |
| → | Skipped |
| ⊘ | Cancelled |

---

### cloudosctl compare

Compare two deployments side-by-side.

```bash
cloudosctl compare <app-id> <from-number> <to-number>
```

**Arguments:**

| Argument | Description |
| :------- | :---------- |
| `app-id` | Application identifier |
| `from-number` | First deployment number |
| `to-number` | Second deployment number |

**Output:**

```
Compare #41 vs #42

Changes:
  Commit
    #41  a1b2c3d
    ↓
    #42  e5f6g7h  ✓ Changed
  Duration
    #41  6.1s
    ↓
    #42  8.2s  ✓ Changed

Unchanged:
  Health   Healthy  (no change)
  Runtime  Go  (no change)
  Steps    6  (no change)
```

**What's compared:**

| Dimension | Detects |
| :-------- | :------ |
| Commit SHA | Code change |
| Duration | Performance regression |
| Health | Health status change |
| Build | Runtime/Buildpack change |
| Steps | Workflow structure change |
| Per-step | Individual node results |

---

### cloudosctl version

Show the installed CloudOS version.

```bash
cloudosctl version
```

**Output:**

```
CloudOS v0.6.0-rc1
Commit:     a1b2c3d
Built:      2026-07-01
Platform:   linux/amd64
Go version: go1.24.4
```

---

## Exit Codes

| Code | Meaning |
| :--: | :------ |
| 0 | Success |
| 1 | Error (API error, missing arguments, etc.) |

---

## Examples Cheat Sheet

```bash
# Environment
cloudosctl doctor                          # Check everything
cloudosctl version                         # Show version

# Deploy
cloudosctl deploy https://github.com/user/app  # Deploy from git

# Observe
cloudosctl logs app                        # Recent logs
cloudosctl logs app -f                     # Stream logs
cloudosctl logs app -d                     # Download logs
cloudosctl status app                      # App dashboard
cloudosctl status app --watch              # Live dashboard
cloudosctl ps                              # All apps

# Debug
cloudosctl timeline app                    # Deployment steps
cloudosctl timeline app -n 3               # Specific deployment
cloudosctl compare app 41 42              # Diff two deployments

# Access
cloudosctl open app                        # Open browser
```

---

## Full Command Summary

| Command | Description |
| :------ | :---------- |
| `doctor` | Check environment readiness (14 checks) |
| `logs` | View, stream, or download application logs |
| `deploy` | Deploy an application from a Git URL |
| `ps` | List all running applications |
| `open` | Open application in default browser |
| `status` | Show application dashboard (terminal or JSON) |
| `timeline` | Show deployment step-by-step timeline |
| `compare` | Compare two deployments side-by-side |
| `version` | Show installed CloudOS version |
| `help` | Print usage information |
