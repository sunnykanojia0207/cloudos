import * as React from 'react';
import type { AppResource } from '@/hooks/useApplications';
import { DeployProgress } from '@/components/applications/DeployProgress';
import { cn, truncate, relativeTime } from '@/lib/utils';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { HealthIndicator } from '@/components/ui/health-indicator';
import { Separator } from '@/components/ui/separator';
import { Button } from '@/components/ui/button';
import {
  GitBranch,
  GitCommitHorizontal,
  Terminal,
  Package,
  Globe,
  Box,
  Clock,
  Calendar,
  Copy,
  ExternalLink,
  Workflow,
  Layers,
  CheckCircle2,
  BookOpen,
  Cloud,
  Cpu,
} from 'lucide-react';

/* ── Copy Button ──────────────────────────────────────── */
function CopyButton({ text, label }: { text: string; label?: string }) {
  const [copied, setCopied] = React.useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(text).catch(() => {});
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <button
      type="button"
      onClick={handleCopy}
      className={cn(
        'ml-1.5 inline-flex items-center justify-center rounded p-0.5',
        'text-text-muted hover:text-text-secondary transition-colors',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
      )}
      aria-label={label ?? 'Copy to clipboard'}
    >
      {copied ? (
        <CheckCircle2 className="h-3.5 w-3.5 text-success" />
      ) : (
        <Copy className="h-3.5 w-3.5" />
      )}
    </button>
  );
}

/* ── Detail Row ───────────────────────────────────────── */
interface DetailRowProps {
  icon: React.ReactNode;
  label: string;
  value: React.ReactNode;
  copyValue?: string;
}

function DetailRow({ icon, label, value, copyValue }: DetailRowProps) {
  return (
    <div className="flex items-start gap-3 py-1.5">
      <span className="mt-0.5 shrink-0 text-text-muted">{icon}</span>
      <div className="min-w-0 flex-1">
        <span className="block text-caption text-text-secondary">{label}</span>
        <span className="mt-0.5 block text-small font-medium text-foreground break-all">
          {value ?? '\u2014'}
          {copyValue && <CopyButton text={copyValue} />}
        </span>
      </div>
    </div>
  );
}

/* ── Status Summary Row ───────────────────────────────── */
interface StatusRowProps {
  label: string;
  children: React.ReactNode;
}

function StatusRow({ label, children }: StatusRowProps) {
  return (
    <div className="flex items-center justify-between rounded-md border border-border bg-surface px-3 py-2.5">
      <span className="text-small text-text-secondary">{label}</span>
      <span className="flex items-center gap-1.5">{children}</span>
    </div>
  );
}

/* ── Format helper ────────────────────────────────────── */
function formatDateTime(dateStr?: string): string {
  if (!dateStr) return '\u2014';
  try {
    return new Date(dateStr).toLocaleString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return dateStr;
  }
}

/* ── Props ────────────────────────────────────────────── */
export interface OverviewTabProps {
  app: AppResource;
}

/* ── Component ────────────────────────────────────────── */
export function OverviewTab({ app }: OverviewTabProps) {
  const report = app.status?.lastReport;
  const phase = app.status?.phase ?? 'Stopped';
  const url = app.status?.url;
  const deploymentCount = app.status?.deploymentCount ?? 0;
  const env = app.spec?.settings?.environment ?? report?.environment;

  return (
    <div className="space-y-5">
      {/* Live deployment progress / status — shown for active deployments */}
      <DeployProgress app={app} />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
      {/* ──── LEFT COLUMN ──── */}

      {/* Deployment Report */}
      <div className="space-y-5">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-body font-semibold flex items-center gap-2">
              <BookOpen className="h-4 w-4 text-text-muted" />
              Deployment Report
            </CardTitle>
          </CardHeader>
          <CardContent>
            {report ? (
              <div className="space-y-0.5">
                <DetailRow
                  icon={<GitBranch className="h-4 w-4" />}
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
                  copyValue={report.commitSha}
                />

                <Separator className="my-2" />

                <DetailRow
                  icon={<Terminal className="h-4 w-4" />}
                  label="Detected Runtime"
                  value={report.detectedRuntime || '\u2014'}
                />
                <DetailRow
                  icon={<Package className="h-4 w-4" />}
                  label="Buildpack"
                  value={report.buildpack || '\u2014'}
                />
                <DetailRow
                  icon={<Terminal className="h-4 w-4" />}
                  label="Runtime Version"
                  value={report.runtimeVersion || '\u2014'}
                />

                <Separator className="my-2" />

                <DetailRow
                  icon={<Globe className="h-4 w-4" />}
                  label="Environment"
                  value={report.environment || '\u2014'}
                />
                <DetailRow
                  icon={<Box className="h-4 w-4" />}
                  label="Artifact Type"
                  value={report.artifactType || '\u2014'}
                />

                <Separator className="my-2" />

                <DetailRow
                  icon={<Clock className="h-4 w-4" />}
                  label="Duration"
                  value={report.duration || '\u2014'}
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
              </div>
            ) : (
              <p className="text-small text-text-muted py-4 text-center">
                No deployment report available.
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* ──── RIGHT COLUMN ──── */}

      {/* Status Summary */}
      <div className="space-y-5">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-body font-semibold flex items-center gap-2">
              <Cloud className="h-4 w-4 text-text-muted" />
              Status Summary
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {/* Phase */}
            <StatusRow label="Phase">
              <Badge variant={
                phase === 'Running' ? 'subtle-success' :
                phase === 'Deploying' ? 'subtle-accent' :
                phase === 'Failed' ? 'subtle-danger' :
                'subtle-neutral'
              } className={cn('gap-1.5', phase === 'Deploying' && 'animate-pulse')}>
                {phase === 'Deploying' && (
                  <span className="relative flex h-1.5 w-1.5">
                    <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-current opacity-75" />
                    <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-current" />
                  </span>
                )}
                {phase}
              </Badge>
            </StatusRow>

            {/* Health */}
            <StatusRow label="Health">
              <HealthIndicator
                status={
                  phase === 'Running' ? 'running' :
                  phase === 'Deploying' ? 'deploying' :
                  phase === 'Failed' ? 'failed' :
                  'stopped'
                }
                showLabel
                size="sm"
              />
            </StatusRow>

            {/* URL */}
            {url && (
              <StatusRow label="URL">
                <Button variant="secondary" size="sm" className="h-7 gap-1 text-small" asChild>
                  <a href={url} target="_blank" rel="noopener noreferrer">
                    <ExternalLink className="h-3.5 w-3.5" />
                    Open
                  </a>
                </Button>
              </StatusRow>
            )}

            {/* Deployments */}
            <StatusRow label="Deployments">
              <span className="text-small font-semibold text-foreground tabular-nums">
                {deploymentCount}
              </span>
            </StatusRow>

            {/* Environment */}
            {env && (
              <StatusRow label="Environment">
                <Badge variant="subtle-neutral" className="text-caption uppercase tracking-wider">
                  {env}
                </Badge>
              </StatusRow>
            )}
          </CardContent>
        </Card>

        {/* Workflow Summary */}
        {report?.workflowId && (
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-body font-semibold flex items-center gap-2">
                <Workflow className="h-4 w-4 text-text-muted" />
                Workflow
              </CardTitle>
            </CardHeader>
            <CardContent>
              <DetailRow
                icon={<Layers className="h-4 w-4" />}
                label="Workflow ID"
                value={report.workflowId}
              />
              <DetailRow
                icon={<Cpu className="h-4 w-4" />}
                label="Steps"
                value={String(report.workflowSteps ?? '\u2014')}
              />
              {report.warnings && report.warnings.length > 0 && (
                <DetailRow
                  icon={<Box className="h-4 w-4" />}
                  label="Warnings"
                  value={report.warnings.length}
                />
              )}
            </CardContent>
          </Card>
        )}

        {/* Resource Info */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-body font-semibold flex items-center gap-2">
              <Box className="h-4 w-4 text-text-muted" />
              Resource Info
            </CardTitle>
          </CardHeader>
          <CardContent>
            <DetailRow
              icon={<BookOpen className="h-4 w-4" />}
              label="API Version"
              value={app.apiVersion || '\u2014'}
            />
            <DetailRow
              icon={<Box className="h-4 w-4" />}
              label="Kind"
              value={app.kind || '\u2014'}
            />
            <DetailRow
              icon={<Terminal className="h-4 w-4" />}
              label="Runtime"
              value={app.spec?.runtime?.type || '\u2014'}
            />
            <DetailRow
              icon={<GitBranch className="h-4 w-4" />}
              label="Source Type"
              value={app.spec?.source?.type || '\u2014'}
            />
          </CardContent>
        </Card>
      </div>
    </div>
    </div>
  );
}
