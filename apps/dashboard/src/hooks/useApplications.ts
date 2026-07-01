import { useQuery } from '@tanstack/react-query';
import { cloudos } from '@/lib/sdk';

// Types for Application resources from the API
export interface AppSource {
  type: string;
  url?: string;
  branch?: string;
  path?: string;
}

export interface AppRuntime {
  type: string;
  command?: string;
  port?: number;
}

export interface AppSpec {
  source: AppSource;
  runtime: AppRuntime;
  deployment?: { port?: number };
  settings?: Record<string, string>;
}

export interface DeploymentReport {
  deploymentNumber: number;
  startedAt: string;
  completedAt: string;
  duration: string;
  repository: string;
  branch: string;
  commitSha?: string;
  detectedRuntime?: string;
  buildpack?: string;
  buildSuccess: boolean;
  runtimeName?: string;
  runtimeVersion?: string;
  environment?: string;
  artifactType?: string;
  healthStatus: string;
  endpoint: string;
  workflowId?: string;
  workflowSteps: number;
  warnings: string[];
  errors: string[];
}

export interface AppStatus {
  phase: string;
  health: string;
  url?: string;
  deploymentCount: number;
  currentDeploymentId?: string;
  lastReport?: DeploymentReport;
  deploymentHistory?: DeploymentReport[];
}

export interface AppResource {
  apiVersion: string;
  kind: string;
  metadata: { id: string; name: string; labels?: Record<string, string>; createdAt?: string };
  spec: AppSpec;
  status: AppStatus;
}

// Hooks

export function useApplications() {
  return useQuery({
    queryKey: ['applications'],
    queryFn: async () => {
      const result = await cloudos.getResources('Application');
      return (result?.items ?? []) as unknown as AppResource[];
    },
    refetchInterval: 10_000,
  });
}

export function useApplication(id: string) {
  return useQuery({
    queryKey: ['applications', id],
    queryFn: async () => {
      const result = await cloudos.getResource('Application', id);
      return result as unknown as AppResource;
    },
    enabled: !!id,
    refetchInterval: 10_000,
  });
}
