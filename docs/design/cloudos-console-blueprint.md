# CloudOS Console — Product Blueprint

> **Phase 1:** Information Architecture & UX Blueprint
> **Status:** Approved — ready for visual design
> **Next:** Prompt #2 — Visual Design System

---

## Product Philosophy

CloudOS is not another cloud dashboard. It does not care about VMs, subnets,
IAM roles, or availability zones. It cares about one thing: **getting your
application from a Git URL to a running, observable endpoint as fast as
possible.**

Every design decision flows from this constraint.

### Design Tenets

1. **Applications are the atomic unit.** Every page, every navigation item,
   every action exists to serve applications. Infrastructure is plumbing —
   invisible unless something breaks.

2. **Deployment is a journey, not a click.** The console should make every
   deployment transparent: you see the clone, the build, the health check,
   the URL. No loading spinners. No "wait 30 seconds and hope."

3. **Observability is the default.** Logs, timelines, health, and metrics
   are not tabs you navigate to — they're surfaces you can see from the
   moment a deployment starts.

4. **Failure is informative.** When something breaks, the console shows you
   what broke, why it broke, and how to fix it — not a red banner that says
   "Error."

5. **No dead ends.** Every screen offers a next action. Every empty state
   offers a starting point. Every error offers remediation.

---

## User Personas

### 1. Solo Developer — "Alex"

| Attribute | Detail |
|-----------|--------|
| **Role** | Independent developer, freelancer, student |
| **Goal** | Deploy personal projects from GitHub with zero friction |
| **Skills** | Comfortable with CLI, avoids cloud consoles |
| **Pain** | Spends more time on CI/CD than on actual code |
| **Journey** | `git push → cloudosctl deploy → open URL` — wants the console to show what happened |
| **Uses console for** | Checking deployment status, watching logs, sharing URLs |

### 2. Team Lead — "Jordan"

| Attribute | Detail |
|-----------|--------|
| **Role** | Engineering manager, tech lead at a small team (2-10 devs) |
| **Goal** | Understand what's deployed, compare versions, review health |
| **Skills** | Comfortable with CLI and basic dashboards |
| **Pain** | Can't quickly tell if a deployment succeeded or what changed |
| **Journey** | `deploy → review timeline → compare to previous → check health` |
| **Uses console for** | Deployment comparison, timeline review, health monitoring, team collaboration |

### 3. Platform Engineer — "Morgan"

| Attribute | Detail |
|-----------|--------|
| **Role** | DevOps / platform engineer configuring CloudOS for a team |
| **Goal** | Configure runtimes, buildpacks, and system settings |
| **Skills** | Expert — comfortable with kernel internals |
| **Pain** | Needs visibility into the engine without reading source code |
| **Journey** | `configure runtime → enable buildpack → verify via controller health` |
| **Uses console for** | System settings, runtime configuration, buildpack management, plugin SDK |

---

## Primary Navigation

```
┌──────────────────────────────────────────────────────────┐
│  CloudOS                                                  │
│                                                           │
│  ◆  Applications              ← PRIMARY — always visible │
│  ◆  Deployments               ← SECONDARY — cross-app    │
│  ◆  Projects                  ← ORGANIZATION             │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─                              │
│  ◆  Monitoring                ← OBSERVABILITY            │
│  ◆  Workflows                 ← ENGINE                   │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─                              │
│  ◆  System                    ← ADMIN                    │
│  ◆  Settings                  ← CONFIGURATION            │
│  ◆  Plugins                   ← EXTENSIBILITY            │
└──────────────────────────────────────────────────────────┘
```

### Navigation Rationale

| Item | Why It Exists | Primary User |
|------|---------------|--------------|
| **Applications** | Core entity — deploy, observe, manage | All |
| **Deployments** | Cross-app deployment history and search | Team Lead |
| **Projects** | Group applications into logical collections | Team Lead |
| **Monitoring** | Global health, metrics, and alerts | Team Lead |
| **Workflows** | Visualize the deployment engine itself | Platform Engineer |
| **System** | Kernel health, controllers, providers | Platform Engineer |
| **Settings** | Runtimes, buildpacks, configuration | Platform Engineer |
| **Plugins** | Discover and manage extensions | Platform Engineer |

Applications is the **home screen** — not a dashboard of widgets, but a
purposeful list of everything you've deployed. This is what CloudOS does.

---

## Information Architecture

```
CloudOS Console
│
├── Applications ◄── DEFAULT LANDING
│   │
│   ├── Application List
│   │   Every deployed application. Status, health, URL, last deployment.
│   │   Search, filter by status, sort by last deployed.
│   │   Empty state: "Deploy your first application" with CLI command.
│   │
│   └── Application Detail
│       │
│       ├── Overview
│       │   Status badge, health indicator, URL, uptime, quick actions.
│       │   Latest deployment summary (duration, steps, result).
│       │
│       ├── Deployments
│       │   Scrollable list of every deployment (#1, #2, ...).
│       │   Each row: number, timestamp, duration, status (success/fail).
│       │   Compare any two deployments side-by-side.
│       │
│       ├── Timeline
│       │   Visual step-by-step view of a single deployment.
│       │   6 standard steps: Validate → Clone → Build → Deploy → Check → Complete.
│       │   Each step: status icon, duration, output, error if failed.
│       │
│       ├── Logs
│       │   Live streaming log viewer (SSE).
│       │   Pause / resume, search/filter, download as file.
│       │   Log levels: info, warn, error with color coding.
│       │
│       └── Settings
│           Application configuration (read-only for now).
│           Source URL, runtime type, environment variables.
│           Actions: re-deploy, stop, destroy (future).
│
├── Deployments (cross-app)
│   │
│   ├── Deployment Feed
│   │   Chronological list of every deployment across all apps.
│   │   Useful for team leads reviewing what changed recently.
│   │
│   └── Deployment Detail
│       Timeline view for a specific deployment.
│       Reuses the same Timeline component from Application Detail.
│
├── Projects
│   │
│   ├── Project List
│   │   Logical groups of applications (e.g., "frontend", "backend", "docs").
│   │   Each project: name, description, member count, app count.
│   │
│   └── Project Detail
│       All applications belonging to this project.
│       Project-level health summary.
│
├── Monitoring
│   │
│   ├── Global Health
│   │   Status overview: all applications, their health, uptime.
│   │   Recent failures, degraded services.
│   │
│   └── Metrics (future)
│       CPU, memory, request rate (requires runtime support).
│
├── Workflows
│   │
│   ├── Workflow Definitions
│   │   List of registered workflow definitions.
│   │   Each: name, steps, version, status.
│   │
│   └── Workflow Executions
│       Recent and in-progress workflow runs.
│       Useful for debugging deployment engine issues.
│
├── System
│   │
│   ├── Kernel
│   │   Version, uptime, Go version, platform.
│   │   Resource registry status.
│   │
│   ├── Controllers
│   │   Registered controllers and their health.
│   │   Application Controller, Namespace Controller, etc.
│   │
│   ├── Providers
│   │   Registered providers and their capabilities.
│   │   Runtime providers, storage providers, etc.
│   │
│   └── Capabilities
│       Available capabilities and which providers implement them.
│
├── Settings
│   │
│   ├── General
│   │   API address, default runtime, log level.
│   │
│   ├── Runtimes
│   │   Configured runtime implementations (Local, OCI/Docker).
│   │   Active runtime, fallback behavior.
│   │
│   └── Buildpacks
│       Registered buildpacks and their detection order.
│       Enable/disable specific buildpacks.
│
└── Plugins
    │
    ├── Plugin Catalog (future)
    │   Discover available plugins.
    │
    └── Installed Plugins
        Manage installed plugins.
        Each plugin: name, version, status, configuration.
```

---

## Screen Hierarchy

```
┌─ GLOBAL LAYOUT ──────────────────────────────────────┐
│  ┌─────────────────────────────────────────────────┐ │
│  │  Top Navigation Bar                             │ │
│  │  [Logo] [Search] [Notifications] [User Avatar]  │ │
│  └─────────────────────────────────────────────────┘ │
│  ┌──────────┬──────────────────────────────────────┐ │
│  │          │                                       │ │
│  │ Sidebar  │           Main Content                │ │
│  │          │                                       │ │
│  │  ◆ Apps  │   (varies by route)                   │ │
│  │  ◆ Depl  │                                       │ │
│  │  ◆ Proj  │                                       │ │
│  │  ──────  │                                       │ │
│  │  ◆ Mon   │                                       │ │
│  │  ◆ Wf    │                                       │ │
│  │  ──────  │                                       │ │
│  │  ◆ Sys   │                                       │ │
│  │  ◆ Set   │                                       │ │
│  │  ◆ Plug  │                                       │ │
│  │          │                                       │ │
│  └──────────┴──────────────────────────────────────┘ │
└──────────────────────────────────────────────────────┘
```

### Top Navigation Bar

| Element | Purpose |
|---------|---------|
| **Logo** | Brand + click → home (Applications) |
| **Global Search** | Search across applications, deployments, projects |
| **Status Indicator** | Green/yellow/red dot — kernel health at a glance |
| **Notifications** | Bell icon — deployment completions, failures, updates (future) |
| **User Avatar** | User menu — preferences, sign out (future) |

### Sidebar

The sidebar uses **section groups** with visual separators:

1. **Primary** — Applications, Deployments, Projects (what users do)
2. **Observability** — Monitoring, Workflows (what users watch)
3. **Admin** — System, Settings, Plugins (what users configure)

Active item is highlighted. Each item shows a subtle icon + label.
No expand/collapse — single-level navigation keeps it fast.

---

## Page Hierarchy

```
/                                    → Applications (default landing)
├── /applications                    → Application List
├── /applications/deploy             → Deploy New (inline, no separate page)
├── /applications/:id                → Application Detail
│   ├── /applications/:id/overview   → Overview (default tab)
│   ├── /applications/:id/deployments → Deployments (tab)
│   ├── /applications/:id/timeline   → Timeline (tab)
│   ├── /applications/:id/logs       → Logs (tab)
│   └── /applications/:id/settings   → Settings (tab)
│
├── /deployments                     → Cross-app Deployment Feed
├── /deployments/:id                 → Deployment Detail (timeline)
│
├── /projects                        → Project List
├── /projects/:id                    → Project Detail
│
├── /monitoring                      → Global Health Dashboard
│
├── /workflows                       → Workflow Definitions
├── /workflows/:id                   → Workflow Detail
├── /workflows/:id/executions        → Workflow Executions
│
├── /system                          → System Overview
├── /system/kernel                   → Kernel Details
├── /system/controllers              → Controller List
├── /system/controllers/:id          → Controller Detail
├── /system/providers                → Provider List
├── /system/providers/:id            → Provider Detail
├── /system/capabilities             → Capability List
├── /system/capabilities/:id         → Capability Detail
│
├── /settings                        → Settings
├── /settings/general                → General Config
├── /settings/runtimes               → Runtime Configuration
├── /settings/buildpacks             → Buildpack Management
│
└── /plugins                         → Plugin Management
    ├── /plugins/catalog              → Plugin Catalog
    └── /plugins/installed            → Installed Plugins
```

---

## Navigation Map

```
                        ┌─────────────────────┐
                        │   APPLICATIONS       │ ◄── HOME
                        │   (application list) │
                        └──────────┬──────────┘
                                   │ click app
                                   ▼
              ┌────────────────────────────────────┐
              │         APPLICATION DETAIL          │
              │                                      │
              │  ┌────┬──────┬──────┬────┬────────┐ │
              │  │OV  │ DEPL │ TIME │ LOG│ SETTNG │ │
              │  │view│oymnt │ line │ s  │ s      │ │
              │  └────┴──────┴──────┴────┴────────┘ │
              │                                      │
              │  ┌─ deployment click ──────────────┐ │
              │  │  /apps/:id/deployments          │ │
              │  │  Lists all deployments          │ │
              │  │  "Compare" selects two          │ │
              │  └─────────────────────────────────┘ │
              │                                      │
              │  ┌─ timeline click ────────────────┐ │
              │  │  /apps/:id/timeline             │ │
              │  │  Visual step-by-step view       │ │
              │  └─────────────────────────────────┘ │
              │                                      │
              │  ┌─ logs tab ──────────────────────┐ │
              │  │  /apps/:id/logs                 │ │
              │  │  Live streaming + search        │ │
              │  └─────────────────────────────────┘ │
              └──────────────────────────────────────┘
                         │
                         │ "All Deployments"
                         ▼
              ┌────────────────────────────────────┐
              │         DEPLOYMENTS (cross-app)     │
              │  Chronological feed of all deploys  │
              └────────────────────────────────────┘
                         │
                         │ sidebar click
                         ▼
              ┌────────────────────────────────────┐
              │           MONITORING                │
              │  Global health dashboard            │
              └────────────────────────────────────┘
                         │
                         │ sidebar click
                         ▼
              ┌────────────────────────────────────┐
              │           WORKFLOWS                 │
              │  Engine visibility                  │
              └────────────────────────────────────┘
                         │
                         │ sidebar click
                         ▼
              ┌────────────────────────────────────┐
              │            SYSTEM                   │
              │  Kernel, Controllers, Providers    │
              └────────────────────────────────────┘
                         │
                         │ sidebar click
                         ▼
              ┌────────────────────────────────────┐
              │           SETTINGS                  │
              │  Runtimes, Buildpacks, Config      │
              └────────────────────────────────────┘
```

---

## User Journeys

### Journey 1: First Deployment (Alex — Solo Developer)

```
1. Install CloudOS → start kernel → open http://localhost:8080

2. See empty state on Applications page:
   "No applications yet. Deploy your first app:"
   cloudosctl deploy https://github.com/your/repo

3. Switch to terminal, run the deploy command.

4. Switch back to browser. The app appears in the list
   within seconds — status: "Deploying."

5. Click the app → auto-land on Logs tab.
   See live streaming logs: clone → build → deploy → health check.

6. See status change to "Running" with a green health badge.

7. Click the URL → opens the app in a new tab.

RESULT: Under 5 minutes from clone to running URL.
```

### Journey 2: Investigating a Failed Deployment (Jordan — Team Lead)

```
1. Open Applications page. See one app showing "Degraded" health.

2. Click the app → see Overview tab:
   - Status: Running (degraded)
   - Last deployment: #42 — Completed with errors

3. Switch to Deployments tab. See deployment #42.
   Duration: 8.2s. Status: "Completed with warnings."

4. Click deployment #42 → Timeline tab loads with #42 selected.
   See: 5/6 steps succeeded.
   Step 5 (Health Check): ✗ Failed — HTTP 503

5. Switch to Logs tab. Search for "error" or "503".
   See the application's stdout: "Cannot connect to database."

6. Click "Compare" → select deployment #41 and #42.
   See that the commit changed from a1b2c3d to e5f6g7h.
   Duration increased from 6.1s to 8.2s.

RESULT: Diagnosis in under 60 seconds, no terminal needed.
```

### Journey 3: Platform Configuration (Morgan — Platform Engineer)

```
1. Open System → Kernel: see version 0.6.0-rc1, uptime, Go version.

2. Open System → Controllers: see Application Controller is healthy.

3. Open System → Providers: see LocalRuntime and OCI Runtime registered.

4. Open Settings → Runtimes: see active runtime is OCI (Docker).
   Fallback: LocalRuntime.

5. Open Settings → Buildpacks: see 7 buildpacks registered.
   Detection order: Go → Node → React → Next.js → Python → Laravel → Static.

6. Open Workflows: see standard deployment workflow definition.
   6 steps: Validate → Clone → Build → Deploy → Check → Complete.

RESULT: Full platform visibility without reading source code.
```

---

## Page Purpose and User Goals

### Applications List (`/applications`)

| Aspect | Detail |
|--------|--------|
| **Why it exists** | This is the home screen. Every user action starts here. |
| **User goal** | See everything deployed at a glance. Find a specific app. Understand the overall health of the system. |
| **Primary action** | Click an app to see details. |
| **Secondary action** | Search, filter, sort to find an app quickly. |
| **Empty state** | "No applications yet." Shows deploy command. |
| **Error state** | "Kernel not reachable." Shows how to start the kernel. |

### Application Detail (`/applications/:id`)

| Tab | Why It Exists | User Goal |
|-----|---------------|-----------|
| **Overview** | At-a-glance status of a specific app | Confirm it's running and healthy. Get the URL. See latest deployment summary. |
| **Deployments** | Full deployment history | Review past deployments. Compare any two. See who deployed what. |
| **Timeline** | Step-by-step deployment visibility | Understand exactly what happened during a deployment. Debug failures. See durations. |
| **Logs** | Real-time and historical logs | Watch a deployment live. Search for errors. Download for offline analysis. |
| **Settings** | Application configuration | See source URL, runtime info. Re-deploy, stop (future). |

### Deployments Feed (`/deployments`)

| Aspect | Detail |
|--------|--------|
| **Why it exists** | Cross-app visibility for team leads reviewing recent changes. |
| **User goal** | See everything that was deployed recently across all applications. |
| **Why not just per-app?** | When you manage 5+ apps, you need a single chronological view. |

### Projects (`/projects`)

| Aspect | Detail |
|--------|--------|
| **Why it exists** | Organizational grouping for teams with multiple applications. |
| **User goal** | Group apps by team, service, or environment. |
| **Relationship to apps** | Projects contain applications. An app belongs to exactly one project. |

### Monitoring (`/monitoring`)

| Aspect | Detail |
|--------|--------|
| **Why it exists** | Global health visibility across all applications. |
| **User goal** | Spot problems before users report them. See which apps are healthy, degraded, or down. |
| **When to visit** | On notification of a failure; or as a daily check-in. |

### Workflows (`/workflows`)

| Aspect | Detail |
|--------|--------|
| **Why it exists** | Visibility into the deployment engine itself. |
| **User goal** | Understand what happens when you deploy. Debug engine-level issues. |
| **Who uses it** | Platform engineers and curious developers. |

### System (`/system`)

| Aspect | Detail |
|--------|--------|
| **Why it exists** | Kernel and infrastructure visibility without CLI. |
| **User goal** | Check kernel health, controller status, registered providers and capabilities. |
| **Who uses it** | Platform engineers debugging configuration issues. |

### Settings (`/settings`)

| Aspect | Detail |
|--------|--------|
| **Why it exists** | Configure CloudOS behavior. |
| **User goal** | Change active runtime, enable/disable buildpacks, configure logging. |
| **Who uses it** | Platform engineers. |

### Plugins (`/plugins`)

| Aspect | Detail |
|--------|--------|
| **Why it exists** | Future extensibility. |
| **User goal** | Discover and manage plugins (custom buildpacks, runtimes, integrations). |
| **Who uses it** | Platform engineers extending CloudOS. |

---

## Design Principles Applied to Layout

### Sidebar

```
┌──────────────────────┐
│  ☁  CloudOS          │ ← Logo + click → home
│                      │
│  ◆  Applications     │ ← Active state
│  ◆  Deployments      │
│  ◆  Projects         │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─  │ ← Visual separator
│  ◆  Monitoring       │
│  ◆  Workflows        │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─  │ ← Visual separator
│  ◆  System           │
│  ◆  Settings         │
│  ◆  Plugins          │
│                      │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│  [version badge]     │ ← v0.6.0-rc1 at bottom
└──────────────────────┘
```

### Application Detail — Tab Layout

```
┌──────────────────────────────────────────────────────┐
│ ← Back to Applications               Status: ● Running │
│                                                      │
│  go-api                                              │
│  github.com/cloudos-examples/go-api                  │
│                                                      │
│  ┌──────────┬──────────┬─────────┬──────┬──────────┐ │
│  │ Overview │ Deploy'ts │ Timeline│ Logs │ Settings │ │
│  └──────────┴──────────┴─────────┴──────┴──────────┘ │
│                                                      │
│  [content for selected tab]                          │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### Deployments Tab

```
┌──────────────────────────────────────────────────────┐
│ Deployments                                          │
│                                                      │
│ ┌── Filter/Search ────────────────────────────────┐  │
│ │ [Search...]                    [Compare ▾]      │  │
│ └─────────────────────────────────────────────────┘  │
│                                                      │
│ ┌── Deployment List ──────────────────────────────┐  │
│ │ #42  2 min ago   8.2s  ● Completed  a1b2c3d  ☐  │  │
│ │ #41  15 min ago  6.1s  ● Completed  e5f6g7h  ☐  │  │
│ │ #40  1 hour ago  7.0s  ✗ Failed     f8g9h0i     │  │
│ └─────────────────────────────────────────────────┘  │
│                                                      │
│ [Compare Selected] (visible when 2 checked)           │
└──────────────────────────────────────────────────────┘
```

### Timeline View

```
┌──────────────────────────────────────────────────────┐
│ Timeline: Deployment #42                             │
│ Duration: 8.2s  |  Completed: 2 min ago              │
│                                                      │
│   ✓  Validate Application              0.3s          │
│      Configuration valid                             │
│                                                      │
│   ✓  Clone Source Repository            1.2s          │
│      Cloned 42 commits                               │
│                                                      │
│   ✓  Build Artifact                     3.1s          │
│      Build completed, binary=app                     │
│                                                      │
│   ✓  Deploy Application                 2.0s          │
│      Deployed to runtime                             │
│                                                      │
│   ✗  Health Check                        1.6s         │
│      HTTP 503 Service Unavailable                    │
│      → The application did not respond to the        │
│        health check. Ensure your app listens on      │
│        the PORT environment variable.                │
│                                                      │
│   ✓  Complete Deployment                0.0s          │
│      Deployment #42 complete (with errors)           │
└──────────────────────────────────────────────────────┘
```

### Logs View

```
┌──────────────────────────────────────────────────────┐
│ Logs — go-api                                        │
│                                                      │
│ [🔍 Search logs...]        [⏸ Pause] [⬇ Download]    │
│                                                      │
│ ┌── Log Stream ────────────────────────────────────┐ │
│ │ 14:30:00 • App [build] Cloning repository...      │ │
│ │ 14:30:01 • App [build] Detecting runtime: Go 1.24 │ │
│ │ 14:30:03 • App [build] Building binary...          │ │
│ │ 14:30:05 • App [deploy] Deploying application...   │ │
│ │ 14:30:06 • App [health] Health check... HTTP 200   │ │
│ │ 14:30:06 • App [deploy] Deployment #42 complete    │ │
│ │ 14:30:07 • App        Server listening on :31245   │ │
│ └─────────────────────────────────────────────────┘  │
│                                                      │
│ ● Streaming live (Ctrl+C to stop)                    │
└──────────────────────────────────────────────────────┘
```

---

## Route Map Summary

```
Route                              Page Component          Sidebar Active
──────────────────────────────────────────────────────────────────────────
/                                   ApplicationsList       Applications
/applications                       ApplicationsList       Applications
/applications/:id                   ApplicationDetail      Applications
/applications/:id/overview          OverviewTab            —
/applications/:id/deployments       DeploymentsTab         —
/applications/:id/timeline          TimelineTab            —
/applications/:id/logs              LogsTab                —
/applications/:id/settings          SettingsTab            —

/deployments                        DeploymentsFeed        Deployments
/deployments/:id                    DeploymentDetail       Deployments

/projects                           ProjectList            Projects
/projects/:id                       ProjectDetail          Projects

/monitoring                         MonitoringPage         Monitoring

/workflows                          WorkflowList           Workflows
/workflows/:id                      WorkflowDetail         Workflows
/workflows/:id/executions           WorkflowExecutions     Workflows

/system                             SystemOverview         System
/system/kernel                      KernelDetail           System
/system/controllers                 ControllerList         System
/system/controllers/:id             ControllerDetail       System
/system/providers                   ProviderList           System
/system/providers/:id               ProviderDetail         System
/system/capabilities                CapabilityList         System
/system/capabilities/:id            CapabilityDetail       System

/settings                           SettingsPage           Settings
/settings/general                   GeneralSettings        Settings
/settings/runtimes                  RuntimeSettings        Settings
/settings/buildpacks                BuildpackSettings      Settings

/plugins                            PluginList             Plugins
/plugins/catalog                    PluginCatalog          Plugins
/plugins/installed                  InstalledPlugins       Plugins
```

---

## Implementation Priority

### Phase 1 — v0.6 (Current)

Already exists and should be preserved:

| Page | Status |
|------|--------|
| Applications List | ✅ Exists as default landing |
| Application Detail (5 tabs) | ✅ Exists |
| Deployments Tab | ✅ Exists |
| Timeline Tab | ✅ Exists |
| Logs Tab | ✅ Exists |
| Settings Tab | ✅ Exists |
| System pages | ✅ Exists (Kernel, Controllers, Providers, Capabilities) |

### Phase 2 — v0.7 (Next)

New pages to build:

| Page | Effort | Priority |
|------|--------|----------|
| Cross-app Deployments Feed | Low | High — missing visibility |
| Deployments landing page | Low | High — need a container |
| Project List + Detail | Medium | Medium — organizational |
| Monitoring / Global Health | Medium | Medium — observability |
| Workflow Definitions + Executions | Medium | Medium — engine visibility |
| Settings (General, Runtimes, Buildpacks) | Low | High — configuration UX |
| Plugins (Catalog + Installed) | Low | Low — future SDK |

### Phase 3 — v1.0 (Future)

| Page | Effort | Priority |
|------|--------|----------|
| Plugin Marketplace | High | Depends on SDK |
| Global Search | Medium | High — usability |
| Notifications | Medium | Medium — engagement |
| User auth / multi-tenant | High | Medium — requires backend |

---

## Key UX Patterns

### Empty States

Every list page must have an intentional empty state:

| Page | Empty State |
|------|-------------|
| Applications | "Deploy your first application" with CLI command shown |
| Deployments | "No deployments yet. Deploy an app first." |
| Projects | "Create your first project to organize applications." |
| Monitoring | "Deploy an application to see health metrics." |
| Workflows | No empty state — workflows always exist in the kernel. |
| Plugins | "No plugins installed. The Plugin SDK is coming in v0.7." |

### Loading States

| Pattern | Behavior |
|---------|----------|
| Page load | Skeleton screen matching page layout. No spinners. |
| Tab switch | Content loads immediately. If data is stale, show cached data with a subtle refresh indicator. |
| Log stream | Stream starts immediately. Show recent history, then live tail. |
| Deployment | Polling shows real-time status changes. No progress bars — show actual step transitions. |

### Error States

| Error | UX |
|-------|-----|
| Kernel not reachable | Full-page error with "Start the kernel" instructions. Retry button. |
| Application not found | 404 page with "Check the application name" and link to Applications list. |
| Deployment failed | Timeline highlights the failed step in red. Error message includes remediation. |
| API error | Inline error banner with error code and message. "Retry" button. |
| Network disconnected | Banner: "Connection lost. Reconnecting..." Auto-reconnects. |

---

## Ready for Implementation

This blueprint defines every page CloudOS needs, why it exists, and how
users move between them. The navigation is application-centric. The
information architecture matches how developers think: applications first,
infrastructure hidden.

An engineer reading this document can now:

1. Name every page in the console
2. Explain why each page exists
3. Describe the primary user journey through each page
4. Identify empty, loading, and error states for every surface
5. Understand the priority order for implementation

**The next step is the Visual Design System** — colors, typography, spacing,
component tokens, and layout grid. That's Prompt #2.
