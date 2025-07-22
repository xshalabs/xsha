import { request } from './request';
import type {
  CreateTaskRequest,
  CreateTaskResponse,
  UpdateTaskRequest,
  UpdateTaskStatusRequest,
  UpdatePullRequestStatusRequest,
  TaskListResponse,
  TaskDetailResponse,
  TaskStatsResponse,
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

  // 更新任务状态
  updateStatus: async (id: number, data: UpdateTaskStatusRequest): Promise<{ message: string }> => {
    return request<{ message: string }>(`/tasks/${id}/status`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // 更新PR状态
  updatePullRequestStatus: async (id: number, data: UpdatePullRequestStatusRequest): Promise<{ message: string }> => {
    return request<{ message: string }>(`/tasks/${id}/pr-status`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // 获取任务统计
  getStats: async (projectId: number): Promise<TaskStatsResponse> => {
    return request<TaskStatsResponse>(`/tasks/stats?project_id=${projectId}`);
  },
}; 