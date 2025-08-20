import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Download, Eye, Trash2, FileText, Image, AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { attachmentApi, type Attachment } from '@/lib/api/attachments';
import { AttachmentPreviewModal } from './AttachmentPreviewModal';

interface AttachmentListProps {
  attachments: Attachment[];
  onAttachmentDelete?: (attachmentId: number) => void;
  readonly?: boolean;
  className?: string;
}

export function AttachmentList({
  attachments,
  onAttachmentDelete,
  readonly = false,
  className
}: AttachmentListProps) {
  const { t } = useTranslation();
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [previewAttachment, setPreviewAttachment] = useState<Attachment | null>(null);

  const handleDownload = async (attachment: Attachment) => {
    try {
      await attachmentApi.downloadAttachment(attachment.id, attachment.original_name);
    } catch (error) {
      console.error('Download failed:', error);
    }
  };

  const handlePreview = (attachment: Attachment) => {
    if (attachment.type === 'image') {
      setPreviewAttachment(attachment);
    } else {
      // For PDF files, open in new tab
      const previewUrl = attachmentApi.getPreviewUrl(attachment.id);
      window.open(previewUrl, '_blank');
    }
  };

  const handleDelete = async (attachmentId: number) => {
    setDeletingId(attachmentId);
    try {
      await attachmentApi.deleteAttachment(attachmentId);
      onAttachmentDelete?.(attachmentId);
    } catch (error) {
      console.error('Delete failed:', error);
    } finally {
      setDeletingId(null);
    }
  };

  const getFileIcon = (type: string) => {
    if (type === 'image') {
      return <Image className="h-4 w-4" />;
    } else if (type === 'pdf') {
      return <FileText className="h-4 w-4" />;
    }
    return <FileText className="h-4 w-4" />;
  };

  const getFileTypeLabel = (type: string) => {
    switch (type) {
      case 'image':
        return t('attachment.type_image', 'Image');
      case 'pdf':
        return t('attachment.type_pdf', 'PDF');
      default:
        return t('attachment.type_file', 'File');
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getDisplayName = (attachment: Attachment, index: number) => {
    const typeCount = attachments.filter(a => a.type === attachment.type).indexOf(attachment) + 1;
    return attachment.type === 'image' ? `[image${typeCount}]` : `[pdf${typeCount}]`;
  };

  if (attachments.length === 0) {
    return null;
  }

  return (
    <>
      <div className={cn('space-y-3', className)}>
        {/* Compact attachment cards in a flexible layout */}
        <div className="flex flex-wrap gap-2">
          {attachments.map((attachment, index) => (
            <div
              key={attachment.id}
              className="relative group inline-flex items-center gap-2 px-3 py-2 border border-orange-200 bg-orange-50 hover:bg-orange-100 rounded-lg transition-colors max-w-[240px] cursor-pointer"
              onClick={() => handlePreview(attachment)}
            >
              <div className="flex-shrink-0 text-orange-600">
                {getFileIcon(attachment.type)}
              </div>
              
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-1">
                  <span className="text-sm font-medium text-gray-900 truncate">
                    {getDisplayName(attachment, index)}
                  </span>
                </div>
                <div className="text-xs text-gray-600">
                  {getFileTypeLabel(attachment.type).toUpperCase()} · {formatFileSize(attachment.file_size)}
                </div>
              </div>

              {/* Action buttons - visible on hover */}
              {!readonly && (
                <div className="opacity-0 group-hover:opacity-100 flex items-center gap-1 transition-opacity">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDownload(attachment);
                    }}
                    className="h-6 w-6 p-0 hover:bg-orange-200"
                    title={t('attachment.download', 'Download')}
                  >
                    <Download className="h-3 w-3" />
                  </Button>
                  
                  <AlertDialog>
                    <AlertDialogTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => e.stopPropagation()}
                        className="h-6 w-6 p-0 text-destructive hover:text-destructive hover:bg-orange-200"
                        title={t('attachment.delete', 'Delete')}
                        disabled={deletingId === attachment.id}
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>
                          {t('attachment.delete_confirm_title', 'Delete Attachment')}
                        </AlertDialogTitle>
                        <AlertDialogDescription>
                          {t('attachment.delete_confirm_message', 
                            'Are you sure you want to delete this attachment? This action cannot be undone.'
                          )}
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>
                          {t('common.cancel', 'Cancel')}
                        </AlertDialogCancel>
                        <AlertDialogAction
                          onClick={() => handleDelete(attachment.id)}
                          className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        >
                          {t('common.delete', 'Delete')}
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Preview Modal */}
      <AttachmentPreviewModal
        attachment={previewAttachment}
        open={!!previewAttachment}
        onOpenChange={(open) => !open && setPreviewAttachment(null)}
      />
    </>
  );
}
