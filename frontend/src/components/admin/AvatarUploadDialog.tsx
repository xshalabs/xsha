import { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { UserAvatar } from '@/components/ui/user-avatar';
import { adminApi, type Admin } from '@/lib/api';
import { toast } from 'sonner';
import { Upload, X } from 'lucide-react';
import { cn } from '@/lib/utils';

interface AvatarUploadDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  admin: Admin;
  onSuccess: () => void;
}

export function AvatarUploadDialog({
  open,
  onOpenChange,
  admin,
  onSuccess,
}: AvatarUploadDialogProps) {
  const { t } = useTranslation();
  const [uploading, setUploading] = useState(false);
  const [dragActive, setDragActive] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string>('');

  const handleFileSelect = useCallback((file: File) => {
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

    setSelectedFile(file);
    
    // Create preview URL
    const reader = new FileReader();
    reader.onload = (e) => {
      setPreviewUrl(e.target?.result as string);
    };
    reader.readAsDataURL(file);
  }, [t]);

  const handleFileInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    const file = e.dataTransfer.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  }, [handleFileSelect]);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
  }, []);

  const handleUpload = async () => {
    if (!selectedFile) return;

    try {
      setUploading(true);
      const response = await adminApi.uploadAvatar(selectedFile);
      
      // Update admin's avatar using avatar UUID
      await adminApi.updateAdminAvatar(response.data.uuid, admin.id);
      
      toast.success(t('admin.avatar.uploadSuccess'));
      clearSelection();
      onSuccess();
      onOpenChange(false);
    } catch (error) {
      console.error('Failed to upload avatar:', error);
      toast.error((error as Error).message || t('admin.avatar.uploadFailed'));
    } finally {
      setUploading(false);
    }
  };


  const clearSelection = () => {
    setSelectedFile(null);
    setPreviewUrl('');
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!uploading) {
      if (!newOpen) {
        clearSelection();
      }
      onOpenChange(newOpen);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('admin.avatar.change')}</DialogTitle>
          <DialogDescription>
            Update the avatar for {admin.name} ({admin.username})
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Current Avatar */}
          <div className="flex flex-col items-center space-y-3">
            <div className="text-sm font-medium">Current Avatar</div>
            <UserAvatar 
              user={admin.username}
              name={admin.name}
              avatar={admin.avatar}
              size="lg"
            />
          </div>

          {/* Upload Area */}
          <div className="space-y-4">
            <div className="text-sm font-medium">Upload New Avatar</div>
            
            {!selectedFile ? (
              <div
                className={cn(
                  "border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors",
                  dragActive 
                    ? "border-primary bg-primary/5" 
                    : "border-muted-foreground/25 hover:border-muted-foreground/50"
                )}
                onDrop={handleDrop}
                onDragOver={handleDragOver}
                onDragEnter={handleDragEnter}
                onDragLeave={handleDragLeave}
                onClick={() => document.getElementById('avatar-file-input')?.click()}
              >
                <Upload className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
                <p className="text-sm text-muted-foreground mb-2">
                  Drag and drop an image here, or click to select
                </p>
                <p className="text-xs text-muted-foreground">
                  Supports: JPG, PNG, GIF, WebP (max 5MB)
                </p>
                <input
                  id="avatar-file-input"
                  type="file"
                  accept="image/*"
                  onChange={handleFileInputChange}
                  className="hidden"
                />
              </div>
            ) : (
              <div className="space-y-4">
                <div className="flex items-center justify-center">
                  <img 
                    src={previewUrl} 
                    alt="Preview" 
                    className="h-24 w-24 rounded-lg object-cover border"
                  />
                </div>
                <div className="text-center space-y-1">
                  <p className="text-sm font-medium">{selectedFile.name}</p>
                  <p className="text-xs text-muted-foreground">
                    {(selectedFile.size / 1024 / 1024).toFixed(2)} MB
                  </p>
                </div>
                <div className="flex space-x-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={clearSelection}
                    disabled={uploading}
                    className="flex-1"
                  >
                    <X className="h-4 w-4 mr-2" />
                    Clear
                  </Button>
                  <Button
                    variant="outline" 
                    size="sm"
                    onClick={() => document.getElementById('avatar-file-input')?.click()}
                    disabled={uploading}
                    className="flex-1"
                  >
                    <Upload className="h-4 w-4 mr-2" />
                    Choose Different
                  </Button>
                </div>
                <input
                  id="avatar-file-input"
                  type="file"
                  accept="image/*"
                  onChange={handleFileInputChange}
                  className="hidden"
                />
              </div>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={uploading}
          >
            Cancel
          </Button>
          <Button
            onClick={handleUpload}
            disabled={!selectedFile || uploading}
          >
            {uploading ? (
              <>
                <Upload className="h-4 w-4 mr-2 animate-spin" />
                {t('admin.avatar.uploading')}
              </>
            ) : (
              <>
                <Upload className="h-4 w-4 mr-2" />
                {t('admin.avatar.confirm')}
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}