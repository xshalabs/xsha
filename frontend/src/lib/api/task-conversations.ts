import { request } from "./request";
import type {
  CreateConversationRequest,
  CreateConversationResponse,
  ConversationListResponse,
  ConversationWithResultAndLogResponse,
  ConversationGitDiffParams,
  ConversationGitDiffResponse,
  ConversationGitDiffFileParams,
  ConversationGitDiffFileResponse,
} from "@/types/task-conversation";
import type { ExecutionActionResponse } from "@/types/task-execution-log";

export const taskConversationsApi = {
  create: async (
    projectId: number,
    taskId: number,
    data: CreateConversationRequest
  ): Promise<CreateConversationResponse> => {
    return request<CreateConversationResponse>(`/projects/${projectId}/tasks/${taskId}/conversations`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  list: async (
    projectId: number,
    taskId: number,
    params?: { page?: number; page_size?: number }
  ): Promise<ConversationListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.page_size)
      searchParams.set("page_size", params.page_size.toString());

    const queryString = searchParams.toString();
    const url = `/projects/${projectId}/tasks/${taskId}/conversations${queryString ? `?${queryString}` : ""}`;
    return request<ConversationListResponse>(url);
  },

  getDetails: async (
    projectId: number,
    taskId: number,
    conversationId: number
  ): Promise<ConversationWithResultAndLogResponse> => {
    return request<ConversationWithResultAndLogResponse>(
      `/projects/${projectId}/tasks/${taskId}/conversations/${conversationId}/details`
    );
  },

  delete: async (
    projectId: number,
    taskId: number,
    conversationId: number
  ): Promise<{ message: string }> => {
    return request<{ message: string }>(`/projects/${projectId}/tasks/${taskId}/conversations/${conversationId}`, {
      method: "DELETE",
    });
  },


  getGitDiff: async (
    projectId: number,
    taskId: number,
    conversationId: number,
    params?: ConversationGitDiffParams
  ): Promise<ConversationGitDiffResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.include_content) {
      searchParams.set("include_content", "true");
    }

    const url = `/projects/${projectId}/tasks/${taskId}/conversations/${conversationId}/git-diff${
      searchParams.toString() ? `?${searchParams.toString()}` : ""
    }`;
    return request<ConversationGitDiffResponse>(url, {
      method: "GET",
    });
  },

  getGitDiffFile: async (
    projectId: number,
    taskId: number,
    conversationId: number,
    params: ConversationGitDiffFileParams
  ): Promise<ConversationGitDiffFileResponse> => {
    const searchParams = new URLSearchParams();
    searchParams.set("file_path", params.file_path);

    const url = `/projects/${projectId}/tasks/${taskId}/conversations/${conversationId}/git-diff/file?${searchParams.toString()}`;
    return request<ConversationGitDiffFileResponse>(url, {
      method: "GET",
    });
  },

  cancelExecution: async (
    projectId: number,
    taskId: number,
    conversationId: number
  ): Promise<ExecutionActionResponse> => {
    return request<ExecutionActionResponse>(
      `/projects/${projectId}/tasks/${taskId}/conversations/${conversationId}/execution/cancel`,
      {
        method: "POST",
      }
    );
  },

  retryExecution: async (
    projectId: number,
    taskId: number,
    conversationId: number
  ): Promise<ExecutionActionResponse> => {
    return request<ExecutionActionResponse>(
      `/projects/${projectId}/tasks/${taskId}/conversations/${conversationId}/execution/retry`,
      {
        method: "POST",
      }
    );
  },
};
