import { request } from './request';
import type {
  CreateDevEnvironmentRequest,
  CreateDevEnvironmentResponse,
  UpdateDevEnvironmentRequest,
  DevEnvironmentDetailResponse,
  DevEnvironmentListResponse,
  DevEnvironmentVarsResponse,
  DevEnvironmentListParams
} from '@/types/dev-environment';

export const devEnvironmentsApi = {
  // 创建开发环境
  create: async (data: CreateDevEnvironmentRequest): Promise<CreateDevEnvironmentResponse> => {
    return request<CreateDevEnvironmentResponse>('/dev-environments', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // 获取开发环境列表
  list: async (params?: DevEnvironmentListParams): Promise<DevEnvironmentListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.type) searchParams.set('type', params.type);
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.page_size) searchParams.set('page_size', params.page_size.toString());
    if (params?.name) searchParams.set('name', params.name);
    
    const queryString = searchParams.toString();
    const url = queryString ? `/dev-environments?${queryString}` : '/dev-environments';
    
    return request<DevEnvironmentListResponse>(url);
  },

  // 获取单个开发环境详情
  get: async (id: number): Promise<DevEnvironmentDetailResponse> => {
    return request<DevEnvironmentDetailResponse>(`/dev-environments/${id}`);
  },

  // 更新开发环境
  update: async (id: number, data: UpdateDevEnvironmentRequest): Promise<{ message: string }> => {
    return request<{ message: string }>(`/dev-environments/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // 删除开发环境
  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/dev-environments/${id}`, {
      method: 'DELETE',
    });
  },



  // 获取环境变量
  getEnvVars: async (id: number): Promise<DevEnvironmentVarsResponse> => {
    return request<DevEnvironmentVarsResponse>(`/dev-environments/${id}/env-vars`);
  },

  // 更新环境变量
  updateEnvVars: async (id: number, envVars: Record<string, string>): Promise<{ message: string }> => {
    return request<{ message: string }>(`/dev-environments/${id}/env-vars`, {
      method: 'PUT',
      body: JSON.stringify(envVars),
    });
  },
}; 