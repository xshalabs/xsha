import { memo, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { User, MoreHorizontal, Eye, FileText, Terminal, RotateCcw, X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { getConversationStatusColor, formatTime } from "./utils";

interface ConversationItemProps {
  conversation: any;
  taskId: number;
  isExpanded: boolean;
  shouldShowExpandButton: boolean;
  onToggleExpanded: (id: number) => void;
  onViewDetails: (taskId: number, conversationId: number) => void;
  onViewGitDiff: (conversationId: number) => void;
  onViewLogs: (conversationId: number) => void;
  onRetry: (conversationId: number) => void;
  onCancel: (conversationId: number) => void;
}

export const ConversationItem = memo<ConversationItemProps>(
  ({
    conversation,
    taskId,
    isExpanded,
    shouldShowExpandButton,
    onToggleExpanded,
    onViewDetails,
    onViewGitDiff,
    onViewLogs,
    onRetry,
    onCancel,
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

    // Check if git diff should be disabled (when commit_hash is empty)
    const isGitDiffDisabled = !conversation.commit_hash;
    
    // Check if retry should be enabled (only for failed or cancelled conversations)
    const isRetryEnabled = conversation.status === 'failed' || conversation.status === 'cancelled';
    
    // Check if cancel should be enabled (only for running conversations)
    const isCancelEnabled = conversation.status === 'running';

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
                <DropdownMenuItem onClick={handleViewDetails}>
                  <Eye className="mr-2 h-4 w-4" />
                  {t("taskConversations.actions.viewDetails")}
                </DropdownMenuItem>
                <DropdownMenuItem 
                  onClick={handleViewGitDiff}
                  disabled={isGitDiffDisabled}
                  className={isGitDiffDisabled ? "opacity-50 cursor-not-allowed" : ""}
                >
                  <FileText className="mr-2 h-4 w-4" />
                  {t("taskConversations.actions.viewGitDiff")}
                </DropdownMenuItem>
                <DropdownMenuItem onClick={handleViewLogs}>
                  <Terminal className="mr-2 h-4 w-4" />
                  {t("taskConversations.actions.viewLogs")}
                </DropdownMenuItem>
                {isCancelEnabled && (
                  <DropdownMenuItem 
                    onClick={handleCancel}
                    className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
                  >
                    <X className="mr-2 h-4 w-4" />
                    {t("taskConversations.actions.cancel")}
                  </DropdownMenuItem>
                )}
                {isRetryEnabled && (
                  <DropdownMenuItem 
                    onClick={handleRetry}
                    className="text-orange-600 hover:text-orange-700 hover:bg-orange-50 dark:hover:bg-orange-950"
                  >
                    <RotateCcw className="mr-2 h-4 w-4" />
                    {t("taskConversations.actions.retry")}
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    );
  }
);

ConversationItem.displayName = "ConversationItem";
