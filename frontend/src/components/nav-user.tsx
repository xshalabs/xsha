import {
  IconDotsVertical,
  IconLogout,
  IconLanguage,
} from "@tabler/icons-react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { useAuth } from "@/contexts/AuthContext";
import { SUPPORTED_LANGUAGES } from "@/lib/constants";
import { apiService } from "@/lib/api";

export function NavUser() {
  const { isMobile } = useSidebar();
  const { t, i18n } = useTranslation();
  const { user, logout, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await logout();
      navigate("/login");
    } catch (error) {
      console.error("Logout failed:", error);
      // 即使登出失败，也要跳转到登录页面，因为logout函数会清除本地状态
      navigate("/login");
    }
  };

  const handleLanguageChange = async (languageCode: string) => {
    // 切换前端语言
    i18n.changeLanguage(languageCode);

    // 如果用户已登录，同步设置后端语言偏好
    if (isAuthenticated) {
      try {
        await apiService.setLanguagePreference(languageCode);
      } catch (error) {
        console.warn("Failed to sync language preference with backend:", error);
        // 不阻止前端语言切换，即使后端同步失败
      }
    }
  };

  // 如果没有用户信息，不渲染组件
  if (!user) {
    return null;
  }

  // 生成用户名首字母作为头像备用显示
  const userInitials = user.charAt(0).toUpperCase();

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <Avatar className="h-8 w-8 rounded-lg">
                <AvatarImage src="" alt={user} />
                <AvatarFallback className="rounded-lg">
                  {userInitials}
                </AvatarFallback>
              </Avatar>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">{user}</span>
                <span className="text-muted-foreground truncate text-xs">
                  {t("navigation.profile")}
                </span>
              </div>
              <IconDotsVertical className="ml-auto size-4" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            side={isMobile ? "bottom" : "right"}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuLabel className="p-0 font-normal">
              <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                <Avatar className="h-8 w-8 rounded-lg">
                  <AvatarImage src="" alt={user} />
                  <AvatarFallback className="rounded-lg">
                    {userInitials}
                  </AvatarFallback>
                </Avatar>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">{user}</span>
                  <span className="text-muted-foreground truncate text-xs">
                    {t("navigation.profile")}
                  </span>
                </div>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              {/* <DropdownMenuItem>
                <IconUserCircle />
                {t("navigation.profile")}
              </DropdownMenuItem> */}
              {/* <DropdownMenuItem>
                <IconSettings />
                {t("navigation.settings")}
              </DropdownMenuItem> */}
              <DropdownMenuSub>
                <DropdownMenuSubTrigger>
                  <IconLanguage />
                  {t("navigation.language")}
                </DropdownMenuSubTrigger>
                <DropdownMenuSubContent>
                  <DropdownMenuRadioGroup
                    value={i18n.language}
                    onValueChange={handleLanguageChange}
                  >
                    {SUPPORTED_LANGUAGES.map((language) => (
                      <DropdownMenuRadioItem
                        key={language.code}
                        value={language.code}
                      >
                        <span className="mr-2">{language.flag}</span>
                        {language.name}
                      </DropdownMenuRadioItem>
                    ))}
                  </DropdownMenuRadioGroup>
                </DropdownMenuSubContent>
              </DropdownMenuSub>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout}>
              <IconLogout />
              {t("auth.logout")}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
