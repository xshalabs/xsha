import { ColumnDef } from "@tanstack/react-table";
import { MoreHorizontal, Edit, Trash2, FolderOpen } from "lucide-react";
import { TFunction } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

import { Badge } from "@/components/ui/badge";
import { DataTableColumnHeader } from "@/components/ui/data-table/data-table-column-header";
import { Project } from "@/types/project";

interface ProjectColumnsProps {
  t: TFunction;
  onEdit: (project: Project) => void;
  onDelete: (id: number) => void;
  onManageTasks: (project: Project) => void;
}

export const createProjectColumns = ({
  t,
  onEdit,
  onDelete,
  onManageTasks,
}: ProjectColumnsProps): ColumnDef<Project>[] => [
  {
    accessorKey: "name",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t("projects.name")} />
    ),
    cell: ({ row }) => (
      <div className="font-medium">{row.getValue("name")}</div>
    ),
    enableSorting: true,
    enableHiding: false,
  },
  {
    accessorKey: "description",
    header: t("projects.description"),
    cell: ({ row }) => {
      const description = row.getValue("description") as string;
      return (
        <div className="max-w-[300px] truncate text-muted-foreground">
          {description || t("common.noDescription")}
        </div>
      );
    },
  },
  {
    accessorKey: "repo_url",
    header: t("projects.repoUrl"),
    cell: ({ row }) => {
      const repoUrl = row.getValue("repo_url") as string;
      return (
        <div className="max-w-[200px] truncate font-mono text-sm">
          {repoUrl}
        </div>
      );
    },
  },
  {
    id: "hasCredential",
    accessorFn: (row) => row.credential_id ? "true" : "false",
    header: t("projects.credential"),
    cell: ({ row }) => {
      const hasCredential = row.getValue("hasCredential") === "true";
      return (
        <Badge variant={hasCredential ? "success" : "secondary"}>
          {hasCredential ? t("common.yes") : t("common.no")}
        </Badge>
      );
    },
    filterFn: (row, id, value) => {
      return value.includes(row.getValue(id));
    },
  },
  {
    accessorKey: "task_count",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t("projects.taskCount")} />
    ),
    cell: ({ row }) => {
      const count = row.getValue("task_count") as number;
      return (
        <Badge variant="secondary" className="font-mono">
          {count || 0}
        </Badge>
      );
    },
    enableSorting: true,
  },
  {
    accessorKey: "created_at",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t("common.created")} />
    ),
    cell: ({ row }) => {
      const date = new Date(row.getValue("created_at"));
      return (
        <div className="text-sm text-muted-foreground">
          {date.toLocaleDateString()}
        </div>
      );
    },
    enableSorting: true,
  },
  {
    id: "actions",
    header: t("common.actions"),
    cell: ({ row }) => {
      const project = row.original;

      return (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="h-8 w-8 p-0">
              <span className="sr-only">{t("common.open_menu")}</span>
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>{t("common.actions")}</DropdownMenuLabel>
            <DropdownMenuItem onClick={() => onManageTasks(project)}>
              <FolderOpen className="mr-2 h-4 w-4" />
              {t("projects.tasksManagement")}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onEdit(project)}>
              <Edit className="mr-2 h-4 w-4" />
              {t("common.edit")}
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => onDelete(project.id)}
              className="text-destructive"
            >
              <Trash2 className="mr-2 h-4 w-4" />
              {t("common.delete")}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      );
    },
    enableSorting: false,
    enableHiding: false,
  },
];
