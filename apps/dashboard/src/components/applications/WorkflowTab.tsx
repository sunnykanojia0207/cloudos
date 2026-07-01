import * as React from 'react';
import type { AppResource } from '@/hooks/useApplications';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { EmptyState } from '@/components/ui/empty-state';
import { Terminal, TerminalLine } from '@/components/ui/terminal';
import { GraphNode, DEFAULT_NODES, type GraphNodeStatus } from './workflow-node';
import {
  Heart,
  Workflow,
  GitCommitHorizontal,
  Layers,
  Clock,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  RotateCcw,
} from 'lucide-react';

/* ── Props ────────────────────────────────────────────── */
export interface WorkflowTabProps {
  app: AppResource;
}

/* ── Node graph data (definitions from workflow-node) ──── */
// GraphNode and DEFAULT_NODES imported from './workflow-node'

/* ── Component ────────────────────────────────────────── */
export function WorkflowTab({ app }: WorkflowTabProps) {
  const report = app.status?.lastReport;
  const workflowId = report?.workflowId;
  const hasData = app.status?.deploymentCount != null && app.status.deploymentCount > 0;

  if (!hasData) {
    return (
      <EmptyState
        icon={Heart}
        title="No workflow data"
        description="Deploy the application to see its workflow execution."
      />
    );
  }

  // Derive node statuses from the last report (simplified — real data would come from a workflow execution API)
  // We show the node graph with pending/success based on whether the deployment succeeded
  const buildSuccess = report?.buildSuccess ?? false;
  const hasErrors = (report?.errors?.length ?? 0) > 0;
  const hasWarnings = (report?.warnings?.length ?? 0) > 0;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
      {/* ── Node Graph ── */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <Workflow className="h-4 w-4 text-text-muted" />
            Workflow Execution
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-0">
            {DEFAULT_NODES.map((node, index) => {
              let status: GraphNodeStatus = 'pending';
              if (node.key === 'complete') {
                status = report ? (buildSuccess ? 'success' : 'failed') : 'pending';
              } else if (node.key === 'health') {
                status = report ? (buildSuccess && !hasErrors ? 'success' : 'failed') : 'pending';
              } else if (node.key === 'build') {
                status = report ? (buildSuccess ? 'success' : 'failed') : 'pending';
              } else if (['validate', 'clone', 'detect', 'install'].includes(node.key)) {
                status = report ? 'success' : 'pending';
              }
              return (
                <GraphNode
                  key={node.key}
                  label={node.label}
                  status={status}
                  isLast={index === DEFAULT_NODES.length - 1}
                />
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* ── Execution Details ── */}
      <div className="space-y-5">
        {/* Summary */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-body font-semibold flex items-center gap-2">
              <Layers className="h-4 w-4 text-text-muted" />
              Execution Summary
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {/* Workflow ID */}
            {workflowId && (
              <div className="flex items-center justify-between">
                <span className="text-small text-text-secondary">Workflow ID</span>
                <span className="text-small font-mono text-foreground">{workflowId}</span>
              </div>
            )}

            {/* Duration */}
            {report?.duration && (
              <div className="flex items-center justify-between">
                <span className="text-small text-text-secondary">Duration</span>
                <span className="text-small text-foreground font-medium tabular-nums">{report.duration}</span>
              </div>
            )}

            {/* Steps */}
            <div className="flex items-center justify-between">
              <span className="text-small text-text-secondary">Steps</span>
              <span className="text-small text-foreground font-medium">{report?.workflowSteps ?? '\u2014'}</span>
            </div>

            {/* Retry */}
            <div className="flex items-center justify-between">
              <span className="text-small text-text-secondary">Retry Count</span>
              <span className="text-small text-foreground font-medium tabular-nums">0</span>
            </div>

            {/* Status */}
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
                Output
              </CardTitle>
            </CardHeader>
            <CardContent>
              {report?.errors && report.errors.length > 0 && (
                <div className="mb-3">
                  <span className="text-caption font-medium text-danger mb-1 block">Errors</span>
                  <ul className="space-y-1">
                    {report.errors.map((err, i) => (
                      <li key={i} className="text-small text-danger flex items-start gap-1.5">
                        <XCircle className="h-3.5 w-3.5 shrink-0 mt-0.5" />
                        <span>{err}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
              {report?.warnings && report.warnings.length > 0 && (
                <div>
                  <span className="text-caption font-medium text-warning mb-1 block">Warnings</span>
                  <ul className="space-y-1">
                    {report.warnings.map((warn, i) => (
                      <li key={i} className="text-small text-warning flex items-start gap-1.5">
                        <AlertTriangle className="h-3.5 w-3.5 shrink-0 mt-0.5" />
                        <span>{warn}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
