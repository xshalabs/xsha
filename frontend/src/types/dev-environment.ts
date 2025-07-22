// 开发环境状态枚举
export type DevEnvironmentStatus = 
  | 'stopped'    // 已停止
  | 'starting'   // 启动中
  | 'running'    // 运行中
  | 'stopping'   // 停止中
  | 'error';     // 错误状态

// 开发环境类型枚举
export type DevEnvironmentType = 
  | 'claude_code'  // Claude Code环境
  | 'gemini_cli'   // Gemini CLI环境
  | 'opencode';    // OpenCode环境

// 环境控制操作类型
export type EnvironmentAction = 
  | 'start'    // 启动
  | 'stop'     // 停止
  | 'restart'; // 重启

// 开发环境基础接口
export interface DevEnvironment {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  type: DevEnvironmentType;
  status: DevEnvironmentStatus;
  cpu_limit: number;
  memory_limit: number;
  env_vars: string; // JSON字符串
  created_by: string;
  last_used: string | null;
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

// 环境控制请求接口
export interface EnvironmentControlRequest {
  action: EnvironmentAction;
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

export interface UseDevEnvironmentResponse {
  message: string;
  environment: DevEnvironment;
}

export interface DevEnvironmentVarsResponse {
  env_vars: Record<string, string>;
}

// 列表查询参数接口
export interface DevEnvironmentListParams {
  page?: number;
  page_size?: number;
  type?: DevEnvironmentType;
  status?: DevEnvironmentStatus;
}

// 环境类型选项
export interface DevEnvironmentTypeOption {
  value: DevEnvironmentType;
  label: string;
  description: string;
}

// 环境状态选项
export interface DevEnvironmentStatusOption {
  value: DevEnvironmentStatus;
  label: string;
  color: string;
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
  by_status: Record<DevEnvironmentStatus, number>;
  resource_usage: ResourceUsageStats;
} 