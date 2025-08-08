import { Table } from "@tanstack/react-table";
import { useTranslation } from "react-i18next";
import { DataTableToolbar } from "@/components/ui/data-table";
import { Project } from "@/types/project";

export interface ProjectDataTableToolbarProps {
  table: Table<Project>;
}

export function ProjectDataTableToolbar({
  table,
}: ProjectDataTableToolbarProps) {
  const { t } = useTranslation();

  return (
    <DataTableToolbar
      table={table}
      filterColumn="name"
      filterPlaceholder={t("projects.filter.placeholder", "Filter projects...")}
    />
  );
}
