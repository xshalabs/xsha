import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useTranslation } from "react-i18next";
import {
  Square,
  RotateCcw,
  CheckCircle,
  XCircle,
  AlertCircle,
  Terminal,
  Eye,
  EyeOff,
} from "lucide-react";
import { taskExecutionLogsApi } from "@/lib/api/task-execution-logs";
import type { TaskExecutionLog } from "@/types/task-execution-log";
import type { ConversationStatus } from "@/types/task-conversation";

interface TaskExecutionLogProps {
  conversationId: number;
  conversationStatus: ConversationStatus;
  onStatusChange?: (newStatus: ConversationStatus) => void;
}

export function TaskExecutionLog({
  conversationId,
  conversationStatus,
  onStatusChange,
}: TaskExecutionLogProps) {
  const { t } = useTranslation();
  const [executionLog, setExecutionLog] = useState<TaskExecutionLog | null>(
    null
  );
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState<"cancel" | "retry" | null>(
    null
  );
  const [showLogs, setShowLogs] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 检查是否可以取消 - 基于对话状态
  const canCancel = (conversationStatus: ConversationStatus) => {
    return conversationStatus === "pending" || conversationStatus === "running";
  };

  // 检查是否可以重试 - 基于对话状态
  const canRetry = (conversationStatus: ConversationStatus) => {
    return (
      conversationStatus === "failed" || conversationStatus === "cancelled"
    );
  };

  // 加载执行日志
  const loadExecutionLog = async () => {
    if (!conversationId) return;

    setLoading(true);
    setError(null);

    try {
      const log = await taskExecutionLogsApi.getExecutionLog(conversationId);
      setExecutionLog(log);
    } catch (error) {
      console.error("Failed to load execution log:", error);
      setError("Failed to load execution log");
    } finally {
      setLoading(false);
    }
  };

  // 取消执行
  const handleCancel = async () => {
    if (!conversationId) return;

    setActionLoading("cancel");
    try {
      await taskExecutionLogsApi.cancelExecution(conversationId);
      await loadExecutionLog();
      onStatusChange?.("cancelled");
    } catch (error) {
      console.error("Failed to cancel execution:", error);
      setError("Failed to cancel execution");
    } finally {
      setActionLoading(null);
    }
  };

  // 重试执行
  const handleRetry = async () => {
    if (!conversationId) return;

    setActionLoading("retry");
    try {
      await taskExecutionLogsApi.retryExecution(conversationId);
      await loadExecutionLog();
      onStatusChange?.("running");
    } catch (error) {
      console.error("Failed to retry execution:", error);
      setError("Failed to retry execution");
    } finally {
      setActionLoading(null);
    }
  };

  // 格式化时间
  const formatTime = (dateString: string | null) => {
    if (!dateString) return "-";
    return new Date(dateString).toLocaleString();
  };

  // 格式化持续时间
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

  // 初始加载
  useEffect(() => {
    if (conversationStatus !== "pending") {
      loadExecutionLog();
    }
  }, [conversationId, conversationStatus]);

  // 如果对话状态是 pending，不显示执行日志
  if (conversationStatus === "pending") {
    return null;
  }

  if (loading) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-center py-4">
            <Terminal className="w-5 h-5 mr-2 animate-spin" />
            <span>{t("taskConversation.execution.messages.loading")}</span>
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
            <Button variant="outline" size="sm" onClick={loadExecutionLog}>
              {t("taskConversation.execution.actions.retry")}
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
            <p>{t("taskConversation.execution.messages.notFound")}</p>
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
              {t("taskConversation.execution.title")}
            </CardTitle>
            <CardDescription>
              {t("taskConversation.execution.subtitle")}
            </CardDescription>
          </div>

          <div className="flex items-center space-x-2">
            {/* 操作按钮 */}
            {canCancel(conversationStatus) && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleCancel}
                disabled={actionLoading === "cancel"}
              >
                <Square className="w-4 h-4 mr-1" />
                {t("taskConversation.execution.actions.cancel")}
              </Button>
            )}

            {canRetry(conversationStatus) && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleRetry}
                disabled={actionLoading === "retry"}
              >
                <RotateCcw className="w-4 h-4 mr-1" />
                {t("taskConversation.execution.actions.retry")}
              </Button>
            )}
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-4 flex-1 overflow-y-auto">
        {/* 执行信息 */}
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-gray-500">
              {t("taskConversation.execution.info.started")}:
            </span>
            <span className="ml-2">{formatTime(executionLog.started_at)}</span>
          </div>
          <div>
            <span className="text-gray-500">
              {t("taskConversation.execution.info.completed")}:
            </span>
            <span className="ml-2">
              {formatTime(executionLog.completed_at)}
            </span>
          </div>
          <div>
            <span className="text-gray-500">
              {t("taskConversation.execution.info.duration")}:
            </span>
            <span className="ml-2">
              {formatDuration(
                executionLog.started_at,
                executionLog.completed_at
              )}
            </span>
          </div>

        </div>

        {/* 错误信息 */}
        {executionLog.error_message && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
            <div className="flex items-center mb-2">
              <XCircle className="w-4 h-4 text-red-500 mr-2" />
              <span className="font-medium text-red-700">
                {t("taskConversation.execution.info.errorMessage")}
              </span>
            </div>
            <pre className="text-sm text-red-600 whitespace-pre-wrap">
              {executionLog.error_message}
            </pre>
          </div>
        )}

        {/* Docker 命令 */}
        {executionLog.docker_command && (
          <div className="p-3 bg-gray-50 border border-gray-200 rounded-lg">
            <div className="flex items-center mb-2">
              <Terminal className="w-4 h-4 text-gray-500 mr-2" />
              <span className="font-medium text-gray-700">
                {t("taskConversation.execution.info.dockerCommand")}
              </span>
            </div>
            <pre className="text-sm text-gray-600 font-mono whitespace-pre-wrap">
              {executionLog.docker_command}
            </pre>
          </div>
        )}

        {/* 提交哈希 */}
        {executionLog.commit_hash && (
          <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
            <div className="flex items-center mb-2">
              <CheckCircle className="w-4 h-4 text-green-500 mr-2" />
              <span className="font-medium text-green-700">
                {t("taskConversation.execution.info.commitHash")}
              </span>
            </div>
            <span className="text-sm text-green-600 font-mono">
              {executionLog.commit_hash}
            </span>
          </div>
        )}

        {/* 执行日志 */}
        {executionLog.execution_logs && (
          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="font-medium text-gray-700">
                {t("taskConversation.execution.info.executionLogs")}
              </span>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowLogs(!showLogs)}
              >
                {showLogs ? (
                  <>
                    <EyeOff className="w-4 h-4 mr-1" />
                    {t("taskConversation.execution.actions.hideLogs")}
                  </>
                ) : (
                  <>
                    <Eye className="w-4 h-4 mr-1" />
                    {t("taskConversation.execution.actions.showLogs")}
                  </>
                )}
              </Button>
            </div>

            {showLogs && (
              <div className="p-4 bg-black text-green-400 rounded-lg font-mono text-xs overflow-x-auto max-h-80 overflow-y-auto">
                <pre className="whitespace-pre-wrap">
                  {executionLog.execution_logs}
                </pre>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
