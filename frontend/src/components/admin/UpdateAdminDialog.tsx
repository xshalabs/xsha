import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { adminApi, type Admin, type AdminRole } from '@/lib/api';
import { usePermissions } from '@/hooks/usePermissions';
import { toast } from 'sonner';

const formSchema = z.object({
  username: z
    .string()
    .min(3, 'Username must be at least 3 characters')
    .max(50, 'Username must be at most 50 characters')
    .regex(/^[a-zA-Z0-9_]+$/, 'Username can only contain letters, numbers and underscores'),
  name: z
    .string()
    .min(2, 'Name must be at least 2 characters')
    .max(100, 'Name must be at most 100 characters'),
  email: z.string().email('Invalid email address').optional().or(z.literal('')),
  is_active: z.boolean(),
  role: z.enum(['super_admin', 'admin', 'developer']).optional(),
});

type FormData = z.infer<typeof formSchema>;

interface UpdateAdminDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  admin: Admin;
  onSuccess: () => void;
}

export function UpdateAdminDialog({
  open,
  onOpenChange,
  admin,
  onSuccess,
}: UpdateAdminDialogProps) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const permissions = usePermissions();

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      name: '',
      email: '',
      is_active: true,
      role: 'admin',
    },
  });

  // Reset form when admin changes
  useEffect(() => {
    if (admin) {
      form.reset({
        username: admin.username,
        name: admin.name,
        email: admin.email || '',
        is_active: admin.is_active,
        role: admin.role,
      });
    }
  }, [admin, form]);


  const handleSubmit = async (data: FormData) => {
    try {
      setLoading(true);
      await adminApi.updateAdmin(admin.id, {
        username: data.username !== admin.username ? data.username : undefined,
        name: data.name !== admin.name ? data.name : undefined,
        email: data.email !== admin.email ? data.email : undefined,
        is_active: data.is_active !== admin.is_active ? data.is_active : undefined,
        role: admin.created_by !== 'system' && data.role !== admin.role ? data.role as AdminRole : undefined,
      });
      toast.success(t('admin.messages.updateSuccess'));
      onSuccess();
    } catch (error: any) {
      console.error('Failed to update admin:', error);
      toast.error(error.message || t('admin.errors.updateFailed'));
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
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{t('admin.dialogs.update.title')}</DialogTitle>
          <DialogDescription>
            {t('admin.dialogs.update.description', { username: admin.username })}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="username"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('admin.fields.username')}</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      placeholder={t('admin.placeholders.username')}
                      disabled={loading}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('admin.fields.name')}</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      placeholder={t('admin.placeholders.name')}
                      disabled={loading}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('admin.fields.email')} ({t('common.optional')})</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      type="email"
                      placeholder={t('admin.placeholders.email')}
                      disabled={loading}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {permissions.canManageAdminRole() && admin.created_by !== 'system' &&
              <FormField
                control={form.control}
                name="role"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('admin.fields.role')}</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger disabled={loading}>
                          <SelectValue placeholder={t('admin.placeholders.role')} />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="admin">
                          {t('admin.roles.admin')}
                        </SelectItem>
                        <SelectItem value="developer">
                          {t('admin.roles.developer')}
                        </SelectItem>
                        <SelectItem value="super_admin">
                          {t('admin.roles.super_admin')}
                        </SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            }

            <FormField
              control={form.control}
              name="is_active"
              render={({ field }) => (
                <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                  <div className="space-y-0.5">
                    <FormLabel className="text-base">
                      {t('admin.fields.status')}
                    </FormLabel>
                    <div className="text-sm text-muted-foreground">
                      {t('admin.fields.statusDescription')}
                    </div>
                  </div>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                      disabled={loading}
                    />
                  </FormControl>
                </FormItem>
              )}
            />

            <div className="flex justify-end space-x-2 pt-4">
              <Button
                type="button"
                variant="outline"
                onClick={() => handleOpenChange(false)}
                disabled={loading}
              >
                {t('common.cancel')}
              </Button>
              <Button type="submit" disabled={loading}>
                {loading ? t('common.updating') : t('common.update')}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}