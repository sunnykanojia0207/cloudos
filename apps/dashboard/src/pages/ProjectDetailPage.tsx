import { useState, useMemo, useCallback } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { motion, type Variants } from 'framer-motion';
import {
  ArrowLeft,
  Edit3,
  Trash2,
  Box,
  Cpu,
  Activity,
  Settings,
  Globe,
  Layers,
  Clock,
  Tag,
  Server,
  Database,
  Lock,
  Brain,
  AlertTriangle,
  AlertCircle,
  RotateCcw,
  Circle,
  CalendarDays,
  FileText,
  List,
  Zap,
  Rocket,
  Heart,
  ExternalLink,
  Terminal,
} from 'lucide-react';
import { usePageTitle } from '@/hooks/usePageTitle';
import {
  useProject,
  useControllers,
  useResourceKinds,
  useDeleteProject,
} from '@/hooks/useCloudOS';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
  DialogClose,
} from '@/components/ui/dialog';
import { cn, relativeTime, formatNumber } from '@/lib/utils';
import { DeployDialog } from '@/components/applications/DeployDialog';
import { ApplicationCard } from '@/components/applications/ApplicationCard';
import { useApplications } from '@/hooks/useApplications';
import type {
  ProjectPhase,
  ProjectHealth,
  ResourceKindDto,
  ControllerDTO,
} from '@cloudos/sdk';

// ── Animation Variants ────────────────────────────────────────────────────

const containerVariants: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.06 },
  },
};

const fadeUpVariants: Variants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease: [0.25, 0.1, 0.25, 1] },
  },
};

const staggerCardVariants: Variants = {
  hidden: { opacity: 0, y: 16 },
  visible: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { duration: 0.35, delay: i * 0.05, ease: [0.25, 0.1, 0.25, 1] },
  }),
};

const tabContentVariants: Variants = {
  hidden: { opacity: 0, y: 12 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.3, ease: [0.25, 0.1, 0.25, 1] },
  },
};

// ── Constants ─────────────────────────────────────────────────────────────

const ACTIVE_TABS = [
  { value: 'overview', label: 'Overview', icon: Layers },
  { value: 'resources', label: 'Resources', icon: Server },
  { value: 'controllers', label: 'Controllers', icon: Cpu },
  { value: 'activity', label: 'Activity', icon: Activity },
  { value: 'settings', label: 'Settings', icon: Settings },
] as const;

const FUTURE_TABS = [
  { value: 'deployments', label: 'Deployments', icon: Zap },
  { value: 'storage', label: 'Storage', icon: Database },
  { value: 'database', label: 'Database', icon: Database },
  { value: 'domains', label: 'Domains', icon: Globe },
  { value: 'secrets', label: 'Secrets', icon: Lock },
  { value: 'ai', label: 'AI', icon: Brain },
] as const;

// ── Helpers ───────────────────────────────────────────────────────────────

function getPhaseBadgeVariant(phase: ProjectPhase | undefined) {
  switch (phase) {
    case 'Active':
      return 'success' as const;
    case 'Creating':
      return 'warning' as const;
    case 'Archived':
      return 'secondary' as const;
    case 'Deleting':
      return 'destructive' as const;
    default:
      return 'outline' as const;
  }
}

function getHealthColor(health: ProjectHealth | undefined): string {
  switch (health) {
    case 'Healthy':
      return 'bg-emerald-500';
    case 'Degraded':
      return 'bg-amber-500';
    case 'Unhealthy':
      return 'bg-red-500';
    default:
      return 'bg-muted-foreground/40';
  }
}

function getPhaseColor(phase: ProjectPhase | undefined): string {
  switch (phase) {
    case 'Active':
      return 'bg-emerald-500';
    case 'Creating':
      return 'bg-amber-500';
    case 'Archived':
      return 'bg-muted-foreground/40';
    case 'Deleting':
      return 'bg-red-500';
    default:
      return 'bg-muted-foreground/40';
  }
}

function formatDate(dateStr: string | undefined | null): string {
  if (!dateStr) return '\u2014';
  try {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return '\u2014';
  }
}

function getEnvironmentBadgeVariant(
  env: string | undefined,
): 'default' | 'secondary' | 'outline' | 'success' | 'warning' {
  switch (env) {
    case 'production':
      return 'success' as const;
    case 'staging':
      return 'warning' as const;
    case 'development':
      return 'secondary' as const;
    case 'testing':
      return 'outline' as const;
    default:
      return 'secondary' as const;
  }
}

function getControllerStateColor(state: string): string {
  switch (state) {
    case 'running':
    case 'active':
    case 'healthy':
      return 'bg-emerald-500';
    case 'degraded':
    case 'warning':
      return 'bg-amber-500';
    case 'stopped':
    case 'error':
    case 'unhealthy':
      return 'bg-red-500';
    default:
      return 'bg-muted-foreground/40';
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// SUB-COMPONENTS
// ═══════════════════════════════════════════════════════════════════════════

// ── Loading Skeleton ──────────────────────────────────────────────────────

function PageSkeleton() {
  return (
    <div className="space-y-8">
      {/* Header skeleton */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex items-start gap-4">
          <Skeleton className="h-9 w-9 rounded-md" />
          <div className="space-y-2">
            <Skeleton className="h-8 w-56" />
            <Skeleton className="h-4 w-36" />
            <div className="flex items-center gap-2 pt-1">
              <Skeleton className="h-5 w-16 rounded-full" />
              <Skeleton className="h-5 w-24 rounded-full" />
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Skeleton className="h-9 w-9 rounded-md" />
          <Skeleton className="h-9 w-9 rounded-md" />
        </div>
      </div>

      {/* Tabs skeleton */}
      <Skeleton className="h-10 w-full rounded-md" />

      {/* Stats grid skeleton */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Card key={i}>
            <CardHeader className="pb-2">
              <Skeleton className="h-3.5 w-16" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-7 w-20" />
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Details card skeleton */}
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-24" />
        </CardHeader>
        <CardContent className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-start gap-2">
              <Skeleton className="mt-0.5 h-4 w-4 rounded" />
              <div className="flex-1 space-y-1">
                <Skeleton className="h-3 w-20" />
                <Skeleton className="h-4 w-40" />
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}

// ── Error State ───────────────────────────────────────────────────────────

function ErrorState({
  message,
  onRetry,
  projectId,
}: {
  message: string;
  onRetry: () => void;
  projectId: string;
}) {
  return (
    <motion.div
      variants={fadeUpVariants}
      initial="hidden"
      animate="visible"
      className="space-y-4"
    >
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Failed to load project</AlertTitle>
        <AlertDescription className="flex flex-col gap-3">
          <p className="text-sm">{message}</p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              className="w-fit gap-1.5"
              onClick={onRetry}
            >
              <RotateCcw className="h-3.5 w-3.5" />
              Retry
            </Button>
            <Button variant="ghost" size="sm" className="gap-1.5" asChild>
              <Link to="/projects">
                <ArrowLeft className="h-3.5 w-3.5" />
                Back to Projects
              </Link>
            </Button>
          </div>
        </AlertDescription>
      </Alert>

      <Card className="border-dashed border-border/40 bg-card/20">
        <CardContent className="flex flex-col items-center py-12 text-center">
          <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl border border-border/30 bg-muted/20">
            <AlertTriangle className="h-6 w-6 text-muted-foreground/40" />
          </div>
          <p className="text-sm text-muted-foreground">
            Project <span className="font-mono text-xs text-foreground/60">{projectId}</span> could not be retrieved.
          </p>
        </CardContent>
      </Card>
    </motion.div>
  );
}

// ── Status Badge ──────────────────────────────────────────────────────────

function StatusBadge({ phase }: { phase: ProjectPhase | undefined }) {
  const variant = getPhaseBadgeVariant(phase);
  const isPulsing = phase === 'Creating' || phase === 'Deleting';

  return (
    <Badge
      variant={variant}
      className={cn(
        'rounded-full px-2.5 py-0.5 text-[11px] font-medium capitalize',
        isPulsing && 'animate-pulse',
      )}
    >
      {phase ?? 'Unknown'}
    </Badge>
  );
}

// ── Stat Card ─────────────────────────────────────────────────────────────

interface StatCardProps {
  label: string;
  value: string | number;
  icon: React.ElementType;
  color?: string;
  dotColor?: string;
}

function StatCard({ label, value, icon: Icon, color, dotColor }: StatCardProps) {
  return (
    <Card className="border-border/50 bg-card/50 transition-colors hover:border-border/80">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-xs font-medium text-muted-foreground">
          {label}
        </CardTitle>
        <div
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded-md',
            color ?? 'bg-muted/30',
          )}
        >
          <Icon
            className={cn(
              'h-3.5 w-3.5',
              color ? 'text-current' : 'text-muted-foreground/60',
            )}
          />
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-2">
          {dotColor && (
            <span
              className={cn('h-2 w-2 rounded-full', dotColor)}
              aria-hidden="true"
            />
          )}
          <span className="text-xl font-semibold tracking-tight">
            {value}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}

// ── Empty Tab State ──────────────────────────────────────────────────────

interface EmptyTabStateProps {
  icon: React.ElementType;
  title: string;
  description: string;
}

function EmptyTabState({ icon: Icon, title, description }: EmptyTabStateProps) {
  return (
    <motion.div
      variants={fadeUpVariants}
      initial="hidden"
      animate="visible"
      className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border/40 bg-card/10 px-6 py-16 text-center"
    >
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl border border-border/30 bg-muted/20">
        <Icon className="h-6 w-6 text-muted-foreground/40" />
      </div>
      <h3 className="mb-1 text-base font-medium text-foreground/80">
        {title}
      </h3>
      <p className="max-w-xs text-sm text-muted-foreground">{description}</p>
    </motion.div>
  );
}

// ── Details Row ───────────────────────────────────────────────────────────

interface DetailsRowProps {
  icon: React.ElementType;
  label: string;
  children: React.ReactNode;
}

function DetailsRow({ icon: Icon, label, children }: DetailsRowProps) {
  return (
    <div className="flex items-start gap-3">
      <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-muted/30">
        <Icon className="h-4 w-4 text-muted-foreground/60" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-xs font-medium text-muted-foreground">{label}</p>
        <div className="mt-0.5 text-sm text-foreground/90">{children}</div>
      </div>
    </div>
  );
}

// ── Delete Confirmation Dialog ────────────────────────────────────────────

interface DeleteProjectDialogProps {
  projectId: string;
  projectName: string;
  onConfirm: () => void;
  isDeleting: boolean;
  error?: string;
}

function DeleteProjectDialog({
  projectId,
  projectName,
  onConfirm,
  isDeleting,
  error,
}: DeleteProjectDialogProps) {
  return (
    <Dialog>
      <DialogTrigger>
        <Button
          variant="ghost"
          size="icon"
          className="text-muted-foreground/60 hover:text-red-400 hover:bg-red-500/10"
          aria-label="Delete project"
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Delete project</DialogTitle>
          <DialogDescription className="pt-1">
            Are you sure you want to delete{' '}
            <span className="font-medium text-foreground">{projectName}</span>?
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-start gap-3 rounded-lg border border-red-500/20 bg-red-500/5 p-3">
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-red-400" />
          <div className="text-xs leading-relaxed text-muted-foreground">
            <p className="mb-1 font-medium text-red-300">Warning</p>
            <p>
              This will permanently delete this project and all its associated
              resources, deployments, and configurations. This action cannot be
              undone.
            </p>
          </div>
        </div>

        <div className="rounded-lg bg-muted/30 px-3 py-2">
          <p className="text-xs text-muted-foreground">Project ID</p>
          <p className="font-mono text-xs text-foreground/70">{projectId}</p>
        </div>

        {error && (
          <div className="flex items-start gap-2 rounded-md border border-danger/30 bg-danger/5 px-3 py-2 text-small text-danger" role="alert">
            <AlertCircle className="h-4 w-4 mt-0.5 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <DialogFooter>
          <DialogClose>
            <Button variant="outline" size="sm" disabled={isDeleting}>
              Cancel
            </Button>
          </DialogClose>
          <Button
            variant="destructive"
            size="sm"
            className="gap-1.5"
            disabled={isDeleting}
            onClick={onConfirm}
          >
            {isDeleting ? (
              <>
                <RotateCcw className="h-3.5 w-3.5 animate-spin" />
                Deleting...
              </>
            ) : (
              <>
                <Trash2 className="h-3.5 w-3.5" />
                Delete Project
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ── Resource Kind Card ────────────────────────────────────────────────────

interface ResourceKindCardProps {
  kind: ResourceKindDto;
  index: number;
}

function ResourceKindCard({ kind, index }: ResourceKindCardProps) {
  return (
    <motion.div
      variants={staggerCardVariants}
      initial="hidden"
      animate="visible"
      custom={index}
    >
      <Link
        to={`/resources/${kind.name}`}
        className="group block rounded-lg border border-border/50 bg-card/50 p-4 transition-all hover:border-border/80 hover:bg-card/80"
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted/40 group-hover:bg-muted/60 transition-colors">
              <Box className="h-4 w-4 text-muted-foreground/60 group-hover:text-foreground/80 transition-colors" />
            </div>
            <div>
              <p className="text-sm font-medium text-foreground/90 group-hover:text-foreground transition-colors">
                {kind.name}
              </p>
              <p className="text-xs text-muted-foreground">
                {kind.namespaced ? 'Namespaced' : 'Cluster-scoped'}
                {kind.versions?.length ? ` \u00B7 ${kind.versions.length} version(s)` : ''}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {kind.versions?.slice(0, 2).map((v) => (
              <Badge key={v} variant="outline" className="rounded-md px-1.5 py-0 text-[10px] font-mono">
                {v}
              </Badge>
            ))}
            <List className="h-4 w-4 text-muted-foreground/30 group-hover:text-muted-foreground/60 transition-colors" />
          </div>
        </div>
      </Link>
    </motion.div>
  );
}

// ── Controller Row ────────────────────────────────────────────────────────

interface ControllerRowProps {
  controller: ControllerDTO;
  index: number;
}

function ControllerRow({ controller, index }: ControllerRowProps) {
  const health = controller.health;
  const stateColor = getControllerStateColor(health?.state ?? controller.state);

  return (
    <motion.div
      variants={staggerCardVariants}
      initial="hidden"
      animate="visible"
      custom={index}
    >
      <Link
        to={`/controllers/${controller.name}`}
        className="group block rounded-lg border border-border/50 bg-card/50 p-4 transition-all hover:border-border/80 hover:bg-card/80"
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted/40 group-hover:bg-muted/60 transition-colors">
              <Cpu className="h-4 w-4 text-muted-foreground/60 group-hover:text-foreground/80 transition-colors" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <p className="text-sm font-medium text-foreground/90 group-hover:text-foreground transition-colors">
                  {controller.name}
                </p>
                <Badge
                  variant="outline"
                  className="rounded-md px-1.5 py-0 text-[10px] font-normal text-muted-foreground"
                >
                  {controller.kind}
                </Badge>
              </div>
              <div className="mt-0.5 flex items-center gap-3 text-xs text-muted-foreground">
                <span className="inline-flex items-center gap-1">
                  <span className={cn('h-1.5 w-1.5 rounded-full', stateColor)} />
                  {health?.state ?? controller.state}
                </span>
                {health?.reconcileCount !== undefined && (
                  <>
                    <span className="h-3 w-px bg-border/40" aria-hidden="true" />
                    <span>{formatNumber(health.reconcileCount)} reconcile{health.reconcileCount !== 1 ? 's' : ''}</span>
                  </>
                )}
                {health?.errorCount !== undefined && health.errorCount > 0 && (
                  <>
                    <span className="h-3 w-px bg-border/40" aria-hidden="true" />
                    <span className="text-red-400">{health.errorCount} error{health.errorCount !== 1 ? 's' : ''}</span>
                  </>
                )}
              </div>
            </div>
          </div>
          {health?.lastReconciled && (
            <span className="hidden text-xs text-muted-foreground sm:block">
              {relativeTime(health.lastReconciled)}
            </span>
          )}
        </div>
        {controller.message && (
          <p className="mt-2 truncate text-xs text-muted-foreground/60">
            {controller.message}
          </p>
        )}
      </Link>
    </motion.div>
  );
}

// ═══════════════════════════════════════════════════════════════════════════
// MAIN PAGE
// ═══════════════════════════════════════════════════════════════════════════

export default function ProjectDetailPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const navigate = useNavigate();

  // ── Data ──
  const {
    data: project,
    isLoading,
    error,
    refetch,
  } = useProject(projectId ?? '');

  usePageTitle(project ? `${project.spec.displayName} • Projects` : 'Project Detail');
  const { data: controllersData } = useControllers();
  const { data: resourceKindsData } = useResourceKinds();
  const deleteProject = useDeleteProject();

  // ── Local state ──
  const [activeTab, setActiveTab] = useState('overview');
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState('');
  const [deployOpen, setDeployOpen] = useState(false);

  // ── All applications (filtered to this project on the overview) ──
  const { data: allApplications = [] } = useApplications();
  const applications = useMemo(
    () => allApplications.filter((app) => !app.spec.projectId || app.spec.projectId === projectId),
    [allApplications, projectId],
  );
  const appCount = applications.length;

  // ── Derived ──
  const allControllers = controllersData?.controllers ?? [];
  const allResourceKinds = resourceKindsData?.kinds ?? [];

  // Filter controllers that reference this project (via labels or naming conventions)
  const projectControllers = useMemo(() => {
    if (!projectId) return [];
    const id = projectId.toLowerCase();
    return allControllers.filter(
      (c) =>
        c.name.toLowerCase().includes(id) ||
        (c.health?.name && c.health.name.toLowerCase().includes(id)) ||
        c.kind.toLowerCase().includes(id),
    );
  }, [allControllers, projectId]);

  // ── Handlers ──
  const handleDelete = useCallback(async () => {
    if (!projectId) return;
    setIsDeleting(true);
    setDeleteError('');
    try {
      await deleteProject.mutateAsync(projectId);
      navigate('/projects', { replace: true });
    } catch (err) {
      setDeleteError((err as Error)?.message || 'Failed to delete project');
      setIsDeleting(false);
    }
  }, [projectId, deleteProject, navigate]);

  const handleEditClick = useCallback(() => {
    if (!projectId) return;
    navigate(`/projects/${projectId}/edit`);
  }, [projectId, navigate]);

  // ── Guard: no project ID ──
  if (!projectId) {
    return (
      <motion.div
        className="space-y-6"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Missing project ID</AlertTitle>
          <AlertDescription>
            No project ID was provided in the URL.
          </AlertDescription>
        </Alert>
        <Button variant="outline" size="sm" className="gap-1.5" asChild>
          <Link to="/projects">
            <ArrowLeft className="h-3.5 w-3.5" />
            Back to Projects
          </Link>
        </Button>
      </motion.div>
    );
  }

  // ── Loading ──
  if (isLoading) {
    return (
      <motion.div
        className="space-y-6"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        <PageSkeleton />
      </motion.div>
    );
  }

  // ── Error ──
  if (error || !project) {
    return (
      <motion.div
        className="space-y-6"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        <ErrorState
          message={
            (error as Error)?.message ||
            'An unexpected error occurred while loading the project.'
          }
          onRetry={() => refetch()}
          projectId={projectId}
        />
      </motion.div>
    );
  }

  const { spec, status, metadata } = project;
  const tags = spec.tags ? Object.entries(spec.tags) : [];

  // ── Render ──
  return (
    <motion.div
      className="mx-auto max-w-6xl space-y-6 pb-12 pt-2"
      variants={containerVariants}
      initial="hidden"
      animate="visible"
    >
      {/* ═════════════════════════════════════════════════════════════════
          PAGE HEADER
         ═════════════════════════════════════════════════════════════════ */}
      <motion.div
        variants={fadeUpVariants}
        className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
      >
        {/* Left: Back + Title + Meta */}
        <div className="flex items-start gap-4 min-w-0">
          <Button variant="ghost" size="icon" className="mt-0.5 shrink-0" asChild>
            <Link to="/projects" aria-label="Back to projects">
              <ArrowLeft className="h-4 w-4" />
            </Link>
          </Button>

          <div className="min-w-0">
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl truncate">
                {spec.displayName}
              </h1>
              <StatusBadge phase={status?.phase} />
              <Badge
                variant={getEnvironmentBadgeVariant(spec.environment)}
                className="rounded-full px-2.5 py-0.5 text-[11px] font-medium capitalize"
              >
                {spec.environment ?? 'unknown'}
              </Badge>
            </div>

            <div className="mt-1 flex items-center gap-3 text-sm text-muted-foreground">
              <span className="font-mono text-xs text-muted-foreground/60">
                {metadata.id}
              </span>
              {metadata.createdAt && (
                <>
                  <span className="h-3 w-px bg-border/40" aria-hidden="true" />
                  <span className="inline-flex items-center gap-1 text-xs">
                    <CalendarDays className="h-3 w-3" />
                    Created {formatDate(metadata.createdAt)}
                  </span>
                </>
              )}
            </div>

            {/* Mobile action buttons */}
            <div className="mt-3 flex items-center gap-2 sm:hidden">
              <DeployDialog open={deployOpen} onOpenChange={setDeployOpen} projectId={projectId}>
                <Button
                  variant="primary"
                  size="sm"
                  className="gap-1.5"
                >
                  <Rocket className="h-3.5 w-3.5" />
                  Deploy
                </Button>
              </DeployDialog>
              <Button
                variant="outline"
                size="sm"
                className="gap-1.5"
                onClick={handleEditClick}
              >
                <Edit3 className="h-3.5 w-3.5" />
                Edit
              </Button>
              <DeleteProjectDialog
                projectId={projectId}
                projectName={spec.displayName}
                onConfirm={handleDelete}
                isDeleting={isDeleting}
                error={deleteError}
              />
            </div>
          </div>
        </div>

        {/* Right: Desktop action buttons */}
        <div className="hidden items-center gap-2 sm:flex">
          <DeployDialog open={deployOpen} onOpenChange={setDeployOpen} projectId={projectId}>
            <Button
              variant="primary"
              size="sm"
              className="gap-1.5"
            >
              <Rocket className="h-3.5 w-3.5" />
              Deploy Application
            </Button>
          </DeployDialog>
          <Button
            variant="outline"
            size="sm"
            className="gap-1.5"
            onClick={handleEditClick}
          >
            <Edit3 className="h-3.5 w-3.5" />
            Edit
          </Button>
          <DeleteProjectDialog
            projectId={projectId}
            projectName={spec.displayName}
            onConfirm={handleDelete}
            isDeleting={isDeleting}
            error={deleteError}
          />
        </div>
      </motion.div>

      {/* ═════════════════════════════════════════════════════════════════
          TAB NAVIGATION
         ═════════════════════════════════════════════════════════════════ */}
      <motion.div variants={fadeUpVariants}>
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <div className="overflow-x-auto -mx-4 px-4">
            <TabsList className="mb-1 h-auto w-full justify-start overflow-x-auto rounded-none bg-transparent p-0">
              {ACTIVE_TABS.map((tab) => {
                const Icon = tab.icon;
                return (
                  <TabsTrigger
                    key={tab.value}
                    value={tab.value}
                    className={cn(
                      'relative inline-flex items-center gap-1.5 rounded-none border-b-2 px-4 py-2.5 text-xs font-medium transition-colors',
                      'data-[state=active]:border-primary data-[state=active]:text-foreground',
                      'data-[state=inactive]:border-transparent data-[state=inactive]:text-muted-foreground data-[state=inactive]:hover:text-foreground/70',
                    )}
                  >
                    <Icon className="h-3.5 w-3.5" />
                    {tab.label}
                  </TabsTrigger>
                );
              })}

              <Separator
                orientation="vertical"
                className="mx-2 h-6 bg-border/30"
              />

              {FUTURE_TABS.map((tab) => {
                const Icon = tab.icon;
                return (
                  <button
                    key={tab.value}
                    type="button"
                    disabled
                    className={cn(
                      'inline-flex items-center gap-1.5 rounded-none border-b-2 border-transparent px-4 py-2.5',
                      'text-xs font-medium text-muted-foreground/40 cursor-not-allowed',
                    )}
                    aria-disabled
                  >
                    <Icon className="h-3.5 w-3.5" />
                    {tab.label}
                  </button>
                );
              })}
            </TabsList>
          </div>

          <Separator className="bg-border/30" />

          {/* ─────────────────────────────────────────────────────────────
              OVERVIEW TAB
             ───────────────────────────────────────────────────────────── */}
          <TabsContent value="overview">
            <motion.div
              key="overview"
              variants={tabContentVariants}
              initial="hidden"
              animate="visible"
              className="space-y-6 pt-6"
            >
              {/* Stats Grid */}
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <motion.div
                  variants={staggerCardVariants}
                  initial="hidden"
                  animate="visible"
                  custom={0}
                >
                  <StatCard
                    label="Phase"
                    value={status?.phase ?? 'Unknown'}
                    icon={Layers}
                    color="bg-indigo-500/10 text-indigo-400"
                    dotColor={getPhaseColor(status?.phase)}
                  />
                </motion.div>

                <motion.div
                  variants={staggerCardVariants}
                  initial="hidden"
                  animate="visible"
                  custom={1}
                >
                  <StatCard
                    label="Health"
                    value={status?.health ?? 'Unknown'}
                    icon={Activity}
                    color="bg-emerald-500/10 text-emerald-400"
                    dotColor={getHealthColor(status?.health)}
                  />
                </motion.div>

                <motion.div
                  variants={staggerCardVariants}
                  initial="hidden"
                  animate="visible"
                  custom={2}
                >
                  <StatCard
                    label="Resources"
                    value={status?.resourceCount ?? 0}
                    icon={Server}
                    color="bg-blue-500/10 text-blue-400"
                  />
                </motion.div>

                <motion.div
                  variants={staggerCardVariants}
                  initial="hidden"
                  animate="visible"
                  custom={3}
                >
                  <StatCard
                    label="Deployments"
                    value={status?.deploymentCount ?? 0}
                    icon={Zap}
                    color="bg-amber-500/10 text-amber-400"
                  />
                </motion.div>
              </div>

              {/* Details Card */}
              <motion.div
                variants={staggerCardVariants}
                initial="hidden"
                animate="visible"
                custom={4}
              >
                <Card className="border-border/50 bg-card/50">
                  <CardHeader>
                    <CardTitle className="text-sm font-medium">
                      Project Details
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <DetailsRow icon={FileText} label="Description">
                      {spec.description ? (
                        <p className="leading-relaxed">{spec.description}</p>
                      ) : (
                        <span className="italic text-muted-foreground/60">
                          No description provided
                        </span>
                      )}
                    </DetailsRow>

                    <Separator className="bg-border/30" />

                    <DetailsRow icon={Globe} label="Environment">
                      <Badge
                        variant={getEnvironmentBadgeVariant(spec.environment)}
                        className="rounded-full px-2 py-0 text-[11px] font-medium capitalize"
                      >
                        {spec.environment ?? 'unknown'}
                      </Badge>
                    </DetailsRow>

                    <Separator className="bg-border/30" />

                    <DetailsRow icon={Globe} label="Default Region">
                      <span>{spec.defaultRegion ?? '\u2014'}</span>
                    </DetailsRow>

                    <Separator className="bg-border/30" />

                    <DetailsRow icon={Tag} label="Tags">
                      {tags.length > 0 ? (
                        <div className="flex flex-wrap gap-1.5">
                          {tags.map(([key, value]) => (
                            <Badge
                              key={key}
                              variant="secondary"
                              className="rounded-md px-2 py-0 text-[10px] font-normal"
                            >
                              {key}: {value as string}
                            </Badge>
                          ))}
                        </div>
                      ) : (
                        <span className="italic text-muted-foreground/60">
                          No tags
                        </span>
                      )}
                    </DetailsRow>

                    <Separator className="bg-border/30" />

                    <DetailsRow icon={CalendarDays} label="Created At">
                      <span>{formatDate(metadata.createdAt)}</span>
                    </DetailsRow>

                    <Separator className="bg-border/30" />

                    <DetailsRow icon={Clock} label="Last Activity">
                      <span>
                        {status?.lastActivity
                          ? relativeTime(status.lastActivity)
                          : '\u2014'}
                      </span>
                    </DetailsRow>
                  </CardContent>
                </Card>
              </motion.div>

              {/* ── Applications ───────────────────────────────────── */}
              <motion.div
                variants={staggerCardVariants}
                initial="hidden"
                animate="visible"
                custom={5}
              >
                <Card className="border-border/50 bg-card/50">
                  <CardHeader className="flex flex-row items-center justify-between">
                    <div>
                      <CardTitle className="text-sm font-medium">
                        Applications
                      </CardTitle>
                      <CardDescription>
                        {appCount} application{appCount !== 1 ? 's' : ''} deployed
                      </CardDescription>
                    </div>
                    <DeployDialog open={deployOpen} onOpenChange={setDeployOpen} projectId={projectId}>
                      <Button variant="primary" size="sm" className="gap-1.5">
                        <Rocket className="h-3.5 w-3.5" />
                        Deploy
                      </Button>
                    </DeployDialog>
                  </CardHeader>
                  <CardContent>
                    {applications.length > 0 ? (
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                        {applications.map((app) => (
                          <ApplicationCard key={app.metadata.id} app={app} />
                        ))}
                      </div>
                    ) : (
                      <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-border/40 bg-muted/10 px-6 py-10 text-center">
                        <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-lg border border-border/30 bg-muted/20">
                          <Box className="h-5 w-5 text-muted-foreground/50" />
                        </div>
                        <h4 className="mb-1 text-sm font-medium text-foreground/70">
                          No applications yet
                        </h4>
                        <p className="mb-4 max-w-xs text-xs text-muted-foreground">
                          Deploy your first application from a Git repository.
                        </p>
                        <DeployDialog open={deployOpen} onOpenChange={setDeployOpen} projectId={projectId}>
                          <Button variant="primary" size="sm" className="gap-1.5">
                            <Rocket className="h-3.5 w-3.5" />
                            Deploy Application
                          </Button>
                        </DeployDialog>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </motion.div>

              {/* Conditions (if any) */}
              {status?.conditions && status.conditions.length > 0 && (
                <motion.div
                  variants={staggerCardVariants}
                  initial="hidden"
                  animate="visible"
                  custom={6}
                >
                  <Card className="border-border/50 bg-card/50">
                    <CardHeader>
                      <CardTitle className="text-sm font-medium">
                        Conditions
                      </CardTitle>
                      <CardDescription>
                        Recent conditions affecting this project
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2">
                        {status.conditions.map((condition, i) => (
                          <div
                            key={i}
                            className="flex items-start gap-3 rounded-lg border border-border/30 bg-muted/20 px-3 py-2"
                          >
                            <div className="mt-0.5">
                              <Circle
                                className={cn(
                                  'h-2 w-2',
                                  condition.status === 'True'
                                    ? 'fill-emerald-500 text-emerald-500'
                                    : condition.status === 'False'
                                      ? 'fill-red-500 text-red-500'
                                      : 'fill-amber-500 text-amber-500',
                                )}
                              />
                            </div>
                            <div className="min-w-0 flex-1">
                              <p className="text-xs font-medium text-foreground/80">
                                {condition.type}
                              </p>
                              {condition.message && (
                                <p className="mt-0.5 text-xs text-muted-foreground/70">
                                  {condition.message}
                                </p>
                              )}
                              {condition.lastTransitionTime && (
                                <p className="mt-0.5 text-[10px] text-muted-foreground/50">
                                  {relativeTime(condition.lastTransitionTime)}
                                </p>
                              )}
                            </div>
                            <Badge
                              variant="outline"
                              className="shrink-0 rounded-md px-1.5 py-0 text-[10px] font-normal capitalize"
                            >
                              {condition.status}
                            </Badge>
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                </motion.div>
              )}
            </motion.div>
          </TabsContent>

          {/* ─────────────────────────────────────────────────────────────
              RESOURCES TAB
             ───────────────────────────────────────────────────────────── */}
          <TabsContent value="resources">
            <motion.div
              key="resources"
              variants={tabContentVariants}
              initial="hidden"
              animate="visible"
              className="space-y-6 pt-6"
            >
              {allResourceKinds.length > 0 ? (
                <>
                  <div className="flex items-center justify-between">
                    <p className="text-sm text-muted-foreground">
                      {allResourceKinds.length} resource kind{allResourceKinds.length !== 1 ? 's' : ''} available
                    </p>
                    <Button variant="outline" size="sm" className="gap-1.5" asChild>
                      <Link to="/resources">
                        <Server className="h-3.5 w-3.5" />
                        View All Resources
                      </Link>
                    </Button>
                  </div>
                  <div className="grid gap-3 sm:grid-cols-2">
                    {allResourceKinds.map((kind, i) => (
                      <ResourceKindCard key={kind.name} kind={kind} index={i} />
                    ))}
                  </div>
                </>
              ) : (
                <EmptyTabState
                  icon={Server}
                  title="No resources in this project"
                  description="Resources will appear here once they are deployed to this project."
                />
              )}
            </motion.div>
          </TabsContent>

          {/* ─────────────────────────────────────────────────────────────
              CONTROLLERS TAB
             ───────────────────────────────────────────────────────────── */}
          <TabsContent value="controllers">
            <motion.div
              key="controllers"
              variants={tabContentVariants}
              initial="hidden"
              animate="visible"
              className="space-y-6 pt-6"
            >
              {projectControllers.length > 0 ? (
                <>
                  <div className="flex items-center justify-between">
                    <p className="text-sm text-muted-foreground">
                      {projectControllers.length} controller{projectControllers.length !== 1 ? 's' : ''} associated with this project
                    </p>
                    <Button variant="outline" size="sm" className="gap-1.5" asChild>
                      <Link to="/controllers">
                        <Cpu className="h-3.5 w-3.5" />
                        All Controllers
                      </Link>
                    </Button>
                  </div>
                  <div className="space-y-2">
                    {projectControllers.map((controller, i) => (
                      <ControllerRow
                        key={controller.name}
                        controller={controller}
                        index={i}
                      />
                    ))}
                  </div>
                </>
              ) : (
                <EmptyTabState
                  icon={Cpu}
                  title="No controllers found"
                  description="Controllers that are relevant to this project will appear here."
                />
              )}
            </motion.div>
          </TabsContent>

          {/* ─────────────────────────────────────────────────────────────
              ACTIVITY TAB
             ───────────────────────────────────────────────────────────── */}
          <TabsContent value="activity">
            <motion.div
              key="activity"
              variants={tabContentVariants}
              initial="hidden"
              animate="visible"
              className="space-y-6 pt-6"
            >
              <motion.div
                variants={fadeUpVariants}
                initial="hidden"
                animate="visible"
                className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border/40 bg-card/10 px-6 py-20 text-center"
              >
                <div className="mb-5 flex h-14 w-14 items-center justify-center rounded-2xl border border-border/40 bg-muted/30">
                  <Activity className="h-7 w-7 text-muted-foreground/50" />
                </div>
                <h3 className="mb-1.5 text-lg font-semibold tracking-tight text-foreground/90">
                  Activity tracking coming soon
                </h3>
                <p className="mb-4 max-w-sm text-sm leading-relaxed text-muted-foreground">
                  We&apos;re working on a comprehensive activity feed that will
                  show all events, changes, and operations performed on this
                  project.
                </p>
                <div className="flex items-center gap-2 rounded-full border border-border/30 bg-muted/20 px-4 py-1.5">
                  <Brain className="h-3.5 w-3.5 text-muted-foreground/50" />
                  <span className="text-xs text-muted-foreground/60">
                    Timeline view &mdash; Coming in v2.1
                  </span>
                </div>
              </motion.div>
            </motion.div>
          </TabsContent>

          {/* ─────────────────────────────────────────────────────────────
              SETTINGS TAB
             ───────────────────────────────────────────────────────────── */}
          <TabsContent value="settings">
            <motion.div
              key="settings"
              variants={tabContentVariants}
              initial="hidden"
              animate="visible"
              className="space-y-6 pt-6"
            >
              {/* Read-only project spec */}
              <Card className="border-border/50 bg-card/50">
                <CardHeader>
                  <CardTitle className="text-sm font-medium">
                    Project Configuration
                  </CardTitle>
                  <CardDescription>
                    Current project specification values (read-only)
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* displayName */}
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-muted-foreground">
                      Display Name
                    </p>
                    <div className="rounded-md border border-border/40 bg-muted/20 px-3 py-2 font-mono text-sm text-foreground/80">
                      {spec.displayName}
                    </div>
                  </div>

                  {/* description */}
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-muted-foreground">
                      Description
                    </p>
                    <div className="rounded-md border border-border/40 bg-muted/20 px-3 py-2 font-mono text-sm text-foreground/80">
                      {spec.description || (
                        <span className="italic text-muted-foreground/50">
                          No description
                        </span>
                      )}
                    </div>
                  </div>

                  {/* environment */}
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-muted-foreground">
                      Environment
                    </p>
                    <div className="rounded-md border border-border/40 bg-muted/20 px-3 py-2 font-mono text-sm text-foreground/80 capitalize">
                      {spec.environment ?? 'unknown'}
                    </div>
                  </div>

                  {/* defaultRegion */}
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-muted-foreground">
                      Default Region
                    </p>
                    <div className="rounded-md border border-border/40 bg-muted/20 px-3 py-2 font-mono text-sm text-foreground/80">
                      {spec.defaultRegion || '\u2014'}
                    </div>
                  </div>

                  {/* tags */}
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-muted-foreground">
                      Tags
                    </p>
                    <div className="rounded-md border border-border/40 bg-muted/20 px-3 py-2 text-sm">
                      {tags.length > 0 ? (
                        <div className="flex flex-wrap gap-1.5">
                          {tags.map(([key, value]) => (
                            <Badge
                              key={key}
                              variant="secondary"
                              className="rounded-md px-2 py-0 text-[10px] font-mono font-normal"
                            >
                              {key}: {value as string}
                            </Badge>
                          ))}
                        </div>
                      ) : (
                        <span className="italic text-muted-foreground/50">
                          No tags configured
                        </span>
                      )}
                    </div>
                  </div>

                  {/* quota */}
                  {spec.quota && Object.keys(spec.quota).length > 0 && (
                    <div className="space-y-1">
                      <p className="text-xs font-medium text-muted-foreground">
                        Quota
                      </p>
                      <pre className="overflow-x-auto rounded-md border border-border/40 bg-muted/20 px-3 py-2 font-mono text-xs text-foreground/70">
                        {JSON.stringify(spec.quota, null, 2)}
                      </pre>
                    </div>
                  )}

                  {/* settings */}
                  {spec.settings && Object.keys(spec.settings).length > 0 && (
                    <div className="space-y-1">
                      <p className="text-xs font-medium text-muted-foreground">
                        Settings
                      </p>
                      <pre className="overflow-x-auto rounded-md border border-border/40 bg-muted/20 px-3 py-2 font-mono text-xs text-foreground/70">
                        {JSON.stringify(spec.settings, null, 2)}
                      </pre>
                    </div>
                  )}
                </CardContent>
              </Card>

              {/* Coming soon notice */}
              <motion.div
                variants={fadeUpVariants}
                className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border/40 bg-card/10 px-6 py-12 text-center"
              >
                <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl border border-border/30 bg-muted/20">
                  <Settings className="h-6 w-6 text-muted-foreground/40" />
                </div>
                <h3 className="mb-1 text-base font-medium text-foreground/80">
                  Settings management coming in a future release
                </h3>
                <p className="max-w-xs text-sm text-muted-foreground">
                  You will be able to modify project settings, quotas, and
                  configurations directly from this panel.
                </p>
              </motion.div>
            </motion.div>
          </TabsContent>
        </Tabs>
      </motion.div>
    </motion.div>
  );
}
