import { request } from './request';
import type {
  CreateTaskRequest,
  CreateTaskResponse,
  UpdateTaskRequest,
  TaskListResponse,
  TaskDetailResponse,
  TaskListParams
} from '@/types/task';

export const tasksApi = {
  // 创建任务
  create: async (data: CreateTaskRequest): Promise<CreateTaskResponse> => {
    return request<CreateTaskResponse>('/tasks', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // 获取任务列表
  list: async (params?: TaskListParams): Promise<TaskListResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.page_size) searchParams.set('page_size', params.page_size.toString());
    if (params?.project_id) searchParams.set('project_id', params.project_id.toString());
    if (params?.status) searchParams.set('status', params.status);
    if (params?.title) searchParams.set('title', params.title);
    if (params?.branch) searchParams.set('branch', params.branch);
    if (params?.dev_environment_id) searchParams.set('dev_environment_id', params.dev_environment_id.toString());
    
    const queryString = searchParams.toString();
    const url = queryString ? `/tasks?${queryString}` : '/tasks';
    
    return request<TaskListResponse>(url);
  },

  // 获取单个任务详情
  get: async (id: number): Promise<TaskDetailResponse> => {
    return request<TaskDetailResponse>(`/tasks/${id}`);
  },

  // 更新任务
  update: async (id: number, data: UpdateTaskRequest): Promise<{ message: string }> => {
    return request<{ message: string }>(`/tasks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // 删除任务
  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/tasks/${id}`, {
      method: 'DELETE',
    });
  },
}; 