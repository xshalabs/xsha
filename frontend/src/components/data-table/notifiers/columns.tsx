import type { ColumnDef } from "@tanstack/react-table";
import {
  Edit,
  TestTube,
  MessageSquare,
  MessageCircle,
  Slack,
  Webhook,
  Hash,
  ToggleLeft,
  ToggleRight,
} from "lucide-react";
import { QuickActions } from "@/components/ui/quick-actions";
import { Badge } from "@/components/ui/badge";
import type { Notifier } from "@/types/notifier";
import { NotifierType } from "@/types/notifier";

interface NotifierColumnsProps {
  onEdit: (notifier: Notifier) => void;
  onDelete: (id: number) => void;
  onTest: (id: number) => void;
  onToggleStatus: (id: number, enabled: boolean) => void;
  t: (key: string) => string;
  canEditNotifier: (resourceAdminId?: number) => boolean;
  canDeleteNotifier: (resourceAdminId?: number) => boolean;
  canTestNotifier: (resourceAdminId?: number) => boolean;
}

export const createNotifierColumns = ({
  onEdit,
  onDelete,
  onTest,
  onToggleStatus,
  t,
  canEditNotifier,
  canDeleteNotifier,
  canTestNotifier,
}: NotifierColumnsProps): ColumnDef<Notifier>[] => [
  {
    id: "search",
    accessorFn: (row) => `${row.name} ${row.description || ""}`,
    header: () => null,
    cell: () => null,
    enableSorting: false,
    enableHiding: false,
    enableColumnFilter: true,
    filterFn: (row, _, value) => {
      if (!value) return true;

      const searchValue = value.toLowerCase();
      const name = row.original.name.toLowerCase();
      const description = (row.original.description || "").toLowerCase();

      return name.includes(searchValue) ||
             description.includes(searchValue);
    },
    size: 0,
    maxSize: 0,
  },
  {
    accessorKey: "name",
    header: t("notifiers.columns.name"),
    cell: ({ row }) => (
      <div className="font-medium">{row.getValue("name")}</div>
    ),
  },
  {
    accessorKey: "description",
    header: t("notifiers.columns.description"),
    cell: ({ row }) => {
      const description = row.getValue("description") as string;
      return (
        <div className="max-w-[300px] truncate text-muted-foreground">
          {description || t("notifiers.columns.noDescription")}
        </div>
      );
    },
  },
  {
    accessorKey: "type",
    header: t("notifiers.columns.type"),
    cell: ({ row }) => {
      const type = row.getValue("type") as NotifierType;
      const getTypeIcon = () => {
        switch (type) {
          case NotifierType.WECHAT_WORK:
            return <MessageSquare className="w-4 h-4" />;
          case NotifierType.DINGTALK:
            return <MessageCircle className="w-4 h-4" />;
          case NotifierType.FEISHU:
            return <MessageSquare className="w-4 h-4" />;
          case NotifierType.SLACK:
            return <Slack className="w-4 h-4" />;
          case NotifierType.DISCORD:
            return <Hash className="w-4 h-4" />;
          case NotifierType.WEBHOOK:
            return <Webhook className="w-4 h-4" />;
          default:
            return <Webhook className="w-4 h-4" />;
        }
      };

      const getTypeName = () => {
        switch (type) {
          case NotifierType.WECHAT_WORK:
            return t("notifiers.types.wechat_work");
          case NotifierType.DINGTALK:
            return t("notifiers.types.dingtalk");
          case NotifierType.FEISHU:
            return t("notifiers.types.feishu");
          case NotifierType.SLACK:
            return t("notifiers.types.slack");
          case NotifierType.DISCORD:
            return t("notifiers.types.discord");
          case NotifierType.WEBHOOK:
            return t("notifiers.types.webhook");
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
    filterFn: (row, id, value) => {
      if (!Array.isArray(value) || value.length === 0) {
        return true;
      }
      const type = row.getValue(id) as string;
      return value.includes(type);
    },
  },
  {
    accessorKey: "is_enabled",
    header: t("notifiers.columns.status"),
    cell: ({ row }) => {
      const isEnabled = row.getValue("is_enabled") as boolean;
      return (
        <Badge variant={isEnabled ? "default" : "secondary"} className="flex items-center gap-1">
          {isEnabled ? (
            <ToggleRight className="w-4 h-4" />
          ) : (
            <ToggleLeft className="w-4 h-4" />
          )}
          {isEnabled ? t("notifiers.status.enabled") : t("notifiers.status.disabled")}
        </Badge>
      );
    },
    filterFn: (row, id, value) => {
      if (!Array.isArray(value) || value.length === 0) {
        return true;
      }

      const isEnabled = row.getValue(id) as boolean;
      return value.includes(isEnabled ? "enabled" : "disabled");
    },
  },
  {
    accessorKey: "created_by",
    header: t("notifiers.columns.createdBy"),
    cell: ({ row }) => {
      const createdBy = row.getValue("created_by") as string;
      return (
        <div className="text-sm text-muted-foreground">{createdBy || "N/A"}</div>
      );
    },
  },
  {
    accessorKey: "created_at",
    header: t("notifiers.columns.created"),
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
      const notifier = row.original;

      // Build actions based on permissions
      const actions = [];

      // Add test action if user has permission
      if (canTestNotifier(notifier.admin_id)) {
        actions.push({
          id: "test",
          label: t("notifiers.test"),
          icon: TestTube,
          onClick: () => onTest(notifier.id),
        });
      }

      // Add edit action if user has permission
      if (canEditNotifier(notifier.admin_id)) {
        actions.push({
          id: "edit",
          label: t("notifiers.edit"),
          icon: Edit,
          onClick: () => onEdit(notifier),
        });
      }

      // Add toggle status action if user has permission
      if (canEditNotifier(notifier.admin_id)) {
        actions.push({
          id: "toggle",
          label: notifier.is_enabled
            ? t("notifiers.disable")
            : t("notifiers.enable"),
          icon: notifier.is_enabled ? ToggleLeft : ToggleRight,
          onClick: () => onToggleStatus(notifier.id, !notifier.is_enabled),
        });
      }

      // Only show delete action if user has permission
      const deleteAction = canDeleteNotifier(notifier.admin_id) ? {
        title: notifier.name,
        confirmationValue: notifier.name,
        submitAction: async () => {
          await onDelete(notifier.id);
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