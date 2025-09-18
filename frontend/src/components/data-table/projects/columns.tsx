import type { ColumnDef } from "@tanstack/react-table";
import { Edit, Columns, Users, Bell } from "lucide-react";
import type { TFunction } from "i18next";
import { QuickActions } from "@/components/ui/quick-actions";

import { Badge } from "@/components/ui/badge";
import { DataTableColumnHeader } from "@/components/ui/data-table/data-table-column-header";
import type { Project } from "@/types/project";

interface ProjectColumnsProps {
  t: TFunction;
  onEdit: (project: Project) => void;
  onDelete: (id: number) => void;
  onKanban: (project: Project) => void;
  onManageAdmins: (project: Project) => void;
  onManageNotifiers: (project: Project) => void;
  canEditProject: (resourceAdminId?: number) => boolean;
  canDeleteProject: (resourceAdminId?: number) => boolean;
}

export const createProjectColumns = ({
  t,
  onEdit,
  onDelete,
  onKanban,
  onManageAdmins,
  onManageNotifiers,
  canEditProject,
  canDeleteProject,
}: ProjectColumnsProps): ColumnDef<Project>[] => [
  {
    accessorKey: "name",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t("projects.name")} />
    ),
    cell: ({ row }) => {
      const project = row.original;
      return (
        <button
          onClick={() => onKanban(project)}
          className="font-medium text-left text-primary hover:text-primary/80 underline hover:no-underline cursor-pointer transition-colors underline-offset-4"
        >
          {row.getValue("name")}
        </button>
      );
    },
    enableSorting: true,
    enableHiding: false,
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
    accessorKey: "admin_count",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t("projects.adminCount")} />
    ),
    cell: ({ row }) => {
      const count = row.getValue("admin_count") as number;
      return (
        <Badge variant="outline" className="font-mono">
          {count || 0}
        </Badge>
      );
    },
    enableSorting: true,
  },
  {
    accessorKey: "created_by",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t("projects.createdBy")} />
    ),
    cell: ({ row }) => {
      const createdBy = row.getValue("created_by") as string;
      return (
        <div className="text-sm">
          {createdBy}
        </div>
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
    meta: {
      headerClassName: "sticky right-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-sm",
      cellClassName: "sticky right-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-sm",
    },
    cell: ({ row }) => {
      const project = row.original;

      const actions = [
        {
          id: "kanban",
          label: t("projects.kanban"),
          icon: Columns,
          onClick: () => onKanban(project),
        },
        // Only show edit action if user has permission
        ...(canEditProject(project.admin_id) ? [{
          id: "edit",
          label: t("common.edit"),
          icon: Edit,
          onClick: () => onEdit(project),
        }] : []),
        // Only show manage admins action if user has permission
        ...(canEditProject(project.admin_id) ? [{
          id: "manage-admins",
          label: t("projects.admin.manage"),
          icon: Users,
          onClick: () => onManageAdmins(project),
        }] : []),
        // Only show manage notifiers action if user has permission
        ...(canEditProject(project.admin_id) ? [{
          id: "manage-notifiers",
          label: t("projects.notifier.manage"),
          icon: Bell,
          onClick: () => onManageNotifiers(project),
        }] : []),
      ];

      // Only show delete action if user has permission
      const deleteAction = canDeleteProject(project.admin_id) ? {
        title: project.name,
        confirmationValue: project.name,
        submitAction: async () => {
          await onDelete(project.id);
        },
      } : undefined;

      return (
        <QuickActions 
          actions={actions} 
          deleteAction={deleteAction}
        />
      );
    },
    enableSorting: false,
    enableHiding: false,
  },
];
