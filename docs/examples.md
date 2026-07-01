# CloudOS Examples Guide

CloudOS ships with seven official sample applications. These are the canonical
starting point for learning the platform.

## Deploy + Open Workflow

Every example follows the same workflow:

```bash
# 1. Deploy the application
cloudosctl deploy https://github.com/cloudos-examples/go-api

# 2. CloudOS builds and deploys automatically.
#    When complete, it asks: "Open in browser? [y/N]"
#    Type 'y' to open the application URL.

# 3. Or open it later:
cloudosctl open go-api
```

The `cloudosctl open` command resolves the application URL and launches
your default browser automatically.

## Recommended Order

Deploy the examples in this order. Each builds on the previous:

### 1. Go API

**Why first?** Go is a compiled language with no runtime dependencies. The
buildpack detects `go.mod`, compiles a binary, and deploys it. No `npm install`,
no `pip install` — just `go build`.

```bash
cloudosctl deploy https://github.com/cloudos-examples/go-api
```

**What to observe:** Fastest possible compile-and-deploy cycle. Great for
verifying the core pipeline works.

### 2. Static Site

**Why second?** Static sites have zero build steps. The buildpack serves
`index.html` directly. It's the simplest possible deployment.

```bash
cloudosctl deploy https://github.com/cloudos-examples/static-site
```

**What to observe:** The fallback buildpack in action. No detection file needed.

### 3. Node.js API

**Why third?** Introduces dependency management (`npm install`) without the
complexity of a build pipeline.

```bash
cloudosctl deploy https://github.com/cloudos-examples/node-api
```

**What to observe:** `npm install` runs during the build phase. The server starts
from a plain JavaScript file with no transpilation step.

### 4. React App

**Why fourth?** Introduces a build pipeline (Vite). Source code is transpiled
into static HTML/JS/CSS output.

```bash
cloudosctl deploy https://github.com/cloudos-examples/react-app
```

**What to observe:** The buildpack detects `react` in `package.json` and runs
`npm run build`. Output goes to a `dist/` directory.

### 5. Next.js Blog

**Why fifth?** Server-side rendering with API routes. Demonstrates the most
complex detection (Next.js is checked before generic Node.js).

```bash
cloudosctl deploy https://github.com/cloudos-examples/nextjs-blog
```

**What to observe:** Next.js detection via `next` dependency. `next build`
produces both static pages and server-side routes.

### 6. Python API

**Why sixth?** Interpreted runtime with different execution model. Demonstrates
`requirements.txt` detection and `pip install`.

```bash
cloudosctl deploy https://github.com/cloudos-examples/python-api
```

**What to observe:** Python buildpack runs `pip install -r requirements.txt`,
then starts the Flask app with `python app.py`. The example includes a fallback
to stdlib `http.server` if Flask isn't available.

### 7. Laravel API

**Why seventh?** PHP runtime with artisan CLI. Demonstrates `composer.json`
detection and `composer install`.

```bash
cloudosctl deploy https://github.com/cloudos-examples/laravel-api
```

**What to observe:** Composer dependency resolution. The `artisan serve` command
starts PHP's built-in development server.

---

## Quick Reference

| Example | Stack | Detection File | Buildpack | Runtime | Build Step | Health |
|---------|-------|----------------|-----------|---------|------------|--------|
| Go API | Go | `go.mod` | Go | `go` | `go build` | `/health` |
| Static Site | HTML/CSS | (fallback) | Static | `static` | None | N/A |
| Node.js API | Node.js | `package.json` | Node | `node` | `npm install` | `/health` |
| React App | React + Vite | `package.json` (react) | React | `static` | `npm run build` | N/A |
| Next.js Blog | Next.js | `package.json` (next) | Next.js | `static` | `next build` | `/api/health` |
| Python API | Flask | `requirements.txt` | Python | `python` | `pip install` | `/health` |
| Laravel API | PHP | `composer.json` | Laravel | `php` | `composer install` | `/health` |

---

## Certification

Each example is compatible with the CloudOS certification test suite. The
certification harness creates equivalent project structures and validates
that every step of the pipeline succeeds:

1. **Detect** — the correct buildpack is selected
2. **Plan** — the build plan includes the right commands
3. **Build** — the artifact is produced without errors
4. **Runtime** — the application starts and binds to a port
5. **Health** — the application responds to HTTP requests
6. **Logs** — structured logs are emitted during deployment

See [tests/certification/](../tests/certification/) for the test suite.

---

## Adding a New Example

To contribute a new example:

1. Create a new directory under `cloudos-examples/`
2. Include the detection file for your stack (see table above)
3. Provide a `README.md` with deploy command, expected URL, and folder structure
4. Include a health endpoint where applicable
5. Ensure it deploys with a single `cloudosctl deploy` command
