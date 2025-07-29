import { API_CONFIG } from "@/lib/constants";

export const getApiBaseUrl = (): string => {
  const baseUrl = import.meta.env.VITE_API_BASE_URL;
  if (!baseUrl) {
    console.warn(
      "VITE_API_BASE_URL not found in environment variables, using default"
    );
    return API_CONFIG.baseUrl;
  }
  return baseUrl;
};

export const API_BASE_URL = getApiBaseUrl();
