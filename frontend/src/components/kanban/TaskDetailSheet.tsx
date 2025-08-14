import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  GitBranch,
  User,
  MessageSquare,
  FileText,
  Eye,
  MoreHorizontal,
  Clock,
  Monitor,
  Send,
  Calendar,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Textarea } from "@/components/ui/textarea";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { PushBranchDialog } from "@/components/PushBranchDialog";
import { useTaskConversations } from "@/hooks/useTaskConversations";
import type { Task, TaskStatus } from "@/types/task";

interface TaskDetailSheetProps {
  task: Task | null;
  isOpen: boolean;
  onClose: () => void;
}

export function TaskDetailSheet({
  task,
  isOpen,
  onClose,
}: TaskDetailSheetProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [isPushDialogOpen, setIsPushDialogOpen] = useState(false);

  const {
    conversations,
    conversationsLoading,
    newMessage,
    setNewMessage,
    executionTime,
    setExecutionTime,
    sending,
    loadConversations,
    handleSendMessage,
    toggleExpanded,
    isConversationExpanded,
    canSendMessage,
    shouldShowExpandButton,
    isTaskCompleted,
    hasPendingOrRunningConversations,
  } = useTaskConversations(task);

  const getStatusBadgeClass = (status: TaskStatus) => {
    switch (status) {
      case "todo":
        return "bg-gray-100 text-gray-800";
      case "in_progress":
        return "bg-blue-100 text-blue-800";
      case "done":
        return "bg-green-100 text-green-800";
      case "cancelled":
        return "bg-red-100 text-red-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  const getConversationStatusColor = (status: string) => {
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

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const handleViewConversationGitDiff = (conversationId: number) => {
    if (!task) return;
    navigate(`/tasks/${task.id}/conversations/${conversationId}/git-diff`);
  };

  const handlePushBranch = () => {
    setIsPushDialogOpen(true);
  };

  const handleViewTaskGitDiff = () => {
    if (!task) return;
    navigate(`/tasks/${task.id}/git-diff`);
  };

  if (!task) return null;

  return (
    <Sheet open={isOpen} onOpenChange={onClose}>
      <SheetContent className="w-full sm:w-[800px] sm:max-w-[800px] flex flex-col">
        <SheetHeader className="border-b sticky top-0 bg-background">
          <SheetTitle className="text-foreground font-semibold">
            {task.title}
          </SheetTitle>
          <SheetDescription className="text-muted-foreground text-sm">
            {t("tasks.details")}
          </SheetDescription>
        </SheetHeader>

        <div className="flex-1 flex flex-col p-4 space-y-6 overflow-y-auto">
          {/* 基础信息板块 */}
          <div className="space-y-4">
            <h3 className="font-medium text-foreground text-lg flex items-center gap-2">
              <FileText className="h-5 w-5" />
              {t("tasks.tabs.basic")}
            </h3>
            
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="font-medium text-foreground">
                  {t("tasks.status.label")}:
                </span>
                <Badge
                  className={`ml-2 ${getStatusBadgeClass(task.status)}`}
                >
                  {t(`tasks.status.${task.status}`)}
                </Badge>
              </div>

              <div className="flex items-center">
                <GitBranch className="h-4 w-4 mr-1" />
                <span className="font-medium text-foreground">
                  {t("tasks.workBranch")}:
                </span>
                <span className="ml-2 font-mono text-xs">
                  {task.work_branch}
                </span>
              </div>

              <div className="flex items-center">
                <GitBranch className="h-4 w-4 mr-1" />
                <span className="font-medium text-foreground">
                  {t("tasks.startBranch")}:
                </span>
                <span className="ml-2 font-mono text-xs">
                  {task.start_branch}
                </span>
              </div>

              <div className="flex items-center">
                <Monitor className="h-4 w-4 mr-1" />
                <span className="font-medium text-foreground">
                  {t("tasks.environment")}:
                </span>
                <span className="ml-2">
                  {task.dev_environment?.name || "-"}
                </span>
              </div>

              <div className="flex items-center">
                <Clock className="h-4 w-4 mr-1" />
                <span className="font-medium text-foreground">
                  {t("tasks.executionTime")}:
                </span>
                <span className="ml-2">
                  {task.latest_execution_time 
                    ? formatTime(task.latest_execution_time)
                    : t("common.notSet")
                  }
                </span>
              </div>

              <div className="flex items-center">
                <User className="h-4 w-4 mr-1" />
                <span className="font-medium text-foreground">
                  {t("tasks.createdBy")}:
                </span>
                <span className="ml-2">{task.created_by}</span>
              </div>

              <div className="flex items-center">
                <Calendar className="h-4 w-4 mr-1" />
                <span className="font-medium text-foreground">
                  {t("tasks.createdAt")}:
                </span>
                <span className="ml-2">
                  {new Date(task.created_at).toLocaleDateString()}
                </span>
              </div>
            </div>

            {/* Actions */}
            <div className="border-t pt-4">
              <h4 className="font-medium text-foreground mb-3">
                {t("tasks.actions.title")}
              </h4>
              <div className="flex flex-wrap gap-3">
                <Button
                  onClick={handlePushBranch}
                  className="flex items-center gap-2"
                  disabled={
                    task.status === "done" || task.status === "cancelled"
                  }
                >
                  <GitBranch className="h-4 w-4" />
                  {t("tasks.actions.pushBranch")}
                </Button>

                <Button
                  onClick={handleViewTaskGitDiff}
                  variant="outline"
                  className="flex items-center gap-2"
                >
                  <Eye className="h-4 w-4" />
                  {t("tasks.actions.viewGitDiff")}
                </Button>
              </div>
            </div>
          </div>

          {/* 对话信息板块 */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="font-medium text-foreground text-lg flex items-center gap-2">
                <MessageSquare className="h-5 w-5" />
                {t("taskConversations.list.title")}
                {task.conversation_count > 0 && (
                  <Badge variant="outline" className="ml-1 text-xs">
                    {task.conversation_count}
                  </Badge>
                )}
              </h3>
              <Button
                variant="outline"
                size="sm"
                onClick={loadConversations}
                disabled={conversationsLoading}
                className="flex items-center space-x-2"
              >
                <MessageSquare className={`w-4 h-4 ${conversationsLoading ? "animate-spin" : ""}`} />
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
                conversations.map((conversation) => (
                  <div
                    key={conversation.id}
                    className="p-4 rounded-lg border border-border bg-card"
                  >
                    <div className="flex items-start justify-between gap-4">
                      {/* 左侧对话内容 */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center space-x-2 mb-2">
                          <User className="w-4 h-4" />
                          <span className="font-medium text-sm">
                            {conversation.created_by}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {formatTime(conversation.created_at)}
                          </span>
                        </div>
                        <div
                          className={`text-sm whitespace-pre-wrap ${
                            isConversationExpanded(conversation.id)
                              ? ""
                              : shouldShowExpandButton(conversation.content)
                              ? "line-clamp-3"
                              : ""
                          }`}
                        >
                          {conversation.content}
                        </div>
                        {shouldShowExpandButton(conversation.content) && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => toggleExpanded(conversation.id)}
                            className="mt-1 h-6 px-1 text-xs text-muted-foreground hover:bg-muted"
                          >
                            {isConversationExpanded(conversation.id) 
                              ? t("common.showLess") 
                              : t("common.showMore")
                            }
                          </Button>
                        )}
                      </div>

                      {/* 右侧状态和菜单 */}
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
                            >
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() => navigate(`/tasks/${task.id}/conversations/${conversation.id}`)}
                            >
                              <Eye className="mr-2 h-4 w-4" />
                              {t("taskConversations.actions.viewDetails")}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() => handleViewConversationGitDiff(conversation.id)}
                            >
                              <FileText className="mr-2 h-4 w-4" />
                              {t("taskConversations.actions.viewGitDiff")}
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* 发送对话消息板块 */}
          <div className="space-y-4 border-t pt-4">
            <h3 className="font-medium text-foreground text-lg flex items-center gap-2">
              <Send className="h-5 w-5" />
              {t("taskConversations.newMessage")}
            </h3>
            
            <div className="space-y-4">
              <div className="space-y-2">
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

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  {t("taskConversations.executionTime")}:
                </label>
                <DateTimePicker
                  value={executionTime}
                  onChange={setExecutionTime}
                  placeholder={t("taskConversations.executionTimePlaceholder")}
                  label=""
                />
                <p className="text-xs text-muted-foreground">
                  {t("taskConversations.executionTimeHint")}
                </p>
              </div>

              <div className="flex items-center justify-between">
                <div className="text-xs text-muted-foreground">
                  {t("taskConversations.shortcut")}
                </div>
                <Button
                  onClick={handleSendMessage}
                  disabled={!newMessage.trim() || sending || !canSendMessage()}
                  className="flex items-center space-x-2"
                >
                  <MessageSquare className="w-4 h-4" />
                  <span>
                    {sending ? t("common.sending") : t("common.send")}
                  </span>
                </Button>
              </div>

              {!canSendMessage() && (
                <div className="text-sm text-amber-600 bg-amber-50 p-3 rounded-lg border border-amber-200">
                  {isTaskCompleted() 
                    ? t("taskConversations.taskCompletedMessage")
                    : t("taskConversations.hasPendingMessage")
                  }
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Push Branch Dialog */}
        <PushBranchDialog
          open={isPushDialogOpen}
          onOpenChange={setIsPushDialogOpen}
          task={task}
          onSuccess={() => {
            // Could refresh task data here if needed
          }}
        />
      </SheetContent>
    </Sheet>
  );
}
