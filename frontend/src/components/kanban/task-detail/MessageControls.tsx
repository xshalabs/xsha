import { memo, useCallback, useRef } from "react";
import { useTranslation } from "react-i18next";
import { Paperclip } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ExecutionTimeControl } from "./ExecutionTimeControl";
import { ModelSelector } from "./ModelSelector";
import type { Task } from "@/types/task";

interface MessageControlsProps {
  task: Task;
  executionTime?: Date;
  model: string;
  attachmentCount: number;
  sending: boolean;
  uploading: boolean;
  onExecutionTimeChange: (time: Date | undefined) => void;
  onModelChange: (model: string) => void;
  onFileSelect: () => void;
}

export const MessageControls = memo<MessageControlsProps>(
  ({
    task,
    executionTime,
    model,
    attachmentCount,
    sending,
    uploading,
    onExecutionTimeChange,
    onModelChange,
    onFileSelect,
  }) => {
    const { t } = useTranslation();
    const fileInputRef = useRef<HTMLInputElement>(null);
    
    // State to manage which control is open - only one at a time
    const closeAllControls = useCallback(() => {
      // This callback will be passed to individual controls to close others
    }, []);

    const handleAttachmentClick = useCallback(() => {
      onFileSelect();
      fileInputRef.current?.click();
    }, [onFileSelect]);

    return (
      <div className="absolute bottom-3 left-3 right-3 flex items-end gap-3">
        {/* Attachment Control */}
        <div className="relative">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleAttachmentClick}
            disabled={sending || uploading}
            className={`h-7 w-7 p-0 rounded-md transition-colors ${
              attachmentCount > 0
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
        </div>

        {/* Execution Time Control */}
        <ExecutionTimeControl
          executionTime={executionTime}
          onChange={onExecutionTimeChange}
          onCloseOtherControls={closeAllControls}
        />

        {/* Model Selection - Only show for claude-code environments */}
        {task.dev_environment?.type === "claude-code" && (
          <ModelSelector
            model={model}
            disabled={sending}
            onChange={onModelChange}
            onCloseOtherControls={closeAllControls}
          />
        )}
      </div>
    );
  }
);

MessageControls.displayName = "MessageControls";
