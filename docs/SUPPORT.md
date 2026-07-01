# CloudOS Feature Support by Stack

> **Last updated:** 2026-07-01
>
> This document describes the **intended feature support** for each stack.
> It reflects what each stack _should_ be able to do when fully implemented.
>
> **For current certification status** (which stacks have passed the automated
> test suite through the full Detect → Plan → Build → Runtime → Health →
> Logs → Metrics pipeline), see [COMPATIBILITY.md](COMPATIBILITY.md).
>
> A feature listed as ✅ here may show ⏳ (in progress) in COMPATIBILITY.md
> if it has not yet been certified through the full pipeline.

## Go

| Feature | Status | Notes |
| :------ | :----: | :---- |
| Hot Reload | ❌ | Not yet implemented |
| Live Logs | ✅ | LocalRuntime captures stdout/stderr |
| Health Checks | ✅ | HTTP /health endpoint |
| Metrics | ✅ | CPU, memory, uptime via Runtime |
| Auto Restart | ✅ | LocalRuntime supports restart |
| Env Variables | ✅ | Runtime passes env vars |
| PORT Allocation | ✅ | Auto-allocated by Runtime.Prepare |
| Static File Serving | ❌ | Use a framework or serve static separately |

## Static (HTML/CSS/JS)

| Feature | Status | Notes |
| :------ | :----: | :---- |
| npx serve | ✅ | Served via `serve` npm package |
| Live Logs | ❌ | Static files don't produce application logs |
| Health Checks | ❌ | Static sites have no /health endpoint by default |
| PORT Allocation | ✅ | Auto-allocated |
| Custom Start Command | ❌ | Falls through to npx serve default |
| Directory Listing | ✅ | Default behavior of serve |

## Node.js

| Feature | Status | Notes |
| :------ | :----: | :---- |
| Hot Reload | ❌ | Not yet implemented |
| Live Logs | ✅ | LocalRuntime captures stdout/stderr |
| Health Checks | ✅ | Configurable per application |
| PORT Allocation | ✅ | Auto-allocated via process.env.PORT |
| npm install | ✅ | InstallCmd runs before build |
| Custom Start Script | ✅ | package.json scripts.start is detected |

## React (Vite)

| Feature | Status | Notes |
| :------ | :----: | :---- |
| Hot Reload | ❌ | Not yet implemented |
| Static Hosting | ✅ | Built to dist/, served via npx serve |
| SSR | ❌ | Client-side only |
| Live Logs | ❌ | Static output, no server logs |
| Build Output | ✅ | vite build produces dist/ |
| npm install | ✅ | Full dependency resolution |
| PORT Allocation | ✅ | Auto-allocated |

## Next.js

| Feature | Status | Notes |
| :------ | :----: | :---- |
| SSR | ✅ | Next.js SSR via next start |
| API Routes | ✅ | /api/* endpoints supported |
| Static Generation | ✅ | next build produces .next/ |
| Live Logs | ✅ | Node.js process logs |
| Health Checks | ✅ | /api/health endpoint |
| PORT Allocation | ✅ | Auto-allocated via process.env.PORT |
| Edge Runtime | ❌ | Future capability |
| Middleware | ✅ | Supported by Next.js natively |

## Python

| Feature | Status | Notes |
| :------ | :----: | :---- |
| Flask / Gunicorn | ✅ | Detected and configured |
| Live Logs | ✅ | stdout/stderr captured |
| Health Checks | ✅ | /health endpoint |
| PORT Allocation | ✅ | Auto-allocated via PORT env |
| pip install | ✅ | requirements.txt support |
| Virtual Env | ❌ | Not yet supported (runs system Python) |
| Django | 📋 | Planned, not yet certified |
| FastAPI | 📋 | Planned, not yet certified |

## Laravel / PHP

| Feature | Status | Notes |
| :------ | :----: | :---- |
| php artisan serve | ✅ | Development server |
| Live Logs | ✅ | stdout/stderr captured |
| Health Checks | ✅ | /health endpoint |
| PORT Allocation | ✅ | Auto-allocated |
| Composer | ✅ | composer install runs if composer is on PATH |
| Nginx / FPM | ❌ | Not yet supported (development only) |
| Octane / Swoole | ❌ | Future capability |

## Cross-Stack Features

| Feature | Supported Stacks | Notes |
| :------ | :--------------- | :---- |
| Auto-detection | All | Buildpack Engine detects stack from project files |
| Auto PORT | All | Runtime.Prepare allocates port |
| Process Lifecycle | All | Start / Stop / Restart / Destroy |
| Workflow Visibility | All | SSE endpoint for live deployment progress |
| Log Streaming | Go, Node, Next.js, Python | Runtime.Logs() returns LogStream |
| Metrics | Go | CPU, memory, uptime via Runtime.Metrics() |
| Health Checks | Go, Node, Next.js, Python | Configurable HealthPolicy |
