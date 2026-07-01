# Prerequisites

> **CloudOS v0.6**

Before deploying applications with CloudOS, you need a few tools
on your machine. This guide walks through each one.

---

## Git

Git is required to clone source repositories.

**Verify Git is installed:**

```bash
git --version
```

Expected output:

```
git version 2.40.0
```

**Install Git:**

| Platform | Command |
| :--- | :--- |
| macOS | `brew install git` |
| Ubuntu/Debian | `sudo apt install git` |
| Fedora | `sudo dnf install git` |
| Windows (WSL2) | `sudo apt install git` |

> Any modern version of Git (2.30+) will work.

---

## Docker

Docker is required to run applications in isolated containers.

**Verify Docker is installed:**

```bash
docker --version
```

Expected output:

```
Docker version 25.0.0, build abcdef1
```

**Verify Docker is running:**

```bash
docker info
```

If Docker is not running, you'll see an error like:

```
Cannot connect to the Docker daemon.
```

**Install Docker:**

| Platform | Instructions |
| :--- | :--- |
| macOS | Download [Docker Desktop for Mac](https://docs.docker.com/desktop/install/mac-install/) |
| Linux | Follow the [Docker Engine install guide](https://docs.docker.com/engine/install/) |
| Windows (WSL2) | Install Docker Desktop with [WSL2 backend](https://docs.docker.com/desktop/wsl/) |

> CloudOS falls back to the LocalRuntime if Docker is not available.
> Applications still work, but without container isolation.

**Troubleshooting:**

If you see `permission denied` when running `docker ps`, add your
user to the `docker` group:

```bash
sudo usermod -aG docker $USER
```

Then log out and back in.

---

## Go

Go is required to build Go applications. It is optional for other
stacks.

**Verify Go is installed (optional):**

```bash
go version
```

Expected output:

```
go version go1.24.0 linux/amd64
```

**Install Go:**

| Platform | Instructions |
| :--- | :--- |
| All | Download from [go.dev/dl](https://go.dev/dl/) and follow the installation guide |
| macOS | `brew install go` |
| Ubuntu/Debian | `sudo apt install golang-go` (may be outdated — prefer the official installer) |

> CloudOS requires Go 1.24 or later. Older versions will not compile
> the toolchain.

---

## Node.js

Node.js is required to build JavaScript, TypeScript, React, and
Next.js applications. It is optional for other stacks.

**Verify Node.js is installed (optional):**

```bash
node --version
```

Expected output:

```
v22.0.0
```

**Install Node.js:**

| Platform | Instructions |
| :--- | :--- |
| All | Download from [nodejs.org](https://nodejs.org/) (LTS recommended) |
| macOS | `brew install node` |
| Ubuntu/Debian | `sudo apt install nodejs npm` (use [nodesource](https://github.com/nodesource/distributions) for recent versions) |

> CloudOS works with Node.js 18+. Node 22 LTS is recommended.

---

## Python

Python is required to build Python/Flask applications. It is
optional for other stacks.

**Verify Python is installed (optional):**

```bash
python3 --version
```

Expected output:

```
Python 3.12.0
```

**Install Python:**

| Platform | Instructions |
| :--- | :--- |
| macOS | `brew install python@3.12` |
| Ubuntu/Debian | `sudo apt install python3 python3-pip python3-venv` |
| Fedora | `sudo dnf install python3 python3-pip` |
| Windows (WSL2) | `sudo apt install python3 python3-pip python3-venv` |

> Python 3.10+ is supported. Python 3.12 is recommended.

---

## PHP

PHP is required to build Laravel applications. It is optional
for other stacks.

**Verify PHP is installed (optional):**

```bash
php --version
```

Expected output:

```
PHP 8.3.0 (cli) (built: ...)
```

**Install PHP:**

| Platform | Instructions |
| :--- | :--- |
| macOS | `brew install php` |
| Ubuntu/Debian | `sudo apt install php-cli php-mbstring php-xml php-curl composer` |
| Fedora | `sudo dnf install php-cli php-mbstring php-xml php-curl composer` |
| Windows (WSL2) | `sudo apt install php-cli php-mbstring php-xml php-curl composer` |

**Verify Composer is installed:**

```bash
composer --version
```

Expected output:

```
Composer version 2.7.0 ...
```

> PHP 8.1+ and Composer 2.x are required for Laravel deployments.

---

## WSL2 (Windows Only)

If you are on Windows, CloudOS requires WSL2 (Windows Subsystem for
Linux).

**Install WSL2:**

```powershell
# In PowerShell as Administrator:
wsl --install -d Ubuntu-22.04
```

This installs WSL2 with Ubuntu 22.04 LTS.

**After installation:**

1. Launch Ubuntu from the Start menu
2. Create a Linux username and password
3. Update packages: `sudo apt update && sudo apt upgrade -y`
4. Inside WSL2, install the tools listed above

**Important notes:**

- Run CloudOS commands **inside** your WSL2 terminal, not PowerShell
- Docker Desktop must be configured with the WSL2 backend
- Files in `/mnt/c/` (Windows filesystem) are slower — keep your
  projects inside the Linux filesystem (`~/projects/`)

---

## Summary

| Toolchain | Required? | Verification Command |
| :--- | :---: | :--- |
| Git | Yes | `git --version` |
| Docker | Yes | `docker --version` + `docker info` |
| Go | For Go apps | `go version` |
| Node.js | For JS/TS apps | `node --version` |
| Python | For Python apps | `python3 --version` |
| PHP | For Laravel apps | `php --version` |

After installing all required tools, run:

```bash
cloudosctl doctor
```

This verifies everything is set up correctly.
