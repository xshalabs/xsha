import { useState, useEffect, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { User, Settings, Activity, BarChart3, ChevronDown, ChevronUp } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { taskConversationsApi } from "@/lib/api/task-conversations";
import { getConversationStatusColor, formatTime } from "@/components/kanban/task-detail/utils";

interface ConversationDetailModalProps {
  conversationId: number | null;
  isOpen: boolean;
  onClose: () => void;
}

export const ConversationDetailModal: React.FC<ConversationDetailModalProps> = ({
  conversationId,
  isOpen,
  onClose,
}) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [details, setDetails] = useState<any>(null);
  const [isContentExpanded, setIsContentExpanded] = useState(false);

  useEffect(() => {
    if (isOpen && conversationId) {
      loadConversationDetails();
    }
  }, [isOpen, conversationId]);

  const loadConversationDetails = async () => {
    if (!conversationId) return;

    setLoading(true);
    try {
      const response = await taskConversationsApi.getDetails(conversationId);
      setDetails(response.data);
    } catch (error) {
      console.error("Failed to load conversation details:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setDetails(null);
    setIsContentExpanded(false);
    onClose();
  };

  // 解析Usage Statistics获取token信息
  const parseUsageStats = useMemo(() => {
    if (!details?.result?.usage) return null;
    
    try {
      const usage = JSON.parse(details.result.usage);
      return {
        inputTokens: usage.input_tokens || 0,
        outputTokens: usage.output_tokens || 0,
      };
    } catch (error) {
      return null;
    }
  }, [details?.result?.usage]);

  // 格式化token数量为人类友好显示
  const formatTokens = (count: number): string => {
    if (count >= 1000000) {
      return `${(count / 1000000).toFixed(1)}M`;
    } else if (count >= 1000) {
      return `${(count / 1000).toFixed(1)}K`;
    }
    return count.toString();
  };

  // 检查Content是否需要展开功能
  const shouldShowExpandButton = (content: string): boolean => {
    const lines = content.split('\n');
    return lines.length > 3;
  };

  // 获取显示的Content内容
  const getDisplayContent = (content: string): string => {
    if (!shouldShowExpandButton(content) || isContentExpanded) {
      return content;
    }
    const lines = content.split('\n');
    return lines.slice(0, 3).join('\n');
  };

  const renderConversationInfo = () => {
    if (!details?.conversation) return null;

    const conversation = details.conversation;

    return (
      <Card className="w-full min-w-0">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <User className="h-5 w-5" />
            {t("taskConversations.details.conversationInfo")}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4 w-full min-w-0">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-w-0">
            <span className="text-sm text-muted-foreground min-w-0">
              {t("taskConversations.details.status")}:
            </span>
            <Badge className={`${getConversationStatusColor(conversation.status)} self-start sm:self-auto`}>
              {t(`taskConversations.status.${conversation.status}`)}
            </Badge>
          </div>

          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-w-0">
            <span className="text-sm text-muted-foreground min-w-0">
              {t("taskConversations.details.createdBy")}:
            </span>
            <span className="font-medium break-words">{conversation.created_by}</span>
          </div>

          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-w-0">
            <span className="text-sm text-muted-foreground min-w-0">
              {t("taskConversations.details.createdAt")}:
            </span>
            <span className="text-sm break-words">{formatTime(conversation.created_at)}</span>
          </div>

          {conversation.execution_time && (
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-w-0">
              <span className="text-sm text-muted-foreground min-w-0">
                {t("taskConversations.details.executionTime")}:
              </span>
              <span className="text-sm break-words">{formatTime(conversation.execution_time)}</span>
            </div>
          )}

          {conversation.commit_hash && (
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-w-0">
              <span className="text-sm text-muted-foreground min-w-0">
                {t("taskConversations.details.commitHash")}:
              </span>
              <span className="text-sm font-mono text-xs break-all">
                {conversation.commit_hash.substring(0, 8)}
              </span>
            </div>
          )}

          <Separator />

          <div className="w-full min-w-0">
            <span className="text-sm text-muted-foreground">
              {t("taskConversations.details.content")}:
            </span>
            <div className="mt-2 p-3 bg-muted rounded-md text-sm whitespace-pre-wrap break-words w-full min-w-0 overflow-hidden">
              {getDisplayContent(conversation.content)}
            </div>
            {shouldShowExpandButton(conversation.content) && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsContentExpanded(!isContentExpanded)}
                className="mt-2 h-8 px-2 text-xs text-muted-foreground hover:bg-muted"
              >
                {isContentExpanded ? (
                  <>
                    <ChevronUp className="h-3 w-3 mr-1" />
                    {t("common.showLess")}
                  </>
                ) : (
                  <>
                    <ChevronDown className="h-3 w-3 mr-1" />
                    {t("common.showMore")}
                  </>
                )}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    );
  };

  const renderResultInfo = () => {
    if (!details?.result) return null;

    const result = details.result;

    return (
      <Card className="w-full min-w-0">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <BarChart3 className="h-5 w-5" />
            {t("taskConversations.details.executionResult")}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4 w-full min-w-0">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 w-full min-w-0">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-w-0">
              <span className="text-sm text-muted-foreground min-w-0">
                {t("taskConversations.details.resultType")}:
              </span>
              <Badge variant={result.is_error ? "destructive" : "default"} className="self-start sm:self-auto">
                {result.subtype}
              </Badge>
            </div>

            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-w-0">
              <span className="text-sm text-muted-foreground min-w-0">
                {t("taskConversations.details.duration")}:
              </span>
              <span className="text-sm font-medium">{(result.duration_ms / 1000).toFixed(2)}s</span>
            </div>

            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-w-0">
              <span className="text-sm text-muted-foreground min-w-0">
                {t("taskConversations.details.numTurns")}:
              </span>
              <span className="text-sm font-medium">{result.num_turns}</span>
            </div>

            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-w-0">
              <span className="text-sm text-muted-foreground min-w-0">
                {t("taskConversations.details.sessionId")}:
              </span>
              <span className="text-xs font-mono break-all">{result.session_id.substring(0, 8)}...</span>
            </div>

            {parseUsageStats && (
              <>
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-w-0">
                  <span className="text-sm text-muted-foreground min-w-0">
                    {t("taskConversations.details.inputTokens")}:
                  </span>
                  <span className="text-sm font-medium">{formatTokens(parseUsageStats.inputTokens)}</span>
                </div>

                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-w-0">
                  <span className="text-sm text-muted-foreground min-w-0">
                    {t("taskConversations.details.outputTokens")}:
                  </span>
                  <span className="text-sm font-medium">{formatTokens(parseUsageStats.outputTokens)}</span>
                </div>
              </>
            )}
          </div>

          <Separator />

          <div className="w-full min-w-0">
            <span className="text-sm text-muted-foreground">
              {t("taskConversations.details.result")}:
            </span>
            <div className="mt-2 p-3 bg-muted rounded-md text-sm whitespace-pre-wrap break-words w-full min-w-0 overflow-hidden max-h-60 overflow-y-auto">
              {result.result}
            </div>
          </div>
        </CardContent>
      </Card>
    );
  };

  const renderNoResult = () => (
    <Card className="w-full min-w-0">
      <CardContent className="flex items-center justify-center py-8 text-center w-full min-w-0">
        <div className="space-y-2 w-full min-w-0">
          <Activity className="h-12 w-12 mx-auto opacity-50" />
          <p className="text-muted-foreground">
            {t("taskConversations.details.noResult")}
          </p>
          <p className="text-sm text-muted-foreground">
            {t("taskConversations.details.noResultDescription")}
          </p>
        </div>
      </CardContent>
    </Card>
  );

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent 
        className="w-full max-w-[95vw] sm:max-w-[90vw] md:max-w-4xl lg:max-w-6xl xl:max-w-7xl max-h-[90vh] overflow-y-auto overflow-x-hidden p-4 sm:p-6"
      >
        <DialogHeader className="pb-4 w-full min-w-0">
          <DialogTitle className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 w-full min-w-0">
            <span className="flex items-center gap-2 min-w-0">
              <Settings className="h-5 w-5" />
              <span className="truncate">{t("taskConversations.details.title")}</span>
            </span>
          </DialogTitle>
          <DialogDescription className="text-left w-full min-w-0">
            {conversationId && (
              <>
                <span className="block sm:inline">ID: {conversationId}</span>
                <span className="hidden sm:inline"> | </span>
                <span className="block sm:inline break-words">{t("taskConversations.details.description")}</span>
              </>
            )}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 w-full min-w-0">
          {loading ? (
            <div className="space-y-4">
              <Skeleton className="h-48 w-full" />
              <Skeleton className="h-48 w-full" />
            </div>
          ) : (
            <>
              {renderConversationInfo()}
              {details?.result ? renderResultInfo() : renderNoResult()}
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};
