// Git 凭据类型定义
export const GitCredentialType = {
  PASSWORD: 'password',
  TOKEN: 'token',
  SSH_KEY: 'ssh_key'
} as const;

export type GitCredentialType = typeof GitCredentialType[keyof typeof GitCredentialType];

// 基础 Git 凭据接口
export interface GitCredential {
  id: number;
  name: string;
  description: string;
  type: GitCredentialType;
  username: string;
  created_by: string;
  public_key?: string;
  created_at: string;
  updated_at: string;
}

// 创建 Git 凭据请求
export interface CreateGitCredentialRequest {
  name: string;
  description: string;
  type: GitCredentialType;
  username: string;
  secret_data: Record<string, string>;
}

// 更新 Git 凭据请求
export interface UpdateGitCredentialRequest {
  name?: string;
  description?: string;
  username?: string;
  secret_data?: Record<string, string>;
}

// Git 凭据列表响应
export interface GitCredentialListResponse {
  message: string;
  credentials: GitCredential[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// Git 凭据详情响应
export interface GitCredentialDetailResponse {
  credential: GitCredential;
}

// 创建凭据响应
export interface CreateGitCredentialResponse {
  message: string;
  credential: GitCredential;
}



// 列表查询参数
export interface GitCredentialListParams {
  type?: GitCredentialType;
  page?: number;
  page_size?: number;
}

// 表单数据类型
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