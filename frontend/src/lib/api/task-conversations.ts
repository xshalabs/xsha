import { request } from "./request";
import type {
  CreateConversationRequest,
  CreateConversationResponse,
  UpdateConversationRequest,
  ConversationListResponse,
  ConversationDetailResponse,
  LatestConversationResponse,
  ConversationListParams,
} from "@/types/task-conversation";

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

  get: async (id: number): Promise<ConversationDetailResponse> => {
    return request<ConversationDetailResponse>(`/conversations/${id}`);
  },

  update: async (
    id: number,
    data: UpdateConversationRequest
  ): Promise<{ message: string }> => {
    return request<{ message: string }>(`/conversations/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/conversations/${id}`, {
      method: "DELETE",
    });
  },

  getLatest: async (taskId: number): Promise<LatestConversationResponse> => {
    return request<LatestConversationResponse>(
      `/conversations/latest?task_id=${taskId}`
    );
  },
};
