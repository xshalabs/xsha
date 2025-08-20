import { memo, useCallback, useState, useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { Send, MessageSquare, Clock, Zap, Calendar, Sparkles, X, Paperclip, Image, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Task } from "@/types/task";
import { attachmentApi, type Attachment } from "@/lib/api/attachments";

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
    const [isTimePickerOpen, setIsTimePickerOpen] = useState(false);
    const [isModelSelectorOpen, setIsModelSelectorOpen] = useState(false);
    const [attachments, setAttachments] = useState<Attachment[]>([]);
    const [uploading, setUploading] = useState(false);
    const timePickerRef = useRef<HTMLDivElement>(null);
    const modelSelectorRef = useRef<HTMLDivElement>(null);
    const fileInputRef = useRef<HTMLInputElement>(null);

    // Explicitly acknowledge the parameter to avoid linter warning
    // This parameter is used implicitly through canSendMessage logic
    void _hasPendingOrRunningConversations;

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

    // Update content with attachment tags when attachments change
    useEffect(() => {
      const tags = generateAttachmentTags(attachments);
      if (tags) {
        // Remove existing tags from content first
        let currentContent = newMessage.replace(/\[(image|pdf)\d+\]\s*/g, '').trim();
        const newContent = currentContent ? `${currentContent} ${tags}` : tags;
        if (newContent !== newMessage) {
          onMessageChange(newContent);
        }
      } else if (newMessage.match(/\[(image|pdf)\d+\]/)) {
        // If no attachments but content has tags, remove them
        const cleanContent = newMessage.replace(/\[(image|pdf)\d+\]\s*/g, '').trim();
        if (cleanContent !== newMessage) {
          onMessageChange(cleanContent);
        }
      }
    }, [attachments, generateAttachmentTags, newMessage, onMessageChange]);

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
      const attachmentIds = attachments.map(a => a.id);
      onSendMessage(attachmentIds.length > 0 ? attachmentIds : undefined);
      
      // Clear attachments after sending (they will be associated with the new conversation)
      setAttachments([]);
    }, [onSendMessage, attachments]);

    const handleTimePickerToggle = useCallback(() => {
      setIsTimePickerOpen(!isTimePickerOpen);
      setIsModelSelectorOpen(false); // Close model selector when opening time picker
    }, [isTimePickerOpen]);

    const handleModelSelectorToggle = useCallback(() => {
      setIsModelSelectorOpen(!isModelSelectorOpen);
      setIsTimePickerOpen(false); // Close time picker when opening model selector
    }, [isModelSelectorOpen]);

    const handleAttachmentClick = useCallback(() => {
      fileInputRef.current?.click();
    }, []);

    const handleTimeChange = useCallback((time: Date | undefined) => {
      onExecutionTimeChange(time);
      // Don't auto-close to allow multiple time adjustments
    }, [onExecutionTimeChange]);

    const handleModelChange = useCallback((newModel: string) => {
      onModelChange(newModel);
      setIsModelSelectorOpen(false); // Close after selection
    }, [onModelChange]);

    // Handle file input change
    const handleFileInputChange = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
      const files = event.target.files;
      if (!files || files.length === 0) return;

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

          return await attachmentApi.uploadAttachment(file);
        });

        const uploadedAttachments = await Promise.all(uploadPromises);
        
        // Update attachments state
        setAttachments(prev => [...prev, ...uploadedAttachments]);
        
      } catch (error) {
        console.error('Failed to upload files:', error);
        // You might want to show a toast error here
      } finally {
        setUploading(false);
        // Reset file input
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
      }
    }, []);

    // Handle attachment removal
    const handleAttachmentRemove = useCallback(async (attachment: Attachment) => {
      try {
        await attachmentApi.deleteAttachment(attachment.id);
        setAttachments(prev => prev.filter(a => a.id !== attachment.id));
      } catch (error) {
        console.error('Failed to delete attachment:', error);
      }
    }, []);

    // Handle click outside to close popups - simplified approach
    useEffect(() => {
      const handleClickOutside = (event: MouseEvent) => {
        const target = event.target as Element;
        
        // Only close if clicking completely outside our components and not on any portal/popup content
        const isClickOnPortal = target.closest('[data-radix-popper-content-wrapper], [data-radix-portal], [data-sonner-toaster]');
        const isClickOnTimePicker = timePickerRef.current?.contains(target as Node);
        const isClickOnModelSelector = modelSelectorRef.current?.contains(target as Node);
        
        if (!isClickOnPortal && !isClickOnTimePicker && !isClickOnModelSelector) {
          setIsTimePickerOpen(false);
          setIsModelSelectorOpen(false);
        }
      };

      // Use a timeout to avoid immediate closure
      const timeoutId = setTimeout(() => {
        document.addEventListener('mousedown', handleClickOutside);
      }, 100);

      return () => {
        clearTimeout(timeoutId);
        document.removeEventListener('mousedown', handleClickOutside);
      };
    }, []);

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
            {attachments.length > 0 && (
              <div className="space-y-2 p-3 bg-muted/30 rounded-lg border">
                <div className="flex items-center gap-2">
                  <Paperclip className="h-3 w-3 text-muted-foreground" />
                  <span className="text-xs font-medium text-muted-foreground">
                    {t("taskConversations.attachments", "Attachments")} ({attachments.length})
                  </span>
                </div>
                <div className="grid gap-2">
                  {attachments.map((attachment) => (
                    <div
                      key={attachment.id}
                      className="flex items-center gap-2 p-2 bg-background rounded border"
                    >
                      <div className="flex-shrink-0">
                        {attachment.type === 'image' ? 
                          <Image className="h-3 w-3 text-blue-500" /> : 
                          <FileText className="h-3 w-3 text-red-500" />
                        }
                      </div>
                      <span className="flex-1 text-xs truncate">
                        {attachment.original_name}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {Math.round(attachment.file_size / 1024)}KB
                      </span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleAttachmentRemove(attachment)}
                        className="h-5 w-5 p-0 text-muted-foreground hover:text-destructive"
                      >
                        <X className="h-3 w-3" />
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            )}
            
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
              <div className="absolute bottom-3 left-3 right-3 flex items-end gap-3">
                {/* Execution Time Control */}
                <div className="relative" ref={timePickerRef}>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={handleTimePickerToggle}
                    className={`h-7 w-7 p-0 rounded-md transition-colors ${
                      executionTime 
                        ? 'bg-blue-100 text-blue-600 hover:bg-blue-200 dark:bg-blue-900/50 dark:text-blue-400' 
                        : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                    }`}
                    title={executionTime ? t("taskConversations.executionTime") + ": " + executionTime.toLocaleString() : t("taskConversations.executionTime")}
                  >
                    {executionTime ? <Calendar className="h-3.5 w-3.5" /> : <Clock className="h-3.5 w-3.5" />}
                  </Button>
                  
                                      {isTimePickerOpen && (
                      <div className="absolute bottom-full left-0 mb-2 p-3 bg-background border rounded-lg shadow-lg z-10 min-w-[200px]">
                        <div className="flex items-center justify-between mb-2">
                          <Label className="text-xs font-medium">{t("taskConversations.executionTime")}</Label>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setIsTimePickerOpen(false)}
                            className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground"
                          >
                            <X className="h-3 w-3" />
                          </Button>
                        </div>
                        <div className="space-y-2">
                          <DateTimePicker
                            value={executionTime}
                            onChange={handleTimeChange}
                            placeholder={t("taskConversations.executionTimePlaceholder")}
                            label=""
                            className="h-8 text-xs"
                          />
                          <p className="text-xs text-muted-foreground">
                            {t("taskConversations.executionTimeHint")}
                          </p>
                        </div>
                      </div>
                    )}
                </div>

                {/* Model Selection - Only show for claude-code environments */}
                {task.dev_environment?.type === "claude-code" && (
                  <div className="relative" ref={modelSelectorRef}>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={handleModelSelectorToggle}
                      className={`h-7 w-7 p-0 rounded-md transition-colors ${
                        model && model !== 'default'
                          ? 'bg-purple-100 text-purple-600 hover:bg-purple-200 dark:bg-purple-900/50 dark:text-purple-400'
                          : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                      }`}
                      title={model ? t("taskConversations.selectModel") + ": " + model : t("taskConversations.selectModel")}
                    >
                      {model && model !== 'default' ? <Sparkles className="h-3.5 w-3.5" /> : <Zap className="h-3.5 w-3.5" />}
                    </Button>
                    
                    {isModelSelectorOpen && (
                      <div className="absolute bottom-full left-0 mb-2 p-3 bg-background border rounded-lg shadow-lg z-10 min-w-[200px]">
                        <div className="flex items-center justify-between mb-2">
                          <Label className="text-xs font-medium">{t("taskConversations.selectModel")}</Label>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setIsModelSelectorOpen(false)}
                            className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground"
                          >
                            <X className="h-3 w-3" />
                          </Button>
                        </div>
                        <div className="space-y-2">
                          <Select
                            value={model}
                            onValueChange={handleModelChange}
                            disabled={sending}
                          >
                            <SelectTrigger className="h-8 text-xs">
                              <SelectValue placeholder={t("taskConversations.selectModel")} />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="default">
                                <div className="flex flex-col items-start">
                                  <span className="font-medium text-xs">{t("taskConversations.model.default")}</span>
                                  <span className="text-xs text-muted-foreground">
                                    {t("taskConversations.model.defaultDescription")}
                                  </span>
                                </div>
                              </SelectItem>
                              <SelectItem value="sonnet">
                                <div className="flex flex-col items-start">
                                  <span className="font-medium text-xs">{t("taskConversations.model.sonnet")}</span>
                                  <span className="text-xs text-muted-foreground">
                                    Sonnet
                                  </span>
                                </div>
                              </SelectItem>
                              <SelectItem value="opus">
                                <div className="flex flex-col items-start">
                                  <span className="font-medium text-xs">{t("taskConversations.model.opus")}</span>
                                  <span className="text-xs text-muted-foreground">
                                    Opus
                                  </span>
                                </div>
                              </SelectItem>
                            </SelectContent>
                          </Select>
                          <p className="text-xs text-muted-foreground">
                            {t("taskConversations.modelHint")}
                          </p>
                        </div>
                      </div>
                    )}
                  </div>
                )}

                {/* Attachment Control */}
                <div className="relative">
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={handleAttachmentClick}
                    disabled={sending || uploading}
                    className={`h-7 w-7 p-0 rounded-md transition-colors ${
                      attachments.length > 0
                        ? 'bg-green-100 text-green-600 hover:bg-green-200 dark:bg-green-900/50 dark:text-green-400'
                        : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                    } ${uploading ? 'opacity-50' : ''}`}
                    title={uploading ? t("attachment.uploading", "Uploading...") : t("taskConversations.attachments", "Attachments")}
                  >
                    {uploading ? (
                      <div className="animate-spin rounded-full h-3.5 w-3.5 border-b border-current"></div>
                    ) : (
                      <Paperclip className="h-3.5 w-3.5" />
                    )}
                  </Button>
                  
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
              </div>
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
