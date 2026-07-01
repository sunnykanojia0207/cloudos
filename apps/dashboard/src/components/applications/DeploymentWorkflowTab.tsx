import * as React from 'react';
import type { DeploymentReport } from '@/hooks/useApplications';
import type { TimelineResponse } from '@/hooks/useDeployments';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { EmptyState } from '@/components/ui/empty-state';
import { Separator } from '@/components/ui/separator';
import { GraphNode, DEFAULT_NODES, type GraphNodeStatus } from './workflow-node';
import {
  Heart,
  Layers,
  Clock,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  RotateCcw,
  Timer,
  Activity,
  GitCommitHorizontal,
} from 'lucide-react';

/* ── Graph Node and DEFAULT_NODES imported from ./workflow-node ── */

/* ── Props ────────────────────────────────────────────── */
export interface DeploymentWorkflowTabProps {
  report: DeploymentReport;
  timeline?: TimelineResponse;
}

/* ── Component ────────────────────────────────────────── */
export function DeploymentWorkflowTab({ report, timeline }: DeploymentWorkflowTabProps) {
  const buildSuccess = report.buildSuccess;
  const hasErrors = (report.errors?.length ?? 0) > 0;
  const hasWarnings = (report.warnings?.length ?? 0) > 0;
  const steps = timeline?.steps ?? [];

  // Derive node status from available data
  const getNodeStatus = (key: string): GraphNodeStatus => {
    // If timeline has this step, use its status
    if (steps.length > 0) {
      const step = steps.find((s) => s.action?.toLowerCase() === key || s.name?.toLowerCase().includes(key));
      if (step) {
        const lower = step.status?.toLowerCase() ?? '';
        if (['success', 'succeeded'].includes(lower)) return 'success';
        if (['failure', 'failed', 'error'].includes(lower)) return 'failed';
        if (lower === 'running') return 'running';
        if (lower === 'skipped') return 'skipped';
      }
    }

    // Fallback: derive from report
    if (key === 'complete') return buildSuccess ? 'success' : 'failed';
    if (key === 'health') return buildSuccess && !hasErrors ? 'success' : 'failed';
    if (key === 'build') return buildSuccess ? 'success' : 'failed';
    if (['validate', 'clone', 'detect', 'install'].includes(key)) return 'success';
    return 'pending';
  };

  return (
    <div className="space-y-5">
      {/* Node Graph */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <Activity className="h-4 w-4 text-text-muted" />
            Workflow Execution
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-0">
            {DEFAULT_NODES.map((node, index) => (
              <GraphNode
                key={node.key}
                label={node.label}
                status={getNodeStatus(node.key)}
                isLast={index === DEFAULT_NODES.length - 1}
              />
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Execution Details */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <Layers className="h-4 w-4 text-text-muted" />
            Execution Details
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-small text-text-secondary">Workflow ID</span>
            <span className="text-small font-mono text-foreground">{report.workflowId || '\u2014'}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-small text-text-secondary">Duration</span>
            <span className="text-small text-foreground font-medium tabular-nums">{report.duration || '\u2014'}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-small text-text-secondary">Steps</span>
            <span className="text-small text-foreground font-medium">{(report.workflowSteps ?? steps.length) || '\u2014'}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-small text-text-secondary">Retry Count</span>
            <span className="text-small text-foreground font-medium tabular-nums">0</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-small text-text-secondary">Queue Time</span>
            <span className="text-small text-foreground font-medium">—</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-small text-text-secondary">Status</span>
            <Badge variant={buildSuccess ? 'subtle-success' : 'subtle-danger'} className="gap-1">
              {buildSuccess ? <CheckCircle2 className="h-3.5 w-3.5" /> : <XCircle className="h-3.5 w-3.5" />}
              {buildSuccess ? 'Completed' : 'Failed'}
            </Badge>
          </div>
        </CardContent>
      </Card>

      {/* Errors & Warnings */}
      {(hasErrors || hasWarnings) && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-body font-semibold flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-text-muted" />
              Issues
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {report.errors?.map((err, i) => (
              <div key={i} className="flex items-start gap-2">
                <XCircle className="h-3.5 w-3.5 text-danger shrink-0 mt-0.5" />
                <span className="text-small text-danger">{err}</span>
              </div>
            ))}
            {report.warnings?.map((warn, i) => (
              <div key={i} className="flex items-start gap-2">
                <AlertTriangle className="h-3.5 w-3.5 text-warning shrink-0 mt-0.5" />
                <span className="text-small text-warning">{warn}</span>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
