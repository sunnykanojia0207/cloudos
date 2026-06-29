import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { cloudos } from '@/lib/sdk';

// ── System ────────────────────────────────────────────────────────────────

export function useHealth() {
  return useQuery({
    queryKey: ['health'],
    queryFn: () => cloudos.getHealth(),
    refetchInterval: 30_000,
  });
}

export function useVersion() {
  return useQuery({
    queryKey: ['version'],
    queryFn: () => cloudos.getVersion(),
    staleTime: 5 * 60 * 1000,
  });
}

export function useKernel() {
  return useQuery({
    queryKey: ['kernel'],
    queryFn: () => cloudos.getKernel(),
    refetchInterval: 10_000,
  });
}

export function useSystem() {
  return useQuery({
    queryKey: ['system'],
    queryFn: () => cloudos.getSystem(),
    staleTime: 5 * 60 * 1000,
  });
}

export function useReady() {
  return useQuery({
    queryKey: ['ready'],
    queryFn: () => cloudos.getReady(),
    refetchInterval: 10_000,
  });
}

export function useLive() {
  return useQuery({
    queryKey: ['live'],
    queryFn: () => cloudos.getLive(),
    refetchInterval: 10_000,
  });
}

// ── Capabilities ──────────────────────────────────────────────────────────

export function useCapabilities() {
  return useQuery({
    queryKey: ['capabilities'],
    queryFn: () => cloudos.getCapabilities(),
    staleTime: 30_000,
  });
}

export function useCapability(id: string) {
  return useQuery({
    queryKey: ['capabilities', id],
    queryFn: () => cloudos.getCapability(id),
    enabled: !!id,
  });
}

// ── Providers ─────────────────────────────────────────────────────────────

export function useProviders() {
  return useQuery({
    queryKey: ['providers'],
    queryFn: () => cloudos.getProviders(),
    staleTime: 30_000,
  });
}

export function useProvider(id: string) {
  return useQuery({
    queryKey: ['providers', id],
    queryFn: () => cloudos.getProvider(id),
    enabled: !!id,
  });
}

export function useProviderHealth(id: string) {
  return useQuery({
    queryKey: ['providers', id, 'health'],
    queryFn: () => cloudos.getProviderHealth(id),
    enabled: !!id,
    refetchInterval: 30_000,
  });
}

// ── Resources (Resource Engine) ─────────────────────────────────────────

export function useResourceKinds() {
  return useQuery({
    queryKey: ['resourceKinds'],
    queryFn: () => cloudos.getResourceKinds(),
    staleTime: 5 * 60 * 1000,
  });
}

export function useResources(kind: string) {
  return useQuery({
    queryKey: ['resources', kind],
    queryFn: () => cloudos.getResources(kind),
    enabled: !!kind,
    staleTime: 30_000,
  });
}

export function useResource(kind: string, id: string) {
  return useQuery({
    queryKey: ['resources', kind, id],
    queryFn: () => cloudos.getResource(kind, id),
    enabled: !!kind && !!id,
  });
}

// ── Controllers (Controller Runtime) ────────────────────────────────────

export function useControllers() {
  return useQuery({
    queryKey: ['controllers'],
    queryFn: () => cloudos.getControllers(),
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useController(id: string) {
  return useQuery({
    queryKey: ['controllers', id],
    queryFn: () => cloudos.getController(id),
    enabled: !!id,
    staleTime: 10_000,
  });
}

export function useControllerHealth(id: string) {
  return useQuery({
    queryKey: ['controllers', id, 'health'],
    queryFn: () => cloudos.getControllerHealth(id),
    enabled: !!id,
    refetchInterval: 15_000,
  });
}

// ── Projects ─────────────────────────────────────────────────────────────

export function useProjects() {
  return useQuery({
    queryKey: ['projects'],
    queryFn: () => cloudos.getProjects(),
    staleTime: 10_000,
    refetchInterval: 15_000,
  });
}

export function useProject(id: string) {
  return useQuery({
    queryKey: ['projects', id],
    queryFn: () => cloudos.getProject(id),
    enabled: !!id,
    staleTime: 10_000,
    refetchInterval: 15_000,
  });
}

export function useCreateProject() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { id: string; displayName: string; description?: string; environment?: string }) =>
      cloudos.createProject(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
}

export function useUpdateProject() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<{ displayName: string; description: string; environment: string }> }) =>
      cloudos.updateProject(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
}

export function useDeleteProject() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => cloudos.deleteProject(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
}
