// 任务执行状态类型
export type TaskExecutionStatus = 'pending' | 'running' | 'success' | 'failed' | 'cancelled';

// 任务执行日志接口
export interface TaskExecutionLog {
  id: number;
  conversation_id: number;
  status: TaskExecutionStatus;
  workspace_path: string;
  docker_command: string;
  execution_logs: string;
  error_message: string;
  commit_hash: string;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
  conversation?: {
    id: number;
    content: string;
    role: string;
    status: string;
  };
}

// 执行状态响应
export interface ExecutionStatusResponse {
  running_count: number;
  max_concurrency: number;
  can_execute: boolean;
}

// API响应类型
export interface ExecutionLogResponse {
  message: string;
  data: TaskExecutionLog;
}

export interface ExecutionActionResponse {
  message: string;
}

// 执行日志操作类型
export type ExecutionAction = 'cancel' | 'retry';

// 执行日志统计接口
export interface ExecutionStats {
  total_logs: number;
  running_count: number;
  success_count: number;
  failed_count: number;
  cancelled_count: number;
} 