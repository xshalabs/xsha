// 管理员操作类型
export const AdminOperationType = {
  CREATE: 'create',
  READ: 'read', 
  UPDATE: 'update',
  DELETE: 'delete',
  LOGIN: 'login',
  LOGOUT: 'logout'
} as const;

export type AdminOperationType = typeof AdminOperationType[keyof typeof AdminOperationType];

// 管理员操作日志接口
export interface AdminOperationLog {
  id: number;
  created_at: string;
  updated_at: string;
  username: string;
  operation: AdminOperationType;
  resource: string;
  resource_id: string;
  description: string;
  details: string;
  success: boolean;
  error_msg: string;
  ip: string;
  user_agent: string;
  method: string;
  path: string;
  operation_time: string;
}

// 登录日志接口
export interface LoginLog {
  id: number;
  created_at: string;
  updated_at: string;
  username: string;
  success: boolean;
  ip: string;
  user_agent: string;
  reason: string;
  login_time: string;
}

// 操作日志列表请求参数
export interface AdminOperationLogListParams {
  username?: string;
  resource?: string;
  operation?: AdminOperationType;
  success?: boolean;
  start_time?: string;
  end_time?: string;
  page?: number;
  page_size?: number;
}

// 登录日志列表请求参数
export interface LoginLogListParams {
  username?: string;
  page?: number;
  page_size?: number;
}

// 操作日志列表响应
export interface AdminOperationLogListResponse {
  message: string;
  logs: AdminOperationLog[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// 登录日志列表响应
export interface LoginLogListResponse {
  message: string;
  logs: LoginLog[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// 操作日志详情响应
export interface AdminOperationLogDetailResponse {
  message: string;
  log: AdminOperationLog;
}

// 操作统计响应
export interface AdminOperationStatsResponse {
  message: string;
  operation_stats: Record<string, number>;
  resource_stats: Record<string, number>;
  start_time: string;
  end_time: string;
}

// 操作统计请求参数
export interface AdminOperationStatsParams {
  username?: string;
  start_time?: string;
  end_time?: string;
} 