import { request } from "./request";
import type {
  CreateDevEnvironmentRequest,
  CreateDevEnvironmentResponse,
  UpdateDevEnvironmentRequest,
  DevEnvironmentDetailResponse,
  DevEnvironmentListResponse,
  DevEnvironmentVarsResponse,
  DevEnvironmentListParams,
} from "@/types/dev-environment";

export const devEnvironmentsApi = {
  create: async (
    data: CreateDevEnvironmentRequest
  ): Promise<CreateDevEnvironmentResponse> => {
    return request<CreateDevEnvironmentResponse>("/dev-environments", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  list: async (
    params?: DevEnvironmentListParams
  ): Promise<DevEnvironmentListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.type) searchParams.set("type", params.type);
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.page_size)
      searchParams.set("page_size", params.page_size.toString());
    if (params?.name) searchParams.set("name", params.name);

    const queryString = searchParams.toString();
    const url = queryString
      ? `/dev-environments?${queryString}`
      : "/dev-environments";

    return request<DevEnvironmentListResponse>(url);
  },

  get: async (id: number): Promise<DevEnvironmentDetailResponse> => {
    return request<DevEnvironmentDetailResponse>(`/dev-environments/${id}`);
  },

  update: async (
    id: number,
    data: UpdateDevEnvironmentRequest
  ): Promise<{ message: string }> => {
    return request<{ message: string }>(`/dev-environments/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/dev-environments/${id}`, {
      method: "DELETE",
    });
  },

  getEnvVars: async (id: number): Promise<DevEnvironmentVarsResponse> => {
    return request<DevEnvironmentVarsResponse>(
      `/dev-environments/${id}/env-vars`
    );
  },

  updateEnvVars: async (
    id: number,
    envVars: Record<string, string>
  ): Promise<{ message: string }> => {
    return request<{ message: string }>(`/dev-environments/${id}/env-vars`, {
      method: "PUT",
      body: JSON.stringify(envVars),
    });
  },

  getAvailableTypes: async (): Promise<{ types: { name: string; image: string }[] }> => {
    return request<{ types: { name: string; image: string }[] }>("/dev-environments/available-types");
  },
};
