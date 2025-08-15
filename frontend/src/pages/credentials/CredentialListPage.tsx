import React, {
  useState,
  useEffect,
  useMemo,
  useCallback,
  useRef,
} from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
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
import { CredentialFormSheet } from "@/components/CredentialFormSheet";
import { DataTable } from "@/components/ui/data-table/data-table";
import { DataTablePaginationServer } from "@/components/ui/data-table/data-table-pagination-server";
import { createGitCredentialColumns } from "@/components/data-table/credentials/columns";
import { GitCredentialDataTableToolbar } from "@/components/data-table/credentials/data-table-toolbar";

import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type {
  GitCredential,
  GitCredentialListParams,
} from "@/types/credentials";
import { Plus, Save } from "lucide-react";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";

const CredentialListPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { setItems } = useBreadcrumb();
  const { setActions } = usePageActions();

  const [credentials, setCredentials] = useState<GitCredential[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);

  // Sheet state management
  const [isCreateSheetOpen, setIsCreateSheetOpen] = useState(false);
  const [isEditSheetOpen, setIsEditSheetOpen] = useState(false);
  const [editingCredential, setEditingCredential] = useState<GitCredential | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Add request deduplication
  const lastRequestRef = useRef<string>("");
  const isRequestInProgress = useRef(false);

  const pageSize = 10;

  usePageTitle(t("common.pageTitle.gitCredentials"));

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
  }, [setActions, setItems, t]);

  const loadCredentialsData = useCallback(
    async (page: number, filters: ColumnFiltersState, updateUrl = true) => {
      // Create a unique request key for deduplication
      const requestKey = JSON.stringify({ page, filters, updateUrl });

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
        const apiParams: GitCredentialListParams = {
          page,
          page_size: pageSize,
        };

        // Handle column filters
        filters.forEach((filter) => {
          if (filter.id === "name" && filter.value) {
            apiParams.name = filter.value as string;
          }
        });

        const response = await apiService.gitCredentials.list(apiParams);
        setCredentials(response.credentials);
        setTotal(response.total);
        setTotalPages(response.total_pages);
        setCurrentPage(page);

        // Update URL parameters
        if (updateUrl) {
          const params = new URLSearchParams();

          // Add filter parameters
          filters.forEach((filter) => {
            if (filter.id === "name" && filter.value) {
              params.set(filter.id, String(filter.value));
            }
          });

          // Add page parameter (only if not page 1)
          if (page > 1) {
            params.set("page", String(page));
          }

          // Update URL without causing navigation
          setSearchParams(params, { replace: true });
        }
      } catch (error) {
        logError(error as Error, "Failed to load credentials");
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
    [pageSize, setSearchParams]
  );

  // Initialize from URL on component mount (only once)
  const [isInitialized, setIsInitialized] = useState(false);

  useEffect(() => {
    // Get URL params directly to avoid dependency issues
    const nameParam = searchParams.get("name");
    const pageParam = searchParams.get("page");

    const initialFilters: ColumnFiltersState = [];

    if (nameParam) {
      initialFilters.push({ id: "name", value: nameParam });
    }

    const initialPage = pageParam ? parseInt(pageParam, 10) : 1;

    // Set state first
    setColumnFilters(initialFilters);
    setCurrentPage(initialPage);

    // Load initial data using the unified function
    loadCredentialsData(initialPage, initialFilters, false).then(() => {
      setIsInitialized(true);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Empty dependency array - only run once on mount

  // Track previous filter values to detect actual filter changes
  const previousFiltersRef = useRef<ColumnFiltersState>([]);
  
  // Handle column filter changes (skip initial load)
  useEffect(() => {
    if (!isInitialized) {
      return;
    }

    // Check if filters actually changed in content, not just reference
    const currentFilterValues = columnFilters.map(f => ({ id: f.id, value: f.value }));
    const previousFilterValues = previousFiltersRef.current.map(f => ({ id: f.id, value: f.value }));
    
    const hasActualFilterChange = JSON.stringify(currentFilterValues) !== JSON.stringify(previousFilterValues);
    
    if (hasActualFilterChange) {
      // Only reset to page 1 when filter content actually changes
      loadCredentialsData(1, columnFilters);
    }
    
    // Update the previous filters reference
    previousFiltersRef.current = columnFilters;
  }, [columnFilters, isInitialized, loadCredentialsData]);

  const handlePageChange = useCallback(
    (page: number) => {
      loadCredentialsData(page, columnFilters);
    },
    [columnFilters, loadCredentialsData]
  );

  const handleEdit = useCallback(
    (credential: GitCredential) => {
      setEditingCredential(credential);
      setIsEditSheetOpen(true);
    },
    []
  );

  const handleDelete = useCallback(
    async (id: number) => {
      try {
        await apiService.gitCredentials.delete(id);
        await loadCredentialsData(currentPage, columnFilters);
      } catch (error) {
        // Re-throw error to let QuickActions handle the user notification
        throw error;
      }
    },
    [loadCredentialsData, currentPage, columnFilters]
  );




  // Sheet handlers
  const handleCreateCredential = async (credential: GitCredential) => {
    try {
      setIsSubmitting(true);
      // Refresh the credential list
      await loadCredentialsData(currentPage, columnFilters);
      // Close the sheet
      setIsCreateSheetOpen(false);
      // Show success message
      toast.success(t("gitCredentials.messages.createSuccess"));
      console.log("Credential created successfully:", credential);
    } catch (error) {
      console.error("Failed to create credential:", error);
      logError(error as Error, "Failed to create credential");
      throw error;
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleUpdateCredential = async (credential: GitCredential) => {
    try {
      setIsSubmitting(true);
      // Refresh the credential list
      await loadCredentialsData(currentPage, columnFilters);
      // Close the sheet
      setIsEditSheetOpen(false);
      setEditingCredential(null);
      // Show success message
      toast.success(t("gitCredentials.messages.updateSuccess"));
      console.log("Credential updated successfully:", credential);
    } catch (error) {
      console.error("Failed to update credential:", error);
      logError(error as Error, "Failed to update credential");
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
    setEditingCredential(null);
  };

  const columns = useMemo(
    () =>
      createGitCredentialColumns({
        onEdit: handleEdit,
        onDelete: handleDelete,
        t: (key: string) => t(key),
      }),
    [handleEdit, handleDelete, t]
  );



  return (
    <div>
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
          <div className="space-y-4">
            <DataTable
              columns={columns}
              data={credentials}
              toolbarComponent={GitCredentialDataTableToolbar}
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

      {/* Create Credential Sheet */}
      <FormSheet
        open={isCreateSheetOpen}
        onOpenChange={setIsCreateSheetOpen}
      >
        <FormSheetContent className="w-full sm:w-[600px] sm:max-w-[600px]">
          <FormSheetHeader>
            <FormSheetTitle>{t("gitCredentials.create")}</FormSheetTitle>
            <FormSheetDescription>
              {t("gitCredentials.createDescription")}
            </FormSheetDescription>
          </FormSheetHeader>
          <FormCardGroup className="overflow-y-auto">
            <FormCard className="border-none overflow-auto">
              <FormCardContent>
                <CredentialFormSheet
                  onSubmit={handleCreateCredential}
                  onCancel={handleCloseCreateSheet}
                  formId="credential-create-sheet-form"
                />
              </FormCardContent>
            </FormCard>
          </FormCardGroup>
          <FormSheetFooter>
            <Button
              type="submit"
              form="credential-create-sheet-form"
              disabled={isSubmitting}
            >
              <Save className="w-4 h-4 mr-2" />
              {isSubmitting
                ? t("common.saving")
                : t("gitCredentials.create")}
            </Button>
          </FormSheetFooter>
        </FormSheetContent>
      </FormSheet>

      {/* Edit Credential Sheet */}
      <FormSheet
        open={isEditSheetOpen}
        onOpenChange={setIsEditSheetOpen}
      >
        <FormSheetContent className="w-full sm:w-[600px] sm:max-w-[600px]">
          <FormSheetHeader>
            <FormSheetTitle>
              {t("gitCredentials.edit")} - {editingCredential?.name || ""}
            </FormSheetTitle>
            <FormSheetDescription>
              {t("gitCredentials.editDescription")}
            </FormSheetDescription>
          </FormSheetHeader>
          <FormCardGroup className="overflow-y-auto">
            <FormCard className="border-none overflow-auto">
              <FormCardContent>
                {editingCredential && (
                  <CredentialFormSheet
                    credential={editingCredential}
                    onSubmit={handleUpdateCredential}
                    onCancel={handleCloseEditSheet}
                    formId="credential-edit-sheet-form"
                  />
                )}
              </FormCardContent>
            </FormCard>
          </FormCardGroup>
          <FormSheetFooter>
            <Button
              type="submit"
              form="credential-edit-sheet-form"
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

export default CredentialListPage;
