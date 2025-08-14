import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  ArrowLeft,
  Plus,
  Settings,
  Save,
} from "lucide-react";
import { Button } from "@/components/ui/button";
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
  KanbanBoardExtraMargin,
} from "@/components/kanban";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { usePageTitle } from "@/hooks/usePageTitle";
import { logError } from "@/lib/errors";
import type { Task, TaskFormData } from "@/types/task";
import { TaskFormCreateNew } from "@/components/TaskFormCreateNew";
import { useKanbanData } from "@/hooks/useKanbanData";
import { KanbanColumn } from "@/components/kanban/KanbanColumn";
import { TaskDetailSheet } from "@/components/kanban/TaskDetailSheet";
import { apiService } from "@/lib/api/index";

import type { TaskStatus } from "@/types/task";

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

export default function ProjectKanbanPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId } = useParams<{ projectId: string }>();

  // Use custom hooks for data management
  const {
    project,
    projects,
    tasks,
    loading,
    handleDropOverColumn,
    refreshKanbanData,
  } = useKanbanData(projectId);

  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [isCreateTaskSheetOpen, setIsCreateTaskSheetOpen] = useState(false);
  const [isCreatingTask, setIsCreatingTask] = useState(false);

  usePageTitle(
    project ? `${project.name} - ${t("common.kanban")}` : t("common.kanban")
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

      // Refresh kanban data using the hook function
      await refreshKanbanData();

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
                onDropOverColumn={(dataTransferData: string) =>
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
