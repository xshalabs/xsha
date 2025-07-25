import { request } from './request';
import type {
  TaskExecutionLog,
  ExecutionActionResponse,
  ExecutionStatusResponse
} from '@/types/task-execution-log';

export const taskExecutionLogsApi = {
  // 获取执行日志
  getExecutionLog: async (conversationId: number): Promise<TaskExecutionLog> => {
    const response = await request<TaskExecutionLog>(`/task-conversations/${conversationId}/execution-log`);
    return response;
  },

  // 取消任务执行
  cancelExecution: async (conversationId: number): Promise<ExecutionActionResponse> => {
    return request<ExecutionActionResponse>(`/task-conversations/${conversationId}/execution/cancel`, {
      method: 'POST',
    });
  },

  // 重试任务执行
  retryExecution: async (conversationId: number): Promise<ExecutionActionResponse> => {
    return request<ExecutionActionResponse>(`/task-conversations/${conversationId}/execution/retry`, {
      method: 'POST',
    });
  },

  // 获取执行状态统计（可选，用于监控面板）
  getExecutionStatus: async (): Promise<ExecutionStatusResponse> => {
    return request<ExecutionStatusResponse>('/execution/status');
  },
}; 