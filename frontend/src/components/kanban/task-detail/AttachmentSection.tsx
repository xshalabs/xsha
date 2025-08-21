import { memo } from "react";
import { useTranslation } from "react-i18next";
import { X, Paperclip, Image, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { Attachment } from "@/lib/api/attachments";

interface AttachmentSectionProps {
  attachments: Attachment[];
  onRemove: (attachment: Attachment) => void;
}

export const AttachmentSection = memo<AttachmentSectionProps>(
  ({ attachments, onRemove }) => {
    const { t } = useTranslation();

    const getDisplayName = (attachment: Attachment) => {
      const typeCount = attachments.filter(a => a.type === attachment.type).indexOf(attachment) + 1;
      return attachment.type === 'image' ? `[image${typeCount}]` : `[pdf${typeCount}]`;
    };

    if (attachments.length === 0) {
      return null;
    }

    return (
      <div className="space-y-2 p-3 bg-muted/30 rounded-lg border">
        <div className="flex items-center gap-2">
          <Paperclip className="h-3 w-3 text-muted-foreground" />
          <span className="text-xs font-medium text-muted-foreground">
            {t("taskConversations.attachments", "Attachments")} ({attachments.length})
          </span>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2">
          {attachments.map((attachment) => (
            <div
              key={attachment.id}
              className="flex items-center gap-2 p-2 bg-background rounded border min-w-0"
            >
              <div className="flex-shrink-0">
                {attachment.type === 'image' ? 
                  <Image className="h-3 w-3 text-blue-500" /> : 
                  <FileText className="h-3 w-3 text-red-500" />
                }
              </div>
              <span className="flex-1 text-xs truncate">
                {getDisplayName(attachment)}
              </span>
              <span className="text-xs text-muted-foreground whitespace-nowrap">
                {Math.round(attachment.file_size / 1024)}KB
              </span>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onRemove(attachment)}
                className="h-5 w-5 p-0 text-muted-foreground hover:text-destructive flex-shrink-0"
              >
                <X className="h-3 w-3" />
              </Button>
            </div>
          ))}
        </div>
      </div>
    );
  }
);

AttachmentSection.displayName = "AttachmentSection";
