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

export interface UpdateSystemConfigRequest {
  config_value?: string;
  description?: string;
  category?: string;
  is_editable?: boolean;
}

export interface SystemConfigDetailResponse {
  config: SystemConfig;
}

export interface SystemConfigListResponse {
  message: string;
  configs: SystemConfig[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface UpdateSystemConfigResponse {
  message: string;
}



export interface DevEnvironmentType {
  name: string;
  image: string;
}

export interface DevEnvironmentTypesResponse {
  env_types: DevEnvironmentType[];
}

export interface UpdateDevEnvironmentTypesRequest {
  env_types: DevEnvironmentType[];
}

export interface UpdateDevEnvironmentTypesResponse {
  message: string;
} 