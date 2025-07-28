// 结果类型
export type ResultType = 'result';

// 结果子类型
export type ResultSubtype = 'success' | 'error';

// 使用统计接口
export interface UsageStats {
  input_tokens: number;
  cache_creation_input_tokens: number;
  cache_read_input_tokens: number;
  output_tokens: number;
  server_tool_use?: {
    web_search_requests: number;
  };
  service_tier?: string;
}

// 任务对话结果基础接口
export interface TaskConversationResult {
  id: number;
  conversation_id: number;
  type: ResultType;
  subtype: ResultSubtype;
  is_error: boolean;
  duration_ms: number;
  duration_api_ms: number;
  num_turns: number;
  result: string;
  session_id: string;
  total_cost_usd: number;
  usage: string; // JSON字符串，包含UsageStats
  created_at: string;
  updated_at: string;
  conversation?: {
    id: number;
    content: string;
    task?: {
      id: number;
      title: string;
      project?: {
        id: number;
        name: string;
      };
    };
  };
}

// 解析后的使用统计
export interface ParsedTaskConversationResult extends Omit<TaskConversationResult, 'usage'> {
  usage: UsageStats;
}

// 创建结果请求
export interface CreateResultRequest {
  conversation_id: number;
  result_data: {
    type: string;
    subtype: string;
    is_error: boolean;
    duration_ms?: number;
    duration_api_ms?: number;
    num_turns?: number;
    result: string;
    session_id: string;
    total_cost_usd?: number;
    usage?: UsageStats;
  };
}

// 从JSON处理结果请求
export interface ProcessResultFromJSONRequest {
  conversation_id: number;
  json_data: string;
}

// 更新结果请求
export interface UpdateResultRequest {
  updates: {
    type?: string;
    subtype?: string;
    is_error?: boolean;
    duration_ms?: number;
    duration_api_ms?: number;
    num_turns?: number;
    result?: string;
    session_id?: string;
    total_cost_usd?: number;
    usage?: string;
  };
}

// 结果列表查询参数
export interface ResultListByTaskParams {
  task_id: number;
  page?: number;
  page_size?: number;
}

export interface ResultListByProjectParams {
  project_id: number;
  page?: number;
  page_size?: number;
}

// 统计信息接口
export interface TaskStats {
  success_rate: number;
  total_cost_usd: number;
  average_duration_ms: number;
}

export interface ProjectStats {
  total_conversations: number;
  success_count: number;
  error_count: number;
  success_rate: number;
  total_cost_usd: number;
  average_duration_ms: number;
}

// API响应类型
export interface CreateResultResponse {
  message: string;
  data: TaskConversationResult;
}

export interface ProcessResultResponse {
  message: string;
  data: TaskConversationResult;
}

export interface ResultListResponse {
  message: string;
  data: {
    items: TaskConversationResult[];
    total: number;
    page: number;
    page_size: number;
  };
}

export interface ResultDetailResponse {
  message: string;
  data: TaskConversationResult;
}

export interface TaskStatsResponse {
  message: string;
  data: TaskStats;
}

export interface ProjectStatsResponse {
  message: string;
  data: ProjectStats;
}

// 表单数据接口
export interface ResultFormData {
  type: ResultType;
  subtype: ResultSubtype;
  is_error: boolean;
  duration_ms: number;
  duration_api_ms: number;
  num_turns: number;
  result: string;
  session_id: string;
  total_cost_usd: number;
  usage: UsageStats;
}

// 工具函数类型
export interface ResultUtils {
  parseUsage: (usageString: string) => UsageStats | null;
  formatDuration: (ms: number) => string;
  formatCost: (cost: number) => string;
  getSuccessRateColor: (rate: number) => string;
} 