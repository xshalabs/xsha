import { useState, useEffect, useCallback } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  ArrowLeft,
  Plus,
  Settings,
  Calendar,
  GitBranch,
  User,
  Save,
  MessageSquare,
  FileText,
  Eye,
  MoreHorizontal,
  Clock,
  Monitor,
  Send,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  FormSheet,
  FormSheetContent,
  FormSheetHeader,
  FormSheetTitle,
  FormSheetDescription,
  FormSheetFooter,
  FormCardGroup,
} from "@/components/forms/form-sheet";
import { FormCard, FormCardContent } from "@/components/forms/form-card";
import {
  KanbanBoardProvider,
  KanbanBoard,
  KanbanBoardColumn,
  KanbanBoardColumnHeader,
  KanbanBoardColumnTitle,
  KanbanColorCircle,
  KanbanBoardColumnList,
  KanbanBoardColumnListItem,
  KanbanBoardCard,
  KanbanBoardCardTitle,
  KanbanBoardExtraMargin,
} from "@/components/kanban";

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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { usePageTitle } from "@/hooks/usePageTitle";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { Task, TaskStatus, TaskFormData } from "@/types/task";
import type { Project } from "@/types/project";
import type {
  TaskConversation as TaskConversationInterface,
} from "@/types/task-conversation";
import { TaskFormCreateNew } from "@/components/TaskFormCreateNew";

import { PushBranchDialog } from "@/components/PushBranchDialog";

const KANBAN_COLUMNS = [
  { id: "todo", title: "Todo", status: "todo" as TaskStatus },
  {
    id: "in_progress",
    title: "In Progress",
    status: "in_progress" as TaskStatus,
  },
  { id: "done", title: "Done", status: "done" as TaskStatus },
  { id: "cancelled", title: "Cancelled", status: "cancelled" as TaskStatus },
];

// Task Card Component
function TaskCard({ task, onClick }: { task: Task; onClick?: () => void }) {
  const handleClick = () => {
    onClick?.();
  };

  return (
    <KanbanBoardCard data={{ id: task.id.toString() }} onClick={handleClick}>
      <KanbanBoardCardTitle>{task.title}</KanbanBoardCardTitle>
      {task.conversation_count > 0 && (
        <div className="flex items-center justify-between mt-2">
          <Badge variant="outline" className="text-xs">
            {task.conversation_count} conversations
          </Badge>
        </div>
      )}
    </KanbanBoardCard>
  );
}

// Task Detail Sheet Component
function TaskDetailSheet({
  task,
  isOpen,
  onClose,
}: {
  task: Task | null;
  isOpen: boolean;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const [conversations, setConversations] = useState<
    TaskConversationInterface[]
  >([]);
  const [conversationsLoading, setConversationsLoading] = useState(false);
  const [isPushDialogOpen, setIsPushDialogOpen] = useState(false);
  const [newMessage, setNewMessage] = useState("");
  const [executionTime, setExecutionTime] = useState<Date | undefined>(undefined);
  const [sending, setSending] = useState(false);
  const [expandedConversations, setExpandedConversations] = useState<Set<number>>(new Set());

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

  const shouldShowExpandButton = (content: string) => {
    return content.split("\n").length > 3 || content.length > 150;
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

  const handleSendMessage = async () => {
    if (!task || !newMessage.trim() || !canSendMessage()) return;

    setSending(true);
    try {
      await apiService.taskConversations.create({
        task_id: task.id,
        content: newMessage.trim(),
        execution_time: executionTime?.toISOString(),
      });

      // Clear form and refresh conversations list
      setNewMessage("");
      setExecutionTime(undefined);
      await loadConversations();
    } catch (error) {
      console.error("Failed to send message:", error);
      throw error;
    } finally {
      setSending(false);
    }
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

  // Reset state and load conversations when task changes
  useEffect(() => {
    if (task) {
      // Reset all conversation-related state when task changes
      setConversations([]);
      setNewMessage("");
      setExecutionTime(undefined);
      setSending(false);
      setExpandedConversations(new Set());
      
      // Load conversations for the new task
      loadConversations();
    }
  }, [task?.id, loadConversations]); // Use task.id to ensure it triggers when task changes

  // 在所有hooks调用后进行条件性返回
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

// Kanban Column Component
function KanbanColumn({
  title,
  status,
  tasks,
  onTaskClick,
  onDropOverColumn,
}: {
  title: string;
  status: TaskStatus;
  tasks: Task[];
  onTaskClick: (task: Task) => void;
  onDropOverColumn: (dataTransferData: string) => void;
}) {
  const { t } = useTranslation();

  const getColumnColor = (status: TaskStatus) => {
    switch (status) {
      case "todo":
        return "gray";
      case "in_progress":
        return "blue";
      case "done":
        return "green";
      case "cancelled":
        return "red";
      default:
        return "gray";
    }
  };

  return (
    <KanbanBoardColumn columnId={status} onDropOverColumn={onDropOverColumn}>
      <KanbanBoardColumnHeader>
        <KanbanBoardColumnTitle columnId={status}>
          <KanbanColorCircle color={getColumnColor(status)} />
          <span>{title}</span>
          <Badge variant="secondary" className="text-xs ml-2">
            {tasks.length}
          </Badge>
        </KanbanBoardColumnTitle>
      </KanbanBoardColumnHeader>

      <KanbanBoardColumnList>
        {tasks.map((task) => (
          <KanbanBoardColumnListItem
            key={task.id}
            cardId={task.id.toString()}
            onDropOverListItem={(dataTransferData, dropDirection) => {
              // Handle task reordering within the same column
              console.log("Drop over task:", dataTransferData, dropDirection);
            }}
          >
            <TaskCard task={task} onClick={() => onTaskClick(task)} />
          </KanbanBoardColumnListItem>
        ))}
        {tasks.length === 0 && (
          <li className="text-center text-muted-foreground text-sm py-8">
            {t("tasks.no_tasks")}
          </li>
        )}
      </KanbanBoardColumnList>
    </KanbanBoardColumn>
  );
}

export default function ProjectKanbanPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId } = useParams<{ projectId: string }>();

  const [project, setProject] = useState<Project | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [tasks, setTasks] = useState<{
    todo: Task[];
    in_progress: Task[];
    done: Task[];
    cancelled: Task[];
  }>({
    todo: [],
    in_progress: [],
    done: [],
    cancelled: [],
  });
  const [loading, setLoading] = useState(true);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [isCreateTaskSheetOpen, setIsCreateTaskSheetOpen] = useState(false);
  const [isCreatingTask, setIsCreatingTask] = useState(false);

  usePageTitle(
    project ? `${project.name} - ${t("common.kanban")}` : t("common.kanban")
  );

  // Load project and projects list
  useEffect(() => {
    const loadData = async () => {
      if (!projectId) return;

      try {
        setLoading(true);

        console.log("Loading kanban data for project:", projectId);

        // Load current project first
        const projectResponse = await apiService.projects.get(
          parseInt(projectId)
        );
        setProject(projectResponse.project);
        console.log("Project loaded:", projectResponse.project);

        // Load projects list and kanban data in parallel
        const [projectsResponse, kanbanResponse] = await Promise.all([
          apiService.projects.list(),
          apiService.tasks.getKanbanTasks(parseInt(projectId)),
        ]);

        setProjects(projectsResponse.projects);
        setTasks(kanbanResponse.data);
        console.log("Kanban data loaded:", kanbanResponse.data);
      } catch (error) {
        console.error("Failed to load project data:", error);
        logError(error as Error, "Failed to load project data");

        // If project loading failed, try to load just the projects list for the dropdown
        try {
          const projectsResponse = await apiService.projects.list();
          setProjects(projectsResponse.projects);
        } catch (projectsError) {
          console.error("Failed to load projects list:", projectsError);
        }
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [projectId]);

  // Handle task drop on column
  const handleDropOverColumn = useCallback(
    async (dataTransferData: string, targetStatus: TaskStatus) => {
      try {
        const draggedTaskData = JSON.parse(dataTransferData);
        const taskId = parseInt(draggedTaskData.id);

        // Find the task in current state
        const sourceTask = Object.values(tasks)
          .flat()
          .find((task) => task.id === taskId);

        if (!sourceTask || sourceTask.status === targetStatus) {
          return;
        }

        // Update task status via API
        await apiService.tasks.batchUpdateStatus({
          task_ids: [taskId],
          status: targetStatus,
        });

        // Update local state
        setTasks((prev) => {
          const newTasks = { ...prev };

          // Remove task from old column
          newTasks[sourceTask.status] = newTasks[sourceTask.status].filter(
            (t) => t.id !== taskId
          );

          // Add task to new column with updated status
          const updatedTask = {
            ...sourceTask,
            status: targetStatus,
          };
          newTasks[targetStatus] = [...newTasks[targetStatus], updatedTask];

          return newTasks;
        });
      } catch (error) {
        console.error("Failed to update task status:", error);
        logError(error as Error, "Failed to update task status");
      }
    },
    [tasks]
  );

  const handleGoBack = () => {
    navigate(`/projects/${projectId}/tasks`);
  };

  const handleAddTask = () => {
    setIsCreateTaskSheetOpen(true);
  };

  const handleProjectSettings = () => {
    navigate(`/projects/${projectId}/edit`);
  };

  const handleProjectChange = (newProjectId: string) => {
    navigate(`/projects/${newProjectId}/kanban`);
  };

  const handleTaskClick = (task: Task) => {
    setSelectedTask(task);
    setIsSheetOpen(true);
  };

  const handleCloseSheet = () => {
    setIsSheetOpen(false);
    setSelectedTask(null);
  };

  const handleCreateTask = async (taskData: TaskFormData) => {
    if (!projectId) return;

    try {
      setIsCreatingTask(true);

      // Convert TaskFormData to CreateTaskRequest
      const createRequest = {
        title: taskData.title,
        start_branch: taskData.start_branch,
        project_id: taskData.project_id,
        dev_environment_id: taskData.dev_environment_id,
        requirement_desc: taskData.requirement_desc,
        include_branches: taskData.include_branches,
        execution_time: taskData.execution_time
          ? taskData.execution_time.toISOString()
          : undefined,
      };

      const response = await apiService.tasks.create(createRequest);

      // Refresh kanban data
      const kanbanResponse = await apiService.tasks.getKanbanTasks(
        parseInt(projectId)
      );
      setTasks(kanbanResponse.data);

      // Close the sheet
      setIsCreateTaskSheetOpen(false);

      // Show success message (you might want to add toast notification here)
      console.log("Task created successfully:", response.data);
    } catch (error) {
      console.error("Failed to create task:", error);
      logError(error as Error, "Failed to create task");
      // Error will be handled by the TaskFormCreate component
      throw error;
    } finally {
      setIsCreatingTask(false);
    }
  };

  const handleCloseCreateTaskSheet = () => {
    setIsCreateTaskSheetOpen(false);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background">
        <div className="border-b">
          <div className="flex h-16 items-center px-6">
            <div className="flex items-center space-x-4">
              <Skeleton className="h-8 w-8" />
              <Skeleton className="h-6 w-48" />
            </div>
            <div className="ml-auto flex items-center space-x-4">
              <Skeleton className="h-8 w-24" />
              <Skeleton className="h-8 w-8" />
            </div>
          </div>
        </div>
        <div className="p-6">
          <div className="flex space-x-6 h-[calc(100vh-8rem)]">
            {[1, 2, 3, 4].map((i) => (
              <Skeleton key={i} className="min-w-80 h-full" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (!loading && !project) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center space-y-4">
          <h1 className="text-2xl font-bold mb-2">
            {t("errors.projectNotFound")}
          </h1>
          <p className="text-muted-foreground">
            {projectId
              ? `Project with ID ${projectId} was not found.`
              : "No project ID provided."}
          </p>
          <div className="space-x-2">
            <Button onClick={() => navigate("/projects")}>
              {t("common.backToProjects")}
            </Button>
            {projectId && (
              <Button
                variant="outline"
                onClick={() => window.location.reload()}
              >
                {t("common.retry")}
              </Button>
            )}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="flex sticky top-0 bg-background h-14 shrink-0 items-center gap-2 border-b px-2 z-10">
        <div className="flex flex-1 items-center gap-2 px-3">
          <Button
            variant="ghost"
            size="icon"
            onClick={handleGoBack}
            className="h-8 w-8 -ml-1"
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <Separator orientation="vertical" className="mr-2 h-4" />

          <Select value={projectId} onValueChange={handleProjectChange}>
            <SelectTrigger className="min-w-48 border-none shadow-none text-sm font-semibold bg-transparent">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {projects.map((proj) => (
                <SelectItem key={proj.id} value={proj.id.toString()}>
                  {proj.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="ml-auto px-3">
          <div className="flex items-center gap-2">
            <Button onClick={handleAddTask} size="sm">
              <Plus className="h-4 w-4 mr-2" />
              {t("tasks.addTask")}
            </Button>
            <Separator orientation="vertical" className="h-4" />
            <Button
              variant="outline"
              size="icon"
              onClick={handleProjectSettings}
              className="h-8 w-8"
            >
              <Settings className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      {/* Kanban Board */}
      <main className="p-6">
        <KanbanBoardProvider>
          <KanbanBoard className="h-[calc(100vh-8rem)]">
            {KANBAN_COLUMNS.map((column) => (
              <KanbanColumn
                key={column.id}
                title={t(`tasks.status.${column.status}`)}
                status={column.status}
                tasks={tasks[column.status] || []}
                onTaskClick={handleTaskClick}
                onDropOverColumn={(dataTransferData) =>
                  handleDropOverColumn(dataTransferData, column.status)
                }
              />
            ))}
            <KanbanBoardExtraMargin />
          </KanbanBoard>
        </KanbanBoardProvider>

        {/* Task Detail Sheet */}
        <TaskDetailSheet
          task={selectedTask}
          isOpen={isSheetOpen}
          onClose={handleCloseSheet}
        />

        {/* Create Task Sheet */}
        <FormSheet
          open={isCreateTaskSheetOpen}
          onOpenChange={setIsCreateTaskSheetOpen}
        >
          <FormSheetContent className="w-full sm:w-[800px] sm:max-w-[800px]">
            <FormSheetHeader>
              <FormSheetTitle>{t("tasks.actions.create")}</FormSheetTitle>
              <FormSheetDescription>
                {t("tasks.form.createDescription")}
              </FormSheetDescription>
            </FormSheetHeader>
            <FormCardGroup className="overflow-y-auto">
              <FormCard className="border-none overflow-auto">
                <FormCardContent>
                  {project && (
                    <TaskFormCreateNew
                      defaultProjectId={project.id}
                      currentProject={project}
                      onSubmit={handleCreateTask}
                      onCancel={handleCloseCreateTaskSheet}
                      formId="task-create-sheet-form"
                    />
                  )}
                </FormCardContent>
              </FormCard>
            </FormCardGroup>
            <FormSheetFooter>
              <Button
                type="submit"
                form="task-create-sheet-form"
                disabled={isCreatingTask}
              >
                <Save className="w-4 h-4 mr-2" />
                {isCreatingTask
                  ? t("common.saving")
                  : t("tasks.actions.create")}
              </Button>
            </FormSheetFooter>
          </FormSheetContent>
        </FormSheet>
      </main>
    </div>
  );
}
