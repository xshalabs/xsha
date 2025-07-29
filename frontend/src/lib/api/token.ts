import { STORAGE_KEYS } from "@/lib/constants";

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
  },
};

export const getCurrentLanguage = (): string => {
  return localStorage.getItem(STORAGE_KEYS.language) || "zh-CN";
};
