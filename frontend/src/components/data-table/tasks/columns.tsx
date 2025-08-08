import { ColumnDef } from "@tanstack/react-table";
import { MoreHorizontal, Edit, Trash2, MessageSquare, GitCompare, GitBranch, CheckCircle, Clock, Play, X } from "lucide-react";
import { TFunction } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

import { DataTableColumnHeader } from "@/components/ui/data-table/data-table-column-header";
import { Checkbox } from "@/components/ui/checkbox";
import { Task, TaskStatus } from "@/types/task";

interface TaskColumnsProps {
  t: TFunction;
  onEdit: (task: Task) => void;
  onDelete: (id: number) => void;
  onViewConversation?: (task: Task) => void;
  onViewGitDiff?: (task: Task) => void;
  onPushBranch?: (task: Task) => void;
  hideProjectColumn?: boolean;
}

const getStatusIcon = (status: TaskStatus) => {
  switch (status) {
    case "todo":
      return <Clock className="w-4 h-4 text-gray-500" />;
    case "in_progress":
      return <Play className="w-4 h-4 text-blue-500" />;
    case "done":
      return <CheckCircle className="w-4 h-4 text-green-500" />;
    case "cancelled":
      return <X className="w-4 h-4 text-red-500" />;
    default:
      return <Clock className="w-4 h-4 text-gray-500" />;
  }
};

const getStatusTextColor = (status: TaskStatus) => {
  switch (status) {
    case "todo":
      return "text-gray-600";
    case "in_progress":
      return "text-blue-600";
    case "done":
      return "text-green-600";
    case "cancelled":
      return "text-red-600";
    default:
      return "text-gray-600";
  }
};

export const createTaskColumns = ({
  t,
  onEdit,
  onDelete,
  onViewConversation,
  onViewGitDiff,
  onPushBranch,
  hideProjectColumn = false,
}: TaskColumnsProps): ColumnDef<Task>[] => {
  const columns: ColumnDef<Task>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllPageRowsSelected() ||
            (table.getIsSomePageRowsSelected() && "indeterminate")
          }
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select row"
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "title",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t("tasks.table.title")} />
      ),
      cell: ({ row }) => {
        const task = row.original;
        return (
          <div className="max-w-[300px]">
            <div
              className="font-medium text-blue-600 hover:text-blue-800 underline cursor-pointer transition-colors truncate"
              onClick={() => onViewConversation?.(task)}
            >
              {task.title}
            </div>
            <div className="text-xs text-muted-foreground">
              {t("common.createdAt")}: {new Date(task.created_at).toLocaleString()}
            </div>
          </div>
        );
      },
      enableSorting: true,
      enableHiding: false,
    },
  ];

  // Add project column conditionally
  if (!hideProjectColumn) {
    columns.push({
      accessorKey: "project.name",
      header: t("tasks.table.project"),
      cell: ({ row }) => {
        const projectName = row.original.project?.name;
        return projectName ? (
          <span className="text-blue-600">{projectName}</span>
        ) : (
          <span className="text-muted-foreground">-</span>
        );
      },
      enableSorting: false,
    });
  }

  // Add remaining columns
  columns.push(
    {
      accessorKey: "status",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t("tasks.table.status")} />
      ),
      cell: ({ row }) => {
        const status = row.getValue("status") as TaskStatus;
        const statusDisplay = {
          todo: t("tasks.status.todo"),
          in_progress: t("tasks.status.in_progress"),
          done: t("tasks.status.done"),
          cancelled: t("tasks.status.cancelled"),
        }[status] || status;

        return (
          <div className="flex items-center space-x-2">
            {getStatusIcon(status)}
            <span className={`text-sm ${getStatusTextColor(status)}`}>
              {statusDisplay}
            </span>
          </div>
        );
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
      enableSorting: false,
    },
    {
      accessorKey: "conversation_count",
      header: t("tasks.table.conversations"),
      cell: ({ row }) => {
        const count = row.getValue("conversation_count") as number;
        return (
          <div className="flex items-center space-x-2">
            <MessageSquare className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm font-medium">{count}</span>
          </div>
        );
      },
      enableSorting: true,
    },
    {
      accessorKey: "start_branch",
      header: t("tasks.table.branch"),
      cell: ({ row }) => {
        const task = row.original;
        return (
          <div className="flex items-center space-x-2">
            <GitBranch className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm">{task.start_branch}</span>
            {task.has_pull_request && (
              <Badge variant="outline" className="text-xs">
                <GitBranch className="w-3 h-3 mr-1" />
                PR
              </Badge>
            )}
          </div>
        );
      },
      enableSorting: false,
    },
    {
      accessorKey: "dev_environment.name",
      header: t("tasks.table.environment"),
      cell: ({ row }) => {
        const devEnv = row.original.dev_environment;
        return devEnv ? (
          <div className="flex items-center space-x-2">
            <div className="w-2 h-2 rounded-full bg-blue-500"></div>
            <span className="text-sm">{devEnv.name}</span>
          </div>
        ) : (
          <span className="text-xs text-muted-foreground">-</span>
        );
      },
      filterFn: (row, id, value) => {
        const devEnvId = row.original.dev_environment?.id;
        return value.includes(devEnvId?.toString() || "");
      },
      enableSorting: false,
    },
    {
      accessorKey: "updated_at",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t("tasks.table.updated")} />
      ),
      cell: ({ row }) => {
        const date = new Date(row.getValue("updated_at"));
        return (
          <div className="text-xs text-muted-foreground">
            {date.toLocaleString()}
          </div>
        );
      },
      enableSorting: true,
    },
    {
      id: "actions",
      header: t("common.actions"),
      cell: ({ row }) => {
        const task = row.original;

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

              {onViewConversation && (
                <DropdownMenuItem onClick={() => onViewConversation(task)}>
                  <MessageSquare className="h-4 w-4 mr-2" />
                  {t("tasks.actions.viewConversation")}
                </DropdownMenuItem>
              )}

              {onViewGitDiff && task.work_branch && (
                <DropdownMenuItem onClick={() => onViewGitDiff(task)}>
                  <GitCompare className="h-4 w-4 mr-2" />
                  {t("tasks.actions.viewGitDiff")}
                </DropdownMenuItem>
              )}

              {onPushBranch && (
                <DropdownMenuItem onClick={() => onPushBranch(task)}>
                  <GitBranch className="h-4 w-4 mr-2" />
                  {t("tasks.actions.pushBranch")}
                </DropdownMenuItem>
              )}

              <DropdownMenuItem onClick={() => onEdit(task)}>
                <Edit className="h-4 w-4 mr-2" />
                {t("common.edit")}
              </DropdownMenuItem>

              <DropdownMenuItem
                onClick={() => onDelete(task.id)}
                className="text-destructive"
              >
                <Trash2 className="h-4 w-4 mr-2" />
                {t("common.delete")}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
      enableSorting: false,
      enableHiding: false,
    }
  );

  return columns;
};
