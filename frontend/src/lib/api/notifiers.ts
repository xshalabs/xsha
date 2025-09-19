import { request } from "./request";
import type {
  CreateNotifierRequest,
  CreateNotifierResponse,
  UpdateNotifierRequest,
  NotifierListResponse,
  NotifierDetailResponse,
  NotifierListParams,
  NotifierTypesResponse,
  ProjectNotifiersResponse,
  AddNotifierToProjectRequest,
  TestNotifierResponse,
  ApiResponse,
} from "@/types/notifier";

export const notifiersApi = {
  // Get available notifier types
  getTypes: async (): Promise<NotifierTypesResponse> => {
    return request<NotifierTypesResponse>("/notifiers/types");
  },

  // Create a new notifier
  create: async (
    data: CreateNotifierRequest
  ): Promise<CreateNotifierResponse> => {
    return request<CreateNotifierResponse>("/notifiers", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  // List notifiers with filtering and pagination
  list: async (params?: NotifierListParams): Promise<NotifierListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.name) searchParams.set("name", params.name);
    if (params?.type) searchParams.set("type", params.type);
    if (params?.is_enabled !== undefined)
      searchParams.set("is_enabled", params.is_enabled.toString());
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.page_size)
      searchParams.set("page_size", params.page_size.toString());

    const queryString = searchParams.toString();
    const url = queryString ? `/notifiers?${queryString}` : "/notifiers";

    return request<NotifierListResponse>(url);
  },

  // Get a notifier by ID
  get: async (id: number): Promise<NotifierDetailResponse> => {
    return request<NotifierDetailResponse>(`/notifiers/${id}`);
  },

  // Update a notifier
  update: async (
    id: number,
    data: UpdateNotifierRequest
  ): Promise<ApiResponse> => {
    return request<ApiResponse>(`/notifiers/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  // Delete a notifier
  delete: async (id: number): Promise<ApiResponse> => {
    return request<ApiResponse>(`/notifiers/${id}`, {
      method: "DELETE",
    });
  },

  // Test a notifier
  test: async (id: number): Promise<TestNotifierResponse> => {
    return request<TestNotifierResponse>(`/notifiers/${id}/test`, {
      method: "POST",
    });
  },

  // Project-related methods

  // Get notifiers associated with a project
  getProjectNotifiers: async (
    projectId: number
  ): Promise<ProjectNotifiersResponse> => {
    return request<ProjectNotifiersResponse>(`/projects/${projectId}/notifiers`);
  },

  // Add a notifier to a project
  addToProject: async (
    projectId: number,
    data: AddNotifierToProjectRequest
  ): Promise<ApiResponse> => {
    return request<ApiResponse>(`/projects/${projectId}/notifiers`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  // Remove a notifier from a project
  removeFromProject: async (
    projectId: number,
    notifierId: number
  ): Promise<ApiResponse> => {
    return request<ApiResponse>(
      `/projects/${projectId}/notifiers/${notifierId}`,
      {
        method: "DELETE",
      }
    );
  },
};