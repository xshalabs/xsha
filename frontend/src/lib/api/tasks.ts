import { request } from "./request";
import type {
  CreateTaskRequest,
  CreateTaskResponse,
  UpdateTaskRequest,
  TaskListResponse,
  TaskDetailResponse,
  TaskListParams,
} from "@/types/task";

export const tasksApi = {
  create: async (data: CreateTaskRequest): Promise<CreateTaskResponse> => {
    return request<CreateTaskResponse>("/tasks", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  list: async (params?: TaskListParams): Promise<TaskListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.page_size)
      searchParams.set("page_size", params.page_size.toString());
    if (params?.project_id)
      searchParams.set("project_id", params.project_id.toString());
    if (params?.status) searchParams.set("status", params.status);
    if (params?.title) searchParams.set("title", params.title);
    if (params?.branch) searchParams.set("branch", params.branch);
    if (params?.dev_environment_id)
      searchParams.set(
        "dev_environment_id",
        params.dev_environment_id.toString()
      );

    const queryString = searchParams.toString();
    const url = queryString ? `/tasks?${queryString}` : "/tasks";

    return request<TaskListResponse>(url);
  },

  get: async (id: number): Promise<TaskDetailResponse> => {
    return request<TaskDetailResponse>(`/tasks/${id}`);
  },

  update: async (
    id: number,
    data: UpdateTaskRequest
  ): Promise<{ message: string }> => {
    return request<{ message: string }>(`/tasks/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/tasks/${id}`, {
      method: "DELETE",
    });
  },

  batchUpdateStatus: async (data: {
    task_ids: number[];
    status: string;
  }): Promise<{
    message: string;
    data: {
      success_count: number;
      failed_count: number;
      success_ids: number[];
      failed_ids: number[];
    };
  }> => {
    return request(`/tasks/batch/status`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  getTaskGitDiff: async (
    taskId: number,
    params?: TaskGitDiffParams
  ): Promise<TaskGitDiffResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.include_content) {
      searchParams.set("include_content", "true");
    }

    const url = `/tasks/${taskId}/git-diff${
      searchParams.toString() ? `?${searchParams.toString()}` : ""
    }`;
    return request<TaskGitDiffResponse>(url, {
      method: "GET",
    });
  },

  getTaskGitDiffFile: async (
    taskId: number,
    params: TaskGitDiffFileParams
  ): Promise<TaskGitDiffFileResponse> => {
    const searchParams = new URLSearchParams();
    searchParams.set("file_path", params.file_path);

    const url = `/tasks/${taskId}/git-diff/file?${searchParams.toString()}`;
    return request<TaskGitDiffFileResponse>(url, {
      method: "GET",
    });
  },

  pushTaskBranch: async (
    taskId: number,
    forcePush: boolean = false
  ): Promise<{
    message: string;
    data: {
      output: string;
    };
  }> => {
    return request(`/tasks/${taskId}/push`, {
      method: "POST",
      body: JSON.stringify({ force_push: forcePush }),
    });
  },
};

export interface GitDiffFile {
  path: string;
  status: "added" | "modified" | "deleted" | "renamed";
  additions: number;
  deletions: number;
  is_binary: boolean;
  old_path?: string;
  diff_content?: string;
}

export interface GitDiffSummary {
  total_files: number;
  total_additions: number;
  total_deletions: number;
  files: GitDiffFile[];
  commits_behind: number;
  commits_ahead: number;
}

export interface TaskGitDiffParams {
  include_content?: boolean;
}

export interface TaskGitDiffFileParams {
  file_path: string;
}

export interface TaskGitDiffResponse {
  data: GitDiffSummary;
}

export interface TaskGitDiffFileResponse {
  data: {
    file_path: string;
    diff_content: string;
  };
}
