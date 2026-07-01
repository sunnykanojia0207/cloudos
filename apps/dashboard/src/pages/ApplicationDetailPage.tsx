import { useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useApplication, useApplications } from '@/hooks/useApplications';
import { usePageTitle } from '@/hooks/usePageTitle';
import { useTimeline } from '@/hooks/useDeployments';
import { OverviewTab } from '@/components/applications/OverviewTab';
import { DeploymentsTab } from '@/components/applications/DeploymentsTab';
import { TimelineTab } from '@/components/applications/TimelineTab';
import { LogsTab } from '@/components/applications/LogsTab';
import { WorkflowTab } from '@/components/applications/WorkflowTab';
import { MonitoringTab } from '@/components/applications/MonitoringTab';
import { SettingsTab } from '@/components/applications/SettingsTab';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { HealthIndicator } from '@/components/ui/health-indicator';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import { Separator } from '@/components/ui/separator';
import {
  ArrowLeft,
  ExternalLink,
  Rocket,
  BookOpen,
  Layers,
  Terminal,
  Activity,
  Heart,
  Monitor,
  Settings,
  Globe,
  GitBranch,
  GitCommitHorizontal,
  Clock,
  Box,
  History,
  ChevronRight,
  AlertCircle,
} from 'lucide-react';
import { cn, relativeTime, truncate } from '@/lib/utils';

/* ── Tab config ─────────────────────────────────────────── */
const APP_TABS = [
  { value: 'overview',    label: 'Overview',    icon: BookOpen },
  { value: 'deployments', label: 'Deployments', icon: Layers },
  { value: 'timeline',    label: 'Timeline',    icon: Activity },
  { value: 'logs',        label: 'Logs',        icon: Terminal },
  { value: 'workflow',    label: 'Workflow',    icon: Heart },
  { value: 'monitoring',  label: 'Monitoring',  icon: Monitor },
  { value: 'settings',    label: 'Settings',    icon: Settings },
] as const;

/* ── Helpers ────────────────────────────────────────────── */
function phaseToHealth(phase: string) {
  switch (phase) {
    case 'Running':   return 'running' as const;
    case 'Deploying': return 'deploying' as const;
    case 'Failed':    return 'failed' as const;
    case 'Stopped':   return 'stopped' as const;
    default:          return 'stopped' as const;
  }
}

/* ── Loading Skeleton ───────────────────────────────────── */
function PageSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      {/* Breadcrumb */}
      <Skeleton className="h-4 w-40" />

      {/* Summary bar */}
      <div className="flex flex-col gap-4">
        <div className="flex items-start gap-4">
          <Skeleton className="h-9 w-9 rounded-md shrink-0" />
          <div className="min-w-0 flex-1 space-y-2">
            <Skeleton className="h-7 w-56" />
            <Skeleton className="h-4 w-80" />
          </div>
          <div className="flex gap-2 shrink-0">
            <Skeleton className="h-8 w-28 rounded-md" />
            <Skeleton className="h-8 w-28 rounded-md" />
          </div>
        </div>
      </div>

      {/* Tabs + Content */}
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_260px] gap-6">
        <div className="space-y-4">
          <Skeleton className="h-10 w-full rounded-md" />
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <Skeleton className="h-64 rounded-md" />
            <Skeleton className="h-64 rounded-md" />
          </div>
        </div>

        {/* Right panel skeleton */}
        <div className="space-y-4">
          <Skeleton className="h-48 rounded-md" />
          <Skeleton className="h-32 rounded-md" />
        </div>
      </div>
    </div>
  );
}

/* ── Not Found State ────────────────────────────────────── */
function NotFoundState({ appId }: { appId: string }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-surface px-6 py-20 text-center">
      <div className="mb-5 flex h-14 w-14 items-center justify-center rounded-2xl border border-border bg-surface-elevated">
        <Box className="h-7 w-7 text-text-muted" />
      </div>
      <h3 className="mb-1.5 text-h3 text-foreground">Application not found</h3>
      <p className="mb-6 max-w-sm text-small text-text-secondary">
        The application <span className="font-mono text-code text-text-muted">{appId}</span> could not be
        found. It may have been removed or the URL may be incorrect.
      </p>
      <Button variant="secondary" size="sm" className="gap-1.5" asChild>
        <Link to="/applications">
          <ArrowLeft className="h-3.5 w-3.5" />
          Back to Applications
        </Link>
      </Button>
    </div>
  );
}

/* ── Right Panel ────────────────────────────────────────── */
function RightPanel({ appId }: { appId: string }) {
  const { data: app } = useApplication(appId);
  const { data: allApps } = useApplications();

  if (!app) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-48 rounded-md" />
        <Skeleton className="h-32 rounded-md" />
      </div>
    );
  }

  const phase = app.status?.phase ?? 'Stopped';
  const health = app.status?.health ?? 'Unknown';
  const url = app.status?.url;
  const lastReport = app.status?.lastReport;
  const deploymentCount = app.status?.deploymentCount ?? 0;
  const runtime = app.spec?.runtime?.type ?? lastReport?.detectedRuntime;

  // Recent activity: last 3 deployment reports + other apps for context
  const history = app.status?.deploymentHistory ?? [];
  const recentDeployments = [...history]
    .sort((a, b) => b.deploymentNumber - a.deploymentNumber)
    .slice(0, 3);

  return (
    <aside className="space-y-4" aria-label="Application quick summary">
      {/* ── Quick Summary Card ── */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-small font-semibold text-text-secondary flex items-center gap-1.5">
            <Box className="h-3.5 w-3.5" />
            Quick Summary
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {/* Health */}
          <div className="flex items-center justify-between">
            <span className="text-small text-text-secondary">Health</span>
            <HealthIndicator status={phaseToHealth(phase)} showLabel size="sm" />
          </div>

          {/* Status */}
          <div className="flex items-center justify-between">
            <span className="text-small text-text-secondary">Status</span>
            <Badge variant={
              phase === 'Running' ? 'subtle-success' :
              phase === 'Deploying' ? 'subtle-accent' :
              phase === 'Failed' ? 'subtle-danger' :
              'subtle-neutral'
            } className="text-caption font-medium">
              {phase}
            </Badge>
          </div>

          {/* Runtime */}
          {runtime && (
            <div className="flex items-center justify-between">
              <span className="text-small text-text-secondary">Runtime</span>
              <span className="text-small text-foreground font-medium">{runtime}</span>
            </div>
          )}

          {/* Deployment # */}
          <div className="flex items-center justify-between">
            <span className="text-small text-text-secondary">Deployments</span>
            <span className="text-small text-foreground font-medium tabular-nums">{deploymentCount}</span>
          </div>

          {/* URL */}
          {url && (
            <div className="flex items-center justify-between min-w-0">
              <span className="text-small text-text-secondary shrink-0 mr-2">URL</span>
              <a
                href={url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-small text-accent hover:text-accent-hover truncate inline-flex items-center gap-1 min-w-0"
              >
                <span className="truncate">{truncate(url.replace(/^https?:\/\//, ''), 20)}</span>
                <ExternalLink className="h-3 w-3 shrink-0" />
              </a>
            </div>
          )}
        </CardContent>
      </Card>

      {/* ── Recent Activity ── */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-small font-semibold text-text-secondary flex items-center gap-1.5">
            <History className="h-3.5 w-3.5" />
            Recent Activity
          </CardTitle>
        </CardHeader>
        <CardContent>
          {recentDeployments.length === 0 ? (
            <p className="text-small text-text-muted py-2 text-center">
              No recent deployments
            </p>
          ) : (
            <div className="space-y-2">
              {recentDeployments.map((dep) => (
                <div key={dep.deploymentNumber} className="flex items-center gap-2 min-w-0">
                  <div className={cn(
                    'h-2 w-2 shrink-0 rounded-full',
                    dep.buildSuccess ? 'bg-success' : 'bg-danger',
                  )} aria-hidden="true" />
                  <span className="text-caption text-text-muted tabular-nums shrink-0">
                    #{dep.deploymentNumber}
                  </span>
                  <span className="text-caption text-text-muted shrink-0">
                    {dep.duration}
                  </span>
                  <span className="text-caption text-text-muted truncate">
                    {relativeTime(dep.startedAt)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </aside>
  );
}

/* ── Main Page ──────────────────────────────────────────── */
export default function ApplicationDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const {
    data: app,
    isLoading,
    error,
    refetch,
  } = useApplication(id ?? '');

  usePageTitle(app ? `${app.metadata.name} • Applications` : 'Application Detail');

  const {
    data: timelineData,
    isLoading: timelineLoading,
  } = useTimeline(id ?? '');

  const [activeTab, setActiveTab] = useState('overview');

  // ── Guard: no ID ──
  if (!id) {
    return (
      <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-surface px-6 py-20 text-center">
        <AlertCircle className="h-10 w-10 text-text-muted mb-4" />
        <h3 className="text-h3 text-foreground mb-1.5">Missing application ID</h3>
        <p className="text-small text-text-secondary mb-6 max-w-sm">
          No application ID was provided in the URL.
        </p>
        <Button variant="secondary" size="sm" className="gap-1.5" asChild>
          <Link to="/applications">
            <ArrowLeft className="h-3.5 w-3.5" />
            Back to Applications
          </Link>
        </Button>
      </div>
    );
  }

  // ── Loading ──
  if (isLoading) {
    return (
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="flex flex-col gap-6"
      >
        <PageSkeleton />
      </motion.div>
    );
  }

  // ── Error ──
  if (error) {
    return (
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="flex flex-col gap-6"
      >
        <ErrorState
          title="Failed to load application"
          message={
            (error as Error)?.message ||
            'An unexpected error occurred while loading the application. Ensure the CloudOS kernel is running.'
          }
          onRetry={() => refetch()}
        />
      </motion.div>
    );
  }

  // ── Not found ──
  if (!app) {
    return (
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="flex flex-col gap-6"
      >
        <NotFoundState appId={id} />
      </motion.div>
    );
  }

  // ── Derived data ──
  const phase = app.status?.phase ?? 'Stopped';
  const health = app.status?.health ?? 'Unknown';
  const url = app.status?.url;
  const repoUrl = app.spec?.source?.url;
  const branch = app.spec?.source?.branch ?? app.status?.lastReport?.branch;
  const commitSha = app.status?.lastReport?.commitSha;
  const env = app.spec?.settings?.environment ?? app.status?.lastReport?.environment;
  const runtimeType = app.spec?.runtime?.type ?? app.status?.lastReport?.detectedRuntime;
  const deploymentHistory = app.status?.deploymentHistory ?? [];

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="flex flex-col gap-6"
    >
      {/* ════════════════════ BREADCRUMB ════════════════════ */}
      <nav className="flex items-center gap-1.5 text-small text-text-secondary" aria-label="Breadcrumb">
        <Link to="/applications" className="hover:text-foreground transition-colors">
          Applications
        </Link>
        <ChevronRight className="h-3.5 w-3.5 text-text-muted" aria-hidden="true" />
        <span className="text-foreground font-medium truncate max-w-[300px]">
          {app.metadata.name}
        </span>
      </nav>

      {/* ════════════════════ TOP SUMMARY BAR ════════════════ */}
      <div className="flex flex-col gap-4">
        {/* Row 1: Back + Name + Badges + Buttons */}
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3 min-w-0 flex-1">
            <Button
              variant="icon-ghost"
              size="icon-sm"
              className="mt-0.5 shrink-0"
              asChild
            >
              <Link to="/applications" aria-label="Back to applications">
                <ArrowLeft className="h-4 w-4" />
              </Link>
            </Button>

            <div className="min-w-0 flex-1">
              {/* Name row */}
              <div className="flex items-center gap-2.5 flex-wrap">
                <h1 className="text-h1 text-foreground truncate">
                  {app.metadata.name}
                </h1>

                {/* Environment badge */}
                {env && (
                  <Badge variant="subtle-neutral" className="text-caption uppercase tracking-wider font-medium">
                    {env}
                  </Badge>
                )}

                {/* Runtime */}
                {runtimeType && (
                  <span className="hidden sm:inline-flex items-center gap-1 text-small text-text-secondary">
                    <Terminal className="h-3.5 w-3.5" />
                    {runtimeType}
                  </span>
                )}
              </div>

              {/* Meta row: ID + repo + branch + commit */}
              <div className="flex items-center gap-3 mt-1 flex-wrap text-small text-text-muted">
                <span className="font-mono text-code-sm">{app.metadata.id}</span>

                {repoUrl && (
                  <span className="inline-flex items-center gap-1 min-w-0">
                    <Globe className="h-3.5 w-3.5 shrink-0" />
                    <span className="truncate">{truncate(repoUrl.replace(/^https?:\/\//, ''), 35)}</span>
                  </span>
                )}

                {branch && (
                  <span className="inline-flex items-center gap-1">
                    <GitBranch className="h-3.5 w-3.5 shrink-0" />
                    {branch}
                  </span>
                )}

                {commitSha && (
                  <span className="inline-flex items-center gap-1 font-mono text-code-sm">
                    <GitCommitHorizontal className="h-3.5 w-3.5" />
                    {truncate(commitSha, 7)}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Action buttons */}
          <div className="flex items-center gap-2 shrink-0">
            {/* Phase + Health indicators (desktop) */}
            <div className="hidden md:flex items-center gap-2 mr-1">
              <HealthIndicator status={phaseToHealth(phase)} showLabel size="sm" />
            </div>

            {url && (
              <Button variant="secondary" size="sm" className="gap-1.5" asChild>
                <a href={url} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-3.5 w-3.5" />
                  Open
                </a>
              </Button>
            )}
            <Button variant="primary" size="sm" className="gap-1.5">
              <Rocket className="h-3.5 w-3.5" />
              Deploy
            </Button>
          </div>
        </div>
      </div>

      {/* ════════════════════ MAIN CONTENT ════════════════════ */}
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_260px] gap-6">
        {/* ── Left: Tabs ── */}
        <div className="min-w-0">
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <div className="overflow-x-auto -mx-1">
              <TabsList className="mb-0 h-auto w-full justify-start overflow-x-auto rounded-none bg-transparent p-0 border-b border-border gap-0">
                {APP_TABS.map((tab) => {
                  const Icon = tab.icon;
                  return (
                    <TabsTrigger
                      key={tab.value}
                      value={tab.value}
                      className={cn(
                        'relative inline-flex items-center gap-1.5 rounded-none px-3 py-2.5 text-small font-medium transition-colors shrink-0',
                        'data-[state=active]:text-foreground data-[state=active]:after:absolute data-[state=active]:after:bottom-0 data-[state=active]:after:left-0 data-[state=active]:after:right-0 data-[state=active]:after:h-0.5 data-[state=active]:after:bg-accent',
                        'data-[state=inactive]:text-text-secondary data-[state=inactive]:hover:text-foreground',
                      )}
                    >
                      <Icon className="h-3.5 w-3.5" />
                      {tab.label}
                    </TabsTrigger>
                  );
                })}
              </TabsList>
            </div>

            {/* Tab content */}
            <div className="pt-5">
              <motion.div
                key={activeTab}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.15, ease: [0, 0, 0.2, 1] }}
              >
                <TabsContent value="overview"><OverviewTab app={app} /></TabsContent>
                <TabsContent value="deployments"><DeploymentsTab appId={id} deploymentHistory={deploymentHistory} /></TabsContent>
                <TabsContent value="timeline"><TimelineTab timeline={timelineData} loading={timelineLoading} /></TabsContent>
                <TabsContent value="logs"><LogsTab appId={id} /></TabsContent>
                <TabsContent value="workflow"><WorkflowTab app={app} /></TabsContent>
                <TabsContent value="monitoring"><MonitoringTab app={app} /></TabsContent>
                <TabsContent value="settings"><SettingsTab app={app} /></TabsContent>
              </motion.div>
            </div>
          </Tabs>
        </div>

        {/* ── Right: Panel ── */}
        <div className="hidden lg:block">
          <div className="sticky top-6">
            <RightPanel appId={id} />
          </div>
        </div>
      </div>
    </motion.div>
  );
}
