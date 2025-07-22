import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { ModeToggle } from "@/components/mode-toggle";
import { useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";

export function SiteHeader() {
  const location = useLocation();
  const { t } = useTranslation();

  // 路径到页面标题的映射
  const getPageTitle = (pathname: string): string => {
    switch (pathname) {
      case "/dashboard":
        return t("common.pageTitle.dashboard");
      case "/projects":
        return t("common.pageTitle.projects");
      case "/git-credentials":
        return t("common.pageTitle.gitCredentials");
      case "/dev-environments":
        return t("navigation.dev_environments");
      case "/admin/logs":
        return t("common.pageTitle.adminLogs");
      case "/login":
        return t("common.pageTitle.login");
      default:
        return t("common.app.name", "Sleep0");
    }
  };

  return (
    <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b border-border transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-(--header-height)">
      <div className="flex w-full items-center gap-1 px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-1" />
        <Separator
          orientation="vertical"
          className="mx-2 data-[orientation=vertical]:h-4"
        />
        <h1 className="text-base font-medium">
          {getPageTitle(location.pathname)}
        </h1>
        <div className="ml-auto flex items-center gap-2">
          <ModeToggle />
          <Button variant="ghost" asChild size="sm" className="hidden sm:flex">
            <a
              href="https://github.com/shadcn-ui/ui/tree/main/apps/v4/app/(examples)/dashboard"
              rel="noopener noreferrer"
              target="_blank"
              className="dark:text-foreground"
            >
              GitHub
            </a>
          </Button>
        </div>
      </div>
    </header>
  );
}
