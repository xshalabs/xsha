import React from "react";
import { useTheme } from "@/components/ThemeProvider";
import { useSidebar } from "@/components/ui/sidebar";
import xshaLightLogo from "@/assets/xsha_light.png";
import xshaDarkLogo from "@/assets/xsha_dark.png";
import logoImage from "@/assets/logo.png";

interface LogoProps {
  className?: string;
  alt?: string;
}

export const Logo: React.FC<LogoProps> = ({ 
  className = "h-8 w-auto", 
  alt = "XSHA" 
}) => {
  const { theme } = useTheme();
  
  // Try to get sidebar state, but handle gracefully if not in a SidebarProvider context
  let sidebarState: "expanded" | "collapsed" | null = null;
  try {
    const { state } = useSidebar();
    sidebarState = state;
  } catch (error) {
    // useSidebar hook is not available (e.g., on login page), that's fine
    sidebarState = null;
  }

  // 当侧边栏收缩时使用 logo.png，否则根据主题使用相应的logo
  // If no sidebar context, default to theme-based logo
  const logoSrc = sidebarState === "collapsed" 
    ? logoImage 
    : (theme === 'dark' ? xshaDarkLogo : xshaLightLogo);

  return (
    <img
      src={logoSrc}
      alt={alt}
      className={className}
    />
  );
};
