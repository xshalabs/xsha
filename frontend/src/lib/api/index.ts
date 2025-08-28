export { API_BASE_URL, getApiBaseUrl } from "./config";
export { tokenManager, getCurrentLanguage } from "./token";
export { request } from "./request";

export type {
  LoginRequest,
  LoginResponse,
  UserResponse,
  ApiErrorResponse,
  Admin,
  CreateAdminRequest,
  UpdateAdminRequest,
  ChangePasswordRequest,
  ChangeOwnPasswordRequest,
  AdminListResponse,
  AdminResponse,
  CreateAdminResponse,
} from "./types";

import { authApi } from "./auth";
import { adminApi } from "./admin";
import { gitCredentialsApi } from "./credentials";
import { adminLogsApi } from "./admin-logs";
import { projectsApi } from "./projects";
import { devEnvironmentsApi } from "./environments";
import { tasksApi } from "./tasks";
import { taskConversationsApi } from "./task-conversations";
import { taskExecutionLogsApi } from "./task-execution-logs";
import { dashboardApi } from "./dashboard";
import { attachmentApi } from "./attachments";

export {
  authApi,
  adminApi,
  gitCredentialsApi,
  adminLogsApi,
  projectsApi,
  devEnvironmentsApi,
  tasksApi,
  taskConversationsApi,
  taskExecutionLogsApi,
  dashboardApi,
  attachmentApi,
};

export const apiService = {
  login: authApi.login,
  logout: authApi.logout,
  getCurrentUser: authApi.getCurrentUser,
  changeOwnPassword: authApi.changeOwnPassword,
  healthCheck: authApi.healthCheck,

  admin: adminApi,

  gitCredentials: gitCredentialsApi,

  adminLogs: adminLogsApi,

  projects: projectsApi,

  devEnvironments: devEnvironmentsApi,

  tasks: tasksApi,

  taskConversations: taskConversationsApi,


  taskExecutionLogs: taskExecutionLogsApi,

  dashboard: dashboardApi,

  attachments: attachmentApi,
};
