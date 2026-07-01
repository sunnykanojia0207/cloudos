'use client';

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  useMemo,
  useRef,
  type ReactNode,
} from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  GitBranch,
  Heart,
  GitMerge,
  Cpu,
  Settings,
  Puzzle,
  Folder,
  PlusCircle,
  Activity,
  Terminal,
  Rocket,
  Search,
  Globe,
  Database,
  Layers,
  Monitor,
  BookOpen,
  RotateCcw,
  Star,
  StarOff,
  Clock,
  Trash2,
  FileQuestion,
  RefreshCw,
  Zap,
} from 'lucide-react';
import {
  CommandDialog,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
  CommandShortcut,
} from '@/components/ui/command';
import { Badge } from '@/components/ui/badge';
import { HealthIndicator } from '@/components/ui/health-indicator';
import { useApplications } from '@/hooks/useApplications';
import { useWorkflows } from '@/hooks/useWorkflows';
import { useProjects, useControllers } from '@/hooks/useCloudOS';
import { useRecentItems, useFavorites, type RecentItem } from '@/hooks/useRecentFavorites';
import { cn, fuzzyMatch, truncate, type FuzzyMatchResult } from '@/lib/utils';
import { highlightMatch } from '@/lib/search-utils';

/* ════════════════════════════════════════════════════════════
   TYPES
   ════════════════════════════════════════════════════════════ */

interface SearchItem {
  id: string;
  type: 'page' | 'application' | 'deployment' | 'workflow' | 'project' | 'controller' | 'resource' | 'action';
  label: string;
  subtitle: string;
  path: string;
  icon: React.ElementType;
  keywords: string[];
        badge?: { label: string; variant: 'success' | 'warning' | 'destructive' | 'default' | 'secondary' | 'outline' };
  health?: string;
  runtime?: string;
  environment?: string;
}

type SearchResult = SearchItem & { score: number; matchIndices: number[] };

/* ════════════════════════════════════════════════════════════
   ICON STRING MAP (for recent items serialization)
   ════════════════════════════════════════════════════════════ */

const ICON_MAP: Record<string, React.ElementType> = {
  box: Box,
  'git-branch': GitBranch,
  heart: Heart,
  'git-merge': GitMerge,
  cpu: Cpu,
  settings: Settings,
  puzzle: Puzzle,
  folder: Folder,
  'plus-circle': PlusCircle,
  activity: Activity,
  terminal: Terminal,
  rocket: Rocket,
  globe: Globe,
  database: Database,
  layers: Layers,
  monitor: Monitor,
  'book-open': BookOpen,
  'rotate-ccw': RotateCcw,
  search: Search,
  'file-question': FileQuestion,
  'refresh-cw': RefreshCw,
  clock: Clock,
  zap: Zap,
};

export function iconFromString(name: string): React.ElementType {
  return ICON_MAP[name] ?? Box;
}

/* ════════════════════════════════════════════════════════════
   CONTEXT
   ════════════════════════════════════════════════════════════ */

interface CommandPaletteContextValue {
  open: boolean;
  setOpen: (open: boolean) => void;
  toggle: () => void;
}

const CommandPaletteContext = createContext<CommandPaletteContextValue | null>(null);

export function useCommandPalette(): CommandPaletteContextValue {
  const ctx = useContext(CommandPaletteContext);
  if (!ctx) throw new Error('useCommandPalette must be used within a <CommandPaletteProvider />');
  return ctx;
}

/* ════════════════════════════════════════════════════════════
   PAGES (static)
   ════════════════════════════════════════════════════════════ */

const PAGE_ITEMS: SearchItem[] = [
  { id: 'home', type: 'page', label: 'Applications', subtitle: 'Manage your applications', path: '/', icon: Box, keywords: ['apps', 'home', 'app list'] },
  { id: 'deployments', type: 'page', label: 'Deployments', subtitle: 'All deployment history', path: '/deployments', icon: GitBranch, keywords: ['deploys', 'releases', 'rollback'] },
  { id: 'monitoring', type: 'page', label: 'Monitoring', subtitle: 'Application health and metrics', path: '/monitoring', icon: Heart, keywords: ['health', 'metrics', 'status', 'alerts'] },
  { id: 'workflows', type: 'page', label: 'Workflows', subtitle: 'Execution engine and pipelines', path: '/workflows', icon: GitMerge, keywords: ['executions', 'engine', 'pipeline', 'ci', 'cd'] },
  { id: 'system', type: 'page', label: 'System', subtitle: 'Kernel and infrastructure info', path: '/system', icon: Cpu, keywords: ['kernel', 'controllers', 'infra', 'info'] },
  { id: 'projects', type: 'page', label: 'Projects', subtitle: 'Organize resources by project', path: '/projects', icon: Folder, keywords: ['groups', 'teams', 'organization'] },
  { id: 'controllers', type: 'page', label: 'Controllers', subtitle: 'Controller runtime management', path: '/controllers', icon: Activity, keywords: ['runtime', 'reconciler', 'operator'] },
  { id: 'resources', type: 'page', label: 'Resources', subtitle: 'Resource engine kinds', path: '/resources', icon: Database, keywords: ['kinds', 'crds', 'schemas'] },
  { id: 'plugins', type: 'page', label: 'Plugins', subtitle: 'CloudOS plugin ecosystem', path: '/plugins', icon: Puzzle, keywords: ['extensions', 'addons', 'marketplace'] },
  { id: 'settings', type: 'page', label: 'Settings', subtitle: 'CloudOS configuration', path: '/settings', icon: Settings, keywords: ['config', 'preferences', 'options'] },
];

/* ════════════════════════════════════════════════════════════
   QUICK ACTIONS
   ════════════════════════════════════════════════════════════ */

interface QuickAction {
  label: string;
  subtitle: string;
  icon: React.ElementType;
  keywords: string[];
  action: () => void;
}

function createQuickActions(navigate: ReturnType<typeof useNavigate>, setOpen: (v: boolean) => void): QuickAction[] {
  return [
    {
      label: 'Create Project',
      subtitle: 'Start a new project',
      icon: PlusCircle,
      keywords: ['new project', 'add project', 'create'],
      action: () => { navigate('/projects'); setOpen(false); },
    },
    {
      label: 'Deploy Application',
      subtitle: 'Trigger a new deployment',
      icon: Rocket,
      keywords: ['deploy', 'release', 'publish'],
      action: () => { navigate('/'); setOpen(false); },
    },
    {
      label: 'Open Monitoring',
      subtitle: 'View application health',
      icon: Heart,
      keywords: ['monitor', 'health', 'metrics', 'status'],
      action: () => { navigate('/monitoring'); setOpen(false); },
    },
    {
      label: 'Open Workflow Explorer',
      subtitle: 'Browse workflow executions',
      icon: GitMerge,
      keywords: ['workflow', 'executions', 'pipeline'],
      action: () => { navigate('/workflows'); setOpen(false); },
    },
    {
      label: 'Run Doctor',
      subtitle: 'System diagnostics',
      icon: RefreshCw,
      keywords: ['diagnostics', 'health check', 'system check'],
      action: () => { navigate('/system'); setOpen(false); },
    },
    {
      label: 'Open Documentation',
      subtitle: 'CloudOS docs',
      icon: BookOpen,
      keywords: ['docs', 'help', 'guide', 'manual'],
      action: () => { window.open('https://cloudos.dev/docs', '_blank'); setOpen(false); },
    },
    {
      label: 'View Logs',
      subtitle: 'Browse application logs',
      icon: Terminal,
      keywords: ['logs', 'logging', 'output'],
      action: () => { navigate('/'); setOpen(false); },
    },
    {
      label: 'Restart Application',
      subtitle: 'Restart an application',
      icon: RotateCcw,
      keywords: ['restart', 'reboot', 'reload'],
      action: () => { navigate('/'); setOpen(false); },
    },
  ];
}

/* ════════════════════════════════════════════════════════════
   PROVIDER
   ════════════════════════════════════════════════════════════ */

interface CommandPaletteProviderProps {
  children: ReactNode;
}

export function CommandPaletteProvider({ children }: CommandPaletteProviderProps) {
  const [open, setOpen] = useState(false);

  const toggle = useCallback(() => setOpen((v) => !v), []);

  // Cmd+K / Ctrl+K
  useEffect(() => {
    const handler = (e: globalThis.KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setOpen((v) => !v);
      }
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, []);

  // / shortcut → focus search
  useEffect(() => {
    const handler = () => setOpen(true);
    window.addEventListener('cloudos:focus-search', handler);
    return () => window.removeEventListener('cloudos:focus-search', handler);
  }, []);

  return (
    <CommandPaletteContext.Provider value={{ open, setOpen, toggle }}>
      {children}
      <CommandPalette />
    </CommandPaletteContext.Provider>
  );
}

/* ════════════════════════════════════════════════════════════
   DEBOUNCE
   ════════════════════════════════════════════════════════════ */

function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(id);
  }, [value, delay]);
  return debounced;
}

/* ════════════════════════════════════════════════════════════
   PALETTE COMPONENT
   ════════════════════════════════════════════════════════════ */

const DEBOUNCE_MS = 150;
const MAX_RECENT_DISPLAY = 8;
const FAVORITE_BOOST = 500;

function CommandPalette() {
  const { open, setOpen } = useCommandPalette();
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, DEBOUNCE_MS);
  const q = debouncedSearch.trim().toLowerCase();
  const navigate = useNavigate();

  // Data sources
  const { data: applications } = useApplications();
  const { workflows } = useWorkflows();
  const { data: projects } = useProjects();
  const { data: controllers } = useControllers();

  // Recent + Favorites
  const { items: recentItems, pushItem, removeItem, clearAll } = useRecentItems();
  const { items: favItems, isFavorite, toggleFavorite } = useFavorites();

  const quickActions = useMemo(() => createQuickActions(navigate, setOpen), [navigate, setOpen]);

  /* ── Build search index ────────────────────────────────── */
  const searchIndex = useMemo<SearchItem[]>(() => {
    const items: SearchItem[] = [...PAGE_ITEMS];

    // Applications
    for (const app of applications ?? []) {
      items.push({
        id: `app-${app.metadata.id}`,
        type: 'application',
        label: app.metadata.name,
        subtitle: `ID: ${truncate(app.metadata.id, 20)}`,
        path: `/applications/${app.metadata.id}`,
        icon: Box,
        keywords: [app.metadata.id, app.metadata.name, app.status?.health ?? '', 'app'],
        health: app.status?.health,
        badge: { label: app.status?.phase ?? 'unknown', variant: app.status?.health === 'Healthy' ? 'success' : app.status?.health === 'Warning' ? 'warning' : 'secondary' },
        runtime: app.status?.lastReport?.detectedRuntime ?? app.spec.runtime.type,
      });
    }

    // Workflows
    for (const wf of workflows) {
      items.push({
        id: `wf-${wf.id}`,
        type: 'workflow',
        label: truncate(wf.id, 28),
        subtitle: `${wf.appName} #${wf.deploymentNumber}`,
        path: `/workflows/${encodeURIComponent(wf.id)}`,
        icon: Activity,
        keywords: [wf.id, wf.appName, wf.status, `#${wf.deploymentNumber}`],
        badge: { label: wf.status, variant: wf.status === 'succeeded' ? 'success' : wf.status === 'failed' ? 'destructive' : wf.status === 'running' ? 'warning' : 'secondary' },
        runtime: wf.runtime,
        environment: wf.environment,
      });
    }

    // Projects
    for (const proj of projects?.items ?? []) {
      items.push({
        id: `proj-${proj.metadata.id}`,
        type: 'project',
        label: proj.spec.displayName ?? proj.metadata.id,
        subtitle: `ID: ${truncate(proj.metadata.id, 20)}`,
        path: `/projects/${proj.metadata.id}`,
        icon: Folder,
        keywords: [proj.metadata.id, proj.spec.displayName, proj.spec.environment ?? '', 'project'],
        environment: proj.spec.environment,
        badge: { label: proj.status?.phase ?? 'unknown', variant: proj.status?.phase === 'Active' ? 'success' : 'secondary' },
      });
    }

    // Controllers
    for (const ctl of controllers?.controllers ?? []) {
      items.push({
        id: `ctl-${ctl.name}`,
        type: 'controller',
        label: ctl.name,
        subtitle: `Kind: ${ctl.kind}`,
        path: `/controllers/${ctl.name}`,
        icon: Cpu,
        keywords: [ctl.name, ctl.kind, ctl.state ?? ''],
        badge: { label: ctl.state ?? 'unknown', variant: ctl.state === 'running' ? 'success' : ctl.state === 'failed' ? 'destructive' : 'secondary' },
      });
    }

    return items;
  }, [applications, workflows, projects, controllers]);

  /* ── Search (fuzzy match + rank) ───────────────────────── */
  const searchResults = useMemo<SearchResult[]>(() => {
    if (!q) return [];

    const results: SearchResult[] = [];

    for (const item of searchIndex) {
      // Score label
      const labelMatch = fuzzyMatch(item.label, q);
      // Score keywords
      let bestKwScore = 0;
      let bestKwIndices: number[] = [];
      for (const kw of item.keywords) {
        const km = fuzzyMatch(kw, q);
        if (km && km.score > bestKwScore) {
          bestKwScore = km.score;
          bestKwIndices = km.indices;
        }
      }

      const best = labelMatch ?? (bestKwScore > 0 ? { score: bestKwScore, indices: bestKwIndices } : null);
      if (!best) continue;

      // Boost favorites
      const favBoost = favItems.some((f) => f.id === item.id && f.type === item.type) ? FAVORITE_BOOST : 0;

      results.push({
        ...item,
        score: best.score + favBoost,
        matchIndices: best.indices,
      });
    }

    // Sort by score descending, then alphabetically
    results.sort((a, b) => {
      if (a.score !== b.score) return b.score - a.score;
      return a.label.localeCompare(b.label);
    });

    return results;
  }, [q, searchIndex, favItems]);

  /* ── Group results by type ─────────────────────────────── */
  const groupedResults = useMemo(() => {
    const groups: { heading: string; items: SearchResult[] }[] = [];
    const order = ['page', 'application', 'workflow', 'project', 'controller'] as const;
    const headings: Record<string, string> = {
      page: 'Pages',
      application: 'Applications',
      workflow: 'Workflows',
      project: 'Projects',
      controller: 'Controllers',
    };

    for (const type of order) {
      const items = searchResults.filter((r) => r.type === type);
      if (items.length) groups.push({ heading: headings[type], items });
    }

    return groups;
  }, [searchResults]);

  /* ── Filtered quick actions ────────────────────────────── */
  const filteredActions = useMemo(() => {
    if (!q) return quickActions;
    return quickActions.filter(
      (a) => fuzzyMatch(a.label, q) || a.keywords.some((kw) => fuzzyMatch(kw, q)),
    );
  }, [q, quickActions]);

  /* ── Recent items for empty search ─────────────────────── */
  const displayRecent = useMemo(() => {
    if (q) return [];
    return recentItems.slice(0, MAX_RECENT_DISPLAY);
  }, [q, recentItems]);

  /* ── Recent items that match search ────────────────────── */
  const matchingRecents = useMemo(() => {
    if (!q) return [];
    return recentItems
      .filter((r) => fuzzyMatch(r.label, q) || fuzzyMatch(r.subtitle, q))
      .slice(0, 4);
  }, [q, recentItems]);

  /* ── Handle select ─────────────────────────────────────── */
  const handleSelect = useCallback(
    (item: SearchItem | QuickAction) => {
      if ('path' in item && item.path) {
        navigate(item.path);
      } else if ('action' in item) {
        item.action();
        return;
      }
      // Push to recent
      if ('id' in item && 'type' in item) {
        pushItem({
          id: item.id,
          type: item.type,
          label: item.label,
          subtitle: item.subtitle,
          path: item.path ?? '/',
          icon: getIconName(item.icon),
        });
      }
      setOpen(false);
    },
    [navigate, setOpen, pushItem],
  );

  const handleActionSelect = useCallback(
    (action: QuickAction) => {
      action.action();
      setOpen(false);
    },
    [setOpen],
  );

  /* ── Handle recent item select ─────────────────────────── */
  const handleRecentSelect = useCallback(
    (recent: RecentItem) => {
      navigate(recent.path);
      pushItem(recent); // bump to top
      setOpen(false);
    },
    [navigate, pushItem, setOpen],
  );

  /* ── Reset search on close ─────────────────────────────── */
  const handleOpenChange = (next: boolean) => {
    setOpen(next);
    if (!next) setSearch('');
  };

  const inputRef = useRef<HTMLInputElement>(null);
  const announcementRef = useRef<HTMLDivElement>(null);

  // Announce result count for screen readers
  useEffect(() => {
    if (open && announcementRef.current) {
      const count = searchResults.length + filteredActions.length;
      const msg = q
        ? `${count} result${count !== 1 ? 's' : ''} for ${q}`
        : `${displayRecent.length} recent items`;
      announcementRef.current.textContent = msg;
    }
  }, [open, q, searchResults.length, filteredActions.length, displayRecent.length]);

  /* ── Render ────────────────────────────────────────────── */
  const showRecent = !q;
  const hasResults = groupedResults.length > 0 || filteredActions.length > 0 || matchingRecents.length > 0;

  return (
    <CommandDialog open={open} onOpenChange={handleOpenChange}>
      {/* Screen reader live region */}
      <div
        ref={announcementRef}
        role="status"
        aria-live="polite"
        className="sr-only"
      />

      <CommandInput
        ref={inputRef}
        placeholder="Search pages, apps, workflows, projects…"
        value={search}
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearch(e.target.value)}
        aria-label="Search CloudOS"
        role="combobox"
        aria-expanded={open}
        autoFocus
      />

      <CommandList role="listbox" aria-label="Search results">
        <CommandEmpty>
          {q ? 'No results found.' : 'Type to search...'}
        </CommandEmpty>

        {/* ═══ RECENT ITEMS (empty search) ═══ */}
        {showRecent && displayRecent.length > 0 && (
          <CommandGroup heading="Recent">
            {displayRecent.map((recent) => {
              const Icon = iconFromString(recent.icon);
              return (
                <CommandItem
                  key={`recent-${recent.id}-${recent.type}`}
                  value={`recent-${recent.label}`}
                  onSelect={() => handleRecentSelect(recent)}
                >
                  <Icon className="mr-2 h-4 w-4 text-text-muted" />
                  <span className="flex-1">{recent.label}</span>
                  <span className="mr-2 max-w-[160px] truncate text-caption text-text-muted">
                    {recent.subtitle}
                  </span>
                  <Clock className="h-3 w-3 text-text-muted/50" aria-hidden="true" />
                </CommandItem>
              );
            })}
            <CommandItem
              value="clear-recent"
              onSelect={() => clearAll()}
              className="text-caption text-text-muted"
            >
              <Trash2 className="mr-2 h-3 w-3" />
              Clear recent items
            </CommandItem>
          </CommandGroup>
        )}

        {/* ═══ MATCHING RECENTS (when searching) ═══ */}
        {!showRecent && matchingRecents.length > 0 && (
          <CommandGroup heading="Recent Matches">
            {matchingRecents.map((recent) => {
              const Icon = iconFromString(recent.icon);
              const lm = fuzzyMatch(recent.label, q);
              return (
                <CommandItem
                  key={`mrec-${recent.id}`}
                  value={`mrec-${recent.label}`}
                  onSelect={() => handleRecentSelect(recent)}
                >
                  <Icon className="mr-2 h-4 w-4 text-text-muted" />
                  <span className="flex-1">
                    {lm ? highlightMatch(recent.label, lm.indices) : recent.label}
                  </span>
                  <span className="text-caption text-text-muted">{recent.subtitle}</span>
                </CommandItem>
              );
            })}
          </CommandGroup>
        )}

        {/* ═══ SEARCH RESULTS (typed search) ═══ */}
        {groupedResults.map((group) => (
          <CommandGroup key={group.heading} heading={group.heading}>
            {group.items.map((item) => {
              const isFav = favItems.some((f) => f.id === item.id && f.type === item.type);
              const Icon = item.icon;
              const labelEl = item.matchIndices.length
                ? highlightMatch(item.label, item.matchIndices)
                : item.label;
              return (
                <CommandItem
                  key={item.id}
                  value={`${item.type}-${item.id}`}
                  onSelect={() => handleSelect(item)}
                >
                  <Icon className="mr-2 h-4 w-4 shrink-0 text-text-muted" />

                  <div className="flex flex-1 items-center gap-2 min-w-0">
                    <span className="flex-1 truncate">{labelEl}</span>

                    {/* Health indicator */}
                    {item.health && (
                      <HealthIndicator status={item.health.toLowerCase() as any} className="h-2 w-2" />
                    )}

                    {/* Badge */}
                    {item.badge && (
                      <Badge variant={item.badge.variant} className="shrink-0 text-[10px] px-1.5 py-0">
                        {item.badge.label}
                      </Badge>
                    )}

                    {/* Runtime badge */}
                    {item.runtime && (
                      <Badge variant="outline" className="shrink-0 text-[10px] px-1.5 py-0 font-mono">
                        {item.runtime}
                      </Badge>
                    )}

                    {/* Environment badge */}
                    {item.environment && (
                      <Badge
                        variant="outline"
                        className={cn(
                          'shrink-0 text-[10px] px-1.5 py-0',
                          item.environment === 'production' && 'border-emerald-500/40 text-emerald-400',
                          item.environment === 'staging' && 'border-amber-500/40 text-amber-400',
                        )}
                      >
                        {item.environment}
                      </Badge>
                    )}
                  </div>

                  {/* Favorite toggle */}
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      toggleFavorite(item.id, item.type);
                    }}
                    className="ml-1 p-0.5 rounded-sm text-text-muted hover:text-amber-400 transition-colors"
                    aria-label={isFav ? 'Remove from favorites' : 'Add to favorites'}
                    tabIndex={-1}
                  >
                    {isFav ? (
                      <Star className="h-3.5 w-3.5 fill-amber-400 text-amber-400" />
                    ) : (
                      <StarOff className="h-3.5 w-3.5" />
                    )}
                  </button>
                </CommandItem>
              );
            })}
          </CommandGroup>
        ))}

        {/* ═══ QUICK ACTIONS ═══ */}
        {filteredActions.length > 0 && (
          <CommandGroup heading="Quick Actions">
            {filteredActions.map((action) => {
              const ActionIcon = action.icon;
              const lm = q ? fuzzyMatch(action.label, q) : null;
              return (
                <CommandItem
                  key={`action-${action.label}`}
                  value={`action-${action.label}`}
                  onSelect={() => handleActionSelect(action)}
                >
                  <ActionIcon className="mr-2 h-4 w-4 text-primary" />
                  <span className="flex-1">
                    {lm ? highlightMatch(action.label, lm.indices) : action.label}
                  </span>
                  <span className="text-caption text-text-muted">{action.subtitle}</span>
                </CommandItem>
              );
            })}
          </CommandGroup>
        )}
      </CommandList>
    </CommandDialog>
  );
}

/* ── Helper: get icon name string from component ────────── */
function getIconName(icon: React.ElementType): string {
  for (const [name, comp] of Object.entries(ICON_MAP)) {
    if (comp === icon) return name;
  }
  return 'box';
}
