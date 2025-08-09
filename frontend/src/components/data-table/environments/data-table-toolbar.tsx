import type { Table } from "@tanstack/react-table";
import { useTranslation } from "react-i18next";
import { DataTableToolbar } from "@/components/ui/data-table";
import type { DevEnvironmentDisplay } from "@/types/dev-environment";

export interface DevEnvironmentDataTableToolbarProps {
  table: Table<DevEnvironmentDisplay>;
}

export function DevEnvironmentDataTableToolbar({
  table,
}: DevEnvironmentDataTableToolbarProps) {
  const { t } = useTranslation();

  return (
    <DataTableToolbar
      table={table}
      filterColumn="name"
      filterPlaceholder={t("devEnvironments.filter.placeholder", "Filter environments...")}
    />
  );
}
