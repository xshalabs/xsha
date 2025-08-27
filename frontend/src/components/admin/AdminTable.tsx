import { useTranslation } from 'react-i18next';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { DataTablePaginationServer } from '@/components/ui/data-table/data-table-pagination-server';
import { MoreHorizontal, Edit, Key, Trash2, Loader2 } from 'lucide-react';
import { formatDateTime } from '@/lib/utils';
import type { Admin } from '@/lib/api';

interface AdminTableProps {
  admins: Admin[];
  loading: boolean;
  total: number;
  page: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onEdit: (admin: Admin) => void;
  onChangePassword: (admin: Admin) => void;
  onDelete: (admin: Admin) => void;
}

export function AdminTable({
  admins,
  loading,
  total,
  page,
  pageSize,
  onPageChange,
  onEdit,
  onChangePassword,
  onDelete,
}: AdminTableProps) {
  const { t } = useTranslation();

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('admin.table.username')}</TableHead>
              <TableHead>{t('admin.table.email')}</TableHead>
              <TableHead>{t('admin.table.status')}</TableHead>
              <TableHead>{t('admin.table.lastLogin')}</TableHead>
              <TableHead>{t('admin.table.createdBy')}</TableHead>
              <TableHead>{t('admin.table.createdAt')}</TableHead>
              <TableHead className="text-right">{t('common.actions')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {admins.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="h-24 text-center">
                  {t('admin.table.noData')}
                </TableCell>
              </TableRow>
            ) : (
              admins.map((admin) => (
                <TableRow key={admin.id}>
                  <TableCell className="font-medium">{admin.username}</TableCell>
                  <TableCell>{admin.email || '-'}</TableCell>
                  <TableCell>
                    <Badge variant={admin.is_active ? 'default' : 'secondary'}>
                      {admin.is_active ? t('admin.status.active') : t('admin.status.inactive')}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="space-y-1">
                      {admin.last_login_at ? (
                        <>
                          <div className="text-sm">{formatDateTime(admin.last_login_at)}</div>
                          {admin.last_login_ip && (
                            <div className="text-xs text-muted-foreground">IP: {admin.last_login_ip}</div>
                          )}
                        </>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>{admin.created_by}</TableCell>
                  <TableCell>{formatDateTime(admin.created_at)}</TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" className="h-8 w-8 p-0">
                          <span className="sr-only">{t('common.openMenu')}</span>
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => onEdit(admin)}>
                          <Edit className="mr-2 h-4 w-4" />
                          {t('common.edit')}
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => onChangePassword(admin)}>
                          <Key className="mr-2 h-4 w-4" />
                          {t('admin.actions.changePassword')}
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => onDelete(admin)}
                          className="text-destructive"
                        >
                          <Trash2 className="mr-2 h-4 w-4" />
                          {t('common.delete')}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <DataTablePaginationServer
        currentPage={page}
        totalPages={Math.ceil(total / pageSize)}
        total={total}
        onPageChange={onPageChange}
      />
    </div>
  );
}