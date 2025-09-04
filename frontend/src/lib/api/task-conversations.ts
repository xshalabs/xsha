import { request } from "./request";
import type {
  CreateConversationRequest,
  CreateConversationResponse,
  ConversationListResponse,
  ConversationWithResultAndLogResponse,
  ConversationListParams,
  ConversationGitDiffParams,
  ConversationGitDiffResponse,
  ConversationGitDiffFileParams,
  ConversationGitDiffFileResponse,
} from "@/types/task-conversation";
import type { ExecutionActionResponse } from "@/types/task-execution-log";

export const taskConversationsApi = {
  create: async (
    data: CreateConversationRequest
  ): Promise<CreateConversationResponse> => {
    return request<CreateConversationResponse>("/conversations", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  list: async (
    params: ConversationListParams
  ): Promise<ConversationListResponse> => {
    const searchParams = new URLSearchParams();
    searchParams.set("task_id", params.task_id.toString());
    if (params.page) searchParams.set("page", params.page.toString());
    if (params.page_size)
      searchParams.set("page_size", params.page_size.toString());

    const queryString = searchParams.toString();
    return request<ConversationListResponse>(`/conversations?${queryString}`);
  },

  getDetails: async (
    id: number
  ): Promise<ConversationWithResultAndLogResponse> => {
    return request<ConversationWithResultAndLogResponse>(
      `/conversations/${id}/details`
    );
  },

  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/conversations/${id}`, {
      method: "DELETE",
    });
  },


  getGitDiff: async (
    conversationId: number,
    params?: ConversationGitDiffParams
  ): Promise<ConversationGitDiffResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.include_content) {
      searchParams.set("include_content", "true");
    }

    const url = `/conversations/${conversationId}/git-diff${
      searchParams.toString() ? `?${searchParams.toString()}` : ""
    }`;
    return request<ConversationGitDiffResponse>(url, {
      method: "GET",
    });
  },

  getGitDiffFile: async (
    conversationId: number,
    params: ConversationGitDiffFileParams
  ): Promise<ConversationGitDiffFileResponse> => {
    const searchParams = new URLSearchParams();
    searchParams.set("file_path", params.file_path);

    const url = `/conversations/${conversationId}/git-diff/file?${searchParams.toString()}`;
    return request<ConversationGitDiffFileResponse>(url, {
      method: "GET",
    });
  },

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
};
