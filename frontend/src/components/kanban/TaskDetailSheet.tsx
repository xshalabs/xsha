import { useState, useCallback, memo, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { PushBranchDialog } from "@/components/PushBranchDialog";
import { TaskGitDiffModal } from "./TaskGitDiffModal";
import { ConversationGitDiffModal } from "./ConversationGitDiffModal";
import { ConversationDetailModal } from "@/components/ConversationDetailModal";
import { ConversationLogModal } from "./ConversationLogModal";
import { useTaskConversations } from "@/hooks/useTaskConversations";
import { taskExecutionLogsApi } from "@/lib/api/task-execution-logs";
import {
  TaskBasicInfo,
  TaskActions,
  ConversationList,
  NewMessageForm,
} from "./task-detail";
import type { Task } from "@/types/task";

interface TaskDetailSheetProps {
  task: Task | null;
  isOpen: boolean;
  onClose: () => void;
}

export const TaskDetailSheet = memo<TaskDetailSheetProps>(({
  task,
  isOpen,
  onClose,
}) => {
  const { t } = useTranslation();
  const [isPushDialogOpen, setIsPushDialogOpen] = useState(false);
  const [isTaskGitDiffModalOpen, setIsTaskGitDiffModalOpen] = useState(false);
  const [isConversationGitDiffModalOpen, setIsConversationGitDiffModalOpen] = useState(false);
  const [isConversationDetailModalOpen, setIsConversationDetailModalOpen] = useState(false);
  const [isConversationLogModalOpen, setIsConversationLogModalOpen] = useState(false);
  const [selectedConversation, setSelectedConversation] = useState<any>(null);
  const [selectedConversationId, setSelectedConversationId] = useState<number | null>(null);
  const [selectedLogConversationId, setSelectedLogConversationId] = useState<number | null>(null);
  const [retryConversationId, setRetryConversationId] = useState<number | null>(null);
  const [retrying, setRetrying] = useState(false);

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

  // Memoized event handlers
  const handleViewConversationGitDiff = useCallback((conversationId: number) => {
    if (!task) return;
    const conversation = conversations.find(c => c.id === conversationId);
    if (conversation) {
      setSelectedConversation(conversation);
      setIsConversationGitDiffModalOpen(true);
    }
  }, [task, conversations]);

  const handleViewConversationLogs = useCallback((conversationId: number) => {
    setSelectedLogConversationId(conversationId);
    setIsConversationLogModalOpen(true);
  }, []);

  const handlePushBranch = useCallback(() => {
    setIsPushDialogOpen(true);
  }, []);

  const handleViewTaskGitDiff = useCallback(() => {
    if (!task) return;
    setIsTaskGitDiffModalOpen(true);
  }, [task]);

  const handleCloseSheet = useCallback(() => {
    onClose();
  }, [onClose]);

  const handleClosePushDialog = useCallback(() => {
    setIsPushDialogOpen(false);
  }, []);

  const handleCloseTaskGitDiff = useCallback(() => {
    setIsTaskGitDiffModalOpen(false);
  }, []);

  const handleCloseConversationGitDiff = useCallback(() => {
    setIsConversationGitDiffModalOpen(false);
    setSelectedConversation(null);
  }, []);

  const handleViewConversationDetails = useCallback((conversationId: number) => {
    setSelectedConversationId(conversationId);
    setIsConversationDetailModalOpen(true);
  }, []);

  const handleCloseConversationDetails = useCallback(() => {
    setIsConversationDetailModalOpen(false);
    setSelectedConversationId(null);
  }, []);

  const handleCloseConversationLogs = useCallback(() => {
    setIsConversationLogModalOpen(false);
    setSelectedLogConversationId(null);
  }, []);

  const handleRetryConversation = useCallback((conversationId: number) => {
    setRetryConversationId(conversationId);
  }, []);

  const handleConfirmRetry = useCallback(async () => {
    if (!retryConversationId) return;
    
    setRetrying(true);
    try {
      await taskExecutionLogsApi.retryExecution(retryConversationId);
      toast.success(t("taskConversations.execution.actions.retry") + " " + t("common.success"));
      // Refresh conversations to show updated status
      await loadConversations();
    } catch (error) {
      console.error("Failed to retry conversation:", error);
      toast.error(t("common.retry") + " " + t("common.failed"));
    } finally {
      setRetrying(false);
      setRetryConversationId(null);
    }
  }, [retryConversationId, t, loadConversations]);

  const handleCancelRetry = useCallback(() => {
    setRetryConversationId(null);
  }, []);

  // Memoized computed values
  const canSend = useMemo(() => canSendMessage(), [canSendMessage]);
  const taskCompleted = useMemo(() => isTaskCompleted(), [isTaskCompleted]);
  const hasPendingConversations = useMemo(() => hasPendingOrRunningConversations(), [hasPendingOrRunningConversations]);

  if (!task) return null;

  return (
    <>
      <Sheet open={isOpen} onOpenChange={handleCloseSheet}>
        <SheetContent 
          className="w-full sm:w-[800px] sm:max-w-[800px] flex flex-col"
          aria-describedby="task-detail-description"
        >
          <SheetHeader className="border-b sticky top-0 bg-background z-10">
            <SheetTitle className="text-foreground font-semibold">
              {task.title}
            </SheetTitle>
            <SheetDescription 
              id="task-detail-description"
              className="text-muted-foreground text-sm"
            >
              {t("tasks.details")}
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 flex flex-col space-y-6 overflow-y-auto">
            {/* 基础信息板块 */}
            <TaskBasicInfo task={task} />

            {/* Actions */}
            <TaskActions
              task={task}
              onPushBranch={handlePushBranch}
              onViewGitDiff={handleViewTaskGitDiff}
            />

            {/* 对话信息板块 */}
            <ConversationList
              conversations={conversations}
              conversationsLoading={conversationsLoading}
              conversationCount={task.conversation_count}
              taskId={task.id}
              onLoadConversations={loadConversations}
              onViewConversationGitDiff={handleViewConversationGitDiff}
              onViewConversationDetails={handleViewConversationDetails}
              onViewConversationLogs={handleViewConversationLogs}
              onRetryConversation={handleRetryConversation}
              toggleExpanded={toggleExpanded}
              isConversationExpanded={isConversationExpanded}
              shouldShowExpandButton={shouldShowExpandButton}
            />

            {/* 发送对话消息板块 */}
            <NewMessageForm
              newMessage={newMessage}
              executionTime={executionTime}
              sending={sending}
              canSendMessage={canSend}
              isTaskCompleted={taskCompleted}
              _hasPendingOrRunningConversations={hasPendingConversations}
              onMessageChange={setNewMessage}
              onExecutionTimeChange={setExecutionTime}
              onSendMessage={handleSendMessage}
            />
          </div>
        </SheetContent>
      </Sheet>

      {/* Dialogs and Modals */}
      <PushBranchDialog
        open={isPushDialogOpen}
        onOpenChange={handleClosePushDialog}
        task={task}
        onSuccess={() => {
          // Could refresh task data here if needed
        }}
      />

      <TaskGitDiffModal
        isOpen={isTaskGitDiffModalOpen}
        onClose={handleCloseTaskGitDiff}
        task={task}
      />

      <ConversationGitDiffModal
        isOpen={isConversationGitDiffModalOpen}
        onClose={handleCloseConversationGitDiff}
        conversation={selectedConversation}
      />

      <ConversationDetailModal
        conversationId={selectedConversationId}
        isOpen={isConversationDetailModalOpen}
        onClose={handleCloseConversationDetails}
      />

      <ConversationLogModal
        conversationId={selectedLogConversationId}
        isOpen={isConversationLogModalOpen}
        onClose={handleCloseConversationLogs}
      />

      <AlertDialog open={retryConversationId !== null} onOpenChange={handleCancelRetry}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("taskConversations.execution.retry_confirm_title")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("taskConversations.execution.retry_confirm")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={retrying}>
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirmRetry} disabled={retrying}>
              {retrying ? t("common.processing") : t("common.confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
});

TaskDetailSheet.displayName = "TaskDetailSheet";