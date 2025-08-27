import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useBreadcrumb } from '@/contexts/BreadcrumbContext';
import { usePageActions } from '@/contexts/PageActionsContext';
import { usePageTitle } from '@/hooks/usePageTitle';
import { AdminTable } from '@/components/admin/AdminTable';
import { CreateAdminDialog } from '@/components/admin/CreateAdminDialog';
import { UpdateAdminDialog } from '@/components/admin/UpdateAdminDialog';
import { ChangePasswordDialog } from '@/components/admin/ChangePasswordDialog';
import { DeleteAdminDialog } from '@/components/admin/DeleteAdminDialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { adminApi, type Admin } from '@/lib/api';
import { Plus, Search } from 'lucide-react';
import { toast } from 'sonner';

export default function AdminListPage() {
  const { t } = useTranslation();
  const { setItems } = useBreadcrumb();
  const { setActions } = usePageActions();
  
  usePageTitle('admin.pageTitle.list');

  const [admins, setAdmins] = useState<Admin[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [searchUsername, setSearchUsername] = useState('');
  const [statusFilter, setStatusFilter] = useState<boolean | undefined>(undefined);
  
  // Dialog states
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [updateDialogOpen, setUpdateDialogOpen] = useState(false);
  const [changePasswordDialogOpen, setChangePasswordDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedAdmin, setSelectedAdmin] = useState<Admin | null>(null);

  // Set breadcrumbs and page actions
  useEffect(() => {
    setItems([
      { type: 'link', label: t('navigation.dashboard'), href: '/dashboard' },
      { type: 'page', label: t('admin.title') },
    ]);

    setActions([
      <Button key="create" onClick={() => setCreateDialogOpen(true)}>
        <Plus className="w-4 h-4 mr-2" />
        {t('admin.actions.create')}
      </Button>,
    ]);

    return () => {
      setActions([]);
      setItems([]);
    };
  }, [setItems, setActions, t]);

  // Load admins
  const loadAdmins = async () => {
    try {
      setLoading(true);
      const response = await adminApi.getAdmins({
        username: searchUsername || undefined,
        is_active: statusFilter,
        page,
        page_size: pageSize,
      });
      setAdmins(response.admins);
      setTotal(response.total);
    } catch (error) {
      console.error('Failed to load admins:', error);
      toast.error(t('admin.errors.loadFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAdmins();
  }, [page, pageSize, searchUsername, statusFilter]);

  // Handle search
  const handleSearch = (value: string) => {
    setSearchUsername(value);
    setPage(1);
  };

  // Handle status filter change
  const handleStatusFilterChange = (value: string) => {
    if (value === 'all') {
      setStatusFilter(undefined);
    } else {
      setStatusFilter(value === 'active');
    }
    setPage(1);
  };

  // Handle admin actions
  const handleEdit = (admin: Admin) => {
    setSelectedAdmin(admin);
    setUpdateDialogOpen(true);
  };

  const handleChangePassword = (admin: Admin) => {
    setSelectedAdmin(admin);
    setChangePasswordDialogOpen(true);
  };

  const handleDelete = (admin: Admin) => {
    setSelectedAdmin(admin);
    setDeleteDialogOpen(true);
  };

  const handleCreateSuccess = () => {
    setCreateDialogOpen(false);
    loadAdmins();
  };

  const handleUpdateSuccess = () => {
    setUpdateDialogOpen(false);
    setSelectedAdmin(null);
    loadAdmins();
  };

  const handleChangePasswordSuccess = () => {
    setChangePasswordDialogOpen(false);
    setSelectedAdmin(null);
  };

  const handleDeleteSuccess = () => {
    setDeleteDialogOpen(false);
    setSelectedAdmin(null);
    loadAdmins();
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">{t('admin.title')}</h1>
        <p className="text-muted-foreground">{t('admin.description')}</p>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder={t('admin.filters.searchUsername')}
            value={searchUsername}
            onChange={(e) => handleSearch(e.target.value)}
            className="pl-10"
          />
        </div>
        <Select value={statusFilter === undefined ? 'all' : statusFilter ? 'active' : 'inactive'} onValueChange={handleStatusFilterChange}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder={t('admin.filters.status')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('admin.filters.allStatus')}</SelectItem>
            <SelectItem value="active">{t('admin.filters.active')}</SelectItem>
            <SelectItem value="inactive">{t('admin.filters.inactive')}</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Admin Table */}
      <AdminTable
        admins={admins}
        loading={loading}
        total={total}
        page={page}
        pageSize={pageSize}
        onPageChange={setPage}
        onEdit={handleEdit}
        onChangePassword={handleChangePassword}
        onDelete={handleDelete}
      />

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