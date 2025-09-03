import type { ColumnDef } from "@tanstack/react-table";
import { Edit, Key, Shield, User, Users } from "lucide-react";
import { QuickActions } from "@/components/ui/quick-actions";
import { Badge } from "@/components/ui/badge";
import type { GitCredential } from "@/types/credentials";
import { GitCredentialType } from "@/types/credentials";
import { CredentialAdminManagementSheet } from "@/components/credentials/CredentialAdminManagementSheet";

interface GitCredentialColumnsProps {
  onEdit: (credential: GitCredential) => void;
  onDelete: (id: number) => void;
  t: (key: string) => string;
  canEditCredential: (resourceAdminId?: number) => boolean;
  canDeleteCredential: (resourceAdminId?: number) => boolean;
  canManageCredentialAdmins: (resourceAdminId?: number) => boolean;
}

export const createGitCredentialColumns = ({
  onEdit,
  onDelete,
  t,
  canEditCredential,
  canDeleteCredential,
  canManageCredentialAdmins,
}: GitCredentialColumnsProps): ColumnDef<GitCredential>[] => [
  {
    accessorKey: "name",
    header: t("gitCredentials.columns.name"),
    cell: ({ row }) => (
      <div className="font-medium">{row.getValue("name")}</div>
    ),
  },
  {
    accessorKey: "description",
    header: t("gitCredentials.columns.description"),
    cell: ({ row }) => {
      const description = row.getValue("description") as string;
      return (
        <div className="max-w-[300px] truncate text-muted-foreground">
          {description || t("gitCredentials.columns.noDescription")}
        </div>
      );
    },
  },
  {
    accessorKey: "type",
    header: t("gitCredentials.columns.type"),
    cell: ({ row }) => {
      const type = row.getValue("type") as GitCredentialType;
      const getTypeIcon = () => {
        switch (type) {
          case GitCredentialType.PASSWORD:
            return <Key className="w-4 h-4" />;
          case GitCredentialType.TOKEN:
            return <Shield className="w-4 h-4" />;
          case GitCredentialType.SSH_KEY:
            return <User className="w-4 h-4" />;
          default:
            return <Key className="w-4 h-4" />;
        }
      };

      const getTypeName = () => {
        switch (type) {
          case GitCredentialType.PASSWORD:
            return t("gitCredentials.types.password");
          case GitCredentialType.TOKEN:
            return t("gitCredentials.types.token");
          case GitCredentialType.SSH_KEY:
            return t("gitCredentials.types.ssh_key");
          default:
            return "Unknown";
        }
      };

      return (
        <Badge variant="secondary" className="flex items-center gap-1">
          {getTypeIcon()}
          {getTypeName()}
        </Badge>
      );
    },
  },
  {
    accessorKey: "username",
    header: t("gitCredentials.columns.username"),
    cell: ({ row }) => {
      const username = row.getValue("username") as string;
      return (
        <div className="text-sm text-muted-foreground">{username || "N/A"}</div>
      );
    },
  },
  {
    accessorKey: "created_at",
    header: t("gitCredentials.columns.created"),
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
      const credential = row.original;

      // Build actions based on permissions
      const actions = [];

      // Add edit action if user has permission
      if (canEditCredential(credential.admin_id)) {
        actions.push({
          id: "edit",
          label: t("gitCredentials.edit"),
          icon: Edit,
          onClick: () => onEdit(credential),
        });
      }

      // Add manage admins action if user has permission
      if (canManageCredentialAdmins(credential.admin_id)) {
        actions.push({
          id: "manageAdmins",
          label: t("gitCredentials.manageAdmins"),
          icon: Users,
          render: () => (
            <CredentialAdminManagementSheet
              credential={credential}
              trigger={
                <button className="flex items-center gap-2 w-full px-2 py-1.5 text-left text-sm hover:bg-accent rounded-sm">
                  <Users className="h-4 w-4" />
                  {t("gitCredentials.manageAdmins")}
                </button>
              }
            />
          ),
        });
      }

      // Only show delete action if user has permission
      const deleteAction = canDeleteCredential(credential.admin_id) ? {
        title: credential.name,
        confirmationValue: credential.name,
        submitAction: async () => {
          await onDelete(credential.id);
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
