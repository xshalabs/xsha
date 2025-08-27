import { ApiError, NetworkError, logError } from "@/lib/errors";
import { API_BASE_URL } from "./config";
import { tokenManager, getCurrentLanguage } from "./token";
import type { ApiErrorResponse } from "./types";

export const request = async <T>(
  url: string,
  options: RequestInit = {}
): Promise<T> => {
  const token = tokenManager.getToken();
  const currentLanguage = getCurrentLanguage();

  const config: RequestInit = {
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": currentLanguage,
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    },
    ...options,
  };

  try {
    const response = await fetch(`${API_BASE_URL}${url}`, config);

    if (!response.ok) {
      const errorData: ApiErrorResponse = await response.json();
      
      // Handle 401 Unauthorized responses - token is invalid or user is deactivated
      if (response.status === 401) {
        tokenManager.removeToken();
        // Redirect to login page if not already there (using hash routing)
        if (window.location.hash !== '#/login') {
          window.location.hash = '#/login';
          return Promise.reject(new ApiError(
            errorData.error || 'Session expired, please login again',
            response.status,
            undefined,
            errorData.details
          ));
        }
      }
      
      throw new ApiError(
        errorData.error || `HTTP error! status: ${response.status}`,
        response.status,
        undefined,
        errorData.details
      );
    }

    return response.json();
  } catch (error) {
    if (error instanceof ApiError) {
      logError(error, `API request to ${url}`);
      throw error;
    }

    const networkError = new NetworkError("Failed to connect to server");
    logError(networkError, `API request to ${url}`);
    throw networkError;
  }
};
