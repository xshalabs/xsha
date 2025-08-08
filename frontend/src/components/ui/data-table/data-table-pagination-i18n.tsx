import { Table } from "@tanstack/react-table";
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from "lucide-react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";

export interface DataTablePaginationI18nProps<TData> {
  table: Table<TData>;
}

export function DataTablePaginationI18n<TData>({
  table,
}: DataTablePaginationI18nProps<TData>) {
  const { t } = useTranslation();

  return (
    <div className="flex items-center justify-end px-2">
      <div className="flex items-center space-x-2">
        <Button
          variant="outline"
          className="h-8 w-8 p-0"
          onClick={() => table.setPageIndex(0)}
          disabled={!table.getCanPreviousPage()}
        >
          <span className="sr-only">{t("common.goToFirstPage")}</span>
          <ChevronsLeft />
        </Button>
        <Button
          variant="outline"
          className="h-8 w-8 p-0"
          onClick={() => table.previousPage()}
          disabled={!table.getCanPreviousPage()}
        >
          <span className="sr-only">{t("common.goToPreviousPage")}</span>
          <ChevronLeft />
        </Button>
        <Button
          variant="outline"
          className="h-8 w-8 p-0"
          onClick={() => table.nextPage()}
          disabled={!table.getCanNextPage()}
        >
          <span className="sr-only">{t("common.goToNextPage")}</span>
          <ChevronRight />
        </Button>
        <Button
          variant="outline"
          className="h-8 w-8 p-0"
          onClick={() => table.setPageIndex(table.getPageCount() - 1)}
          disabled={!table.getCanNextPage()}
        >
          <span className="sr-only">{t("common.goToLastPage")}</span>
          <ChevronsRight />
        </Button>
      </div>
    </div>
  );
}
