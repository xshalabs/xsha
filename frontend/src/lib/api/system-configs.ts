import { request } from "./request";
import type {
  SystemConfigListResponse,
  BatchUpdateConfigsRequest,
  BatchUpdateConfigsResponse,
} from "@/types/system-config";

export const systemConfigsApi = {
  // 获取所有配置
  listAll: async (): Promise<SystemConfigListResponse> => {
    return request<SystemConfigListResponse>("/system-configs");
  },

  // 批量更新配置
  batchUpdate: async (
    data: BatchUpdateConfigsRequest
  ): Promise<BatchUpdateConfigsResponse> => {
    return request<BatchUpdateConfigsResponse>("/system-configs", {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },
}; 