import React, { useState, useEffect, useMemo } from "react";
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
import {
  MetricCardGroup,
  MetricCardHeader,
  MetricCardTitle,
  MetricCardValue,
  MetricCardButton,
} from "@/components/metric/metric-card";
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
import { Plus, Key, Shield, ListFilter, CheckCircle } from "lucide-react";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";

const GitCredentialListPage: React.FC = () => {
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

  // Calculate statistics
  const statistics = useMemo(() => {
    const passwordCount = credentials.filter(
      (cred) => cred.type === GitCredentialType.PASSWORD
    ).length;
    const tokenCount = credentials.filter(
      (cred) => cred.type === GitCredentialType.TOKEN
    ).length;

    return [
      {
        title: t("gitCredentials.filter.password"),
        value: passwordCount,
        variant: "success" as const,
        type: GitCredentialType.PASSWORD,
        icon: Key,
      },
      {
        title: t("gitCredentials.filter.token"),
        value: tokenCount,
        variant: "warning" as const,
        type: GitCredentialType.TOKEN,
        icon: Shield,
      },
      {
        title: t("common.total"),
        value: total,
        variant: "ghost" as const,
        type: undefined,
        icon: ListFilter,
      },
    ];
  }, [credentials, total, t]);

  const handleStatisticClick = (
    statisticType: GitCredentialType | undefined
  ) => {
    let newColumnFilters = [...columnFilters];
    
    if (statisticType === undefined) {
      // Clear all filters
      newColumnFilters = columnFilters.filter(f => f.id !== "type");
    } else {
      // Toggle filter
      const currentFilter = columnFilters.find(f => f.id === "type");
      const currentValues = (currentFilter?.value as string[]) || [];
      const isActive = currentValues.includes(statisticType);
      
      newColumnFilters = columnFilters.filter(f => f.id !== "type");
      if (!isActive) {
        newColumnFilters.push({ id: "type", value: [statisticType] });
      }
    }
    
    setColumnFilters(newColumnFilters);
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

          <MetricCardGroup>
            {statistics.map((stat) => {
              const currentFilter = columnFilters.find(f => f.id === "type");
              const currentValues = (currentFilter?.value as string[]) || [];
              const isActive = stat.type === undefined 
                ? currentValues.length === 0 
                : currentValues.includes(stat.type);

              // Determine icon based on state (like openstatus)
              let Icon;
              if (stat.type === undefined) {
                // Total always uses ListFilter
                Icon = ListFilter;
              } else {
                // Filter types use CheckCircle when active, type icon when inactive
                Icon = isActive ? CheckCircle : stat.icon;
              }

              return (
                <MetricCardButton
                  key={stat.title}
                  variant={stat.variant}
                  onClick={() => handleStatisticClick(stat.type)}
                >
                  <MetricCardHeader className="flex justify-between items-center gap-2 w-full">
                    <MetricCardTitle className="truncate">
                      {stat.title}
                    </MetricCardTitle>
                    <Icon className="size-4" />
                  </MetricCardHeader>
                  <MetricCardValue>{stat.value}</MetricCardValue>
                </MetricCardButton>
              );
            })}
          </MetricCardGroup>
        </Section>

        <Section>
          {loading ? (
            <div className="flex items-center justify-center h-64">
              <div className="text-gray-500">{t("common.loading")}</div>
            </div>
          ) : credentials.length === 0 ? (
            <div className="text-center py-8">
              <Key className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
              <h3 className="text-lg font-medium text-foreground mb-2">
                {t("gitCredentials.messages.noCredentials")}
              </h3>
              <p className="text-muted-foreground mb-4">
                {t("gitCredentials.messages.noCredentialsDesc")}
              </p>
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

export default GitCredentialListPage;
