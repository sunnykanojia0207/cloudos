# Installation Guide

> **CloudOS v0.6**

CloudOS turns any Git repository into a running application with complete
deployment visibility. This guide walks you through installing CloudOS on
your machine.

## What CloudOS Does

CloudOS is a developer tool that:

- Deploys applications from public or private Git repositories
- Detects the language and framework automatically (Go, Node, Python, etc.)
- Builds, deploys, and health-checks your application
- Streams live logs as the deployment progresses
- Shows a structured timeline of every deployment step
- Lets you compare deployments side-by-side
- Runs entirely on your machine — no cloud account required

## Supported Operating Systems

CloudOS runs on:

| OS | Status |
| :--- | :--- |
| **macOS** | Supported (Intel and Apple Silicon) |
| **Linux** | Supported (x86_64, aarch64) |
| **Windows** | Supported (via WSL2 — see [prerequisites](prerequisites.md)) |

> On Windows, CloudOS requires WSL2 with a Linux distribution. All
> commands and examples in this guide assume a Unix-like terminal.

## Installation Steps

### 1. Build from Source

CloudOS is distributed as source. Build the CLI binary:

```bash
git clone https://github.com/cloudos/cloudos.git
cd cloudos
go build ./tools/cloudosctl/
```

This produces `cloudosctl` in the current directory.

**Install the binary:**

```bash
# macOS / Linux
sudo mv cloudosctl /usr/local/bin/

# Windows (WSL2)
sudo mv cloudosctl /usr/local/bin/
```

### 2. Verify Installation

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

If you see a version string, CloudOS is installed and ready.

### 3. Verify the Environment

Run the doctor to check that all required toolchains are available:

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

If any checks fail, follow the links in the output to install the
missing toolchain, then run `cloudosctl doctor` again.

## Next Steps

- [Quick Start](quick-start.md) — deploy your first application in under 5 minutes
- [Prerequisites](prerequisites.md) — detailed toolchain installation guide
- [Supported Platforms](supported-platforms.md) — full platform compatibility matrix
- [Troubleshooting](troubleshooting.md) — common issues and solutions

## Uninstalling

```bash
rm /usr/local/bin/cloudosctl
rm -rf ~/.cloudos
```

That's it — CloudOS has no system services, daemons, or background
processes.
