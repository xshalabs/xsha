import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ArrowLeft } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { TaskConversation } from "@/components/TaskConversation";
import { TaskExecutionLog } from "@/components/TaskExecutionLog";
import { TaskConversationResult } from "@/components/TaskConversationResult";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { Task } from "@/types/task";
import type {
  TaskConversation as TaskConversationInterface,
  ConversationFormData,
  ConversationStatus,
} from "@/types/task-conversation";

const TaskConversationPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, taskId } = useParams<{
    projectId: string;
    taskId: string;
  }>();

  const [task, setTask] = useState<Task | null>(null);
  const [conversations, setConversations] = useState<
    TaskConversationInterface[]
  >([]);
  const [selectedConversationId, setSelectedConversationId] = useState<
    number | null
  >(null);
  const [conversationsLoading, setConversationsLoading] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"log" | "result">("log");

  usePageTitle(
    task
      ? `${t("tasks.conversation")} - ${task.title}`
      : t("tasks.conversation")
  );

  useEffect(() => {
    const loadTask = async () => {
      if (!taskId) {
        logError(new Error("Task ID is required"), "Invalid task ID");
        navigate(`/projects/${projectId}/tasks`);
        return;
      }

      try {
        setLoading(true);
        setError(null);
        const response = await apiService.tasks.get(parseInt(taskId, 10));
        setTask(response.data);
      } catch (error) {
        logError(error as Error, "Failed to load task");
        console.error("Task loading error:", error);
        setError(
          error instanceof Error
            ? error.message
            : t("tasks.messages.loadFailed")
        );
        setTask(null);
      } finally {
        setLoading(false);
      }
    };

    loadTask();
  }, [taskId, projectId, navigate, t]);

  const loadConversations = async (taskId: number) => {
    try {
      setConversationsLoading(true);
      const response = await apiService.taskConversations.list({
        task_id: taskId,
        page: 1,
        page_size: 100,
      });

      setConversations(response.data.conversations);

      if (response.data.conversations.length > 0 && !selectedConversationId) {
        // 选择最近的对话（数组最后一个元素）
        const latestConversation = response.data.conversations[response.data.conversations.length - 1];
        setSelectedConversationId(latestConversation.id);
      }
    } catch (error) {
      logError(error as Error, "Failed to load conversations");
      console.error("Conversations loading error:", error);
    } finally {
      setConversationsLoading(false);
    }
  };

  useEffect(() => {
    if (task) {
      loadConversations(task.id);
    }
  }, [task]);

  const handleSendMessage = async (data: ConversationFormData) => {
    if (!task) return;

    try {
      await apiService.taskConversations.create({
        task_id: task.id,
        content: data.content,
      });

      await loadConversations(task.id);
    } catch (error) {
      logError(error as Error, "Failed to send message");
      throw error;
    }
  };

  const handleConversationRefresh = () => {
    if (task) {
      loadConversations(task.id);
    }
  };

  const handleDeleteConversation = async (conversationId: number) => {
    try {
      await apiService.taskConversations.delete(conversationId);

      if (selectedConversationId === conversationId) {
        const remainingConversations = conversations.filter(
          (c) => c.id !== conversationId
        );
        setSelectedConversationId(
          remainingConversations.length > 0
            ? remainingConversations[remainingConversations.length - 1].id
            : null
        );
      }

      if (task) {
        loadConversations(task.id);
      }
    } catch (error) {
      logError(error as Error, "Failed to delete conversation");
      throw error;
    }
  };

  const handleConversationStatusChange = (
    conversationId: number,
    newStatus: ConversationStatus
  ) => {
    setConversations((prev) =>
      prev.map((conv) =>
        conv.id === conversationId ? { ...conv, status: newStatus } : conv
      )
    );
  };

  const selectedConversation = conversations.find(
    (c) => c.id === selectedConversationId
  );

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="max-w-7xl mx-auto">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
              <p className="text-muted-foreground">{t("common.loading")}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error && !loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="max-w-7xl mx-auto">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <p className="text-red-600 mb-4">{error}</p>
              <Button
                variant="outline"
                onClick={() => navigate(`/projects/${projectId}/tasks`)}
                className="mr-2"
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                {t("common.back")}
              </Button>
              <Button
                variant="default"
                onClick={() => window.location.reload()}
              >
                {t("common.retry")}
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!task && !loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="max-w-7xl mx-auto">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <p className="text-muted-foreground">
                {t("tasks.messages.loadFailed")}
              </p>
              <Button
                variant="outline"
                onClick={() => navigate(`/projects/${projectId}/tasks`)}
                className="mt-4"
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                {t("common.back")}
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!task) {
    return null;
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              {task.project?.name} - {t("tasks.conversation")}
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">{task.title}</p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="default"
              onClick={() => navigate(`/projects/${projectId}/tasks`)}
              className="mb-4"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              {t("common.back")}
            </Button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 h-[calc(100vh-200px)]">
          <div className="flex flex-col">
            <TaskConversation
              taskTitle={task.title}
              conversations={conversations}
              selectedConversationId={selectedConversationId}
              loading={conversationsLoading}
              onSendMessage={handleSendMessage}
              onRefresh={handleConversationRefresh}
              onDeleteConversation={handleDeleteConversation}
              onSelectConversation={setSelectedConversationId}
              onConversationStatusChange={handleConversationStatusChange}
            />
          </div>

          <div className="flex flex-col">
            {selectedConversation ? (
              <Tabs
                value={activeTab}
                onValueChange={(value) =>
                  setActiveTab(value as "log" | "result")
                }
                className="h-full flex flex-col"
              >
                <TabsList className="grid w-full grid-cols-2">
                  <TabsTrigger value="log">
                    {t("taskConversation.execution_log")}
                  </TabsTrigger>
                  <TabsTrigger value="result">
                    {t("taskConversation.execution_result")}
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="log" className="flex-1 mt-2">
                  <div className="h-full">
                    <TaskExecutionLog
                      conversationId={selectedConversation.id}
                      conversationStatus={selectedConversation.status}
                      conversation={selectedConversation}
                      onStatusChange={(newStatus) =>
                        handleConversationStatusChange(
                          selectedConversation.id,
                          newStatus
                        )
                      }
                    />
                  </div>
                </TabsContent>

                <TabsContent value="result" className="flex-1 mt-2">
                  <div className="h-full overflow-auto">
                    <TaskConversationResult
                      conversationId={selectedConversation.id}
                      showHeader={false}
                    />
                  </div>
                </TabsContent>
              </Tabs>
            ) : (
              <div className="flex items-center justify-center h-full bg-muted rounded-lg border-2 border-dashed border-border">
                <div className="text-center text-muted-foreground">
                  <p className="text-lg font-medium mb-2">
                    {t("taskConversation.noSelection.title")}
                  </p>
                  <p className="text-sm">
                    {t("taskConversation.noSelection.description")}
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default TaskConversationPage;
