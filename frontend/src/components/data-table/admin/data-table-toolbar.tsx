import type { Table } from "@tanstack/react-table";
import { X } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";
import { DataTableFacetedFilter } from "@/components/ui/data-table/data-table-faceted-filter";
import type { Admin } from "@/lib/api";

export interface AdminDataTableToolbarProps {
  table: Table<Admin>;
}

export function AdminDataTableToolbar({
  table,
}: AdminDataTableToolbarProps) {
  const { t } = useTranslation();
  const isFiltered = table.getState().columnFilters.length > 0;

  // Status filter options for faceted filter
  const statusOptions = [
    { label: t("admin.filters.active"), value: "active" },
    { label: t("admin.filters.inactive"), value: "inactive" },
  ];

  // Role filter options
  const roleOptions = [
    { label: t("admin.roles.super_admin"), value: "super_admin" },
    { label: t("admin.roles.admin"), value: "admin" },
    { label: t("admin.roles.developer"), value: "developer" },
  ];

  return (
    <div className="flex flex-col space-y-4">
      <div className="flex flex-1 items-center space-x-2 flex-wrap gap-2">
        {table.getColumn("role") && (
          <DataTableFacetedFilter
            column={table.getColumn("role")}
            title={t("admin.filters.role")}
            options={roleOptions}
          />
        )}
        {table.getColumn("is_active") && (
          <DataTableFacetedFilter
            column={table.getColumn("is_active")}
            title={t("admin.filters.status")}
            options={statusOptions}
          />
        )}
        {isFiltered && (
          <Button
            variant="ghost"
            onClick={() => table.resetColumnFilters()}
            className="h-8 px-2 lg:px-3"
          >
            {t("common.reset")}
            <X className="ml-2 h-4 w-4" />
          </Button>
        )}
      </div>
    </div>
  );
}