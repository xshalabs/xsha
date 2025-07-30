export type ConversationStatus =
  | "pending"
  | "running"
  | "success"
  | "failed"
  | "cancelled";

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

export interface CreateConversationRequest {
  task_id: number;
  content: string;
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

export interface LatestConversationResponse {
  message: string;
  data: TaskConversation;
}

export interface ConversationFormData {
  content: string;
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
