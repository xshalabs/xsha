import { request } from "./request";
import type {
  TaskExecutionLog,
  ExecutionActionResponse,
  ExecutionStatusResponse,
} from "@/types/task-execution-log";

export const taskExecutionLogsApi = {

  cancelExecution: async (
    conversationId: number
  ): Promise<ExecutionActionResponse> => {
    return request<ExecutionActionResponse>(
      `/conversations/${conversationId}/execution/cancel`,
      {
        method: "POST",
      }
    );
  },

  retryExecution: async (
    conversationId: number
  ): Promise<ExecutionActionResponse> => {
    return request<ExecutionActionResponse>(
      `/conversations/${conversationId}/execution/retry`,
      {
        method: "POST",
      }
    );
  },

  getExecutionStatus: async (): Promise<ExecutionStatusResponse> => {
    return request<ExecutionStatusResponse>("/execution/status");
  },
};
