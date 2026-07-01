import { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useWorkflows, type WorkflowExecution, type WorkflowStatus } from '@/hooks/useWorkflows';
import { usePageTitle } from '@/hooks/usePageTitle';
import { Badge, StatusDot } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, type SelectOption } from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import {
  GitMerge,
  Search,
  SlidersHorizontal,
  ArrowUpDown,
  ExternalLink,
  Activity,
  CheckCircle2,
  XCircle,
  Play,
  Clock,
  Timer,
  Layers,
  GitBranch,
  GitCommitHorizontal,
  Terminal,
} from 'lucide-react';
import { cn, relativeTime, truncate } from '@/lib/utils';

/* ── Status config ──────────────────────────────────────── */
const STATUS_LABELS: Record<WorkflowStatus, string> = {
  succeeded: 'Succeeded',
  failed: 'Failed',
  running: 'Running',
  pending: 'Pending',
  cancelled: 'Cancelled',
};

const STATUS_BADGE: Record<WorkflowStatus, string> = {
  succeeded: 'subtle-success',
  failed: 'subtle-danger',
  running: 'subtle-accent',
  pending: 'subtle-neutral',
  cancelled: 'subtle-warning',
};

/* ── Filter options ──────────────────────────────────────── */
const STATUS_OPTIONS: SelectOption[] = [
  { value: '', label: 'All Statuses' },
  { value: 'succeeded', label: 'Succeeded' },
  { value: 'failed', label: 'Failed' },
  { value: 'running', label: 'Running' },
  { value: 'pending', label: 'Pending' },
  { value: 'cancelled', label: 'Cancelled' },
];

const SORT_OPTIONS: SelectOption[] = [
  { value: 'newest', label: 'Newest First' },
  { value: 'oldest', label: 'Oldest First' },
  { value: 'duration', label: 'Longest Duration' },
];

/* ── Duration format ─────────────────────────────────────── */
function formatDuration(dur?: string): string {
  return dur || '\u2014';
}

/* ── Node count display ──────────────────────────────────── */
function NodeCount({ entry }: { entry: WorkflowExecution }) {
  const { nodeCount, completedNodes, failedNodes, status } = entry;
  if (!nodeCount) return <span className="text-caption text-text-muted">\u2014</span>;

  return (
    <span className="inline-flex items-center gap-1.5 text-caption tabular-nums text-text-secondary">
      <span className="text-success font-medium">{completedNodes}</span>
      <span className="text-text-muted">/</span>
      <span className={failedNodes > 0 ? 'text-danger font-medium' : 'text-foreground'}>{nodeCount}</span>
      {failedNodes > 0 && (
        <>
          <span className="text-text-muted">/</span>
          <span className="text-danger font-medium">{failedNodes}</span>
        </>
      )}
    </span>
  );
}

/* ── Desktop Row ────────────────────────────────────────── */
interface RowProps {
  entry: WorkflowExecution;
  onOpen: (id: string) => void;
}

function DesktopRow({ entry, onOpen }: RowProps) {
  const statusBadge = STATUS_BADGE[entry.status] as keyof typeof Badge extends never ? string : never;

  return (
    <div className="hidden md:flex items-center gap-3 px-4 py-2.5 border-b border-border last:border-b-0 hover:bg-accent-subtle/30 transition-colors min-h-[44px]">
      {/* Workflow ID */}
      <div className="w-[160px] shrink-0 min-w-0">
        <button
          type="button"
          onClick={() => onOpen(entry.id)}
          className="text-small font-mono text-accent hover:text-accent-hover truncate text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-sm"
        >
          {truncate(entry.id, 18)}
        </button>
      </div>

      {/* Application */}
      <div className="w-[120px] shrink-0 min-w-0">
        <span className="text-small text-foreground truncate block">{entry.appName}</span>
        <span className="text-caption text-text-muted tabular-nums">#{entry.deploymentNumber}</span>
      </div>

      {/* Status */}
      <div className="w-[110px] shrink-0">
        <Badge variant={statusBadge as any} className="gap-1.5 text-caption">
          <StatusDot status={
            entry.status === 'succeeded' ? 'success' :
            entry.status === 'failed' ? 'danger' :
            entry.status === 'running' ? 'deploying' :
            'pending'
          } pulsing={entry.status === 'running'} />
          {STATUS_LABELS[entry.status]}
        </Badge>
      </div>

      {/* Started */}
      <div className="w-[80px] shrink-0 text-caption text-text-muted tabular-nums">
        {relativeTime(entry.startedAt)}
      </div>

      {/* Duration */}
      <div className="w-[70px] shrink-0 text-caption text-text-muted tabular-nums text-right">
        {formatDuration(entry.duration)}
      </div>

      {/* Nodes */}
      <div className="w-[100px] shrink-0">
        <NodeCount entry={entry} />
      </div>

      {/* Branch / Commit */}
      <div className="w-[100px] shrink-0 min-w-0">
        {entry.branch ? (
          <span className="inline-flex items-center gap-1 text-caption text-text-secondary truncate">
            <GitBranch className="h-3 w-3 shrink-0" />
            <span className="truncate">{entry.branch}</span>
          </span>
        ) : (
          <span className="text-caption text-text-muted">\u2014</span>
        )}
      </div>

      {/* Environment */}
      <div className="w-[70px] shrink-0">
        {entry.environment ? (
          <Badge variant="subtle-neutral" className="text-caption uppercase tracking-wider">
            {entry.environment}
          </Badge>
        ) : (
          <span className="text-caption text-text-muted">\u2014</span>
        )}
      </div>

      {/* Action */}
      <div className="flex-1 flex justify-end">
        <Button variant="icon-ghost" size="icon-sm" aria-label={`View workflow ${entry.id}`} onClick={() => onOpen(entry.id)}>
          <ExternalLink className="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  );
}

/* ── Mobile Card ────────────────────────────────────────── */
function MobileCard({ entry, onOpen }: RowProps) {
  return (
    <div className="md:hidden rounded-md border border-border bg-surface p-3 space-y-2">
      {/* Header */}
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <button
            type="button"
            onClick={() => onOpen(entry.id)}
            className="text-small font-mono text-accent hover:text-accent-hover text-left truncate block w-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-sm"
          >
            {truncate(entry.id, 24)}
          </button>
          <div className="flex items-center gap-2 mt-0.5">
            <span className="text-caption text-text-muted">{entry.appName} #{entry.deploymentNumber}</span>
            {entry.environment && (
              <Badge variant="subtle-neutral" className="text-caption uppercase tracking-wider">
                {entry.environment}
              </Badge>
            )}
          </div>
        </div>
        <Badge variant={STATUS_BADGE[entry.status] as any} className="gap-1.5 text-caption shrink-0">
          <StatusDot status={
            entry.status === 'succeeded' ? 'success' :
            entry.status === 'failed' ? 'danger' :
            entry.status === 'running' ? 'deploying' :
            'pending'
          } pulsing={entry.status === 'running'} />
          {STATUS_LABELS[entry.status]}
        </Badge>
      </div>

      {/* Details */}
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-caption text-text-muted">
        <span className="tabular-nums">{relativeTime(entry.startedAt)}</span>
        <span className="inline-flex items-center gap-1 tabular-nums">
          <Timer className="h-3 w-3" />
          {formatDuration(entry.duration)}
        </span>
        <span className="inline-flex items-center gap-1">
          <Layers className="h-3 w-3" />
          <NodeCount entry={entry} />
        </span>
        {entry.branch && (
          <span className="inline-flex items-center gap-1">
            <GitBranch className="h-3 w-3" />
            {entry.branch}
          </span>
        )}
        {entry.runtime && (
          <span className="inline-flex items-center gap-1">
            <Terminal className="h-3 w-3" />
            {entry.runtime}
          </span>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-1 pt-1 border-t border-border">
        <Button variant="secondary" size="sm" className="h-7 gap-1 text-caption" onClick={() => onOpen(entry.id)}>
          <Activity className="h-3 w-3" />
          View Graph
        </Button>
      </div>
    </div>
  );
}

/* ── Loading skeletons ──────────────────────────────────── */
function DesktopSkeletonRows() {
  return (
    <div className="hidden md:block">
      {Array.from({ length: 8 }).map((_, i) => (
        <div key={i} className="flex items-center gap-3 px-4 py-2.5 border-b border-border min-h-[44px]">
          <Skeleton className="h-4 w-[140px]" />
          <Skeleton className="h-4 w-[100px]" />
          <Skeleton className="h-5 w-[100px] rounded-sm" />
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-4 w-12" />
          <Skeleton className="h-4 w-20" />
          <Skeleton className="h-4 w-[80px]" />
          <Skeleton className="h-4 w-12" />
          <div className="flex-1 flex justify-end">
            <Skeleton className="h-7 w-7 rounded-sm" />
          </div>
        </div>
      ))}
    </div>
  );
}

function MobileSkeletonCards() {
  return (
    <div className="md:hidden space-y-3">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="rounded-md border border-border bg-surface p-3 space-y-2">
          <div className="flex items-start justify-between gap-2">
            <div className="space-y-1.5 flex-1">
              <Skeleton className="h-4 w-36" />
              <Skeleton className="h-3 w-24" />
            </div>
            <Skeleton className="h-5 w-24 rounded-sm" />
          </div>
          <div className="flex gap-2">
            <Skeleton className="h-3 w-16" />
            <Skeleton className="h-3 w-12" />
          </div>
          <div className="flex gap-1 pt-1 border-t border-border">
            <Skeleton className="h-7 w-24 rounded-sm" />
          </div>
        </div>
      ))}
    </div>
  );
}

/* ── Main Page ──────────────────────────────────────────── */
export default function WorkflowsPage() {
  usePageTitle('Workflows');
  const navigate = useNavigate();
  const { workflows, stats, isLoading, error, refetch } = useWorkflows();

  // ── Search & filter state ──
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [appFilter, setAppFilter] = useState('');
  const [runtimeFilter, setRuntimeFilter] = useState('');
  const [envFilter, setEnvFilter] = useState('');
  const [sortOrder, setSortOrder] = useState('newest');
  const [showFilters, setShowFilters] = useState(false);

  // ── Uniq filter options ──
  const filterOptions = useMemo(() => {
    const apps = new Set<string>();
    const runtimes = new Set<string>();
    const envs = new Set<string>();

    for (const w of workflows) {
      apps.add(w.appName);
      if (w.runtime) runtimes.add(w.runtime);
      if (w.environment) envs.add(w.environment);
    }

    return {
      appOptions: [{ value: '', label: 'All Applications' }, ...Array.from(apps).map((a) => ({ value: a, label: a }))],
      runtimeOptions: [{ value: '', label: 'All Runtimes' }, ...Array.from(runtimes).map((r) => ({ value: r, label: r }))],
      envOptions: [{ value: '', label: 'All Environments' }, ...Array.from(envs).map((e) => ({ value: e, label: e }))],
    };
  }, [workflows]);

  // ── Filtered + sorted ──
  const filtered = useMemo(() => {
    let result = [...workflows];

    // Search
    if (search.trim()) {
      const q = search.toLowerCase();
      result = result.filter(
        (w) =>
          w.id.toLowerCase().includes(q) ||
          w.appName.toLowerCase().includes(q) ||
          w.appId.toLowerCase().includes(q) ||
          w.branch?.toLowerCase().includes(q) ||
          w.commitSha?.toLowerCase().includes(q),
      );
    }

    // Status filter
    if (statusFilter) result = result.filter((w) => w.status === statusFilter);

    // App filter
    if (appFilter) result = result.filter((w) => w.appName === appFilter);

    // Runtime filter
    if (runtimeFilter) result = result.filter((w) => w.runtime === runtimeFilter);

    // Environment filter
    if (envFilter) result = result.filter((w) => w.environment === envFilter);

    // Sort
    if (sortOrder === 'oldest') {
      result.sort((a, b) => a.startedAt.localeCompare(b.startedAt));
    } else if (sortOrder === 'duration') {
      result.sort((a, b) => (b.duration || '').localeCompare(a.duration || ''));
    } else {
      result.sort((a, b) => b.startedAt.localeCompare(a.startedAt));
    }

    return result;
  }, [workflows, search, statusFilter, appFilter, runtimeFilter, envFilter, sortOrder]);

  const hasActiveFilters = search || statusFilter || appFilter || runtimeFilter || envFilter;

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="flex flex-col gap-6"
    >
      {/* ══════════ HEADER ══════════ */}
      <div className="flex flex-col gap-1">
        <h1 className="text-h1 text-foreground">Workflows</h1>
        <p className="text-small text-text-secondary">
          {isLoading
            ? 'Loading...'
            : error
              ? 'Unable to load workflow executions'
              : `${stats.total} execution${stats.total === 1 ? '' : 's'} across all applications`}
        </p>
      </div>

      {/* ══════════ STATS ROW ══════════ */}
      {!isLoading && !error && workflows.length > 0 && (
        <div className="flex items-center gap-4">
          <div className="inline-flex items-center gap-4 rounded-md border border-border bg-surface px-4 py-2">
            <span className="flex items-center gap-1.5 text-small text-text-secondary">
              <Activity className="h-3.5 w-3.5" />
              Total
              <span className="font-semibold text-foreground tabular-nums">{stats.total}</span>
            </span>
            <span className="w-px h-4 bg-border" aria-hidden="true" />
            <span className="flex items-center gap-1.5 text-small text-text-secondary">
              <Play className="h-3.5 w-3.5 text-accent" />
              Running
              <span className="font-semibold text-foreground tabular-nums">{stats.running}</span>
            </span>
            <span className="w-px h-4 bg-border" aria-hidden="true" />
            <span className="flex items-center gap-1.5 text-small text-text-secondary">
              <CheckCircle2 className="h-3.5 w-3.5 text-success" />
              Succeeded
              <span className="font-semibold text-foreground tabular-nums">{stats.succeeded}</span>
            </span>
            <span className="w-px h-4 bg-border" aria-hidden="true" />
            <span className="flex items-center gap-1.5 text-small text-text-secondary">
              <XCircle className="h-3.5 w-3.5 text-danger" />
              Failed
              <span className="font-semibold text-foreground tabular-nums">{stats.failed}</span>
            </span>
          </div>
        </div>
      )}

      {/* ══════════ SEARCH + FILTERS ══════════ */}
      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-2">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted pointer-events-none" />
            <Input
              placeholder="Search by workflow ID, app, branch..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-8 h-[34px]"
              aria-label="Search workflows"
            />
          </div>

          <Button
            variant="secondary"
            size="sm"
            onClick={() => setShowFilters(!showFilters)}
            className={cn('gap-1.5', showFilters && 'bg-accent-subtle text-accent border-accent/30')}
            aria-label={showFilters ? 'Hide filters' : 'Show filters'}
            aria-expanded={showFilters}
          >
            <SlidersHorizontal className="h-3.5 w-3.5" />
            Filters
          </Button>

          <Select
            options={SORT_OPTIONS}
            value={sortOrder}
            onChange={(e) => setSortOrder(e.target.value)}
            className="w-36"
            aria-label="Sort order"
          />
        </div>

        {/* Expandable filters */}
        {showFilters && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.15, ease: [0, 0, 0.2, 1] }}
            className="flex flex-wrap items-center gap-3 overflow-hidden"
          >
            <Select label="Status" options={STATUS_OPTIONS} value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} className="w-36" />
            <Select label="Application" options={filterOptions.appOptions} value={appFilter} onChange={(e) => setAppFilter(e.target.value)} className="w-44" />
            <Select label="Runtime" options={filterOptions.runtimeOptions} value={runtimeFilter} onChange={(e) => setRuntimeFilter(e.target.value)} className="w-36" />
            <Select label="Environment" options={filterOptions.envOptions} value={envFilter} onChange={(e) => setEnvFilter(e.target.value)} className="w-36" />
            {hasActiveFilters && (
              <Button variant="ghost" size="sm" onClick={() => { setSearch(''); setStatusFilter(''); setAppFilter(''); setRuntimeFilter(''); setEnvFilter(''); }} className="gap-1 text-small">
                Clear filters
              </Button>
            )}
          </motion.div>
        )}
      </div>

      {/* ══════════ ERROR ══════════ */}
      {!isLoading && error && (
        <ErrorState
          title="Failed to load workflows"
          message={(error as Error)?.message || 'An unexpected error occurred.'}
          onRetry={() => refetch()}
        />
      )}

      {/* ══════════ EMPTY (no workflows) ══════════ */}
      {!isLoading && !error && workflows.length === 0 && (
        <EmptyState
          icon={GitMerge}
          title="No workflow executions"
          description="Workflow executions appear when you deploy an application. Each deployment generates a workflow with execution steps."
        />
      )}

      {/* ══════════ LOADING ══════════ */}
      {isLoading && (
        <>
          <DesktopSkeletonRows />
          <MobileSkeletonCards />
        </>
      )}

      {/* ══════════ WORKFLOW LIST ══════════ */}
      {!isLoading && !error && workflows.length > 0 && (
        <>
          {/* Filtered empty state */}
          {filtered.length === 0 && (
            <EmptyState
              icon={Search}
              title="No matching workflows"
              description="Try adjusting your search or filters."
              actions={[
                { label: 'Clear filters', onClick: () => { setSearch(''); setStatusFilter(''); setAppFilter(''); setRuntimeFilter(''); setEnvFilter(''); }, variant: 'secondary' },
              ]}
            />
          )}

          {/* Desktop table */}
          <div className="hidden md:block rounded-md border border-border overflow-hidden">
            {/* Column headers */}
            <div className="flex items-center gap-3 px-4 py-2 bg-surface border-b border-border text-caption font-medium text-text-secondary uppercase tracking-wider">
              <div className="w-[160px] shrink-0">Workflow ID</div>
              <div className="w-[120px] shrink-0">Application</div>
              <div className="w-[110px] shrink-0">Status</div>
              <div className="w-[80px] shrink-0">Started</div>
              <div className="w-[70px] shrink-0 text-right">Duration</div>
              <div className="w-[100px] shrink-0">Nodes</div>
              <div className="w-[100px] shrink-0">Branch</div>
              <div className="w-[70px] shrink-0">Env</div>
              <div className="flex-1" />
            </div>

            {filtered.map((entry) => (
              <DesktopRow
                key={entry.id}
                entry={entry}
                onOpen={(id) => navigate(`/workflows/${encodeURIComponent(id)}`)}
              />
            ))}
          </div>

          {/* Mobile cards */}
          <div className="md:hidden space-y-2">
            {filtered.map((entry) => (
              <MobileCard
                key={entry.id}
                entry={entry}
                onOpen={(id) => navigate(`/workflows/${encodeURIComponent(id)}`)}
              />
            ))}
          </div>

          {/* Footer */}
          <div className="flex items-center justify-center text-caption text-text-muted pt-1">
            {hasActiveFilters && filtered.length !== workflows.length
              ? `Showing ${filtered.length} of ${workflows.length} executions`
              : `${workflows.length} execution${workflows.length === 1 ? '' : 's'}`}
          </div>
        </>
      )}
    </motion.div>
  );
}
