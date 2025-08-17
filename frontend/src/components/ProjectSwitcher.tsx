import * as React from "react";
import { Building2, ChevronsUpDown, FolderOpen } from "lucide-react";
import { useNavigate } from "react-router-dom";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { useIsMobile } from "@/hooks/use-mobile";
import type { Project } from "@/types/project";

interface ProjectSwitcherProps {
  projects: Project[];
  currentProject: Project;
  onProjectChange: (projectId: string) => void;
  className?: string;
}

export function ProjectSwitcher({
  projects,
  currentProject,
  onProjectChange,
  className,
}: ProjectSwitcherProps) {
  const navigate = useNavigate();
  const isMobile = useIsMobile();

  const handleProjectSelect = (project: Project) => {
    onProjectChange(project.id.toString());
  };

  // Truncate long text for mobile
  const truncateText = (text: string, maxLength: number) => {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + "...";
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className={`flex items-center gap-2 px-2 py-2 h-auto text-left hover:bg-accent/50 data-[state=open]:bg-accent ${
            isMobile ? "min-w-0 max-w-48" : "gap-3 px-3"
          }`}
        >
          <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground flex-shrink-0">
            <Building2 className="size-4" />
          </div>
          <div className="grid flex-1 text-left text-sm leading-tight min-w-0">
            <span className="truncate font-medium text-foreground">
              {isMobile ? truncateText(currentProject.name, 20) : currentProject.name}
            </span>
            {!isMobile && (
              <span className="truncate text-xs text-muted-foreground">
                {currentProject.repo_url}
              </span>
            )}
          </div>
          <ChevronsUpDown className="ml-auto size-4 text-muted-foreground flex-shrink-0" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        className={`${isMobile ? "min-w-72 max-w-80" : "min-w-64"} rounded-lg`}
        align="start"
        side={isMobile ? "bottom" : "right"}
        sideOffset={4}
      >
        <DropdownMenuLabel className="text-xs text-muted-foreground">
          Projects
        </DropdownMenuLabel>
        {projects.map((project) => (
          <DropdownMenuItem
            key={project.id}
            onClick={() => handleProjectSelect(project)}
            className={`gap-3 ${isMobile ? "p-4" : "p-3"} cursor-pointer`}
          >
            <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-muted text-muted-foreground flex-shrink-0">
              <FolderOpen className="size-4" />
            </div>
            <div className="grid flex-1 text-left text-sm leading-tight min-w-0">
              <span className="truncate font-medium">{project.name}</span>
              <span className="truncate text-xs text-muted-foreground">
                {project.repo_url}
              </span>
            </div>
            {currentProject.id === project.id && (
              <div className="size-2 rounded-full bg-primary flex-shrink-0" />
            )}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
