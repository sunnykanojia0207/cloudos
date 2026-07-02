import { useMemo } from 'react';
import { motion } from 'framer-motion';
import { useDeploymentEvents } from '@/hooks/useDeploymentEvents';
import { useApplication } from '@/hooks/useApplications';
import type { AppResource } from '@/hooks/useApplications';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Loader2,
  CheckCircle2,
  XCircle,
  Circle,
  ArrowRight,
  ExternalLink,
  RotateCcw,
  GitCompareArrows,
  Terminal,
  RefreshCw,
} from 'lucide-react';
import { cn } from '@/lib/utils';

/* ── Props ─────────────────────────────────────────────── */
export interface DeployProgressProps {
  app: AppResource;
}

/* ── Step icon helper ──────────────────────────────────── */
function StepIcon({ status }: { status: string }) {
  switch (status) {
    case 'completed':
    case 'succeeded':
      return <CheckCircle2 className="h-4 w-4 text-success" />;
    case 'failed':
      return <XCircle className="h-4 w-4 text-danger" />;
    case 'running':
      return <Loader2 className="h-4 w-4 text-accent animate-spin" />;
    case 'pending':
    case 'skipped':
      return <Circle className="h-4 w-4 text-text-muted" />;
    default:
      return <Circle className="h-4 w-4 text-text-muted" />;
  }
}

/* ── Deployment Steps ──────────────────────────────────── */
const DEFAULT_STEPS = [
  { id: 'validate', name: 'Validate Application', action: 'validate' },
  { id: 'clone', name: 'Clone Source Repository', action: 'source.clone' },
  { id: 'install', name: 'Install Dependencies', action: 'build.install' },
  { id: 'build', name: 'Build Artifact', action: 'build.execute' },
  { id: 'deploy', name: 'Deploy Application', action: 'provider.deploy' },
  { id: 'healthcheck', name: 'Health Check', action: 'health.check' },
  { id: 'complete', name: 'Complete Deployment', action: 'complete' },
];

/* ── Component ─────────────────────────────────────────── */
export function DeployProgress({ app }: DeployProgressProps) {
  const phase = app.status?.phase ?? '';
  const isDeploying = phase === 'Deploying' || phase === 'Creating';
  const isFailed = phase === 'Failed';
  const isRunning = phase === 'Running';
  const currentDeploymentId = app.status?.currentDeploymentId;
  const report = app.status?.lastReport;
  const url = app.status?.url;

  // Live SSE events for active deployment
  const liveState = useDeploymentEvents(
    isDeploying ? currentDeploymentId : null,
  );

  // Determine step statuses
  const steps = useMemo(() => {
    if (report?.workflowSteps && report.workflowSteps > 0) {
      // If we have node results from the report, use those
      const nodeResults = report.errors ?? [];
      return DEFAULT_STEPS.map((step) => {
        // Check if there's an error matching this step
        const hasError = nodeResults.some(
          (e: string) => e.toLowerCase().includes(step.name.toLowerCase()),
        );
        if (isRunning || (report.buildSuccess && !isFailed)) {
          // All steps completed if app is running
          return { ...step, status: 'completed' };
        }
        if (hasError || isFailed) {
          return { ...step, status: 'failed' };
        }
        return { ...step, status: 'pending' };
      });
    }

    // Live deployment - use SSE data
    const completedNodes = liveState.completedNodes ?? [];
    const failedNodes = liveState.failedNodes ?? [];
    const currentNode = liveState.currentNode;

    return DEFAULT_STEPS.map((step) => {
      if (failedNodes.includes(step.id)) return { ...step, status: 'failed' };
      if (completedNodes.includes(step.id)) return { ...step, status: 'completed' };
      if (currentNode === step.id) return { ...step, status: 'running' };
      return { ...step, status: 'pending' };
    });
  }, [report, isRunning, isFailed, liveState, currentDeploymentId]);

  const progress = liveState.progress > 0
    ? liveState.progress
    : isRunning ? 100
    : isFailed ? 0
    : steps.filter((s) => s.status === 'completed').length / Math.max(steps.length, 1) * 100;

  const duration = liveState.duration ?? report?.duration;
  const errorMessage = liveState.error ?? (report?.errors?.length ? report.errors[0] : null);

  // ── Running state ──
  if (isRunning && !isDeploying && !!url) {
    return (
      <div className="space-y-5">
        <Card className="border-success/30 bg-success/5">
          <CardHeader className="pb-2">
            <CardTitle className="text-body font-semibold flex items-center gap-2 text-success">
              <CheckCircle2 className="h-5 w-5" />
              Application Running
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-small text-text-secondary">Endpoint:</span>
              <a
                href={url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-small text-accent hover:text-accent-hover inline-flex items-center gap-1"
              >
                {url}
                <ExternalLink className="h-3.5 w-3.5" />
              </a>
            </div>
            <div className="flex flex-wrap gap-2 mt-3">
              <Button variant="secondary" size="sm" className="gap-1.5" asChild>
                <a href={url} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-3.5 w-3.5" />
                  Open Application
                </a>
              </Button>
              <Button variant="secondary" size="sm" className="gap-1.5">
                <Terminal className="h-3.5 w-3.5" />
                View Logs
              </Button>
              <Button variant="secondary" size="sm" className="gap-1.5">
                <RefreshCw className="h-3.5 w-3.5" />
                Redeploy
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // ── Failed state ──
  if (isFailed && !isDeploying) {
    return (
      <div className="space-y-5">
        <Card className="border-danger/30 bg-danger/5">
          <CardHeader className="pb-2">
            <CardTitle className="text-body font-semibold flex items-center gap-2 text-danger">
              <XCircle className="h-5 w-5" />
              Deployment Failed
            </CardTitle>
          </CardHeader>
          <CardContent>
            {errorMessage && (
              <div className="rounded-md border border-danger/20 bg-danger/10 px-3 py-2 mb-3">
                <p className="text-small text-danger font-medium">{errorMessage}</p>
              </div>
            )}
            <div className="flex flex-wrap gap-2">
              <Button variant="primary" size="sm" className="gap-1.5">
                <RotateCcw className="h-3.5 w-3.5" />
                Retry
              </Button>
              <Button variant="secondary" size="sm" className="gap-1.5">
                <GitCompareArrows className="h-3.5 w-3.5" />
                Compare Previous
              </Button>
              <Button variant="secondary" size="sm" className="gap-1.5">
                <Terminal className="h-3.5 w-3.5" />
                View Logs
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Failed steps timeline */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-body font-semibold">Deployment Steps</CardTitle>
          </CardHeader>
          <CardContent>
            <StepTimeline steps={steps} />
          </CardContent>
        </Card>
      </div>
    );
  }

  // ── Deploying state ──
  if (isDeploying) {
    return (
      <div className="space-y-5">
        {/* Progress bar */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-body font-semibold flex items-center gap-2">
              <Loader2 className="h-4 w-4 text-accent animate-spin" />
              Deploying...
              {duration && (
                <span className="text-small text-text-muted font-normal ml-auto">{duration}</span>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {/* Progress bar */}
            <div className="relative h-2 w-full overflow-hidden rounded-full bg-surface-elevated">
              <motion.div
                className="absolute inset-y-0 left-0 bg-accent rounded-full"
                initial={{ width: 0 }}
                animate={{ width: `${Math.min(progress, 100)}%` }}
                transition={{ duration: 0.3, ease: [0, 0, 0.2, 1] }}
              />
            </div>

            {/* Current step */}
            <div className="flex items-center gap-2 text-small">
              <Loader2 className="h-3.5 w-3.5 text-accent animate-spin shrink-0" />
              <span className="text-foreground font-medium">
                {liveState.currentNode
                  ? DEFAULT_STEPS.find((s) => s.id === liveState.currentNode)?.name ?? liveState.currentNode
                  : 'Starting deployment...'}
              </span>
              {liveState.totalNodes > 0 && (
                <span className="text-text-muted">
                  ({steps.filter((s) => s.status === 'completed').length}/{liveState.totalNodes} steps)
                </span>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Live step timeline */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-body font-semibold">Deployment Steps</CardTitle>
          </CardHeader>
          <CardContent>
            <StepTimeline steps={steps} />
          </CardContent>
        </Card>
      </div>
    );
  }

  // ── Default: no active deployment ──
  return null;
}

/* ── Step Timeline ─────────────────────────────────────── */
function StepTimeline({ steps }: { steps: Array<{ id: string; name: string; status: string }> }) {
  return (
    <div className="space-y-1">
      {steps.map((step) => (
        <div
          key={step.id}
          className={cn(
            'flex items-center gap-3 rounded-md px-3 py-2 transition-colors',
            step.status === 'running' && 'bg-accent/5',
            step.status === 'failed' && 'bg-danger/5',
          )}
        >
          <StepIcon status={step.status} />
          <span className={cn(
            'text-small flex-1',
            step.status === 'completed' && 'text-foreground',
            step.status === 'running' && 'text-accent font-medium',
            step.status === 'failed' && 'text-danger font-medium',
            step.status === 'pending' && 'text-text-muted',
          )}>
            {step.name}
          </span>
          {step.status === 'running' && (
            <span className="text-caption text-accent animate-pulse">In progress...</span>
          )}
          {step.status === 'completed' && (
            <CheckCircle2 className="h-3.5 w-3.5 text-success shrink-0" />
          )}
          {step.status === 'failed' && (
            <XCircle className="h-3.5 w-3.5 text-danger shrink-0" />
          )}
        </div>
      ))}
    </div>
  );
}
