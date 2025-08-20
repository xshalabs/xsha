import { memo, useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { User, MoreHorizontal, Eye, FileText, Terminal, RotateCcw, X, Trash2, Copy, Check, Clock } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { StatusDot, type ConversationStatus } from "./StatusDot";
import { formatTime, formatTimeWithoutSeconds, isFutureExecution } from "./utils";

interface ConversationItemProps {
  conversation: any;
  taskId: number;
  task: any;
  isExpanded: boolean;
  shouldShowExpandButton: boolean;
  isLatest?: boolean;
  onToggleExpanded: (id: number) => void;
  onViewDetails: (taskId: number, conversationId: number) => void;
  onViewGitDiff: (conversationId: number) => void;
  onViewLogs: (conversationId: number) => void;
  onRetry: (conversationId: number) => void;
  onCancel: (conversationId: number) => void;
  onDelete: (conversationId: number) => void;
}

export const ConversationItem = memo<ConversationItemProps>(
  ({
    conversation,
    taskId,
    task,
    isExpanded,
    shouldShowExpandButton,
    isLatest = false,
    onToggleExpanded,
    onViewDetails,
    onViewGitDiff,
    onViewLogs,
    onRetry,
    onCancel,
    onDelete,
  }) => {
    const { t } = useTranslation();
    const [copied, setCopied] = useState(false);

    const handleCopyContent = useCallback(async () => {
      try {
        await navigator.clipboard.writeText(conversation.content);
        toast.success(t("common.copied_to_clipboard"));
        setCopied(true);
        // Reset the copied state after 2 seconds
        setTimeout(() => setCopied(false), 2000);
      } catch (error) {
        console.error("Failed to copy:", error);
        toast.error(t("common.copy_failed"));
      }
    }, [conversation.content, t]);

    const handleToggleExpanded = useCallback(() => {
      onToggleExpanded(conversation.id);
    }, [conversation.id, onToggleExpanded]);

    const handleViewDetails = useCallback(() => {
      onViewDetails(taskId, conversation.id);
    }, [taskId, conversation.id, onViewDetails]);

    const handleViewGitDiff = useCallback(() => {
      onViewGitDiff(conversation.id);
    }, [conversation.id, onViewGitDiff]);

    const handleViewLogs = useCallback(() => {
      onViewLogs(conversation.id);
    }, [conversation.id, onViewLogs]);

    const handleRetry = useCallback(() => {
      onRetry(conversation.id);
    }, [conversation.id, onRetry]);

    const handleCancel = useCallback(() => {
      onCancel(conversation.id);
    }, [conversation.id, onCancel]);

    const handleDelete = useCallback(() => {
      onDelete(conversation.id);
    }, [conversation.id, onDelete]);

    // Check if git diff should be disabled (when commit_hash is empty)
    const isGitDiffDisabled = !conversation.commit_hash;
    
    // Check if view details should be disabled (pending or running conversations)
    const isViewDetailsDisabled = conversation.status === 'pending' || conversation.status === 'running';
    
    // Check if view logs should be disabled (pending conversations have no execution logs yet)
    const isViewLogsDisabled = conversation.status === 'pending';
    
    // Check if retry should be enabled (only for failed or cancelled conversations)
    const isRetryEnabled = conversation.status === 'failed' || conversation.status === 'cancelled';
    
    // Check if cancel should be enabled (only for running conversations)
    const isCancelEnabled = conversation.status === 'running';
    
    // Check if delete should be enabled (only for latest conversation, not running, and task not pending/in progress)
    const isDeleteEnabled = isLatest && conversation.status !== 'running' && 
                           task.status !== 'pending' && task.status !== 'in_progress';
    
    // Check if this conversation is scheduled for future execution
    const isFuture = isFutureExecution(conversation.execution_time);

    return (
      <div className="p-4 rounded-md border border-border bg-card">
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1 min-w-0">
            <div className="flex items-center space-x-2 mb-2">
              <User className="w-4 h-4" />
              <span className="font-medium text-sm">
                {conversation.created_by}
              </span>
              <time
                className="text-xs text-muted-foreground"
                dateTime={conversation.created_at}
              >
                {formatTime(conversation.created_at)}
              </time>
              {/* Future execution indicator */}
              {isFuture && (
                <Badge 
                  variant="outline" 
                  className="text-xs h-5 px-1.5 bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-900/20 dark:text-purple-300 dark:border-purple-800"
                >
                  <Clock className="w-2.5 h-2.5 mr-1" />
                  {t("taskConversations.status.scheduled")}
                </Badge>
              )}
            </div>
            <div className="relative group">
              <div
                className={`text-sm whitespace-pre-wrap pr-8 ${
                  isExpanded ? "" : shouldShowExpandButton ? "line-clamp-3" : ""
                }`}
              >
                {conversation.content}
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={handleCopyContent}
                className="absolute top-0 right-0 h-6 w-6 p-0 text-muted-foreground hover:text-foreground hover:bg-muted opacity-0 group-hover:opacity-100 transition-opacity"
                title={t("common.copy")}
              >
                {copied ? (
                  <Check className="h-3 w-3 text-green-600" />
                ) : (
                  <Copy className="h-3 w-3" />
                )}
              </Button>
            </div>
            
            {/* Execution time info for future scheduled tasks */}
            {isFuture && conversation.execution_time && (
              <div className="mt-2 text-xs text-muted-foreground flex items-center">
                <Clock className="w-3 h-3 mr-1 flex-shrink-0" />
                <span>{t("taskConversations.details.scheduledFor")}: {formatTimeWithoutSeconds(conversation.execution_time)}</span>
              </div>
            )}
            
            {shouldShowExpandButton && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleToggleExpanded}
                className="mt-1 h-6 px-1 text-xs text-muted-foreground hover:bg-muted"
                aria-label={
                  isExpanded ? t("common.showLess") : t("common.showMore")
                }
              >
                {isExpanded ? t("common.showLess") : t("common.showMore")}
              </Button>
            )}
          </div>

          <div className="flex items-center space-x-3 shrink-0">
            <StatusDot status={conversation.status as ConversationStatus} />

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 w-8 p-0"
                  aria-label={t("common.moreActions")}
                >
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem 
                  onClick={isViewDetailsDisabled ? undefined : handleViewDetails}
                  disabled={isViewDetailsDisabled}
                  className={isViewDetailsDisabled ? "opacity-50 cursor-not-allowed" : ""}
                >
                  <Eye className="mr-2 h-4 w-4" />
                  {t("taskConversations.actions.viewDetails")}
                </DropdownMenuItem>
                <DropdownMenuItem 
                  onClick={handleViewLogs}
                  disabled={isViewLogsDisabled}
                  className={isViewLogsDisabled ? "opacity-50 cursor-not-allowed" : ""}
                >
                  <Terminal className="mr-2 h-4 w-4" />
                  {t("taskConversations.actions.logs")}
                </DropdownMenuItem>
                <DropdownMenuItem 
                  onClick={isGitDiffDisabled ? undefined : handleViewGitDiff}
                  disabled={isGitDiffDisabled}
                  className={isGitDiffDisabled ? "opacity-50 cursor-not-allowed" : ""}
                >
                  <FileText className="mr-2 h-4 w-4" />
                  {t("taskConversations.actions.viewGitDiff")}
                </DropdownMenuItem>
                {isRetryEnabled && (
                  <DropdownMenuItem onClick={handleRetry}>
                    <RotateCcw className="mr-2 h-4 w-4" />
                    {t("taskConversations.actions.retry")}
                  </DropdownMenuItem>
                )}
                {isCancelEnabled && (
                  <DropdownMenuItem onClick={handleCancel}>
                    <X className="mr-2 h-4 w-4" />
                    {t("taskConversations.actions.cancel")}
                  </DropdownMenuItem>
                )}
                <DropdownMenuSeparator />
                <DropdownMenuItem 
                  onClick={isDeleteEnabled ? handleDelete : undefined}
                  disabled={!isDeleteEnabled}
                  className={isDeleteEnabled 
                    ? "text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950" 
                    : "opacity-50 cursor-not-allowed"}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  {t("taskConversations.actions.delete")}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    );
  }
);

ConversationItem.displayName = "ConversationItem";
