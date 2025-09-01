export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  message: string;
  user: string;
  token: string;
}

export interface UserResponse {
  user: string;
  admin_id: number;
  name: string;
  authenticated: boolean;
  message: string;
  avatar?: AdminAvatar;
  role: AdminRole;
}

export interface ApiErrorResponse {
  error: string;
  details?: string;
}

// Avatar types
export interface AdminAvatar {
  id: number;
  uuid: string;
  original_name: string;
  file_size: number;
  content_type: string;
  created_at: string;
}

// Admin role types
export type AdminRole = 'super_admin' | 'admin' | 'developer';

// Admin types
export interface Admin {
  id: number;
  created_at: string;
  updated_at: string;
  username: string;
  name: string;
  email: string;
  role: AdminRole;
  is_active: boolean;
  last_login_at?: string;
  last_login_ip?: string;
  created_by: string;
  avatar_id?: number;
  avatar?: AdminAvatar;
}

export interface CreateAdminRequest {
  username: string;
  password: string;
  name: string;
  email?: string;
  role?: AdminRole;
}

export interface UpdateAdminRequest {
  username?: string;
  name?: string;
  email?: string;
  is_active?: boolean;
  role?: AdminRole;
}

export interface ChangePasswordRequest {
  new_password: string;
}

export interface ChangeOwnPasswordRequest {
  current_password: string;
  new_password: string;
}

export interface UpdateOwnAvatarRequest {
  avatar_uuid: string;
}

export interface AdminListResponse {
  message: string;
  admins: Admin[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface AdminResponse {
  admin: Admin;
}

export interface CreateAdminResponse {
  message: string;
  admin: Admin;
}

// Avatar API types
export interface AvatarUploadResponse {
  message: string;
  data: {
    id: number;
    uuid: string;
    original_name: string;
    file_size: number;
    content_type: string;
    preview_url: string;
    created_at: string;
  };
}

export interface UpdateAvatarRequest {
  avatar_id: number;
}

// Role management types
export interface RoleListResponse {
  roles: AdminRole[];
}
