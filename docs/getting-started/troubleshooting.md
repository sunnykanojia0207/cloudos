# Troubleshooting

> **CloudOS v0.6**

This guide covers common issues you might encounter when installing
and running CloudOS.

---

## Docker Not Running

**Problem:**

```
cloudosctl doctor

✗ Docker
```

Or when deploying:

```
Error: Cannot connect to the Docker daemon.
```

**Cause:** Docker is installed but not running.

**Solution:**

Start Docker Desktop:

| Platform | Command / Action |
| :--- | :--- |
| macOS | Open Docker Desktop from Applications |
| Linux | `sudo systemctl start docker` |
| Windows (WSL2) | Open Docker Desktop (must have WSL2 backend enabled) |

Verify Docker is running:

```bash
docker info
```

Then run `cloudosctl doctor` again.

> CloudOS can still deploy applications without Docker by falling back
> to the LocalRuntime, but container isolation will not be available.

---

## Git Not Found

**Problem:**

```
cloudosctl doctor

✗ Git
```

**Cause:** Git is not installed or not in your PATH.

**Solution:**

Install Git:

| Platform | Command |
| :--- | :--- |
| macOS | `brew install git` |
| Ubuntu/Debian | `sudo apt install git` |
| Fedora | `sudo dnf install git` |
| Windows (WSL2) | `sudo apt install git` |

Verify:

```bash
git --version
```

---

## Port Already in Use

**Problem:**

```
Error: port 8080 is already in use
```

Or the CloudOS API fails to start.

**Cause:** Another service is using CloudOS's default port (8080).

**Solution:**

Stop the service on port 8080, or configure CloudOS to use a
different port:

```bash
export CLOUDOS_API=http://localhost:8081
cloudosctl deploy https://github.com/acme/app
```

---

## Build Failed

**Problem:**

```
cloudosctl timeline my-app

✗ Build Artifact
    Build failed: exit code 1
```

**Cause:** The application failed to build. Common reasons include:

- Missing dependencies (e.g., `go.mod`, `package.json`, `requirements.txt`)
- Missing toolchain (e.g., Go not installed for a Go app)
- Build command errors (syntax errors, missing files)

**Solution:**

1. Check the deployment logs for detailed error messages:

```bash
cloudosctl logs my-app
```

2. Verify the required toolchain is installed:

```bash
cloudosctl doctor
```

3. Ensure your application has the correct build configuration:

   - **Go:** needs `go.mod` and `main.go`
   - **Node.js:** needs `package.json` with a `build` script
   - **Python:** needs `requirements.txt` and `app.py`
   - **Laravel:** needs `composer.json` and `artisan`

4. Test the build locally:

```bash
# For Go:
cd my-app && go build ./...

# For Node:
cd my-app && npm install && npm run build
```

---

## Runtime Unavailable

**Problem:**

```
cloudosctl timeline my-app

✗ Deploy Application
    Runtime unavailable: ...
```

**Cause:** The runtime could not start the application. Common
reasons include:

- Docker is installed but not running (for OCI Runtime)
- The application binary is missing or invalid
- The port allocated by the runtime is already in use

**Solution:**

1. Check Docker is running:

```bash
docker info
```

2. Check the deployment logs for runtime details:

```bash
cloudosctl logs my-app
```

3. Ensure your application listens on the port provided by the
   runtime (usually via the `PORT` environment variable).

---

## Application Starts Then Immediately Stops

**Problem:**

The deployment succeeds, but the application status shows as
`Error` or `Degraded`.

**Cause:** The health check failed. The application started but
did not respond to the health check.

**Solution:**

1. Check the logs:

```bash
cloudosctl logs my-app
```

2. Verify your application listens on the `PORT` environment
   variable, not a hardcoded port:

```go
// Go — correct:
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
http.ListenAndServe(":"+port, nil)

// Go — incorrect (will fail health checks):
http.ListenAndServe(":8080", nil)
```

3. Ensure your application has a `/health` endpoint that returns
   HTTP 200:

```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
})
```

---

## Application Not Accessible

**Problem:**

The deployment succeeded, but `cloudosctl open my-app` doesn't
open anything, or the browser shows "connection refused."

**Cause:** The application may not be listening on the expected
port, or the runtime health check may have failed.

**Solution:**

1. Check the application status:

```bash
cloudosctl ps
```

Look for the URL column. If it's empty, the application may not
be running.

2. Check the deployment report for errors:

```bash
cloudosctl timeline my-app
```

3. View the logs:

```bash
cloudosctl logs my-app
```

---

## Doctor Shows Missing Toolchain for a Stack I Don't Use

**Problem:**

```
cloudosctl doctor

✗ PHP
```

But I don't need PHP.

**Cause:** CloudOS checks for all supported toolchains. Missing
toolchains for stacks you don't use are warnings, not errors.
They won't prevent you from deploying other types of applications.

**Solution:**

You can safely ignore this warning. CloudOS only uses the
toolchains required for the application you're deploying.

If you want to suppress the check, you can still deploy without
that toolchain.

---

## Deployment Takes Too Long

**Problem:**

CloudOS deployments are taking significantly longer than expected.

**Cause:** Common causes include:

- Large repository with many commits
- Slow internet connection for cloning
- Docker image build (for Dockerfile-based apps)
- npm install downloading many packages

**Solution:**

- Use smaller sample repositories for testing
- Ensure a fast internet connection
- Use the `go-api` example for the fastest deployment experience

---

## Windows-Specific: WSL2 Issues

**Problem:**

CloudOS commands don't work in PowerShell or Command Prompt.

**Cause:** CloudOS is designed for Unix-like environments and
requires WSL2 on Windows.

**Solution:**

1. Install WSL2 (as Administrator in PowerShell):

```powershell
wsl --install -d Ubuntu-22.04
```

2. Launch Ubuntu from the Start menu
3. Run all CloudOS commands **inside** WSL2
4. Store your projects in the Linux filesystem (`~/projects/`),
   not `/mnt/c/`

---

## Still Stuck?

If the above solutions don't help, try:

1. Run `cloudosctl doctor` and check for any failures
2. Run `cloudosctl logs <my-app>` and look for error messages
3. Run `cloudosctl timeline <my-app>` for a step-by-step view

Open an issue at:
[https://github.com/cloudos/cloudos/issues](https://github.com/cloudos/cloudos/issues)
