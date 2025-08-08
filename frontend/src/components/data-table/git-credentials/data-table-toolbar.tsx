import { Table } from "@tanstack/react-table";
import { useTranslation } from "react-i18next";
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DataTableFacetedFilter } from "@/components/ui/data-table/data-table-faceted-filter";
import { GitCredential, GitCredentialType } from "@/types/git-credentials";

export interface GitCredentialDataTableToolbarProps {
  table: Table<GitCredential>;
}

export function GitCredentialDataTableToolbar({
  table,
}: GitCredentialDataTableToolbarProps) {
  const { t } = useTranslation();
  const isFiltered = table.getState().columnFilters.length > 0;

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-1 items-center space-x-2 flex-wrap">
        <Input
          placeholder={t("gitCredentials.filter.placeholder", "Filter name...")}
          value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
          onChange={(event) =>
            table.getColumn("name")?.setFilterValue(event.target.value)
          }
          className="h-8 w-[150px] lg:w-[250px]"
        />
        {table.getColumn("type") && (
          <DataTableFacetedFilter
            column={table.getColumn("type")}
            title={t("gitCredentials.type")}
            options={[
              { label: t("gitCredentials.filter.password"), value: GitCredentialType.PASSWORD },
              { label: t("gitCredentials.filter.token"), value: GitCredentialType.TOKEN },
              { label: t("gitCredentials.filter.sshKey"), value: GitCredentialType.SSH_KEY },
            ]}
          />
        )}
        {isFiltered && (
          <Button
            variant="ghost"
            onClick={() => table.resetColumnFilters()}
            className="h-8 px-2 lg:px-3"
          >
            {t("common.reset", "Reset")}
            <X />
          </Button>
        )}
      </div>
    </div>
  );
}
