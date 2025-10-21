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
import { ProviderFormSheet } from "@/components/ProviderFormSheet";
import { DataTable } from "@/components/ui/data-table/data-table";
import { DataTablePaginationServer } from "@/components/ui/data-table/data-table-pagination-server";
import { createProviderColumns } from "@/components/data-table/providers/columns";
import { ProviderDataTableToolbar } from "@/components/data-table/providers/data-table-toolbar";
import type { Provider, ProviderListParams, ProviderType } from "@/types/provider";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";

const ProviderListPage: React.FC = () => {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { setItems } = useBreadcrumb();
  const { setActions } = usePageActions();
  const { canCreateProvider, canEditProvider, canDeleteProvider, adminId } =
    usePermissions();

  usePageTitle(t("navigation.providers"));

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
    if (canCreateProvider) {
      setActions(
        <Button onClick={handleCreateNew} size="sm">
          <Plus className="h-4 w-4 mr-2" />
          {t("provider.create")}
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
  }, [setActions, setItems, t, canCreateProvider]);

  const [providers, setProviders] = useState<Provider[]>([]);
  const [loading, setLoading] = useState(true);
  const [providerTypes, setProviderTypes] = useState<ProviderType[]>([]);

  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);

  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);

  // Sheet state management
  const [isCreateSheetOpen, setIsCreateSheetOpen] = useState(false);
  const [isEditSheetOpen, setIsEditSheetOpen] = useState(false);
  const [editingProvider, setEditingProvider] = useState<Provider | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Add request deduplication
  const lastRequestRef = useRef<string>("");
  const isRequestInProgress = useRef(false);

  const pageSize = 10;

  // Load provider types
  useEffect(() => {
    const loadProviderTypes = async () => {
      try {
        const response = await apiService.providers.getTypes();
        setProviderTypes(response.types);
      } catch (error) {
        console.error("Failed to load provider types:", error);
        // Fallback to default type
        setProviderTypes(["claude-code"]);
      }
    };

    loadProviderTypes();
  }, []);

  // Separate URL update function to avoid dependency issues
  const updateUrl = useCallback(
    (page: number, filters: ColumnFiltersState) => {
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
    },
    [setSearchParams]
  );

  const loadProvidersData = useCallback(
    async (
      page: number,
      filters: ColumnFiltersState,
      shouldUpdateUrl = true
    ) => {
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
        const apiParams: ProviderListParams = {
          page,
          page_size: pageSize,
        };

        // Handle column filters
        filters.forEach((filter) => {
          if (filter.id === "name" && filter.value) {
            apiParams.name = filter.value as string;
          } else if (filter.id === "type" && filter.value) {
            apiParams.type = filter.value as ProviderType;
          }
        });

        const response = await apiService.providers.list(apiParams);
        setProviders(response.providers);
        setTotalPages(response.total_pages);
        setTotal(response.total);
        setCurrentPage(page);

        // Update URL parameters after successful data load
        if (shouldUpdateUrl) {
          updateUrl(page, filters);
        }
      } catch (error) {
        logError(error as Error, "Failed to fetch providers");
      } finally {
        setLoading(false);
        isRequestInProgress.current = false;

        // Clear the request key after a short delay to allow legitimate new requests
        setTimeout(() => {
          if (lastRequestRef.current === requestKey) {
            lastRequestRef.current = "";
          }
        }, 500);
      }
    },
    [pageSize, updateUrl]
  );

  const handleDeleteProvider = useCallback(
    async (id: number) => {
      try {
        await apiService.providers.delete(id);
        await loadProvidersData(currentPage, columnFilters);
      } catch (error) {
        // Re-throw error to let QuickActions handle the user notification
        throw error;
      }
    },
    [loadProvidersData, currentPage, columnFilters]
  );

  const handleEditProvider = useCallback(
    async (provider: Provider) => {
      try {
        setIsSubmitting(true);
        // Fetch full provider details
        const response = await apiService.providers.get(provider.id);

        setEditingProvider(response.provider);
        setIsEditSheetOpen(true);
      } catch (error) {
        console.error("Failed to fetch provider details:", error);
        toast.error(t("provider.fetch_details_failed"));
      } finally {
        setIsSubmitting(false);
      }
    },
    [t]
  );

  // Sheet handlers
  const handleCreateProvider = async (provider: Provider) => {
    try {
      setIsSubmitting(true);
      // Refresh the provider list
      await loadProvidersData(currentPage, columnFilters);
      // Close the sheet
      setIsCreateSheetOpen(false);
      // Show success message
      toast.success(t("provider.create_success"));
      console.log("Provider created successfully:", provider);
    } catch (error) {
      console.error("Failed to create provider:", error);
      logError(error as Error, "Failed to create provider");
      throw error;
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleUpdateProvider = async (provider: Provider) => {
    try {
      setIsSubmitting(true);
      // Refresh the provider list
      await loadProvidersData(currentPage, columnFilters);
      // Close the sheet
      setIsEditSheetOpen(false);
      setEditingProvider(null);
      // Show success message
      toast.success(t("provider.update_success"));
      console.log("Provider updated successfully:", provider);
    } catch (error) {
      console.error("Failed to update provider:", error);
      logError(error as Error, "Failed to update provider");
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
    setEditingProvider(null);
  };

  const columns = useMemo(
    () =>
      createProviderColumns({
        onEdit: handleEditProvider,
        onDelete: handleDeleteProvider,
        t,
        canEditProvider,
        canDeleteProvider,
      }),
    [handleEditProvider, handleDeleteProvider, t, canEditProvider, canDeleteProvider]
  );

  // Initialize from URL on component mount (only once)
  const [isInitialized, setIsInitialized] = useState(false);

  // Keep track of previous filter values to detect actual changes
  const previousFiltersRef = useRef<ColumnFiltersState>([]);

  useEffect(() => {
    // Get URL params directly to avoid dependency issues
    const nameParam = searchParams.get("name");
    const typeParam = searchParams.get("type");
    const pageParam = searchParams.get("page");

    const initialFilters: ColumnFiltersState = [];

    if (nameParam) {
      initialFilters.push({ id: "name", value: nameParam });
    }

    if (typeParam) {
      initialFilters.push({ id: "type", value: typeParam });
    }

    const initialPage = pageParam ? parseInt(pageParam, 10) : 1;

    // Set state first
    setColumnFilters(initialFilters);
    setCurrentPage(initialPage);
    previousFiltersRef.current = initialFilters;

    // Load initial data using the unified function
    loadProvidersData(initialPage, initialFilters, false).then(() => {
      setIsInitialized(true);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Empty dependency array - only run once on mount

  // Handle column filter changes (skip initial load)
  useEffect(() => {
    if (!isInitialized) return;

    // Check if the filter values actually changed
    const filtersChanged =
      JSON.stringify(previousFiltersRef.current) !==
      JSON.stringify(columnFilters);

    if (filtersChanged) {
      previousFiltersRef.current = columnFilters;
      loadProvidersData(1, columnFilters); // Reset to page 1 when filtering
    }
  }, [columnFilters, isInitialized, loadProvidersData]);

  const handlePageChange = useCallback(
    (page: number) => {
      loadProvidersData(page, columnFilters);
    },
    [columnFilters, loadProvidersData]
  );

  return (
    <div className="min-h-screen bg-background">
      <SectionGroup>
        <Section>
          <SectionHeader>
            <SectionTitle>{t("navigation.providers")}</SectionTitle>
            <SectionDescription>
              {t("provider.page_description")}
            </SectionDescription>
          </SectionHeader>
        </Section>
        <Section>
          <div className="space-y-4">
            <DataTable
              columns={columns}
              data={providers}
              toolbarComponent={(props) => (
                <ProviderDataTableToolbar
                  {...props}
                  providerTypes={providerTypes}
                />
              )}
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

      {/* Create Provider Sheet */}
      <FormSheet open={isCreateSheetOpen} onOpenChange={setIsCreateSheetOpen}>
        <FormSheetContent className="w-full sm:w-[600px] sm:max-w-[600px]">
          <FormSheetHeader>
            <FormSheetTitle>{t("provider.create")}</FormSheetTitle>
            <FormSheetDescription>
              {t("provider.create_description")}
            </FormSheetDescription>
          </FormSheetHeader>
          <FormCardGroup className="overflow-y-auto">
            <FormCard className="border-none overflow-auto">
              <FormCardContent>
                <ProviderFormSheet
                  onSubmit={handleCreateProvider}
                  onCancel={handleCloseCreateSheet}
                  formId="provider-create-sheet-form"
                />
              </FormCardContent>
            </FormCard>
          </FormCardGroup>
          <FormSheetFooter>
            <Button
              type="submit"
              form="provider-create-sheet-form"
              disabled={isSubmitting}
            >
              <Save className="w-4 h-4 mr-2" />
              {isSubmitting ? t("common.saving") : t("provider.create")}
            </Button>
          </FormSheetFooter>
        </FormSheetContent>
      </FormSheet>

      {/* Edit Provider Sheet */}
      <FormSheet open={isEditSheetOpen} onOpenChange={setIsEditSheetOpen}>
        <FormSheetContent className="w-full sm:w-[600px] sm:max-w-[600px]">
          <FormSheetHeader>
            <FormSheetTitle>
              {t("provider.edit")} - {editingProvider?.name || ""}
            </FormSheetTitle>
            <FormSheetDescription>
              {t("provider.edit_description")}
            </FormSheetDescription>
          </FormSheetHeader>
          <FormCardGroup className="overflow-y-auto">
            <FormCard className="border-none overflow-auto">
              <FormCardContent>
                {editingProvider && (
                  <ProviderFormSheet
                    provider={editingProvider}
                    onSubmit={handleUpdateProvider}
                    onCancel={handleCloseEditSheet}
                    formId="provider-edit-sheet-form"
                  />
                )}
              </FormCardContent>
            </FormCard>
          </FormCardGroup>
          <FormSheetFooter>
            <Button
              type="submit"
              form="provider-edit-sheet-form"
              disabled={isSubmitting}
            >
              <Save className="w-4 h-4 mr-2" />
              {isSubmitting ? t("common.saving") : t("common.save")}
            </Button>
          </FormSheetFooter>
        </FormSheetContent>
      </FormSheet>
    </div>
  );
};

export default ProviderListPage;
