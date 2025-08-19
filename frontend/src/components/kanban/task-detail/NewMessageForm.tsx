import { memo, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { Send, MessageSquare, Clock, Zap } from "lucide-react";
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
  onSendMessage: () => void;
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
      onSendMessage();
    }, [onSendMessage]);

    const isDisabled = !newMessage.trim() || sending || !canSendMessage;

    return (
      <div className="space-y-6 border-t p-6">
        <h3 className="font-medium text-foreground text-base flex items-center gap-2">
          <Send className="h-4 w-4" />
          {t("taskConversations.newMessage")}
        </h3>

        <div className="space-y-6">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <MessageSquare className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="message-content" className="text-sm font-medium">
                {t("taskConversations.content")}:
              </Label>
            </div>
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
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="execution-time" className="text-sm font-medium">
                {t("taskConversations.executionTime")}:
              </Label>
            </div>
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

          {/* Model Selection - Only show for claude-code environments */}
          {task.dev_environment?.type === "claude-code" && (
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Zap className="h-4 w-4 text-muted-foreground" />
                <Label htmlFor="model" className="text-sm font-medium">
                  {t("tasks.fields.model")}:
                </Label>
              </div>
              <Select
                value={model}
                onValueChange={onModelChange}
                disabled={sending}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t("taskConversations.selectModel")} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="default">
                    <div className="flex items-center justify-between w-full">
                      <span className="font-medium">{t("taskConversations.model.default")}</span>
                      <span className="text-xs text-muted-foreground ml-2">
                        {t("taskConversations.model.defaultDescription")}
                      </span>
                    </div>
                  </SelectItem>
                  <SelectItem value="sonnet">
                    <div className="flex items-center justify-between w-full">
                      <span className="font-medium">{t("taskConversations.model.sonnet")}</span>
                      <span className="text-xs text-muted-foreground ml-2">
                        Sonnet
                      </span>
                    </div>
                  </SelectItem>
                  <SelectItem value="opus">
                    <div className="flex items-center justify-between w-full">
                      <span className="font-medium">{t("taskConversations.model.opus")}</span>
                      <span className="text-xs text-muted-foreground ml-2">
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
          )}

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
