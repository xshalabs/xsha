import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { toast } from 'sonner';
import { ArrowLeft, Lock } from 'lucide-react';
import { usePageTitle } from '@/hooks/usePageTitle';
import { apiService } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import {
  Section,
  SectionDescription,
  SectionGroup,
  SectionHeader,
  SectionTitle,
} from '@/components/content/section';
import {
  FormCard,
  FormCardContent,
  FormCardDescription,
  FormCardFooter,
  FormCardHeader,
  FormCardTitle,
} from '@/components/forms/form-card';

const formSchema = z.object({
  current_password: z
    .string()
    .min(1, 'Current password is required'),
  new_password: z
    .string()
    .min(6, 'Password must be at least 6 characters')
    .max(128, 'Password must be at most 128 characters'),
  confirm_password: z.string(),
}).refine((data) => data.new_password === data.confirm_password, {
  message: "Passwords don't match",
  path: ["confirm_password"],
});

type FormData = z.infer<typeof formSchema>;

export default function ChangePasswordPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  
  usePageTitle(t('user.changePasswordPage.title'));

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      current_password: '',
      new_password: '',
      confirm_password: '',
    },
  });

  const handleSubmit = async (data: FormData) => {
    try {
      setLoading(true);
      await apiService.changeOwnPassword({
        current_password: data.current_password,
        new_password: data.new_password,
      });
      toast.success(t('user.changePasswordPage.success'));
      navigate(-1); // Go back to previous page
    } catch (error) {
      console.error('Failed to change password:', error);
      const errorMessage = error instanceof Error ? error.message : String(error);
      toast.error(errorMessage || t('user.changePasswordPage.error'));
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    navigate(-1);
  };

  return (
    <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
      <Section>
        <SectionHeader>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleCancel}
              className="h-8 w-8 p-0"
            >
              <ArrowLeft className="h-4 w-4" />
            </Button>
            <div>
              <SectionTitle className="flex items-center gap-2">
                <Lock className="h-5 w-5" />
                {t('user.changePasswordPage.title')}
              </SectionTitle>
              <SectionDescription>
                {t('user.changePasswordPage.description')}
              </SectionDescription>
            </div>
          </div>
        </SectionHeader>

        <SectionGroup>
          <FormCard>
            <FormCardHeader>
              <FormCardTitle>{t('user.changePasswordPage.formTitle')}</FormCardTitle>
              <FormCardDescription>
                {t('user.changePasswordPage.formDescription')}
              </FormCardDescription>
            </FormCardHeader>

            <Form {...form}>
              <form onSubmit={form.handleSubmit(handleSubmit)}>
                <FormCardContent className="space-y-4">
                  <FormField
                    control={form.control}
                    name="current_password"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t('user.changePasswordPage.currentPassword')}</FormLabel>
                        <FormControl>
                          <Input
                            {...field}
                            type="password"
                            placeholder={t('user.changePasswordPage.currentPasswordPlaceholder')}
                            disabled={loading}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="new_password"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t('user.changePasswordPage.newPassword')}</FormLabel>
                        <FormControl>
                          <Input
                            {...field}
                            type="password"
                            placeholder={t('user.changePasswordPage.newPasswordPlaceholder')}
                            disabled={loading}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="confirm_password"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t('user.changePasswordPage.confirmPassword')}</FormLabel>
                        <FormControl>
                          <Input
                            {...field}
                            type="password"
                            placeholder={t('user.changePasswordPage.confirmPasswordPlaceholder')}
                            disabled={loading}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </FormCardContent>

                <FormCardFooter>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleCancel}
                    disabled={loading}
                  >
                    {t('common.cancel')}
                  </Button>
                  <Button type="submit" disabled={loading}>
                    {loading ? t('user.changePasswordPage.changing') : t('user.changePasswordPage.submit')}
                  </Button>
                </FormCardFooter>
              </form>
            </Form>
          </FormCard>
        </SectionGroup>
      </Section>
    </div>
  );
}