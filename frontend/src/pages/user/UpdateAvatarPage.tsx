import { useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Camera, Upload } from 'lucide-react';
import { usePageTitle } from '@/hooks/usePageTitle';
import { apiService, adminApi, type AvatarUploadResponse } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { UserAvatar } from '@/components/ui/user-avatar';
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
import { useAuth } from '@/contexts/AuthContext';

export default function UpdateAvatarPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const { user, name, avatar, checkAuth } = useAuth();
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [newAvatar, setNewAvatar] = useState<AvatarUploadResponse | null>(null);
  
  usePageTitle(t('user.updateAvatarPage.title'));

  const handleFileSelect = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!file.type.startsWith('image/')) {
      toast.error(t('user.updateAvatarPage.invalidFileType'));
      return;
    }

    // Validate file size (5MB max)
    const maxSize = 5 * 1024 * 1024;
    if (file.size > maxSize) {
      toast.error(t('user.updateAvatarPage.fileTooLarge'));
      return;
    }

    try {
      setUploading(true);
      const uploadResponse = await adminApi.uploadAvatar(file);
      setNewAvatar(uploadResponse);
      toast.success(t('user.updateAvatarPage.uploadSuccess'));
    } catch (error) {
      console.error('Failed to upload avatar:', error);
      const errorMessage = error instanceof Error ? error.message : String(error);
      toast.error(errorMessage || t('user.updateAvatarPage.uploadError'));
    } finally {
      setUploading(false);
    }
  };

  const handleUpdateAvatar = async () => {
    if (!newAvatar) {
      toast.error(t('user.updateAvatarPage.noAvatarSelected'));
      return;
    }

    try {
      setLoading(true);
      await apiService.updateOwnAvatar({
        avatar_uuid: newAvatar.data.uuid,
      });
      toast.success(t('user.updateAvatarPage.updateSuccess'));
      
      // Refresh user data to get the updated avatar
      await checkAuth();
      
      // Clear the new avatar state
      setNewAvatar(null);
      
      navigate(-1); // Go back to previous page
    } catch (error) {
      console.error('Failed to update avatar:', error);
      const errorMessage = error instanceof Error ? error.message : String(error);
      toast.error(errorMessage || t('user.updateAvatarPage.updateError'));
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    setNewAvatar(null);
    navigate(-1);
  };

  return (
    <SectionGroup>
      <Section>
        <SectionHeader>
          <SectionTitle>{t('user.updateAvatarPage.title')}</SectionTitle>
          <SectionDescription>
            {t('user.updateAvatarPage.description')}
          </SectionDescription>
        </SectionHeader>
        <FormCardGroup>
          <FormCard>
            <FormCardHeader>
              <FormCardTitle>{t('user.updateAvatarPage.formTitle')}</FormCardTitle>
              <FormCardDescription>
                {t('user.updateAvatarPage.formDescription')}
              </FormCardDescription>
            </FormCardHeader>
            <FormCardContent className="space-y-6">
              <div className="space-y-4">
                <Label className="text-sm font-medium">
                  {t('user.updateAvatarPage.currentAvatar')}
                </Label>
                <div className="flex items-center space-x-4">
                  <UserAvatar 
                    user={user || undefined}
                    name={name || undefined}
                    avatar={avatar || undefined}
                    size="lg"
                  />
                  <div className="flex flex-col">
                    <span className="text-sm font-medium">{user}</span>
                    <span className="text-xs text-muted-foreground">{name || "Admin"}</span>
                  </div>
                </div>
              </div>

              <Separator />

              <div className="space-y-4">
                <Label className="text-sm font-medium">
                  {t('user.updateAvatarPage.newAvatar')}
                </Label>
                
                {newAvatar ? (
                  <div className="space-y-4">
                    <div className="flex items-center space-x-4">
                      <img
                        src={newAvatar.data.preview_url}
                        alt={t('user.updateAvatarPage.newAvatarPreview')}
                        className="h-12 w-12 rounded-lg object-cover"
                      />
                      <div className="flex flex-col">
                        <span className="text-sm font-medium">{newAvatar.data.original_name}</span>
                        <span className="text-xs text-muted-foreground">
                          {(newAvatar.data.file_size / 1024).toFixed(1)} KB
                        </span>
                      </div>
                    </div>
                    <Button
                      type="button"
                      variant="outline"
                      onClick={handleFileSelect}
                      disabled={uploading}
                      className="w-full"
                    >
                      <Camera className="mr-2 h-4 w-4" />
                      {t('user.updateAvatarPage.chooseAnother')}
                    </Button>
                  </div>
                ) : (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleFileSelect}
                    disabled={uploading}
                    className="w-full h-32 border-dashed"
                  >
                    {uploading ? (
                      <div className="flex flex-col items-center">
                        <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-current"></div>
                        <span className="mt-2">{t('user.updateAvatarPage.uploading')}</span>
                      </div>
                    ) : (
                      <div className="flex flex-col items-center">
                        <Upload className="h-8 w-8 mb-2" />
                        <span>{t('user.updateAvatarPage.selectFile')}</span>
                        <span className="text-xs text-muted-foreground mt-1">
                          {t('user.updateAvatarPage.fileHint')}
                        </span>
                      </div>
                    )}
                  </Button>
                )}

                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  onChange={handleFileChange}
                  className="hidden"
                />
              </div>
            </FormCardContent>
            <FormCardFooter className="flex justify-end space-x-2">
              <Button
                type="button"
                variant="outline"
                onClick={handleCancel}
                disabled={loading || uploading}
              >
                {t('common.cancel')}
              </Button>
              <Button 
                onClick={handleUpdateAvatar} 
                disabled={loading || uploading || !newAvatar}
              >
                {loading ? t('user.updateAvatarPage.updating') : t('user.updateAvatarPage.update')}
              </Button>
            </FormCardFooter>
          </FormCard>
        </FormCardGroup>
      </Section>
    </SectionGroup>
  );
}