import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { adminApi, type Admin } from '@/lib/api';
import { toast } from 'sonner';

interface DeleteAdminDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  admin: Admin;
  onSuccess: () => void;
}

export function DeleteAdminDialog({
  open,
  onOpenChange,
  admin,
  onSuccess,
}: DeleteAdminDialogProps) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);

  const handleDelete = async () => {
    try {
      setLoading(true);
      await adminApi.deleteAdmin(admin.id);
      toast.success(t('admin.messages.deleteSuccess'));
      onSuccess();
    } catch (error: any) {
      console.error('Failed to delete admin:', error);
      toast.error(error.message || t('admin.errors.deleteFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!loading) {
      onOpenChange(newOpen);
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t('admin.dialogs.delete.title')}</AlertDialogTitle>
          <AlertDialogDescription className="space-y-2">
            <div>
              {t('admin.dialogs.delete.description', { username: admin.username })}
            </div>
            <div className="text-destructive font-medium">
              {t('admin.dialogs.delete.warning')}
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={loading}>
            {t('common.cancel')}
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={loading}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {loading ? t('common.deleting') : t('common.delete')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}