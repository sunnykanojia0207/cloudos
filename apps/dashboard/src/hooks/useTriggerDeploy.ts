import { useMutation, useQueryClient } from '@tanstack/react-query';
import { cloudos } from '@/lib/sdk';
import type { AppResource } from './useApplications';

export function useTriggerDeploy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (appId: string) => {
      const result = await cloudos.triggerDeploy(appId);
      return result as unknown as AppResource;
    },
    onSuccess: (_data, appId) => {
      // Invalidate both the app detail and the list
      queryClient.invalidateQueries({ queryKey: ['applications', appId] });
      queryClient.invalidateQueries({ queryKey: ['applications'] });
    },
  });
}
