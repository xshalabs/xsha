import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DataTable } from "@/components/ui/data-table/data-table";
import { DataTablePaginationSimple } from "@/components/ui/data-table/data-table-pagination";
import { createGitCredentialColumns } from "@/components/data-table/git-credentials/columns";
import { GitCredentialDataTableToolbar } from "@/components/data-table/git-credentials/data-table-toolbar";
import { GitCredentialDataTableActionBar } from "@/components/data-table/git-credentials/data-table-action-bar";
import type { GitCredential, GitCredentialType } from "@/types/git-credentials";
import { GitCredentialType as CredentialTypes } from "@/types/git-credentials";
import { 
  Key, 
  RefreshCw, 
  Shield, 
  User, 
  Clock, 
  MoreHorizontal, 
  Edit, 
  Trash2,
  ChevronLeft,
  ChevronRight 
} from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface GitCredentialListProps {
  credentials: GitCredential[];
  loading: boolean;
  currentPage: number;
  totalPages: number;
  total: number;
  typeFilter?: GitCredentialType;
  onPageChange: (page: number) => void;
  onTypeFilterChange: (type: GitCredentialType | undefined) => void;
  onEdit: (credential: GitCredential) => void;
  onDelete: (id: number) => void;
  onBatchDelete?: (ids: number[]) => void;
  onRefresh: () => void;
}

export const GitCredentialList: React.FC<GitCredentialListProps> = ({
  credentials,
  loading,
  currentPage,
  totalPages,
  total,
  typeFilter,
  onPageChange,
  onTypeFilterChange,
  onEdit,
  onDelete,
  onBatchDelete,
  onRefresh,
}) => {
  const { t } = useTranslation();
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);

  const columns = createGitCredentialColumns({
    onEdit,
    onDelete,
  });

  const ActionBar = ({ table }: { table: any }) => {
    if (!onBatchDelete) return null;
    return <GitCredentialDataTableActionBar table={table} onBatchDelete={onBatchDelete} />;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">{t("common.loading")}</div>
      </div>
    );
  }

  if (credentials.length === 0) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="text-center py-8">
            <Key className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium text-foreground mb-2">
              {t("gitCredentials.messages.noCredentials")}
            </h3>
            <p className="text-muted-foreground mb-4">
              {t("gitCredentials.messages.noCredentialsDesc")}
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <DataTable
      columns={columns}
      data={credentials}
      actionBar={onBatchDelete ? ActionBar : undefined}
      toolbarComponent={GitCredentialDataTableToolbar}
      paginationComponent={DataTablePaginationSimple}
      columnFilters={columnFilters}
      setColumnFilters={setColumnFilters}
      sorting={sorting}
      setSorting={setSorting}
    />
  );
};

export default GitCredentialList;
