// 应用配置常量
export const APP_CONFIG = {
  name: import.meta.env.VITE_APP_NAME || 'Sleep0',
  version: import.meta.env.VITE_APP_VERSION || '1.0.0',
} as const;

// API配置常量
export const API_CONFIG = {
  // 开发环境使用相对路径，生产环境使用完整 URL
  baseUrl: import.meta.env.VITE_API_BASE_URL || 
    (import.meta.env.DEV ? '/api/v1' : 'http://localhost:8080/api/v1'),
  timeout: 10000, // 10秒超时
} as const;

// 本地存储键名常量
export const STORAGE_KEYS = {
  authToken: 'auth_token',
  language: 'i18nextLng',
  user: 'user_info',
} as const;

// 路由路径常量
export const ROUTES = {
  home: '/',
  login: '/login',
  dashboard: '/dashboard',
  gitCredentials: '/git-credentials',
  settings: '/settings',
  profile: '/profile',
} as const;

// 支持的语言
export const SUPPORTED_LANGUAGES = [
  { code: 'zh-CN', name: '中文', flag: '🇨🇳' },
  { code: 'en-US', name: 'English', flag: '🇺🇸' },
] as const; 