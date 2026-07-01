import { useQuery } from '@tanstack/react-query';
import { cloudos } from '@/lib/sdk';

export interface TimelineStep {
  id: string;
  name: string;
  action: string;
  status: string;
  result?: string;
  error?: string;
}

export interface TimelineResponse {
  application: string;
  deploymentNumber: number;
  workflowId: string;
  overallStatus: string;
  startedAt?: string;
  completedAt?: string;
  duration?: string;
  steps: TimelineStep[];
}

export interface DeploymentSummary {
  deploymentNumber: number;
  startedAt?: string;
  completedAt?: string;
  duration?: string;
  repository?: string;
  branch?: string;
  commitSha?: string;
  detectedRuntime?: string;
  buildpack?: string;
  buildSuccess: boolean;
  runtimeName?: string;
  environment?: string;
  artifactType?: string;
  healthStatus: string;
  endpoint?: string;
  workflowSteps: number;
  errors: string[];
}

export interface NodeComparison {
  id: string;
  name: string;
  action: string;
  fromStatus: string;
  toStatus: string;
  fromResult?: string;
  toResult?: string;
  fromError?: string;
  toError?: string;
  changed: boolean;
}

export interface ComparisonResponse {
  from: DeploymentSummary;
  to: DeploymentSummary;
  nodeComparison: NodeComparison[];
  summary: {
    statusChanged: boolean;
    healthChanged: boolean;
    durationChanged: boolean;
    durationDiff?: string;
    commitChanged: boolean;
    buildChanged: boolean;
    totalStepsMatch: boolean;
    changedNodeCount: number;
  };
}

const BASE = import.meta.env.VITE_CLOUDOS_API_URL ?? '';

async function directFetch<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { Accept: 'application/json' },
  });
  const body = await res.json();
  if (!body.success) throw new Error(body.error?.message ?? 'API error');
  return body.data as T;
}

export function useTimeline(appId: string, deploymentNumber?: number) {
  return useQuery({
    queryKey: ['timeline', appId, deploymentNumber],
    queryFn: async () => {
      if (deploymentNumber) {
        return directFetch<TimelineResponse>(
          `/api/v1/applications/${encodeURIComponent(appId)}/deployments/${deploymentNumber}/timeline`,
        );
      }
      // Get latest deployment number from application
      const app = await cloudos.getResource('Application', appId);
      const status = (app as unknown as { status: { deploymentCount?: number } }).status;
      const latest = status?.deploymentCount ?? 0;
      if (latest < 1) throw new Error('No deployments found');
      return directFetch<TimelineResponse>(
        `/api/v1/applications/${encodeURIComponent(appId)}/deployments/${latest}/timeline`,
      );
    },
    enabled: !!appId,
    staleTime: 5_000,
  });
}

export function useCompare(appId: string, fromNum: number, toNum: number) {
  return useQuery({
    queryKey: ['compare', appId, fromNum, toNum],
    queryFn: () =>
      directFetch<ComparisonResponse>(
        `/api/v1/applications/${encodeURIComponent(appId)}/deployments/compare?from=${fromNum}&to=${toNum}`,
      ),
    enabled: !!appId && fromNum > 0 && toNum > 0,
    staleTime: 10_000,
  });
}
