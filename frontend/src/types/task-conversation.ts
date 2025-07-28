// 对话状态类型
export type ConversationStatus = 'pending' | 'running' | 'success' | 'failed' | 'cancelled';

// 任务对话基础接口
export interface TaskConversation {
  id: number;
  task_id: number;
  content: string;
  status: ConversationStatus;
  commit_hash: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  task?: {
    id: number;
    title: string;
  };
}

// 创建对话请求
export interface CreateConversationRequest {
  task_id: number;
  content: string;
}

// 更新对话请求
export interface UpdateConversationRequest {
  content?: string;
}



// 对话列表查询参数
export interface ConversationListParams {
  task_id: number;
  page?: number;
  page_size?: number;
}

// API响应类型
export interface CreateConversationResponse {
  message: string;
  data: TaskConversation;
}

export interface ConversationListResponse {
  message: string;
  data: {
    conversations: TaskConversation[];
    total: number;
    page: number;
    page_size: number;
  };
}

export interface ConversationDetailResponse {
  message: string;
  data: TaskConversation;
}

export interface LatestConversationResponse {
  message: string;
  data: TaskConversation;
}

// 对话表单数据
export interface ConversationFormData {
  content: string;
} 