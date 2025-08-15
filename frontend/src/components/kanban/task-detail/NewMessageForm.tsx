import { memo, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { Send, MessageSquare } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { DateTimePicker } from "@/components/ui/datetime-picker";

interface NewMessageFormProps {
  newMessage: string;
  executionTime: Date | undefined;
  sending: boolean;
  canSendMessage: boolean;
  isTaskCompleted: boolean;
  hasPendingOrRunningConversations: boolean;
  onMessageChange: (message: string) => void;
  onExecutionTimeChange: (time: Date | undefined) => void;
  onSendMessage: () => void;
}

export const NewMessageForm = memo<NewMessageFormProps>(
  ({
    newMessage,
    executionTime,
    sending,
    canSendMessage,
    isTaskCompleted,
    hasPendingOrRunningConversations,
    onMessageChange,
    onExecutionTimeChange,
    onSendMessage,
  }) => {
    const { t } = useTranslation();

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
      onSendMessage();
    }, [onSendMessage]);

    const isDisabled = !newMessage.trim() || sending || !canSendMessage;

    return (
      <div className="space-y-4 border-t pt-4">
        <h3 className="font-medium text-foreground text-lg flex items-center gap-2">
          <Send className="h-4 w-4" />
          {t("taskConversations.newMessage")}
        </h3>

        <div className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="message-content" className="text-sm font-medium">
              {t("taskConversations.content")}:
            </label>
            <Textarea
              id="message-content"
              className="min-h-[120px] resize-none"
              placeholder={t("taskConversations.contentPlaceholder")}
              value={newMessage}
              onChange={handleMessageChange}
              onKeyDown={handleKeyDown}
              aria-describedby="message-shortcut-hint"
            />
          </div>

          <div className="space-y-2">
            <label htmlFor="execution-time" className="text-sm font-medium">
              {t("taskConversations.executionTime")}:
            </label>
            <DateTimePicker
              id="execution-time"
              value={executionTime}
              onChange={onExecutionTimeChange}
              placeholder={t("taskConversations.executionTimePlaceholder")}
              label=""
              aria-describedby="execution-time-hint"
            />
            <p
              id="execution-time-hint"
              className="text-xs text-muted-foreground"
            >
              {t("taskConversations.executionTimeHint")}
            </p>
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
