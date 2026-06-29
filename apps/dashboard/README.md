# CloudOS Dashboard

The CloudOS Dashboard is a modern, dark-first React application that provides a visual interface for the CloudOS Control Plane API.

## Tech Stack

| Layer | Choice |
|-------|--------|
| Framework | React 19 |
| Language | TypeScript (strict) |
| Build tool | Vite 6 |
| Routing | React Router 7 |
| Server state | TanStack Query 5 |
| Styling | Tailwind CSS 3 + CSS variables |
| Components | shadcn/ui (custom) |
| Icons | Lucide React |
| API client | `@cloudos/sdk` (local package) |

## Architecture

```
apps/dashboard/
├── src/
│   ├── components/
│   │   ├── ui/          # Reusable primitives (Button, Card, Badge, etc.)
│   │   ├── layout/      # AppShell, TopNav, Sidebar, StatusBar
│   │   ├── theme/       # ThemeProvider (dark/light/system)
│   │   └── error/       # ErrorBoundary, LoadingScreen
│   ├── hooks/           # TanStack Query hooks wrapping the SDK
│   ├── lib/             # SDK singleton, utility functions
│   └── pages/           # Route-level page components
├── packages/sdk/        # CloudOS SDK (typed API client)
└── package.json
```

## Getting Started

### Prerequisites

- Node.js 20+
- The CloudOS kernel running on `http://localhost:8080`

### Start the kernel (in a separate terminal)

```bash
cd ../..
go run ./tools/cloudos
```

### Start the dashboard

```bash
# Install dependencies
npm install

# Start the dev server (proxies /api → localhost:8080)
npm run dev
```

The dashboard runs at **http://localhost:5173**.

## Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start the Vite dev server |
| `npm run build` | TypeScript check + production build |
| `npm run preview` | Preview the production build |
| `npm run typecheck` | Run TypeScript type checking |

## API Proxy

During development, the Vite dev server proxies `/api/*` requests to the Go kernel at `http://localhost:8080`. This avoids CORS issues. In production, set `VITE_CLOUDOS_API_URL` to the base URL of the API server.

## Project Status

| Page | Status | Backed by API |
|------|--------|---------------|
| Dashboard | ✅ Done | `/api/v1/health`, `/version`, `/kernel`, `/capabilities`, `/providers` |
| Kernel | ✅ Done | `/api/v1/kernel` |
| System | ✅ Done | `/api/v1/system` |
| Capabilities | ✅ Done | `/api/v1/capabilities`, `/api/v1/capabilities/:id` |
| Providers | ✅ Done | `/api/v1/providers` |
| Plugins | 🔜 Stub | Plugin Discovery API (Sprint 1) |
| Settings | 🔜 Stub | Configuration API (Sprint 1) |

## SDK

All API calls go through `@cloudos/sdk`, a typed TypeScript client located at `../../packages/sdk/src/`. Every client (dashboard, CLI, desktop, mobile, AI, automation) uses the same SDK — only the `CloudOSClient` base URL changes.

```typescript
import { CloudOSClient } from '@cloudos/sdk';

const client = new CloudOSClient('/api');
const caps = await client.getCapabilities();
const health = await client.getHealth();
```
