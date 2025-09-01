import { useMemo } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import type { AdminRole } from '@/lib/api/types';

export interface PermissionConfig {
  allowedRoles: AdminRole[];
  requireOwnership?: boolean;
  resourceOwnerId?: string;
}

export const usePermissions = () => {
  const { role, user } = useAuth();

  const hasRole = (allowedRoles: AdminRole[]): boolean => {
    if (!role) return false;
    return allowedRoles.includes(role);
  };

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

  return {
    role,
    user,
    hasRole,
    hasPermission,
    isSuperAdmin,
    isAdmin,
    isDeveloper,
    canCreateAdmin,
    canEditAdmin,
    canDeleteAdmin,
    canChangeAdminPassword,
    canManageAdminRole,
  };
};