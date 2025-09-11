import { memo, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { MessageSquare } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { ConversationItem } from "./ConversationItem";
import { StatusDot } from "./StatusDot";

interface ConversationListProps {
  conversations: any[];
  conversationCount: number;
  taskId: number;
  projectId: number;
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
    conversationCount,
    taskId,
    projectId,
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
            <>
              {conversations.map((conversation, index) => (
                <ConversationItem
                  key={conversation.id}
                  conversation={conversation}
                  taskId={taskId}
                  projectId={projectId}
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
              ))}
              
              {/* Status Legend */}
              <div className="mt-6 pt-4 border-t border-border">
                <div className="text-xs text-muted-foreground mb-3 font-medium">
                  {t("taskConversations.statusLegend")}:
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-x-4 gap-y-2 text-xs">
                  <div className="flex items-center space-x-2 min-w-0">
                    <StatusDot status="pending" />
                    <span className="text-muted-foreground truncate">
                      {t("taskConversations.status.pending")}
                    </span>
                  </div>
                  <div className="flex items-center space-x-2 min-w-0">
                    <StatusDot status="running" />
                    <span className="text-muted-foreground truncate">
                      {t("taskConversations.status.running")}
                    </span>
                  </div>
                  <div className="flex items-center space-x-2 min-w-0">
                    <StatusDot status="success" />
                    <span className="text-muted-foreground truncate">
                      {t("taskConversations.status.success")}
                    </span>
                  </div>
                  <div className="flex items-center space-x-2 min-w-0">
                    <StatusDot status="failed" />
                    <span className="text-muted-foreground truncate">
                      {t("taskConversations.status.failed")}
                    </span>
                  </div>
                  <div className="flex items-center space-x-2 min-w-0">
                    <StatusDot status="cancelled" />
                    <span className="text-muted-foreground truncate">
                      {t("taskConversations.status.cancelled")}
                    </span>
                  </div>
                  <div className="flex items-center space-x-2 min-w-0">
                    <div className="w-3 h-3 rounded-full bg-purple-500" />
                    <span className="text-muted-foreground truncate">
                      {t("taskConversations.status.scheduled")}
                    </span>
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    );
  }
);

ConversationList.displayName = "ConversationList";
