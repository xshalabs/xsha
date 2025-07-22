import { ApiError, NetworkError, logError } from '@/lib/errors';
import { API_BASE_URL } from './config';
import { tokenManager, getCurrentLanguage } from './token';
import type { ApiErrorResponse } from './types';

// HTTP 请求工具函数
export const request = async <T>(
  url: string, 
  options: RequestInit = {}
): Promise<T> => {
  const token = tokenManager.getToken();
  const currentLanguage = getCurrentLanguage();
  
  const config: RequestInit = {
    headers: {
      'Content-Type': 'application/json',
      'Accept-Language': currentLanguage,
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    },
    ...options,
  };

  try {
    const response = await fetch(`${API_BASE_URL}${url}`, config);
    
    if (!response.ok) {
      const errorData: ApiErrorResponse = await response.json();
      throw new ApiError(
        errorData.error || `HTTP error! status: ${response.status}`,
        response.status
      );
    }
    
    return response.json();
  } catch (error) {
    if (error instanceof ApiError) {
      logError(error, `API request to ${url}`);
      throw error;
    }
    
    // 网络错误或其他错误
    const networkError = new NetworkError('Failed to connect to server');
    logError(networkError, `API request to ${url}`);
    throw networkError;
  }
}; 