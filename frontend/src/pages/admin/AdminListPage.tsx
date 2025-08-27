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
import { DeleteAdminDialog } from '@/components/admin/DeleteAdminDialog';
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
import { Plus } from 'lucide-react';
import { toast } from 'sonner';
import { logError } from '@/lib/errors';
import type { ColumnFiltersState, SortingState } from '@tanstack/react-table';

export default function AdminListPage() {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { setItems } = useBreadcrumb();
  const { setActions } = usePageActions();
  
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
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedAdmin, setSelectedAdmin] = useState<Admin | null>(null);

  const loadAdminsData = useCallback(
    async (page: number, filters: ColumnFiltersState, sortingState: SortingState, updateUrl = true) => {
      // Create a unique request key for deduplication
      const requestKey = JSON.stringify({ page, filters, sortingState, updateUrl });

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
        const apiParams: any = {
          page,
          page_size: pageSize,
        };

        // Handle column filters
        filters.forEach((filter) => {
          if (filter.id === "username" && filter.value) {
            apiParams.username = filter.value as string;
          } else if (filter.id === "is_active" && filter.value !== undefined) {
            if (filter.value !== "all") {
              apiParams.is_active = filter.value === "active";
            }
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
            if (filter.value && filter.value !== "all") {
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
        console.error('Failed to load admins:', error);
        toast.error(t('admin.errors.loadFailed'));
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
    },
    [pageSize, setSearchParams, t]
  );

  // Initialize from URL on component mount (only once)
  const [isInitialized, setIsInitialized] = useState(false);

  // Initialize from URL parameters
  useEffect(() => {
    const usernameParam = searchParams.get("username");
    const statusParam = searchParams.get("is_active");
    const pageParam = searchParams.get("page");

    const initialFilters: ColumnFiltersState = [];

    if (usernameParam) {
      initialFilters.push({ id: "username", value: usernameParam });
    }

    if (statusParam && statusParam !== "all") {
      initialFilters.push({ id: "is_active", value: statusParam });
    }

    const initialPage = pageParam ? parseInt(pageParam, 10) : 1;
    const initialSorting: SortingState = [];

    // Set state first
    setColumnFilters(initialFilters);
    setCurrentPage(initialPage);
    setSorting(initialSorting);

    // Load initial data using the unified function
    loadAdminsData(initialPage, initialFilters, initialSorting, false).then(() => {
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

    setActions(
      <Button onClick={handleCreateNew} size="sm">
        <Plus className="h-4 w-4 mr-2" />
        {t('admin.actions.create')}
      </Button>
    );

    // Clear breadcrumb items (we're at the root level)
    setItems([]);

    // Cleanup when component unmounts
    return () => {
      setActions(null);
      setItems([]);
    };
  }, [setActions, setItems, t]);

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

  const handleDelete = useCallback(
    async (admin: Admin) => {
      try {
        // Delete logic will be handled by the QuickActions component
        // This should not be called directly, but kept for consistency
        setSelectedAdmin(admin);
        setDeleteDialogOpen(true);
      } catch (error) {
        // Re-throw error to let QuickActions handle the user notification
        throw error;
      }
    },
    []
  );

  const handleCreateSuccess = () => {
    setCreateDialogOpen(false);
    loadAdminsData(currentPage, columnFilters, sorting);
  };

  const handleUpdateSuccess = () => {
    setUpdateDialogOpen(false);
    setSelectedAdmin(null);
    loadAdminsData(currentPage, columnFilters, sorting);
  };

  const handleChangePasswordSuccess = () => {
    setChangePasswordDialogOpen(false);
    setSelectedAdmin(null);
  };

  const handleDeleteSuccess = () => {
    setDeleteDialogOpen(false);
    setSelectedAdmin(null);
    loadAdminsData(currentPage, columnFilters, sorting);
  };

  const columns = useMemo(
    () =>
      createAdminColumns({
        t,
        onEdit: handleEdit,
        onChangePassword: handleChangePassword,
        onDelete: handleDelete,
      }),
    [t, handleEdit, handleChangePassword, handleDelete]
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
          <DeleteAdminDialog
            open={deleteDialogOpen}
            onOpenChange={setDeleteDialogOpen}
            admin={selectedAdmin}
            onSuccess={handleDeleteSuccess}
          />
        </>
      )}
    </div>
  );
}