export type ResultType = "result";
export type ResultSubtype = "success" | "error";

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
  usage: string;
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

export interface ParsedTaskConversationResult
  extends Omit<TaskConversationResult, "usage"> {
  usage: UsageStats;
}

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

export interface ResultUtils {
  parseUsage: (usageString: string) => UsageStats | null;
  formatDuration: (ms: number) => string;
  formatCost: (cost: number) => string;
  getSuccessRateColor: (rate: number) => string;
}
