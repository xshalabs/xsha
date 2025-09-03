export const GitCredentialType = {
  PASSWORD: "password",
  TOKEN: "token",
  SSH_KEY: "ssh_key",
} as const;

export type GitCredentialType =
  (typeof GitCredentialType)[keyof typeof GitCredentialType];

export interface GitCredential {
  id: number;
  name: string;
  description: string;
  type: GitCredentialType;
  username: string;
  admin_id?: number;
  admin?: MinimalAdminResponse;
  admins?: MinimalAdminResponse[];
  created_by: string;
  public_key?: string;
  created_at: string;
  updated_at: string;
}

export interface MinimalAdminResponse {
  id: number;
  username: string;
  name: string;
  email: string;
  avatar?: AdminAvatarMinimal;
}

export interface AdminAvatarMinimal {
  uuid: string;
  original_name: string;
}

export interface CreateGitCredentialRequest {
  name: string;
  description: string;
  type: GitCredentialType;
  username: string;
  secret_data: Record<string, string>;
}

export interface UpdateGitCredentialRequest {
  name?: string;
  description?: string;
  username?: string;
  secret_data?: Record<string, string>;
}

export interface GitCredentialListResponse {
  message: string;
  credentials: GitCredential[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface GitCredentialDetailResponse {
  credential: GitCredential;
}

export interface CreateGitCredentialResponse {
  message: string;
  credential: GitCredential;
}

export interface GitCredentialListParams {
  name?: string;
  page?: number;
  page_size?: number;
}

export interface GitCredentialFormData {
  name: string;
  description: string;
  type: GitCredentialType;
  username: string;
  password?: string;
  token?: string;
  private_key?: string;
  public_key?: string;
}
