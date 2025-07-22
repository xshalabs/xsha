import React, { createContext, useContext, useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { apiService, tokenManager } from '@/lib/api/index';
import type { UserResponse } from '@/lib/api/index';
import { handleApiError } from '@/lib/errors';

interface AuthContextType {
  user: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<string | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  // 检查认证状态
  const checkAuth = async () => {
    try {
      if (!tokenManager.isTokenPresent()) {
        setIsAuthenticated(false);
        setUser(null);
        return;
      }

      const response: UserResponse = await apiService.getCurrentUser();
      setUser(response.user);
      setIsAuthenticated(true);
    } catch (error) {
      console.error('Auth check failed:', handleApiError(error));
      setIsAuthenticated(false);
      setUser(null);
      tokenManager.removeToken(); // 清除无效token
    } finally {
      setIsLoading(false);
    }
  };

  // 登录函数
  const login = async (username: string, password: string) => {
    setIsLoading(true);
    try {
      const response = await apiService.login({ username, password });
      setUser(response.user);
      setIsAuthenticated(true);
    } catch (error) {
      setIsAuthenticated(false);
      setUser(null);
      // 抛出带有后端国际化消息的错误
      throw new Error(handleApiError(error));
    } finally {
      setIsLoading(false);
    }
  };

  // 登出函数
  const logout = async () => {
    setIsLoading(true);
    try {
      await apiService.logout();
    } catch (error) {
      console.error('Logout failed:', handleApiError(error));
      // 即使API调用失败，也要清除本地状态
    } finally {
      setIsAuthenticated(false);
      setUser(null);
      setIsLoading(false);
    }
  };

  // 组件挂载时检查认证状态
  useEffect(() => {
    checkAuth();
  }, []);

  const value: AuthContextType = {
    user,
    isAuthenticated,
    isLoading,
    login,
    logout,
    checkAuth,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}; 