import { useState, useCallback, memo, useMemo, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Edit3 } from "lucide-react";
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
import { Button } from "@/components/ui/button";
import { PushBranchDialog } from "@/components/PushBranchDialog";
import { TaskGitDiffModal } from "./TaskGitDiffModal";
import { ConversationGitDiffModal } from "./ConversationGitDiffModal";
import { ConversationDetailModal } from "@/components/ConversationDetailModal";
import { ConversationLogModal } from "./ConversationLogModal";
import { TaskTitleEditDialog } from "./TaskTitleEditDialog";
import { useTaskConversations } from "@/hooks/useTaskConversations";
import { taskConversationsApi } from "@/lib/api/task-conversations";
import { tasksApi } from "@/lib/api/tasks";
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
  onTaskDeleted?: () => void;
}

export const TaskDetailSheet = memo<TaskDetailSheetProps>(({
  task,
  isOpen,
  onClose,
  onTaskDeleted,
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
  const [cancelConversationId, setCancelConversationId] = useState<number | null>(null);
  const [cancelling, setCancelling] = useState(false);
  const [deleteConversationId, setDeleteConversationId] = useState<number | null>(null);
  const [deletingConversation, setDeletingConversation] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [isTitleEditDialogOpen, setIsTitleEditDialogOpen] = useState(false);

  const {
    conversations,
    newMessage,
    setNewMessage,
    executionTime,
    setExecutionTime,
    model,
    setModel,
    isPlanMode,
    setIsPlanMode,
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

  // Auto-refresh conversations every second when sheet is open
  useEffect(() => {
    if (!isOpen || !task) return;
    
    const interval = setInterval(() => {
      // Only refresh if not currently sending a message
      if (!sending) {
        loadConversations();
      }
    }, 1000);
    
    return () => clearInterval(interval);
  }, [isOpen, task, sending, loadConversations]);

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
    if (!retryConversationId || !task) return;
    
    setRetrying(true);
    try {
      await taskConversationsApi.retryExecution(task.project_id, task.id, retryConversationId);
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
  }, [retryConversationId, task, t, loadConversations]);

  const handleCancelRetry = useCallback(() => {
    setRetryConversationId(null);
  }, []);

  const handleCancelConversation = useCallback((conversationId: number) => {
    setCancelConversationId(conversationId);
  }, []);

  const handleConfirmCancel = useCallback(async () => {
    if (!cancelConversationId || !task) return;
    
    setCancelling(true);
    try {
      await taskConversationsApi.cancelExecution(task.project_id, task.id, cancelConversationId);
      toast.success(t("taskConversations.execution.actions.cancel") + " " + t("common.success"));
      // Refresh conversations to show updated status
      await loadConversations();
    } catch (error) {
      console.error("Failed to cancel conversation:", error);
      toast.error(t("common.cancel") + " " + t("common.failed"));
    } finally {
      setCancelling(false);
      setCancelConversationId(null);
    }
  }, [cancelConversationId, task, t, loadConversations]);

  const handleCancelCancel = useCallback(() => {
    setCancelConversationId(null);
  }, []);

  const handleDeleteConversation = useCallback((conversationId: number) => {
    setDeleteConversationId(conversationId);
  }, []);

  const handleConfirmDeleteConversation = useCallback(async () => {
    if (!deleteConversationId || !task) return;
    
    setDeletingConversation(true);
    try {
      await taskConversationsApi.delete(task.project_id, task.id, deleteConversationId);
      toast.success(t("taskConversations.delete.deleteSuccess"));
      // Refresh conversations to show updated list
      await loadConversations();
    } catch (error) {
      console.error("Failed to delete conversation:", error);
      
      // Extract specific error message from API response
      let errorMessage = t("taskConversations.delete.deleteFailed");
      if (error instanceof Error) {
        errorMessage = error.message;
      } else if (typeof error === "string") {
        errorMessage = error;
      } else if (error && typeof error === "object" && "message" in error) {
        errorMessage = String(error.message);
      }
      
      toast.error(errorMessage);
    } finally {
      setDeletingConversation(false);
      setDeleteConversationId(null);
    }
  }, [deleteConversationId, task, t, loadConversations]);

  const handleCancelDeleteConversation = useCallback(() => {
    setDeleteConversationId(null);
  }, []);

  const handleDeleteTask = useCallback(() => {
    setIsDeleteDialogOpen(true);
  }, []);

  const handleConfirmDelete = useCallback(async () => {
    if (!task) return;
    
    setDeleting(true);
    try {
      await tasksApi.delete(task.project_id, task.id);
      toast.success(t("tasks.delete.deleteSuccess"));
      // Close the sheet first
      onClose();
      // Then notify parent about task deletion
      onTaskDeleted?.();
    } catch (error) {
      console.error("Failed to delete task:", error);
      
      // Extract specific error message from API response
      let errorMessage = t("tasks.delete.deleteFailed");
      if (error instanceof Error) {
        errorMessage = error.message;
      } else if (typeof error === "string") {
        errorMessage = error;
      } else if (error && typeof error === "object" && "message" in error) {
        errorMessage = String(error.message);
      }
      
      toast.error(errorMessage);
    } finally {
      setDeleting(false);
      setIsDeleteDialogOpen(false);
    }
  }, [task, t, onClose, onTaskDeleted]);

  const handleCancelDelete = useCallback(() => {
    setIsDeleteDialogOpen(false);
  }, []);

  const handleOpenTitleEditDialog = useCallback(() => {
    setIsTitleEditDialogOpen(true);
  }, []);

  const handleCloseTitleEditDialog = useCallback(() => {
    setIsTitleEditDialogOpen(false);
  }, []);

  const handleTitleUpdateSuccess = useCallback((updatedTitle: string) => {
    if (task) {
      task.title = updatedTitle;
    }
  }, [task]);

  // Memoized computed values
  const canSend = useMemo(() => canSendMessage(), [canSendMessage]);
  const taskCompleted = useMemo(() => isTaskCompleted(), [isTaskCompleted]);
  const hasPendingConversations = useMemo(() => hasPendingOrRunningConversations(), [hasPendingOrRunningConversations]);

  if (!task) return null;

  return (
    <>
      <Sheet open={isOpen} onOpenChange={handleCloseSheet}>
        <SheetContent 
          className="w-full sm:w-[800px] sm:max-w-[800px] flex flex-col focus:outline-none focus-visible:outline-none [&[data-state=open]]:focus:outline-none [&[data-state=open]]:focus-visible:outline-none"
          aria-describedby="task-detail-description"
          style={{ outline: "none" }}
        >
          <SheetHeader className="border-b sticky top-0 bg-background z-10">
            <div className="w-full pr-12">
              <div className="flex items-center gap-2">
                <SheetTitle className="text-foreground font-semibold truncate">
                  {task.title}
                </SheetTitle>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleOpenTitleEditDialog}
                  className="h-6 w-6 p-0 text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:text-gray-300 dark:hover:bg-gray-800 shrink-0"
                >
                  <Edit3 className="h-3 w-3" />
                </Button>
              </div>
            </div>
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
              onDelete={handleDeleteTask}
            />

            {/* 对话信息板块 */}
            <ConversationList
              conversations={conversations}
              conversationCount={task.conversation_count}
              taskId={task.id}
              projectId={task.project_id}
              onViewConversationGitDiff={handleViewConversationGitDiff}
              onViewConversationDetails={handleViewConversationDetails}
              onViewConversationLogs={handleViewConversationLogs}
              onRetryConversation={handleRetryConversation}
              onCancelConversation={handleCancelConversation}
              onDeleteConversation={handleDeleteConversation}
              toggleExpanded={toggleExpanded}
              isConversationExpanded={isConversationExpanded}
              shouldShowExpandButton={shouldShowExpandButton}
            />

            {/* 发送对话消息板块 - 仅在任务未完成且未取消时显示 */}
            {task.status !== "done" && task.status !== "cancelled" && (
              <NewMessageForm
                task={task}
                newMessage={newMessage}
                executionTime={executionTime}
                model={model}
                isPlanMode={isPlanMode}
                sending={sending}
                canSendMessage={canSend}
                isTaskCompleted={taskCompleted}
                _hasPendingOrRunningConversations={hasPendingConversations}
                onMessageChange={setNewMessage}
                onExecutionTimeChange={setExecutionTime}
                onModelChange={setModel}
                onPlanModeChange={setIsPlanMode}
                onSendMessage={handleSendMessage}
              />
            )}
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
        projectId={task?.project_id}
      />

      <ConversationDetailModal
        conversationId={selectedConversationId}
        projectId={task?.project_id}
        taskId={task?.id}
        isOpen={isConversationDetailModalOpen}
        onClose={handleCloseConversationDetails}
      />

      <ConversationLogModal
        conversationId={selectedLogConversationId}
        projectId={task?.project_id}
        taskId={task?.id}
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
            <AlertDialogAction 
              onClick={handleConfirmRetry} 
              disabled={retrying}
              className="bg-red-600 hover:bg-red-700 focus:ring-red-600 text-white"
            >
              {retrying ? t("common.processing") : t("common.confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={cancelConversationId !== null} onOpenChange={handleCancelCancel}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("taskConversations.execution.cancel_confirm_title")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("taskConversations.execution.cancel_confirm")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={cancelling}>
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction 
              onClick={handleConfirmCancel} 
              disabled={cancelling}
              className="bg-red-600 hover:bg-red-700 focus:ring-red-600 text-white"
            >
              {cancelling ? t("common.processing") : t("common.confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={deleteConversationId !== null} onOpenChange={handleCancelDeleteConversation}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("taskConversations.delete.confirmTitle")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("taskConversations.delete.confirmDescription")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deletingConversation}>
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction 
              onClick={handleConfirmDeleteConversation} 
              disabled={deletingConversation}
              className="bg-red-600 hover:bg-red-700 focus:ring-red-600 text-white"
            >
              {deletingConversation ? t("common.processing") : t("common.confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={isDeleteDialogOpen} onOpenChange={handleCancelDelete}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("tasks.delete.confirmTitle")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("tasks.delete.confirmDescription")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleting}>
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction 
              onClick={handleConfirmDelete} 
              disabled={deleting}
              className="bg-red-600 hover:bg-red-700 focus:ring-red-600 text-white"
            >
              {deleting ? t("common.processing") : t("common.confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <TaskTitleEditDialog
        open={isTitleEditDialogOpen}
        onOpenChange={handleCloseTitleEditDialog}
        task={task}
        onSuccess={handleTitleUpdateSuccess}
      />
    </>
  );
});

TaskDetailSheet.displayName = "TaskDetailSheet";