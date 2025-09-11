import {
  useEffect,
  useState,
  useMemo,
  useCallback,
  useRef,
} from 'react';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import { useBreadcrumb } from '@/contexts/BreadcrumbContext';
import { usePageActions } from '@/contexts/PageActionsContext';
import { usePageTitle } from '@/hooks/usePageTitle';
import { CreateAdminDialog } from '@/components/admin/CreateAdminDialog';
import { UpdateAdminDialog } from '@/components/admin/UpdateAdminDialog';
import { ChangePasswordDialog } from '@/components/admin/ChangePasswordDialog';
import { AvatarUploadDialog } from '@/components/admin/AvatarUploadDialog';
import { Button } from '@/components/ui/button';
import { DataTable } from '@/components/ui/data-table/data-table';
import { DataTablePaginationServer } from '@/components/ui/data-table/data-table-pagination-server';
import {
  Section,
  SectionDescription,
  SectionGroup,
  SectionHeader,
  SectionTitle,
} from '@/components/content/section';
import { createAdminColumns } from '@/components/data-table/admin/columns';
import { AdminDataTableToolbar } from '@/components/data-table/admin/data-table-toolbar';
import { adminApi, type Admin } from '@/lib/api';
import { usePermissions } from '@/hooks/usePermissions';
import { Plus } from 'lucide-react';
import { toast } from 'sonner';
import { logError, handleApiError } from '@/lib/errors';
import type { ColumnFiltersState, SortingState } from '@tanstack/react-table';

export default function AdminListPage() {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { setItems } = useBreadcrumb();
  const { setActions } = usePageActions();
  const permissions = usePermissions();
  
  usePageTitle('admin.pageTitle.list');

  const [admins, setAdmins] = useState<Admin[]>([]);
  const [loading, setLoading] = useState(true);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);

  // Add request deduplication
  const lastRequestRef = useRef<string>("");
  const isRequestInProgress = useRef(false);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 20;
  
  // Dialog states
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [updateDialogOpen, setUpdateDialogOpen] = useState(false);
  const [changePasswordDialogOpen, setChangePasswordDialogOpen] = useState(false);
  const [avatarDialogOpen, setAvatarDialogOpen] = useState(false);
  const [selectedAdmin, setSelectedAdmin] = useState<Admin | null>(null);

  const loadAdminsData = useCallback(
    async (page: number, filters: ColumnFiltersState, sortingState: SortingState, shouldDebounce = true, updateUrl = true) => {
      // Create a unique request key for deduplication
      const requestKey = JSON.stringify({ page, filters, sortingState, updateUrl });

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

        // Convert DataTable filters to API parameters
        const apiParams: any = {
          page,
          page_size: pageSize,
        };

        // Handle column filters
        filters.forEach((filter) => {
          if (filter.id === "search" && filter.value) {
            // Handle search filter for email, name, username
            apiParams.search = filter.value as string;
          } else if (filter.id === "is_active" && Array.isArray(filter.value) && filter.value.length > 0) {
            // Handle faceted filter with array values
            if (filter.value.length === 1) {
              // Single selection
              apiParams.is_active = filter.value[0] === "active";
            } else if (filter.value.length === 2) {
              // Both active and inactive selected, don't filter
              // apiParams.is_active remains undefined
            }
          } else if (filter.id === "role" && Array.isArray(filter.value) && filter.value.length > 0) {
            // Handle role filter
            apiParams.role = filter.value;
          }
        });

        // Handle sorting (basic implementation for now)
        // Note: Backend might need updates to support sorting

        const response = await adminApi.getAdmins(apiParams);
        setAdmins(response.admins);
        setTotal(response.total);
        setTotalPages(Math.ceil(response.total / pageSize));
        setCurrentPage(page);

        // Update URL parameters
        if (updateUrl) {
          const params = new URLSearchParams();

          // Add filter parameters
          filters.forEach((filter) => {
            if (filter.value) {
              if (filter.id === "search") {
                // Handle search parameter
                params.set(filter.id, String(filter.value));
              } else if (filter.id === "is_active" && Array.isArray(filter.value) && filter.value.length > 0) {
                // Only set parameter if not both values are selected (which means no filter)
                if (filter.value.length === 1) {
                  params.set(filter.id, filter.value[0]);
                }
              } else if (filter.id === "role" && Array.isArray(filter.value) && filter.value.length > 0) {
                // Handle role filter - join multiple values with comma
                params.set(filter.id, filter.value.join(","));
              }
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
        console.error('Failed to load admins:', error);
        toast.error(handleApiError(error));
        logError(error as Error, "Failed to load admins");
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
    [pageSize, setSearchParams, t]
  );

  // Initialize from URL on component mount (only once)
  const [isInitialized, setIsInitialized] = useState(false);

  // Initialize from URL parameters
  useEffect(() => {
    const searchParam = searchParams.get("search");
    const statusParam = searchParams.get("is_active");
    const roleParam = searchParams.get("role");
    const pageParam = searchParams.get("page");

    const initialFilters: ColumnFiltersState = [];

    if (searchParam) {
      initialFilters.push({ id: "search", value: searchParam });
    }

    if (statusParam) {
      initialFilters.push({ id: "is_active", value: [statusParam] });
    }

    if (roleParam) {
      initialFilters.push({ id: "role", value: roleParam.split(",") });
    }

    const initialPage = pageParam ? parseInt(pageParam, 10) : 1;
    const initialSorting: SortingState = [];

    // Set state first
    setColumnFilters(initialFilters);
    setCurrentPage(initialPage);
    setSorting(initialSorting);

    // Load initial data using the unified function
    loadAdminsData(initialPage, initialFilters, initialSorting, false, false).then(() => {
      setIsInitialized(true);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Empty dependency array - only run once on mount

  // Handle filter and sorting changes (skip initial load)
  useEffect(() => {
    if (isInitialized) {
      loadAdminsData(1, columnFilters, sorting); // Reset to page 1 when filtering or sorting
    }
  }, [columnFilters, sorting, isInitialized, loadAdminsData]);

  const handlePageChange = useCallback(
    (page: number) => {
      loadAdminsData(page, columnFilters, sorting);
    },
    [columnFilters, sorting, loadAdminsData]
  );

  // Set page actions (Create button in header) and clear breadcrumb
  useEffect(() => {
    const handleCreateNew = () => {
      setCreateDialogOpen(true);
    };

    // Only show create button if user has permission
    const createButton = permissions.canCreateAdmin ? (
      <Button onClick={handleCreateNew} size="sm">
        <Plus className="h-4 w-4 mr-2" />
        {t('admin.actions.create')}
      </Button>
    ) : null;

    setActions(createButton);

    // Clear breadcrumb items (we're at the root level)
    setItems([]);

    // Cleanup when component unmounts
    return () => {
      setActions(null);
      setItems([]);
    };
  }, [setActions, setItems, t, permissions.canCreateAdmin]);

  // Handle admin actions
  const handleEdit = useCallback(
    (admin: Admin) => {
      setSelectedAdmin(admin);
      setUpdateDialogOpen(true);
    },
    []
  );

  const handleChangePassword = useCallback(
    (admin: Admin) => {
      setSelectedAdmin(admin);
      setChangePasswordDialogOpen(true);
    },
    []
  );

  const handleAvatarClick = useCallback(
    (admin: Admin) => {
      setSelectedAdmin(admin);
      setAvatarDialogOpen(true);
    },
    []
  );

  const handleDelete = useCallback(
    async (admin: Admin) => {
      await adminApi.deleteAdmin(admin.id);
      await loadAdminsData(currentPage, columnFilters, sorting, false);
    },
    [loadAdminsData, currentPage, columnFilters, sorting]
  );

  const handleCreateSuccess = () => {
    setCreateDialogOpen(false);
    loadAdminsData(currentPage, columnFilters, sorting, false);
  };

  const handleUpdateSuccess = () => {
    setUpdateDialogOpen(false);
    setSelectedAdmin(null);
    loadAdminsData(currentPage, columnFilters, sorting, false);
  };

  const handleChangePasswordSuccess = () => {
    setChangePasswordDialogOpen(false);
    setSelectedAdmin(null);
  };

  const handleAvatarUploadSuccess = () => {
    setAvatarDialogOpen(false);
    setSelectedAdmin(null);
    loadAdminsData(currentPage, columnFilters, sorting, false);
  };


  const columns = useMemo(
    () =>
      createAdminColumns({
        t,
        onEdit: handleEdit,
        onChangePassword: handleChangePassword,
        onDelete: handleDelete,
        onAvatarClick: handleAvatarClick,
        permissions: {
          canEditAdmin: permissions.canEditAdmin,
          canChangeAdminPassword: permissions.canChangeAdminPassword,
          canDeleteAdmin: permissions.canDeleteAdmin,
        },
      }),
    [t, handleEdit, handleChangePassword, handleDelete, handleAvatarClick, permissions]
  );

  return (
    <div className="min-h-screen bg-background">
      <SectionGroup>
        <Section>
          <SectionHeader>
            <SectionTitle>{t('admin.title')}</SectionTitle>
            <SectionDescription>
              {t('admin.description')}
            </SectionDescription>
          </SectionHeader>
        </Section>
        <Section>
          <div className="space-y-4">
            <DataTable
              columns={columns}
              data={admins}
              toolbarComponent={AdminDataTableToolbar}
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

      {/* Dialogs */}
      <CreateAdminDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSuccess={handleCreateSuccess}
      />

      {selectedAdmin && (
        <>
          <UpdateAdminDialog
            open={updateDialogOpen}
            onOpenChange={setUpdateDialogOpen}
            admin={selectedAdmin}
            onSuccess={handleUpdateSuccess}
          />
          <ChangePasswordDialog
            open={changePasswordDialogOpen}
            onOpenChange={setChangePasswordDialogOpen}
            admin={selectedAdmin}
            onSuccess={handleChangePasswordSuccess}
          />
          <AvatarUploadDialog
            open={avatarDialogOpen}
            onOpenChange={setAvatarDialogOpen}
            admin={selectedAdmin}
            onSuccess={handleAvatarUploadSuccess}
          />
        </>
      )}
    </div>
  );
}