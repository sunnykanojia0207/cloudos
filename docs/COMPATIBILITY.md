# CloudOS Compatibility Matrix

> **Last updated:** 2026-07-01
>
> This matrix is generated from [certification test results](../tests/certification/).
> Each row represents a stack validated through the full CloudOS deployment pipeline:
>
> `Detect → Plan → Build → Artifact → Runtime → Health → Logs → Metrics`
>
> **For feature-level support** (what each stack can do beyond the certification
> pipeline), see [SUPPORT.md](SUPPORT.md). Certification status is more
> restrictive than feature support — a feature may work in the LocalRuntime
> but not yet be certified through the full pipeline.

## Legend

| Icon | Meaning |
| :--: | ------- |
|  ✅  | Certified — passes certification test |
|  ⏳  | In progress — implementation exists, not yet certified |
|  📋  | Planned — on the roadmap |
|  ❌  | Not supported |

## Certification Matrix

| Stack | Detect | Plan | Build | Runtime | Health | Logs | Metrics | Status |
| :---- | :----: | :--: | :---: | :-----: | :----: | :--: | :-----: | :----: |
| Go | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | **Certified** |
| Static | ✅ | ✅ | ✅ | ✅ | ✅ | ⌛ | ⌛ | Detection Verified |
| Node | ✅ | ✅ | ⏳ | ⏳ | ⏳ | ❌ | ❌ | Detection Verified |
| React | ✅ | ✅ | ⏳ | ⏳ | ⏳ | ❌ | ❌ | Detection Verified |
| Next.js | ✅ | ✅ | ⏳ | ⏳ | ⏳ | ❌ | ❌ | Detection Verified |
| Python | ✅ | ✅ | ⏳ | ⏳ | ⏳ | ❌ | ❌ | Detection Verified |
| Laravel | ✅ | ✅ | ⏳ | ⏳ | ⏳ | ❌ | ❌ | Detection Verified |
| Docker | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | Buildpack Only |

> **Note:** Logs and Metrics require Runtime-level support. Each Runtime
> implementation (Local, Docker, SSH, K8s) will add Logs and Metrics as
> they become certified.

## Runtimes

| Runtime | Local | Docker | SSH | Kubernetes | Status |
| :------ | :---: | :----: | :-: | :--------: | :----: |
| LocalRuntime | ✅ | — | — | — | **Active** |
| OCI Runtime (Docker) | — | ✅ | — | — | **Active** |
| SSH | — | — | 📋 | — | Planned |
| Kubernetes | — | — | — | 📋 | Planned |

> **Note:** OCI Runtime is the second Runtime implementation, proving the
> Runtime interface (ADR-0009) is stable and reusable. It uses a
> `ContainerEngine` abstraction so Docker, Podman, containerd, and nerdctl
> are interchangeable. OCI certification requires a running Docker daemon
> (`tests/certification/oci_test.go`).

## Certification Tests

Each row in the matrix maps to a certification test in
[`tests/certification/`](../tests/certification/):

| Stack | Test File | Run Command |
| :---- | :-------- | :---------- |
| Go | `go_test.go` | `go test ./tests/certification/ -run TestCertify_Go -v` |
| Static | `static_test.go` | `go test ./tests/certification/ -run TestCertify_Static -v` |
| Node | `node_test.go` | `go test ./tests/certification/ -run TestCertify_Node -v` |
| React | `react_test.go` | `go test ./tests/certification/ -run TestCertify_React -v` |
| Next.js | `nextjs_test.go` | `go test ./tests/certification/ -run TestCertify_NextJS -v` |
| Python | `python_test.go` | `go test ./tests/certification/ -run TestCertify_Python -v` |
| Laravel | `laravel_test.go` | `go test ./tests/certification/ -run TestCertify_Laravel -v` |
| OCI (Docker) | `oci_test.go` | `go test ./tests/certification/ -run TestCertify_OCI_Docker -v` |

## Becoming Certified

A buildpack or runtime becomes "CloudOS Certified" when it passes all
applicable certification tests. Certification proves:

- The stack is detected, planned, built, and deployed by the core pipeline
- Health checks confirm the application responds correctly
- The deployment can be cleaned up without leaking resources

To add a new certification:

1. Create a test file in `tests/certification/` using the `TestHarness`
2. Create sample project helpers in `harness.go`
3. Update this matrix
4. Run the test and ensure it passes

## How Tests are Run

```bash
# Run all certification tests (skips tests missing toolchains)
go test ./tests/certification/ -v -count=1

# Run certification tests in short mode (no toolchains required)
go test ./tests/certification/ -short

# Run a single stack certification
go test ./tests/certification/ -run TestCertify_Go -v -count=1
```
