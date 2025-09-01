import React, { createContext, useContext, useEffect, useState } from "react";
import type { ReactNode } from "react";
import { apiService, tokenManager } from "@/lib/api/index";
import type { UserResponse } from "@/lib/api/index";
import type { AdminAvatar, AdminRole } from "@/lib/api/types";
import { handleApiError } from "@/lib/errors";

interface AuthContextType {
  user: string | null;
  name: string | null;
  avatar: AdminAvatar | null;
  role: AdminRole | null;
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
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<string | null>(null);
  const [name, setName] = useState<string | null>(null);
  const [avatar, setAvatar] = useState<AdminAvatar | null>(null);
  const [role, setRole] = useState<AdminRole | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const checkAuth = async () => {
    try {
      if (!tokenManager.isTokenPresent()) {
        setIsAuthenticated(false);
        setUser(null);
        return;
      }

      const response: UserResponse = await apiService.getCurrentUser();
      setUser(response.user);
      setName(response.name);
      setAvatar(response.avatar || null);
      setRole(response.role);
      setIsAuthenticated(true);
    } catch (error) {
      console.error("Auth check failed:", handleApiError(error));
      setIsAuthenticated(false);
      setUser(null);
      setName(null);
      setAvatar(null);
      setRole(null);
      tokenManager.removeToken();
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (username: string, password: string) => {
    setIsLoading(true);
    try {
      const response = await apiService.login({ username, password });
      setUser(response.user);
      setIsAuthenticated(true);
      
      // Fetch current user information to get complete user details including name, avatar, and role
      const userInfo = await apiService.getCurrentUser();
      setName(userInfo.name);
      setAvatar(userInfo.avatar || null);
      setRole(userInfo.role);
    } catch (error) {
      setIsAuthenticated(false);
      setUser(null);
      setName(null);
      setAvatar(null);
      setRole(null);
      throw new Error(handleApiError(error));
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    setIsLoading(true);
    try {
      await apiService.logout();
    } catch (error) {
      console.error("Logout failed:", handleApiError(error));
    } finally {
      setIsAuthenticated(false);
      setUser(null);
      setName(null);
      setAvatar(null);
      setRole(null);
      setIsLoading(false);
    }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  const value: AuthContextType = {
    user,
    name,
    avatar,
    role,
    isAuthenticated,
    isLoading,
    login,
    logout,
    checkAuth,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
