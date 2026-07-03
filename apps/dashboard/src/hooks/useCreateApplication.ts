import { useMutation, useQueryClient } from '@tanstack/react-query';
import { cloudos } from '@/lib/sdk';
import type { AppResource } from './useApplications';

export interface CreateApplicationInput {
  id: string;
  name: string;
  projectId?: string;
  source: {
    url: string;
    branch?: string;
    path?: string;
  };
  runtime: {
    type: string;
    command?: string;
    port?: number;
  };
  build?: {
    command?: string;
    outputDir?: string;
    installCmd?: string;
  };
  environment?: Record<string, string>;
}

export function useCreateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateApplicationInput) => {
      const result = await cloudos.createApplication(data);
      return result as unknown as AppResource;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['applications'] });
    },
  });
}
