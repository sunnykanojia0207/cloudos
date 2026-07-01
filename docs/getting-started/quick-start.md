# Quick Start

> **CloudOS v0.6**
>
> Deploy an application in under 5 minutes.

This guide walks through the entire journey: from nothing to a running
application with live logs, deployment timeline, and a URL you can open
in your browser.

---

## Prerequisites

Make sure you have the following installed:

- **Git** — `git --version`
- **Docker** — `docker --version` (Docker Desktop must be running)

See the [prerequisites guide](prerequisites.md) if you need help
installing these.

---

## Step 1: Install CloudOS

```bash
curl -fsSL https://cloudos.io/install.sh | sh
```

Or, if you're building from source:

```bash
git clone https://github.com/cloudos/cloudos.git
cd cloudos
go build ./tools/cloudosctl/
sudo mv cloudosctl /usr/local/bin/
```

---

## Step 2: Verify Installation

```bash
cloudosctl version
```

Expected output:

```
CloudOS v0.6.0-rc1
Commit:     a1b2c3d
Built:      2026-06-30
Platform:   linux/amd64
Go version: go1.24.4
```

---

## Step 3: Run Doctor

Before deploying, verify your environment is ready:

```bash
cloudosctl doctor
```

Expected output (all checks pass):

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

If any checks fail, follow the instructions in the output to install the
missing toolchain, then run `cloudosctl doctor` again.

---

## Step 4: Deploy a Sample Application

CloudOS includes sample applications. Deploy the Go API example:

```bash
cloudosctl deploy https://github.com/cloudos-examples/go-api
```

You'll see output like:

```
Deploying go-api from https://github.com/cloudos-examples/go-api...
Watch logs: cloudosctl logs go-api -f
✓ Deployment #1 completed (8.2s)
Open in browser? [y/N]
```

The deployment begins automatically. CloudOS will:

1. **Validate** the application configuration
2. **Clone** the source repository
3. **Detect** the runtime (Go 1.24)
4. **Build** the binary
5. **Deploy** to the runtime
6. **Check** health (HTTP 200)
7. **Complete** the deployment

This usually takes 15–30 seconds.

---

## Step 5: Watch Live Logs

Stream the deployment and application logs in real time:

```bash
cloudosctl logs go-api -f
```

You'll see each step as it happens:

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

Press `Ctrl+C` to stop following logs.

---

## Step 6: Open the Application

Open the deployed application in your browser:

```bash
cloudosctl open go-api
```

Your browser opens to `http://localhost:31245` (or whatever port
was allocated). You should see the application's welcome page.

---

## Step 7: View Application Status

See everything about a deployment in one place:

```bash
cloudosctl status go-api
```

Expected output:

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

Use `--watch` to refresh every 2 seconds:

```bash
cloudosctl status go-api --watch
```

Or `--json` for structured output:

```bash
cloudosctl status go-api --json
```

---

## Step 8: List All Applications

See all running applications at a glance:

```bash
cloudosctl ps
```

Expected output:

```
ID         PHASE      HEALTH     URL
go-api     Running    Healthy    http://localhost:31245
```

---

## Step 9: View the Deployment Timeline

See every step of the deployment in detail:

```bash
cloudosctl timeline go-api
```

Expected output:

```
Timeline: Deployment #1
Status:   Completed
Duration: 8.2 seconds
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

---

## All Done

You've successfully deployed an application with CloudOS.

You saw:

- `cloudosctl version` — verify installation
- `cloudosctl doctor` — verify environment
- `cloudosctl deploy` — deploy from a Git URL
- `cloudosctl logs -f` — watch live logs
- `cloudosctl open` — open in browser
- `cloudosctl status` — application dashboard
- `cloudosctl ps` — list applications
- `cloudosctl timeline` — view deployment steps

### What's Next?

| Topic | Guide |
| :--- | :--- |
| Deploy your own app | `cloudosctl deploy <your-git-url>` |
| Compare deployments | `cloudosctl compare go-api 1 2` |
| Troubleshoot issues | [Troubleshooting Guide](troubleshooting.md) |
| Understand concepts | [Concepts Guide](../../concepts/overview.md) |
| Plugin SDK | Coming in v0.7 |
