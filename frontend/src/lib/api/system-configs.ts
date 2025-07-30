import { request } from "./request";
import type {
  UpdateSystemConfigRequest,
  SystemConfigDetailResponse,
  SystemConfigListResponse,
  UpdateSystemConfigResponse,
  DevEnvironmentTypesResponse,
  UpdateDevEnvironmentTypesRequest,
  UpdateDevEnvironmentTypesResponse,
} from "@/types/system-config";

export interface SystemConfigListParams {
  category?: string;
  page?: number;
  page_size?: number;
}

export const systemConfigsApi = {
  list: async (
    params?: SystemConfigListParams
  ): Promise<SystemConfigListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.category) searchParams.set("category", params.category);
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.page_size)
      searchParams.set("page_size", params.page_size.toString());

    const queryString = searchParams.toString();
    const url = queryString
      ? `/system-configs?${queryString}`
      : "/system-configs";

    return request<SystemConfigListResponse>(url);
  },

  get: async (id: number): Promise<SystemConfigDetailResponse> => {
    return request<SystemConfigDetailResponse>(`/system-configs/${id}`);
  },

  update: async (
    id: number,
    data: UpdateSystemConfigRequest
  ): Promise<UpdateSystemConfigResponse> => {
    return request<UpdateSystemConfigResponse>(`/system-configs/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  getDevEnvironmentTypes: async (): Promise<DevEnvironmentTypesResponse> => {
    return request<DevEnvironmentTypesResponse>("/system-configs/dev-environment-types");
  },

  updateDevEnvironmentTypes: async (
    data: UpdateDevEnvironmentTypesRequest
  ): Promise<UpdateDevEnvironmentTypesResponse> => {
    return request<UpdateDevEnvironmentTypesResponse>("/system-configs/dev-environment-types", {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },
}; 