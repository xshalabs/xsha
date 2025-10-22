import { request } from "./request";
import type {
  CreateProviderRequest,
  CreateProviderResponse,
  UpdateProviderRequest,
  ProviderDetailResponse,
  ProviderListResponse,
  ProviderListParams,
  ProviderTypesResponse,
} from "@/types/provider";

export const providerApi = {
  getTypes: async (): Promise<ProviderTypesResponse> => {
    return request<ProviderTypesResponse>("/providers/types");
  },

  create: async (
    data: CreateProviderRequest
  ): Promise<CreateProviderResponse> => {
    return request<CreateProviderResponse>("/providers", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  list: async (params?: ProviderListParams): Promise<ProviderListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.page_size)
      searchParams.set("page_size", params.page_size.toString());
    if (params?.name) searchParams.set("name", params.name);
    if (params?.type) searchParams.set("type", params.type);

    const queryString = searchParams.toString();
    const url = queryString ? `/providers?${queryString}` : "/providers";

    return request<ProviderListResponse>(url);
  },

  get: async (id: number): Promise<ProviderDetailResponse> => {
    return request<ProviderDetailResponse>(`/providers/${id}`);
  },

  update: async (
    id: number,
    data: UpdateProviderRequest
  ): Promise<{ message: string }> => {
    return request<{ message: string }>(`/providers/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/providers/${id}`, {
      method: "DELETE",
    });
  },
};
