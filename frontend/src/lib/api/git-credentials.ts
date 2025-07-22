import { request } from './request';
import type {
  CreateGitCredentialRequest,
  CreateGitCredentialResponse,
  UpdateGitCredentialRequest,
  GitCredentialListResponse,
  GitCredentialDetailResponse,
  UseGitCredentialResponse,
  GitCredentialListParams
} from '@/types/git-credentials';

export const gitCredentialsApi = {
  // 创建 Git 凭据
  create: async (data: CreateGitCredentialRequest): Promise<CreateGitCredentialResponse> => {
    return request<CreateGitCredentialResponse>('/git-credentials', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // 获取 Git 凭据列表
  list: async (params?: GitCredentialListParams): Promise<GitCredentialListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.type) searchParams.set('type', params.type);
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.page_size) searchParams.set('page_size', params.page_size.toString());
    
    const queryString = searchParams.toString();
    const url = queryString ? `/git-credentials?${queryString}` : '/git-credentials';
    
    return request<GitCredentialListResponse>(url);
  },

  // 获取单个 Git 凭据详情
  get: async (id: number): Promise<GitCredentialDetailResponse> => {
    return request<GitCredentialDetailResponse>(`/git-credentials/${id}`);
  },

  // 更新 Git 凭据
  update: async (id: number, data: UpdateGitCredentialRequest): Promise<{ message: string }> => {
    return request<{ message: string }>(`/git-credentials/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // 删除 Git 凭据
  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/git-credentials/${id}`, {
      method: 'DELETE',
    });
  },

  // 切换凭据激活状态
  toggle: async (id: number, isActive: boolean): Promise<{ message: string }> => {
    return request<{ message: string }>(`/git-credentials/${id}/toggle`, {
      method: 'PATCH',
      body: JSON.stringify({ is_active: isActive }),
    });
  },

  // 使用凭据（获取解密后的凭据信息）
  use: async (id: number): Promise<UseGitCredentialResponse> => {
    return request<UseGitCredentialResponse>(`/git-credentials/${id}/use`, {
      method: 'POST',
    });
  },
}; 