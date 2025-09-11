import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { toast } from 'sonner';
import { usePageTitle } from '@/hooks/usePageTitle';
import { apiService } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
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
import { FormCardGroup } from '@/components/forms/form-sheet';

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

  return (
    <SectionGroup>
      <Section>
        <SectionHeader>
          <SectionTitle>{t('user.changePasswordPage.title')}</SectionTitle>
          <SectionDescription>
            {t('user.changePasswordPage.description')}
          </SectionDescription>
        </SectionHeader>
        <form onSubmit={form.handleSubmit(handleSubmit)}>
          <FormCardGroup>
            <FormCard>
              <FormCardHeader>
                <FormCardTitle>{t('user.changePasswordPage.formTitle')}</FormCardTitle>
                <FormCardDescription>
                  {t('user.changePasswordPage.formDescription')}
                </FormCardDescription>
              </FormCardHeader>
              <FormCardContent className="space-y-6">
                <Form {...form}>
                  <div className="space-y-2">
                    <Label
                      htmlFor="current_password"
                      className="text-sm font-medium"
                    >
                      {t('user.changePasswordPage.currentPassword')}
                    </Label>
                    <FormField
                      control={form.control}
                      name="current_password"
                      render={({ field }) => (
                        <FormItem>
                          <FormControl>
                            <Input
                              {...field}
                              id="current_password"
                              type="password"
                              placeholder={t('user.changePasswordPage.currentPasswordPlaceholder')}
                              disabled={loading}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>

                  <Separator />

                  <div className="space-y-2">
                    <Label
                      htmlFor="new_password"
                      className="text-sm font-medium"
                    >
                      {t('user.changePasswordPage.newPassword')}
                    </Label>
                    <FormField
                      control={form.control}
                      name="new_password"
                      render={({ field }) => (
                        <FormItem>
                          <FormControl>
                            <Input
                              {...field}
                              id="new_password"
                              type="password"
                              placeholder={t('user.changePasswordPage.newPasswordPlaceholder')}
                              disabled={loading}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>

                  <Separator />

                  <div className="space-y-2">
                    <Label
                      htmlFor="confirm_password"
                      className="text-sm font-medium"
                    >
                      {t('user.changePasswordPage.confirmPassword')}
                    </Label>
                    <FormField
                      control={form.control}
                      name="confirm_password"
                      render={({ field }) => (
                        <FormItem>
                          <FormControl>
                            <Input
                              {...field}
                              id="confirm_password"
                              type="password"
                              placeholder={t('user.changePasswordPage.confirmPasswordPlaceholder')}
                              disabled={loading}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                </Form>
              </FormCardContent>
              <FormCardFooter>
                <Button type="submit" disabled={loading}>
                  {loading ? t('user.changePasswordPage.changing') : t('user.changePasswordPage.submit')}
                </Button>
              </FormCardFooter>
            </FormCard>
          </FormCardGroup>
        </form>
      </Section>
    </SectionGroup>
  );
}