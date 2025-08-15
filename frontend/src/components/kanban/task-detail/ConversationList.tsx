import { memo, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { MessageSquare, RefreshCcw } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ConversationItem } from "./ConversationItem";

interface ConversationListProps {
  conversations: any[];
  conversationsLoading: boolean;
  conversationCount: number;
  task: any;
  taskId: number;
  onLoadConversations: () => void;
  onViewConversationGitDiff: (conversationId: number) => void;
  onViewConversationDetails: (conversationId: number) => void;
  onViewConversationLogs: (conversationId: number) => void;
  onRetryConversation: (conversationId: number) => void;
  onCancelConversation: (conversationId: number) => void;
  onDeleteConversation: (conversationId: number) => void;
  toggleExpanded: (id: number) => void;
  isConversationExpanded: (id: number) => boolean;
  shouldShowExpandButton: (content: string) => boolean;
}

export const ConversationList = memo<ConversationListProps>(
  ({
    conversations,
    conversationsLoading,
    conversationCount,
    task,
    taskId,
    onLoadConversations,
    onViewConversationGitDiff,
    onViewConversationDetails,
    onViewConversationLogs,
    onRetryConversation,
    onCancelConversation,
    onDeleteConversation,
    toggleExpanded,
    isConversationExpanded,
    shouldShowExpandButton,
  }) => {
    const { t } = useTranslation();

    const handleViewDetails = useCallback(
      (_taskId: number, conversationId: number) => {
        onViewConversationDetails(conversationId);
      },
      [onViewConversationDetails]
    );

    const handleLoadConversations = useCallback(() => {
      onLoadConversations();
    }, [onLoadConversations]);

    return (
      <div className="space-y-6 px-6">
        <div className="flex items-center justify-between">
          <h3 className="font-medium text-foreground text-base flex items-center gap-2">
            <MessageSquare className="h-4 w-4" />
            {t("taskConversations.list.title")}
            {conversationCount > 0 && (
              <Badge variant="outline" className="ml-1 text-xs">
                {conversationCount}
              </Badge>
            )}
          </h3>
          <Button
            variant="outline"
            size="sm"
            onClick={handleLoadConversations}
            disabled={conversationsLoading}
            className="flex items-center space-x-2"
            aria-label={t("common.refresh")}
          >
            <RefreshCcw
              className={`w-4 h-4 ${
                conversationsLoading ? "animate-spin" : ""
              }`}
            />
            <span>{t("common.refresh")}</span>
          </Button>
        </div>

        <div className="space-y-3">
          {conversations.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <MessageSquare className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>{t("taskConversations.empty.title")}</p>
              <p className="text-sm">
                {t("taskConversations.empty.description")}
              </p>
            </div>
          ) : (
            conversations.map((conversation, index) => (
              <ConversationItem
                key={conversation.id}
                conversation={conversation}
                task={task}
                taskId={taskId}
                isExpanded={isConversationExpanded(conversation.id)}
                shouldShowExpandButton={shouldShowExpandButton(
                  conversation.content
                )}
                isLatest={index === conversations.length - 1}
                onToggleExpanded={toggleExpanded}
                onViewDetails={handleViewDetails}
                onViewGitDiff={onViewConversationGitDiff}
                onViewLogs={onViewConversationLogs}
                onRetry={onRetryConversation}
                onCancel={onCancelConversation}
                onDelete={onDeleteConversation}
              />
            ))
          )}
        </div>
      </div>
    );
  }
);

ConversationList.displayName = "ConversationList";
