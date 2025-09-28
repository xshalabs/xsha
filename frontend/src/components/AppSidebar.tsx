import * as React from "react";
import { useTranslation } from "react-i18next";
import {
  LayoutGrid,
  Folder,
  Key,
  Container,
  Bell,
  Settings2,
  Cog,
  Shield,
  TrendingUp,
  Activity,
  Users,
} from "lucide-react";
import { usePermissions } from "@/hooks/usePermissions";

import { NavMain } from "@/components/NavMain";
import { NavSecondary } from "@/components/NavSecondary";
import { NavUser } from "@/components/NavUser";
import { Logo } from "@/components/Logo";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { t } = useTranslation();
  const { canViewLogs, canAccessAdminPanel, canAccessSettings, canCreateNotifier, canCreateMCP } = usePermissions();

  const data = {
    navGroups: [
      {
        title: t("navigation.groups.workspace"),
        items: [
          {
            title: t("navigation.dashboard"),
            url: "/dashboard",
            icon: LayoutGrid,
          },
          {
            title: t("navigation.projects"),
            url: "/projects",
            icon: Folder,
          },
          {
            title: t("navigation.gitCredentials"),
            url: "/credentials",
            icon: Key,
          },
          {
            title: t("navigation.environments"),
            url: "/environments",
            icon: Container,
          },
          ...(canCreateMCP ? [{
            title: t("navigation.mcp"),
            url: "/mcp",
            icon: Settings2,
          }] : []),
        ],
      },
      // Only show logs section for super admins
      ...(canViewLogs ? [{
        title: t("navigation.groups.logs"),
        items: [
          {
            title: t("navigation.logs.loginLogs"),
            url: "/logs/login-logs",
            icon: Shield,
          },
          {
            title: t("navigation.logs.operationLogs"),
            url: "/logs/operation-logs",
            icon: Activity,
          },
          {
            title: t("navigation.logs.stats"),
            url: "/logs/stats",
            icon: TrendingUp,
          },
        ],
      }] : []),
      // Only show admin section for super admins
      ...(canAccessAdminPanel || canAccessSettings || canCreateNotifier ? [{
        title: t("navigation.groups.admin"),
        items: [
          // Admin users management - only for super admin
          ...(canAccessAdminPanel ? [{
            title: t("navigation.admin.users"),
            url: "/admin",
            icon: Users,
          }] : []),
          // Notifier management - for admin and super admin
          ...(canCreateNotifier ? [{
            title: t("navigation.notifiers"),
            url: "/notifiers",
            icon: Bell,
          }] : []),
          // System settings - only for super admin
          ...(canAccessSettings ? [{
            title: t("navigation.settings"),
            url: "/settings",
            icon: Cog,
          }] : []),
        ],
      }] : []),
    ],
    navSecondary: [],
  };

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader className="border-b py-1 h-14 flex justify-center">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <a href="/" className="flex items-center gap-2">
                <Logo className="h-8 w-auto" />
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate text-xs">AI Dev Platform</span>
                </div>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain groups={data.navGroups} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter className="border-t">
        <NavUser />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
