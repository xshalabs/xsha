import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { usePageTitle } from "@/hooks/usePageTitle";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";
import { usePageActions } from "@/contexts/PageActionsContext";
import { Button } from "@/components/ui/button";


import {
  Section,
  SectionGroup,
  SectionHeader,
  SectionTitle,
  SectionDescription,
} from "@/components/content/section";
import { DataTable } from "@/components/ui/data-table/data-table";
import { DataTablePaginationServer } from "@/components/ui/data-table/data-table-pagination-server";
import { createGitCredentialColumns } from "@/components/data-table/credentials/columns";
import { GitCredentialDataTableToolbar } from "@/components/data-table/credentials/data-table-toolbar";
import { GitCredentialDataTableActionBar } from "@/components/data-table/credentials/data-table-action-bar";
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import type {
  GitCredential,
  GitCredentialListParams,
} from "@/types/credentials";
import { GitCredentialType } from "@/types/credentials";
import { Plus } from "lucide-react";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";

const CredentialListPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { setItems } = useBreadcrumb();
  const { setActions } = usePageActions();

  const [credentials, setCredentials] = useState<GitCredential[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);



  const pageSize = 10;

  usePageTitle(t("common.pageTitle.gitCredentials"));

  // Set page actions (Create button in header) and clear breadcrumb
  useEffect(() => {
    const handleCreateNew = () => {
      navigate("/credentials/create");
    };

    setActions(
      <Button onClick={handleCreateNew} size="sm">
        <Plus className="h-4 w-4 mr-2" />
        {t("gitCredentials.create")}
      </Button>
    );

    // Clear breadcrumb items (we're at root level)
    setItems([]);

    // Cleanup when component unmounts
    return () => {
      setActions(null);
      setItems([]);
    };
  }, [navigate, setActions, setItems, t]);

  const loadCredentials = async (page = currentPage, filters = columnFilters) => {
    try {
      setLoading(true);
      
      // Convert DataTable filters to API parameters
      const apiParams: GitCredentialListParams = {
        page,
        page_size: pageSize,
      };

      // Handle column filters
      filters.forEach((filter) => {
        if (filter.id === "type" && Array.isArray(filter.value) && filter.value.length > 0) {
          apiParams.type = filter.value[0] as GitCredentialType;
        }
      });

      const response = await apiService.gitCredentials.list(apiParams);

      setCredentials(response.credentials);
      setTotal(response.total);
      setTotalPages(response.total_pages);
      setCurrentPage(page);
    } catch (err: any) {
      const errorMessage =
        err.message || t("gitCredentials.messages.loadFailed");
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCredentials().then(() => setIsInitialized(true));
  }, []);

  // Handle column filter changes (skip initial empty state)
  const [isInitialized, setIsInitialized] = useState(false);
  
  useEffect(() => {
    if (isInitialized) {
      loadCredentials(1, columnFilters); // Reset to page 1 when filtering
    }
  }, [columnFilters, isInitialized]);

  const handlePageChange = (page: number) => {
    loadCredentials(page);
  };

  const handleEdit = (credential: GitCredential) => {
    navigate(`/credentials/${credential.id}/edit`);
  };

  const handleDelete = async (id: number) => {
    try {
      await apiService.gitCredentials.delete(id);
      await loadCredentials();
    } catch (error) {
      // Re-throw error to let QuickActions handle the user notification
      throw error;
    }
  };



  const handleBatchDelete = async (ids: number[]) => {
    try {
      await Promise.all(ids.map((id) => apiService.gitCredentials.delete(id)));
      toast.success(
        t(
          "gitCredentials.messages.batchDeleteSuccess",
          `Successfully deleted ${ids.length} credentials`
        )
      );
      await loadCredentials();
    } catch (err: any) {
      const errorMessage =
        err.message ||
        t(
          "gitCredentials.messages.batchDeleteFailed",
          "Failed to delete credentials"
        );
      toast.error(errorMessage);
    }
  };



  return (
    <>
      <SectionGroup>
        <Section>
          <SectionHeader>
            <SectionTitle>{t("gitCredentials.title")}</SectionTitle>
            <SectionDescription>
              {t(
                "gitCredentials.subtitle",
                "Manage your Git repository access credentials"
              )}
            </SectionDescription>
          </SectionHeader>


        </Section>

        <Section>
          {loading ? (
            <div className="flex items-center justify-center h-64">
              <div className="text-gray-500">{t("common.loading")}</div>
            </div>
          ) : (
            <div className="space-y-4">
              <DataTable
                columns={createGitCredentialColumns({
                  onEdit: handleEdit,
                  onDelete: handleDelete,
                })}
                data={credentials}
                actionBar={({ table }: { table: any }) => (
                  <GitCredentialDataTableActionBar
                    table={table}
                    onBatchDelete={handleBatchDelete}
                  />
                )}
                toolbarComponent={GitCredentialDataTableToolbar}
                columnFilters={columnFilters}
                setColumnFilters={setColumnFilters}
                sorting={sorting}
                setSorting={setSorting}
              />
              <DataTablePaginationServer
                currentPage={currentPage}
                totalPages={totalPages}
                total={total}
                onPageChange={handlePageChange}
              />
            </div>
          )}
        </Section>
      </SectionGroup>
    </>
  );
};

export default CredentialListPage;
