import type { Table } from "@tanstack/react-table";
import { X, Search } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useState, useEffect, useRef } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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

  // Local state for the input value
  const [searchValue, setSearchValue] = useState("");
  const isUserInput = useRef(false);

  // Sync searchValue with table filter value
  useEffect(() => {
    const currentFilter = table.getColumn("username")?.getFilterValue() as
      | string
      | undefined;

    // Only update if it's not from user input to avoid conflicts
    if (!isUserInput.current) {
      setSearchValue(currentFilter || "");
    }

    // Reset the flag after processing
    isUserInput.current = false;
  }, [table.getState().columnFilters]);

  // Debounce the filter value update when user types
  useEffect(() => {
    if (!isUserInput.current) {
      return;
    }

    const debounceTimer = setTimeout(() => {
      if (isUserInput.current) {
        // Double check before setting filter
        table.getColumn("username")?.setFilterValue(searchValue || undefined);
        isUserInput.current = false;
      }
    }, 500); // 500ms debounce for consistency with operation logs

    return () => clearTimeout(debounceTimer);
  }, [searchValue, table]);

  const handleInputChange = (value: string) => {
    setSearchValue(value);
    isUserInput.current = true;
  };

  // Status filter options for faceted filter
  const statusOptions = [
    { label: t("admin.filters.active"), value: "active" },
    { label: t("admin.filters.inactive"), value: "inactive" },
  ];

  return (
    <div className="flex flex-col space-y-4">
      <div className="flex flex-1 items-center space-x-2 flex-wrap gap-2">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder={t("admin.filters.searchUsername")}
            value={searchValue}
            onChange={(event) => handleInputChange(event.target.value)}
            className="h-8 w-[150px] lg:w-[250px] pl-10"
          />
        </div>
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