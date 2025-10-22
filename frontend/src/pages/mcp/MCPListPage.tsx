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
import { createMCPColumns } from "@/components/data-table/mcp/columns";
import { MCPDataTableToolbar } from "@/components/data-table/mcp/data-table-toolbar";
import { MCPFormSheet } from "@/components/MCPFormSheet";
import { MCPTemplates } from "@/components/mcp/MCPTemplates";

import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { MCP, MCPListParams } from "@/types/mcp";
import { Plus } from "lucide-react";
import { usePermissions } from "@/hooks/usePermissions";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";

const MCPListPage: React.FC = () => {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { setItems } = useBreadcrumb();
  const { setActions } = usePageActions();
  const { canCreateMCP, canEditMCP, canDeleteMCP } = usePermissions();

  const [mcps, setMCPs] = useState<MCP[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);

  // Sheet state management
  const [isCreateSheetOpen, setIsCreateSheetOpen] = useState(false);
  const [isEditSheetOpen, setIsEditSheetOpen] = useState(false);
  const [editingMCP, setEditingMCP] = useState<MCP | null>(null);
  const [isLoadingEdit, setIsLoadingEdit] = useState(false);

  // Add request deduplication
  const lastRequestRef = useRef<string>("");
  const isRequestInProgress = useRef(false);

  const pageSize = 10;

  usePageTitle("common.pageTitle.mcp");

  // Check for action parameter to auto-open create sheet
  useEffect(() => {
    const actionParam = searchParams.get("action");
    if (actionParam === "create") {
      // Clear any existing editing state before opening create sheet
      setEditingMCP(null);
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
      setEditingMCP(null);
      setIsEditSheetOpen(false);
      setIsCreateSheetOpen(true);
    };

    // Only show create button if user has permission
    if (canCreateMCP) {
      setActions(
        <Button onClick={handleCreateNew} size="sm">
          <Plus className="h-4 w-4 mr-2" />
          {t("mcp.create")}
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
  }, [setActions, setItems, t, canCreateMCP]);

  // Build search parameters from filters
  const buildSearchParams = useCallback(
    (page: number, filters: ColumnFiltersState) => {
      const params: MCPListParams = {
        page,
        page_size: pageSize,
      };

      // Handle column filters
      filters.forEach((filter) => {
        if (filter.id === "search" && filter.value) {
          // Handle search filter for name and description
          params.name = filter.value as string;
        } else if (
          filter.id === "enabled" &&
          Array.isArray(filter.value) &&
          filter.value.length > 0
        ) {
          // Handle enabled filter
          if (filter.value.length === 1) {
            // Single selection
            params.enabled = filter.value[0] === "enabled";
          }
          // Both enabled and disabled selected means no filter
        }
      });

      return params;
    },
    [pageSize]
  );

  // Load MCPs data
  const loadMCPsData = useCallback(
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

          const response = await apiService.mcp.list(params);

          setMCPs(response.mcps || []);
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
                  filter.id === "enabled" &&
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
          logError(error, "Failed to load MCP configurations");
          toast.error(t("mcp.errors.loadFailed"));
          setMCPs([]);
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
    const enabledParam = searchParams.get("enabled");
    const pageParam = searchParams.get("page");

    const initialFilters: ColumnFiltersState = [];

    if (searchParam) {
      initialFilters.push({ id: "search", value: searchParam });
    }

    if (enabledParam) {
      initialFilters.push({ id: "enabled", value: [enabledParam] });
    }

    const initialPage = pageParam ? parseInt(pageParam, 10) : 1;
    const initialSorting: SortingState = [];

    // Set state first
    setColumnFilters(initialFilters);
    setCurrentPage(initialPage);
    setSorting(initialSorting);

    // Load initial data using the unified function
    loadMCPsData(
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
      loadMCPsData(1, columnFilters, sorting); // Reset to page 1 when filtering or sorting
    }
  }, [columnFilters, sorting, isInitialized, loadMCPsData]);

  const handlePageChange = useCallback(
    (page: number) => {
      loadMCPsData(page, columnFilters, sorting);
    },
    [columnFilters, sorting, loadMCPsData]
  );

  // Handle edit MCP
  const handleEdit = useCallback(
    async (mcp: MCP) => {
      // Prevent multiple clicks while loading
      if (isLoadingEdit) {
        return;
      }

      // Clear any existing create sheet state before opening edit sheet
      setIsCreateSheetOpen(false);

      setIsLoadingEdit(true);
      try {
        const response = await apiService.mcp.get(mcp.id);
        // Response is the MCP object directly, not wrapped in data
        setEditingMCP(response);
        setIsEditSheetOpen(true);
      } catch (error) {
        logError(error, "Failed to load MCP details");
        toast.error(t("mcp.errors.loadDetailsFailed"));
      } finally {
        setIsLoadingEdit(false);
      }
    },
    [isLoadingEdit, t]
  );

  // Handle delete MCP
  const handleDelete = useCallback(
    async (id: number) => {
      try {
        await apiService.mcp.delete(id);
        toast.success(t("mcp.deleteSuccess"));
        await loadMCPsData(currentPage, columnFilters, sorting, false);
      } catch (error) {
        logError(error, "Failed to delete MCP configuration");
        toast.error(t("mcp.errors.deleteFailed"));
        // Re-throw error to let QuickActions handle the user notification
        throw error;
      }
    },
    [currentPage, columnFilters, sorting, loadMCPsData, t]
  );

  // Handle toggle MCP status
  const handleToggleStatus = useCallback(
    async (id: number, enabled: boolean) => {
      try {
        await apiService.mcp.update(id, { enabled: enabled });
        toast.success(
          enabled ? t("mcp.enableSuccess") : t("mcp.disableSuccess")
        );
        await loadMCPsData(currentPage, columnFilters, sorting, false);
      } catch (error) {
        logError(error, "Failed to update MCP status");
        toast.error(t("mcp.errors.updateFailed"));
      }
    },
    [currentPage, columnFilters, sorting, loadMCPsData, t]
  );

  // Permission check helpers
  const canEditMCPFunc = useCallback(
    (resourceAdminId?: number) => canEditMCP(resourceAdminId),
    [canEditMCP]
  );

  const canDeleteMCPFunc = useCallback(
    (resourceAdminId?: number) => canDeleteMCP(resourceAdminId),
    [canDeleteMCP]
  );

  // Create table columns with handlers
  const columns = useMemo(
    () =>
      createMCPColumns({
        onEdit: handleEdit,
        onDelete: handleDelete,
        onToggleStatus: handleToggleStatus,
        t,
        canEditMCP: canEditMCPFunc,
        canDeleteMCP: canDeleteMCPFunc,
      }),
    [
      handleEdit,
      handleDelete,
      handleToggleStatus,
      t,
      canEditMCPFunc,
      canDeleteMCPFunc,
    ]
  );

  return (
    <SectionGroup>
      <SectionHeader>
        <SectionTitle>{t("mcp.title")}</SectionTitle>
        <SectionDescription>{t("mcp.description")}</SectionDescription>
      </SectionHeader>

      <Section>
        <div className="space-y-4">
          <DataTable
            columns={columns}
            data={mcps}
            toolbarComponent={MCPDataTableToolbar}
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

      {/* MCP Templates Section */}
      <MCPTemplates
        onMCPCreated={async () => {
          await loadMCPsData(currentPage, columnFilters, sorting, false);
        }}
      />

      {/* Create MCP Form */}
      <MCPFormSheet
        isOpen={isCreateSheetOpen}
        onClose={() => setIsCreateSheetOpen(false)}
        onSuccess={async () => {
          toast.success(t("mcp.createSuccess"));
          await loadMCPsData(currentPage, columnFilters, sorting, false);
        }}
      />

      {/* Edit MCP Form */}
      <MCPFormSheet
        isOpen={isEditSheetOpen}
        onClose={() => {
          setIsEditSheetOpen(false);
          setEditingMCP(null);
        }}
        onSuccess={async () => {
          toast.success(t("mcp.updateSuccess"));
          await loadMCPsData(currentPage, columnFilters, sorting, false);
        }}
        mcp={editingMCP || undefined}
      />
    </SectionGroup>
  );
};

export default MCPListPage;
