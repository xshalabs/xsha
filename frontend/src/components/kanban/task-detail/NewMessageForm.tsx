import { memo, useCallback, useRef } from "react";
import { useTranslation } from "react-i18next";
import { Send, MessageSquare } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import type { Task } from "@/types/task";
import { useAttachments } from "@/hooks/useAttachments";
import { AttachmentSection } from "./AttachmentSection";
import { MessageControls } from "./MessageControls";
import type { Attachment } from "@/lib/api/attachments";

interface NewMessageFormProps {
  task: Task;
  newMessage: string;
  executionTime: Date | undefined;
  model: string;
  sending: boolean;
  canSendMessage: boolean;
  isTaskCompleted: boolean;
  _hasPendingOrRunningConversations: boolean;
  onMessageChange: (message: string) => void;
  onExecutionTimeChange: (time: Date | undefined) => void;
  onModelChange: (model: string) => void;
  onSendMessage: (attachmentIds?: number[]) => void;
}

export const NewMessageForm = memo<NewMessageFormProps>(
  ({
    task,
    newMessage,
    executionTime,
    model,
    sending,
    canSendMessage,
    isTaskCompleted,
    _hasPendingOrRunningConversations,
    onMessageChange,
    onExecutionTimeChange,
    onModelChange,
    onSendMessage,
  }) => {
    const { t } = useTranslation();
    const fileInputRef = useRef<HTMLInputElement>(null);

    // Use attachment hook
    const {
      attachments,
      uploading,
      uploadFiles,
      removeAttachment,
      clearAttachments,
      getAttachmentIds,
    } = useAttachments();

    // Explicitly acknowledge the parameter to avoid linter warning
    // This parameter is used implicitly through canSendMessage logic
    void _hasPendingOrRunningConversations;





    const handleMessageChange = useCallback(
      (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        onMessageChange(e.target.value);
      },
      [onMessageChange]
    );

    const handleKeyDown = useCallback(
      (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
          e.preventDefault();
          onSendMessage();
        }
      },
      [onSendMessage]
    );

    const handleSendMessage = useCallback(() => {
      const attachmentIds = getAttachmentIds();
      onSendMessage(attachmentIds.length > 0 ? attachmentIds : undefined);
      
      // Clear attachments after sending (they will be associated with the new conversation)
      clearAttachments();
    }, [onSendMessage, getAttachmentIds, clearAttachments]);

    // Handle file input change
    const handleFileInputChange = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
      const files = event.target.files;
      if (!files || files.length === 0) return;

      try {
        const uploadedAttachments = await uploadFiles(files);
        
        // Add attachment tags to the message content
        if (uploadedAttachments && uploadedAttachments.length > 0) {
          // Generate tags only for newly uploaded attachments
          const newTags: string[] = [];
          
          uploadedAttachments.forEach((attachment, index) => {
            if (attachment.type === 'image') {
              // Count existing images to get correct index
              const imageCount = attachments.filter(a => a.type === 'image').length + 
                               uploadedAttachments.slice(0, index + 1).filter(a => a.type === 'image').length;
              newTags.push(`[image${imageCount}]`);
            } else if (attachment.type === 'pdf') {
              // Count existing PDFs to get correct index  
              const pdfCount = attachments.filter(a => a.type === 'pdf').length + 
                             uploadedAttachments.slice(0, index + 1).filter(a => a.type === 'pdf').length;
              newTags.push(`[pdf${pdfCount}]`);
            }
          });
          
          // Add new tags to the end of existing content
          const newContent = newMessage.trim() ? 
            `${newMessage.trim()} ${newTags.join(' ')}` : 
            newTags.join(' ');
          
          onMessageChange(newContent);
        }
      } catch (error) {
        console.error('Failed to upload files:', error);
        // You might want to show a toast error here
      } finally {
        // Reset file input
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
      }
    }, [uploadFiles, attachments, newMessage, onMessageChange]);

    // Handle file select trigger
    const handleFileSelect = useCallback(() => {
      fileInputRef.current?.click();
    }, []);

    // Handle attachment removal with tag update
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
          
          // Remove the tag from the message
          let updatedMessage = newMessage.replace(new RegExp(`\\s*${tagToRemove.replace(/[[\]]/g, '\\$&')}\\s*`, 'g'), ' ');
          
          // Renumber remaining tags of the same type
          for (let i = attachmentIndex + 1; i < sameTypeAttachments.length; i++) {
            const oldTag = attachment.type === 'image' ? `[image${i + 1}]` : `[pdf${i + 1}]`;
            const newTag = attachment.type === 'image' ? `[image${i}]` : `[pdf${i}]`;
            updatedMessage = updatedMessage.replace(new RegExp(oldTag.replace(/[[\]]/g, '\\$&'), 'g'), newTag);
          }
          
          // Clean up extra spaces
          updatedMessage = updatedMessage.replace(/\s+/g, ' ').trim();
          
          // Update the message
          if (updatedMessage !== newMessage) {
            onMessageChange(updatedMessage);
          }
        }
        
        // Remove the attachment
        await removeAttachment(attachment);
      } catch (error) {
        console.error('Failed to remove attachment:', error);
      }
    }, [removeAttachment, attachments, newMessage, onMessageChange]);

    const isDisabled = !newMessage.trim() || sending || !canSendMessage;

    return (
      <div className="space-y-6 border-t p-6">
        <h3 className="font-medium text-foreground text-base flex items-center gap-2">
          <Send className="h-4 w-4" />
          {t("taskConversations.newMessage")}
        </h3>

        <div className="space-y-4">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <MessageSquare className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="message-content" className="text-sm font-medium">
                {t("taskConversations.content")}:
              </Label>
            </div>
            
            {/* Uploaded Attachments Display - Above the textarea */}
            <AttachmentSection
              attachments={attachments}
              onRemove={handleAttachmentRemove}
            />
            
            <div className="relative">
              <Textarea
                id="message-content"
                className="min-h-[120px] resize-none pr-4 pb-16"
                placeholder={t("taskConversations.contentPlaceholder")}
                value={newMessage}
                onChange={handleMessageChange}
                onKeyDown={handleKeyDown}
                aria-describedby="message-shortcut-hint"
              />
              
              {/* Interactive Controls positioned at the bottom left of the textarea */}
              <MessageControls
                task={task}
                executionTime={executionTime}
                model={model}
                attachmentCount={attachments.length}
                sending={sending}
                uploading={uploading}
                onExecutionTimeChange={onExecutionTimeChange}
                onModelChange={onModelChange}
                onFileSelect={handleFileSelect}
              />
                  
                  {/* Hidden file input */}
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/*,.pdf"
                    multiple
                    onChange={handleFileInputChange}
                    className="hidden"
                  />
            </div>
            
            {/* Hint for interactive controls */}
            <div className="text-xs text-muted-foreground">
              {t("taskConversations.clickIconsToConfigureHint", "Click icons in the text area to configure execution settings")}
            </div>
          </div>

          <div className="flex items-center justify-between">
            <div
              id="message-shortcut-hint"
              className="text-xs text-muted-foreground"
            >
              {t("taskConversations.shortcut")}
            </div>
            <Button
              onClick={handleSendMessage}
              disabled={isDisabled}
              className="flex items-center space-x-2"
              aria-label={sending ? t("common.sending") : t("common.send")}
            >
              <MessageSquare className="w-4 h-4" />
              <span>{sending ? t("common.sending") : t("common.send")}</span>
            </Button>
          </div>

          {!canSendMessage && (
            <div
              className="text-sm text-amber-600 bg-amber-50 p-3 rounded-lg border border-amber-200 dark:bg-amber-900/20 dark:border-amber-800 dark:text-amber-200"
              role="alert"
              aria-live="polite"
            >
              {isTaskCompleted
                ? t("taskConversations.taskCompletedMessage")
                : t("taskConversations.hasPendingMessage")}
            </div>
          )}
        </div>
      </div>
    );
  }
);

NewMessageForm.displayName = "NewMessageForm";
