import { request } from './request';
import type {
  CreateAdminRequest,
  UpdateAdminRequest,
  ChangePasswordRequest,
  AdminListResponse,
  AdminResponse,
  CreateAdminResponse,
  AvatarUploadResponse,
  RoleListResponse
} from './types';

export const adminApi = {
  // Get all admins with pagination and filtering
  getAdmins: async (params?: {
    search?: string;
    username?: string; // kept for backward compatibility
    is_active?: boolean;
    page?: number;
    page_size?: number;
  }): Promise<AdminListResponse> => {
    const searchParams = new URLSearchParams();
    
    // Use 'search' parameter if available, fallback to 'username' for backward compatibility
    if (params?.search) {
      searchParams.append('search', params.search);
    } else if (params?.username) {
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

  // Update admin's avatar by avatar UUID
  updateAdminAvatar: async (avatarUuid: string, adminId: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/admin/avatar/${avatarUuid}`, {
      method: 'PUT',
      body: JSON.stringify({ admin_id: adminId }),
    });
  },

  // Get available roles
  getRoles: async (): Promise<RoleListResponse> => {
    return request<RoleListResponse>('/admin/roles');
  },

  // Get all admins from v1 API endpoint
  getV1Admins: async (params?: {
    search?: string;
    username?: string;
    is_active?: boolean;
    page?: number;
    page_size?: number;
  }): Promise<AdminListResponse> => {
    const searchParams = new URLSearchParams();
    
    if (params?.search) {
      searchParams.append('search', params.search);
    } else if (params?.username) {
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

    const url = `/admins${searchParams.toString() ? `?${searchParams.toString()}` : ''}`;
    return request<AdminListResponse>(url);
  },
};