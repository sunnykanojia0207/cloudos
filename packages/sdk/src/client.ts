import { CloudOSError } from './errors';
import type {
  ApiResponse,
  HealthResponse,
  ReadinessResponse,
  LivenessResponse,
  VersionResponse,
  KernelResponse,
  SystemResponse,
  ResourceList,
  ResourceObject,
  CapabilitySpec,
  CapabilityStatus,
  ProviderSpec,
  ProviderStatus,
  ProviderHealthResponse,
  ProviderCapabilitiesResponse,
  ResourceKindListResponse,
  ResourceSpec,
  ResourceStatus,
  ControllerListResponse,
  ControllerDTO,
  ControllerHealthDTO,
  ProjectDTO,
  ProjectListResponse,
  ApplicationDTO,
  CreateApplicationRequest,
} from './types';

export class CloudOSClient {
  private baseUrl: string;

  constructor(baseUrl = 'http://localhost:8080') {
    this.baseUrl = baseUrl.replace(/\/+$/, '');
  }

  private async request<T>(path: string, init?: RequestInit): Promise<T> {
    const url = `${this.baseUrl}${path}`;
    const response = await fetch(url, {
      ...init,
      headers: {
        Accept: 'application/json',
        ...init?.headers,
      },
    });

    let body: ApiResponse<T>;
    try {
      body = await response.json();
    } catch {
      throw new CloudOSError(
        'PARSE_ERROR',
        `Failed to parse response from ${path}`,
      );
    }

    if (!body.success || body.error) {
      throw new CloudOSError(
        body.error?.code || 'UNKNOWN',
        body.error?.message || `Request to ${path} failed`,
      );
    }

    return body.data as T;
  }

  // ── System ────────────────────────────────────────────────────────────

  getHealth(): Promise<HealthResponse> {
    return this.request<HealthResponse>('/api/v1/health');
  }

  getReady(): Promise<ReadinessResponse> {
    return this.request<ReadinessResponse>('/api/v1/ready');
  }

  getLive(): Promise<LivenessResponse> {
    return this.request<LivenessResponse>('/api/v1/live');
  }

  getVersion(): Promise<VersionResponse> {
    return this.request<VersionResponse>('/api/v1/version');
  }

  getKernel(): Promise<KernelResponse> {
    return this.request<KernelResponse>('/api/v1/kernel');
  }

  getSystem(): Promise<SystemResponse> {
    return this.request<SystemResponse>('/api/v1/system');
  }

  // ── Capabilities ──────────────────────────────────────────────────────

  getCapabilities(): Promise<ResourceList<CapabilitySpec, CapabilityStatus>> {
    return this.request<ResourceList<CapabilitySpec, CapabilityStatus>>('/api/v1/capabilities');
  }

  getCapability(id: string): Promise<ResourceObject<CapabilitySpec, CapabilityStatus>> {
    return this.request<ResourceObject<CapabilitySpec, CapabilityStatus>>(`/api/v1/capabilities/${encodeURIComponent(id)}`);
  }

  // ── Providers ─────────────────────────────────────────────────────────

  getProviders(): Promise<ResourceList<ProviderSpec, ProviderStatus>> {
    return this.request<ResourceList<ProviderSpec, ProviderStatus>>('/api/v1/providers');
  }

  getProvider(id: string): Promise<ResourceObject<ProviderSpec, ProviderStatus>> {
    return this.request<ResourceObject<ProviderSpec, ProviderStatus>>(`/api/v1/providers/${encodeURIComponent(id)}`);
  }

  getProviderHealth(id: string): Promise<ProviderHealthResponse> {
    return this.request<ProviderHealthResponse>(`/api/v1/providers/${encodeURIComponent(id)}/health`);
  }

  getProviderCapabilities(id: string): Promise<ProviderCapabilitiesResponse> {
    return this.request<ProviderCapabilitiesResponse>(`/api/v1/providers/${encodeURIComponent(id)}/capabilities`);
  }

  // ── Resources (Resource Engine) ─────────────────────────────────────────

  getResourceKinds(): Promise<ResourceKindListResponse> {
    return this.request<ResourceKindListResponse>('/api/v1/resources');
  }

  getResources(kind: string): Promise<ResourceList<ResourceSpec, ResourceStatus>> {
    return this.request<ResourceList<ResourceSpec, ResourceStatus>>(`/api/v1/resources/${encodeURIComponent(kind)}`);
  }

  getResource(kind: string, id: string): Promise<ResourceObject<ResourceSpec, ResourceStatus>> {
    return this.request<ResourceObject<ResourceSpec, ResourceStatus>>(
      `/api/v1/resources/${encodeURIComponent(kind)}/${encodeURIComponent(id)}`,
    );
  }

  // ── Controllers (Controller Runtime) ───────────────────────────────────

  getControllers(): Promise<ControllerListResponse> {
    return this.request<ControllerListResponse>('/api/v1/controllers');
  }

  getController(id: string): Promise<ControllerDTO> {
    return this.request<ControllerDTO>(`/api/v1/controllers/${encodeURIComponent(id)}`);
  }

  getControllerHealth(id: string): Promise<ControllerHealthDTO> {
    return this.request<ControllerHealthDTO>(`/api/v1/controllers/${encodeURIComponent(id)}/health`);
  }

  // ── Projects ──────────────────────────────────────────────────────────

  getProjects(): Promise<ProjectListResponse> {
    return this.request<ProjectListResponse>('/api/v1/projects');
  }

  getProject(id: string): Promise<ProjectDTO> {
    return this.request<ProjectDTO>(`/api/v1/projects/${encodeURIComponent(id)}`);
  }

  createProject(data: {
    id: string;
    displayName: string;
    description?: string;
    environment?: string;
  }): Promise<ProjectDTO> {
    return this.request<ProjectDTO>('/api/v1/projects', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
  }

  updateProject(id: string, data: Partial<{
    displayName: string;
    description: string;
    environment: string;
  }>): Promise<ProjectDTO> {
    return this.request<ProjectDTO>(`/api/v1/projects/${encodeURIComponent(id)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
  }

  deleteProject(id: string): Promise<void> {
    return this.request<void>(`/api/v1/projects/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    });
  }

  // ── Applications ──────────────────────────────────────────────────────────

  createApplication(data: CreateApplicationRequest): Promise<ApplicationDTO> {
    return this.request<ApplicationDTO>('/api/v1/applications', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
  }

  triggerDeploy(id: string): Promise<ApplicationDTO> {
    return this.request<ApplicationDTO>(`/api/v1/applications/${encodeURIComponent(id)}/deploy`, {
      method: 'POST',
    });
  }
}
