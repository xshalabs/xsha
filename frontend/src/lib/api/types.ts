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
  authenticated: boolean;
  message: string;
}

export interface ApiErrorResponse {
  error: string;
  details?: string;
}

// Admin types
export interface Admin {
  id: number;
  created_at: string;
  updated_at: string;
  username: string;
  name: string;
  email: string;
  is_active: boolean;
  last_login_at?: string;
  last_login_ip?: string;
  created_by: string;
}

export interface CreateAdminRequest {
  username: string;
  password: string;
  name: string;
  email?: string;
}

export interface UpdateAdminRequest {
  username?: string;
  name?: string;
  email?: string;
  is_active?: boolean;
}

export interface ChangePasswordRequest {
  new_password: string;
}

export interface ChangeOwnPasswordRequest {
  current_password: string;
  new_password: string;
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


