import { useState, useCallback, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { Upload, X, FileText, Image, AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { cn } from '@/lib/utils';
import { attachmentApi, type Attachment } from '@/lib/api/attachments';

interface AttachmentUploaderProps {
  conversationId: number | null;
  onUploadSuccess?: (attachment: Attachment) => void;
  onUploadError?: (error: string) => void;
  disabled?: boolean;
  className?: string;
}

interface UploadFile {
  id: string;
  file: File;
  progress: number;
  status: 'pending' | 'uploading' | 'success' | 'error';
  error?: string;
  attachment?: Attachment;
}

const ACCEPTED_TYPES = {
  'image/*': ['.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp'],
  'application/pdf': ['.pdf']
};

const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

export function AttachmentUploader({
  conversationId,
  onUploadSuccess,
  onUploadError,
  disabled = false,
  className
}: AttachmentUploaderProps) {
  const { t } = useTranslation();
  const [uploadFiles, setUploadFiles] = useState<UploadFile[]>([]);
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const validateFile = (file: File): string | null => {
    // Check file size
    if (file.size > MAX_FILE_SIZE) {
      return t('attachment.file_too_large', 'File is too large (max 10MB)');
    }

    // Check file type
    const isImage = file.type.startsWith('image/');
    const isPdf = file.type === 'application/pdf';
    
    if (!isImage && !isPdf) {
      return t('attachment.unsupported_file_type', 'Only images and PDF files are supported');
    }

    return null;
  };

  const generateUploadId = () => Math.random().toString(36).substr(2, 9);

  const addFiles = useCallback((files: FileList | File[]) => {
    if (!conversationId) return;

    const fileArray = Array.from(files);
    const newUploadFiles: UploadFile[] = [];

    for (const file of fileArray) {
      const error = validateFile(file);
      if (error) {
        onUploadError?.(error);
        continue;
      }

      newUploadFiles.push({
        id: generateUploadId(),
        file,
        progress: 0,
        status: 'pending',
      });
    }

    setUploadFiles(prev => [...prev, ...newUploadFiles]);

    // Start uploading immediately
    newUploadFiles.forEach(uploadFile => {
      uploadFileAsync(uploadFile);
    });
  }, [conversationId, onUploadError]);

  const uploadFileAsync = async (uploadFile: UploadFile) => {
    if (!conversationId) return;

    setUploadFiles(prev =>
      prev.map(f => f.id === uploadFile.id ? { ...f, status: 'uploading' as const } : f)
    );

    try {
      const attachment = await attachmentApi.uploadAttachment(conversationId, uploadFile.file);
      
      setUploadFiles(prev =>
        prev.map(f => f.id === uploadFile.id ? { 
          ...f, 
          status: 'success' as const, 
          progress: 100,
          attachment 
        } : f)
      );

      onUploadSuccess?.(attachment);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Upload failed';
      
      setUploadFiles(prev =>
        prev.map(f => f.id === uploadFile.id ? { 
          ...f, 
          status: 'error' as const, 
          error: errorMessage 
        } : f)
      );

      onUploadError?.(errorMessage);
    }
  };

  const removeUploadFile = useCallback((uploadId: string) => {
    setUploadFiles(prev => prev.filter(f => f.id !== uploadId));
  }, []);

  const handleFileInputChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (files && files.length > 0) {
      addFiles(files);
    }
    // Reset input value to allow selecting the same file again
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  }, [addFiles]);

  const handleDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.stopPropagation();
    setIsDragOver(true);
  }, []);

  const handleDragLeave = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.stopPropagation();
    setIsDragOver(false);
  }, []);

  const handleDrop = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.stopPropagation();
    setIsDragOver(false);

    const files = event.dataTransfer.files;
    if (files && files.length > 0) {
      addFiles(files);
    }
  }, [addFiles]);

  const handleUploadClick = () => {
    fileInputRef.current?.click();
  };

  const getFileIcon = (file: File) => {
    if (file.type.startsWith('image/')) {
      return <Image className="h-4 w-4" />;
    } else if (file.type === 'application/pdf') {
      return <FileText className="h-4 w-4" />;
    }
    return <FileText className="h-4 w-4" />;
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (!conversationId) {
    return (
      <Alert>
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          {t('attachment.conversation_required', 'Please create a conversation first to upload attachments')}
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className={cn('space-y-4', className)}>
      {/* Upload Area */}
      <div
        className={cn(
          'border-2 border-dashed rounded-lg p-6 transition-colors',
          isDragOver ? 'border-primary bg-primary/5' : 'border-muted-foreground/25',
          disabled && 'opacity-50 pointer-events-none'
        )}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        <div className="flex flex-col items-center justify-center space-y-2">
          <Upload className="h-8 w-8 text-muted-foreground" />
          <div className="text-center">
            <Button
              type="button"
              variant="ghost"
              onClick={handleUploadClick}
              disabled={disabled}
              className="text-primary hover:text-primary/80"
            >
              {t('attachment.click_to_upload', 'Click to upload')}
            </Button>
            <span className="text-sm text-muted-foreground">
              {t('attachment.or_drag_and_drop', ' or drag and drop')}
            </span>
          </div>
          <p className="text-xs text-muted-foreground text-center">
            {t('attachment.supported_formats', 'Images (JPG, PNG, GIF, WebP) and PDF files up to 10MB')}
          </p>
        </div>
      </div>

      {/* File Input */}
      <input
        ref={fileInputRef}
        type="file"
        accept={Object.keys(ACCEPTED_TYPES).join(',')}
        multiple
        onChange={handleFileInputChange}
        className="hidden"
      />

      {/* Upload Progress */}
      {uploadFiles.length > 0 && (
        <div className="space-y-2">
          {uploadFiles.map((uploadFile) => (
            <div
              key={uploadFile.id}
              className="flex items-center space-x-3 p-3 border rounded-lg bg-muted/30"
            >
              <div className="flex-shrink-0">
                {getFileIcon(uploadFile.file)}
              </div>
              
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <p className="text-sm font-medium truncate">
                    {uploadFile.file.name}
                  </p>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => removeUploadFile(uploadFile.id)}
                    className="h-6 w-6 p-0 text-muted-foreground hover:text-foreground"
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
                
                <p className="text-xs text-muted-foreground">
                  {formatFileSize(uploadFile.file.size)}
                </p>
                
                {uploadFile.status === 'uploading' && (
                  <Progress value={uploadFile.progress} className="mt-2 h-1" />
                )}
                
                {uploadFile.status === 'error' && (
                  <p className="text-xs text-destructive mt-1">
                    {uploadFile.error}
                  </p>
                )}
                
                {uploadFile.status === 'success' && (
                  <p className="text-xs text-green-600 mt-1">
                    {t('attachment.upload_success', 'Upload successful')}
                  </p>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
