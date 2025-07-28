// 配置和工具
export { API_BASE_URL, getApiBaseUrl } from './config';
export { tokenManager, getCurrentLanguage } from './token';
export { request } from './request';

// 类型定义
export type {
  LoginRequest,
  LoginResponse,
  UserResponse,
  ApiErrorResponse,
  LanguagesResponse
} from './types';

// API 模块导入
import { authApi } from './auth';
import { gitCredentialsApi } from './git-credentials';
import { adminLogsApi } from './admin-logs';
import { projectsApi } from './projects';
import { devEnvironmentsApi } from './dev-environments';
import { tasksApi } from './tasks';
import { taskConversationsApi } from './task-conversations';
import { taskConversationResultsApi } from './task-conversation-results';
import { taskExecutionLogsApi } from './task-execution-logs';

// API 模块导出
export { authApi, gitCredentialsApi, adminLogsApi, projectsApi, devEnvironmentsApi, tasksApi, taskConversationsApi, taskConversationResultsApi, taskExecutionLogsApi };

// 兼容性导出 - 保持原有的 apiService 结构
export const apiService = {
  // 认证相关
  login: authApi.login,
  logout: authApi.logout,
  getCurrentUser: authApi.getCurrentUser,
  healthCheck: authApi.healthCheck,
  getSupportedLanguages: authApi.getSupportedLanguages,
  setLanguagePreference: authApi.setLanguagePreference,

  // Git 凭据
  gitCredentials: gitCredentialsApi,

  // 管理员日志
  adminLogs: adminLogsApi,

  // 项目
  projects: projectsApi,

  // 开发环境
  devEnvironments: devEnvironmentsApi,

  // 任务管理
  tasks: tasksApi,

  // 任务对话
  taskConversations: taskConversationsApi,

  // 任务对话结果
  taskConversationResults: taskConversationResultsApi,

  // 任务执行日志
  taskExecutionLogs: taskExecutionLogsApi,
}; 