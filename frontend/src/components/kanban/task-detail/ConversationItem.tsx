import { memo, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { User, MoreHorizontal, Eye, FileText, Terminal, RotateCcw, X, Trash2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { getConversationStatusColor, formatTime } from "./utils";

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
            </div>
            <div
              className={`text-sm whitespace-pre-wrap ${
                isExpanded ? "" : shouldShowExpandButton ? "line-clamp-3" : ""
              }`}
            >
              {conversation.content}
            </div>
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

          <div className="flex items-center space-x-2 shrink-0">
            <Badge className={getConversationStatusColor(conversation.status)}>
              {t(`taskConversations.status.${conversation.status}`)}
            </Badge>

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
