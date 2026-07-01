import { useState, useMemo } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useApplication } from '@/hooks/useApplications';
import { usePageTitle } from '@/hooks/usePageTitle';
import { useTimeline } from '@/hooks/useDeployments';
import type { DeploymentReport } from '@/hooks/useApplications';
import type { TimelineResponse } from '@/hooks/useDeployments';
import { TimelineStepContent } from '@/components/applications/DeploymentTimelineTab';
import { DeploymentLogsTab } from '@/components/applications/DeploymentLogsTab';
import { DeploymentWorkflowTab } from '@/components/applications/DeploymentWorkflowTab';
import { DeploymentArtifactsTab } from '@/components/applications/DeploymentArtifactsTab';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { HealthIndicator } from '@/components/ui/health-indicator';
import { Separator } from '@/components/ui/separator';
import { ErrorState } from '@/components/ui/error-state';
import { EmptyState } from '@/components/ui/empty-state';
import {
  ArrowLeft,
  ExternalLink,
  GitCompare,
  RotateCcw,
  ChevronRight,
  Clock,
  GitBranch,
  GitCommitHorizontal,
  Globe,
  Terminal,
  Layers,
  Activity,
  Heart,
  Monitor,
  Box,
  Calendar,
  BookOpen,
  Package,
  Hash,
  AlertCircle,
  CheckCircle2,
  XCircle,
} from 'lucide-react';
import { cn, relativeTime, truncate } from '@/lib/utils';

/* ── Tab config ─────────────────────────────────────────── */
const DEPLOY_TABS = [
  { value: 'timeline',    label: 'Timeline',    icon: Activity },
  { value: 'logs',        label: 'Logs',        icon: Terminal },
  { value: 'workflow',    label: 'Workflow',    icon: Heart },
  { value: 'artifacts',   label: 'Artifacts',   icon: Box },
] as const;

/* ── Helpers ────────────────────────────────────────────── */
function mapTimelineStatus(status: string) {
  const lower = status?.toLowerCase() ?? '';
  if (['success', 'succeeded'].includes(lower)) return 'succeeded' as const;
  if (['failure', 'failed', 'error'].includes(lower)) return 'failed' as const;
  if (lower === 'running') return 'running' as const;
  if (lower === 'skipped') return 'skipped' as const;
  if (lower === 'cancelled') return 'cancelled' as const;
  return 'pending' as const;
}

function formatDateTime(dateStr?: string): string {
  if (!dateStr) return '\u2014';
  try {
    return new Date(dateStr).toLocaleString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  } catch {
    return dateStr;
  }
}

/* ── Loading Skeleton ───────────────────────────────────── */
function PageSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <Skeleton className="h-4 w-48" />
      <div className="flex flex-col gap-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-4">
            <Skeleton className="h-9 w-9 rounded-md shrink-0" />
            <div className="min-w-0 space-y-2">
              <Skeleton className="h-7 w-48" />
              <div className="flex gap-2">
                <Skeleton className="h-5 w-20 rounded-sm" />
                <Skeleton className="h-5 w-24 rounded-sm" />
                <Skeleton className="h-5 w-16 rounded-sm" />
              </div>
            </div>
          </div>
          <div className="flex gap-2">
            <Skeleton className="h-8 w-24 rounded-md" />
            <Skeleton className="h-8 w-24 rounded-md" />
          </div>
        </div>
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_280px] gap-6">
        <div className="space-y-4">
          <Skeleton className="h-10 w-full rounded-md" />
          <Skeleton className="h-80 rounded-md" />
        </div>
        <div className="space-y-4">
          <Skeleton className="h-64 rounded-md" />
        </div>
      </div>
    </div>
  );
}

/* ── Not Found State ────────────────────────────────────── */
function NotFoundState({ appId }: { appId: string }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-surface px-6 py-20 text-center">
      <Box className="h-12 w-12 text-text-muted mb-4" />
      <h3 className="text-h3 text-foreground mb-1.5">Deployment not found</h3>
      <p className="text-small text-text-secondary mb-6 max-w-sm">
        The deployment could not be found for application{' '}
        <span className="font-mono text-code-sm text-text-muted">{appId}</span>.
      </p>
      <Button variant="secondary" size="sm" className="gap-1.5" asChild>
        <Link to={`/applications/${appId}`}>
          <ArrowLeft className="h-3.5 w-3.5" />
          Back to Application
        </Link>
      </Button>
    </div>
  );
}

/* ── Detail Row (for right panel) ───────────────────────── */
function DetailRow({ icon, label, value }: { icon: React.ReactNode; label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-start gap-2.5 py-1.5">
      <span className="mt-0.5 shrink-0 text-text-muted">{icon}</span>
      <div className="min-w-0 flex-1">
        <span className="block text-caption text-text-secondary">{label}</span>
        <span className="mt-0.5 block text-small text-foreground font-medium break-all">
          {value ?? '\u2014'}
        </span>
      </div>
    </div>
  );
}

/* ── Deployment Summary (right panel) ──────────────────── */
function DeploymentSummaryPanel({ report, timeline }: { report: DeploymentReport; timeline?: TimelineResponse }) {
  const hasErrors = (report.errors?.length ?? 0) > 0;
  const hasWarnings = (report.warnings?.length ?? 0) > 0;

  return (
    <aside className="space-y-4" aria-label="Deployment summary">
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-small font-semibold text-text-secondary flex items-center gap-1.5">
            <Box className="h-3.5 w-3.5" />
            Deployment Summary
          </CardTitle>
        </CardHeader>
        <CardContent>
          <DetailRow
            icon={<Hash className="h-4 w-4" />}
            label="Deployment"
            value={<span className="tabular-nums">#{report.deploymentNumber}</span>}
          />
          <DetailRow
            icon={<Calendar className="h-4 w-4" />}
            label="Started"
            value={formatDateTime(report.startedAt)}
          />
          <DetailRow
            icon={<Calendar className="h-4 w-4" />}
            label="Completed"
            value={formatDateTime(report.completedAt)}
          />
          <DetailRow
            icon={<Clock className="h-4 w-4" />}
            label="Duration"
            value={report.duration || '\u2014'}
          />

          <Separator className="my-2" />

          <DetailRow
            icon={<Globe className="h-4 w-4" />}
            label="Repository"
            value={report.repository || '\u2014'}
          />
          <DetailRow
            icon={<GitBranch className="h-4 w-4" />}
            label="Branch"
            value={report.branch || '\u2014'}
          />
          <DetailRow
            icon={<GitCommitHorizontal className="h-4 w-4" />}
            label="Commit"
            value={report.commitSha ? truncate(report.commitSha, 7) : '\u2014'}
          />

          <Separator className="my-2" />

          <DetailRow
            icon={<Terminal className="h-4 w-4" />}
            label="Runtime"
            value={report.detectedRuntime || report.runtimeName || '\u2014'}
          />
          <DetailRow
            icon={<Package className="h-4 w-4" />}
            label="Buildpack"
            value={report.buildpack || '\u2014'}
          />
          <DetailRow
            icon={<Layers className="h-4 w-4" />}
            label="Workflow"
            value={report.workflowId || '\u2014'}
          />

          <Separator className="my-2" />

          <DetailRow
            icon={<ExternalLink className="h-4 w-4" />}
            label="Endpoint"
            value={report.endpoint || '\u2014'}
          />
          <DetailRow
            icon={<Heart className="h-4 w-4" />}
            label="Health"
            value={
              <Badge variant={
                report.healthStatus === 'Healthy' ? 'subtle-success' :
                report.healthStatus === 'Degraded' ? 'subtle-warning' :
                'subtle-danger'
              } className="text-caption">
                {report.healthStatus || '\u2014'}
              </Badge>
            }
          />
          <DetailRow
            icon={<Globe className="h-4 w-4" />}
            label="Environment"
            value={report.environment || '\u2014'}
          />
        </CardContent>
      </Card>

      {/* Errors & Warnings summary */}
      {(hasErrors || hasWarnings) && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-small font-semibold text-text-secondary flex items-center gap-1.5">
              <AlertCircle className="h-3.5 w-3.5" />
              Issues
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {hasErrors && (
              <div className="flex items-center gap-2">
                <XCircle className="h-3.5 w-3.5 text-danger shrink-0" />
                <span className="text-small text-danger">{report.errors.length} error{report.errors.length > 1 ? 's' : ''}</span>
              </div>
            )}
            {hasWarnings && (
              <div className="flex items-center gap-2">
                <AlertCircle className="h-3.5 w-3.5 text-warning shrink-0" />
                <span className="text-small text-warning">{report.warnings.length} warning{report.warnings.length > 1 ? 's' : ''}</span>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </aside>
  );
}

/* ── Main Page ──────────────────────────────────────────── */
export default function DeploymentDetailPage() {
  const { appId, deploymentNumber: depNumParam } = useParams<{
    appId: string;
    deploymentNumber: string;
  }>();
  const navigate = useNavigate();

  const deploymentNumber = depNumParam ? parseInt(depNumParam, 10) : NaN;

  const {
    data: app,
    isLoading: appLoading,
    error: appError,
    refetch,
  } = useApplication(appId ?? '');

  usePageTitle(
    isNaN(deploymentNumber) || !app
      ? 'Deployment Detail'
      : `#${deploymentNumber} • ${app.metadata.name} • Deployments`,
  );

  const {
    data: timeline,
    isLoading: timelineLoading,
  } = useTimeline(appId ?? '', isNaN(deploymentNumber) ? undefined : deploymentNumber);

  const [activeTab, setActiveTab] = useState('timeline');

  // ── Derive deployment report from app data ──
  const report = useMemo<DeploymentReport | null>(() => {
    if (!app || isNaN(deploymentNumber)) return null;
    const history = app.status?.deploymentHistory ?? [];
    const found = history.find((d) => d.deploymentNumber === deploymentNumber);
    return found ?? null;
  }, [app, deploymentNumber]);

  // ── Validate params ──
  if (!appId || isNaN(deploymentNumber)) {
    return (
      <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-surface px-6 py-20 text-center">
        <AlertCircle className="h-10 w-10 text-text-muted mb-4" />
        <h3 className="text-h3 text-foreground mb-1.5">Invalid URL</h3>
        <p className="text-small text-text-secondary mb-6 max-w-sm">
          Missing or invalid application ID or deployment number.
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
  if (appLoading) {
    return (
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="flex flex-col gap-6">
        <PageSkeleton />
      </motion.div>
    );
  }

  // ── Error ──
  if (appError) {
    return (
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="flex flex-col gap-6">
        <ErrorState
          title="Failed to load deployment"
          message={
            (appError as Error)?.message ||
            'An unexpected error occurred while loading the deployment.'
          }
          onRetry={() => refetch()}
        />
      </motion.div>
    );
  }

  // ── App not found ──
  if (!app) {
    return (
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="flex flex-col gap-6">
        <NotFoundState appId={appId} />
      </motion.div>
    );
  }

  // ── Deployment not found ──
  if (!report) {
    return (
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="flex flex-col gap-6">
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-surface px-6 py-20 text-center">
          <Box className="h-12 w-12 text-text-muted mb-4" />
          <h3 className="text-h3 text-foreground mb-1.5">Deployment #{deploymentNumber} not found</h3>
          <p className="text-small text-text-secondary mb-6 max-w-sm">
            This deployment does not exist in the application history.
          </p>
          <Button variant="secondary" size="sm" className="gap-1.5" asChild>
            <Link to={`/applications/${appId}`}>
              <ArrowLeft className="h-3.5 w-3.5" />
              Back to Application
            </Link>
          </Button>
        </div>
      </motion.div>
    );
  }

  const appName = app.metadata.name;
  const prevDeployment = deploymentNumber > 1;

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
        <Link to={`/applications/${appId}`} className="hover:text-foreground transition-colors truncate max-w-[150px]">
          {appName}
        </Link>
        <ChevronRight className="h-3.5 w-3.5 text-text-muted" aria-hidden="true" />
        <span className="text-foreground font-medium">Deployment #{deploymentNumber}</span>
      </nav>

      {/* ════════════════════ HEADER ════════════════════ */}
      <div className="flex flex-col gap-4">
        {/* Row 1: Back + Title + Status + Actions */}
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3 min-w-0 flex-1">
            <Button
              variant="icon-ghost"
              size="icon-sm"
              className="mt-0.5 shrink-0"
              asChild
            >
              <Link to={`/applications/${appId}`} aria-label="Back to application">
                <ArrowLeft className="h-4 w-4" />
              </Link>
            </Button>

            <div className="min-w-0 flex-1">
              {/* Title row */}
              <div className="flex items-center gap-2.5 flex-wrap">
                <h1 className="text-h1 text-foreground truncate">
                  Deployment <span className="text-accent">#{deploymentNumber}</span>
                </h1>

                {/* Build status */}
                <Badge
                  variant={report.buildSuccess ? 'subtle-success' : 'subtle-danger'}
                  className="gap-1.5"
                >
                  {report.buildSuccess
                    ? <CheckCircle2 className="h-3.5 w-3.5" />
                    : <XCircle className="h-3.5 w-3.5" />
                  }
                  {report.buildSuccess ? 'Succeeded' : 'Failed'}
                </Badge>

                {/* Health */}
                <Badge
                  variant={
                    report.healthStatus === 'Healthy' ? 'subtle-success' :
                    report.healthStatus === 'Degraded' ? 'subtle-warning' :
                    'subtle-danger'
                  }
                  className="gap-1.5"
                >
                  <HealthIndicator
                    status={
                      report.healthStatus === 'Healthy' ? 'healthy' :
                      report.healthStatus === 'Degraded' ? 'degraded' :
                      'failed'
                    }
                    size="sm"
                  />
                  {report.healthStatus || 'Unknown'}
                </Badge>
              </div>

              {/* Meta row */}
              <div className="flex items-center gap-3 mt-1 flex-wrap text-small text-text-muted">
                <span className="inline-flex items-center gap-1 tabular-nums">
                  <Clock className="h-3.5 w-3.5" />
                  {report.duration || '\u2014'}
                </span>

                {report.commitSha && (
                  <span className="inline-flex items-center gap-1 font-mono text-code-sm">
                    <GitCommitHorizontal className="h-3.5 w-3.5" />
                    {truncate(report.commitSha, 7)}
                  </span>
                )}

                {report.branch && (
                  <span className="inline-flex items-center gap-1">
                    <GitBranch className="h-3.5 w-3.5" />
                    {report.branch}
                  </span>
                )}

                {report.detectedRuntime && (
                  <span className="inline-flex items-center gap-1">
                    <Terminal className="h-3.5 w-3.5" />
                    {report.detectedRuntime}
                  </span>
                )}

                {report.environment && (
                  <Badge variant="subtle-neutral" className="text-caption uppercase tracking-wider">
                    {report.environment}
                  </Badge>
                )}

                {report.workflowId && (
                  <span className="inline-flex items-center gap-1 text-code-sm text-text-muted">
                    <Layers className="h-3.5 w-3.5" />
                    {truncate(report.workflowId, 20)}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Action buttons */}
          <div className="flex items-center gap-2 shrink-0">
            {prevDeployment && (
              <Button
                variant="secondary"
                size="sm"
                className="gap-1.5"
                onClick={() => navigate(`/applications/${appId}/compare?from=${deploymentNumber - 1}&to=${deploymentNumber}`)}
              >
                <GitCompare className="h-3.5 w-3.5" />
                Compare
              </Button>
            )}
            <Button
              variant="primary"
              size="sm"
              className="gap-1.5"
              onClick={() => {
                // Redeploy - open a confirmation or trigger redeploy
                navigate(`/applications/${appId}/deployments/${deploymentNumber}/redeploy`);
              }}
            >
              <RotateCcw className="h-3.5 w-3.5" />
              Redeploy
            </Button>
          </div>
        </div>
      </div>

      {/* ════════════════════ MAIN CONTENT ════════════════════ */}
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_280px] gap-6">
        {/* ── Left: Tabs ── */}
        <div className="min-w-0">
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <div className="overflow-x-auto -mx-1">
              <TabsList className="mb-0 h-auto w-full justify-start overflow-x-auto rounded-none bg-transparent p-0 border-b border-border gap-0">
                {DEPLOY_TABS.map((tab) => {
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
                <TabsContent value="timeline">
                  <TimelineStepContent timeline={timeline} loading={timelineLoading} />
                </TabsContent>
                <TabsContent value="logs">
                  <DeploymentLogsTab appId={appId} />
                </TabsContent>
                <TabsContent value="workflow">
                  <DeploymentWorkflowTab report={report} timeline={timeline} />
                </TabsContent>
                <TabsContent value="artifacts">
                  <DeploymentArtifactsTab report={report} appName={appName} />
                </TabsContent>
              </motion.div>
            </div>
          </Tabs>
        </div>

        {/* ── Right: Summary Panel ── */}
        <div className="hidden lg:block">
          <div className="sticky top-6">
            <DeploymentSummaryPanel report={report} timeline={timeline} />
          </div>
        </div>
      </div>
    </motion.div>
  );
}
