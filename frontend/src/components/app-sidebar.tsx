import * as React from "react";
import { useTranslation } from "react-i18next";
import {
  LayoutGrid,
  Folder,
  Key,
  FileText,
  Container,
  Settings,
  Cog,
} from "lucide-react";

import { NavMain } from "@/components/nav-main";
import { NavSecondary } from "@/components/nav-secondary";
import { NavUser } from "@/components/nav-user";
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

  const data = {
    navMain: [
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
        url: "/git-credentials",
        icon: Key,
      },
      {
        title: t("navigation.dev_environments"),
        url: "/dev-environments",
        icon: Container,
      },
      {
        title: t("navigation.adminLogs"),
        url: "/admin/logs",
        icon: FileText,
      },
      {
        title: t("navigation.systemConfigs"),
        url: "/system-configs",
        icon: Cog,
      },
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
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                  <Settings className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">XSHA</span>
                  <span className="truncate text-xs">AI DevOps Platform</span>
                </div>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter className="border-t">
        <NavUser />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
