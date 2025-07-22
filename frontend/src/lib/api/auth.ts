import { request } from './request';
import { tokenManager } from './token';
import { API_BASE_URL } from './config';
import type { LoginRequest, LoginResponse, UserResponse, LanguagesResponse } from './types';

export const authApi = {
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