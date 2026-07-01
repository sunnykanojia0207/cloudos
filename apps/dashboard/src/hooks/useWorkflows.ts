import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { cloudos } from '@/lib/sdk';
import { useApplications, type DeploymentReport } from './useApplications';
import type { TimelineResponse } from './useDeployments';

/* ── Types ──────────────────────────────────────────────── */

export type WorkflowStatus = 'succeeded' | 'failed' | 'running' | 'pending' | 'cancelled';

export interface WorkflowExecution {
  /** The workflow execution ID (from DeploymentReport.workflowId) */
  id: string;
  appId: string;
  appName: string;
  deploymentNumber: number;
  status: WorkflowStatus;
  startedAt: string;
  completedAt?: string;
  duration: string;
  nodeCount: number;
  completedNodes: number;
  failedNodes: number;
  queueTime?: string;
  branch?: string;
  commitSha?: string;
  runtime?: string;
  environment?: string;
}

export interface WorkflowDetail extends WorkflowExecution {
  timeline: TimelineResponse;
}

/* ── Helpers ────────────────────────────────────────────── */

function deriveStatus(report: DeploymentReport): WorkflowStatus {
  if (report.buildSuccess) return 'succeeded';
  // If not completed yet, check timestamps
  if (!report.completedAt && report.startedAt) return 'running';
  return 'failed';
}

/* ── Aggregate all workflow executions ──────────────────── */
export function useWorkflows() {
  const { data: applications, isLoading, error, refetch } = useApplications();

  const workflows = useMemo<WorkflowExecution[]>(() => {
    if (!applications) return [];

    const entries: WorkflowExecution[] = [];

    for (const app of applications) {
      const history = app.status?.deploymentHistory ?? [];
      for (const report of history) {
        // Skip entries without a workflowId
        if (!report.workflowId) continue;

        const status = deriveStatus(report);
        const completedNodes = status === 'failed'
          ? (report.workflowSteps ?? 0) - (report.errors?.length ?? 0)
          : status === 'succeeded'
            ? (report.workflowSteps ?? 0)
            : 0;
        const failedNodes = status === 'failed' ? (report.errors?.length ?? 0) : 0;

        entries.push({
          id: report.workflowId,
          appId: app.metadata.id,
          appName: app.metadata.name,
          deploymentNumber: report.deploymentNumber,
          status,
          startedAt: report.startedAt || app.metadata.createdAt || '',
          completedAt: report.completedAt || undefined,
          duration: report.duration,
          nodeCount: report.workflowSteps,
          completedNodes,
          failedNodes,
          branch: report.branch,
          commitSha: report.commitSha,
          runtime: report.detectedRuntime,
          environment: report.environment,
        });
      }
    }

    // Sort newest first
    entries.sort((a, b) => b.startedAt.localeCompare(a.startedAt));

    return entries;
  }, [applications]);

  // ── Stats ──
  const stats = useMemo(() => {
    const total = workflows.length;
    const running = workflows.filter((w) => w.status === 'running').length;
    const succeeded = workflows.filter((w) => w.status === 'succeeded').length;
    const failed = workflows.filter((w) => w.status === 'failed').length;
    return { total, running, succeeded, failed };
  }, [workflows]);

  return { workflows, stats, isLoading, error, refetch };
}

/* ── Single workflow detail ─────────────────────────────── */
export function useWorkflow(workflowId: string) {
  const { workflows, isLoading: loadingApps } = useWorkflows();

  return useQuery({
    queryKey: ['workflow', workflowId],
    queryFn: async () => {
      // Find which app + deployment this workflow belongs to
      const match = workflows.find((w) => w.id === workflowId);
      if (!match) throw new Error(`Workflow ${workflowId} not found`);

      // Fetch the timeline for this deployment
      const BASE = import.meta.env.VITE_CLOUDOS_API_URL ?? '';
      const url = `${BASE}/api/v1/applications/${encodeURIComponent(match.appId)}/deployments/${match.deploymentNumber}/timeline`;
      const res = await fetch(url, { headers: { Accept: 'application/json' } });
      const body = await res.json();
      if (!body.success) throw new Error(body.error?.message ?? 'Failed to load timeline');

      const timeline = body.data as TimelineResponse;

      const detail: WorkflowDetail = {
        ...match,
        timeline,
      };

      return detail;
    },
    enabled: !!workflowId && !loadingApps,
    staleTime: 5_000,
  });
}
