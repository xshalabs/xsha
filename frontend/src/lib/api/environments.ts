import { request } from "./request";
import type {
  CreateDevEnvironmentRequest,
  CreateDevEnvironmentResponse,
  UpdateDevEnvironmentRequest,
  DevEnvironmentDetailResponse,
  DevEnvironmentListResponse,
  DevEnvironmentListParams,
  DevEnvironmentImageConfig,
} from "@/types/dev-environment";

export const devEnvironmentsApi = {
  create: async (
    data: CreateDevEnvironmentRequest
  ): Promise<CreateDevEnvironmentResponse> => {
    return request<CreateDevEnvironmentResponse>("/environments", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  list: async (
    params?: DevEnvironmentListParams
  ): Promise<DevEnvironmentListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.page_size)
      searchParams.set("page_size", params.page_size.toString());
    if (params?.name) searchParams.set("name", params.name);
    if (params?.docker_image) searchParams.set("docker_image", params.docker_image);

    const queryString = searchParams.toString();
    const url = queryString
      ? `/environments?${queryString}`
      : "/environments";

    return request<DevEnvironmentListResponse>(url);
  },

  get: async (id: number): Promise<DevEnvironmentDetailResponse> => {
    return request<DevEnvironmentDetailResponse>(`/environments/${id}`);
  },

  update: async (
    id: number,
    data: UpdateDevEnvironmentRequest
  ): Promise<{ message: string }> => {
    return request<{ message: string }>(`/environments/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/environments/${id}`, {
      method: "DELETE",
    });
  },


  getAvailableImages: async (): Promise<{
    images: DevEnvironmentImageConfig[];
  }> => {
    return request<{ images: DevEnvironmentImageConfig[] }>(
      "/environments/available-images"
    );
  },

};
