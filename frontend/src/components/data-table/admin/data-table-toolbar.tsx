import type { Table } from "@tanstack/react-table";
import { X, Search } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useState, useEffect, useRef } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
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
    }, 300); // 300ms debounce

    return () => clearTimeout(debounceTimer);
  }, [searchValue, table]);

  const handleInputChange = (value: string) => {
    setSearchValue(value);
    isUserInput.current = true;
  };

  const handleStatusFilterChange = (value: string) => {
    const column = table.getColumn("is_active");
    if (value === "all") {
      column?.setFilterValue(undefined);
    } else {
      column?.setFilterValue(value);
    }
  };

  const getCurrentStatusFilter = () => {
    const value = table.getColumn("is_active")?.getFilterValue();
    if (value === undefined || value === "all") return "all";
    return value === true ? "active" : "inactive";
  };

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-1 items-center space-x-2">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder={t("admin.filters.searchUsername")}
            value={searchValue}
            onChange={(event) => handleInputChange(event.target.value)}
            className="h-8 w-[150px] lg:w-[250px] pl-10"
          />
        </div>
        <Select 
          value={getCurrentStatusFilter()} 
          onValueChange={handleStatusFilterChange}
        >
          <SelectTrigger className="h-8 w-[150px]">
            <SelectValue placeholder={t("admin.filters.status")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t("admin.filters.allStatus")}</SelectItem>
            <SelectItem value="active">{t("admin.filters.active")}</SelectItem>
            <SelectItem value="inactive">{t("admin.filters.inactive")}</SelectItem>
          </SelectContent>
        </Select>
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