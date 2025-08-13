import type { ColumnDef } from "@tanstack/react-table";
import { Edit, MessageSquare, GitCompare, GitBranch, Copy } from "lucide-react";
import { useTranslation } from "react-i18next";

type TFunction = ReturnType<typeof useTranslation>['t'];
import { QuickActions } from "@/components/ui/quick-actions";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

import { DataTableColumnHeader } from "@/components/ui/data-table/data-table-column-header";
import { Checkbox } from "@/components/ui/checkbox";
import type { Task, TaskStatus } from "@/types/task";

interface TaskColumnsProps {
  t: TFunction;
  onEdit: (task: Task) => void;
  onDelete: (id: number) => void;
  onViewConversation?: (task: Task) => void;
  onViewGitDiff?: (task: Task) => void;
  onPushBranch?: (task: Task) => void;
  hideProjectColumn?: boolean;
}



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
              className="font-medium text-primary hover:text-primary/80 underline cursor-pointer transition-colors truncate"
              onClick={() => onViewConversation?.(task)}
            >
              {task.title}
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
          <span className={`text-sm font-medium ${getStatusTextColor(status)}`}>
            {statusDisplay}
          </span>
        );
      },
      filterFn: (row, _id, value) => {
        return value.includes(row.getValue(_id));
      },
      enableSorting: true,
    },
    {
      accessorKey: "conversation_count",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t("tasks.table.conversations")} />
      ),
      cell: ({ row }) => {
        const count = row.getValue("conversation_count") as number;
        return (
          <span className="text-sm font-medium">{count}</span>
        );
      },
      enableSorting: true,
    },
    {
      accessorKey: "start_branch",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t("tasks.table.branch")} />
      ),
      cell: ({ row }) => {
        const task = row.original;
        return (
          <div className="flex items-center space-x-2">
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
      enableSorting: true,
    },
    {
      accessorKey: "work_branch",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t("tasks.table.workBranch")} />
      ),
      cell: ({ row }) => {
        const task = row.original;
        
        if (!task.work_branch) {
          return <span className="text-xs text-muted-foreground">-</span>;
        }
        
        const copyToClipboard = async () => {
          try {
            await navigator.clipboard.writeText(task.work_branch);
            toast.success(t("common.copied_to_clipboard"));
          } catch (err) {
            toast.error(t("common.copy_failed"));
          }
        };

        return (
          <div className="flex items-center gap-2 max-w-[200px]">
            <span className="text-sm font-mono truncate">
              {task.work_branch}
            </span>
            <Button
              variant="ghost"
              size="sm"
              onClick={copyToClipboard}
              className="h-6 w-6 p-0 hover:bg-muted/50 flex-shrink-0"
              title={t("common.copy")}
            >
              <Copy className="h-3 w-3" />
            </Button>
          </div>
        );
      },
      enableSorting: true,
    },
    {
      id: "dev_environment.name",
      accessorFn: (row) => row.dev_environment?.name || "",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t("tasks.table.environment")} />
      ),
      cell: ({ row }) => {
        const devEnv = row.original.dev_environment;
        return devEnv ? (
          <span className="text-sm">{devEnv.name}</span>
        ) : (
          <span className="text-xs text-muted-foreground">-</span>
        );
      },
      filterFn: (row, _id, value) => {
        const devEnvId = row.original.dev_environment?.id;
        return value.includes(devEnvId?.toString() || "");
      },
      enableSorting: true,
    },
    {
      accessorKey: "created_at",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t("tasks.table.created")} />
      ),
      cell: ({ row }) => {
        const date = new Date(row.getValue("created_at"));
        return (
          <div className="text-xs text-muted-foreground">
            {date.toLocaleString()}
          </div>
        );
      },
      enableSorting: true,
    },
    {
      accessorKey: "latest_execution_time",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t("tasks.table.executionTime")} />
      ),
      cell: ({ row }) => {
        const executionTime = row.getValue("latest_execution_time") as string | null;
        if (!executionTime) {
          return (
            <span className="text-xs text-muted-foreground">-</span>
          );
        }
        const date = new Date(executionTime);
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
      meta: {
        headerClassName: "sticky right-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-sm",
        cellClassName: "sticky right-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-sm",
      },
      cell: ({ row }) => {
        const task = row.original;

        return (
          (() => {
            const actions = [];

            if (onViewConversation) {
              actions.push({
                id: "view-conversation",
                label: t("tasks.actions.viewConversation"),
                icon: MessageSquare,
                onClick: () => onViewConversation(task),
              });
            }

            if (onViewGitDiff && task.work_branch) {
              actions.push({
                id: "view-git-diff",
                label: t("tasks.actions.viewGitDiff"),
                icon: GitCompare,
                onClick: () => onViewGitDiff(task),
              });
            }

            if (onPushBranch) {
              actions.push({
                id: "push-branch",
                label: t("tasks.actions.pushBranch"),
                icon: GitBranch,
                onClick: () => onPushBranch(task),
              });
            }

            actions.push({
              id: "edit",
              label: t("common.edit"),
              icon: Edit,
              onClick: () => onEdit(task),
            });

            const deleteAction = {
              title: task.title,
              confirmationValue: task.title,
              submitAction: async () => {
                await onDelete(task.id);
              },
            };

            return (
              <QuickActions 
                actions={actions} 
                deleteAction={deleteAction}
                className="w-7 h-7"
              />
            );
          })()
        );
      },
      enableSorting: false,
      enableHiding: false,
    }
  );

  return columns;
};
