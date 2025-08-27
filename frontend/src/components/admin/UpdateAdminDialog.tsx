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
import { adminApi, type Admin } from '@/lib/api';
import { toast } from 'sonner';

const formSchema = z.object({
  username: z
    .string()
    .min(3, 'Username must be at least 3 characters')
    .max(50, 'Username must be at most 50 characters')
    .regex(/^[a-zA-Z0-9_]+$/, 'Username can only contain letters, numbers and underscores'),
  email: z.string().email('Invalid email address').optional().or(z.literal('')),
  is_active: z.boolean(),
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

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      email: '',
      is_active: true,
    },
  });

  // Reset form when admin changes
  useEffect(() => {
    if (admin) {
      form.reset({
        username: admin.username,
        email: admin.email || '',
        is_active: admin.is_active,
      });
    }
  }, [admin, form]);

  const handleSubmit = async (data: FormData) => {
    try {
      setLoading(true);
      await adminApi.updateAdmin(admin.id, {
        username: data.username !== admin.username ? data.username : undefined,
        email: data.email !== admin.email ? data.email : undefined,
        is_active: data.is_active !== admin.is_active ? data.is_active : undefined,
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