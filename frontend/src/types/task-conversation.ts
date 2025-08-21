import type { TaskExecutionLog } from "./task-execution-log";

export type ConversationStatus =
  | "pending"
  | "running"
  | "success"
  | "failed"
  | "cancelled";

export interface TaskConversationResult {
  id: number;
  conversation_id: number;
  type: string;
  subtype: string;
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
}

export interface TaskConversation {
  id: number;
  task_id: number;
  content: string;
  status: ConversationStatus;
  execution_time?: string; // ISO 8601 date string
  commit_hash: string;
  env_params: string; // JSON string containing environment parameters
  created_by: string;
  created_at: string;
  updated_at: string;
  task?: {
    id: number;
    title: string;
  };
}

export interface CreateConversationRequest {
  task_id: number;
  content: string;
  execution_time?: string; // ISO 8601 date string
  env_params?: string; // JSON string containing environment parameters
  attachment_ids?: number[]; // Optional array of attachment IDs
}

export interface UpdateConversationRequest {
  content?: string;
}

export interface ConversationListParams {
  task_id: number;
  page?: number;
  page_size?: number;
}

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

export interface ConversationWithResultAndLogResponse {
  message: string;
  data: {
    conversation: TaskConversation;
    result?: TaskConversationResult;
    execution_log?: TaskExecutionLog;
  };
}

export interface LatestConversationResponse {
  message: string;
  data: TaskConversation;
}

export interface ConversationFormData {
  content: string;
  execution_time?: Date; // Date object for form handling
  model?: string; // Model selection for claude-code environments
}

// Git diff types for conversations
export interface ConversationGitDiffParams {
  include_content?: boolean;
}

export interface ConversationGitDiffFileParams {
  file_path: string;
}

export interface ConversationGitDiffResponse {
  data: GitDiffSummary;
}

export interface ConversationGitDiffFileResponse {
  data: {
    file_path: string;
    diff_content: string;
  };
}

// Import GitDiffSummary type from tasks
export interface GitDiffFile {
  path: string;
  status: 'added' | 'modified' | 'deleted' | 'renamed';
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

// Claude model options for claude-code environments
export type ClaudeModel = "default" | "sonnet" | "opus";

export interface ClaudeModelOption {
  value: ClaudeModel;
  label: string;
}

// Environment parameters interface
export interface EnvironmentParams {
  model?: ClaudeModel;
}
