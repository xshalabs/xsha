import { request } from "./request";
import type {
  CreateProjectRequest,
  CreateProjectResponse,
  UpdateProjectRequest,
  ProjectListResponse,
  ProjectDetailResponse,
  CompatibleCredentialsResponse,
  ProjectListParams,
  FetchRepositoryBranchesRequest,
  FetchRepositoryBranchesResponse,
} from "@/types/project";

export const projectsApi = {
  create: async (
    data: CreateProjectRequest
  ): Promise<CreateProjectResponse> => {
    return request<CreateProjectResponse>("/projects", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  list: async (params?: ProjectListParams): Promise<ProjectListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.name) searchParams.set("name", params.name);
    if (params?.protocol) searchParams.set("protocol", params.protocol);
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.page_size)
      searchParams.set("page_size", params.page_size.toString());
    if (params?.sort_by) searchParams.set("sort_by", params.sort_by);
    if (params?.sort_direction) searchParams.set("sort_direction", params.sort_direction);

    const queryString = searchParams.toString();
    const url = queryString ? `/projects?${queryString}` : "/projects";

    return request<ProjectListResponse>(url);
  },

  get: async (id: number): Promise<ProjectDetailResponse> => {
    return request<ProjectDetailResponse>(`/projects/${id}`);
  },

  update: async (
    id: number,
    data: UpdateProjectRequest
  ): Promise<{ message: string }> => {
    return request<{ message: string }>(`/projects/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/projects/${id}`, {
      method: "DELETE",
    });
  },

  getCompatibleCredentials: async (
    repoUrl: string
  ): Promise<CompatibleCredentialsResponse> => {
    return request<CompatibleCredentialsResponse>(
      `/projects/credentials?repo_url=${encodeURIComponent(repoUrl)}`
    );
  },


  fetchBranches: async (
    data: FetchRepositoryBranchesRequest
  ): Promise<FetchRepositoryBranchesResponse> => {
    return request<FetchRepositoryBranchesResponse>("/projects/branches", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

};
