import { request } from "./request";
import { tokenManager } from "./token";
import { API_BASE_URL } from "./config";
import type {
  LoginRequest,
  LoginResponse,
  UserResponse,
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


};
