import { useState, useCallback, useEffect } from "react";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { Task } from "@/types/task";
import type {
  TaskConversation as TaskConversationInterface,
} from "@/types/task-conversation";

export function useTaskConversations(task: Task | null) {
  const [conversations, setConversations] = useState<TaskConversationInterface[]>([]);
  const [conversationsLoading, setConversationsLoading] = useState(false);
  const [newMessage, setNewMessage] = useState("");
  const [executionTime, setExecutionTime] = useState<Date | undefined>(undefined);
  const [model, setModel] = useState("default");
  const [isPlanMode, setIsPlanMode] = useState(false);
  const [sending, setSending] = useState(false);
  const [expandedConversations, setExpandedConversations] = useState<Set<number>>(new Set());

  const loadConversations = useCallback(async () => {
    if (!task) return;

    setConversationsLoading(true);
    try {
      const response = await apiService.taskConversations.list({
        task_id: task.id,
      });
      setConversations(response.data.conversations);
    } catch (error) {
      console.error("Failed to load conversations:", error);
      logError(error as Error, "Failed to load conversations");
    } finally {
      setConversationsLoading(false);
    }
  }, [task]);

  const handleSendMessage = async (attachmentIds?: number[]) => {
    if (!task || !newMessage.trim() || !canSendMessage()) return;

    setSending(true);
    try {
      // Prepare env_params based on task environment
      let envParams = "{}";
      if (task.dev_environment?.type === "claude-code") {
        const params: { model?: string; is_plan_mode?: boolean } = {};
        
        if (model && model !== "default") {
          params.model = model;
        }
        
        if (isPlanMode === true) {
          params.is_plan_mode = isPlanMode;
        }
        
        if (Object.keys(params).length > 0) {
          envParams = JSON.stringify(params);
        }
      }

      await apiService.taskConversations.create({
        task_id: task.id,
        content: newMessage.trim(),
        execution_time: executionTime?.toISOString(),
        env_params: envParams,
        attachment_ids: attachmentIds,
      });

      // Clear form and refresh conversations list
      setNewMessage("");
      setExecutionTime(undefined);
      setModel("default");
      await loadConversations();
    } catch (error) {
      console.error("Failed to send message:", error);
      throw error;
    } finally {
      setSending(false);
    }
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

  const hasPendingOrRunningConversations = () => {
    return conversations.some(
      (conv) => conv.status === "pending" || conv.status === "running"
    );
  };

  const isTaskCompleted = () => {
    return task?.status === "done" || task?.status === "cancelled";
  };

  const canSendMessage = () => {
    return !hasPendingOrRunningConversations() && !isTaskCompleted();
  };

  const shouldShowExpandButton = (content: string) => {
    return content.split("\n").length > 3 || content.length > 150;
  };

  // Reset state when task changes
  useEffect(() => {
    if (task) {
      setConversations([]);
      setNewMessage("");
      setExecutionTime(undefined);
      setModel("default");
      setIsPlanMode(false);
      setSending(false);
      setExpandedConversations(new Set());
      loadConversations();
    }
  }, [task?.id, loadConversations]);

  return {
    conversations,
    conversationsLoading,
    newMessage,
    setNewMessage,
    executionTime,
    setExecutionTime,
    model,
    setModel,
    isPlanMode,
    setIsPlanMode,
    sending,
    expandedConversations,
    loadConversations,
    handleSendMessage,
    toggleExpanded,
    isConversationExpanded,
    hasPendingOrRunningConversations,
    isTaskCompleted,
    canSendMessage,
    shouldShowExpandButton,
  };
}
