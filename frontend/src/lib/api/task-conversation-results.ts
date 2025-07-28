import { request } from './request';
import type {
  CreateResultRequest,
  CreateResultResponse,
  ProcessResultFromJSONRequest,
  ProcessResultResponse,
  UpdateResultRequest,
  ResultListResponse,
  ResultDetailResponse,
  TaskStatsResponse,
  ProjectStatsResponse,
  ResultListByTaskParams,
  ResultListByProjectParams
} from '@/types/task-conversation-result';

export const taskConversationResultsApi = {
  // 创建结果
  create: async (data: CreateResultRequest): Promise<CreateResultResponse> => {
    return request<CreateResultResponse>('/conversation-results', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // 从JSON处理结果
  processFromJSON: async (data: ProcessResultFromJSONRequest): Promise<ProcessResultResponse> => {
    return request<ProcessResultResponse>('/conversation-results/process-json', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // 根据任务ID获取结果列表
  listByTask: async (params: ResultListByTaskParams): Promise<ResultListResponse> => {
    const searchParams = new URLSearchParams();
    searchParams.set('task_id', params.task_id.toString());
    if (params.page) searchParams.set('page', params.page.toString());
    if (params.page_size) searchParams.set('page_size', params.page_size.toString());
    
    const queryString = searchParams.toString();
    return request<ResultListResponse>(`/conversation-results?${queryString}`);
  },

  // 根据项目ID获取结果列表
  listByProject: async (params: ResultListByProjectParams): Promise<ResultListResponse> => {
    const searchParams = new URLSearchParams();
    searchParams.set('project_id', params.project_id.toString());
    if (params.page) searchParams.set('page', params.page.toString());
    if (params.page_size) searchParams.set('page_size', params.page_size.toString());
    
    const queryString = searchParams.toString();
    return request<ResultListResponse>(`/conversation-results/by-project?${queryString}`);
  },

  // 获取单个结果详情
  get: async (id: number): Promise<ResultDetailResponse> => {
    return request<ResultDetailResponse>(`/conversation-results/${id}`);
  },

  // 根据对话ID获取结果
  getByConversationId: async (conversationId: number): Promise<ResultDetailResponse> => {
    return request<ResultDetailResponse>(`/conversation-results/by-conversation/${conversationId}`);
  },

  // 更新结果
  update: async (id: number, data: UpdateResultRequest): Promise<{ message: string }> => {
    return request<{ message: string }>(`/conversation-results/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // 删除结果
  delete: async (id: number): Promise<{ message: string }> => {
    return request<{ message: string }>(`/conversation-results/${id}`, {
      method: 'DELETE',
    });
  },

  // 获取任务统计信息
  getTaskStats: async (taskId: number): Promise<TaskStatsResponse> => {
    return request<TaskStatsResponse>(`/stats/tasks/${taskId}`);
  },

  // 获取项目统计信息
  getProjectStats: async (projectId: number): Promise<ProjectStatsResponse> => {
    return request<ProjectStatsResponse>(`/stats/projects/${projectId}`);
  },
}; 