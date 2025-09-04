import { request } from "./request";
import type {
  CreateTaskApiRequest,
  CreateTaskResponse,
  UpdateTaskRequest,
  Task,
} from "@/types/task";

export const tasksApi = {
  create: async (
    projectId: number,
    data: CreateTaskApiRequest
  ): Promise<CreateTaskResponse> => {
    return request<CreateTaskResponse>(`/projects/${projectId}/tasks`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  update: async (
    projectId: number,
    taskId: number,
    data: UpdateTaskRequest
  ): Promise<{ message: string }> => {
    return request<{ message: string }>(
      `/projects/${projectId}/tasks/${taskId}`,
      {
        method: "PUT",
        body: JSON.stringify(data),
      }
    );
  },

  delete: async (
    projectId: number,
    taskId: number
  ): Promise<{ message: string }> => {
    return request<{ message: string }>(
      `/projects/${projectId}/tasks/${taskId}`,
      {
        method: "DELETE",
      }
    );
  },

  batchUpdateStatus: async (
    projectId: number,
    data: {
      task_ids: number[];
      status: string;
    }
  ): Promise<{
    message: string;
    data: {
      success_count: number;
      failed_count: number;
      success_ids: number[];
      failed_ids: number[];
    };
  }> => {
    return request(`/projects/${projectId}/tasks/batch/status`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  getTaskGitDiff: async (
    projectId: number,
    taskId: number,
    params?: TaskGitDiffParams
  ): Promise<TaskGitDiffResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.include_content) {
      searchParams.set("include_content", "true");
    }

    const url = `/projects/${projectId}/tasks/${taskId}/git-diff${
      searchParams.toString() ? `?${searchParams.toString()}` : ""
    }`;
    return request<TaskGitDiffResponse>(url, {
      method: "GET",
    });
  },

  getTaskGitDiffFile: async (
    projectId: number,
    taskId: number,
    params: TaskGitDiffFileParams
  ): Promise<TaskGitDiffFileResponse> => {
    const searchParams = new URLSearchParams();
    searchParams.set("file_path", params.file_path);

    const url = `/projects/${projectId}/tasks/${taskId}/git-diff/file?${searchParams.toString()}`;
    return request<TaskGitDiffFileResponse>(url, {
      method: "GET",
    });
  },

  pushTaskBranch: async (
    projectId: number,
    taskId: number,
    forcePush: boolean = false
  ): Promise<{
    message: string;
    data: {
      output: string;
    };
  }> => {
    return request(`/projects/${projectId}/tasks/${taskId}/push`, {
      method: "POST",
      body: JSON.stringify({ force_push: forcePush }),
    });
  },

  getKanbanTasks: async (projectId: number): Promise<KanbanTasksResponse> => {
    return request<KanbanTasksResponse>(`/projects/${projectId}/kanban`);
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

export interface KanbanTasksResponse {
  message: string;
  data: {
    todo: Task[];
    in_progress: Task[];
    done: Task[];
    cancelled: Task[];
  };
}
