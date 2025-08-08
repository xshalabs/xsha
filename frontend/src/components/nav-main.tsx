import { Link, useLocation } from "react-router-dom";
import type { LucideIcon } from "lucide-react";

import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";

export function NavMain({
  items,
}: {
  items: {
    title: string
    url: string
    icon?: LucideIcon
  }[]
}) {
  const location = useLocation();
  const { setOpenMobile } = useSidebar();

  const isActive = (itemUrl: string) => {
    if (itemUrl === "/projects") {
      return location.pathname.startsWith("/projects");
    }
    if (itemUrl === "/git-credentials") {
      return location.pathname.startsWith("/git-credentials");
    }
    if (itemUrl === "/dev-environments") {
      return location.pathname.startsWith("/dev-environments");
    }
    if (itemUrl === "/admin/logs") {
      return location.pathname.startsWith("/admin/logs");
    }
    if (itemUrl === "/system-configs") {
      return location.pathname.startsWith("/system-configs");
    }
    return location.pathname === itemUrl;
  };

  return (
    <SidebarGroup>
      <SidebarGroupLabel>Platform</SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) => (
          <SidebarMenuItem key={item.title}>
            <SidebarMenuButton 
              tooltip={item.title}
              asChild
              isActive={isActive(item.url)}
            >
              <Link to={item.url} onClick={() => setOpenMobile(false)}>
                {item.icon && <item.icon />}
                <span>{item.title}</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
