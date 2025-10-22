import type { Admin } from "@/lib/api/types";

export type ProviderType = "claude-code";

export interface Provider {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  type: ProviderType;
  config: string;
  admin_id?: number;
  admin?: Admin;
  created_by: string;
}

export interface CreateProviderRequest {
  name: string;
  description?: string;
  type: ProviderType;
  config: string;
}

export interface UpdateProviderRequest {
  name?: string;
  description?: string;
  config?: string;
}

export interface CreateProviderResponse {
  message: string;
  provider: Provider;
}

export interface ProviderDetailResponse {
  provider: Provider;
}

export interface ProviderListResponse {
  message: string;
  providers: Provider[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ProviderListParams {
  page?: number;
  page_size?: number;
  name?: string;
  type?: ProviderType;
}

export interface ProviderTypesResponse {
  types: ProviderType[];
}

// ProviderSelection represents provider information for selection dropdowns
// SECURITY: This type intentionally excludes the config field to prevent sensitive data exposure
export interface ProviderSelection {
  id: number;
  name: string;
  description: string;
  type: ProviderType;
}
