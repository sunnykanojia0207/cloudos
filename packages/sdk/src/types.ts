export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: ApiError;
}

export interface ApiError {
  code: string;
  message: string;
}

// ── Health ──────────────────────────────────────────────────────────────

export interface HealthReport {
  status: string;
  message: string;
  timestamp: string;
}

export interface HealthResponse {
  overall: HealthReport;
  components: Record<string, HealthReport>;
}

// ── Readiness ───────────────────────────────────────────────────────────

export interface ReadinessResponse {
  ready: boolean;
  state: string;
  message?: string;
}

// ── Liveness ────────────────────────────────────────────────────────────

export interface LivenessResponse {
  alive: boolean;
  state: string;
}

// ── Version ─────────────────────────────────────────────────────────────

export interface BuildMetadata {
  version: string;
  commit: string;
  date: string;
  builtBy: string;
  goVersion: string;
}

export interface VersionResponse {
  number: string;
  commit: string;
  date: string;
  build: BuildMetadata;
}

// ── Kernel ──────────────────────────────────────────────────────────────

export interface KernelResponse {
  state: string;
  uptime: string;
  uptimeNs: number;
  startedAt: string;
  subsystems?: string[];
}

// ── System ──────────────────────────────────────────────────────────────

export interface SystemResponse {
  os: string;
  arch: string;
  goVersion: string;
  numCpu: number;
  numGoroutine: number;
  compiler: string;
  hostname?: string;
}

// ── Resource Objects ────────────────────────────────────────────────────

export interface ResourceMeta {
  id: string;
  name: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  createdAt?: string;
  updatedAt?: string;
  resourceVersion?: string;
}

export interface ResourceObject<TSpec = Record<string, unknown>, TStatus = Record<string, unknown>> {
  apiVersion: string;
  kind: string;
  metadata: ResourceMeta;
  spec: TSpec;
  status?: TStatus;
}

export interface ListMeta {
  total: number;
  page?: number;
  limit?: number;
}

export interface ResourceList<TSpec = Record<string, unknown>, TStatus = Record<string, unknown>> {
  apiVersion: string;
  kind: string;
  metadata: ListMeta;
  items: ResourceObject<TSpec, TStatus>[];
}

// ── Capability ──────────────────────────────────────────────────────────

export interface Operation {
  name: string;
  description?: string;
  httpMethod?: string;
  path?: string;
}

export interface CapabilitySpec {
  displayName: string;
  description: string;
  version: string;
  category: string;
  tags?: string[];
  operations: Operation[];
  dependencies?: string[];
}

export interface CapabilityStatus {
  status: string;
  available: boolean;
  providerCount: number;
}

// ── Provider ────────────────────────────────────────────────────────────

export interface Platform {
  os: string;
  arch: string;
}

export interface CapabilityClaim {
  id: string;
  version: string;
  operations: string[];
  features?: string[];
  limits?: Record<string, string>;
  extensions?: string[];
}

export interface ProviderSpec {
  displayName?: string;
  description: string;
  version: string;
  providerType: string;
  author?: string;
  license?: string;
  homepage?: string;
  documentationUrl?: string;
  sourceRepository?: string;
  tags?: string[];
  experimental: boolean;
  enterprise: boolean;
  supportedPlatforms?: Platform[];
  dependencies?: string[];
  capabilities: CapabilityClaim[];
}

export interface ProviderStatus {
  status: string;
  healthy: boolean;
  ready: boolean;
  message?: string;
}

export interface ProviderHealthResponse {
  status: string;
  version: string;
  state: string;
  available: boolean;
  dependencies?: Record<string, string>;
  lastCheck: string;
  message?: string;
}

export interface ProviderCapabilitiesResponse {
  providerId: string;
  providerName: string;
  capabilities: CapabilityClaim[];
}

// ── Resource Engine ──────────────────────────────────────────────────────

export interface ResourceKindDto {
  name: string;
  namespaced: boolean;
  versions?: string[];
}

export interface ResourceKindListResponse {
  kinds: ResourceKindDto[];
  total: number;
}

export type ResourceSpec = Record<string, unknown>;
export type ResourceStatus = Record<string, unknown>;

// ── Controller Runtime ───────────────────────────────────────────────────

export interface ControllerHealthDTO {
  name: string;
  kind: string;
  state: string;
  message?: string;
  lastReconciled?: string;
  reconcileCount: number;
  errorCount: number;
}

export interface ControllerDTO {
  name: string;
  kind: string;
  state: string;
  message?: string;
  health: ControllerHealthDTO;
}

export interface ControllerListResponse {
  controllers: ControllerDTO[];
  total: number;
}

// ── Projects ────────────────────────────────────────────────────────────

export type ProjectPhase = 'Creating' | 'Active' | 'Archived' | 'Deleting';
export type ProjectHealth = 'Healthy' | 'Degraded' | 'Unhealthy' | 'Unknown';
export type ProjectEnvironment = 'development' | 'staging' | 'production' | 'testing';

export interface ProjectCondition {
  type: string;
  status: string;
  reason?: string;
  message?: string;
  lastTransitionTime?: string;
}

export interface ProjectSpec {
  displayName: string;
  description?: string;
  environment: ProjectEnvironment;
  defaultRegion?: string;
  tags?: Record<string, string>;
  quota?: Record<string, unknown>;
  settings?: Record<string, unknown>;
}

export interface ProjectStatus {
  phase: ProjectPhase;
  health: ProjectHealth;
  conditions?: ProjectCondition[];
  lastActivity?: string;
  resourceCount?: number;
  deploymentCount?: number;
}

export interface ProjectDTO {
  apiVersion: string;
  kind: string;
  metadata: ResourceMeta;
  spec: ProjectSpec;
  status?: ProjectStatus;
}

export interface ProjectListResponse {
  apiVersion: string;
  kind: string;
  items: ProjectDTO[];
  metadata: ListMeta;
}

// ── Activity Items ──────────────────────────────────────────────────────

export interface ActivityItem {
  id: string;
  type: string;
  action: string;
  resourceKind: string;
  resourceId: string;
  timestamp: string;
  message?: string;
}
