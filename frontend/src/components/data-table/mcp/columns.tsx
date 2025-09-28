import type { ColumnDef } from "@tanstack/react-table";
import {
  Edit,
  Settings2,
  ToggleLeft,
  ToggleRight,
} from "lucide-react";
import { QuickActions } from "@/components/ui/quick-actions";
import { Badge } from "@/components/ui/badge";
import type { MCP } from "@/types/mcp";

interface MCPColumnsProps {
  onEdit: (mcp: MCP) => void;
  onDelete: (id: number) => void;
  onToggleStatus: (id: number, enabled: boolean) => void;
  t: (key: string) => string;
  canEditMCP: (resourceAdminId?: number) => boolean;
  canDeleteMCP: (resourceAdminId?: number) => boolean;
}

export const createMCPColumns = ({
  onEdit,
  onDelete,
  onToggleStatus,
  t,
  canEditMCP,
  canDeleteMCP,
}: MCPColumnsProps): ColumnDef<MCP>[] => [
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
    header: t("mcp.columns.name"),
    cell: ({ row }) => (
      <div className="flex items-center gap-2">
        <Settings2 className="w-4 h-4 text-muted-foreground" />
        <div className="font-medium">{row.getValue("name")}</div>
      </div>
    ),
  },
  {
    accessorKey: "description",
    header: t("mcp.columns.description"),
    cell: ({ row }) => {
      const description = row.getValue("description") as string;
      return (
        <div className="max-w-[300px] truncate text-muted-foreground">
          {description || t("mcp.columns.noDescription")}
        </div>
      );
    },
  },
  {
    accessorKey: "enabled",
    header: t("mcp.columns.status"),
    cell: ({ row }) => {
      const isEnabled = row.getValue("enabled") as boolean;
      return (
        <Badge variant={isEnabled ? "default" : "secondary"} className="flex items-center gap-1">
          {isEnabled ? (
            <ToggleRight className="w-4 h-4" />
          ) : (
            <ToggleLeft className="w-4 h-4" />
          )}
          {isEnabled ? t("mcp.status.enabled") : t("mcp.status.disabled")}
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
    header: t("mcp.columns.createdBy"),
    cell: ({ row }) => {
      const createdBy = row.getValue("created_by") as string;
      return (
        <div className="text-sm text-muted-foreground">{createdBy || "N/A"}</div>
      );
    },
  },
  {
    accessorKey: "created_at",
    header: t("mcp.columns.created"),
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
      const mcp = row.original;

      // Build actions based on permissions
      const actions = [];

      // Add edit action if user has permission
      if (canEditMCP(mcp.admin_id)) {
        actions.push({
          id: "edit",
          label: t("mcp.edit"),
          icon: Edit,
          onClick: () => onEdit(mcp),
        });
      }

      // Add toggle status action if user has permission
      if (canEditMCP(mcp.admin_id)) {
        actions.push({
          id: "toggle",
          label: mcp.enabled
            ? t("mcp.disable")
            : t("mcp.enable"),
          icon: mcp.enabled ? ToggleLeft : ToggleRight,
          onClick: () => onToggleStatus(mcp.id, !mcp.enabled),
        });
      }

      // Only show delete action if user has permission
      const deleteAction = canDeleteMCP(mcp.admin_id) ? {
        title: mcp.name,
        confirmationValue: mcp.name,
        submitAction: async () => {
          await onDelete(mcp.id);
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