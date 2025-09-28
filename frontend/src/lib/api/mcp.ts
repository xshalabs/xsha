import { request } from "./request";
import type {
  CreateMCPRequest,
  CreateMCPResponse,
  UpdateMCPRequest,
  MCPListResponse,
  MCPDetailResponse,
  MCPListParams,
  ProjectMCPsResponse,
  EnvironmentMCPsResponse,
  AddMCPToProjectRequest,
  AddMCPToEnvironmentRequest,
  MCPProjectsResponse,
  MCPEnvironmentsResponse,
  ApiResponse,
} from "@/types/mcp";

export const mcpApi = {
  // CRUD operations

  // Create a new MCP configuration
  create: async (
    data: CreateMCPRequest
  ): Promise<CreateMCPResponse> => {
    return request<CreateMCPResponse>("/mcp", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  // List MCP configurations with filtering and pagination
  list: async (params?: MCPListParams): Promise<MCPListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.name) searchParams.set("name", params.name);
    if (params?.enabled !== undefined)
      searchParams.set("enabled", params.enabled.toString());
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.page_size)
      searchParams.set("page_size", params.page_size.toString());

    const queryString = searchParams.toString();
    const url = queryString ? `/mcp?${queryString}` : "/mcp";

    return request<MCPListResponse>(url);
  },

  // Get a MCP configuration by ID
  get: async (id: number): Promise<MCPDetailResponse> => {
    const response = await request<{mcp: MCPDetailResponse}>(`/mcp/${id}`);
    return response.mcp;
  },

  // Update a MCP configuration
  update: async (
    id: number,
    data: UpdateMCPRequest
  ): Promise<ApiResponse> => {
    return request<ApiResponse>(`/mcp/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  // Delete a MCP configuration
  delete: async (id: number): Promise<ApiResponse> => {
    return request<ApiResponse>(`/mcp/${id}`, {
      method: "DELETE",
    });
  },

  // Project-related methods

  // Get MCP configurations associated with a project
  getProjectMCPs: async (
    projectId: number
  ): Promise<ProjectMCPsResponse> => {
    return request<ProjectMCPsResponse>(`/projects/${projectId}/mcp`);
  },

  // Add a MCP configuration to a project
  addToProject: async (
    projectId: number,
    data: AddMCPToProjectRequest
  ): Promise<ApiResponse> => {
    return request<ApiResponse>(`/projects/${projectId}/mcp`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  // Remove a MCP configuration from a project
  removeFromProject: async (
    projectId: number,
    mcpId: number
  ): Promise<ApiResponse> => {
    return request<ApiResponse>(
      `/projects/${projectId}/mcp/${mcpId}`,
      {
        method: "DELETE",
      }
    );
  },

  // Environment-related methods

  // Get MCP configurations associated with an environment
  getEnvironmentMCPs: async (
    environmentId: number
  ): Promise<EnvironmentMCPsResponse> => {
    return request<EnvironmentMCPsResponse>(`/environments/${environmentId}/mcp`);
  },

  // Add a MCP configuration to an environment
  addToEnvironment: async (
    environmentId: number,
    data: AddMCPToEnvironmentRequest
  ): Promise<ApiResponse> => {
    return request<ApiResponse>(`/environments/${environmentId}/mcp`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  // Remove a MCP configuration from an environment
  removeFromEnvironment: async (
    environmentId: number,
    mcpId: number
  ): Promise<ApiResponse> => {
    return request<ApiResponse>(
      `/environments/${environmentId}/mcp/${mcpId}`,
      {
        method: "DELETE",
      }
    );
  },

  // Utility methods

  // Get projects associated with a MCP configuration
  getMCPProjects: async (mcpId: number): Promise<MCPProjectsResponse> => {
    return request<MCPProjectsResponse>(`/mcp/${mcpId}/projects`);
  },

  // Get environments associated with a MCP configuration
  getMCPEnvironments: async (mcpId: number): Promise<MCPEnvironmentsResponse> => {
    return request<MCPEnvironmentsResponse>(`/mcp/${mcpId}/environments`);
  },
};