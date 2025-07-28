// 开发环境类型枚举
export type DevEnvironmentType = 
  | 'claude_code'  // Claude Code环境
  | 'gemini_cli'   // Gemini CLI环境
  | 'opencode';    // OpenCode环境

// 开发环境基础接口
export interface DevEnvironment {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  type: DevEnvironmentType;
  cpu_limit: number;
  memory_limit: number;
  env_vars: string; // JSON字符串
  created_by: string;
}

// 开发环境显示接口（包含解析后的环境变量）
export interface DevEnvironmentDisplay extends Omit<DevEnvironment, 'env_vars'> {
  env_vars_map: Record<string, string>;
}

// 创建开发环境请求接口
export interface CreateDevEnvironmentRequest {
  name: string;
  description?: string;
  type: DevEnvironmentType;
  cpu_limit: number;
  memory_limit: number;
  env_vars?: Record<string, string>;
}

// 更新开发环境请求接口
export interface UpdateDevEnvironmentRequest {
  name?: string;
  description?: string;
  cpu_limit?: number;
  memory_limit?: number;
  env_vars?: Record<string, string>;
}



// API响应接口
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

// 列表查询参数接口
export interface DevEnvironmentListParams {
  page?: number;
  page_size?: number;
  type?: DevEnvironmentType;
  name?: string;
}

// 环境类型选项
export interface DevEnvironmentTypeOption {
  value: DevEnvironmentType;
  label: string;
  description: string;
}



// 资源使用统计
export interface ResourceUsageStats {
  total_cpu: number;
  used_cpu: number;
  total_memory: number;
  used_memory: number;
  running_count: number;
  total_count: number;
}

// 环境统计信息
export interface EnvironmentStats {
  by_type: Record<DevEnvironmentType, number>;
  resource_usage: ResourceUsageStats;
} 