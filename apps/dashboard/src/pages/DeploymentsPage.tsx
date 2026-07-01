import { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useApplications, type DeploymentReport } from '@/hooks/useApplications';
import { usePageTitle } from '@/hooks/usePageTitle';
import { HealthIndicator } from '@/components/ui/health-indicator';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, type SelectOption } from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import {
  GitBranch,
  GitCommitHorizontal,
  Layers,
  Clock,
  Terminal,
  Rocket,
  Search,
  SlidersHorizontal,
  ArrowUpDown,
  ExternalLink,
  Eye,
  GitCompare,
  RotateCcw,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Calendar,
} from 'lucide-react';
import { cn, relativeTime, truncate } from '@/lib/utils';

/* ── Enriched deployment entry ────────────────────────── */
interface DeploymentEntry {
  appId: string;
  appName: string;
  report: DeploymentReport;
  sortTime: string; // ISO string for sorting
}

/* ── Time group helper ───────────────────────────────── */
type TimeGroup = 'today' | 'yesterday' | 'earlier';

function getTimeGroup(dateStr: string): TimeGroup {
  if (!dateStr) return 'earlier';
  const d = new Date(dateStr);
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);

  if (d >= today) return 'today';
  if (d >= yesterday) return 'yesterday';
  return 'earlier';
}

const GROUP_LABELS: Record<TimeGroup, string> = {
  today: 'Today',
  yesterday: 'Yesterday',
  earlier: 'Earlier',
};

/* ── Filter options ──────────────────────────────────── */
const STATUS_OPTIONS: SelectOption[] = [
  { value: '', label: 'All Statuses' },
  { value: 'success', label: 'Succeeded' },
  { value: 'failed', label: 'Failed' },
  { value: 'running', label: 'Running' },
];

const SORT_OPTIONS: SelectOption[] = [
  { value: 'newest', label: 'Newest First' },
  { value: 'oldest', label: 'Oldest First' },
];

/* ── Format deployment duration (parse "8.2s" etc.) ──── */
function formatDuration(dur?: string): string {
  return dur || '\u2014';
}

/* ── Desktop Row ──────────────────────────────────────── */
interface DesktopRowProps {
  entry: DeploymentEntry;
  onOpen: (appId: string, num: number) => void;
  onLogs: (appId: string, num: number) => void;
  onTimeline: (appId: string, num: number) => void;
  onCompare: (appId: string, num: number) => void;
  onRedeploy: (appId: string, num: number) => void;
}

function DesktopRow({ entry, onOpen, onLogs, onTimeline, onCompare, onRedeploy }: DesktopRowProps) {
  const { appId, appName, report } = entry;
  const success = report.buildSuccess;

  return (
    <div className="hidden md:flex items-center gap-3 px-4 py-2.5 border-b border-border last:border-b-0 hover:bg-accent-subtle/30 transition-colors min-h-[44px]">
      {/* App name + deployment # */}
      <div className="flex items-center gap-2 w-[180px] shrink-0 min-w-0">
        <button
          type="button"
          onClick={() => onOpen(appId, report.deploymentNumber)}
          className="text-small font-medium text-accent hover:text-accent-hover truncate text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-sm"
        >
          {appName}
        </button>
        <span className="text-caption text-text-muted tabular-nums shrink-0">
          #{report.deploymentNumber}
        </span>
      </div>

      {/* Status */}
      <div className="w-[100px] shrink-0">
        <HealthIndicator
          status={success ? 'healthy' : 'failed'}
          showLabel
          size="sm"
        />
      </div>

      {/* Commit */}
      <div className="w-[80px] shrink-0">
        {report.commitSha ? (
          <span className="inline-flex items-center gap-1 text-caption font-mono text-text-secondary">
            <GitCommitHorizontal className="h-3 w-3" />
            {truncate(report.commitSha, 7)}
          </span>
        ) : (
          <span className="text-caption text-text-muted">\u2014</span>
        )}
      </div>

      {/* Branch */}
      <div className="w-[100px] shrink-0 min-w-0">
        {report.branch ? (
          <span className="inline-flex items-center gap-1 text-caption text-text-secondary truncate">
            <GitBranch className="h-3 w-3 shrink-0" />
            <span className="truncate">{report.branch}</span>
          </span>
        ) : (
          <span className="text-caption text-text-muted">\u2014</span>
        )}
      </div>

      {/* Runtime */}
      <div className="w-[80px] shrink-0">
        {report.detectedRuntime ? (
          <span className="inline-flex items-center gap-1 text-caption text-text-secondary">
            <Terminal className="h-3 w-3" />
            {report.detectedRuntime}
          </span>
        ) : (
          <span className="text-caption text-text-muted">\u2014</span>
        )}
      </div>

      {/* Environment */}
      <div className="w-[80px] shrink-0">
        {report.environment ? (
          <Badge variant="subtle-neutral" className="text-caption uppercase tracking-wider">
            {report.environment}
          </Badge>
        ) : (
          <span className="text-caption text-text-muted">\u2014</span>
        )}
      </div>

      {/* Duration */}
      <div className="w-[70px] shrink-0 text-caption text-text-muted tabular-nums text-right">
        {formatDuration(report.duration)}
      </div>

      {/* Started */}
      <div className="w-[90px] shrink-0 text-caption text-text-muted tabular-nums text-right">
        {relativeTime(report.startedAt)}
      </div>

      {/* Actions */}
      <div className="flex-1 flex items-center justify-end gap-0.5">
        <Button variant="icon-ghost" size="icon-sm" aria-label="View timeline" onClick={() => onTimeline(appId, report.deploymentNumber)}>
          <Eye className="h-3.5 w-3.5" />
        </Button>
        <Button variant="icon-ghost" size="icon-sm" aria-label="View logs" onClick={() => onLogs(appId, report.deploymentNumber)}>
          <Terminal className="h-3.5 w-3.5" />
        </Button>
        <Button variant="icon-ghost" size="icon-sm" aria-label="Compare" onClick={() => onCompare(appId, report.deploymentNumber)} disabled={report.deploymentNumber <= 1}>
          <GitCompare className="h-3.5 w-3.5" />
        </Button>
        <Button variant="icon-ghost" size="icon-sm" aria-label="Redeploy" onClick={() => onRedeploy(appId, report.deploymentNumber)}>
          <RotateCcw className="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  );
}

/* ── Mobile Card ──────────────────────────────────────── */
function MobileCard({ entry, onOpen, onLogs, onTimeline }: DesktopRowProps) {
  const { appId, appName, report } = entry;
  const success = report.buildSuccess;

  return (
    <div className="md:hidden rounded-md border border-border bg-surface p-3 space-y-2">
      {/* Header */}
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <button
            type="button"
            onClick={() => onOpen(appId, report.deploymentNumber)}
            className="text-small font-medium text-accent hover:text-accent-hover text-left truncate block w-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-sm"
          >
            {appName}
          </button>
          <div className="flex items-center gap-2 mt-0.5">
            <span className="text-caption text-text-muted tabular-nums">#{report.deploymentNumber}</span>
            {report.environment && (
              <Badge variant="subtle-neutral" className="text-caption uppercase tracking-wider">
                {report.environment}
              </Badge>
            )}
          </div>
        </div>
        <HealthIndicator status={success ? 'healthy' : 'failed'} showLabel size="sm" />
      </div>

      {/* Details */}
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-caption text-text-muted">
        {report.commitSha && (
          <span className="inline-flex items-center gap-1 font-mono">
            <GitCommitHorizontal className="h-3 w-3" />
            {truncate(report.commitSha, 7)}
          </span>
        )}
        {report.branch && (
          <span className="inline-flex items-center gap-1">
            <GitBranch className="h-3 w-3" />
            {report.branch}
          </span>
        )}
        {report.detectedRuntime && (
          <span className="inline-flex items-center gap-1">
            <Terminal className="h-3 w-3" />
            {report.detectedRuntime}
          </span>
        )}
        <span className="inline-flex items-center gap-1 tabular-nums">
          <Clock className="h-3 w-3" />
          {formatDuration(report.duration)}
        </span>
        <span className="tabular-nums">{relativeTime(report.startedAt)}</span>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-1 pt-1 border-t border-border">
        <Button variant="secondary" size="sm" className="h-7 gap-1 text-caption" onClick={() => onTimeline(appId, report.deploymentNumber)}>
          <Eye className="h-3 w-3" />
          Timeline
        </Button>
        <Button variant="secondary" size="sm" className="h-7 gap-1 text-caption" onClick={() => onLogs(appId, report.deploymentNumber)}>
          <Terminal className="h-3 w-3" />
          Logs
        </Button>
      </div>
    </div>
  );
}

/* ── Loading skeleton (desktop) ───────────────────────── */
function DesktopSkeletonRows() {
  return (
    <div className="hidden md:block">
      {Array.from({ length: 8 }).map((_, i) => (
        <div key={i} className="flex items-center gap-3 px-4 py-2.5 border-b border-border min-h-[44px]">
          <Skeleton className="h-4 w-[160px]" />
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-4 w-20" />
          <Skeleton className="h-4 w-[70px]" />
          <Skeleton className="h-4 w-[90px]" />
          <Skeleton className="h-4 w-[70px]" />
          <Skeleton className="h-4 w-[70px]" />
          <Skeleton className="h-4 w-12" />
          <Skeleton className="h-4 w-[80px]" />
          <div className="flex-1 flex justify-end gap-1">
            <Skeleton className="h-7 w-7 rounded-sm" />
            <Skeleton className="h-7 w-7 rounded-sm" />
            <Skeleton className="h-7 w-7 rounded-sm" />
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
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-3 w-20" />
            </div>
            <Skeleton className="h-4 w-16" />
          </div>
          <div className="flex gap-2">
            <Skeleton className="h-3 w-16" />
            <Skeleton className="h-3 w-12" />
          </div>
          <div className="flex gap-1 pt-1 border-t border-border">
            <Skeleton className="h-7 w-20 rounded-sm" />
            <Skeleton className="h-7 w-16 rounded-sm" />
          </div>
        </div>
      ))}
    </div>
  );
}

/* ── Main Page ──────────────────────────────────────────── */
export default function DeploymentsPage() {
  usePageTitle('Deployments');
  const navigate = useNavigate();
  const { data: applications, isLoading, error, refetch } = useApplications();

  // ── Aggregate all deployments across apps ──
  const allDeployments = useMemo<DeploymentEntry[]>(() => {
    if (!applications) return [];
    const entries: DeploymentEntry[] = [];

    for (const app of applications) {
      const history = app.status?.deploymentHistory ?? [];
      for (const report of history) {
        entries.push({
          appId: app.metadata.id,
          appName: app.metadata.name,
          report,
          sortTime: report.startedAt || report.completedAt || app.metadata.createdAt || '',
        });
      }
    }

    // Sort newest first by default
    entries.sort((a, b) => b.sortTime.localeCompare(a.sortTime));

    return entries;
  }, [applications]);

  // ── Stats ──
  const stats = useMemo(() => {
    const total = allDeployments.length;
    const succeeded = allDeployments.filter((d) => d.report.buildSuccess).length;
    const failed = allDeployments.filter((d) => !d.report.buildSuccess).length;
    return { total, succeeded, failed };
  }, [allDeployments]);

  // ── Uniq values for filters ──
  const filterOptions = useMemo(() => {
    const apps = new Set<string>();
    const runtimes = new Set<string>();
    const envs = new Set<string>();

    for (const d of allDeployments) {
      apps.add(d.appName);
      if (d.report.detectedRuntime) runtimes.add(d.report.detectedRuntime);
      if (d.report.environment) envs.add(d.report.environment);
    }

    return {
      appOptions: [{ value: '', label: 'All Applications' }, ...Array.from(apps).map((a) => ({ value: a, label: a }))],
      runtimeOptions: [{ value: '', label: 'All Runtimes' }, ...Array.from(runtimes).map((r) => ({ value: r, label: r }))],
      envOptions: [{ value: '', label: 'All Environments' }, ...Array.from(envs).map((e) => ({ value: e, label: e }))],
    };
  }, [allDeployments]);

  // ── Search & filter state ──
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [appFilter, setAppFilter] = useState('');
  const [runtimeFilter, setRuntimeFilter] = useState('');
  const [envFilter, setEnvFilter] = useState('');
  const [sortOrder, setSortOrder] = useState('newest');
  const [showFilters, setShowFilters] = useState(false);

  // ── Filtered + sorted ──
  const filtered = useMemo(() => {
    let result = [...allDeployments];

    // Search
    if (search.trim()) {
      const q = search.toLowerCase();
      result = result.filter(
        (d) =>
          d.appName.toLowerCase().includes(q) ||
          d.appId.toLowerCase().includes(q) ||
          d.report.commitSha?.toLowerCase().includes(q) ||
          d.report.repository?.toLowerCase().includes(q) ||
          d.report.branch?.toLowerCase().includes(q),
      );
    }

    // Status filter
    if (statusFilter === 'success') result = result.filter((d) => d.report.buildSuccess);
    else if (statusFilter === 'failed') result = result.filter((d) => !d.report.buildSuccess);

    // App filter
    if (appFilter) result = result.filter((d) => d.appName === appFilter);

    // Runtime filter
    if (runtimeFilter) result = result.filter((d) => d.report.detectedRuntime === runtimeFilter);

    // Environment filter
    if (envFilter) result = result.filter((d) => d.report.environment === envFilter);

    // Sort
    if (sortOrder === 'oldest') {
      result.sort((a, b) => a.sortTime.localeCompare(b.sortTime));
    } else {
      result.sort((a, b) => b.sortTime.localeCompare(a.sortTime));
    }

    return result;
  }, [allDeployments, search, statusFilter, appFilter, runtimeFilter, envFilter, sortOrder]);

  // ── Group by time ──
  const grouped = useMemo(() => {
    const groups: Record<TimeGroup, DeploymentEntry[]> = {
      today: [],
      yesterday: [],
      earlier: [],
    };

    for (const entry of filtered) {
      const group = getTimeGroup(entry.report.startedAt || entry.report.completedAt);
      groups[group].push(entry);
    }

    return groups;
  }, [filtered]);

  const hasActiveFilters = search || statusFilter || appFilter || runtimeFilter || envFilter;

  // ── Actions ──
  const openDeployment = (appId: string, num: number) => {
    navigate(`/applications/${appId}/deployments/${num}`);
  };

  const openLogs = (appId: string, _num: number) => {
    navigate(`/applications/${appId}/timeline`);
  };

  const openTimeline = (appId: string, num: number) => {
    navigate(`/applications/${appId}/deployments/${num}`);
  };

  const openCompare = (appId: string, num: number) => {
    if (num <= 1) return;
    navigate(`/applications/${appId}/compare?from=${num - 1}&to=${num}`);
  };

  const openRedeploy = (appId: string, num: number) => {
    navigate(`/applications/${appId}/deployments/${num}/redeploy`);
  };

  /* ════════════════ RENDER ════════════════ */
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="flex flex-col gap-6"
    >
      {/* ══════════ HEADER ══════════ */}
      <div className="flex flex-col gap-1">
        <h1 className="text-h1 text-foreground">Deployments</h1>
        <p className="text-small text-text-secondary">
          {isLoading
            ? 'Loading...'
            : error
              ? 'Unable to load deployments'
              : `${stats.total} deployment${stats.total === 1 ? '' : 's'} across all applications`}
        </p>
      </div>

      {/* ══════════ STATS ROW ══════════ */}
      {!isLoading && !error && allDeployments.length > 0 && (
        <div className="flex items-center gap-4">
          <div className="inline-flex items-center gap-4 rounded-md border border-border bg-surface px-4 py-2">
            <span className="flex items-center gap-1.5 text-small text-text-secondary">
              <Rocket className="h-3.5 w-3.5" />
              Total
              <span className="font-semibold text-foreground tabular-nums">{stats.total}</span>
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
              placeholder="Search by app, commit, repo, or branch..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-8 h-[34px]"
              aria-label="Search deployments"
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
          title="Failed to load deployments"
          message={(error as Error)?.message || 'An unexpected error occurred.'}
          onRetry={() => refetch()}
        />
      )}

      {/* ══════════ EMPTY (no apps) ══════════ */}
      {!isLoading && !error && allDeployments.length === 0 && (
        <EmptyState
          icon={Rocket}
          title="No deployments yet"
          description="Deploy an application first. Deployments from all applications will appear here."
          actions={[
            { label: 'Deploy Application', icon: Rocket, onClick: () => navigate('/applications/deploy'), variant: 'primary' },
          ]}
        />
      )}

      {/* ══════════ LOADING ══════════ */}
      {isLoading && (
        <>
          <DesktopSkeletonRows />
          <MobileSkeletonCards />
        </>
      )}

      {/* ══════════ DEPLOYMENT LIST ══════════ */}
      {!isLoading && !error && allDeployments.length > 0 && (
        <>
          {/* Filtered empty state */}
          {filtered.length === 0 && (
            <EmptyState
              icon={Search}
              title="No matching deployments"
              description="Try adjusting your search or filters."
              actions={[
                { label: 'Clear filters', onClick: () => { setSearch(''); setStatusFilter(''); setAppFilter(''); setRuntimeFilter(''); setEnvFilter(''); }, variant: 'secondary' },
              ]}
            />
          )}

          {/* Desktop grouped table */}
          <div className="hidden md:block rounded-md border border-border overflow-hidden">
            {/* Column headers */}
            <div className="flex items-center gap-3 px-4 py-2 bg-surface border-b border-border text-caption font-medium text-text-secondary uppercase tracking-wider">
              <div className="w-[180px] shrink-0">Application</div>
              <div className="w-[100px] shrink-0">Status</div>
              <div className="w-[80px] shrink-0">Commit</div>
              <div className="w-[100px] shrink-0">Branch</div>
              <div className="w-[80px] shrink-0">Runtime</div>
              <div className="w-[80px] shrink-0">Env</div>
              <div className="w-[70px] shrink-0 text-right">Duration</div>
              <div className="w-[90px] shrink-0 text-right">When</div>
              <div className="flex-1" />
            </div>

            {/* Grouped rows */}
            {(['today', 'yesterday', 'earlier'] as TimeGroup[]).map((group) => {
              const entries = grouped[group];
              if (entries.length === 0) return null;
              return (
                <div key={group}>
                  <div className="px-4 py-1.5 bg-surface-elevated/50 border-b border-border">
                    <span className="text-caption font-semibold text-text-muted flex items-center gap-1.5">
                      <Calendar className="h-3 w-3" />
                      {GROUP_LABELS[group]}
                      <span className="tabular-nums font-normal">({entries.length})</span>
                    </span>
                  </div>
                  {entries.map((entry) => (
                    <DesktopRow
                      key={`${entry.appId}-${entry.report.deploymentNumber}`}
                      entry={entry}
                      onOpen={openDeployment}
                      onLogs={openLogs}
                      onTimeline={openTimeline}
                      onCompare={openCompare}
                      onRedeploy={openRedeploy}
                    />
                  ))}
                </div>
              );
            })}
          </div>

          {/* Mobile grouped cards */}
          <div className="md:hidden space-y-4">
            {(['today', 'yesterday', 'earlier'] as TimeGroup[]).map((group) => {
              const entries = grouped[group];
              if (entries.length === 0) return null;
              return (
                <div key={group}>
                  <h3 className="text-small font-semibold text-text-muted flex items-center gap-1.5 mb-2 px-1">
                    <Calendar className="h-3.5 w-3.5" />
                    {GROUP_LABELS[group]}
                    <span className="tabular-nums font-normal">({entries.length})</span>
                  </h3>
                  <div className="space-y-2">
                    {entries.map((entry) => (
                      <MobileCard
                        key={`${entry.appId}-${entry.report.deploymentNumber}`}
                        entry={entry}
                        onOpen={openDeployment}
                        onLogs={openLogs}
                        onTimeline={openTimeline}
                        onCompare={openCompare}
                        onRedeploy={openRedeploy}
                      />
                    ))}
                  </div>
                </div>
              );
            })}
          </div>

          {/* Footer */}
          <div className="flex items-center justify-center text-caption text-text-muted pt-1">
            {hasActiveFilters && filtered.length !== allDeployments.length
              ? `Showing ${filtered.length} of ${allDeployments.length} deployments`
              : `${allDeployments.length} deployment${allDeployments.length === 1 ? '' : 's'}`}
          </div>
        </>
      )}
    </motion.div>
  );
}
