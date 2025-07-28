// 应用配置常量
export const APP_CONFIG = {
  name: import.meta.env.VITE_APP_NAME || 'XSHA',
  version: import.meta.env.VITE_APP_VERSION || '1.0.0',
} as const;

// 本地存储键
export const STORAGE_KEYS = {
  authToken: 'xsha_auth_token',
  language: 'xsha_language',
} as const;

// API配置常量
export const API_CONFIG = {
  // 开发环境使用相对路径，生产环境使用完整 URL
  baseUrl: import.meta.env.VITE_API_BASE_URL || 
    (import.meta.env.DEV ? '/api/v1' : 'http://localhost:8080/api/v1'),
  timeout: 10000, // 10秒超时
} as const;

// UI配置常量
export const UI_CONFIG = {
  pageSize: 20,
  maxRetries: 3,
} as const;

// 路由路径常量
export const ROUTES = {
  home: '/',
  login: '/login',
  dashboard: '/dashboard',
  
  // 项目管理
  projects: '/projects',
  projectCreate: '/projects/create',
  projectEdit: (id: number) => `/projects/${id}/edit`,
  
  // 开发环境
  devEnvironments: '/dev-environments',
  devEnvironmentCreate: '/dev-environments/create',
  devEnvironmentEdit: (id: number) => `/dev-environments/${id}/edit`,
  
  // Git凭据
  gitCredentials: '/git-credentials',
  gitCredentialCreate: '/git-credentials/create',
  gitCredentialEdit: (id: number) => `/git-credentials/${id}/edit`,
  
  // 任务管理
  projectTasks: (projectId: number) => `/projects/${projectId}/tasks`,
  taskCreate: (projectId: number) => `/projects/${projectId}/tasks/create`,
  taskEdit: (projectId: number, taskId: number) => `/projects/${projectId}/tasks/${taskId}/edit`,
  taskConversation: (projectId: number, taskId: number) => `/projects/${projectId}/tasks/${taskId}/conversation`,
  
  // 管理员功能
  adminLogs: '/admin/logs',
  
  // 设置和个人资料
  settings: '/settings',
  profile: '/profile',
} as const;

// 支持的语言
export const SUPPORTED_LANGUAGES = [
  { code: 'zh-CN', name: '中文', flag: '🇨🇳' },
  { code: 'en-US', name: 'English', flag: '🇺🇸' },
] as const; 