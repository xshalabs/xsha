import { request } from './request';
import type {
  CreateProjectRequest,
  CreateProjectResponse,
  UpdateProjectRequest,
  ProjectListResponse,
  ProjectDetailResponse,
  CompatibleCredentialsResponse,
  ProjectListParams,
  ParseRepositoryURLResponse,
  FetchRepositoryBranchesRequest,
  FetchRepositoryBranchesResponse,
  ValidateRepositoryAccessRequest,
  ValidateRepositoryAccessResponse
} from '@/types/project';

export const projectsApi = {
  // 创建项目
  create: async (data: CreateProjectRequest): Promise<CreateProjectResponse> => {
    return request<CreateProjectResponse>('/projects', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // 获取项目列表
  list: async (params?: ProjectListParams): Promise<ProjectListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.protocol) searchParams.set('protocol', params.protocol);
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.page_size) searchParams.set('page_size', params.page_size.toString());
    
    const queryString = searchParams.toString();
    const url = queryString ? `/projects?${queryString}` : '/projects';
    
    return request<ProjectListResponse>(url);
  },

  // 获取单个项目详情
  get: async (id: number): Promise<ProjectDetailResponse> => {
    return request<ProjectDetailResponse>(`/projects/${id}`);
  },

  // 更新项目
  update: async (id: number, data: UpdateProjectRequest): Promise<{ message: string }> => {
    return request<{ message: string }>(`/projects/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // 删除项目
  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/projects/${id}`, {
      method: 'DELETE',
    });
  },

  // 获取与协议兼容的凭据列表
  getCompatibleCredentials: async (protocol: string): Promise<CompatibleCredentialsResponse> => {
    return request<CompatibleCredentialsResponse>(`/projects/credentials?protocol=${protocol}`);
  },

  // 解析仓库URL
  parseUrl: async (repoUrl: string): Promise<ParseRepositoryURLResponse> => {
    return request<ParseRepositoryURLResponse>('/projects/parse-url', {
      method: 'POST',
      body: JSON.stringify({ repo_url: repoUrl }),
    });
  },

  // 获取仓库分支列表
  fetchBranches: async (data: FetchRepositoryBranchesRequest): Promise<FetchRepositoryBranchesResponse> => {
    return request<FetchRepositoryBranchesResponse>('/projects/branches', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // 验证仓库访问权限
  validateAccess: async (data: ValidateRepositoryAccessRequest): Promise<ValidateRepositoryAccessResponse> => {
    return request<ValidateRepositoryAccessResponse>('/projects/validate-access', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },
}; 