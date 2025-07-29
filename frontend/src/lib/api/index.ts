export { API_BASE_URL, getApiBaseUrl } from "./config";
export { tokenManager, getCurrentLanguage } from "./token";
export { request } from "./request";

export type {
  LoginRequest,
  LoginResponse,
  UserResponse,
  ApiErrorResponse,
  LanguagesResponse,
} from "./types";

import { authApi } from "./auth";
import { gitCredentialsApi } from "./git-credentials";
import { adminLogsApi } from "./admin-logs";
import { projectsApi } from "./projects";
import { devEnvironmentsApi } from "./dev-environments";
import { tasksApi } from "./tasks";
import { taskConversationsApi } from "./task-conversations";
import { taskConversationResultsApi } from "./task-conversation-results";
import { taskExecutionLogsApi } from "./task-execution-logs";

export {
  authApi,
  gitCredentialsApi,
  adminLogsApi,
  projectsApi,
  devEnvironmentsApi,
  tasksApi,
  taskConversationsApi,
  taskConversationResultsApi,
  taskExecutionLogsApi,
};

export const apiService = {
  login: authApi.login,
  logout: authApi.logout,
  getCurrentUser: authApi.getCurrentUser,
  healthCheck: authApi.healthCheck,
  getSupportedLanguages: authApi.getSupportedLanguages,
  setLanguagePreference: authApi.setLanguagePreference,

  gitCredentials: gitCredentialsApi,

  adminLogs: adminLogsApi,

  projects: projectsApi,

  devEnvironments: devEnvironmentsApi,

  tasks: tasksApi,

  taskConversations: taskConversationsApi,

  taskConversationResults: taskConversationResultsApi,

  taskExecutionLogs: taskExecutionLogsApi,
};
