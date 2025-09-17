export type FormType = 'input' | 'textarea' | 'switch' | 'select' | 'number' | 'password';

export interface SystemConfig {
  id: number;
  created_at: string;
  updated_at: string;
  config_key: string;
  config_value: string;
  name: string;
  description: string;
  category: string;
  form_type: FormType;
  is_editable: boolean;
  sort_order: number;
}

export interface ConfigUpdateItem {
  config_key: string;
  config_value: string;
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