import { useState, useMemo, useCallback } from 'react';
import { motion, AnimatePresence, type Variants } from 'framer-motion';
import {
  Search,
  Grid3X3,
  List,
  Plus,
  Layers,
  AlertCircle,
  RotateCcw,
  Check,
  SortAsc,
} from 'lucide-react';
import { useProjects } from '@/hooks/useCloudOS';
import { ProjectCard } from '@/components/projects/ProjectCard';
import { CreateProjectDialog } from '@/components/projects/CreateProjectDialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
} from '@/components/ui/dropdown-menu';
import type { ProjectDTO } from '@cloudos/sdk';

// ── Types ────────────────────────────────────────────────────────────────

type ViewMode = 'grid' | 'list';
type SortKey = 'name-asc' | 'name-desc' | 'created' | 'updated';
type EnvironmentFilter = 'all' | 'development' | 'staging' | 'production' | 'testing';

// ── Constants ─────────────────────────────────────────────────────────────

const ENVIRONMENTS: { value: EnvironmentFilter; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'development', label: 'Development' },
  { value: 'staging', label: 'Staging' },
  { value: 'production', label: 'Production' },
  { value: 'testing', label: 'Testing' },
];

const SORT_OPTIONS: { value: SortKey; label: string }[] = [
  { value: 'name-asc', label: 'Name (A\u2013Z)' },
  { value: 'name-desc', label: 'Name (Z\u2013A)' },
  { value: 'created', label: 'Recently Created' },
  { value: 'updated', label: 'Recently Updated' },
];

const SORT_LABEL_MAP: Record<SortKey, string> = {
  'name-asc': 'Name (A\u2013Z)',
  'name-desc': 'Name (Z\u2013A)',
  created: 'Recently Created',
  updated: 'Recently Updated',
};

// ── Animation Variants ────────────────────────────────────────────────────

const containerVariants: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.06 },
  },
};

const headerItemVariants: Variants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease: [0.25, 0.1, 0.25, 1] },
  },
};

const fadeUpVariants: Variants = {
  hidden: { opacity: 0, y: 24 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.5, ease: [0.25, 0.1, 0.25, 1] },
  },
};

// ── Helpers ────────────────────────────────────────────────────────────────

function filterProjects(
  projects: ProjectDTO[],
  query: string,
  envFilter: EnvironmentFilter,
): ProjectDTO[] {
  if (!query && envFilter === 'all') return projects;

  return projects.filter((p) => {
    if (envFilter !== 'all' && p.spec.environment !== envFilter) return false;
    if (!query) return true;
    const q = query.toLowerCase();
    return (
      p.spec.displayName.toLowerCase().includes(q) ||
      p.metadata.id.toLowerCase().includes(q) ||
      (p.spec.description ?? '').toLowerCase().includes(q)
    );
  });
}

function sortProjects(projects: ProjectDTO[], sortKey: SortKey): ProjectDTO[] {
  const sorted = [...projects];
  switch (sortKey) {
    case 'name-asc':
      sorted.sort((a, b) =>
        a.spec.displayName.localeCompare(b.spec.displayName),
      );
      break;
    case 'name-desc':
      sorted.sort((a, b) =>
        b.spec.displayName.localeCompare(a.spec.displayName),
      );
      break;
    case 'created':
      sorted.sort((a, b) => {
        const da = a.metadata.createdAt ?? '';
        const db = b.metadata.createdAt ?? '';
        return db.localeCompare(da); // newest first
      });
      break;
    case 'updated':
      sorted.sort((a, b) => {
        const da = a.metadata.updatedAt ?? a.metadata.createdAt ?? '';
        const db = b.metadata.updatedAt ?? b.metadata.createdAt ?? '';
        return db.localeCompare(da);
      });
      break;
  }
  return sorted;
}

function projectCountLabel(count: number): string {
  return `${count} project${count !== 1 ? 's' : ''}`;
}

// ═══════════════════════════════════════════════════════════════════════════
// SUB-COMPONENTS
// ═══════════════════════════════════════════════════════════════════════════

// ── Loading Skeleton ──────────────────────────────────────────────────────

function ProjectCardSkeleton({ view }: { view: ViewMode }) {
  if (view === 'grid') {
    return (
      <div className="rounded-lg border border-border/50 bg-card/50">
        <div className="flex items-start justify-between gap-3 p-6 pb-3">
          <div className="flex items-center gap-2.5">
            <Skeleton className="h-9 w-9 rounded-lg" />
            <div className="space-y-1.5">
              <Skeleton className="h-4 w-28" />
              <Skeleton className="h-3 w-20" />
            </div>
          </div>
          <Skeleton className="h-5 w-16 rounded-full" />
        </div>
        <div className="px-6 pb-3">
          <Skeleton className="mb-2 h-3 w-full" />
          <Skeleton className="h-3 w-3/4" />
          <div className="mt-2">
            <Skeleton className="h-5 w-16 rounded-md" />
          </div>
        </div>
        <div className="mx-6 h-px bg-border/50" />
        <div className="flex items-center justify-between p-6 pt-3">
          <Skeleton className="h-3 w-20" />
          <Skeleton className="h-3 w-16" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-4 rounded-lg border border-border/50 bg-card/50 px-4 py-3">
      <Skeleton className="h-8 w-8 shrink-0 rounded-md" />
      <div className="flex-1 space-y-1.5">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-3 w-20" />
      </div>
      <Skeleton className="hidden h-5 w-14 rounded-md sm:block" />
      <Skeleton className="h-5 w-16 rounded-full" />
      <Skeleton className="hidden h-3 w-16 md:block" />
      <Skeleton className="hidden h-3 w-14 lg:block" />
    </div>
  );
}

function LoadingState({ view }: { view: ViewMode }) {
  return (
    <div
      className={
        view === 'grid'
          ? 'grid gap-4 sm:grid-cols-2 xl:grid-cols-3'
          : 'flex flex-col gap-2'
      }
    >
      {Array.from({ length: 6 }).map((_, i) => (
        <ProjectCardSkeleton key={i} view={view} />
      ))}
    </div>
  );
}

// ── Empty State ───────────────────────────────────────────────────────────

function EmptyState({ onCreate }: { onCreate: () => void }) {
  return (
    <motion.div
      variants={fadeUpVariants}
      initial="hidden"
      animate="visible"
      className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border/60 bg-card/20 px-6 py-20 text-center"
    >
      <div className="mb-5 flex h-14 w-14 items-center justify-center rounded-2xl border border-border/40 bg-muted/30">
        <Layers className="h-7 w-7 text-muted-foreground/50" />
      </div>
      <h3 className="mb-1.5 text-lg font-semibold tracking-tight text-foreground/90">
        No projects yet
      </h3>
      <p className="mb-6 max-w-sm text-sm leading-relaxed text-muted-foreground">
        Create your first project to get started with CloudOS.
      </p>
      <Button onClick={onCreate} size="sm" className="gap-1.5">
        <Plus className="h-4 w-4" />
        Create Project
      </Button>
    </motion.div>
  );
}

// ── No Search/Filter Results ───────────────────────────────────────────────

function NoResultsState() {
  return (
    <motion.div
      variants={fadeUpVariants}
      initial="hidden"
      animate="visible"
      className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border/40 bg-card/10 px-6 py-16 text-center"
    >
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl border border-border/30 bg-muted/20">
        <Search className="h-6 w-6 text-muted-foreground/40" />
      </div>
      <h3 className="mb-1 text-base font-medium text-foreground/80">
        No matching projects
      </h3>
      <p className="max-w-xs text-sm text-muted-foreground">
        Try adjusting your search or filter to find what you&apos;re looking
        for.
      </p>
    </motion.div>
  );
}

// ── Error State ───────────────────────────────────────────────────────────

function ErrorState({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <motion.div variants={fadeUpVariants} initial="hidden" animate="visible">
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Failed to load projects</AlertTitle>
        <AlertDescription className="flex flex-col gap-3">
          <p className="text-sm">{message}</p>
          <Button
            variant="outline"
            size="sm"
            className="w-fit gap-1.5"
            onClick={onRetry}
          >
            <RotateCcw className="h-3.5 w-3.5" />
            Retry
          </Button>
        </AlertDescription>
      </Alert>
    </motion.div>
  );
}

// ── Stats Bar ─────────────────────────────────────────────────────────────

function StatsBar({
  total,
  active,
  archived,
}: {
  total: number;
  active: number;
  archived: number;
}) {
  if (total === 0) return null;

  return (
    <motion.div
      variants={headerItemVariants}
      className="flex items-center gap-4 text-xs text-muted-foreground"
    >
      <span className="inline-flex items-center gap-1.5">
        <span className="h-1.5 w-1.5 rounded-full bg-muted-foreground/30" />
        {projectCountLabel(total)} total
      </span>

      <span className="h-3 w-px bg-border/50" aria-hidden="true" />

      <span className="inline-flex items-center gap-1.5">
        <span className="h-1.5 w-1.5 rounded-full bg-emerald-500/70" />
        {active} Active
      </span>

      <span className="h-3 w-px bg-border/50" aria-hidden="true" />

      <span className="inline-flex items-center gap-1.5">
        <span className="h-1.5 w-1.5 rounded-full bg-muted-foreground/20" />
        {archived} Archived
      </span>
    </motion.div>
  );
}

// ── Environment Filter Buttons ─────────────────────────────────────────────

function EnvironmentFilters({
  value,
  onChange,
}: {
  value: EnvironmentFilter;
  onChange: (v: EnvironmentFilter) => void;
}) {
  return (
    <div className="hidden items-center gap-1 sm:flex">
      {ENVIRONMENTS.map((env) => {
        const isActive = value === env.value;
        return (
          <button
            key={env.value}
            type="button"
            onClick={() => onChange(env.value)}
            className={cn(
              'relative inline-flex items-center rounded-md px-2.5 py-1.5 text-xs font-medium transition-colors',
              isActive
                ? 'bg-accent text-accent-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground hover:bg-accent/50',
            )}
            aria-pressed={isActive}
          >
            {env.label}
          </button>
        );
      })}
    </div>
  );
}

// ── View Toggle ────────────────────────────────────────────────────────────

function ViewToggle({
  value,
  onChange,
}: {
  value: ViewMode;
  onChange: (v: ViewMode) => void;
}) {
  return (
    <div className="flex items-center overflow-hidden rounded-lg border border-border/50 bg-card/30 p-0.5">
      <button
        type="button"
        onClick={() => onChange('grid')}
        className={cn(
          'flex h-8 w-8 items-center justify-center rounded-md transition-colors',
          value === 'grid'
            ? 'bg-accent text-accent-foreground shadow-sm'
            : 'text-muted-foreground hover:text-foreground',
        )}
        aria-label="Grid view"
        aria-pressed={value === 'grid'}
      >
        <Grid3X3 className="h-4 w-4" />
      </button>
      <button
        type="button"
        onClick={() => onChange('list')}
        className={cn(
          'flex h-8 w-8 items-center justify-center rounded-md transition-colors',
          value === 'list'
            ? 'bg-accent text-accent-foreground shadow-sm'
            : 'text-muted-foreground hover:text-foreground',
        )}
        aria-label="List view"
        aria-pressed={value === 'list'}
      >
        <List className="h-4 w-4" />
      </button>
    </div>
  );
}

// ═══════════════════════════════════════════════════════════════════════════
// MAIN PAGE
// ═══════════════════════════════════════════════════════════════════════════

export default function ProjectsPage() {
  // ── Data ──
  const { data, isLoading, error, refetch } = useProjects();
  const projects = data?.items ?? [];

  // ── Local state ──
  const [searchQuery, setSearchQuery] = useState('');
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [sortBy, setSortBy] = useState<SortKey>('created');
  const [environmentFilter, setEnvironmentFilter] =
    useState<EnvironmentFilter>('all');
  const [dialogOpen, setDialogOpen] = useState(false);

  // ── Derived ──
  const filteredAndSorted = useMemo(() => {
    if (isLoading || error) return [];
    const filtered = filterProjects(projects, searchQuery, environmentFilter);
    return sortProjects(filtered, sortBy);
  }, [projects, searchQuery, environmentFilter, sortBy, isLoading, error]);

  const activeProjectCount = useMemo(
    () => projects.filter((p) => p.status?.phase === 'Active').length,
    [projects],
  );

  const archivedProjectCount = useMemo(
    () => projects.filter((p) => p.status?.phase === 'Archived').length,
    [projects],
  );

  const isFiltered = environmentFilter !== 'all' || searchQuery.length > 0;

  const handleCreateClick = useCallback(() => {
    setDialogOpen(true);
  }, []);

  const environmentFilterLabel =
    environmentFilter === 'all'
      ? 'All'
      : environmentFilter.charAt(0).toUpperCase() + environmentFilter.slice(1);

  const hasProjects = projects.length > 0;
  const showResults = filteredAndSorted.length > 0;

  // ── Render ──
  return (
    <>
      <motion.div
        className="mx-auto max-w-6xl space-y-6 pb-12 pt-2"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        {/* ═══════════════════════════════════════════════════════════════
            PAGE HEADER
           ═══════════════════════════════════════════════════════════════ */}
        <motion.div
          variants={headerItemVariants}
          className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
        >
          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
                Projects
              </h1>
              {!isLoading && !error && hasProjects && (
                <Badge
                  variant="secondary"
                  className="rounded-full px-2.5 py-0.5 text-[11px] font-medium"
                >
                  {projects.length}
                </Badge>
              )}
            </div>
            <p className="text-sm text-muted-foreground sm:text-base">
              Manage your infrastructure projects
            </p>
          </div>

          {/* Header actions */}
          <div className="flex items-center gap-2">
            <ViewToggle value={viewMode} onChange={setViewMode} />

            <Button
              onClick={handleCreateClick}
              size="sm"
              className="gap-1.5"
            >
              <Plus className="h-4 w-4" />
              <span className="hidden sm:inline">Create Project</span>
            </Button>
          </div>
        </motion.div>

        {/* ═══════════════════════════════════════════════════════════════
            STATS BAR
           ═══════════════════════════════════════════════════════════════ */}
        {!isLoading && !error && (
          <StatsBar
            total={projects.length}
            active={activeProjectCount}
            archived={archivedProjectCount}
          />
        )}

        {/* ═══════════════════════════════════════════════════════════════
            SEARCH & FILTER BAR
           ═══════════════════════════════════════════════════════════════ */}
        <motion.div
          variants={headerItemVariants}
          className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
        >
          {/* Search */}
          <div className="relative flex-1 max-w-md">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/60" />
            <Input
              placeholder="Search projects..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="h-9 border-border/50 bg-card/30 pl-9 text-sm placeholder:text-muted-foreground/50 focus-visible:border-border/80"
              aria-label="Search projects"
            />
          </div>

          {/* Sort + Filter */}
          <div className="flex items-center gap-2">
            {/* Sort dropdown */}
            <DropdownMenu>
              <DropdownMenuTrigger
                className={cn(
                  'inline-flex items-center justify-center gap-2',
                  'rounded-md border border-input bg-background px-3 py-1.5',
                  'text-xs font-medium text-muted-foreground',
                  'hover:bg-accent hover:text-accent-foreground',
                  'transition-colors',
                  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
                )}
                aria-label="Sort projects"
              >
                <SortAsc className="h-3.5 w-3.5" />
                <span className="hidden sm:inline">
                  {SORT_LABEL_MAP[sortBy]}
                </span>
                <span className="sm:hidden">Sort</span>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
                <DropdownMenuLabel>Sort by</DropdownMenuLabel>
                <DropdownMenuSeparator />
                {SORT_OPTIONS.map((option) => (
                  <DropdownMenuItem
                    key={option.value}
                    onSelect={() => setSortBy(option.value)}
                    className={cn(
                      sortBy === option.value && 'font-medium text-foreground',
                    )}
                  >
                    <span className="flex-1">{option.label}</span>
                    {sortBy === option.value && (
                      <Check className="h-3.5 w-3.5 text-primary" />
                    )}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>

            {/* Separator */}
            <div
              className="hidden h-6 w-px bg-border/50 sm:block"
              aria-hidden="true"
            />

            {/* Environment filters (desktop) */}
            <EnvironmentFilters
              value={environmentFilter}
              onChange={setEnvironmentFilter}
            />

            {/* Environment filter (mobile) */}
            <div className="sm:hidden">
              <DropdownMenu>
                <DropdownMenuTrigger
                  className={cn(
                    'inline-flex items-center justify-center gap-2',
                    'rounded-md border border-input bg-background px-3 py-1.5',
                    'text-xs font-medium text-muted-foreground',
                    'hover:bg-accent hover:text-accent-foreground',
                    'transition-colors',
                    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
                  )}
                  aria-label="Filter by environment"
                >
                  <Layers className="h-3.5 w-3.5" />
                  {environmentFilterLabel}
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-44">
                  <DropdownMenuLabel>Environment</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {ENVIRONMENTS.map((env) => (
                    <DropdownMenuItem
                      key={env.value}
                      onSelect={() => setEnvironmentFilter(env.value)}
                      className={cn(
                        environmentFilter === env.value &&
                          'font-medium text-foreground',
                      )}
                    >
                      <span className="flex-1">{env.label}</span>
                      {environmentFilter === env.value && (
                        <Check className="h-3.5 w-3.5 text-primary" />
                      )}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {/* Active filter badge */}
            {isFiltered && (
              <Badge
                variant="secondary"
                className="gap-1 rounded-full px-2 py-0 text-[11px] font-normal"
              >
                <span className="text-muted-foreground/70">Filtered</span>
              </Badge>
            )}
          </div>
        </motion.div>

        {/* ═══════════════════════════════════════════════════════════════
            CONTENT AREA
           ═══════════════════════════════════════════════════════════════ */}
        <Separator className="bg-border/40" />

        {/* ── Error ── */}
        {error && !isLoading && (
          <ErrorState
            message={
              (error as Error)?.message ||
              'An unexpected error occurred while loading projects.'
            }
            onRetry={() => refetch()}
          />
        )}

        {/* ── Loading ── */}
        {isLoading && <LoadingState view={viewMode} />}

        {/* ── Empty (no projects at all) ── */}
        {!isLoading && !error && !hasProjects && (
          <EmptyState onCreate={handleCreateClick} />
        )}

        {/* ── No results (filtered everything out) ── */}
        {!isLoading && !error && hasProjects && !showResults && (
          <NoResultsState />
        )}

        {/* ── Project grid / list ── */}
        {!isLoading && !error && showResults && (
          <div
            className={
              viewMode === 'grid'
                ? 'grid gap-4 sm:grid-cols-2 xl:grid-cols-3'
                : 'flex flex-col gap-2'
            }
          >
            <AnimatePresence mode="popLayout">
              {filteredAndSorted.map((project) => (
                <ProjectCard
                  key={project.metadata.id}
                  project={project}
                  view={viewMode}
                />
              ))}
            </AnimatePresence>
          </div>
        )}

        {/* ── Result count (when filters active) ── */}
        {!isLoading && !error && showResults && isFiltered && (
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center text-xs text-muted-foreground/60"
          >
            Showing {filteredAndSorted.length} of {projectCountLabel(projects.length)}
          </motion.p>
        )}
      </motion.div>

      {/* ═══════════════════════════════════════════════════════════════
          CREATE PROJECT DIALOG
         ═══════════════════════════════════════════════════════════════ */}
      <CreateProjectDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
      />
    </>
  );
}
