// 项目协议类型
export const GitProtocolType = {
  HTTPS: 'https',
  SSH: 'ssh'
} as const;

export type GitProtocolType = typeof GitProtocolType[keyof typeof GitProtocolType];

// 基础项目接口
export interface Project {
  id: number;
  name: string;
  description: string;
  repo_url: string;
  protocol: GitProtocolType;
  default_branch: string;
  credential_id?: number;
  created_by: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_used?: string;
  credential?: {
    id: number;
    name: string;
    type: string;
  };
}

// 创建项目请求
export interface CreateProjectRequest {
  name: string;
  description?: string;
  repo_url: string;
  protocol: GitProtocolType;
  default_branch?: string;
  credential_id?: number;
}

// 更新项目请求
export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  repo_url?: string;
  protocol?: GitProtocolType;
  default_branch?: string;
  credential_id?: number;
}

// 项目列表响应
export interface ProjectListResponse {
  message: string;
  projects: Project[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// 项目详情响应
export interface ProjectDetailResponse {
  project: Project;
}

// 创建项目响应
export interface CreateProjectResponse {
  message: string;
  project: Project;
}

// 使用项目响应
export interface UseProjectResponse {
  message: string;
  project: Project;
}

// 兼容凭据响应
export interface CompatibleCredentialsResponse {
  message: string;
  credentials: Array<{
    id: number;
    name: string;
    type: string;
    username: string;
    is_active: boolean;
  }>;
}



// 列表查询参数
export interface ProjectListParams {
  protocol?: GitProtocolType;
  page?: number;
  page_size?: number;
}

// 表单数据类型
export interface ProjectFormData {
  name: string;
  description: string;
  repo_url: string;
  protocol: GitProtocolType;
  default_branch: string;
  credential_id?: number;
}

// 解析仓库URL请求
export interface ParseRepositoryURLRequest {
  repo_url: string;
}

// 解析仓库URL响应
export interface ParseRepositoryURLResponse {
  message: string;
  result: {
    protocol: string;
    host: string;
    owner: string;
    repo: string;
    is_valid: boolean;
  };
}

// 获取仓库分支请求
export interface FetchRepositoryBranchesRequest {
  repo_url: string;
  credential_id?: number;
}

// 获取仓库分支响应
export interface FetchRepositoryBranchesResponse {
  message: string;
  result: {
    can_access: boolean;
    error_message: string;
    branches: string[];
  };
}

// 验证仓库访问请求
export interface ValidateRepositoryAccessRequest {
  repo_url: string;
  credential_id?: number;
}

// 验证仓库访问响应
export interface ValidateRepositoryAccessResponse {
  message: string;
  can_access: boolean;
  error?: string;
} 