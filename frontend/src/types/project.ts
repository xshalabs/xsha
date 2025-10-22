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
  system_prompt: string;
  repo_url: string;
  protocol: GitProtocolType;
  credential_id?: number;
  admin_id?: number;
  created_by: string;
  created_at: string;
  updated_at: string;
  task_count?: number;
  admin_count?: number;
  credential?: {
    id: number;
    name: string;
    type: string;
  };
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  system_prompt?: string;
  repo_url: string;
  protocol?: GitProtocolType;
  credential_id?: number;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  system_prompt?: string;
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
  system_prompt: string;
  repo_url: string;
  protocol: GitProtocolType;
  credential_id?: number;
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

// Project admin management types
export interface AddAdminToProjectRequest {
  admin_id: number;
}

export interface ProjectAdminsResponse {
  admins: Array<{
    id: number;
    username: string;
    name: string;
    email: string;
    role: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
    last_login_at?: string;
    avatar?: {
      uuid: string;
      original_name: string;
    };
  }>;
}

