import type { ColumnDef } from "@tanstack/react-table";
import { Edit } from "lucide-react";
import { QuickActions } from "@/components/ui/quick-actions";
import { Badge } from "@/components/ui/badge";
import type { Provider } from "@/types/provider";

interface ProviderColumnsProps {
  onEdit: (provider: Provider) => void;
  onDelete: (id: number) => Promise<void>;
  t: (key: string) => string;
  canEditProvider: (resourceAdminId?: number) => boolean;
  canDeleteProvider: (resourceAdminId?: number) => boolean;
}

export const createProviderColumns = ({
  onEdit,
  onDelete,
  t,
  canEditProvider,
  canDeleteProvider,
}: ProviderColumnsProps): ColumnDef<Provider>[] => {
  return [
    {
      accessorKey: "name",
      header: t("provider.table.name"),
      cell: ({ row }) => (
        <div className="font-medium">{row.getValue("name")}</div>
      ),
    },
    {
      accessorKey: "description",
      header: t("provider.table.description"),
      cell: ({ row }) => {
        const description = row.getValue("description") as string;
        return (
          <div className="max-w-[300px] truncate text-muted-foreground">
            {description || t("provider.table.no_description")}
          </div>
        );
      },
    },
    {
      accessorKey: "type",
      header: t("provider.table.type"),
      cell: ({ row }) => {
        const type = row.getValue("type") as string;
        return (
          <Badge variant="secondary">
            {t(`provider.types.${type}`)}
          </Badge>
        );
      },
    },
    {
      accessorKey: "created_by",
      header: t("provider.table.created_by"),
      cell: ({ row }) => {
        const createdBy = row.getValue("created_by") as string;
        return (
          <div className="text-sm text-muted-foreground">
            {createdBy}
          </div>
        );
      },
    },
    {
      accessorKey: "created_at",
      header: t("provider.table.created"),
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
        headerClassName:
          "sticky right-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-sm",
        cellClassName:
          "sticky right-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-sm",
      },
      cell: ({ row }) => {
        const provider = row.original;

        const actions = [];

        // Only show edit action if user has permission
        if (canEditProvider(provider.admin_id)) {
          actions.push({
            id: "edit",
            label: t("provider.edit"),
            icon: Edit,
            onClick: () => onEdit(provider),
          });
        }

        // Only show delete action if user has permission
        const deleteAction = canDeleteProvider(provider.admin_id)
          ? {
              title: provider.name,
              confirmationValue: provider.name,
              submitAction: async () => {
                await onDelete(provider.id);
              },
            }
          : undefined;

        return <QuickActions actions={actions} deleteAction={deleteAction} />;
      },
      enableSorting: false,
      enableHiding: false,
    },
  ];
};
