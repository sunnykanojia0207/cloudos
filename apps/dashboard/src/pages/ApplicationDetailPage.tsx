import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useApplication } from '@/hooks/useApplications';
import { useTimeline } from '@/hooks/useDeployments';
import { OverviewTab } from '@/components/applications/OverviewTab';
import { DeploymentsTab } from '@/components/applications/DeploymentsTab';
import { TimelineTab } from '@/components/applications/TimelineTab';
import { LogsTab } from '@/components/applications/LogsTab';
import { SettingsTab } from '@/components/applications/SettingsTab';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import {
  ArrowLeft,
  ExternalLink,
  AlertCircle,
  RotateCcw,
  Globe,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  BookOpen,
  Layers,
  Terminal,
  Settings,
  Activity,
} from 'lucide-react';
import { motion, type Variants } from 'framer-motion';
import { cn } from '@/lib/utils';

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

const tabContentVariants: Variants = {
  hidden: { opacity: 0, y: 12 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.3, ease: [0.25, 0.1, 0.25, 1] },
  },
};

// ── Tab configuration ─────────────────────────────────────────────────────

const APP_TABS = [
  { value: 'overview', label: 'Overview', icon: BookOpen },
  { value: 'deployments', label: 'Deployments', icon: Layers },
  { value: 'timeline', label: 'Timeline', icon: Activity },
  { value: 'logs', label: 'Logs', icon: Terminal },
  { value: 'settings', label: 'Settings', icon: Settings },
] as const;

// ── Helpers ───────────────────────────────────────────────────────────────

const phaseBadge: Record<
  string,
  { variant: 'success' | 'warning' | 'secondary' | 'destructive'; label: string }
> = {
  Running: { variant: 'success', label: 'Running' },
  Deploying: { variant: 'warning', label: 'Deploying' },
  Stopped: { variant: 'secondary', label: 'Stopped' },
  Failed: { variant: 'destructive', label: 'Failed' },
};

const healthBadge: Record<
  string,
  { variant: 'success' | 'warning' | 'destructive'; label: string }
> = {
  Healthy: { variant: 'success', label: 'Healthy' },
  Degraded: { variant: 'warning', label: 'Degraded' },
  Error: { variant: 'destructive', label: 'Error' },
};

function getHealthIcon(health: string) {
  switch (health) {
    case 'Healthy':
      return <CheckCircle2 className="h-3.5 w-3.5" />;
    case 'Degraded':
      return <AlertTriangle className="h-3.5 w-3.5" />;
    case 'Error':
      return <XCircle className="h-3.5 w-3.5" />;
    default:
      return <AlertCircle className="h-3.5 w-3.5" />;
  }
}

// ── Loading Skeleton ──────────────────────────────────────────────────────

function PageSkeleton() {
  return (
    <div className="space-y-6">
      {/* Breadcrumb skeleton */}
      <Skeleton className="h-4 w-40" />

      {/* Header skeleton */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex items-start gap-4">
          <Skeleton className="h-9 w-9 rounded-md shrink-0" />
          <div className="space-y-2">
            <Skeleton className="h-8 w-56" />
            <div className="flex items-center gap-2 pt-1">
              <Skeleton className="h-5 w-20 rounded-full" />
              <Skeleton className="h-5 w-24 rounded-full" />
            </div>
          </div>
        </div>
        <Skeleton className="h-9 w-24 rounded-md" />
      </div>

      {/* Tabs skeleton */}
      <Skeleton className="h-10 w-full rounded-md" />

      {/* Content skeleton */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="border-border/50">
          <CardHeader className="pb-3">
            <Skeleton className="h-5 w-40" />
          </CardHeader>
          <CardContent className="space-y-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="flex items-start gap-3">
                <Skeleton className="mt-0.5 h-4 w-4 rounded" />
                <div className="flex-1 space-y-1">
                  <Skeleton className="h-3 w-20" />
                  <Skeleton className="h-4 w-40" />
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
        <div className="space-y-6">
          <Card className="border-border/50">
            <CardHeader className="pb-3">
              <Skeleton className="h-5 w-32" />
            </CardHeader>
            <CardContent className="space-y-4">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full rounded-lg" />
              ))}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

// ── Not Found State ──────────────────────────────────────────────────────

function NotFoundState({ appId }: { appId: string }) {
  return (
    <motion.div
      variants={fadeUpVariants}
      initial="hidden"
      animate="visible"
      className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border/40 bg-card/10 px-6 py-20 text-center"
    >
      <div className="mb-5 flex h-14 w-14 items-center justify-center rounded-2xl border border-border/40 bg-muted/30">
        <Globe className="h-7 w-7 text-muted-foreground/50" />
      </div>
      <h3 className="mb-1.5 text-lg font-semibold tracking-tight text-foreground/90">
        Application not found
      </h3>
      <p className="mb-6 max-w-sm text-sm leading-relaxed text-muted-foreground">
        The application <span className="font-mono text-xs text-foreground/60">{appId}</span> could not be
        found. It may have been removed or the URL may be incorrect.
      </p>
      <Button variant="outline" size="sm" className="gap-1.5" asChild>
        <Link to="/applications">
          <ArrowLeft className="h-3.5 w-3.5" />
          Back to Applications
        </Link>
      </Button>
    </motion.div>
  );
}

// ── Error State ──────────────────────────────────────────────────────────

function ErrorState({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <motion.div
      variants={fadeUpVariants}
      initial="hidden"
      animate="visible"
      className="space-y-4"
    >
      <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-red-500/20 bg-red-500/5 px-6 py-16 text-center">
        <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl border border-red-500/20 bg-red-500/10">
          <AlertCircle className="h-6 w-6 text-red-400" />
        </div>
        <h3 className="mb-1.5 text-base font-semibold text-foreground/90">
          Failed to load application
        </h3>
        <p className="mb-4 max-w-md text-sm text-muted-foreground">
          {message}
        </p>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            className="gap-1.5"
            onClick={onRetry}
          >
            <RotateCcw className="h-3.5 w-3.5" />
            Retry
          </Button>
          <Button variant="ghost" size="sm" className="gap-1.5" asChild>
            <Link to="/applications">
              <ArrowLeft className="h-3.5 w-3.5" />
              Back to Applications
            </Link>
          </Button>
        </div>
      </div>
    </motion.div>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────

export default function ApplicationDetailPage() {
  const { id } = useParams<{ id: string }>();

  const {
    data: app,
    isLoading,
    error,
    refetch,
  } = useApplication(id ?? '');

  const {
    data: timelineData,
    isLoading: timelineLoading,
  } = useTimeline(id ?? '');

  const [activeTab, setActiveTab] = useState('overview');

  // ── Guard: no application ID ──
  if (!id) {
    return (
      <motion.div
        className="flex flex-col gap-6 p-6"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border/40 bg-card/10 px-6 py-20 text-center">
          <div className="mb-5 flex h-14 w-14 items-center justify-center rounded-2xl border border-border/40 bg-muted/30">
            <AlertCircle className="h-7 w-7 text-muted-foreground/50" />
          </div>
          <h3 className="mb-1.5 text-lg font-semibold tracking-tight text-foreground/90">
            Missing application ID
          </h3>
          <p className="mb-6 max-w-sm text-sm leading-relaxed text-muted-foreground">
            No application ID was provided in the URL.
          </p>
          <Button variant="outline" size="sm" className="gap-1.5" asChild>
            <Link to="/applications">
              <ArrowLeft className="h-3.5 w-3.5" />
              Back to Applications
            </Link>
          </Button>
        </div>
      </motion.div>
    );
  }

  // ── Loading state ──
  if (isLoading) {
    return (
      <motion.div
        className="flex flex-col gap-6 p-6"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        <PageSkeleton />
      </motion.div>
    );
  }

  // ── Error state ──
  if (error) {
    return (
      <motion.div
        className="flex flex-col gap-6 p-6"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        <ErrorState
          message={
            (error as Error)?.message ||
            'An unexpected error occurred while loading the application.'
          }
          onRetry={() => refetch()}
        />
      </motion.div>
    );
  }

  // ── Not found state ──
  if (!app) {
    return (
      <motion.div
        className="flex flex-col gap-6 p-6"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        <NotFoundState appId={id} />
      </motion.div>
    );
  }

  const phase = app.status?.phase ?? 'Stopped';
  const health = app.status?.health ?? 'Unknown';
  const phaseConf = phaseBadge[phase] ?? phaseBadge.Stopped;
  const healthConf = healthBadge[health] ?? { variant: 'secondary' as const, label: health };
  const url = app.status?.url;
  const deploymentHistory = app.status?.deploymentHistory ?? [];

  return (
    <motion.div
      className="flex flex-col gap-6 p-6"
      variants={containerVariants}
      initial="hidden"
      animate="visible"
    >
      {/* ── Breadcrumb ─────────────────────────────────────────────── */}
      <motion.nav
        variants={fadeUpVariants}
        className="flex items-center gap-2 text-sm text-muted-foreground"
        aria-label="Breadcrumb"
      >
        <Link
          to="/applications"
          className="hover:text-foreground transition-colors"
        >
          Applications
        </Link>
        <span className="text-muted-foreground/40" aria-hidden="true">
          /
        </span>
        <span className="text-foreground/80 font-medium truncate max-w-[200px] sm:max-w-[400px]">
          {app.metadata.name}
        </span>
      </motion.nav>

      {/* ── Page Header ────────────────────────────────────────────── */}
      <motion.div
        variants={fadeUpVariants}
        className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
      >
        {/* Left: Back + Title + Meta */}
        <div className="flex items-start gap-4 min-w-0">
          <Button
            variant="ghost"
            size="icon"
            className="mt-0.5 shrink-0"
            asChild
          >
            <Link to="/applications" aria-label="Back to applications">
              <ArrowLeft className="h-4 w-4" />
            </Link>
          </Button>

          <div className="min-w-0">
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl truncate">
                {app.metadata.name}
              </h1>

              {/* Phase badge */}
              <Badge
                variant={phaseConf.variant}
                className={cn(
                  'gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-medium',
                  phase === 'Deploying' && 'animate-pulse',
                )}
              >
                {phase === 'Deploying' && (
                  <span className="relative flex h-1.5 w-1.5">
                    <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-current opacity-75" />
                    <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-current" />
                  </span>
                )}
                {phaseConf.label}
              </Badge>

              {/* Health badge */}
              <Badge
                variant={healthConf.variant}
                className="gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-medium"
              >
                {getHealthIcon(health)}
                {healthConf.label}
              </Badge>
            </div>

            <div className="mt-1 flex items-center gap-3 text-sm text-muted-foreground">
              <span className="font-mono text-xs text-muted-foreground/60">
                {app.metadata.id}
              </span>
              {url && (
                <>
                  <span
                    className="h-3 w-px bg-border/40"
                    aria-hidden="true"
                  />
                  <a
                    href={url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-xs text-primary hover:text-primary/80 transition-colors"
                  >
                    <ExternalLink className="h-3 w-3" />
                    Open app
                  </a>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Right: Desktop external link */}
        {url && (
          <div className="hidden sm:flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              className="gap-1.5"
              asChild
            >
              <a href={url} target="_blank" rel="noopener noreferrer">
                <ExternalLink className="h-3.5 w-3.5" />
                Open Application
              </a>
            </Button>
          </div>
        )}
      </motion.div>

      {/* ── Tab Navigation ─────────────────────────────────────────── */}
      <motion.div variants={fadeUpVariants}>
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <div className="overflow-x-auto -mx-4 px-4">
            <TabsList className="mb-1 h-auto w-full justify-start overflow-x-auto rounded-none bg-transparent p-0">
              {APP_TABS.map((tab) => {
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
            </TabsList>
          </div>

          <div className="h-px bg-border/30" />

          {/* ── Overview Tab ──────────────────────────────────────── */}
          <TabsContent value="overview">
            <motion.div
              key="overview"
              variants={tabContentVariants}
              initial="hidden"
              animate="visible"
              className="pt-6"
            >
              <OverviewTab app={app} />
            </motion.div>
          </TabsContent>

          {/* ── Deployments Tab ───────────────────────────────────── */}
          <TabsContent value="deployments">
            <motion.div
              key="deployments"
              variants={tabContentVariants}
              initial="hidden"
              animate="visible"
              className="pt-6"
            >
              <DeploymentsTab
                appId={id}
                deploymentHistory={deploymentHistory}
              />
            </motion.div>
          </TabsContent>

          {/* ── Timeline Tab ──────────────────────────────────────── */}
          <TabsContent value="timeline">
            <motion.div
              key="timeline"
              variants={tabContentVariants}
              initial="hidden"
              animate="visible"
              className="pt-6"
            >
              <TimelineTab
                timeline={timelineData}
                loading={timelineLoading}
              />
            </motion.div>
          </TabsContent>

          {/* ── Logs Tab ──────────────────────────────────────────── */}
          <TabsContent value="logs">
            <motion.div
              key="logs"
              variants={tabContentVariants}
              initial="hidden"
              animate="visible"
              className="pt-6"
            >
              <LogsTab appId={id} />
            </motion.div>
          </TabsContent>

          {/* ── Settings Tab ──────────────────────────────────────── */}
          <TabsContent value="settings">
            <motion.div
              key="settings"
              variants={tabContentVariants}
              initial="hidden"
              animate="visible"
              className="pt-6"
            >
              <SettingsTab app={app} />
            </motion.div>
          </TabsContent>
        </Tabs>
      </motion.div>
    </motion.div>
  );
}
