import type { Table } from "@tanstack/react-table";
import { X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useState, useEffect, useRef } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DataTableFacetedFilter } from "@/components/ui/data-table/data-table-faceted-filter";
import type { Task } from "@/types/task";
import type { DevEnvironment } from "@/types/dev-environment";

export interface TaskDataTableToolbarProps {
  table: Table<Task>;
  devEnvironments?: DevEnvironment[];
}

export function TaskDataTableToolbar({
  table,
  devEnvironments = [],
}: TaskDataTableToolbarProps) {
  const { t } = useTranslation();
  const isFiltered = table.getState().columnFilters.length > 0;

  // Local state for the input values with debouncing
  const [titleSearchValue, setTitleSearchValue] = useState("");
  const [branchSearchValue, setBranchSearchValue] = useState("");
  const isTitleUserInput = useRef(false);
  const isBranchUserInput = useRef(false);

  // Sync search values with table filter values
  useEffect(() => {
    const currentTitleFilter = table.getColumn("title")?.getFilterValue() as string | undefined;
    
    // Only update if it's not from user input to avoid conflicts
    if (!isTitleUserInput.current) {
      setTitleSearchValue(currentTitleFilter || "");
    }
    
    // Reset the flag after processing
    isTitleUserInput.current = false;
  }, [table.getState().columnFilters]);

  useEffect(() => {
    const currentBranchFilter = table.getColumn("start_branch")?.getFilterValue() as string | undefined;
    
    // Only update if it's not from user input to avoid conflicts
    if (!isBranchUserInput.current) {
      setBranchSearchValue(currentBranchFilter || "");
    }
    
    // Reset the flag after processing
    isBranchUserInput.current = false;
  }, [table.getState().columnFilters]);

  // Debounce the filter value update when user types for title
  useEffect(() => {
    if (!isTitleUserInput.current) {
      return;
    }
    
    const debounceTimer = setTimeout(() => {
      if (isTitleUserInput.current) {
        table.getColumn("title")?.setFilterValue(titleSearchValue || undefined);
        isTitleUserInput.current = false;
      }
    }, 500); // 500ms delay

    return () => {
      clearTimeout(debounceTimer);
    };
  }, [titleSearchValue, table]);

  // Debounce the filter value update when user types for branch
  useEffect(() => {
    if (!isBranchUserInput.current) {
      return;
    }
    
    const debounceTimer = setTimeout(() => {
      if (isBranchUserInput.current) {
        table.getColumn("start_branch")?.setFilterValue(branchSearchValue || undefined);
        isBranchUserInput.current = false;
      }
    }, 500); // 500ms delay

    return () => {
      clearTimeout(debounceTimer);
    };
  }, [branchSearchValue, table]);

  const handleTitleSearchChange = (value: string) => {
    isTitleUserInput.current = true;
    setTitleSearchValue(value);
  };

  const handleBranchSearchChange = (value: string) => {
    isBranchUserInput.current = true;
    setBranchSearchValue(value);
  };

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-1 items-center space-x-2 flex-wrap">
        <Input
          placeholder={t("tasks.filters.titlePlaceholder")}
          value={titleSearchValue}
          onChange={(event) => handleTitleSearchChange(event.target.value)}
          className="h-8 w-[150px] lg:w-[250px]"
        />

        {table.getColumn("start_branch") && (
          <Input
            placeholder={t("tasks.filters.branchPlaceholder")}
            value={branchSearchValue}
            onChange={(event) => handleBranchSearchChange(event.target.value)}
            className="h-8 w-[150px] lg:w-[200px]"
          />
        )}

        {table.getColumn("status") && (
          <DataTableFacetedFilter
            column={table.getColumn("status")}
            title={t("tasks.filters.status")}
            options={[
              { label: t("tasks.status.todo"), value: "todo" },
              { label: t("tasks.status.in_progress"), value: "in_progress" },
              { label: t("tasks.status.done"), value: "done" },
              { label: t("tasks.status.cancelled"), value: "cancelled" },
            ]}
          />
        )}

        {table.getColumn("dev_environment.name") && devEnvironments.length > 0 && (
          <DataTableFacetedFilter
            column={table.getColumn("dev_environment.name")}
            title={t("tasks.filters.devEnvironment")}
            options={devEnvironments.map((env) => ({
              label: `${env.name} (${env.type})`,
              value: env.id.toString(),
            }))}
          />
        )}

        {isFiltered && (
          <Button
            variant="ghost"
            onClick={() => {
              table.resetColumnFilters();
              // Search values will be synced automatically via useEffect
            }}
            className="h-8 px-2 lg:px-3"
          >
            {t("common.reset")}
            <X />
          </Button>
        )}
      </div>
    </div>
  );
}
