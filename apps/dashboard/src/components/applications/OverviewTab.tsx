import * as React from 'react';
import type { AppResource } from '@/hooks/useApplications';
import { cn, truncate } from '@/lib/utils';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Button } from '@/components/ui/button';
import {
  GitBranch,
  GitCommitHorizontal,
  Terminal,
  Package,
  Layers,
  Globe,
  Box,
  Clock,
  Calendar,
  Copy,
  ExternalLink,
  Workflow,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  BookOpen,
} from 'lucide-react';

/* ── Helpers ──────────────────────────────────────────── */

const healthBadge: Record<
  string,
  { variant: 'success' | 'warning' | 'destructive'; label: string }
> = {
  Healthy: { variant: 'success', label: 'Healthy' },
  Degraded: { variant: 'warning', label: 'Degraded' },
  Error: { variant: 'destructive', label: 'Error' },
};

const phaseBadge: Record<
  string,
  { variant: 'success' | 'warning' | 'secondary' | 'destructive'; label: string }
> = {
  Running: { variant: 'success', label: 'Running' },
  Deploying: { variant: 'warning', label: 'Deploying' },
  Stopped: { variant: 'secondary', label: 'Stopped' },
  Failed: { variant: 'destructive', label: 'Failed' },
};

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).catch(() => {
    // Clipboard write failed — silently ignore
  });
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

/* ── Props ────────────────────────────────────────────── */

export interface OverviewTabProps {
  app: AppResource;
}

/* ── Copy Button ──────────────────────────────────────── */

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = React.useState(false);

  const handleCopy = () => {
    copyToClipboard(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <button
      type="button"
      onClick={handleCopy}
      className={cn(
        'ml-1.5 inline-flex items-center justify-center rounded p-0.5',
        'text-muted-foreground/50 hover:text-muted-foreground transition-colors',
      )}
      aria-label="Copy to clipboard"
    >
      {copied ? (
        <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />
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
    <div className="flex items-start gap-3 py-2">
      <span className="mt-0.5 shrink-0 text-muted-foreground">{icon}</span>
      <div className="min-w-0 flex-1">
        <span className="block text-xs text-muted-foreground">{label}</span>
        <span className="mt-0.5 block text-sm font-medium text-foreground/90 break-all">
          {value}
          {copyValue && <CopyButton text={copyValue} />}
        </span>
      </div>
    </div>
  );
}

/* ── Component ────────────────────────────────────────── */

export function OverviewTab({ app }: OverviewTabProps) {
  const report = app.status?.lastReport;
  const phase = app.status?.phase ?? 'Stopped';
  const health = app.status?.health ?? 'Unknown';
  const url = app.status?.url;
  const deploymentCount = app.status?.deploymentCount ?? 0;

  const phaseConf = phaseBadge[phase] ?? phaseBadge.Stopped;
  const healthConf = healthBadge[health] ?? { variant: 'destructive' as const, label: health };

  const healthIcon = health === 'Healthy'
    ? <CheckCircle2 className="h-4 w-4 text-emerald-400" />
    : health === 'Degraded'
      ? <AlertTriangle className="h-4 w-4 text-amber-400" />
      : <XCircle className="h-4 w-4 text-red-400" />;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* ──── Deployment Report ──── */}
      <Card className="border-border/50">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            <BookOpen className="h-4 w-4 text-muted-foreground" />
            Deployment Report
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-1">
          {report ? (
            <>
              {/* Repository & Branch */}
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

              {/* Runtime */}
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
                label="Runtime Name"
                value={report.runtimeName || '\u2014'}
              />
              <DetailRow
                icon={<Terminal className="h-4 w-4" />}
                label="Runtime Version"
                value={report.runtimeVersion || '\u2014'}
              />

              <Separator className="my-2" />

              {/* Environment & Artifact */}
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

              {/* Workflow */}
              <DetailRow
                icon={<Workflow className="h-4 w-4" />}
                label="Workflow ID"
                value={report.workflowId || '\u2014'}
              />
              <DetailRow
                icon={<Layers className="h-4 w-4" />}
                label="Workflow Steps"
                value={String(report.workflowSteps ?? '\u2014')}
              />

              <Separator className="my-2" />

              {/* Timing */}
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
            </>
          ) : (
            <p className="text-sm text-muted-foreground py-4 text-center">
              No deployment report available.
            </p>
          )}
        </CardContent>
      </Card>

      {/* ──── Status Summary ──── */}
      <div className="space-y-6">
        {/* Phase & Health */}
        <Card className="border-border/50">
          <CardHeader className="pb-3">
            <CardTitle className="text-base font-semibold flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-muted-foreground" />
              Status Summary
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Phase */}
            <div className="flex items-center justify-between rounded-lg border border-border/50 bg-muted/30 px-4 py-3">
              <span className="text-sm font-medium text-foreground/80">Phase</span>
              <Badge variant={phaseConf.variant} className="gap-1.5">
                {phase === 'Deploying' && (
                  <span className="relative flex h-1.5 w-1.5">
                    <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-current opacity-75" />
                    <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-current" />
                  </span>
                )}
                {phaseConf.label}
              </Badge>
            </div>

            {/* Health */}
            <div className="flex items-center justify-between rounded-lg border border-border/50 bg-muted/30 px-4 py-3">
              <span className="text-sm font-medium text-foreground/80">Health</span>
              <Badge variant={healthConf.variant} className="gap-1.5">
                {healthIcon}
                {healthConf.label}
              </Badge>
            </div>

            {/* URL */}
            {url && (
              <div className="flex items-center justify-between rounded-lg border border-border/50 bg-muted/30 px-4 py-3">
                <span className="text-sm font-medium text-foreground/80">URL</span>
                <Button variant="outline" size="sm" className="h-8 gap-1.5 text-xs" asChild>
                  <a href={url} target="_blank" rel="noopener noreferrer">
                    <ExternalLink className="h-3.5 w-3.5" />
                    Open
                  </a>
                </Button>
              </div>
            )}

            {/* Deployment Count */}
            <div className="flex items-center justify-between rounded-lg border border-border/50 bg-muted/30 px-4 py-3">
              <span className="text-sm font-medium text-foreground/80">Deployments</span>
              <span className="text-sm font-semibold text-foreground/90">
                {deploymentCount}
              </span>
            </div>
          </CardContent>
        </Card>

        {/* Quick resource info */}
        <Card className="border-border/50">
          <CardHeader className="pb-3">
            <CardTitle className="text-base font-semibold flex items-center gap-2">
              <Box className="h-4 w-4 text-muted-foreground" />
              Resource Info
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-1">
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
              label="Runtime Type"
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
  );
}
