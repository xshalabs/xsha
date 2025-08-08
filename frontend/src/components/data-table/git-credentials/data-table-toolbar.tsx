import { Table } from "@tanstack/react-table";
import { useTranslation } from "react-i18next";
import { DataTableToolbar } from "@/components/ui/data-table";
import { GitCredential } from "@/types/git-credentials";

export interface GitCredentialDataTableToolbarProps {
  table: Table<GitCredential>;
}

export function GitCredentialDataTableToolbar({
  table,
}: GitCredentialDataTableToolbarProps) {
  const { t } = useTranslation();

  return (
    <DataTableToolbar
      table={table}
      filterColumn="name"
      filterPlaceholder={t("gitCredentials.filter.placeholder", "Filter credentials...")}
    />
  );
}
