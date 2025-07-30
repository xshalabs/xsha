import * as React from "react";
import { useTranslation } from "react-i18next";
import {
  IconDashboard,
  IconFolder,
  IconKey,
  IconFileText,
  IconInnerShadowTop,
  IconContainer,
  IconSettings,
} from "@tabler/icons-react";

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
} from "@/components/ui/sidebar";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { t } = useTranslation();

  const data = {
    navMain: [
      {
        title: t("navigation.dashboard"),
        url: "/dashboard",
        icon: IconDashboard,
      },
      {
        title: t("navigation.projects"),
        url: "/projects",
        icon: IconFolder,
      },
      {
        title: t("navigation.gitCredentials"),
        url: "/git-credentials",
        icon: IconKey,
      },
      {
        title: t("navigation.dev_environments"),
        url: "/dev-environments",
        icon: IconContainer,
      },
      {
        title: t("navigation.adminLogs"),
        url: "/admin/logs",
        icon: IconFileText,
      },
      {
        title: t("navigation.systemConfigs"),
        url: "/system-configs",
        icon: IconSettings,
      },
    ],
    navSecondary: [
      // {
      //   title: t("navigation.settings"),
      //   url: "#",
      //   icon: IconSettings,
      // },
      // {
      //   title: t("navigation.help"),
      //   url: "#",
      //   icon: IconHelp,
      // },
      // {
      //   title: t("navigation.search"),
      //   url: "#",
      //   icon: IconSearch,
      // },
    ],
  };

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:!p-1.5"
            >
              <a href="/">
                <IconInnerShadowTop className="!size-5" />
                <span className="text-base font-semibold">XSHA</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser />
      </SidebarFooter>
    </Sidebar>
  );
}
