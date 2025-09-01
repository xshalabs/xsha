import { useMemo, useCallback } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import type { AdminRole } from '@/lib/api/types';

export interface PermissionConfig {
  allowedRoles: AdminRole[];
  requireOwnership?: boolean;
  resourceOwnerId?: string;
}

export const usePermissions = () => {
  const { role, user } = useAuth();

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
    if (config.requireOwnership && config.resourceOwnerId) {
      return user === config.resourceOwnerId;
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

  const canEditProject = (projectOwnerUsername?: string) => {
    // Super admin can edit any project
    if (isSuperAdmin) return true;
    
    // Admin can edit their own projects
    if (role === 'admin') {
      return projectOwnerUsername === user;
    }
    
    return false;
  };

  const canDeleteProject = (projectOwnerUsername?: string) => {
    // Super admin can delete any project
    if (isSuperAdmin) return true;
    
    // Admin can delete their own projects
    if (role === 'admin') {
      return projectOwnerUsername === user;
    }
    
    return false;
  };

  // Git Credential permissions
  const canCreateCredential = useMemo(() => 
    hasRole(['admin', 'super_admin']), 
    [hasRole]
  );

  const canEditCredential = (credentialOwnerUsername?: string) => {
    // Super admin can edit any credential
    if (isSuperAdmin) return true;
    
    // Admin can edit their own credentials
    if (role === 'admin') {
      return credentialOwnerUsername === user;
    }
    
    return false;
  };

  const canDeleteCredential = (credentialOwnerUsername?: string) => {
    // Super admin can delete any credential
    if (isSuperAdmin) return true;
    
    // Admin can delete their own credentials
    if (role === 'admin') {
      return credentialOwnerUsername === user;
    }
    
    return false;
  };

  // Development Environment permissions
  const canCreateEnvironment = useMemo(() => 
    hasRole(['admin', 'super_admin']), 
    [hasRole]
  );

  const canEditEnvironment = (environmentOwnerUsername?: string) => {
    // Super admin can edit any environment
    if (isSuperAdmin) return true;
    
    // Admin can edit their own environments
    if (role === 'admin') {
      return environmentOwnerUsername === user;
    }
    
    return false;
  };

  const canDeleteEnvironment = (environmentOwnerUsername?: string) => {
    // Super admin can delete any environment
    if (isSuperAdmin) return true;
    
    // Admin can delete their own environments
    if (role === 'admin') {
      return environmentOwnerUsername === user;
    }
    
    return false;
  };

  // Task permissions
  const canCreateTask = useMemo(() => true, []); // All roles can create tasks

  const canEditTask = (taskOwnerUsername?: string) => {
    // Super admin can edit any task
    if (isSuperAdmin) return true;
    
    // Users can edit their own tasks
    return taskOwnerUsername === user;
  };

  const canDeleteTask = (taskOwnerUsername?: string) => {
    // Super admin can delete any task
    if (isSuperAdmin) return true;
    
    // Admin can delete their own tasks
    if (role === 'admin') {
      return taskOwnerUsername === user;
    }
    
    // Developer cannot delete tasks
    return false;
  };

  // Conversation permissions (same as task permissions)
  const canCreateConversation = useMemo(() => true, []); // All roles can create conversations

  const canEditConversation = (conversationOwnerUsername?: string) => {
    // Super admin can edit any conversation
    if (isSuperAdmin) return true;
    
    // Users can edit their own conversations
    return conversationOwnerUsername === user;
  };

  const canDeleteConversation = (conversationOwnerUsername?: string) => {
    // Super admin can delete any conversation
    if (isSuperAdmin) return true;
    
    // Users can delete their own conversations
    return conversationOwnerUsername === user;
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
    // Environment permissions
    canCreateEnvironment,
    canEditEnvironment,
    canDeleteEnvironment,
    // Task permissions
    canCreateTask,
    canEditTask,
    canDeleteTask,
    // Conversation permissions
    canCreateConversation,
    canEditConversation,
    canDeleteConversation,
    // System access permissions
    canAccessSettings,
    canViewLogs,
    canAccessAdminPanel,
  };
};