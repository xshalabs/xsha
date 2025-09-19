export const NotifierType = {
  WECHAT_WORK: "wechat_work",
  DINGTALK: "dingtalk",
  FEISHU: "feishu",
  SLACK: "slack",
  DISCORD: "discord",
  WEBHOOK: "webhook",
} as const;

export type NotifierType = (typeof NotifierType)[keyof typeof NotifierType];

export interface Notifier {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  type: NotifierType;
  config: string; // JSON string from backend
  is_enabled: boolean;
  admin_id?: number;
  admin?: MinimalAdminResponse;
  created_by: string;
  projects?: Project[];
}

export interface MinimalAdminResponse {
  id: number;
  username: string;
  name: string;
  email: string;
  avatar?: AdminAvatarMinimal;
}

export interface AdminAvatarMinimal {
  uuid: string;
  original_name: string;
}

export interface Project {
  id: number;
  name: string;
  protocol: string;
  clone_url: string;
}

export interface NotifierConfig {
  [key: string]: string | number | boolean | Record<string, unknown> | undefined;
}

// Specific config types for each notifier
export interface WeChatWorkConfig {
  webhook_url: string;
  timeout?: string;
}

export interface DingTalkConfig {
  webhook_url: string;
  secret?: string;
  timeout?: string;
}

export interface FeishuConfig {
  webhook_url: string;
  secret?: string;
  timeout?: string;
}

export interface SlackConfig {
  webhook_url: string;
  timeout?: string;
}

export interface DiscordConfig {
  webhook_url: string;
  timeout?: string;
}

export interface WebhookConfig {
  url: string;
  method?: string;
  headers?: Record<string, string>;
  body_template?: string;
  timeout?: string;
}

export interface CreateNotifierRequest {
  name: string;
  description: string;
  type: NotifierType;
  config: NotifierConfig;
}

export interface UpdateNotifierRequest {
  name?: string;
  description?: string;
  config?: NotifierConfig;
  is_enabled?: boolean;
}

export interface NotifierListResponse {
  data: Notifier[];
  total: number;
  page: number;
  page_size: number;
}

export interface NotifierDetailResponse {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  type: NotifierType;
  config: string; // JSON string from backend, matches actual response
  is_enabled: boolean;
  admin_id?: number;
  admin?: MinimalAdminResponse;
  created_by: string;
}

export interface CreateNotifierResponse {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  type: NotifierType;
  config: NotifierConfig;
  is_enabled: boolean;
  admin_id?: number;
  admin?: MinimalAdminResponse;
  created_by: string;
}

export interface NotifierListParams {
  name?: string;
  type?: NotifierType;
  is_enabled?: boolean;
  page?: number;
  page_size?: number;
}

export interface NotifierTypeInfo {
  type: NotifierType;
  name: string;
  description: string;
  config_schema: {
    name: string;
    type: string;
    required: boolean;
    default?: string;
    description: string;
  }[];
}

export interface NotifierTypesResponse {
  data: NotifierTypeInfo[];
}

export interface ProjectNotifiersResponse {
  data: Notifier[];
}

export interface AddNotifierToProjectRequest {
  notifier_id: number;
}

export interface NotifierFormData {
  name: string;
  description: string;
  type: NotifierType;
  config: NotifierConfig;
}

// Test notification response
export interface TestNotifierResponse {
  message: string;
}

// Common API response format
export interface ApiResponse<T = unknown> {
  message?: string;
  data?: T;
  error?: string;
}