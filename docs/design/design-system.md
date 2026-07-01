# CloudOS Console — Design System

> **Phase 2:** Visual Design Language
> **Status:** Draft — ready for review
> **Next:** Prompt #3 — Component Implementation

---

## Design Philosophy

CloudOS is a developer tool, not a marketing site, not an enterprise dashboard,
not a Bootstrap admin template.

It should feel like opening a well-designed terminal — calm, capable, and
completely in service of the task at hand.

### Visual Principles

| Principle | Meaning |
|-----------|---------|
| **Calm** | No visual noise. No unnecessary gradients, shadows, or animations. Every pixel earns its place. |
| **Professional** | Tight alignment, consistent spacing, deliberate typography. Looks like it was designed by people who care about margins. |
| **Fast** | The interface never feels slow. Skeleton screens, instant transitions, optimistic UI. |
| **Reliable** | Status is always visible. Health is always shown. Nothing is hidden behind a hover unless it's secondary. |
| **Developer-first** | Code blocks are beautiful. Terminals are readable. Logs are searchable. Mono fonts are respected. |

### What CloudOS Does NOT Look Like

| ❌ Avoid | ✅ Instead |
|----------|------------|
| Bright blue primary buttons everywhere | A single muted accent color used sparingly |
| Dense data tables with zebra striping | Clean list layouts with subtle dividers |
| Heavy card shadows and gradients | Flat cards with 1px borders |
| Circular avatars with random colors | Monogram initials in a consistent container |
| Animated background particles | Zero background decoration |
| Icon-only navigation with tooltips | Labeled navigation, always readable |

---

## Color System

### Philosophy

CloudOS uses a **warm, neutral dark palette** as the default (developer tools
live in dark mode) and a **clean light mode** as the alternative. Colors are
deliberately muted — no electric blues, no neon accents. The interface steps
back so the content (applications, logs, timelines) steps forward.

### Dark Mode (Default)

```
Background        #0C0C0D    ← Near-black with subtle warmth
Surface           #151517    ← Cards, panels
Surface Elevated  #1C1C1F    ← Modals, dropdowns, command palette
Sidebar           #111113    ← Slightly darker than surface
Top Nav           #151517    ← Same as surface

Border            #26262B    ← Subtle separation
Border Hover      #34343A    ← On hover, focus

Text Primary      #EDEDEF    ← High-emphasis
Text Secondary    #9D9DA3    ← Low-emphasis, metadata
Text Muted        #5F5F66    ← Placeholders, disabled
Text Inverse      #0C0C0D    ← On colored backgrounds

Accent            #5E6AD2    ← Indigo — primary actions, active states
Accent Hover      #6F7AE0    ← Button hover
Accent Subtle     #1E1F3A    ← Selected row, active nav item

Success           #2B9D5D    ← Healthy, completed
Success Subtle    #0D2B1A    ← Background tint

Warning           #D4A72C    ← Degraded, warnings
Warning Subtle    #2D2209    ← Background tint

Danger            #D45A5A    ← Failed, errors
Danger Subtle     #2D0E0E    ← Background tint

Info              #4A8FE4    ← Informational
Info Subtle       #0E1F33    ← Background tint

Link              #6F8FE4    ← Interactive text links
```

### Light Mode

```
Background        #F7F7F8    ← Off-white with warmth
Surface           #FFFFFF    ← Cards, panels
Surface Elevated  #FFFFFF    ← Modals (with subtle shadow)
Sidebar           #F0F0F1    ← Slightly darker than background
Top Nav           #FFFFFF    ← Same as surface

Border            #E3E3E6    ← Subtle separation
Border Hover      #C9C9CE    ← On hover, focus

Text Primary      #1A1A1D    ← High-emphasis
Text Secondary    #6B6B73    ← Low-emphasis, metadata
Text Muted        #9D9DA3    ← Placeholders, disabled
Text Inverse      #FFFFFF    ← On colored backgrounds

Accent            #5E6AD2    ← Same indigo across modes
Accent Hover      #4D57B8    ← Darker on light background
Accent Subtle     #EEF0FC    ← Selected row, active nav item

Success           #1A8A4A    ← Slightly darker for contrast
Success Subtle    #E8F5ED    ← Background tint

Warning           #B8912A    ← Slightly darker for contrast
Warning Subtle    #FCF5E0    ← Background tint

Danger            #C74444    ← Slightly darker for contrast
Danger Subtle     #FDEEEE    ← Background tint

Info              #3B7DD4    ← Slightly darker for contrast
Info Subtle       #EBF2FC    ← Background tint

Link              #4A6FC4    ← Interactive text links
```

### Usage Rules

| Token | Where to Use |
|-------|--------------|
| `Background` | Page background (main content area) |
| `Surface` | Card backgrounds, form fields, table rows |
| `Surface Elevated` | Modals, dialogs, dropdowns, command palette |
| `Sidebar` | Left navigation panel |
| `Top Nav` | Top navigation bar |
| `Border` | Card borders, table cell borders, dividers |
| `Border Hover` | Input focus ring, hover card border |
| `Accent` | Primary buttons, active tab, active nav item, links |
| `Accent Subtle` | Selected table row, active sidebar item background |
| `Success/Warning/Danger/Info` | Status badges, health indicators, alert banners |

### Health & Status Color Map

```
Status      Color     Background    Icon
──────────────────────────────────────────
Running     Success   Success Bg    ● (filled circle)
Degraded    Warning   Warning Bg    ◐ (half circle)
Failed      Danger    Danger Bg     ○ (empty circle)
Stopped     Muted     Surface       ◌ (dashed circle)
Deploying   Accent    Accent Bg     ◉ (pulsing circle)
```

### Deployment Status Color Map

```
Status      Color     Description
─────────────────────────────────────
Succeeded   Success   ● Green
Failed      Danger    ● Red
Running     Accent    ● Indigo (pulsing)
Pending     Muted     ● Gray
Skipped     Muted     ● Gray (with dash)
Cancelled   Warning   ● Yellow
```

---

## Typography

### Philosophy

System fonts for performance. Mono font for code. No custom typefaces until
v1.0. The typography should feel spacious but dense enough to show information
at a glance — like a well-designed terminal emulator with beautiful font
rendering.

### Font Families

```
UI Font:      -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Inter', Roboto, sans-serif
Mono Font:    'SF Mono', 'Cascadia Code', 'Fira Code', 'JetBrains Mono', Consolas, monospace
```

**Rationale:** System fonts load instantly, render natively on each platform,
and are already familiar to users. Mono fonts should support programming
ligatures (like `→` `>=` `!=`) for beautiful code rendering.

### Type Scale

```
── HEADINGS ──────────────────────────────────────────

H1   24px / 32px   600   ← Page titles (rarely used)
H2   20px / 28px   600   ← Section headers
H3   16px / 24px   600   ← Card titles, modal headers
H4   14px / 20px   600   ← Sub-section headers, tab labels

── BODY ──────────────────────────────────────────────

Body 14px / 20px   400   ← Default text, table cells, descriptions
Small 13px / 18px   400   ← Metadata, timestamps, secondary text
Caption 12px / 16px   400   ← Status labels, badges, helper text

── CODE ──────────────────────────────────────────────

Code 13px / 20px   400   ← Inline code, log output, terminal
Code Small 12px / 16px   400   ← Log timestamps, line numbers

── SPECIAL ───────────────────────────────────────────

Sidebar Item  14px / 20px   500   ← Navigation items
Tab Label     14px / 20px   500   ← Tab bar items
Button Label  14px / 20px   500   ← Button text
Badge         12px / 16px   500   ← Status badges
```

### Line Heights

```
Tight    1.3   ← Headings, badges
Normal   1.5   ← Body text, tables
Relaxed  1.7   ← Empty states, descriptions, paragraphs
```

### Letter Spacing

```
Normal    0em      ← Most text
Wide      0.02em   ← All-caps labels (rare — use sparingly)
Mono      0em      ← Code and terminal output
```

### Typography Rules

1. Never use all-caps for headings. Use sentence case everywhere.
2. Never justify text. Always left-align.
3. Code blocks use the mono font at 13px. Background should be `Surface Elevated`.
4. Inline code uses the mono font with a subtle background tint (`Accent Subtle` at 50% opacity).
5. Links are always colored (`Link` token) and never underlined unless in running text.
6. Headings have no bottom margin larger than 8px.
7. Log output uses mono font at 13px with no line spacing increase.

---

## Spacing Scale

### Philosophy

Spacing follows an **8px grid** with a **4px fine grid** for micro-adjustments.
Every margin, padding, and gap is a multiple of 4px. This creates a
mathematical rhythm that feels intentional.

### Grid

```
── BASE UNIT: 4px ───────────────────────────────────

0    0px
1    4px    ← Micro: icon gaps, inline badges
2    8px    ← Tight: button padding, tag gaps
3    12px   ← Comfortable: card padding (small)
4    16px   ← Default: card padding, list gaps
5    20px   ← Generous: section spacing
6    24px   ← Page padding, modal padding
7    32px   ← Section separators
8    40px   ← Page section margins
9    48px   ← Major sections
10   64px   ← Page top padding
```

### Applied Spacing

| Context | Value |
|---------|-------|
| Sidebar item padding (vertical) | 8px |
| Sidebar item padding (horizontal) | 16px |
| Sidebar group gap | 4px |
| Card padding | 16px |
| Card gap (between children) | 12px |
| Table cell padding | 12px 16px |
| Tab bar padding | 0 16px |
| Tab bar gap | 4px |
| Button padding (horizontal) | 12px |
| Button padding (vertical) | 6px |
| Dialog padding | 24px |
| Page margin (horizontal) | 24px |
| Page margin (top) | 32px |
| Section gap (vertical) | 24px |
| List item gap | 4px |
| Badge padding | 2px 8px |

### Layout Grid

```
Page max width:    1280px   ← Content doesn't stretch infinitely
Sidebar width:     240px    ← Narrow enough to not compete with content
Top nav height:    48px     ← Compact, just logo + search + avatar
Tab bar height:    40px     ← Tabs + optional action buttons
Content padding:   24px     ← Consistent breathing room
```

---

## Radius

```
Sharp       0px    ← Terminal output, log lines, code blocks
Small       4px    ← Inputs, buttons, cards
Medium      6px    ← Modals, dropdowns, command palette
Large       8px    ← Dialog cards, elevated panels
Full        9999px ← Badges, pills, status indicators
```

**Rule:** Avoid radii larger than 8px on any functional element. Large radii
make developer tools feel playful, not professional.

---

## Elevation

CloudOS avoids heavy box shadows. Instead, elevation is communicated through:

1. **Background color** — elevated surfaces use `Surface Elevated` (#1C1C1F dark / #FFFFFF light)
2. **Border** — elevated surfaces get a 1px `Border` stroke
3. **Subtle shadow** — only for the command palette and dropdowns

### Shadow Tokens

```
── DARK MODE ────────────────────────────────────────

Shadow Sm:   0 1px 2px rgba(0,0,0,0.3)   ← Cards (minimal)
Shadow Md:   0 4px 12px rgba(0,0,0,0.4)  ← Dropdowns, tooltips
Shadow Lg:   0 8px 24px rgba(0,0,0,0.5)  ← Modals, command palette

── LIGHT MODE ───────────────────────────────────────

Shadow Sm:   0 1px 2px rgba(0,0,0,0.06)  ← Cards (minimal)
Shadow Md:   0 4px 12px rgba(0,0,0,0.08) ← Dropdowns, tooltips
Shadow Lg:   0 8px 24px rgba(0,0,0,0.12) ← Modals, command palette
```

---

## Borders

| Token | Width | Style | Usage |
|-------|-------|-------|-------|
| Default | 1px | Solid | Card borders, table cells, dividers |
| Strong | 1px | Solid | Active/focus states, selected items |
| Accent | 2px | Solid | Focus rings (`Accent` token) |
| Dashed | 1px | Dashed | Drag-and-drop zones, empty state containers |

**Rule:** Avoid border-radius on table cells. Tables should feel sharp and data-dense.

---

## Motion

### Philosophy

Motion in CloudOS serves a purpose: orienting the user, acknowledging actions,
and communicating state changes. It is never decorative. Every animation is
under 200ms — fast enough to feel instant, slow enough to perceive.

### Duration Tokens

```
Instant     100ms    ← Hover states, color transitions
Fast        150ms    ← Button clicks, tab switches, toggle states
Normal      200ms    ← Panel open/close, modal appear, navigation
Slow        300ms    ← Page transitions (rare — only for major context shifts)
```

### Easing Tokens

```
Standard:    cubic-bezier(0.16, 1, 0.3, 1)   ← Most animations — feels fast, lands smooth
Decelerate:  cubic-bezier(0, 0, 0.2, 1)      ← Elements entering the screen
Accelerate:  cubic-bezier(0.4, 0, 1, 1)      ← Elements leaving the screen
Spring:      cubic-bezier(0.34, 1.56, 0.64, 1) ← Micro-interactions (hearts, toggles)
```

### Animation Patterns

| Element | Animation | Duration | Easing |
|---------|-----------|----------|--------|
| Page enter | Fade in + slight slide up (8px) | 200ms | Decelerate |
| Page exit | Fade out | 100ms | Accelerate |
| Sidebar item hover | Background color change | 100ms | Standard |
| Tab switch | Underline slide + content cross-fade | 150ms | Standard |
| Modal appear | Scale (0.95 → 1) + fade in | 200ms | Decelerate |
| Modal dismiss | Scale (1 → 0.95) + fade out | 150ms | Accelerate |
| Dropdown open | Fade in + slide down (4px) | 150ms | Decelerate |
| Notification in | Slide in from right | 200ms | Decelerate |
| Status change | Icon cross-fade | 200ms | Standard |
| Deployment step | Step icon transition (◉ → ✓) | 300ms | Spring |
| Log line appear | Fade in (no slide) | 100ms | Standard |
| Hover (button) | Background + shadow | 100ms | Standard |
| Hover (card) | Border color | 150ms | Standard |
| Focus ring | Ring appear (no animation — instant) | 0ms | — |

**Rule:** If it moves, it should move in one direction — never bounce, never
wiggle, never stagger unless it's a deployment timeline (where staggered step
completion is intentional and satisfying).

---

## Iconography

### Philosophy

Use a single, consistent icon set. CloudOS uses **Lucide** — the same icon set
used by Vercel and Linear. Lucide icons are clean, consistent, and MIT-licensed.

**Alternative:** Phosphor Icons (used by Railway) — slightly more expressive,
also MIT-licensed.

**Decision:** Use **Lucide** for v0.6. It's already available via npm, has
consistent 24px/24px viewBox, 1.5px stroke width, and round caps/joins.

### Icon Rules

1. All icons are 16px in navigation, 20px in action buttons, 16px inline.
2. All icons use 1.5px stroke width. Never filled variants.
3. Icons never appear without a text label in navigation.
4. Icons in status badges are 12px.
5. Icons in the timeline are 16px.

### Icon Map

| Concept | Lucide Icon | Notes |
|---------|-------------|-------|
| Application | `Box` | A box — represents a deployed app |
| Deployment | `GitBranch` | A branch — represents a version/deployment |
| Workflow | `GitMerge` | A merge — steps coming together |
| Project | `Folder` | A folder — groups applications |
| Health | `Heart` | A heart — represents health/liveness |
| Logs | `Terminal` | A terminal — command-line output |
| Timeline | `List` | A list — step-by-step sequence |
| Plugin | `Puzzle` | A puzzle piece — extensibility |
| Runtime | `Server` | A server rack — execution environment |
| Buildpack | `Package` | A package — build toolchain |
| System | `Cpu` | A CPU — kernel/infrastructure |
| Settings | `Settings` | A gear — configuration |
| Search | `Search` | A magnifying glass |
| Notification | `Bell` | A bell — alerts |
| User | `User` | A person — user menu |
| Deploy (action) | `Rocket` | A rocket — deploy action |
| Open (action) | `ExternalLink` | External link — open app URL |
| Compare (action) | `GitCompare` | Git compare — diff deployments |
| Download (action) | `Download` | Download — log download |
| Close | `X` | Close — dismiss dialogs |
| Back | `ArrowLeft` | Back — navigate up |
| Menu | `PanelLeft` | Hamburger — mobile menu |
| Check | `Check` | Checkmark — success |
| Warning | `AlertTriangle` | Warning triangle |
| Error | `AlertCircle` | Error circle |
| Info | `Info` | Info circle |
| Plus | `Plus` | Add — create action |
| More | `MoreHorizontal` | More — overflow menu |
| Copy | `Copy` | Copy — clipboard |
| Refresh | `RotateCcw` | Refresh — reload data |

---

## Component Visual Language

### Buttons

```
── VARIANTS ─────────────────────────────────────────

Primary:    Accent background, white text, 6px radius
Secondary:  Transparent, 1px Border, Text Primary text
Ghost:      Transparent, text only, no border
Danger:     Danger background, white text
Icon-only:  Ghost variant, 32px square, centered icon

── SIZES ────────────────────────────────────────────

Small:      28px height, 8px horizontal padding
Default:    34px height, 12px horizontal padding
Large:      42px height, 16px horizontal padding

── STATES ───────────────────────────────────────────

Default:    Background (primary) or Border (secondary)
Hover:      Slightly darker background or stronger border
Active:     Scale to 0.97 + darker background
Disabled:   50% opacity, no hover effects
Focus:      2px Accent focus ring with 2px gap
Loading:    Replace icon with spinner, disable interaction
```

### Cards

```
── VARIANTS ─────────────────────────────────────────

Default:    Surface background, 1px Border, 6px radius
Elevated:   Surface Elevated background, 1px Border, 6px radius, Shadow Sm
Interactive: Default + hover: Border Hover + cursor pointer

── CONTENT ──────────────────────────────────────────

Padding:    16px
Title gap:  8px (between title and body)
Section:    12px (between card sections)

── APPLICATION CARD ─────────────────────────────────

Layout:     Icon | Name + URL | Status Badge | Health Dot
Height:     56px (compact) or auto (expanded with preview)
Hover:      Subtle border color change
```

### Tables

```
── STYLE ────────────────────────────────────────────

Header:     Text Secondary, 12px/16px, 500 weight, uppercase NOT ALLOWED
Cells:      Text Primary, 14px/20px, 400 weight
Borders:    1px Border, only between rows (no column borders)
Stripes:    None — every row is Surface background
Hover:      Accent Subtle background
Selected:   Accent Subtle background + 2px Accent left border

── LAYOUT ──────────────────────────────────────────

Cell padding:    12px 16px
Row height:      44px minimum
Column gap:      0 (borders handle separation)
First column:    16px left padding (aligns with card padding)
Last column:     Right-aligned actions
```

### Badges & Status Pills

```
── VARIANTS ─────────────────────────────────────────

Filled:     Background color + white text (strong emphasis)
Subtle:     Background tint + matching text (moderate emphasis)
Dot:        8px circle, no text (minimal — inline status)

── SIZING ──────────────────────────────────────────

Default:    12px/16px font, 2px 8px padding, 4px radius
Dot:        8px × 8px circle

── COLORS ──────────────────────────────────────────

Success:    Success token (green)
Warning:    Warning token (yellow/amber)
Danger:     Danger token (red)
Info:       Info token (blue)
Neutral:    Text Muted + Muted border
Accent:     Accent token (indigo)
```

### Tabs

```
── STYLE ────────────────────────────────────────────

Container:  1px bottom Border across full width
Tab:        Text Secondary, 14px/20px, 500 weight
Active:     Text Primary + 2px Accent bottom border (pill not rounded)
Hover:      Text Primary, no background change
Gap:        4px between tabs
Padding:    8px 16px per tab
Count:      Optional small badge on tab (e.g., "Deployments (12)")
```

### Dialogs

```
── STYLE ────────────────────────────────────────────

Background: Surface Elevated
Radius:     8px
Border:     1px Border
Shadow:     Shadow Lg
Overlay:    Background at 60% opacity (dark) / 40% (light)
Padding:    24px
Width:      480px (default), 640px (wide), 320px (narrow)
Transition: Scale 0.95→1 + fade, 200ms, Decelerate

── STRUCTURE ────────────────────────────────────────

Header:     Close button (top-right), optional title
Body:       Main content with 16px bottom margin
Footer:     Action buttons right-aligned, primary on right
```

### Inputs

```
── STYLE ────────────────────────────────────────────

Background: Surface
Border:     1px Border, 6px radius
Text:       Text Primary, 14px/20px
Padding:    8px 12px
Height:     34px
Placeholder: Text Muted
Focus:      2px Accent focus ring + Accent border
Disabled:   50% opacity
Error:      Danger border + Danger focus ring
Label:      Text Secondary, 13px/18px, 8px gap to input
Helper:     Text Muted, 12px/16px, 4px top gap
```

### Command Palette

```
── STYLE ────────────────────────────────────────────

Trigger:    Cmd+K (Mac) / Ctrl+K (Windows)
Backdrop:   Full-screen transparent overlay
Width:      640px
Max Height: 480px (scrollable)
Input:      Large — 16px/24px, 500 weight, no border, no background
Groups:     Separated by captions (Text Muted, 12px, 500 weight)
Items:      44px height, 14px/20px, optional description
Active:     Accent Subtle background
Radius:     8px
Shadow:     Shadow Lg
Transition: Scale 0.97→1 + fade, 200ms

── SECTIONS ─────────────────────────────────────────

Commands:   "Deploy app", "View logs", "Open dashboard"
Pages:      "Go to Applications", "Go to Settings"
Apps:       "go-api", "my-react-app" (shown as + recent)
Actions:    "Compare #41 and #42"
```

### Sidebar

```
── STYLE ────────────────────────────────────────────

Width:      240px
Background: Sidebar token
Text:       Text Secondary (default), Text Primary (active)
Item height: 36px
Icon:       16px, Text Secondary, 8px right gap
Active:     Accent Subtle background + Accent text
Hover:      Text Primary, no background
Group gap:  16px (between sections)
Divider:    1px Border, 8px margin
Version:    12px/16px, Text Muted, bottom of sidebar
Logo:       20px height, Text Primary, 16px top padding

── SECTION ORDER ────────────────────────────────────

1. Applications       ◆ Box icon
2. Deployments        ◆ GitBranch icon
   ── divider ──
3. Monitoring         ◆ Heart icon
4. Workflows          ◆ GitMerge icon
   ── divider ──
5. System             ◆ Cpu icon
6. Settings           ◆ Settings icon
7. Plugins            ◆ Puzzle icon
```

### Top Navigation

```
── STYLE ────────────────────────────────────────────

Height:     48px
Background: Top Nav token
Border:     1px bottom Border
Padding:    0 24px
Layout:     Logo (left) | Search (center) | Actions (right)

── ELEMENTS ─────────────────────────────────────────

Logo:       "CloudOS" text, 16px/24px, 600 weight
Search:     Cmd+K trigger, 200px wide (expands on focus)
Status:     Health indicator dot (green/yellow/red)
Avatar:     28px circle, Accent Subtle, initials
Bell:       Notification icon, 20px
```

### Timeline

```
── STYLE ────────────────────────────────────────────

Layout:     Vertical list with connecting line
Line:       2px width, Border color, left side (24px from edge)
Nodes:      16px circles on the line, centered
Spacing:    8px vertical gap between steps
Padding:    12px left (before line) + 16px between line and content

── STEP STATES ──────────────────────────────────────

Pending:    ○ Empty circle, Border color, no fill
Running:    ◉ Filled circle, Accent, pulsing animation
Succeeded:  ● Filled circle, Success, checkmark inside
Failed:     ● Filled circle, Danger, X inside
Skipped:    ○ Dashed circle, Muted, dash inside
Cancelled:  ⊘ Circle with slash, Warning

── CONTENT ──────────────────────────────────────────

Each step:
  Title:   14px/20px, 500 weight, Text Primary
  Duration: 12px/16px, Text Muted, right-aligned
  Detail:  13px/18px, Text Secondary, visible on expand
  Error:   13px/18px, Danger text, with remediation hint
```

### Terminal / Log Output

```
── STYLE ────────────────────────────────────────────

Background: #0C0C0D (pure dark, even in light mode)
Font:       13px/20px, Mono font
Text:       #D4D4D4 (classic terminal green-gray)
Padding:    12px 16px
Radius:     4px
Scroll:     Always visible, 8px wide, thin

── LINE ELEMENTS ───────────────────────────────────

Timestamp:  12px/20px, Text Muted, mono
Level:      12px icon (• info, ⚠ warn, ✗ error)
Source:     12px/20px, Accent, uppercase one-letter prefix
Message:    13px/20px, terminal text

── STATES ──────────────────────────────────────────

Streaming:  "● Live" indicator, 12px, Success, bottom-right
Paused:     "⏸ Paused" indicator, 12px, Warning, bottom-right
Error:      "Connection lost. Reconnecting..." inline banner
Empty:      "Waiting for logs..." with pulsing dot
```

---

## Accessibility Rules

### Color Contrast

- All text meets WCAG AA (4.5:1) minimum contrast ratio
- Large text (18px+ bold or 24px+ regular) meets WCAG AA (3:1)
- Status colors (green, yellow, red) are never the sole indicator — always paired with an icon or text label
- Focus rings are always 2px `Accent` with a 2px gap from the element

### Keyboard Navigation

| Key | Action |
|-----|--------|
| Tab | Move to next focusable element |
| Shift+Tab | Move to previous focusable element |
| Enter / Space | Activate focused element |
| Escape | Close modal, dropdown, command palette |
| Arrow Up/Down | Navigate list items, dropdown options |
| Cmd+K / Ctrl+K | Open command palette |
| / | Focus global search |
| Cmd+B | Toggle sidebar (future) |

### Focus Indicators

- All interactive elements have visible focus rings
- Focus rings use `Accent` color at 2px width
- Focus ring offset is 2px from the element
- Never use `outline: none` without providing an alternative focus style
- Custom components (tabs, selects, menus) use `role` and `aria-*` attributes

### Motion Sensitivity

- All animations respect `prefers-reduced-motion`
- When reduced motion is enabled, transitions happen instantly (0ms) or not at all
- The pulsing deployment indicator becomes a static icon
- Page transitions become instant cross-fades (no slide)

---

## Examples

### Application Card (Dark Mode)

```
┌──────────────────────────────────────────────────────────────┐
│  ◻ go-api                                   ● Healthy       │
│  http://localhost:31245                    ● Running        │
│                                                             │
│  Latest: #42 — 8.2s ago — ✓ Succeeded                      │
└──────────────────────────────────────────────────────────────┘

Background: Surface      #151517
Border:     1px          #26262B
Radius:     6px
Padding:    16px
Text:       Text Primary  #EDEDEF (name)
            Text Secondary #9D9DA3 (URL, meta)
Badge:      Success fill  #2B9D5D (Healthy)
            Accent fill   #5E6AD2 (Running)
```

### Status Badge (Subtle Variant)

```
┌──────────────┐
│ ● Healthy     │
└──────────────┘

Background: Success Subtle  #0D2B1A
Text:       Success         #2B9D5D
Dot:        Success         #2B9D5D
Font:       12px/16px, 500 weight
Radius:     4px
Padding:    2px 8px
```

### Deployment Timeline Step

```
   ┌──────────────────────────────────────────────────────┐
   │                                                       │
   │  ●  ✓  Build Artifact                    3.1s        │
   │      │  Build completed, binary=app                   │
   │      │                                                 │
   │  ●  ✗  Health Check                       1.6s        │
   │      │  HTTP 503 Service Unavailable                   │
   │      │  → Ensure your app listens on PORT              │
   │      │                                                 │
   │  ●  ✓  Complete Deployment               0.0s         │
   │                                                       │
   └──────────────────────────────────────────────────────┘

Line:       2px         Border      #26262B
Node:       16px        Success     #2B9D5D (✓)
                         Danger      #D45A5A (✗)
Title:      14px/20px   500         Text Primary  #EDEDEF
Duration:   12px/16px               Text Muted    #5F5F66
Detail:     13px/18px               Text Secondary #9D9DA3
Error:      13px/18px               Danger        #D45A5A
```

### Log Output

```
┌──────────────────────────────────────────────────────────────┐
│ 14:30:00 • App [build] Cloning repository...                 │
│ 14:30:01 • App [build] Detecting runtime: Go 1.24            │
│ 14:30:03 • App [build] Building binary...                    │
│ 14:30:05 ✓ App [deploy] Deploying application...              │
│ 14:30:06 ✓ App [health] Health check... HTTP 200              │
│ 14:30:06 ✓ App [deploy] Deployment #42 complete               │
│ 14:30:07 • App          Server listening on :31245            │
│                                                              │
│  ● Live                                                      │
└──────────────────────────────────────────────────────────────┘

Background:  #0C0C0D       ← Terminal dark
Font:        13px/20px     ← Mono
Timestamp:   12px          Text Muted   #5F5F66
Level icon:  • default     #5F5F66
             ✓ success     #2B9D5D
             ✗ error       #D45A5A
Source:      "App"         Accent       #5E6AD2
Step:        [step]        Text Muted   #5F5F66
Message:     13px          Terminal     #D4D4D4
Live dot:    12px          Success      #2B9D5D
```

---

## Ready for Implementation

This design system defines every visual aspect of CloudOS Console:

- **Color:** Full dark and light palettes with semantic tokens
- **Typography:** System fonts, mono for code, complete type scale
- **Spacing:** 4px base grid with applied values for every context
- **Components:** Visual spec for buttons, cards, tables, badges, tabs, dialogs,
  inputs, command palette, sidebar, top nav, timeline, and terminal
- **Motion:** Duration and easing tokens for every animated element
- **Icons:** Lucide icon set with a complete icon map
- **Accessibility:** Contrast ratios, keyboard navigation, focus indicators,
  reduced-motion support

A designer can recreate CloudOS in Figma using this document as the single
source of truth. An engineer can implement every component without referring
to any other document.

**The next step is Prompt #3 — Component Implementation.**
