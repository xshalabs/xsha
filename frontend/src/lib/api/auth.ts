import { request } from "./request";
import { tokenManager } from "./token";
import { API_BASE_URL } from "./config";
import type {
  LoginRequest,
  LoginResponse,
  UserResponse,
  ChangeOwnPasswordRequest,
} from "./types";

export const authApi = {
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await request<LoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(credentials),
    });

    if (response.token) {
      tokenManager.setToken(response.token);
    }

    return response;
  },

  logout: async (): Promise<{ message: string }> => {
    try {
      const response = await request<{ message: string }>("/auth/logout", {
        method: "POST",
      });

      tokenManager.removeToken();

      return response;
    } catch (error) {
      tokenManager.removeToken();
      throw error;
    }
  },

  getCurrentUser: async (): Promise<UserResponse> => {
    return request<UserResponse>("/user/current");
  },

  healthCheck: async (): Promise<{ status: string }> => {
    const response = await fetch(
      `${API_BASE_URL.replace("/api/v1", "")}/health`
    );
    return response.json();
  },

  changeOwnPassword: async (data: ChangeOwnPasswordRequest): Promise<{ message: string }> => {
    const token = tokenManager.getToken();
    const currentLanguage = localStorage.getItem('i18nextLng') || 'en-US';

    const config: RequestInit = {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": currentLanguage,
        ...(token && { Authorization: `Bearer ${token}` }),
      },
      body: JSON.stringify(data),
    };

    try {
      const response = await fetch(`${API_BASE_URL}/user/change-password`, config);

      if (!response.ok) {
        const errorData = await response.json();
        
        // For password change, don't treat 401 as session expiry
        // It just means the current password is incorrect
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      return response.json();
    } catch (error) {
      // Re-throw the error to be handled by the calling component
      throw error;
    }
  },

};
