// 任务状态类型
export type TaskStatus = 'todo' | 'in_progress' | 'done' | 'cancelled';

// 任务基础接口
export interface Task {
  id: number;
  title: string;
  description: string;
  start_branch: string;
  status: TaskStatus;
  has_pull_request: boolean;
  project_id: number;
  created_by: string;
  created_at: string;
  updated_at: string;
  project?: {
    id: number;
    name: string;
  };
}

// 创建任务请求
export interface CreateTaskRequest {
  title: string;
  description?: string;
  start_branch: string;
  project_id: number;
}

// 更新任务请求
export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  start_branch?: string;
}

// 更新任务状态请求
export interface UpdateTaskStatusRequest {
  status: TaskStatus;
}

// 更新PR状态请求
export interface UpdatePullRequestStatusRequest {
  has_pull_request: boolean;
}

// 任务列表查询参数
export interface TaskListParams {
  page?: number;
  page_size?: number;
  project_id?: number;
  status?: TaskStatus;
}

// 任务统计
export interface TaskStats {
  total: number;
  todo: number;
  in_progress: number;
  done: number;
  cancelled: number;
}

// API响应类型
export interface CreateTaskResponse {
  message: string;
  data: Task;
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

export interface TaskStatsResponse {
  message: string;
  data: TaskStats;
}

// 任务表单数据
export interface TaskFormData {
  title: string;
  description: string;
  start_branch: string;
  project_id: number;
} 