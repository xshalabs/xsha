import { useState, useCallback } from "react";
import { attachmentApi, type Attachment } from "@/lib/api/attachments";

export function useAttachments(projectId: number) {
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [uploading, setUploading] = useState(false);

  // Generate attachment tags for content
  const generateAttachmentTags = useCallback((allAttachments: Attachment[]) => {
    const images = allAttachments.filter(a => a.type === 'image');
    const pdfs = allAttachments.filter(a => a.type === 'pdf');
    
    const tags: string[] = [];
    
    images.forEach((_, index) => {
      tags.push(`[image${index + 1}]`);
    });
    
    pdfs.forEach((_, index) => {
      tags.push(`[pdf${index + 1}]`);
    });
    
    return tags.join(' ');
  }, []);

  // Upload multiple files
  const uploadFiles = useCallback(async (files: FileList): Promise<Attachment[]> => {
    if (files.length === 0) return [];

    setUploading(true);
    
    try {
      const uploadPromises = Array.from(files).map(async (file) => {
        // Validate file
        if (file.size > 10 * 1024 * 1024) { // 10MB
          throw new Error(`File ${file.name} is too large (max 10MB)`);
        }
        
        const isImage = file.type.startsWith('image/');
        const isPdf = file.type === 'application/pdf';
        
        if (!isImage && !isPdf) {
          throw new Error(`File ${file.name} is not supported (only images and PDF)`);
        }

        return await attachmentApi.uploadAttachment(file, projectId);
      });

      const uploadedAttachments = await Promise.all(uploadPromises);
      
      // Update attachments state
      setAttachments(prev => [...prev, ...uploadedAttachments]);
      
      return uploadedAttachments;
    } catch (error) {
      console.error('Failed to upload files:', error);
      throw error; // Re-throw to allow caller to handle
    } finally {
      setUploading(false);
    }
  }, [projectId]);

  // Remove attachment
  const removeAttachment = useCallback(async (attachment: Attachment) => {
    try {
      await attachmentApi.deleteAttachment(attachment.id, projectId);
      setAttachments(prev => prev.filter(a => a.id !== attachment.id));
    } catch (error) {
      console.error('Failed to delete attachment:', error);
      throw error;
    }
  }, [projectId]);

  // Clear all attachments
  const clearAttachments = useCallback(() => {
    setAttachments([]);
  }, []);

  // Get attachment IDs
  const getAttachmentIds = useCallback(() => {
    return attachments.map(a => a.id);
  }, [attachments]);

  return {
    attachments,
    uploading,
    uploadFiles,
    removeAttachment,
    clearAttachments,
    getAttachmentIds,
    generateAttachmentTags,
  };
}
