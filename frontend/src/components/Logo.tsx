import React from "react";
import { useTheme } from "@/components/ThemeProvider";
import xshaLightLogo from "@/assets/xsha_light.png";
import xshaDarkLogo from "@/assets/xsha_dark.png";

interface LogoProps {
  className?: string;
  alt?: string;
}

export const Logo: React.FC<LogoProps> = ({ 
  className = "h-8 w-auto", 
  alt = "XSHA" 
}) => {
  const { theme } = useTheme();

  return (
    <img
      src={theme === 'dark' ? xshaDarkLogo : xshaLightLogo}
      alt={alt}
      className={className}
    />
  );
};
