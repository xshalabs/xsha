import React, {
  useState,
  useEffect,
  useMemo,
  useCallback,
  useRef,
} from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";
import { usePageActions } from "@/contexts/PageActionsContext";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

import { Plus, Save } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { usePermissions } from "@/hooks/usePermissions";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";

import {
  Section,
  SectionDescription,
  SectionGroup,
  SectionHeader,
  SectionTitle,
} from "@/components/content/section";
import {
  FormSheet,
  FormSheetContent,
  FormSheetHeader,
  FormSheetTitle,
  FormSheetDescription,
  FormSheetFooter,
  FormCardGroup,
} from "@/components/forms/form-sheet";
import { FormCard, FormCardContent } from "@/components/forms/form-card";
import { EnvironmentFormSheet } from "@/components/EnvironmentFormSheet";
import { DataTable } from "@/components/ui/data-table/data-table";
import { DataTablePaginationServer } from "@/components/ui/data-table/data-table-pagination-server";
import { createDevEnvironmentColumns } from "@/components/data-table/environments/columns";
import { DevEnvironmentDataTableToolbar } from "@/components/data-table/environments/data-table-toolbar";
import type {
  DevEnvironment,
  DevEnvironmentDisplay,
  DevEnvironmentListParams,
} from "@/types/dev-environment";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";

const EnvironmentListPage: React.FC = () => {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { setItems } = useBreadcrumb();
  const { setActions } = usePageActions();
  const { canCreateEnvironment, canEditEnvironment, canDeleteEnvironment, canManageEnvironmentAdmins, adminId } = usePermissions();
  
  usePageTitle(t("navigation.environments"));

  // Check for action parameter to auto-open create sheet
  useEffect(() => {
    const actionParam = searchParams.get("action");
    if (actionParam === "create") {
      setIsCreateSheetOpen(true);
      // Remove action parameter from URL to keep it clean
      const newSearchParams = new URLSearchParams(searchParams);
      newSearchParams.delete("action");
      setSearchParams(newSearchParams, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  // Set page actions (Create button in header) and clear breadcrumb
  useEffect(() => {
    const handleCreateNew = () => {
      setIsCreateSheetOpen(true);
    };

    // Only show create button if user has permission
    if (canCreateEnvironment) {
      setActions(
        <Button onClick={handleCreateNew} size="sm">
          <Plus className="h-4 w-4 mr-2" />
          {t("devEnvironments.create")}
        </Button>
      );
    } else {
      setActions(null);
    }

    // Clear breadcrumb items (we're at the root level)
    setItems([]);

    // Cleanup when component unmounts
    return () => {
      setActions(null);
      setItems([]);
    };
  }, [setActions, setItems, t, canCreateEnvironment]);

  const [environments, setEnvironments] = useState<DevEnvironment[]>([]);
  const [loading, setLoading] = useState(true);

  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);

  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);

  // Sheet state management
  const [isCreateSheetOpen, setIsCreateSheetOpen] = useState(false);
  const [isEditSheetOpen, setIsEditSheetOpen] = useState(false);
  const [editingEnvironment, setEditingEnvironment] = useState<DevEnvironmentDisplay | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Add request deduplication
  const lastRequestRef = useRef<string>("");
  const isRequestInProgress = useRef(false);
  
  const pageSize = 10;


  // Separate URL update function to avoid dependency issues
  const updateUrl = useCallback((page: number, filters: ColumnFiltersState) => {
    const params = new URLSearchParams();

    // Add filter parameters
    filters.forEach((filter) => {
      if (filter.value) {
        params.set(filter.id, String(filter.value));
      }
    });

    // Add page parameter (only if not page 1)
    if (page > 1) {
      params.set("page", String(page));
    }

    // Update URL without causing navigation
    setSearchParams(params, { replace: true });
  }, [setSearchParams]);

  const loadEnvironmentsData = useCallback(
    async (page: number, filters: ColumnFiltersState, shouldUpdateUrl = true) => {
      // Create a unique request key for deduplication
      const requestKey = JSON.stringify({ page, filters, shouldUpdateUrl });

      // Skip if same request is already in progress or just completed
      if (
        isRequestInProgress.current ||
        lastRequestRef.current === requestKey
      ) {
        return;
      }

      isRequestInProgress.current = true;
      lastRequestRef.current = requestKey;

      try {
        setLoading(true);

        // Convert DataTable filters to API parameters
        const apiParams: DevEnvironmentListParams = {
          page,
          page_size: pageSize,
        };

        // Handle column filters
        filters.forEach((filter) => {
          if (filter.id === "name" && filter.value) {
            apiParams.name = filter.value as string;
          } else if (filter.id === "docker_image" && filter.value) {
            apiParams.docker_image = filter.value as string;
          }
        });

        const response = await apiService.devEnvironments.list(apiParams);
        setEnvironments(response.environments);
        setTotalPages(response.total_pages);
        setTotal(response.total);
        setCurrentPage(page);

        // Update URL parameters after successful data load
        if (shouldUpdateUrl) {
          updateUrl(page, filters);
        }
      } catch (error) {
        logError(error as Error, "Failed to fetch environments");
      } finally {
        setLoading(false);
        isRequestInProgress.current = false;

        // Clear the request key after a short delay to allow legitimate new requests
        setTimeout(() => {
          if (lastRequestRef.current === requestKey) {
            lastRequestRef.current = "";
          }
        }, 500); // Increase delay to prevent rapid duplicate requests
      }
    },
    [pageSize, updateUrl]
  );



  const handleDeleteEnvironment = useCallback(
    async (id: number) => {
      try {
        await apiService.devEnvironments.delete(id);
        await loadEnvironmentsData(currentPage, columnFilters);
      } catch (error) {
        // Re-throw error to let QuickActions handle the user notification
        throw error;
      }
    },
    [loadEnvironmentsData, currentPage, columnFilters]
  );





  const handleEditEnvironment = useCallback(
    async (environment: DevEnvironment) => {
      try {
        setIsSubmitting(true);
        // Fetch full environment details including env_vars
        const response = await apiService.devEnvironments.get(environment.id);
        
        // Transform the environment to include env_vars_map for the form
        let envVarsMap: Record<string, string> = {};
        try {
          if (response.environment.env_vars) {
            envVarsMap = JSON.parse(response.environment.env_vars);
          }
        } catch (error) {
          console.warn("Failed to parse env_vars for environment:", environment.id, error);
        }

        const environmentWithEnvVars: DevEnvironmentDisplay = {
          ...response.environment,
          env_vars_map: envVarsMap,
        };

        setEditingEnvironment(environmentWithEnvVars);
        setIsEditSheetOpen(true);
      } catch (error) {
        console.error("Failed to fetch environment details:", error);
        toast.error(t("devEnvironments.fetch_details_failed"));
      } finally {
        setIsSubmitting(false);
      }
    },
    [apiService.devEnvironments, t]
  );

  // Sheet handlers
  const handleCreateEnvironment = async (environment: DevEnvironmentDisplay) => {
    try {
      setIsSubmitting(true);
      // Refresh the environment list
      await loadEnvironmentsData(currentPage, columnFilters);
      // Close the sheet
      setIsCreateSheetOpen(false);
      // Show success message
      toast.success(t("devEnvironments.create_success"));
      console.log("Environment created successfully:", environment);
    } catch (error) {
      console.error("Failed to create environment:", error);
      logError(error as Error, "Failed to create environment");
      throw error;
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleUpdateEnvironment = async (environment: DevEnvironmentDisplay) => {
    try {
      setIsSubmitting(true);
      // Refresh the environment list
      await loadEnvironmentsData(currentPage, columnFilters);
      // Close the sheet
      setIsEditSheetOpen(false);
      setEditingEnvironment(null);
      // Show success message
      toast.success(t("devEnvironments.update_success"));
      console.log("Environment updated successfully:", environment);
    } catch (error) {
      console.error("Failed to update environment:", error);
      logError(error as Error, "Failed to update environment");
      throw error;
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCloseCreateSheet = () => {
    setIsCreateSheetOpen(false);
  };

  const handleCloseEditSheet = () => {
    setIsEditSheetOpen(false);
    setEditingEnvironment(null);
  };

  const handleAdminChanged = useCallback(() => {
    // Refresh the environments list when admin assignments change
    loadEnvironmentsData(currentPage, columnFilters, false);
  }, [currentPage, columnFilters, loadEnvironmentsData]);

  const columns = useMemo(
    () =>
      createDevEnvironmentColumns({
        onEdit: handleEditEnvironment,
        onDelete: handleDeleteEnvironment,
        onAdminChanged: handleAdminChanged,
        t,
        canEditEnvironment,
        canDeleteEnvironment,
        canManageEnvironmentAdmins,
        currentAdminId: adminId || undefined,
      }),
    [handleEditEnvironment, handleDeleteEnvironment, handleAdminChanged, t, canEditEnvironment, canDeleteEnvironment, canManageEnvironmentAdmins, adminId]
  );

  // Initialize from URL on component mount (only once)
  const [isInitialized, setIsInitialized] = useState(false);
  
  // Keep track of previous filter values to detect actual changes
  const previousFiltersRef = useRef<ColumnFiltersState>([]);

  useEffect(() => {
    // Get URL params directly to avoid dependency issues
    const nameParam = searchParams.get("name");
    const dockerImageParam = searchParams.get("docker_image");
    const pageParam = searchParams.get("page");

    const initialFilters: ColumnFiltersState = [];

    if (nameParam) {
      initialFilters.push({ id: "name", value: nameParam });
    }

    if (dockerImageParam) {
      initialFilters.push({ id: "docker_image", value: dockerImageParam });
    }

    const initialPage = pageParam ? parseInt(pageParam, 10) : 1;

    // Set state first
    setColumnFilters(initialFilters);
    setCurrentPage(initialPage);
    previousFiltersRef.current = initialFilters;

    // Load initial data using the unified function
    loadEnvironmentsData(initialPage, initialFilters, false).then(() => {
      setIsInitialized(true);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Empty dependency array - only run once on mount

  // Handle column filter changes (skip initial load)
  useEffect(() => {
    if (!isInitialized) return;

    // Check if the filter values actually changed
    const filtersChanged = JSON.stringify(previousFiltersRef.current) !== JSON.stringify(columnFilters);
    
    if (filtersChanged) {
      previousFiltersRef.current = columnFilters;
      loadEnvironmentsData(1, columnFilters); // Reset to page 1 when filtering
    }
  }, [columnFilters, isInitialized, loadEnvironmentsData]);

  const handlePageChange = useCallback(
    (page: number) => {
      loadEnvironmentsData(page, columnFilters);
    },
    [columnFilters, loadEnvironmentsData]
  );

  return (
    <div className="min-h-screen bg-background">
      <SectionGroup>
        <Section>
          <SectionHeader>
            <SectionTitle>{t("navigation.environments")}</SectionTitle>
            <SectionDescription>
              {t("devEnvironments.page_description")}
            </SectionDescription>
          </SectionHeader>
        </Section>
        <Section>
          <div className="space-y-4">
            <DataTable
              columns={columns}
              data={environments}
              toolbarComponent={DevEnvironmentDataTableToolbar}
              columnFilters={columnFilters}
              setColumnFilters={setColumnFilters}
              sorting={sorting}
              setSorting={setSorting}
              loading={loading}
            />
            <DataTablePaginationServer
              currentPage={currentPage}
              totalPages={totalPages}
              total={total}
              onPageChange={handlePageChange}
            />
          </div>
        </Section>
      </SectionGroup>

      {/* Create Environment Sheet */}
      <FormSheet
        open={isCreateSheetOpen}
        onOpenChange={setIsCreateSheetOpen}
      >
        <FormSheetContent className="w-full sm:w-[600px] sm:max-w-[600px]">
          <FormSheetHeader>
            <FormSheetTitle>{t("devEnvironments.create")}</FormSheetTitle>
            <FormSheetDescription>
              {t("devEnvironments.create_description")}
            </FormSheetDescription>
          </FormSheetHeader>
          <FormCardGroup className="overflow-y-auto">
            <FormCard className="border-none overflow-auto">
              <FormCardContent>
                <EnvironmentFormSheet
                  onSubmit={handleCreateEnvironment}
                  onCancel={handleCloseCreateSheet}
                  formId="environment-create-sheet-form"
                />
              </FormCardContent>
            </FormCard>
          </FormCardGroup>
          <FormSheetFooter>
            <Button
              type="submit"
              form="environment-create-sheet-form"
              disabled={isSubmitting}
            >
              <Save className="w-4 h-4 mr-2" />
              {isSubmitting
                ? t("common.saving")
                : t("devEnvironments.create")}
            </Button>
          </FormSheetFooter>
        </FormSheetContent>
      </FormSheet>

      {/* Edit Environment Sheet */}
      <FormSheet
        open={isEditSheetOpen}
        onOpenChange={setIsEditSheetOpen}
      >
        <FormSheetContent className="w-full sm:w-[600px] sm:max-w-[600px]">
          <FormSheetHeader>
            <FormSheetTitle>
              {t("devEnvironments.edit")} - {editingEnvironment?.name || ""}
            </FormSheetTitle>
            <FormSheetDescription>
              {t("devEnvironments.edit_description")}
            </FormSheetDescription>
          </FormSheetHeader>
          <FormCardGroup className="overflow-y-auto">
            <FormCard className="border-none overflow-auto">
              <FormCardContent>
                {editingEnvironment && (
                  <EnvironmentFormSheet
                    environment={editingEnvironment}
                    onSubmit={handleUpdateEnvironment}
                    onCancel={handleCloseEditSheet}
                    formId="environment-edit-sheet-form"
                  />
                )}
              </FormCardContent>
            </FormCard>
          </FormCardGroup>
          <FormSheetFooter>
            <Button
              type="submit"
              form="environment-edit-sheet-form"
              disabled={isSubmitting}
            >
              <Save className="w-4 h-4 mr-2" />
              {isSubmitting
                ? t("common.saving")
                : t("common.save")}
            </Button>
          </FormSheetFooter>
        </FormSheetContent>
      </FormSheet>
    </div>
  );
};

export default EnvironmentListPage;