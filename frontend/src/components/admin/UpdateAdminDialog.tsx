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
import { UserAvatar } from '@/components/ui/user-avatar';
import { adminApi, type Admin } from '@/lib/api';
import { toast } from 'sonner';
import { Upload, X } from 'lucide-react';

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
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string>('');
  const [uploadingAvatar, setUploadingAvatar] = useState(false);

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      name: '',
      email: '',
      is_active: true,
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
      });
      setAvatarFile(null);
      setAvatarPreview('');
    }
  }, [admin, form]);

  // Handle avatar file selection
  const handleAvatarSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!file.type.startsWith('image/')) {
      toast.error(t('admin.avatar.unsupportedFormat'));
      return;
    }

    // Validate file size (5MB)
    if (file.size > 5 * 1024 * 1024) {
      toast.error(t('admin.avatar.fileTooLarge'));
      return;
    }

    setAvatarFile(file);
    
    // Create preview URL
    const reader = new FileReader();
    reader.onload = (e) => {
      setAvatarPreview(e.target?.result as string);
    };
    reader.readAsDataURL(file);
  };

  // Handle avatar upload
  const handleAvatarUpload = async () => {
    if (!avatarFile) return;

    try {
      setUploadingAvatar(true);
      const response = await adminApi.uploadAvatar(avatarFile);
      
      // Update admin's avatar
      await adminApi.updateAdminAvatar(admin.id, response.data.id);
      
      toast.success(t('admin.avatar.uploadSuccess'));
      setAvatarFile(null);
      setAvatarPreview('');
      onSuccess(); // Refresh the admin data
    } catch (error: any) {
      console.error('Failed to upload avatar:', error);
      toast.error(error.message || t('admin.avatar.uploadFailed'));
    } finally {
      setUploadingAvatar(false);
    }
  };

  // Clear selected avatar
  const handleClearAvatar = () => {
    setAvatarFile(null);
    setAvatarPreview('');
  };

  const handleSubmit = async (data: FormData) => {
    try {
      setLoading(true);
      await adminApi.updateAdmin(admin.id, {
        username: data.username !== admin.username ? data.username : undefined,
        name: data.name !== admin.name ? data.name : undefined,
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
            {/* Avatar Upload Section */}
            <div className="space-y-4">
              <div className="text-sm font-medium">{t('admin.avatar.title')}</div>
              <div className="flex items-center space-x-4">
                <UserAvatar 
                  user={admin.username}
                  name={admin.name}
                  avatar={admin.avatar}
                  size="lg"
                />
                <div className="flex-1 space-y-2">
                  {!avatarFile ? (
                    <>
                      <input
                        type="file"
                        id="avatar-upload"
                        accept="image/*"
                        onChange={handleAvatarSelect}
                        className="hidden"
                      />
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => document.getElementById('avatar-upload')?.click()}
                        disabled={loading || uploadingAvatar}
                      >
                        <Upload className="h-4 w-4 mr-2" />
                        {t('admin.avatar.upload')}
                      </Button>
                    </>
                  ) : (
                    <div className="space-y-2">
                      <div className="flex items-center space-x-2">
                        <img 
                          src={avatarPreview} 
                          alt="Preview" 
                          className="h-12 w-12 rounded-lg object-cover"
                        />
                        <div className="flex-1">
                          <p className="text-sm text-muted-foreground">
                            {avatarFile.name}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            {(avatarFile.size / 1024 / 1024).toFixed(2)} MB
                          </p>
                        </div>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={handleClearAvatar}
                          disabled={uploadingAvatar}
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      </div>
                      <Button
                        type="button"
                        size="sm"
                        onClick={handleAvatarUpload}
                        disabled={uploadingAvatar}
                      >
                        {uploadingAvatar ? t('admin.avatar.uploading') : t('admin.avatar.confirm')}
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            </div>

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