import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Download, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { attachmentApi, type Attachment } from '@/lib/api/attachments';

interface AttachmentPreviewModalProps {
  attachment: Attachment | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  displayName?: string;
}

export function AttachmentPreviewModal({
  attachment,
  open,
  onOpenChange,
  displayName
}: AttachmentPreviewModalProps) {
  const { t } = useTranslation();
  const [imageLoading, setImageLoading] = useState(true);
  const [imageError, setImageError] = useState(false);
  const [previewUrl, setPreviewUrl] = useState<string>('');

  // Load preview URL when attachment changes
  useEffect(() => {
    let mounted = true;

    const loadPreview = async () => {
      if (!attachment || !open || attachment.type !== 'image') {
        return;
      }

      setImageLoading(true);
      setImageError(false);
      setPreviewUrl('');

      try {
        const blobUrl = await attachmentApi.getPreviewBlob(attachment.id, attachment.project_id);
        if (mounted) {
          setPreviewUrl(blobUrl);
        }
      } catch (error) {
        console.error('Failed to load preview:', error);
        if (mounted) {
          setImageError(true);
        }
      } finally {
        if (mounted) {
          setImageLoading(false);
        }
      }
    };

    loadPreview();

    return () => {
      mounted = false;
      // Clean up blob URL to prevent memory leaks
      if (previewUrl) {
        window.URL.revokeObjectURL(previewUrl);
      }
    };
  }, [attachment?.id, open]);

  // Clean up blob URL when component unmounts or attachment changes
  useEffect(() => {
    return () => {
      if (previewUrl) {
        window.URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

  const handleDownload = async () => {
    if (!attachment) return;
    
    try {
      await attachmentApi.downloadAttachment(attachment.id, attachment.original_name, attachment.project_id);
    } catch (error) {
      console.error('Download failed:', error);
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getDisplayName = (attachment: Attachment) => {
    // Use provided displayName or fall back to original name
    return displayName || attachment.original_name;
  };

  if (!attachment) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl w-full max-h-[90vh] overflow-hidden">
        <DialogHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <DialogTitle className="text-lg font-semibold truncate pr-8">
            {getDisplayName(attachment)}
          </DialogTitle>
          
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleDownload}
              className="flex items-center space-x-2"
            >
              <Download className="h-4 w-4" />
              <span>{t('attachment.download', 'Download')}</span>
            </Button>
          </div>
        </DialogHeader>
        
        <div className="space-y-4">
          {/* Image Preview */}
          {attachment.type === 'image' && (
            <div className="relative bg-muted/30 rounded-lg overflow-hidden">
              {imageLoading && !imageError && (
                <div className="flex items-center justify-center h-96">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                </div>
              )}
              
              {imageError && (
                <div className="flex flex-col items-center justify-center h-96 text-muted-foreground">
                  <X className="h-12 w-12 mb-2" />
                  <p>{t('attachment.image_load_error', 'Failed to load image')}</p>
                </div>
              )}
              
              {previewUrl && (
                <img
                  src={previewUrl}
                  alt={attachment.original_name}
                  className="w-full h-auto max-h-[70vh] object-contain transition-opacity duration-200"
                  onLoad={() => {
                    setImageLoading(false);
                    setImageError(false);
                  }}
                  onError={() => {
                    setImageLoading(false);
                    setImageError(true);
                  }}
                  style={{ display: imageError ? 'none' : 'block' }}
                />
              )}
            </div>
          )}
          
          {/* File Info */}
          <div className="flex justify-between items-center text-sm text-muted-foreground border-t pt-4">
            <div className="space-y-1">
              <p>
                <span className="font-medium">{t('attachment.file_size', 'Size')}: </span>
                {formatFileSize(attachment.file_size)}
              </p>
              <p>
                <span className="font-medium">{t('attachment.file_type', 'Type')}: </span>
                {attachment.content_type}
              </p>
            </div>
            <div className="text-right space-y-1">
              <p>
                <span className="font-medium">{t('attachment.uploaded_by', 'Uploaded by')}: </span>
                {attachment.created_by}
              </p>
              <p>
                <span className="font-medium">{t('attachment.uploaded_at', 'Uploaded')}: </span>
                {new Date(attachment.created_at).toLocaleString()}
              </p>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
