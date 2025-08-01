import { Link, useLocation } from "react-router-dom";
import type { Icon } from "@tabler/icons-react";

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

export function NavMain({
  items,
}: {
  items: {
    title: string
    url: string
    icon?: Icon
  }[]
}) {
  const location = useLocation();

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
    return location.pathname === itemUrl;
  };

  return (
    <SidebarGroup>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton 
                tooltip={item.title}
                asChild
                isActive={isActive(item.url)}
              >
                <Link to={item.url}>
                  {item.icon && <item.icon />}
                  <span>{item.title}</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
