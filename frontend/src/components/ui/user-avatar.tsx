"use client"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";
import type { AdminAvatar } from "@/lib/api/types";

interface UserAvatarProps {
  user?: string;
  name?: string;
  avatar?: AdminAvatar;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const sizeClasses = {
  sm: "h-6 w-6 text-xs",
  md: "h-8 w-8 text-sm", 
  lg: "h-12 w-12 text-base"
};

export function UserAvatar({ 
  user, 
  name, 
  avatar,
  size = 'md',
  className 
}: UserAvatarProps) {
  // Generate initials from user or name
  const getInitials = () => {
    if (name && name.trim()) {
      return name
        .trim()
        .split(' ')
        .map(word => word.charAt(0))
        .join('')
        .slice(0, 2)
        .toUpperCase();
    }
    
    if (user && user.trim()) {
      return user.charAt(0).toUpperCase();
    }
    
    return '?';
  };

  // Build avatar URL with API base URL
  const getAvatarUrl = () => {
    if (avatar?.uuid) {
      return `/api/v1/admin/avatar/preview/${avatar.uuid}`;
    }
    return '';
  };

  const initials = getInitials();
  const avatarUrl = getAvatarUrl();

  return (
    <Avatar className={cn(sizeClasses[size], "rounded-lg", className)}>
      <AvatarImage src={avatarUrl} alt={name || user || 'User'} />
      <AvatarFallback className="rounded-lg">
        {initials}
      </AvatarFallback>
    </Avatar>
  );
}