export type DevEnvironmentType = string;

export interface DevEnvironment {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  type: DevEnvironmentType;
  docker_image: string;
  cpu_limit: number;
  memory_limit: number;
  env_vars: string;
  created_by: string;
}

export interface DevEnvironmentDisplay
  extends Omit<DevEnvironment, "env_vars"> {
  env_vars_map: Record<string, string>;
}

export interface CreateDevEnvironmentRequest {
  name: string;
  description?: string;
  type: DevEnvironmentType;
  docker_image: string;
  cpu_limit: number;
  memory_limit: number;
  env_vars?: Record<string, string>;
}

export interface UpdateDevEnvironmentRequest {
  name?: string;
  description?: string;
  cpu_limit?: number;
  memory_limit?: number;
  env_vars?: Record<string, string>;
}

export interface CreateDevEnvironmentResponse {
  message: string;
  environment: DevEnvironment;
}

export interface DevEnvironmentDetailResponse {
  environment: DevEnvironment;
}

export interface DevEnvironmentListResponse {
  message: string;
  environments: DevEnvironment[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface DevEnvironmentVarsResponse {
  env_vars: Record<string, string>;
}

export interface DevEnvironmentListParams {
  page?: number;
  page_size?: number;
  name?: string;
  docker_image?: string;
}

export interface DevEnvironmentImageConfig {
  image: string;
  name: string;
  type: string;
}

export interface ResourceUsageStats {
  total_cpu: number;
  used_cpu: number;
  total_memory: number;
  used_memory: number;
  running_count: number;
  total_count: number;
}

export interface EnvironmentStats {
  by_type: Record<DevEnvironmentType, number>;
  resource_usage: ResourceUsageStats;
}
