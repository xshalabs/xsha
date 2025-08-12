export const GitProtocolType = {
  HTTPS: "https",
  SSH: "ssh",
} as const;

export type GitProtocolType =
  (typeof GitProtocolType)[keyof typeof GitProtocolType];

export interface Project {
  id: number;
  name: string;
  description: string;
  repo_url: string;
  protocol: GitProtocolType;
  credential_id?: number;
  created_by: string;
  created_at: string;
  updated_at: string;
  task_count?: number;
  credential?: {
    id: number;
    name: string;
    type: string;
  };
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  repo_url: string;
  protocol: GitProtocolType;
  credential_id?: number;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  repo_url?: string;
  protocol?: GitProtocolType;
  credential_id?: number;
}

export interface ProjectListResponse {
  message: string;
  projects: Project[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ProjectDetailResponse {
  project: Project;
}

export interface CreateProjectResponse {
  message: string;
  project: Project;
}

export interface CompatibleCredentialsResponse {
  message: string;
  credentials: Array<{
    id: number;
    name: string;
    type: string;
    username: string;
  }>;
}

export interface ProjectListParams {
  name?: string;
  protocol?: GitProtocolType;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_direction?: 'asc' | 'desc';
}

export interface ProjectFormData {
  name: string;
  description: string;
  repo_url: string;
  protocol: GitProtocolType;
  credential_id?: number;
}

export interface ParseRepositoryURLRequest {
  repo_url: string;
}

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

export interface FetchRepositoryBranchesRequest {
  repo_url: string;
  credential_id?: number;
}

export interface FetchRepositoryBranchesResponse {
  message: string;
  result: {
    can_access: boolean;
    error_message: string;
    branches: string[];
  };
}

export interface ValidateRepositoryAccessRequest {
  repo_url: string;
  credential_id?: number;
}

export interface ValidateRepositoryAccessResponse {
  message: string;
  can_access: boolean;
  error?: string;
}
