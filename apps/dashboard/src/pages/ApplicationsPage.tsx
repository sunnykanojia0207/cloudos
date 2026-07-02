import { useState, useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useApplications } from '@/hooks/useApplications';
import { usePageTitle } from '@/hooks/usePageTitle';
import { DeployDialog } from '@/components/applications/DeployDialog';
import { ApplicationCard } from '@/components/applications/ApplicationCard';
import { Skeleton } from '@/components/ui/skeleton';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, type SelectOption } from '@/components/ui/select';
import {
  Rocket,
  Search,
  SlidersHorizontal,
  ArrowUpDown,
  Layers,
} from 'lucide-react';
import { cn } from '@/lib/utils';

/* ── Types ───────────────────────────────────────────────── */
type SortKey = 'name' | 'updated' | 'duration' | 'health';

/* ── Filter/Sort state ──────────────────────────────────── */
const STATUS_OPTIONS: SelectOption[] = [
  { value: '', label: 'All Statuses' },
  { value: 'Running', label: 'Running' },
  { value: 'Deploying', label: 'Deploying' },
  { value: 'Failed', label: 'Failed' },
  { value: 'Stopped', label: 'Stopped' },
];

const RUNTIME_OPTIONS: SelectOption[] = [
  { value: '', label: 'All Runtimes' },
  { value: 'Go', label: 'Go' },
  { value: 'Node', label: 'Node.js' },
  { value: 'Python', label: 'Python' },
  { value: 'React', label: 'React' },
  { value: 'Next.js', label: 'Next.js' },
  { value: 'Static', label: 'Static' },
];

const SORT_OPTIONS: SelectOption[] = [
  { value: 'updated', label: 'Recently Deployed' },
  { value: 'name', label: 'Alphabetical' },
  { value: 'duration', label: 'Deploy Duration' },
  { value: 'health', label: 'Health' },
];

/* ── Animation variants ─────────────────────────────────── */
const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.05 },
  },
};

/* ── Loading skeleton grid ──────────────────────────────── */
function ApplicationSkeletonGrid() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="rounded-md border border-border bg-surface p-4 space-y-3">
          {/* Top row: icon + name + badge */}
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-center gap-3 min-w-0">
              <Skeleton className="h-9 w-9 rounded-md shrink-0" />
              <div className="min-w-0 space-y-1.5 flex-1">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-3 w-20" />
              </div>
            </div>
            <Skeleton className="h-5 w-16 rounded-sm shrink-0" />
          </div>
          {/* Meta row */}
          <div className="flex flex-wrap gap-3">
            <Skeleton className="h-3.5 w-16" />
            <Skeleton className="h-3.5 w-14" />
            <Skeleton className="h-3.5 w-12" />
          </div>
          {/* Details row */}
          <div className="flex flex-wrap gap-3">
            <Skeleton className="h-3.5 w-24" />
            <Skeleton className="h-3.5 w-16" />
            <Skeleton className="h-3.5 w-14" />
          </div>
          {/* Footer */}
          <div className="flex items-center gap-1.5 pt-1 border-t border-border">
            <Skeleton className="h-7 w-14 rounded-sm" />
            <Skeleton className="h-7 w-14 rounded-sm" />
            <Skeleton className="h-7 w-14 rounded-sm" />
            <div className="flex-1" />
            <Skeleton className="h-7 w-7 rounded-sm" />
          </div>
        </div>
      ))}
    </div>
  );
}

/* ── Main Page ──────────────────────────────────────────── */
export default function ApplicationsPage() {
  usePageTitle('Applications');
  const [deployOpen, setDeployOpen] = useState(false);
  const {
    data: applications,
    isLoading,
    error,
    refetch,
  } = useApplications();

  // ── Search & filter state ──────────────────────────────
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [runtimeFilter, setRuntimeFilter] = useState('');
  const [sortKey, setSortKey] = useState<SortKey>('updated');
  const [showFilters, setShowFilters] = useState(false);

  // ── Filtered + sorted apps ─────────────────────────────
  const filteredApps = useMemo(() => {
    if (!applications) return [];

    let result = [...applications];

    // Search
    if (search.trim()) {
      const q = search.toLowerCase();
      result = result.filter((app) =>
        app.metadata.name.toLowerCase().includes(q) ||
        app.metadata.id.toLowerCase().includes(q) ||
        (app.spec?.source?.url ?? '').toLowerCase().includes(q) ||
        (app.spec?.runtime?.type ?? '').toLowerCase().includes(q) ||
        (app.status?.lastReport?.environment ?? '').toLowerCase().includes(q),
      );
    }

    // Status filter
    if (statusFilter) {
      result = result.filter((app) => app.status?.phase === statusFilter);
    }

    // Runtime filter
    if (runtimeFilter) {
      result = result.filter((app) => {
        const rt = app.spec?.runtime?.type ?? app.status?.lastReport?.detectedRuntime ?? '';
        return rt.toLowerCase() === runtimeFilter.toLowerCase();
      });
    }

    // Sort
    result.sort((a, b) => {
      switch (sortKey) {
        case 'name':
          return a.metadata.name.localeCompare(b.metadata.name);
        case 'updated': {
          const aTime = a.status?.lastReport?.startedAt ?? a.metadata.createdAt ?? '';
          const bTime = b.status?.lastReport?.startedAt ?? b.metadata.createdAt ?? '';
          return bTime.localeCompare(aTime); // newest first
        }
        case 'duration': {
          const aDur = a.status?.lastReport?.duration ?? '';
          const bDur = b.status?.lastReport?.duration ?? '';
          return aDur.localeCompare(bDur);
        }
        case 'health': {
          const order = { Healthy: 0, Running: 0, Degraded: 1, Warning: 1, Failed: 2, Error: 2, Stopped: 3 };
          const aH = order[a.status?.health as keyof typeof order] ?? 99;
          const bH = order[b.status?.health as keyof typeof order] ?? 99;
          return aH - bH;
        }
        default:
          return 0;
      }
    });

    return result;
  }, [applications, search, statusFilter, runtimeFilter, sortKey]);

  const appCount = applications?.length ?? 0;
  const filteredCount = filteredApps.length;
  const hasActiveFilters = search || statusFilter || runtimeFilter;

  return (
    <motion.div
      className="flex flex-col gap-6"
      variants={containerVariants}
      initial="hidden"
      animate="visible"
    >
      {/* ══════════════ PAGE HEADER ══════════════ */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="text-h1 text-foreground">Applications</h1>
          <p className="text-small text-text-secondary">
            {isLoading
              ? 'Loading...'
              : error
                ? 'Unable to load applications'
                : `${appCount} application${appCount === 1 ? '' : 's'} deployed`}
          </p>
        </div>

        <DeployDialog open={deployOpen} onOpenChange={setDeployOpen}>
          <Button
            variant="primary"
            size="default"
            className="gap-1.5 shrink-0"
          >
            <Rocket className="h-4 w-4" />
            Deploy Application
          </Button>
        </DeployDialog>
      </div>

      {/* ══════════════ SEARCH + FILTERS BAR ══════════════ */}
      <div className="flex flex-col gap-3">
        {/* Search row */}
        <div className="flex items-center gap-2">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted pointer-events-none" aria-hidden="true" />
            <Input
              placeholder="Search by name, repo, runtime, or environment..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-8 h-[34px]"
              aria-label="Search applications"
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
            value={sortKey}
            onChange={(e) => setSortKey(e.target.value as SortKey)}
            className="w-44"
            aria-label="Sort applications"
          />
        </div>

        {/* Expandable filter row */}
        <AnimatePresence>
          {showFilters && (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: 'auto', opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={{ duration: 0.15, ease: [0, 0, 0.2, 1] }}
              className="flex flex-wrap items-center gap-3 overflow-hidden"
            >
              <Select
                label="Status"
                options={STATUS_OPTIONS}
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="w-36"
              />
              <Select
                label="Runtime"
                options={RUNTIME_OPTIONS}
                value={runtimeFilter}
                onChange={(e) => setRuntimeFilter(e.target.value)}
                className="w-36"
              />

              {hasActiveFilters && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setSearch('');
                    setStatusFilter('');
                    setRuntimeFilter('');
                  }}
                  className="gap-1 text-small"
                >
                  Clear filters
                </Button>
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* ══════════════ LOADING STATE ══════════════ */}
      {isLoading && <ApplicationSkeletonGrid />}

      {/* ══════════════ ERROR STATE ══════════════ */}
      {!isLoading && error && (
        <ErrorState
          title="Failed to load applications"
          message={
            (error as Error)?.message ||
            'An unexpected error occurred while loading your applications. Ensure the CloudOS kernel is running.'
          }
          onRetry={() => refetch()}
        />
      )}

      {/* ══════════════ EMPTY STATE ══════════════ */}
      {!isLoading && !error && appCount === 0 && (
        <EmptyState
          icon={Layers}
          title="No applications yet"
          description="Deploy your first application from a Git repository. CloudOS will automatically detect the stack, build it, and give you a running URL."
          actions={[
            {
              label: 'Deploy Application',
              icon: Rocket,
              onClick: () => setDeployOpen(true),
              variant: 'primary',
            },
            {
              label: 'Open Quick Start',
              onClick: () => window.open('https://cloudos.io/docs/quick-start', '_blank'),
              variant: 'secondary',
            },
          ]}
        />
      )}

      {/* ══════════════ NO RESULTS (with filters) ══════════════ */}
      {!isLoading && !error && appCount > 0 && filteredCount === 0 && (
        <EmptyState
          icon={Search}
          title="No matching applications"
          description={`No applications match your search or filters. Try adjusting your criteria.`}
          actions={[
            {
              label: 'Clear filters',
              onClick: () => {
                setSearch('');
                setStatusFilter('');
                setRuntimeFilter('');
              },
              variant: 'secondary',
            },
          ]}
        />
      )}

      {/* ══════════════ APPLICATION GRID ══════════════ */}
      {!isLoading && !error && filteredCount > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {filteredApps.map((app) => (
            <ApplicationCard key={app.metadata.id} app={app} />
          ))}
        </div>
      )}

      {/* ══════════════ RESULTS FOOTER ══════════════ */}
      {!isLoading && !error && appCount > 0 && (
        <div className="flex items-center justify-center gap-2 pt-2 pb-6 text-caption text-text-muted">
          <span>
            Showing {filteredCount} of {appCount} application{appCount === 1 ? '' : 's'}
          </span>
          {hasActiveFilters && filteredCount !== appCount && (
            <Button
              variant="ghost"
              size="sm"
              className="h-auto p-0 text-caption text-accent hover:text-accent-hover"
              onClick={() => {
                setSearch('');
                setStatusFilter('');
                setRuntimeFilter('');
              }}
            >
              Clear filters
            </Button>
          )}
        </div>
      )}
    </motion.div>
  );
}
