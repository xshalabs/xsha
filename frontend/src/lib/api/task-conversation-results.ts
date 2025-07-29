import { request } from "./request";
import type {
  UpdateResultRequest,
  ResultListResponse,
  ResultDetailResponse,
  TaskStatsResponse,
  ProjectStatsResponse,
  ResultListByTaskParams,
  ResultListByProjectParams,
} from "@/types/task-conversation-result";

export const taskConversationResultsApi = {
  listByTask: async (
    params: ResultListByTaskParams
  ): Promise<ResultListResponse> => {
    const searchParams = new URLSearchParams();
    searchParams.set("task_id", params.task_id.toString());
    if (params.page) searchParams.set("page", params.page.toString());
    if (params.page_size)
      searchParams.set("page_size", params.page_size.toString());

    const queryString = searchParams.toString();
    return request<ResultListResponse>(`/conversation-results?${queryString}`);
  },

  listByProject: async (
    params: ResultListByProjectParams
  ): Promise<ResultListResponse> => {
    const searchParams = new URLSearchParams();
    searchParams.set("project_id", params.project_id.toString());
    if (params.page) searchParams.set("page", params.page.toString());
    if (params.page_size)
      searchParams.set("page_size", params.page_size.toString());

    const queryString = searchParams.toString();
    return request<ResultListResponse>(
      `/conversation-results/by-project?${queryString}`
    );
  },

  get: async (id: number): Promise<ResultDetailResponse> => {
    return request<ResultDetailResponse>(`/conversation-results/${id}`);
  },

  getByConversationId: async (
    conversationId: number
  ): Promise<ResultDetailResponse> => {
    return request<ResultDetailResponse>(
      `/conversation-results/by-conversation/${conversationId}`
    );
  },

  update: async (
    id: number,
    data: UpdateResultRequest
  ): Promise<{ message: string }> => {
    return request<{ message: string }>(`/conversation-results/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/conversation-results/${id}`, {
      method: "DELETE",
    });
  },

  getTaskStats: async (taskId: number): Promise<TaskStatsResponse> => {
    return request<TaskStatsResponse>(`/stats/tasks/${taskId}`);
  },

  getProjectStats: async (projectId: number): Promise<ProjectStatsResponse> => {
    return request<ProjectStatsResponse>(`/stats/projects/${projectId}`);
  },
};
