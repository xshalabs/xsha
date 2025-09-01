import type { ColumnDef } from "@tanstack/react-table";
import { Edit } from "lucide-react";
import { QuickActions } from "@/components/ui/quick-actions";
import { Badge } from "@/components/ui/badge";
import type { DevEnvironmentDisplay } from "@/types/dev-environment";

interface DevEnvironmentColumnsProps {
  onEdit: (environment: DevEnvironmentDisplay) => void;
  onDelete: (id: number) => void;
  t: (key: string) => string;
  canEditEnvironment: (createdBy?: string) => boolean;
  canDeleteEnvironment: (createdBy?: string) => boolean;
}

export const createDevEnvironmentColumns = ({
  onEdit,
  onDelete,
  t,
  canEditEnvironment,
  canDeleteEnvironment,
}: DevEnvironmentColumnsProps): ColumnDef<DevEnvironmentDisplay>[] => [
  {
    accessorKey: "name",
    header: t("devEnvironments.table.name"),
    cell: ({ row }) => (
      <div className="font-medium">{row.getValue("name")}</div>
    ),
  },
  {
    accessorKey: "description",
    header: t("devEnvironments.table.description"),
    cell: ({ row }) => {
      const description = row.getValue("description") as string;
      return (
        <div className="max-w-[300px] truncate text-muted-foreground">
          {description || t("devEnvironments.table.no_description")}
        </div>
      );
    },
  },
  {
    accessorKey: "cpu_limit",
    header: t("devEnvironments.table.cpu"),
    cell: ({ row }) => {
      const cores = row.getValue("cpu_limit") as number;
      const coreCount = cores || 0;
      return (
        <Badge variant="secondary">
          {coreCount} {coreCount === 1 ? t("devEnvironments.table.core") : t("devEnvironments.table.cores")}
        </Badge>
      );
    },
  },
  {
    accessorKey: "memory_limit",
    header: t("devEnvironments.table.memory"),
    cell: ({ row }) => {
      const memoryMb = row.getValue("memory_limit") as number;
      const memory = memoryMb || 0;
      const formatMemory = (mb: number) => {
        if (mb >= 1024) {
          return `${(mb / 1024).toFixed(1)} GB`;
        }
        return `${mb} MB`;
      };
      return (
        <Badge variant="secondary">
          {formatMemory(memory)}
        </Badge>
      );
    },
  },
  {
    id: "env_vars_count",
    header: t("devEnvironments.table.env_vars"),
    cell: ({ row }) => {
      const envVarsMap = row.original.env_vars_map || {};
      const count = Object.keys(envVarsMap).length;
      return (
        <Badge variant="outline">
          {count}
        </Badge>
      );
    },
  },
  {
    accessorKey: "created_at",
    header: t("devEnvironments.table.created"),
    cell: ({ row }) => {
      const date = new Date(row.getValue("created_at"));
      return (
        <div className="text-sm text-muted-foreground">
          {date.toLocaleDateString()}
        </div>
      );
    },
  },
  {
    id: "actions",
    meta: {
      headerClassName: "sticky right-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-sm",
      cellClassName: "sticky right-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-sm",
    },
    cell: ({ row }) => {
      const environment = row.original;

      // Only show edit action if user has permission
      const actions = canEditEnvironment(environment.created_by) ? [
        {
          id: "edit",
          label: t("devEnvironments.edit"),
          icon: Edit,
          onClick: () => onEdit(environment),
        },
      ] : [];

      // Only show delete action if user has permission
      const deleteAction = canDeleteEnvironment(environment.created_by) ? {
        title: environment.name,
        confirmationValue: environment.name,
        submitAction: async () => {
          await onDelete(environment.id);
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
