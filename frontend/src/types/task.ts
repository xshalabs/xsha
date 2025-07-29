// 任务状态类型
export type TaskStatus = 'todo' | 'in_progress' | 'done' | 'cancelled';

// 任务基础接口
export interface Task {
  id: number;
  title: string;
  start_branch: string;
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

// 创建任务请求
export interface CreateTaskRequest {
  title: string;
  start_branch: string;
  project_id: number;
  dev_environment_id?: number;
  requirement_desc?: string; // 需求描述
  include_branches?: boolean; // 是否返回项目分支信息
}

// 更新任务请求（只允许更新标题）
export interface UpdateTaskRequest {
  title: string;
}



// 任务列表查询参数
export interface TaskListParams {
  page?: number;
  page_size?: number;
  project_id?: number;
  status?: TaskStatus;
}

// 任务列表响应接口
export interface TaskListResponse {
  data: {
    tasks: Task[];
    total: number;
  };
}

// 任务详情响应接口
export interface TaskDetailResponse {
  task: Task;
}

// 创建任务响应接口
export interface CreateTaskResponse {
  task: Task;
  message: string;
}

// 任务表单数据
export interface TaskFormData {
  title: string;
  start_branch: string;
  project_id: number;
  dev_environment_id?: number;
  requirement_desc?: string; // 需求描述，仅在创建时使用
  include_branches?: boolean; // 是否返回项目分支信息
} 