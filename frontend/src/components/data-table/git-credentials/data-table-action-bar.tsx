import type { Table } from "@tanstack/react-table";
import { Trash2 } from "lucide-react";
import * as React from "react";

import {
  DataTableActionBar,
  DataTableActionBarAction,
  DataTableActionBarSelection,
} from "@/components/ui/data-table";
import { Separator } from "@/components/ui/separator";
import { GitCredential } from "@/types/git-credentials";

interface GitCredentialDataTableActionBarProps {
  table: Table<GitCredential>;
  onDelete?: (ids: number[]) => void;
}

export function GitCredentialDataTableActionBar({
  table,
  onDelete,
}: GitCredentialDataTableActionBarProps) {
  const rows = table.getFilteredSelectedRowModel().rows;

  const handleDelete = () => {
    if (onDelete) {
      const ids = rows.map((row) => row.original.id);
      onDelete(ids);
    }
  };

  return (
    <DataTableActionBar table={table} visible={rows.length > 0}>
      <DataTableActionBarSelection table={table} />
      <Separator
        orientation="vertical"
        className="hidden data-[orientation=vertical]:h-5 sm:block"
      />
      <div className="flex items-center gap-1.5">
        <DataTableActionBarAction size="icon" tooltip="Delete credentials" onClick={handleDelete}>
          <Trash2 />
        </DataTableActionBarAction>
      </div>
    </DataTableActionBar>
  );
}
