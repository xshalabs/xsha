import { useCallback, useRef } from "react";
import { useAttachments } from "@/hooks/useAttachments";
import type { Attachment } from "@/lib/api/attachments";

interface UseTaskFormFileHandlingOptions {
  requirementDesc: string;
  onRequirementDescChange: (value: string) => void;
  projectId: number;
}

export function useTaskFormFileHandling({ requirementDesc, onRequirementDescChange, projectId }: UseTaskFormFileHandlingOptions) {
  const {
    attachments,
    uploading: uploadingAttachments,
    uploadFiles,
    removeAttachment,
    clearAttachments,
    getAttachmentIds,
  } = useAttachments(projectId);

  // Use refs to avoid dependency chain issues
  const requirementDescRef = useRef(requirementDesc);
  const onRequirementDescChangeRef = useRef(onRequirementDescChange);
  
  // Update refs when props change
  requirementDescRef.current = requirementDesc;
  onRequirementDescChangeRef.current = onRequirementDescChange;

  const generateAttachmentTags = useCallback((uploadedAttachments: Attachment[]): string[] => {
    const newTags: string[] = [];
    
    uploadedAttachments.forEach((attachment, index) => {
      if (attachment.type === 'image') {
        const imageCount = attachments.filter(a => a.type === 'image').length + 
                         uploadedAttachments.slice(0, index + 1).filter(a => a.type === 'image').length;
        newTags.push(`[image${imageCount}]`);
      } else if (attachment.type === 'pdf') {
        const pdfCount = attachments.filter(a => a.type === 'pdf').length + 
                       uploadedAttachments.slice(0, index + 1).filter(a => a.type === 'pdf').length;
        newTags.push(`[pdf${pdfCount}]`);
      }
    });
    
    return newTags;
  }, [attachments]);

  const addTagsToDescription = useCallback((newTags: string[]) => {
    const currentDesc = requirementDescRef.current || "";
    const newContent = currentDesc.trim() ? 
      `${currentDesc.trim()} ${newTags.join(' ')}` : 
      newTags.join(' ');
    
    onRequirementDescChangeRef.current(newContent);
  }, []);

  const handlePaste = useCallback(async (e: React.ClipboardEvent<HTMLTextAreaElement>) => {
    const items = e.clipboardData?.items;
    if (!items) return;

    const imageFiles: File[] = [];
    
    // Check for image files in clipboard
    for (let i = 0; i < items.length; i++) {
      const item = items[i];
      if (item.type.startsWith('image/')) {
        const file = item.getAsFile();
        if (file) {
          imageFiles.push(file);
        }
      }
    }

    // Upload image files if found
    if (imageFiles.length > 0) {
      try {
        const fileList = new DataTransfer();
        imageFiles.forEach(file => fileList.items.add(file));
        
        const uploadedAttachments = await uploadFiles(fileList.files);
        
        // Add attachment tags to the requirement_desc
        if (uploadedAttachments && uploadedAttachments.length > 0) {
          const newTags = generateAttachmentTags(uploadedAttachments);
          addTagsToDescription(newTags);
        }
      } catch (error) {
        console.error('Failed to upload pasted images:', error);
      }
    }
  }, [uploadFiles, generateAttachmentTags, addTagsToDescription]);

  const handleFileInputChange = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (!files || files.length === 0) return;

    try {
      const uploadedAttachments = await uploadFiles(files);
      
      // Add attachment tags to the requirement_desc
      if (uploadedAttachments && uploadedAttachments.length > 0) {
        const newTags = generateAttachmentTags(uploadedAttachments);
        addTagsToDescription(newTags);
      }
    } catch (error) {
      console.error('Failed to upload files:', error);
    }
  }, [uploadFiles, generateAttachmentTags, addTagsToDescription]);

  const handleAttachmentRemove = useCallback(async (attachment: Attachment) => {
    try {
      // Find the index of the attachment being removed among its type
      const sameTypeAttachments = attachments.filter(a => a.type === attachment.type);
      const attachmentIndex = sameTypeAttachments.findIndex(a => a.id === attachment.id);
      
      if (attachmentIndex !== -1) {
        // Remove the specific tag for this attachment
        const tagToRemove = attachment.type === 'image' ? 
          `[image${attachmentIndex + 1}]` : 
          `[pdf${attachmentIndex + 1}]`;
        
        const currentDesc = requirementDescRef.current || "";
        
        // Remove the tag from the requirement_desc
        let updatedDesc = currentDesc.replace(new RegExp(`\\s*${tagToRemove.replace(/[[\]]/g, '\\$&')}\\s*`, 'g'), ' ');
        
        // Renumber remaining tags of the same type
        for (let i = attachmentIndex + 1; i < sameTypeAttachments.length; i++) {
          const oldTag = attachment.type === 'image' ? `[image${i + 1}]` : `[pdf${i + 1}]`;
          const newTag = attachment.type === 'image' ? `[image${i}]` : `[pdf${i}]`;
          updatedDesc = updatedDesc.replace(new RegExp(oldTag.replace(/[[\]]/g, '\\$&'), 'g'), newTag);
        }
        
        // Clean up extra spaces
        updatedDesc = updatedDesc.replace(/\s+/g, ' ').trim();
        
        // Update the requirement_desc
        if (updatedDesc !== currentDesc) {
          onRequirementDescChangeRef.current(updatedDesc);
        }
      }
      
      // Remove the attachment
      await removeAttachment(attachment);
    } catch (error) {
      console.error("Failed to remove attachment:", error);
    }
  }, [removeAttachment, attachments]);

  return {
    attachments,
    uploadingAttachments,
    handlePaste,
    handleFileInputChange,
    handleAttachmentRemove,
    clearAttachments,
    getAttachmentIds,
  };
}