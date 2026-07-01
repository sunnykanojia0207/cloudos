# Supported Platforms

> **CloudOS v0.6**

## Operating Systems

| OS | Version | Architecture | Status |
| :--- | :--- | :--- | :--- |
| **macOS** | 13+ (Ventura) | x86_64, arm64 | Supported |
| **Linux** | Kernel 5.4+ | x86_64, aarch64 | Supported |
| **Windows** | 10/11 (via WSL2) | x86_64 | Supported |

> CloudOS does not run natively on Windows. It requires WSL2 with a
> Linux distribution (Ubuntu 22.04 LTS recommended). See the
> [prerequisites guide](prerequisites.md) for WSL2 setup instructions.

## Required Toolchains

CloudOS works with your existing development tools. The following
table shows which toolchains are required, optional, or not needed
depending on the stacks you want to deploy.

| Toolchain | Required | Used For |
| :--- | :---: | :--- |
| **Git** | Yes | Cloning source repositories |
| **Docker** | Yes | Running applications in isolated containers |
| **Go** | Optional | Building Go applications (required for Go deploys) |
| **Node.js** | Optional | Building JavaScript/TypeScript applications |
| **Python** | Optional | Building Python applications |
| **PHP** | Optional | Building Laravel/PHP applications |

CloudOS detects which toolchains are available and adapts
accordingly. You only need the toolchains for the stacks you
actually deploy.

## Certified Stacks

The following application stacks have been tested and certified
with CloudOS:

| Stack | Certification Status | Notes |
| :--- | :---: | :--- |
| **Go** | ✅ Certified | Full support — build, run, health, logs, metrics |
| **Static HTML/CSS/JS** | ✅ Detection Verified | Served via `npx serve` |
| **Node.js** | ✅ Detection Verified | npm install + npm start |
| **React (Vite)** | ✅ Detection Verified | vite build → static hosting |
| **Next.js** | ✅ Detection Verified | Build + SSR via `next start` |
| **Python (Flask)** | ✅ Detection Verified | pip install + gunicorn |
| **Laravel / PHP** | ✅ Detection Verified | composer install + artisan serve |
| **Docker** | ✅ Buildpack Only | Buildpack detects Dockerfile, but runtime pending |

## Runtimes

| Runtime | Status | Description |
| :--- | :---: | :--- |
| **LocalRuntime** | ✅ Active | Runs applications as local processes (default) |
| **OCI Runtime (Docker)** | ✅ Active | Runs applications in Docker containers |
| **SSH** | 📋 Planned | Remote deployment via SSH |
| **Kubernetes** | 📋 Planned | Cluster deployment |

The default runtime is LocalRuntime. If Docker is available and
running, the OCI Runtime is used automatically for better isolation.

## Browser Support

CloudOS dashboard is a web application that runs in any modern browser:

| Browser | Status |
| :--- | :---: |
| Chrome 120+ | ✅ Supported |
| Firefox 120+ | ✅ Supported |
| Safari 17+ | ✅ Supported |
| Edge 120+ | ✅ Supported |

The dashboard is available at `http://localhost:8080` when CloudOS
is running.
