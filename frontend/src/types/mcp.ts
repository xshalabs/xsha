export interface MCP {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  config: string; // JSON string from backend
  enabled: boolean;
  admin_id?: number;
  admin?: MinimalAdminResponse;
  created_by: string;
  projects?: Project[];
  environments?: DevEnvironment[];
}

export interface MinimalAdminResponse {
  id: number;
  username: string;
  name: string;
  email: string;
  avatar?: AdminAvatarMinimal;
}

export interface AdminAvatarMinimal {
  uuid: string;
  original_name: string;
}

export interface Project {
  id: number;
  name: string;
  protocol: string;
  clone_url: string;
}

export interface DevEnvironment {
  id: number;
  name: string;
  description: string;
}

export interface MCPConfig {
  [key: string]: string | number | boolean | Record<string, unknown> | undefined;
}

export interface CreateMCPRequest {
  name: string;
  description: string;
  config: string;
  enabled: boolean;
}

export interface UpdateMCPRequest {
  name?: string;
  description?: string;
  config?: string;
  enabled?: boolean;
}

export interface MCPListResponse {
  mcps: MCP[];
  total: number;
  page: number;
  page_size: number;
}

export interface MCPDetailResponse {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  config: string; // JSON string from backend, matches actual response
  enabled: boolean;
  admin_id?: number;
  admin?: MinimalAdminResponse;
  created_by: string;
}

export interface CreateMCPResponse {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  config: MCPConfig;
  enabled: boolean;
  admin_id?: number;
  admin?: MinimalAdminResponse;
  created_by: string;
}

export interface MCPListParams {
  name?: string;
  enabled?: boolean;
  page?: number;
  page_size?: number;
}

export interface ProjectMCPsResponse {
  mcps: MCP[];
}

export interface EnvironmentMCPsResponse {
  mcps: MCP[];
}

export interface AddMCPToProjectRequest {
  mcp_id: number;
}

export interface AddMCPToEnvironmentRequest {
  mcp_id: number;
}

export interface MCPFormData {
  name: string;
  description: string;
  config: string;
}

export interface MCPProjectsResponse {
  projects: Project[];
}

export interface MCPEnvironmentsResponse {
  environments: DevEnvironment[];
}

// Common API response format
export interface ApiResponse<T = unknown> {
  message?: string;
  data?: T;
  error?: string;
}