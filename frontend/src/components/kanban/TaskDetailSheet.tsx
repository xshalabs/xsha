import { useState, useCallback, memo, useMemo, useRef, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Edit3, Check, X } from "lucide-react";
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
import { Input } from "@/components/ui/input";
import { PushBranchDialog } from "@/components/PushBranchDialog";
import { TaskGitDiffModal } from "./TaskGitDiffModal";
import { ConversationGitDiffModal } from "./ConversationGitDiffModal";
import { ConversationDetailModal } from "@/components/ConversationDetailModal";
import { ConversationLogModal } from "./ConversationLogModal";
import { useTaskConversations } from "@/hooks/useTaskConversations";
import { taskExecutionLogsApi } from "@/lib/api/task-execution-logs";
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
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [editingTitle, setEditingTitle] = useState("");
  const [updatingTitle, setUpdatingTitle] = useState(false);
  const titleInputRef = useRef<HTMLInputElement>(null);

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

  const handleCancelConversation = useCallback((conversationId: number) => {
    setCancelConversationId(conversationId);
  }, []);

  const handleConfirmCancel = useCallback(async () => {
    if (!cancelConversationId) return;
    
    setCancelling(true);
    try {
      await taskExecutionLogsApi.cancelExecution(cancelConversationId);
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
  }, [cancelConversationId, t, loadConversations]);

  const handleCancelCancel = useCallback(() => {
    setCancelConversationId(null);
  }, []);

  const handleDeleteConversation = useCallback((conversationId: number) => {
    setDeleteConversationId(conversationId);
  }, []);

  const handleConfirmDeleteConversation = useCallback(async () => {
    if (!deleteConversationId) return;
    
    setDeletingConversation(true);
    try {
      await taskConversationsApi.delete(deleteConversationId);
      toast.success(t("taskConversations.delete.deleteSuccess"));
      // Refresh conversations to show updated list
      await loadConversations();
    } catch (error) {
      console.error("Failed to delete conversation:", error);
      toast.error(t("taskConversations.delete.deleteFailed"));
    } finally {
      setDeletingConversation(false);
      setDeleteConversationId(null);
    }
  }, [deleteConversationId, t, loadConversations]);

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
      await tasksApi.delete(task.id);
      toast.success(t("tasks.delete.deleteSuccess"));
      // Close the sheet first
      onClose();
      // Then notify parent about task deletion
      onTaskDeleted?.();
    } catch (error) {
      console.error("Failed to delete task:", error);
      toast.error(t("tasks.delete.deleteFailed"));
    } finally {
      setDeleting(false);
      setIsDeleteDialogOpen(false);
    }
  }, [task, t, onClose, onTaskDeleted]);

  const handleCancelDelete = useCallback(() => {
    setIsDeleteDialogOpen(false);
  }, []);

  const handleStartEditTitle = useCallback(() => {
    if (!task) return;
    setEditingTitle(task.title);
    setIsEditingTitle(true);
    // Focus the input after state update
    setTimeout(() => {
      titleInputRef.current?.focus();
      titleInputRef.current?.select();
    }, 0);
  }, [task]);

  const handleCancelEditTitle = useCallback(() => {
    setIsEditingTitle(false);
    setEditingTitle("");
  }, []);

  const handleSaveTitle = useCallback(async () => {
    if (!task) return;
    
    const trimmedTitle = editingTitle.trim();
    if (!trimmedTitle) {
      toast.error(t("tasks.validation.titleRequired"));
      return;
    }
    
    if (trimmedTitle === task.title) {
      handleCancelEditTitle();
      return;
    }
    
    setUpdatingTitle(true);
    try {
      await tasksApi.update(task.id, { title: trimmedTitle });
      toast.success(t("tasks.messages.updateSuccess"));
      
      // Update local task data
      task.title = trimmedTitle;
      
      setIsEditingTitle(false);
      setEditingTitle("");
    } catch (error) {
      console.error("Failed to update task title:", error);
      toast.error(t("tasks.messages.updateFailed"));
    } finally {
      setUpdatingTitle(false);
    }
  }, [task, editingTitle, t, handleCancelEditTitle]);

  const handleTitleKeyDown = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleSaveTitle();
    } else if (e.key === "Escape") {
      e.preventDefault();
      handleCancelEditTitle();
    }
  }, [handleSaveTitle, handleCancelEditTitle]);

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
            <div className="w-full">
              {isEditingTitle ? (
                <div className="flex items-center gap-2">
                  <Input
                    ref={titleInputRef}
                    value={editingTitle}
                    onChange={(e) => setEditingTitle(e.target.value)}
                    onKeyDown={handleTitleKeyDown}
                    disabled={updatingTitle}
                    className="flex-1 text-lg font-semibold"
                    placeholder={t("tasks.fields.title")}
                  />
                  <div className="flex items-center gap-1 shrink-0">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={handleSaveTitle}
                      disabled={updatingTitle}
                      className="h-8 w-8 p-0 text-green-600 hover:text-green-700 hover:bg-green-50 dark:hover:bg-green-950"
                    >
                      <Check className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={handleCancelEditTitle}
                      disabled={updatingTitle}
                      className="h-8 w-8 p-0 text-gray-500 hover:text-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <SheetTitle className="text-foreground font-semibold">
                    {task.title}
                  </SheetTitle>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleStartEditTitle}
                    className="h-6 w-6 p-0 text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:text-gray-300 dark:hover:bg-gray-800 shrink-0"
                  >
                    <Edit3 className="h-3 w-3" />
                  </Button>
                </div>
              )}
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
              conversationsLoading={conversationsLoading}
              conversationCount={task.conversation_count}
              task={task}
              taskId={task.id}
              onLoadConversations={loadConversations}
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

            {/* 发送对话消息板块 - 仅在任务未完成时显示 */}
            {task.status !== "done" && (
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
    </>
  );
});

TaskDetailSheet.displayName = "TaskDetailSheet";