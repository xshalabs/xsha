import React, {
  useState,
  useEffect,
  useMemo,
  useCallback,
  useRef,
} from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";
import { usePageTitle } from "@/hooks/usePageTitle";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";
import { usePageActions } from "@/contexts/PageActionsContext";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

import {
  Section,
  SectionGroup,
  SectionHeader,
  SectionTitle,
  SectionDescription,
} from "@/components/content/section";
import { DataTable } from "@/components/ui/data-table/data-table";
import { DataTablePaginationServer } from "@/components/ui/data-table/data-table-pagination-server";
import { createNotifierColumns } from "@/components/data-table/notifiers/columns";
import { NotifierDataTableToolbar } from "@/components/data-table/notifiers/data-table-toolbar";
import { NotifierFormSheet } from "@/components/NotifierFormSheet";

import { apiService } from "@/lib/api/index";
import { logError, handleApiError } from "@/lib/errors";
import type {
  Notifier,
  NotifierListParams,
  NotifierType,
} from "@/types/notifier";
import { Plus } from "lucide-react";
import { usePermissions } from "@/hooks/usePermissions";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";

const NotifierListPage: React.FC = () => {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { setItems } = useBreadcrumb();
  const { setActions } = usePageActions();
  const { canCreateNotifier, canEditNotifier, canDeleteNotifier } =
    usePermissions();

  const [notifiers, setNotifiers] = useState<Notifier[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);

  // Sheet state management
  const [isCreateSheetOpen, setIsCreateSheetOpen] = useState(false);
  const [isEditSheetOpen, setIsEditSheetOpen] = useState(false);
  const [editingNotifier, setEditingNotifier] = useState<Notifier | null>(null);
  const [isLoadingEdit, setIsLoadingEdit] = useState(false);

  // Add request deduplication
  const lastRequestRef = useRef<string>("");
  const isRequestInProgress = useRef(false);

  const pageSize = 10;

  usePageTitle("common.pageTitle.notifiers");

  // Check for action parameter to auto-open create sheet
  useEffect(() => {
    const actionParam = searchParams.get("action");
    if (actionParam === "create") {
      // Clear any existing editing state before opening create sheet
      setEditingNotifier(null);
      setIsEditSheetOpen(false);
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
      // Clear any existing editing state before opening create sheet
      setEditingNotifier(null);
      setIsEditSheetOpen(false);
      setIsCreateSheetOpen(true);
    };

    // Only show create button if user has permission
    if (canCreateNotifier) {
      setActions(
        <Button onClick={handleCreateNew} size="sm">
          <Plus className="h-4 w-4 mr-2" />
          {t("notifiers.create")}
        </Button>
      );
    } else {
      setActions(null);
    }

    setItems([]);

    return () => {
      setActions(null);
      setItems([]);
    };
  }, [setActions, setItems, t, canCreateNotifier]);

  // Build search parameters from filters
  const buildSearchParams = useCallback(
    (page: number, filters: ColumnFiltersState) => {
      const params: NotifierListParams = {
        page,
        page_size: pageSize,
      };

      // Handle column filters
      filters.forEach((filter) => {
        if (filter.id === "search" && filter.value) {
          // Handle search filter for name and description
          params.name = filter.value as string;
        } else if (
          filter.id === "type" &&
          Array.isArray(filter.value) &&
          filter.value.length > 0
        ) {
          // Handle type filter - pass all selected types as comma-separated string
          params.type = filter.value.join(",") as NotifierType;
        } else if (
          filter.id === "is_enabled" &&
          Array.isArray(filter.value) &&
          filter.value.length > 0
        ) {
          // Handle enabled filter
          if (filter.value.length === 1) {
            // Single selection
            params.is_enabled = filter.value[0] === "enabled";
          }
          // Both enabled and disabled selected means no filter
        }
      });

      return params;
    },
    [pageSize]
  );

  // Load notifiers data
  const loadNotifiersData = useCallback(
    async (
      page: number,
      filters: ColumnFiltersState,
      sortingState: SortingState,
      shouldDebounce = true,
      updateUrl = true
    ) => {
      // Create a unique request key for deduplication
      const requestKey = JSON.stringify({
        page,
        filters,
        sortingState,
        updateUrl,
      });

      // Skip if same request is already in progress or just completed
      if (
        isRequestInProgress.current ||
        lastRequestRef.current === requestKey
      ) {
        return;
      }

      if (shouldDebounce) {
        // Debounce to prevent rapid duplicate requests
        const debounceTimer = setTimeout(async () => {
          if (lastRequestRef.current === requestKey) {
            return; // Request was cancelled
          }

          lastRequestRef.current = requestKey;
          await executeRequest();
        }, 500); // Increased delay to prevent rapid duplicate requests

        // Store timer for potential cleanup
        return () => clearTimeout(debounceTimer);
      } else {
        lastRequestRef.current = requestKey;
        await executeRequest();
      }

      async function executeRequest() {
        isRequestInProgress.current = true;

        try {
          setLoading(true);
          const params = buildSearchParams(page, filters);

          const response = await apiService.notifiers.list(params);

          setNotifiers(response.data || []);
          setTotal(response.total || 0);
          setTotalPages(Math.ceil((response.total || 0) / pageSize));
          setCurrentPage(page);

          // Update URL parameters
          if (updateUrl) {
            const urlParams = new URLSearchParams();

            // Add filter parameters
            filters.forEach((filter) => {
              if (filter.value) {
                if (filter.id === "search") {
                  // Handle search parameter
                  urlParams.set(filter.id, String(filter.value));
                } else if (
                  filter.id === "type" &&
                  Array.isArray(filter.value) &&
                  filter.value.length > 0
                ) {
                  // Handle type filter - join multiple values with comma
                  urlParams.set(filter.id, filter.value.join(","));
                } else if (
                  filter.id === "is_enabled" &&
                  Array.isArray(filter.value) &&
                  filter.value.length > 0
                ) {
                  // Only set parameter if not both values are selected (which means no filter)
                  if (filter.value.length === 1) {
                    urlParams.set(filter.id, filter.value[0]);
                  }
                }
              }
            });

            // Add page parameter (only if not page 1)
            if (page > 1) {
              urlParams.set("page", String(page));
            }

            // Update URL without causing navigation
            setSearchParams(urlParams, { replace: true });
          }
        } catch (error) {
          logError(error, "Failed to load notifiers");
          toast.error(t("notifiers.errors.loadFailed"));
          setNotifiers([]);
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
      }
    },
    [buildSearchParams, setSearchParams, t]
  );

  // Initialize from URL on component mount (only once)
  const [isInitialized, setIsInitialized] = useState(false);

  // Initialize from URL parameters
  useEffect(() => {
    const searchParam = searchParams.get("search");
    const typeParam = searchParams.get("type");
    const enabledParam = searchParams.get("is_enabled");
    const pageParam = searchParams.get("page");

    const initialFilters: ColumnFiltersState = [];

    if (searchParam) {
      initialFilters.push({ id: "search", value: searchParam });
    }

    if (typeParam) {
      initialFilters.push({ id: "type", value: typeParam.split(",") });
    }

    if (enabledParam) {
      initialFilters.push({ id: "is_enabled", value: [enabledParam] });
    }

    const initialPage = pageParam ? parseInt(pageParam, 10) : 1;
    const initialSorting: SortingState = [];

    // Set state first
    setColumnFilters(initialFilters);
    setCurrentPage(initialPage);
    setSorting(initialSorting);

    // Load initial data using the unified function
    loadNotifiersData(
      initialPage,
      initialFilters,
      initialSorting,
      false,
      false
    ).then(() => {
      setIsInitialized(true);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Empty dependency array - only run once on mount

  // Handle filter and sorting changes (skip initial load)
  useEffect(() => {
    if (isInitialized) {
      loadNotifiersData(1, columnFilters, sorting); // Reset to page 1 when filtering or sorting
    }
  }, [columnFilters, sorting, isInitialized, loadNotifiersData]);

  const handlePageChange = useCallback(
    (page: number) => {
      loadNotifiersData(page, columnFilters, sorting);
    },
    [columnFilters, sorting, loadNotifiersData]
  );

  // Handle edit notifier
  const handleEdit = useCallback(
    async (notifier: Notifier) => {
      // Prevent multiple clicks while loading
      if (isLoadingEdit) {
        return;
      }

      // Clear any existing create sheet state before opening edit sheet
      setIsCreateSheetOpen(false);

      setIsLoadingEdit(true);
      try {
        const response = await apiService.notifiers.get(notifier.id);
        // Response is the notifier object directly, not wrapped in data
        setEditingNotifier(response);
        setIsEditSheetOpen(true);
      } catch (error) {
        logError(error, "Failed to load notifier details");
        toast.error(t("notifiers.errors.loadDetailsFailed"));
      } finally {
        setIsLoadingEdit(false);
      }
    },
    [isLoadingEdit, t]
  );

  // Handle delete notifier
  const handleDelete = useCallback(
    async (id: number) => {
      try {
        await apiService.notifiers.delete(id);
        toast.success(t("notifiers.deleteSuccess"));
        await loadNotifiersData(currentPage, columnFilters, sorting, false);
      } catch (error) {
        logError(error, "Failed to delete notifier");
        toast.error(t("notifiers.errors.deleteFailed"));
        // Re-throw error to let QuickActions handle the user notification
        throw error;
      }
    },
    [currentPage, columnFilters, sorting, loadNotifiersData, t]
  );

  // Handle test notifier
  const handleTest = useCallback(
    async (id: number) => {
      try {
        await apiService.notifiers.test(id);
        toast.success(t("notifiers.testSuccess"));
      } catch (error) {
        logError(error, "Failed to test notifier");
        const errorMessage = handleApiError(error);
        toast.error(`${t("notifiers.errors.testFailed")}: ${errorMessage}`);
      }
    },
    [t]
  );

  // Handle toggle notifier status
  const handleToggleStatus = useCallback(
    async (id: number, enabled: boolean) => {
      try {
        await apiService.notifiers.update(id, { is_enabled: enabled });
        toast.success(
          enabled ? t("notifiers.enableSuccess") : t("notifiers.disableSuccess")
        );
        await loadNotifiersData(currentPage, columnFilters, sorting, false);
      } catch (error) {
        logError(error, "Failed to update notifier status");
        toast.error(t("notifiers.errors.updateFailed"));
      }
    },
    [currentPage, columnFilters, sorting, loadNotifiersData, t]
  );

  // Permission check helpers
  const canEditNotifierFunc = useCallback(
    (resourceAdminId?: number) => canEditNotifier(resourceAdminId),
    [canEditNotifier]
  );

  const canDeleteNotifierFunc = useCallback(
    (resourceAdminId?: number) => canDeleteNotifier(resourceAdminId),
    [canDeleteNotifier]
  );

  const canTestNotifierFunc = useCallback(
    (resourceAdminId?: number) => canEditNotifier(resourceAdminId),
    [canEditNotifier]
  );

  // Create table columns with handlers
  const columns = useMemo(
    () =>
      createNotifierColumns({
        onEdit: handleEdit,
        onDelete: handleDelete,
        onTest: handleTest,
        onToggleStatus: handleToggleStatus,
        t,
        canEditNotifier: canEditNotifierFunc,
        canDeleteNotifier: canDeleteNotifierFunc,
        canTestNotifier: canTestNotifierFunc,
      }),
    [
      handleEdit,
      handleDelete,
      handleTest,
      handleToggleStatus,
      t,
      canEditNotifierFunc,
      canDeleteNotifierFunc,
      canTestNotifierFunc,
    ]
  );

  return (
    <SectionGroup>
      <SectionHeader>
        <SectionTitle>{t("notifiers.title")}</SectionTitle>
        <SectionDescription>{t("notifiers.description")}</SectionDescription>
      </SectionHeader>

      <Section>
        <div className="space-y-4">
          <DataTable
            columns={columns}
            data={notifiers}
            toolbarComponent={NotifierDataTableToolbar}
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

      {/* Create Notifier Form */}
      <NotifierFormSheet
        isOpen={isCreateSheetOpen}
        onClose={() => setIsCreateSheetOpen(false)}
        onSuccess={async () => {
          toast.success(t("notifiers.createSuccess"));
          await loadNotifiersData(currentPage, columnFilters, sorting, false);
        }}
      />

      {/* Edit Notifier Form */}
      <NotifierFormSheet
        isOpen={isEditSheetOpen}
        onClose={() => {
          setIsEditSheetOpen(false);
          setEditingNotifier(null);
        }}
        onSuccess={async () => {
          toast.success(t("notifiers.updateSuccess"));
          await loadNotifiersData(currentPage, columnFilters, sorting, false);
        }}
        notifier={editingNotifier || undefined}
      />
    </SectionGroup>
  );
};

export default NotifierListPage;
