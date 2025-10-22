import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { ModeToggle } from "@/components/ModeToggle";
import { usePageActions } from "@/contexts/PageActionsContext";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";
import { usePageTitleContext } from "@/contexts/PageTitleContext";
import { NavBreadcrumb } from "@/components/nav/nav-breadcrumb";

export function SiteHeader() {
  const { actions } = usePageActions();
  const { items } = useBreadcrumb();
  const { pageTitle } = usePageTitleContext();

  return (
    <header className="flex sticky top-0 bg-background h-14 shrink-0 items-center gap-2 border-b px-2 z-10">
      <div className="flex flex-1 items-center gap-2 px-3">
        <SidebarTrigger className="-ml-1" />
        <Separator orientation="vertical" className="mr-2 h-4" />
        {items.length > 0 ? (
          <NavBreadcrumb items={items} />
        ) : (
          <h1 className="text-sm font-semibold">
            {pageTitle}
          </h1>
        )}
      </div>
      <div className="ml-auto px-3">
        <div className="flex items-center gap-2">
          {actions}
          {actions && <Separator orientation="vertical" className="h-4" />}
          <ModeToggle />
          <Button variant="ghost" asChild size="sm" className="hidden sm:flex">
            <a
              href="https://github.com/xshaLabs/xsha"
              rel="noopener noreferrer"
              target="_blank"
            >
              GitHub
            </a>
          </Button>
        </div>
      </div>
    </header>
  );
}
