import { API_CONFIG, STORAGE_KEYS } from '@/lib/constants';
import { ApiError, NetworkError, logError } from '@/lib/errors';

// 环境变量配置
const getApiBaseUrl = (): string => {
  const baseUrl = import.meta.env.VITE_API_BASE_URL;
  if (!baseUrl) {
    console.warn('VITE_API_BASE_URL not found in environment variables, using default');
    return API_CONFIG.baseUrl;
  }
  return baseUrl;
};

// API 基础配置
const API_BASE_URL = getApiBaseUrl();

// API 响应类型定义
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  message: string;
  user: string;
  token: string;
}

export interface UserResponse {
  user: string;
  authenticated: boolean;
  message: string;
}

export interface ApiErrorResponse {
  error: string;
}

export interface LanguagesResponse {
  message: string;
  languages: string[];
  current: string;
}

// Token 管理
export const tokenManager = {
  getToken: (): string | null => {
    return localStorage.getItem(STORAGE_KEYS.authToken);
  },
  
  setToken: (token: string): void => {
    localStorage.setItem(STORAGE_KEYS.authToken, token);
  },
  
  removeToken: (): void => {
    localStorage.removeItem(STORAGE_KEYS.authToken);
  },
  
  isTokenPresent: (): boolean => {
    return !!localStorage.getItem(STORAGE_KEYS.authToken);
  }
};

// 获取当前语言
const getCurrentLanguage = (): string => {
  return localStorage.getItem(STORAGE_KEYS.language) || 'zh-CN';
};

// HTTP 请求工具函数
const request = async <T>(
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

// API 服务函数
export const apiService = {
  // 用户登录
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await request<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
    
    // 登录成功后保存token
    if (response.token) {
      tokenManager.setToken(response.token);
    }
    
    return response;
  },

  // 用户登出
  logout: async (): Promise<{ message: string }> => {
    try {
      const response = await request<{ message: string }>('/auth/logout', {
        method: 'POST',
      });
      
      // 登出成功后清除token
      tokenManager.removeToken();
      
      return response;
    } catch (error) {
      // 即使logout API失败，也要清除本地token
      tokenManager.removeToken();
      throw error;
    }
  },

  // 获取当前用户信息
  getCurrentUser: async (): Promise<UserResponse> => {
    return request<UserResponse>('/user/current');
  },

  // 健康检查
  healthCheck: async (): Promise<{ status: string }> => {
    const response = await fetch(`${API_BASE_URL.replace('/api/v1', '')}/health`);
    return response.json();
  },

  // 获取支持的语言列表
  getSupportedLanguages: async (): Promise<LanguagesResponse> => {
    return request<LanguagesResponse>('/languages');
  },

  // 设置语言偏好
  setLanguagePreference: async (language: string): Promise<{ message: string; language: string }> => {
    return request<{ message: string; language: string }>('/language', {
      method: 'POST',
      body: JSON.stringify({ language }),
    });
  },
}; 