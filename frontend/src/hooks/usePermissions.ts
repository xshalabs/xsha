import { useMemo, useCallback } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import type { AdminRole } from '@/lib/api/types';

export interface PermissionConfig {
  allowedRoles: AdminRole[];
  requireOwnership?: boolean;
  resourceOwnerAdminId?: number;
}

export const usePermissions = () => {
  const { role, user, adminId } = useAuth();

  const hasRole = useCallback((allowedRoles: AdminRole[]): boolean => {
    if (!role) return false;
    return allowedRoles.includes(role);
  }, [role]);

  const hasPermission = (config: PermissionConfig): boolean => {
    // Check if user has required role
    if (!hasRole(config.allowedRoles)) {
      return false;
    }

    // If ownership is required, check if user is the owner
    if (config.requireOwnership && config.resourceOwnerAdminId !== undefined) {
      return adminId === config.resourceOwnerAdminId;
    }

    return true;
  };

  const isSuperAdmin = useMemo(() => role === 'super_admin', [role]);
  const isAdmin = useMemo(() => role === 'admin' || role === 'super_admin', [role]);
  const isDeveloper = useMemo(() => role === 'developer', [role]);

  // Common permission checks for admin management
  const canCreateAdmin = useMemo(() => 
    hasRole(['super_admin']), 
    [hasRole]
  );

  const canEditAdmin = (targetAdminUsername?: string) => {
    // Super admin can edit anyone
    if (isSuperAdmin) return true;
    
    // Admin can edit themselves and developers
    if (role === 'admin') {
      return targetAdminUsername === user || !targetAdminUsername;
    }
    
    // Developer can only edit themselves
    if (role === 'developer') {
      return targetAdminUsername === user;
    }
    
    return false;
  };

  const canDeleteAdmin = (targetAdminRole?: AdminRole, createdBy?: string) => {
    // Only super admin can delete admins
    if (!isSuperAdmin) return false;
    
    // Can't delete system-created admins
    if (createdBy === 'system') return false;
    
    // Super admins can't delete other super admins (prevent self-destruction)
    if (targetAdminRole === 'super_admin') return false;
    
    return true;
  };

  const canChangeAdminPassword = (targetAdminUsername?: string, targetAdminRole?: AdminRole) => {
    // Super admin can change anyone's password
    if (isSuperAdmin) return true;
    
    // Admin can change developer passwords and their own
    if (role === 'admin') {
      return targetAdminUsername === user || targetAdminRole === 'developer';
    }
    
    // Developer can only change their own password
    if (role === 'developer') {
      return targetAdminUsername === user;
    }
    
    return false;
  };

  const canManageAdminRole = () => {
    // Only super admin can manage roles
    if (!isSuperAdmin) return false;
    
    // Super admin can manage any role
    return true;
  };

  // Project permissions
  const canCreateProject = useMemo(() => 
    hasRole(['admin', 'super_admin']), 
    [hasRole]
  );

  const canEditProject = (resourceAdminId?: number) => {
    // Super admin can edit any project
    if (isSuperAdmin) return true;
    
    // Admin can edit their own projects
    if (role === 'admin') {
      return resourceAdminId === adminId;
    }
    
    return false;
  };

  const canDeleteProject = (resourceAdminId?: number) => {
    // Super admin can delete any project
    if (isSuperAdmin) return true;
    
    // Admin can delete their own projects
    if (role === 'admin') {
      return resourceAdminId === adminId;
    }
    
    return false;
  };

  // Git Credential permissions
  const canCreateCredential = useMemo(() => 
    hasRole(['developer', 'admin', 'super_admin']), 
    [hasRole]
  );

  const canEditCredential = (resourceAdminId?: number) => {
    // Super admin can edit any credential
    if (isSuperAdmin) return true;
    
    // Admin and Developer can edit their own credentials
    if (role === 'admin' || role === 'developer') {
      return resourceAdminId === adminId;
    }
    
    return false;
  };

  const canDeleteCredential = (resourceAdminId?: number) => {
    // Super admin can delete any credential
    if (isSuperAdmin) return true;
    
    // Admin and Developer can delete their own credentials
    if (role === 'admin' || role === 'developer') {
      return resourceAdminId === adminId;
    }
    
    return false;
  };

  const canManageCredentialAdmins = (resourceAdminId?: number) => {
    // Super admin can manage any credential's admins
    if (isSuperAdmin) return true;
    
    // Admin and Developer can manage admins of their own credentials
    if (role === 'admin' || role === 'developer') {
      return resourceAdminId === adminId;
    }
    
    return false;
  };

  // Development Environment permissions
  const canCreateEnvironment = useMemo(() => 
    hasRole(['developer', 'admin', 'super_admin']), 
    [hasRole]
  );

  const canEditEnvironment = (resourceAdminId?: number, isEnvironmentAdmin?: boolean) => {
    // Super admin can edit any environment
    if (isSuperAdmin) return true;
    
    // Admin and Developer can edit environments they have admin access to
    if (role === 'admin' || role === 'developer') {
      // If isEnvironmentAdmin is explicitly provided, use it; otherwise fall back to resourceAdminId check
      return isEnvironmentAdmin === true || resourceAdminId === adminId;
    }
    
    return false;
  };

  const canDeleteEnvironment = (resourceAdminId?: number, isEnvironmentAdmin?: boolean) => {
    // Super admin can delete any environment
    if (isSuperAdmin) return true;
    
    // Admin and Developer can delete environments they have admin access to
    if (role === 'admin' || role === 'developer') {
      // If isEnvironmentAdmin is explicitly provided, use it; otherwise fall back to resourceAdminId check
      return isEnvironmentAdmin === true || resourceAdminId === adminId;
    }
    
    return false;
  };

  const canManageEnvironmentAdmins = (resourceAdminId?: number, isEnvironmentAdmin?: boolean) => {
    // Super admin can manage any environment's admins
    if (isSuperAdmin) return true;
    
    // Admin and Developer can manage admins of environments they have admin access to
    if (role === 'admin' || role === 'developer') {
      // If isEnvironmentAdmin is explicitly provided, use it; otherwise fall back to resourceAdminId check
      return isEnvironmentAdmin === true || resourceAdminId === adminId;
    }
    
    return false;
  };

  // Task permissions
  const canCreateTask = useMemo(() => true, []); // All roles can create tasks

  const canEditTask = (resourceAdminId?: number) => {
    // Super admin can edit any task
    if (isSuperAdmin) return true;
    
    // Users can edit their own tasks
    return resourceAdminId === adminId;
  };

  const canDeleteTask = (resourceAdminId?: number) => {
    // Super admin can delete any task
    if (isSuperAdmin) return true;
    
    // Admin can delete their own tasks
    if (role === 'admin') {
      return resourceAdminId === adminId;
    }
    
    // Developer cannot delete tasks
    return false;
  };

  // Conversation permissions (same as task permissions)
  const canCreateConversation = useMemo(() => true, []); // All roles can create conversations

  const canEditConversation = (resourceAdminId?: number) => {
    // Super admin can edit any conversation
    if (isSuperAdmin) return true;
    
    // Users can edit their own conversations
    return resourceAdminId === adminId;
  };

  const canDeleteConversation = (resourceAdminId?: number) => {
    // Super admin can delete any conversation
    if (isSuperAdmin) return true;
    
    // Users can delete their own conversations
    return resourceAdminId === adminId;
  };

  // Notifier permissions
  const canCreateNotifier = useMemo(() =>
    hasRole(['admin', 'super_admin']),
    [hasRole]
  );

  const canEditNotifier = (resourceAdminId?: number) => {
    // Super admin can edit any notifier
    if (isSuperAdmin) return true;

    // Admin can edit their own notifiers
    if (role === 'admin') {
      return resourceAdminId === adminId;
    }

    return false;
  };

  const canDeleteNotifier = (resourceAdminId?: number) => {
    // Super admin can delete any notifier
    if (isSuperAdmin) return true;

    // Admin can delete their own notifiers
    if (role === 'admin') {
      return resourceAdminId === adminId;
    }

    return false;
  };

  // MCP permissions
  const canCreateMCP = useMemo(() =>
    hasRole(['admin', 'super_admin']),
    [hasRole]
  );

  const canEditMCP = (resourceAdminId?: number) => {
    // Super admin can edit any MCP configuration
    if (isSuperAdmin) return true;

    // Admin can edit their own MCP configurations
    if (role === 'admin') {
      return resourceAdminId === adminId;
    }

    return false;
  };

  const canDeleteMCP = (resourceAdminId?: number) => {
    // Super admin can delete any MCP configuration
    if (isSuperAdmin) return true;

    // Admin can delete their own MCP configurations
    if (role === 'admin') {
      return resourceAdminId === adminId;
    }

    return false;
  };

  // System access permissions
  const canAccessSettings = useMemo(() =>
    hasRole(['super_admin']),
    [hasRole]
  );

  const canViewLogs = useMemo(() =>
    hasRole(['super_admin']),
    [hasRole]
  );

  const canAccessAdminPanel = useMemo(() =>
    hasRole(['super_admin']),
    [hasRole]
  );

  return {
    role,
    user,
    adminId,
    hasRole,
    hasPermission,
    isSuperAdmin,
    isAdmin,
    isDeveloper,
    // Admin permissions
    canCreateAdmin,
    canEditAdmin,
    canDeleteAdmin,
    canChangeAdminPassword,
    canManageAdminRole,
    // Project permissions
    canCreateProject,
    canEditProject,
    canDeleteProject,
    // Credential permissions
    canCreateCredential,
    canEditCredential,
    canDeleteCredential,
    canManageCredentialAdmins,
    // Environment permissions
    canCreateEnvironment,
    canEditEnvironment,
    canDeleteEnvironment,
    canManageEnvironmentAdmins,
    // Task permissions
    canCreateTask,
    canEditTask,
    canDeleteTask,
    // Conversation permissions
    canCreateConversation,
    canEditConversation,
    canDeleteConversation,
    // Notifier permissions
    canCreateNotifier,
    canEditNotifier,
    canDeleteNotifier,
    // MCP permissions
    canCreateMCP,
    canEditMCP,
    canDeleteMCP,
    // System access permissions
    canAccessSettings,
    canViewLogs,
    canAccessAdminPanel,
  };
};