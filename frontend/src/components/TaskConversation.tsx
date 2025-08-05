import { useState, useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
  Send,
  User,
  Clock,
  CheckCircle,
  XCircle,
  RotateCcw,
  RefreshCw,
  MessageSquare,
  Play,
  Trash2,
  GitCommit,
  ChevronDown,
  ChevronUp,
  Copy,
} from "lucide-react";
import type {
  TaskConversation as TaskConversationInterface,
  ConversationStatus,
  ConversationFormData,
} from "@/types/task-conversation";
import type { TaskStatus } from "@/types/task";

interface TaskConversationProps {
  conversations: TaskConversationInterface[];
  selectedConversationId: number | null;
  loading: boolean;
  taskStatus?: TaskStatus;
  onSendMessage: (data: ConversationFormData) => Promise<void>;
  onRefresh: () => void;
  onSelectConversation: (conversationId: number) => void;
  onConversationStatusChange?: (
    conversationId: number,
    newStatus: ConversationStatus
  ) => void;
  onDeleteConversation?: (conversationId: number) => Promise<void>;
  onViewConversationGitDiff?: (conversationId: number) => void;
}

export function TaskConversation({
  conversations,
  selectedConversationId,
  loading,
  taskStatus,
  onSendMessage,
  onRefresh,
  onSelectConversation,
  onDeleteConversation,
  onViewConversationGitDiff,
}: TaskConversationProps) {
  const { t } = useTranslation();
  const [newMessage, setNewMessage] = useState("");

  const getStatusText = (status: ConversationStatus) => {
    const statusMap = {
      'pending': t("taskConversations.status.pending"),
      'running': t("taskConversations.status.running"),
      'success': t("taskConversations.status.success"),
      'failed': t("taskConversations.status.failed"),
      'cancelled': t("taskConversations.status.cancelled"),
    } as const;
    
    return statusMap[status] || status;
  };
  const [sending, setSending] = useState(false);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [conversationToDelete, setConversationToDelete] = useState<
    number | null
  >(null);
  const [expandedConversations, setExpandedConversations] = useState<
    Set<number>
  >(new Set());
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [conversations]);

  const getStatusIcon = (status: ConversationStatus) => {
    switch (status) {
      case "pending":
        return <Clock className="w-3 h-3" />;
      case "running":
        return <Play className="w-3 h-3" />;
      case "success":
        return <CheckCircle className="w-3 h-3" />;
      case "failed":
        return <XCircle className="w-3 h-3" />;
      case "cancelled":
        return <RotateCcw className="w-3 h-3" />;
      default:
        return <Clock className="w-3 h-3" />;
    }
  };

  const hasPendingOrRunningConversations = () => {
    return conversations.some(
      (conv) => conv.status === "pending" || conv.status === "running"
    );
  };

  const isTaskCompleted = () => {
    return taskStatus === "done" || taskStatus === "cancelled";
  };

  const canSendMessage = () => {
    return !hasPendingOrRunningConversations() && !isTaskCompleted();
  };

  const getDisabledReason = () => {
    if (isTaskCompleted()) {
      return t("taskConversations.taskCompletedMessage");
    }
    if (hasPendingOrRunningConversations()) {
      return t("taskConversations.hasPendingMessage");
    }
    return "";
  };

  const handleSendMessage = async () => {
    if (!newMessage.trim()) return;
    if (!canSendMessage()) return;

    setSending(true);
    try {
      await onSendMessage({
        content: newMessage.trim(),
      });
      setNewMessage("");
    } catch (error) {
      console.error("Failed to send message:", error);
    } finally {
      setSending(false);
    }
  };

  const handleDeleteConversation = (conversationId: number) => {
    if (!onDeleteConversation) return;
    setConversationToDelete(conversationId);
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = async () => {
    if (!onDeleteConversation || !conversationToDelete) return;

    try {
      setDeletingId(conversationToDelete);
      await onDeleteConversation(conversationToDelete);
    } catch (error) {
      console.error("Failed to delete conversation:", error);
    } finally {
      setDeletingId(null);
      setDeleteDialogOpen(false);
      setConversationToDelete(null);
    }
  };

  const handleCancelDelete = () => {
    setDeleteDialogOpen(false);
    setConversationToDelete(null);
  };

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const getStatusColor = (status: ConversationStatus) => {
    switch (status) {
      case "pending":
        return "bg-yellow-100 text-yellow-800";
      case "running":
        return "bg-blue-100 text-blue-800";
      case "success":
        return "bg-green-100 text-green-800";
      case "failed":
        return "bg-red-100 text-red-800";
      case "cancelled":
        return "bg-gray-100 text-gray-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  const isLatestConversation = (conversationId: number) => {
    if (conversations.length === 0) return false;
    const sortedConversations = [...conversations].sort(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    );
    return sortedConversations[0]?.id === conversationId;
  };

  const toggleExpanded = (conversationId: number) => {
    const newExpanded = new Set(expandedConversations);
    if (newExpanded.has(conversationId)) {
      newExpanded.delete(conversationId);
    } else {
      newExpanded.add(conversationId);
    }
    setExpandedConversations(newExpanded);
  };

  const isConversationExpanded = (conversationId: number) => {
    return expandedConversations.has(conversationId);
  };

  const shouldShowExpandButton = (content: string) => {
    return content.split("\n").length > 3 || content.length > 150;
  };

  const handleCopyContent = async (content: string) => {
    try {
      await navigator.clipboard.writeText(content);
      toast.success(t("common.copied_to_clipboard"));
    } catch (error) {
      console.error("Failed to copy:", error);
      toast.error(t("common.copy_failed"));
    }
  };

  return (
    <div className="space-y-6 h-full flex flex-col">
      <Card className="flex-1 flex flex-col">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <div>
            <CardTitle className="text-xl">
              {t("taskConversations.list.title")}
            </CardTitle>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            disabled={loading}
            className="flex items-center space-x-2 text-foreground hover:text-foreground"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
            <span>{t("common.refresh")}</span>
          </Button>
        </CardHeader>

        <CardContent className="flex-1 flex flex-col">
          <div className="space-y-3 flex-1 overflow-y-auto max-h-[2000px]">
            {conversations.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <MessageSquare className="w-12 h-12 mx-auto mb-4 opacity-50" />
                <p>{t("taskConversations.empty.title")}</p>
                <p className="text-sm">
                  {t("taskConversations.empty.description")}
                </p>
              </div>
            ) : (
              conversations.map((conversation) => (
                <div
                  key={conversation.id}
                  className={`p-4 rounded-lg border border-border cursor-pointer transition-colors ${
                    selectedConversationId === conversation.id
                      ? "border-primary bg-primary/10"
                      : "border-border"
                  }`}
                  onClick={() => onSelectConversation(conversation.id)}
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      <div className="p-1 rounded-full">
                        <User className="w-4 h-4" />
                      </div>
                      <span className="font-medium">
                        {t("taskConversations.message")}
                      </span>
                      <span className="text-xs text-gray-500">
                        {conversation.created_by}
                      </span>
                      {isLatestConversation(conversation.id) && (
                        <Badge variant="outline" className="text-xs px-1 py-0">
                          {t("taskConversations.latest")}
                        </Badge>
                      )}
                    </div>

                    <div className="flex items-center space-x-2">
                      <Badge className={getStatusColor(conversation.status)}>
                        <div className="flex items-center space-x-1">
                          {getStatusIcon(conversation.status)}
                          <span>
                            {getStatusText(conversation.status)}
                          </span>
                        </div>
                      </Badge>
                      {conversation.status !== "running" &&
                        isLatestConversation(conversation.id) &&
                        onDeleteConversation && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDeleteConversation(conversation.id);
                            }}
                            disabled={deletingId === conversation.id}
                            className="h-6 w-6 p-0 text-red-500 hover:text-red-700 hover:bg-red-50"
                            title={t("taskConversations.actions.delete")}
                          >
                            <Trash2 className="w-3 h-3" />
                          </Button>
                        )}
                    </div>
                  </div>

                  <div className="mt-2 text-sm">
                    <div className="flex justify-between items-start gap-2">
                      <div
                        className={`flex-1 whitespace-pre-wrap ${
                          isConversationExpanded(conversation.id)
                            ? ""
                            : shouldShowExpandButton(conversation.content)
                            ? "line-clamp-3"
                            : ""
                        }`}
                      >
                        {conversation.content}
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleCopyContent(conversation.content);
                        }}
                        className="h-6 w-6 p-0 text-gray-500 hover:text-gray-700 hover:bg-gray-100 shrink-0"
                        title={t("common.copy")}
                      >
                        <Copy className="w-3 h-3" />
                      </Button>
                    </div>
                    {shouldShowExpandButton(conversation.content) && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          toggleExpanded(conversation.id);
                        }}
                        className="mt-1 h-6 px-1 text-xs text-gray-500 hover:bg-blue-50"
                      >
                        {isConversationExpanded(conversation.id) ? (
                          <>
                            <ChevronUp className="w-3 h-3 mr-1" />
                            {t("common.showLess")}
                          </>
                        ) : (
                          <>
                            <ChevronDown className="w-3 h-3 mr-1" />
                            {t("common.showMore")}
                          </>
                        )}
                      </Button>
                    )}
                  </div>

                  <div className="flex items-center justify-between mt-2">
                    <div className="text-xs text-gray-400">
                      {formatTime(conversation.created_at)}
                    </div>
                    {conversation.commit_hash && (
                      <div className="flex items-center space-x-2">
                        <div className="flex items-center text-xs text-gray-500">
                          <GitCommit className="w-3 h-3 mr-1" />
                          <span className="font-mono">
                            {conversation.commit_hash.substring(0, 8)}
                          </span>
                        </div>
                        {onViewConversationGitDiff && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              onViewConversationGitDiff(conversation.id);
                            }}
                            className="h-6 px-2 text-xs text-blue-600 hover:text-blue-800 hover:bg-blue-50"
                            title={t("taskConversations.actions.viewGitDiff")}
                          >
                            {t("taskConversations.actions.viewChanges")}
                          </Button>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              ))
            )}
            <div ref={messagesEndRef} />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">
            {t("taskConversations.newMessage")}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex flex-col gap-3">
              <label className="text-sm font-medium">
                {t("taskConversations.content")}:
              </label>
              <Textarea
                className="min-h-[120px] resize-none"
                placeholder={t("taskConversations.contentPlaceholder")}
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
                    e.preventDefault();
                    handleSendMessage();
                  }
                }}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="text-xs text-muted-foreground">
                {t("taskConversations.shortcut")}
              </div>
              <div className="flex flex-1 justify-end">
                <Button
                  onClick={handleSendMessage}
                  disabled={!newMessage.trim() || sending || !canSendMessage()}
                  className="flex items-center space-x-2"
                >
                  <Send className="w-4 h-4" />
                  <span>
                    {sending ? t("common.sending") : t("common.send")}
                  </span>
                </Button>
              </div>
            </div>

            {!canSendMessage() && (
              <div className="text-sm text-amber-600 bg-amber-50 p-3 rounded-lg border border-amber-200">
                {getDisabledReason()}
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-foreground">
              {t("taskConversations.delete_confirm_title")}
            </DialogTitle>
            <DialogDescription className="text-muted-foreground">
              {t("taskConversations.deleteConfirm")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              className="text-foreground hover:text-foreground"
              onClick={handleCancelDelete}
            >
              {t("common.cancel")}
            </Button>
            <Button
              variant="destructive"
              className="text-foreground"
              onClick={handleConfirmDelete}
              disabled={deletingId === conversationToDelete}
            >
              {t("common.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
