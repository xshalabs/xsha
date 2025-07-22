// 自定义错误类型
export class ApiError extends Error {
  status?: number;
  code?: string;

  constructor(
    message: string,
    status?: number,
    code?: string
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

export class NetworkError extends Error {
  constructor(message = 'Network error occurred') {
    super(message);
    this.name = 'NetworkError';
  }
}

export class AuthError extends Error {
  constructor(message = 'Authentication failed') {
    super(message);
    this.name = 'AuthError';
  }
}

// 错误处理工具函数 - 优先使用后端返回的国际化消息
export const handleApiError = (error: unknown): string => {
  if (error instanceof ApiError) {
    // 直接返回后端的国际化错误消息
    return error.message;
  }
  
  if (error instanceof NetworkError) {
    return 'Network connection failed. Please check your internet connection.';
  }
  
  if (error instanceof AuthError) {
    return 'Authentication failed. Please login again.';
  }
  
  if (error instanceof Error) {
    return error.message;
  }
  
  return 'An unknown error occurred';
};

// 开发环境错误日志
export const logError = (error: unknown, context?: string): void => {
  if (import.meta.env.NODE_ENV === 'development') {
    console.error(`[Error${context ? ` in ${context}` : ''}]:`, error);
  }
}; 