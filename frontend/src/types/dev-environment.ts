import type { Admin, AdminAvatar } from "@/lib/api/types";

export type DevEnvironmentType = string;

export interface DevEnvironment {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  system_prompt: string;
  type: DevEnvironmentType;
  docker_image: string;
  cpu_limit: number;
  memory_limit: number;
  admin_id?: number;
  created_by: string;
  admins?: Admin[];
}

export interface DevEnvironmentDisplay extends DevEnvironment {
  env_vars_map?: Record<string, string>;
}

export interface CreateDevEnvironmentRequest {
  name: string;
  description?: string;
  system_prompt?: string;
  type: DevEnvironmentType;
  docker_image: string;
  cpu_limit: number;
  memory_limit: number;
  env_vars?: Record<string, string>;
}

export interface UpdateDevEnvironmentRequest {
  name?: string;
  description?: string;
  system_prompt?: string;
  cpu_limit?: number;
  memory_limit?: number;
  env_vars?: Record<string, string>;
}

export interface CreateDevEnvironmentResponse {
  message: string;
  environment: DevEnvironment;
}

export interface DevEnvironmentWithVars extends DevEnvironment {
  env_vars: string;
}

export interface DevEnvironmentDetailResponse {
  environment: DevEnvironmentWithVars;
}

export interface DevEnvironmentListResponse {
  message: string;
  environments: DevEnvironment[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
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

// Admin management types
export interface AddAdminToEnvironmentRequest {
  admin_id: number;
}

export interface EnvironmentAdminsResponse {
  admins: Admin[];
}

export interface AdminInfo {
  id: number;
  username: string;
  name: string;
  email: string;
  avatar_id?: number;
  avatar?: AdminAvatar;
}
