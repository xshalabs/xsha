import type { ColumnDef } from "@tanstack/react-table";
import { Edit, Key } from "lucide-react";
import type { TFunction } from "i18next";
import { QuickActions } from "@/components/ui/quick-actions";
import { Badge } from "@/components/ui/badge";
import { DataTableColumnHeader } from "@/components/ui/data-table/data-table-column-header";
import { UserAvatar } from "@/components/ui/user-avatar";
import { formatDateTime } from "@/lib/utils";
import type { Admin, AdminRole } from "@/lib/api";

interface AdminColumnsProps {
  t: TFunction;
  onEdit: (admin: Admin) => void;
  onChangePassword: (admin: Admin) => void;
  onDelete: (admin: Admin) => Promise<void>;
  onAvatarClick?: (admin: Admin) => void;
  permissions?: {
    canEditAdmin: (username?: string) => boolean;
    canChangeAdminPassword: (username?: string, role?: AdminRole) => boolean;
    canDeleteAdmin: (role?: AdminRole, createdBy?: string) => boolean;
  };
}

export const createAdminColumns = ({
  t,
  onEdit,
  onChangePassword,
  onDelete,
  onAvatarClick,
  permissions,
}: AdminColumnsProps): ColumnDef<Admin>[] => [
  {
    accessorKey: "username",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t("admin.table.username")} />
    ),
    cell: ({ row }) => {
      const admin = row.original;
      return (
        <div className="flex items-center space-x-3">
          <UserAvatar 
            user={admin.username}
            name={admin.name}
            avatar={admin.avatar}
            size="sm"
            onClick={onAvatarClick ? () => onAvatarClick(admin) : undefined}
          />
          <div className="font-medium">{row.getValue("username")}</div>
        </div>
      );
    },
    enableSorting: true,
    enableHiding: false,
  },
  {
    accessorKey: "name",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t("admin.table.name")} />
    ),
    cell: ({ row }) => (
      <div className="font-medium">{row.getValue("name")}</div>
    ),
    enableSorting: true,
    enableHiding: false,
  },
  {
    accessorKey: "email",
    header: t("admin.table.email"),
    cell: ({ row }) => {
      const email = row.getValue("email") as string;
      return (
        <div className="text-muted-foreground">
          {email || "-"}
        </div>
      );
    },
  },
  {
    accessorKey: "role",
    header: t("admin.table.role"),
    cell: ({ row }) => {
      const role = row.getValue("role") as string;
      const roleVariant = role === 'super_admin' ? 'destructive' : 
                         role === 'admin' ? 'default' : 'secondary';
      return (
        <Badge variant={roleVariant}>
          {t(`admin.roles.${role}`)}
        </Badge>
      );
    },
    filterFn: (row, id, value) => {
      if (!Array.isArray(value) || value.length === 0) {
        return true;
      }
      const role = row.getValue(id) as string;
      return value.includes(role);
    },
  },
  {
    accessorKey: "is_active",
    header: t("admin.table.status"),
    cell: ({ row }) => {
      const isActive = row.getValue("is_active") as boolean;
      return (
        <Badge variant={isActive ? "default" : "secondary"}>
          {isActive ? t("admin.status.active") : t("admin.status.inactive")}
        </Badge>
      );
    },
    filterFn: (row, id, value) => {
      // value 是一个数组，如 ["active"] 或 ["inactive"] 或 ["active", "inactive"]
      if (!Array.isArray(value) || value.length === 0) {
        return true; // 没有筛选时显示所有
      }
      
      const isActive = row.getValue(id) as boolean;
      
      // 检查当前行的状态是否在选中的值中
      return value.includes(isActive ? "active" : "inactive");
    },
  },
  {
    accessorKey: "last_login_at",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t("admin.table.lastLogin")} />
    ),
    cell: ({ row }) => {
      const lastLogin = row.getValue("last_login_at") as string | undefined;
      const admin = row.original;
      return (
        <div className="space-y-1">
          {lastLogin ? (
            <>
              <div className="text-sm">{formatDateTime(lastLogin)}</div>
              {admin.last_login_ip && (
                <div className="text-xs text-muted-foreground">IP: {admin.last_login_ip}</div>
              )}
            </>
          ) : (
            <span className="text-muted-foreground">-</span>
          )}
        </div>
      );
    },
    enableSorting: true,
  },
  {
    accessorKey: "created_by",
    header: t("admin.table.createdBy"),
    cell: ({ row }) => (
      <div className="text-sm text-muted-foreground">
        {row.getValue("created_by")}
      </div>
    ),
  },
  {
    accessorKey: "created_at",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t("admin.table.createdAt")} />
    ),
    cell: ({ row }) => (
      <div className="text-sm text-muted-foreground">
        {formatDateTime(row.getValue("created_at"))}
      </div>
    ),
    enableSorting: true,
  },
  {
    id: "actions",
    meta: {
      headerClassName: "sticky right-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-sm",
      cellClassName: "sticky right-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-sm",
    },
    cell: ({ row }) => {
      const admin = row.original;

      const actions = [];

      // Edit action - check permission
      if (!permissions || permissions.canEditAdmin(admin.username)) {
        actions.push({
          id: "edit",
          label: t("common.edit"),
          icon: Edit,
          onClick: () => onEdit(admin),
        });
      }

      // Change password action - check permission
      if (!permissions || permissions.canChangeAdminPassword(admin.username, admin.role)) {
        actions.push({
          id: "changePassword",
          label: t("admin.actions.changePassword"),
          icon: Key,
          onClick: () => onChangePassword(admin),
        });
      }

      // Delete action - check permission and system constraint
      const canDelete = permissions 
        ? permissions.canDeleteAdmin(admin.role, admin.created_by)
        : admin.created_by !== 'system';

      const deleteAction = canDelete ? {
        title: admin.username,
        confirmationValue: admin.username,
        submitAction: async () => {
          await onDelete(admin);
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