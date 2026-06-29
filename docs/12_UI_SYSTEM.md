# UI/UX Design

> **Document:** 08_UI_UX.md
> **Status:** Draft v0.1
> **Depends On:** [07_API.md](./07_API.md)

---

## 1. Design Principles

1. **Clarity over Complexity** — Every screen has one primary action. Hide advanced options behind progressive disclosure.
2. **Mobile-First** — Design for 5" screens first, then scale up. Every feature must work on mobile.
3. **Dark-First** — Dark mode is the default and primary experience. Light mode is secondary.
4. **Data-Dense but Readable** — Show as much data as needed without overwhelming. Use visual hierarchy.
5. **Instant Feedback** — Every action shows result within 100ms (optimistic) or 1s (confirmed).

---

## 2. Application Surfaces

| Surface | Framework | Form Factor | Primary Users |
|---------|-----------|-------------|---------------|
| **Dashboard** | React 19 + Tailwind v4 | Desktop + Tablet | All users |
| **Mobile** | React Native / Expo 52 | Phone | Alex, Riley |
| **Desktop** | Tauri 2 + React 19 | Desktop (native) | Jordan, Taylor |
| **CLI** | Go (charm.sh/bubbletea) | Terminal | Alex, Jordan |
| **AI Chat** | React 19 + WebSocket | All surfaces | All users |

---

## 3. Design System

### 3.1 Design Tokens

Tokens are defined in `packages/design-tokens/` as CSS custom properties and TypeScript constants:

```css
:root {
  /* Colors — Dark Mode (default) */
  --color-bg-primary: #0a0a0b;
  --color-bg-secondary: #141416;
  --color-bg-tertiary: #1c1c1f;
  --color-surface: #242428;
  --color-border: #2e2e32;
  --color-text-primary: #fafafa;
  --color-text-secondary: #a1a1aa;
  --color-text-tertiary: #71717a;
  --color-accent: #6366f1;
  --color-accent-hover: #818cf8;
  --color-success: #22c55e;
  --color-warning: #eab308;
  --color-error: #ef4444;
  --color-info: #3b82f6;

  /* Typography */
  --font-sans: 'Inter', -apple-system, sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', monospace;
  --font-size-xs: 0.75rem;
  --font-size-sm: 0.875rem;
  --font-size-base: 1rem;
  --font-size-lg: 1.125rem;
  --font-size-xl: 1.25rem;
  --font-size-2xl: 1.5rem;
  --font-size-3xl: 2rem;

  /* Spacing (4px base) */
  --space-1: 0.25rem;
  --space-2: 0.5rem;
  --space-3: 0.75rem;
  --space-4: 1rem;
  --space-6: 1.5rem;
  --space-8: 2rem;
  --space-12: 3rem;
  --space-16: 4rem;

  /* Radius */
  --radius-sm: 0.375rem;
  --radius-md: 0.5rem;
  --radius-lg: 0.75rem;
  --radius-xl: 1rem;
}
```

### 3.2 Component Library

Built in `packages/ui/` with Radix primitives + Tailwind:

- **Layout**: AppShell, Sidebar, Topbar, SplitPane
- **Navigation**: Tabs, Breadcrumbs, CommandPalette
- **Data**: Table, DataGrid, Tree, List, Card, StatCard
- **Forms**: Input, Select, DatePicker, FormField, Button
- **Feedback**: Toast, Alert, Modal, Drawer, Tooltip
- **Visualization**: Chart, Gauge, Sparkline, Heatmap
- **AI**: ChatPanel, SuggestionChip, CommandBar

---

## 4. Key Screens

### 4.1 Dashboard Home
- Project overview cards (deployments, databases, storage)
- Quick action bar (deploy, create DB, add storage)
- Recent activity feed
- Key metrics sparklines
- AI assistant chat (collapsible sidebar)

### 4.2 Project Detail
- Deployment timeline
- Resource tabs (Compute, Database, Storage, Network)
- Environment switcher (production/staging/preview)
- Real-time metrics dashboard
- Log viewer with search/filter

### 4.3 AI Chat Interface
- Persistent chat panel on all screens
- Context-aware queries (auto-detects current project/resource)
- Suggestion chips for common operations
- Action buttons in responses ("Rollback now", "Scale up")
- Source references with clickable links

### 4.4 Plugin Marketplace
- Plugin cards with screenshots and rating
- One-click install
- Configuration wizard on first activation
- Usage statistics and reviews

### 4.5 Mobile Navigation
- Bottom tab bar: Home, Projects, Plugins, AI, Settings
- Swipeable panels
- Pull-to-refresh on all data views
- Haptic feedback on actions
- Widget support on home screen

---

## 5. Accessibility

- WCAG 2.1 AA minimum (AAA where practical)
- Full keyboard navigation
- Screen reader support (ARIA labels, roles, live regions)
- Focus indicators on all interactive elements
- Reduced motion support
- High contrast mode support

---

> **Next:** [09_AI_SYSTEM.md](./09_AI_SYSTEM.md) — AI system design
