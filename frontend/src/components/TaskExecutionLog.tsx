import { useState, useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useTranslation } from "react-i18next";
import {
  Square,
  RotateCcw,
  CheckCircle,
  XCircle,
  AlertCircle,
  Terminal,
  RefreshCw,
} from "lucide-react";
import { taskExecutionLogsApi } from "@/lib/api/task-execution-logs";
import type { TaskExecutionLog } from "@/types/task-execution-log";
import type {
  ConversationStatus,
  TaskConversation,
} from "@/types/task-conversation";
import { formatToLocal } from "@/lib/timezone";

interface TaskExecutionLogProps {
  conversationId: number;
  conversationStatus: ConversationStatus;
  conversation?: TaskConversation;
  onStatusChange?: (newStatus: ConversationStatus) => void;
}

export function TaskExecutionLog({
  conversationId,
  conversationStatus,
  conversation,
  onStatusChange,
}: TaskExecutionLogProps) {
  const { t } = useTranslation();
  const [executionLog, setExecutionLog] = useState<TaskExecutionLog | null>(
    null
  );
  const [loading, setLoading] = useState(false);
  const [refreshLoading, setRefreshLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState<"cancel" | "retry" | null>(
    null
  );
  const [error, setError] = useState<string | null>(null);
  const [cancelDialogOpen, setCancelDialogOpen] = useState(false);
  const [retryDialogOpen, setRetryDialogOpen] = useState(false);
  const logContainerRef = useRef<HTMLDivElement>(null);

  const canCancel = (conversationStatus: ConversationStatus) => {
    return conversationStatus === "pending" || conversationStatus === "running";
  };

  const canRetry = (conversationStatus: ConversationStatus) => {
    return (
      conversationStatus === "failed" || conversationStatus === "cancelled"
    );
  };

  const scrollToBottom = () => {
    if (logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  };

  const loadExecutionLog = async () => {
    if (!conversationId) return;

    setLoading(true);
    setError(null);

    try {
      const log = await taskExecutionLogsApi.getExecutionLog(conversationId);
      setExecutionLog(log);
    } catch (error) {
      console.error("Failed to load execution log:", error);
      setError(t("errors.execution_log_load_failed"));
    } finally {
      setLoading(false);
    }
  };

  const refreshExecutionLog = async () => {
    if (!conversationId) return;

    setRefreshLoading(true);
    setError(null);

    try {
      const log = await taskExecutionLogsApi.getExecutionLog(conversationId);
      setExecutionLog(log);
      // 刷新后自动滚动到底部
      setTimeout(() => {
        if (log?.execution_logs) {
          scrollToBottom();
        }
      }, 100);
    } catch (error) {
      console.error("Failed to refresh execution log:", error);
      setError(t("errors.execution_log_load_failed"));
    } finally {
      setRefreshLoading(false);
    }
  };

  const handleCancelClick = () => {
    setCancelDialogOpen(true);
  };

  const handleConfirmCancel = async () => {
    if (!conversationId) return;

    setActionLoading("cancel");
    try {
      await taskExecutionLogsApi.cancelExecution(conversationId);
      await refreshExecutionLog();
      onStatusChange?.("cancelled");
      setCancelDialogOpen(false);
    } catch (error) {
      console.error("Failed to cancel execution:", error);
      setError(t("errors.execution_cancel_failed"));
      setCancelDialogOpen(false);
    } finally {
      setActionLoading(null);
    }
  };

  const handleCancelCancel = () => {
    setCancelDialogOpen(false);
  };

  const handleRetryClick = () => {
    setRetryDialogOpen(true);
  };

  const handleConfirmRetry = async () => {
    if (!conversationId) return;

    setActionLoading("retry");
    try {
      await taskExecutionLogsApi.retryExecution(conversationId);
      await refreshExecutionLog();
      onStatusChange?.("running");
      setRetryDialogOpen(false);
    } catch (error) {
      console.error("Failed to retry execution:", error);
      setError(t("errors.execution_retry_failed"));
      setRetryDialogOpen(false);
    } finally {
      setActionLoading(null);
    }
  };

  const handleCancelRetry = () => {
    setRetryDialogOpen(false);
  };

  const formatTime = (dateString: string | null) => {
    if (!dateString) return "-";
    return formatToLocal(dateString);
  };

  const formatDuration = (startTime: string | null, endTime: string | null) => {
    if (!startTime) return "-";
    const start = new Date(startTime);
    const end = endTime ? new Date(endTime) : new Date();
    const duration = Math.abs(end.getTime() - start.getTime());
    const seconds = Math.floor(duration / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
      return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds % 60}s`;
    } else {
      return `${seconds}s`;
    }
  };

  useEffect(() => {
    if (conversationStatus !== "pending") {
      loadExecutionLog();
    }
  }, [conversationId, conversationStatus]);

  // 当日志存在时自动滚动到底部
  useEffect(() => {
    if (executionLog?.execution_logs) {
      setTimeout(() => {
        scrollToBottom();
      }, 100);
    }
  }, [executionLog?.execution_logs]);

  if (conversationStatus === "pending") {
    return null;
  }

  if (loading) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-center py-4">
            <Terminal className="w-5 h-5 mr-2 animate-spin" />
            <span>{t("taskConversations.execution.messages.loading")}</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center text-red-600">
              <AlertCircle className="w-5 h-5 mr-2" />
              <span>{error}</span>
            </div>
            <Button variant="outline" size="sm" onClick={refreshExecutionLog}>
              {t("taskConversations.execution.actions.retry")}
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!executionLog) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="text-center py-4 text-gray-500">
            <Terminal className="w-8 h-8 mx-auto mb-2 opacity-50" />
            <p>{t("taskConversations.execution.messages.notFound")}</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <CardTitle className="text-lg flex items-center">
              <Terminal className="w-5 h-5 mr-2" />
              {t("taskConversations.execution.title")}
            </CardTitle>
            <CardDescription>
              {t("taskConversations.execution.subtitle")}
            </CardDescription>
          </div>

          <div className="flex items-center space-x-2">
            {canCancel(conversationStatus) && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleCancelClick}
                className="text-foreground hover:text-foreground"
                disabled={actionLoading === "cancel"}
              >
                <Square className="w-4 h-4 mr-1" />
                {t("taskConversations.execution.actions.cancel")}
              </Button>
            )}

            {canRetry(conversationStatus) && (
              <Button
                variant="outline"
                size="sm"
                className="text-foreground hover:text-foreground"
                onClick={handleRetryClick}
                disabled={actionLoading === "retry"}
              >
                <RotateCcw className="w-4 h-4 mr-1" />
                {t("taskConversations.execution.actions.retry")}
              </Button>
            )}
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-4 flex-1 overflow-y-auto">
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-gray-500">
              {t("taskConversations.execution.info.started")}:
            </span>
            <span className="ml-2">{formatTime(executionLog.started_at)}</span>
          </div>
          <div>
            <span className="text-gray-500">
              {t("taskConversations.execution.info.completed")}:
            </span>
            <span className="ml-2">
              {formatTime(executionLog.completed_at)}
            </span>
          </div>
          <div>
            <span className="text-gray-500">
              {t("taskConversations.execution.info.duration")}:
            </span>
            <span className="ml-2">
              {formatDuration(
                executionLog.started_at,
                executionLog.completed_at
              )}
            </span>
          </div>
        </div>

        {executionLog.error_message && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
            <div className="flex items-center mb-2">
              <XCircle className="w-4 h-4 text-red-500 mr-2" />
              <span className="font-medium text-red-700">
                {t("taskConversations.execution.info.errorMessage")}
              </span>
            </div>
            <pre className="text-sm text-red-600 whitespace-pre-wrap">
              {executionLog.error_message}
            </pre>
          </div>
        )}

        {executionLog.docker_command && (
          <div className="p-3 bg-gray-50 border border-gray-200 rounded-lg">
            <div className="flex items-center mb-2">
              <Terminal className="w-4 h-4 text-gray-500 mr-2" />
              <span className="font-medium text-gray-700">
                {t("taskConversations.execution.info.dockerCommand")}
              </span>
            </div>
            <pre className="text-sm text-gray-600 font-mono whitespace-pre-wrap">
              {executionLog.docker_command}
            </pre>
          </div>
        )}

        {conversation?.commit_hash && (
          <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
            <div className="flex items-center mb-2">
              <CheckCircle className="w-4 h-4 text-green-500 mr-2" />
              <span className="font-medium text-green-700">
                {t("taskConversations.execution.info.commitHash")}
              </span>
            </div>
            <span className="text-sm text-green-600 font-mono">
              {conversation.commit_hash}
            </span>
          </div>
        )}

        {executionLog.execution_logs && (
          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="font-medium text-foreground">
                {t("taskConversations.execution.info.executionLogs")}
              </span>
              {conversationStatus === "running" && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={refreshExecutionLog}
                  disabled={refreshLoading}
                >
                  <RefreshCw className={`w-4 h-4 mr-1 ${refreshLoading ? 'animate-spin' : ''}`} />
                  {t("common.refresh")}
                </Button>
              )}
            </div>

            <div 
              ref={logContainerRef}
              className="p-4 bg-black text-green-400 rounded-lg font-mono text-xs overflow-x-auto max-h-80 overflow-y-auto"
            >
              <pre className="whitespace-pre-wrap">
                {executionLog.execution_logs}
              </pre>
            </div>
          </div>
        )}
      </CardContent>

      <Dialog open={cancelDialogOpen} onOpenChange={setCancelDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-foreground">
              {t("taskConversations.execution.cancel_confirm_title")}
            </DialogTitle>
            <DialogDescription className="text-muted-foreground">
              {t("taskConversations.execution.cancel_confirm")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              className="text-foreground hover:text-foreground"
              onClick={handleCancelCancel}
            >
              {t("common.cancel")}
            </Button>
            <Button
              variant="destructive"
              className="text-foreground"
              onClick={handleConfirmCancel}
              disabled={actionLoading === "cancel"}
            >
              {t("common.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={retryDialogOpen} onOpenChange={setRetryDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-foreground">
              {t("taskConversations.execution.retry_confirm_title")}
            </DialogTitle>
            <DialogDescription className="text-muted-foreground">
              {t("taskConversations.execution.retry_confirm")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              className="text-foreground hover:text-foreground"
              onClick={handleCancelRetry}
            >
              {t("common.cancel")}
            </Button>
            <Button
              onClick={handleConfirmRetry}
              disabled={actionLoading === "retry"}
            >
              {t("common.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
