export interface SystemConfig {
  id: number;
  created_at: string;
  updated_at: string;
  config_key: string;
  config_value: string;
  description: string;
  category: string;
  is_editable: boolean;
}

export interface ConfigUpdateItem {
  config_key: string;
  config_value: string;
  description?: string;
  category?: string;
  is_editable?: boolean;
}

export interface SystemConfigListResponse {
  message: string;
  configs: SystemConfig[];
}

export interface BatchUpdateConfigsRequest {
  configs: ConfigUpdateItem[];
}

export interface BatchUpdateConfigsResponse {
  message: string;
} 