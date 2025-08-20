import type { TaskStatus } from "@/types/task";

/**
 * 获取任务状态的样式类名
 */
export const getStatusBadgeClass = (status: TaskStatus): string => {
  const statusClasses: Record<TaskStatus, string> = {
    todo: "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200",
    in_progress: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
    done: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
    cancelled: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
  };
  
  return statusClasses[status] || statusClasses.todo;
};

/**
 * 获取对话状态的样式类名
 */
export const getConversationStatusColor = (status: string): string => {
  const statusColors: Record<string, string> = {
    pending: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
    running: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
    success: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
    failed: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
    cancelled: "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200",
  };
  
  return statusColors[status] || statusColors.pending;
};

/**
 * 格式化时间显示
 */
export const formatTime = (dateString: string): string => {
  return new Date(dateString).toLocaleString();
};

/**
 * 格式化时间显示（不显示秒）
 */
export const formatTimeWithoutSeconds = (dateString: string): string => {
  return new Date(dateString).toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
};

/**
 * 格式化日期显示
 */
export const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString();
};

/**
 * 检查对话是否是未来执行的
 */
export const isFutureExecution = (executionTime?: string): boolean => {
  if (!executionTime) return false;
  return new Date(executionTime) > new Date();
};
