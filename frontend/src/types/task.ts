export type TaskStatus = "todo" | "in_progress" | "done" | "cancelled";

export interface Task {
  id: number;
  title: string;
  start_branch: string;
  work_branch: string;
  status: TaskStatus;
  has_pull_request: boolean;
  workspace_path: string;
  project_id: number;
  dev_environment_id?: number;
  created_by: string;
  created_at: string;
  updated_at: string;
  project?: {
    id: number;
    name: string;
  };
  dev_environment?: {
    id: number;
    name: string;
    type: string;
    status: string;
  };
}

export interface CreateTaskRequest {
  title: string;
  start_branch: string;
  project_id: number;
  dev_environment_id?: number;
  requirement_desc?: string;
  include_branches?: boolean;
}

export interface UpdateTaskRequest {
  title: string;
}

export interface TaskListParams {
  page?: number;
  page_size?: number;
  project_id?: number;
  status?: TaskStatus;
  title?: string;
  branch?: string;
  dev_environment_id?: number;
}

export interface TaskListResponse {
  message: string;
  data: {
    tasks: Task[];
    total: number;
    page: number;
    page_size: number;
  };
}

export interface TaskDetailResponse {
  message: string;
  data: Task;
}

export interface CreateTaskResponse {
  message: string;
  data: Task;
}

export interface TaskFormData {
  title: string;
  start_branch: string;
  project_id: number;
  dev_environment_id?: number;
  requirement_desc?: string;
  include_branches?: boolean;
}

export interface BatchUpdateStatusRequest {
  task_ids: number[];
  status: TaskStatus;
}

export interface BatchUpdateStatusResponse {
  message: string;
  data: {
    success_count: number;
    failed_count: number;
    success_ids: number[];
    failed_ids: number[];
  };
}
