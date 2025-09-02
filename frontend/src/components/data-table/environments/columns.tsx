import type { ColumnDef } from "@tanstack/react-table";
import { Edit, UserCog } from "lucide-react";
import { QuickActions } from "@/components/ui/quick-actions";
import { Badge } from "@/components/ui/badge";
import { DropdownMenuItem } from "@/components/ui/dropdown-menu";
import { AdminManagementSheet } from "@/components/environments/AdminManagementSheet";
import type { DevEnvironment } from "@/types/dev-environment";

interface DevEnvironmentColumnsProps {
  onEdit: (environment: DevEnvironment) => void;
  onDelete: (id: number) => void;
  onAdminChanged: () => void;
  t: (key: string) => string;
  canEditEnvironment: (resourceAdminId?: number, isEnvironmentAdmin?: boolean) => boolean;
  canDeleteEnvironment: (resourceAdminId?: number, isEnvironmentAdmin?: boolean) => boolean;
  canManageEnvironmentAdmins: (resourceAdminId?: number, isEnvironmentAdmin?: boolean) => boolean;
  currentAdminId?: number;
}

export const createDevEnvironmentColumns = ({
  onEdit,
  onDelete,
  onAdminChanged,
  t,
  canEditEnvironment,
  canDeleteEnvironment,
  canManageEnvironmentAdmins,
  currentAdminId,
}: DevEnvironmentColumnsProps): ColumnDef<DevEnvironment>[] => {
  // Helper function to check if current user is an admin of the environment
  const isEnvironmentAdmin = (environment: DevEnvironment): boolean => {
    if (!currentAdminId) return false;
    
    // Check if user is the legacy admin_id owner
    const isLegacyAdmin = environment.admin_id === currentAdminId;
    
    return isLegacyAdmin;
  };

  return [
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
    id: "admin_count",
    header: t("devEnvironments.table.admins"),
    cell: ({ row }) => {
      const admins = row.original.admins || [];
      const count = admins.length;
      return (
        <Badge variant="secondary">
          {count} {count === 1 ? t("devEnvironments.table.admin") : t("devEnvironments.table.admins")}
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

      const actions = [];

      // Only show edit action if user has permission
      if (canEditEnvironment(environment.admin_id, isEnvironmentAdmin(environment))) {
        actions.push({
          id: "edit",
          label: t("devEnvironments.edit"),
          icon: Edit,
          onClick: () => onEdit(environment),
        });
      }

      // Only show manage admins action if user has permission
      if (canManageEnvironmentAdmins(environment.admin_id, isEnvironmentAdmin(environment))) {
        actions.push({
        id: "manage-admins",
        label: t("devEnvironments.admin.manage"),
        icon: UserCog,
        render: () => (
          <AdminManagementSheet
            environment={environment}
            onAdminChanged={onAdminChanged}
            trigger={
              <DropdownMenuItem 
                className="cursor-pointer"
              >
                <UserCog className="mr-2 h-4 w-4" />
                <span>{t("devEnvironments.admin.manage")}</span>
              </DropdownMenuItem>
            }
          />
        ),
        });
      }

      // Only show delete action if user has permission
      const deleteAction = canDeleteEnvironment(environment.admin_id, isEnvironmentAdmin(environment)) ? {
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
};
