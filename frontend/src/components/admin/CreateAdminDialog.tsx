import { useState } from 'react';
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { adminApi, type AdminRole } from '@/lib/api';
import { usePermissions } from '@/hooks/usePermissions';
import { toast } from 'sonner';

const formSchema = z.object({
  username: z
    .string()
    .min(3, 'Username must be at least 3 characters')
    .max(50, 'Username must be at most 50 characters')
    .regex(/^[a-zA-Z0-9_]+$/, 'Username can only contain letters, numbers and underscores'),
  password: z
    .string()
    .min(6, 'Password must be at least 6 characters')
    .max(128, 'Password must be at most 128 characters'),
  name: z
    .string()
    .min(2, 'Name must be at least 2 characters')
    .max(100, 'Name must be at most 100 characters'),
  email: z.string().email('Invalid email address').optional().or(z.literal('')),
  role: z.enum(['super_admin', 'admin', 'developer']).optional(),
  lang: z.enum(['en-US', 'zh-CN']).optional(),
});

type FormData = z.infer<typeof formSchema>;

interface CreateAdminDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function CreateAdminDialog({
  open,
  onOpenChange,
  onSuccess,
}: CreateAdminDialogProps) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const permissions = usePermissions();

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      password: '',
      name: '',
      email: '',
      role: 'admin',
      lang: 'en-US',
    },
  });

  const handleSubmit = async (data: FormData) => {
    try {
      setLoading(true);
      await adminApi.createAdmin({
        username: data.username,
        password: data.password,
        name: data.name,
        email: data.email || undefined,
        role: data.role as AdminRole,
        lang: data.lang,
      });
      toast.success(t('admin.messages.createSuccess'));
      form.reset();
      onSuccess();
    } catch (error: any) {
      console.error('Failed to create admin:', error);
      toast.error(error.message || t('admin.errors.createFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!loading) {
      onOpenChange(newOpen);
      if (!newOpen) {
        form.reset();
      }
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{t('admin.dialogs.create.title')}</DialogTitle>
          <DialogDescription>
            {t('admin.dialogs.create.description')}
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
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('admin.fields.password')}</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      type="password"
                      placeholder={t('admin.placeholders.password')}
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

            {permissions.canManageAdminRole() && 
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
              name="lang"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('admin.fields.language')} ({t('common.optional')})</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger disabled={loading}>
                        <SelectValue placeholder={t('admin.placeholders.language')} />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="en-US">
                        {t('admin.languages.en-US')}
                      </SelectItem>
                      <SelectItem value="zh-CN">
                        {t('admin.languages.zh-CN')}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
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
                {loading ? t('common.creating') : t('common.create')}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}