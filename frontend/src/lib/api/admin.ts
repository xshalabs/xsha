import { request } from './request';
import type {
  CreateAdminRequest,
  UpdateAdminRequest,
  ChangePasswordRequest,
  AdminListResponse,
  AdminResponse,
  CreateAdminResponse,
  AvatarUploadResponse
} from './types';

export const adminApi = {
  // Get all admins with pagination and filtering
  getAdmins: async (params?: {
    username?: string;
    is_active?: boolean;
    page?: number;
    page_size?: number;
  }): Promise<AdminListResponse> => {
    const searchParams = new URLSearchParams();
    
    if (params?.username) {
      searchParams.append('username', params.username);
    }
    if (params?.is_active !== undefined) {
      searchParams.append('is_active', params.is_active.toString());
    }
    if (params?.page) {
      searchParams.append('page', params.page.toString());
    }
    if (params?.page_size) {
      searchParams.append('page_size', params.page_size.toString());
    }

    const url = `/admin/users${searchParams.toString() ? `?${searchParams.toString()}` : ''}`;
    return request<AdminListResponse>(url);
  },

  // Get admin by ID
  getAdmin: async (id: number): Promise<AdminResponse> => {
    return request<AdminResponse>(`/admin/users/${id}`);
  },

  // Create new admin
  createAdmin: async (data: CreateAdminRequest): Promise<CreateAdminResponse> => {
    return request<CreateAdminResponse>('/admin/users', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // Update admin
  updateAdmin: async (id: number, data: UpdateAdminRequest): Promise<{ message: string }> => {
    return request<{ message: string }>(`/admin/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // Delete admin
  deleteAdmin: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/admin/users/${id}`, {
      method: 'DELETE',
    });
  },

  // Change admin password
  changePassword: async (id: number, data: ChangePasswordRequest): Promise<{ message: string }> => {
    return request<{ message: string }>(`/admin/users/${id}/password`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // Upload avatar
  uploadAvatar: async (file: File): Promise<AvatarUploadResponse> => {
    const formData = new FormData();
    formData.append('file', file);

    return request<AvatarUploadResponse>('/admin/avatar/upload', {
      method: 'POST',
      body: formData,
    });
  },

  // Update admin's avatar
  updateAdminAvatar: async (adminId: number, avatarId: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/admin/users/${adminId}/avatar`, {
      method: 'PUT',
      body: JSON.stringify({ avatar_id: avatarId }),
    });
  },
};