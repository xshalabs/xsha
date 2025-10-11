import type { Table } from "@tanstack/react-table";
import { X, ToggleLeft, ToggleRight } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useState, useEffect, useRef } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DataTableFacetedFilter } from "@/components/ui/data-table/data-table-faceted-filter";
import type { MCP } from "@/types/mcp";

export interface MCPDataTableToolbarProps {
  table: Table<MCP>;
}

export function MCPDataTableToolbar({
  table,
}: MCPDataTableToolbarProps) {
  const { t } = useTranslation();
  const isFiltered = table.getState().columnFilters.length > 0;

  // Local state for the search input value
  const [searchValue, setSearchValue] = useState("");
  const isUserInput = useRef(false);

  // Sync searchValue with table filter value
  useEffect(() => {
    const currentFilter = table.getColumn("search")?.getFilterValue() as
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
        table.getColumn("search")?.setFilterValue(searchValue || undefined);
        isUserInput.current = false;
      }
    }, 500); // 500ms delay

    return () => {
      clearTimeout(debounceTimer);
      // Don't reset isUserInput here as it might interfere with the timer
    };
  }, [searchValue, table]);

  const handleSearchChange = (value: string) => {
    isUserInput.current = true;
    setSearchValue(value);
  };

  // Status filter options
  const statusOptions = [
    {
      label: t("mcp.status.enabled"),
      value: "enabled",
      icon: () => <ToggleRight className="w-4 h-4" />
    },
    {
      label: t("mcp.status.disabled"),
      value: "disabled",
      icon: () => <ToggleLeft className="w-4 h-4" />
    },
  ];

  return (
    <div className="flex flex-col space-y-4">
      <div className="flex flex-1 items-center space-x-2 flex-wrap gap-2">
        <Input
          placeholder={t("mcp.search.placeholder")}
          value={searchValue}
          onChange={(event) => handleSearchChange(event.target.value)}
          className="h-8 w-[150px] lg:w-[250px]"
        />
        {table.getColumn("enabled") && (
          <DataTableFacetedFilter
            column={table.getColumn("enabled")}
            title={t("mcp.filters.status.placeholder")}
            options={statusOptions}
          />
        )}
        {isFiltered && (
          <Button
            variant="ghost"
            onClick={() => {
              table.resetColumnFilters();
              // searchValue will be synced automatically via useEffect
            }}
            className="h-8 px-2 lg:px-3"
          >
            {t("mcp.filters.reset")}
            <X className="ml-2 h-4 w-4" />
          </Button>
        )}
      </div>
    </div>
  );
}