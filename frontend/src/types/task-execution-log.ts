export interface TaskExecutionLog {
  id: number;
  conversation_id: number;
  docker_command: string;
  execution_logs: string;
  error_message: string;
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

export interface ExecutionStatusResponse {
  running_count: number;
  max_concurrency: number;
  can_execute: boolean;
}

export interface ExecutionLogResponse {
  message: string;
  data: TaskExecutionLog;
}

export interface ExecutionActionResponse {
  message: string;
}

export type ExecutionAction = "cancel" | "retry";

export interface ExecutionStats {
  total_logs: number;
  running_count: number;
  success_count: number;
  failed_count: number;
  cancelled_count: number;
}
