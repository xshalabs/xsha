import type { ColumnDef } from "@tanstack/react-table";
import { Edit, Key } from "lucide-react";
import type { TFunction } from "i18next";
import { QuickActions } from "@/components/ui/quick-actions";
import { Badge } from "@/components/ui/badge";
import { DataTableColumnHeader } from "@/components/ui/data-table/data-table-column-header";
import { UserAvatar } from "@/components/ui/user-avatar";
import { formatDateTime } from "@/lib/utils";
import type { Admin } from "@/lib/api";

interface AdminColumnsProps {
  t: TFunction;
  onEdit: (admin: Admin) => void;
  onChangePassword: (admin: Admin) => void;
  onDelete: (admin: Admin) => Promise<void>;
}

export const createAdminColumns = ({
  t,
  onEdit,
  onChangePassword,
  onDelete,
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

      const actions = [
        {
          id: "edit",
          label: t("common.edit"),
          icon: Edit,
          onClick: () => onEdit(admin),
        },
        {
          id: "changePassword",
          label: t("admin.actions.changePassword"),
          icon: Key,
          onClick: () => onChangePassword(admin),
        },
      ];

      // Only allow deletion if not created by system
      const deleteAction = admin.created_by !== 'system' ? {
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