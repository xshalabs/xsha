import type { Table } from "@tanstack/react-table";
import { useTranslation } from "react-i18next";
import { useState, useEffect, useRef } from "react";
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Provider, ProviderType } from "@/types/provider";

export interface ProviderDataTableToolbarProps {
  table: Table<Provider>;
  providerTypes?: ProviderType[];
}

export function ProviderDataTableToolbar({
  table,
  providerTypes = ["claude-code"],
}: ProviderDataTableToolbarProps) {
  const { t } = useTranslation();
  const isFiltered = table.getState().columnFilters.length > 0;

  // Local state for the input value
  const [searchValue, setSearchValue] = useState("");
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const isUserInput = useRef(false);

  // Sync searchValue with table filter value
  useEffect(() => {
    const currentFilter = table.getColumn("name")?.getFilterValue() as
      | string
      | undefined;

    // Only update if it's not from user input to avoid conflicts
    if (!isUserInput.current) {
      setSearchValue(currentFilter || "");
    }

    // Reset the flag after processing
    isUserInput.current = false;
  }, [table.getState().columnFilters]);

  // Sync typeFilter with table filter value
  useEffect(() => {
    const currentTypeFilter = table.getColumn("type")?.getFilterValue() as
      | string
      | undefined;

    // Update local typeFilter state from table filter
    setTypeFilter(currentTypeFilter || "all");
  }, [table.getState().columnFilters]);

  // Debounce the filter value update when user types
  useEffect(() => {
    if (!isUserInput.current) {
      return;
    }

    const debounceTimer = setTimeout(() => {
      if (isUserInput.current) {
        // Double check before setting filter
        table.getColumn("name")?.setFilterValue(searchValue || undefined);
        isUserInput.current = false;
      }
    }, 500); // 500ms delay

    return () => {
      clearTimeout(debounceTimer);
    };
  }, [searchValue, table]);

  const handleSearchChange = (value: string) => {
    isUserInput.current = true;
    setSearchValue(value);
  };

  const handleTypeFilterChange = (value: string) => {
    setTypeFilter(value);
    if (value === "all") {
      table.getColumn("type")?.setFilterValue(undefined);
    } else {
      table.getColumn("type")?.setFilterValue(value);
    }
  };

  const handleResetFilters = () => {
    table.resetColumnFilters();
    setTypeFilter("all");
    // searchValue will be synced automatically via useEffect
  };

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-1 items-center space-x-2 flex-wrap">
        <Input
          placeholder={t("provider.filters.name_placeholder")}
          value={searchValue}
          onChange={(event) => handleSearchChange(event.target.value)}
          className="h-8 w-[150px] lg:w-[250px]"
        />

        <Select value={typeFilter} onValueChange={handleTypeFilterChange}>
          <SelectTrigger className="h-8 w-[150px]">
            <SelectValue placeholder={t("provider.filters.type_placeholder")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t("provider.filters.all_types")}</SelectItem>
            {providerTypes.map((type) => (
              <SelectItem key={type} value={type}>
                {t(`provider.types.${type}`)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {isFiltered && (
          <Button
            variant="ghost"
            onClick={handleResetFilters}
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
