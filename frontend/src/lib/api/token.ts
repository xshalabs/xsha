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
  // First try to get from i18next storage (primary source)
  const i18nextLang = localStorage.getItem("i18nextLng");
  if (i18nextLang && (i18nextLang === "zh-CN" || i18nextLang === "en-US")) {
    return i18nextLang;
  }
  
  // Fallback to our own storage, then default
  return localStorage.getItem(STORAGE_KEYS.language) || "zh-CN";
};
