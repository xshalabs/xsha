import { useState, useEffect } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Plus, Save, ArrowLeft } from "lucide-react";
import { useIsMobile } from "@/hooks/use-mobile";
import { Button } from "@/components/ui/button";
import { Logo } from "@/components/Logo";
import { ProjectSwitcher } from "@/components/ProjectSwitcher";
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
import { Skeleton } from "@/components/ui/skeleton";
import { usePageTitle } from "@/hooks/usePageTitle";
import { logError } from "@/lib/errors";
import type { Task, TaskFormData } from "@/types/task";
import { TaskFormCreateNew } from "@/components/TaskFormCreateNew";
import { useKanbanData } from "@/hooks/useKanbanData";
import { KanbanColumn } from "@/components/kanban/KanbanColumn";
import { TaskDetailSheet } from "@/components/kanban/TaskDetailSheet";
import { apiService } from "@/lib/api/index";
import {
  AppHeader,
  AppHeaderContent,
  AppHeaderActions,
} from "@/components/nav";


import type { TaskStatus } from "@/types/task";
import type { DevEnvironment } from "@/types/dev-environment";

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
  const [searchParams, setSearchParams] = useSearchParams();
  const isMobile = useIsMobile();

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

  // Handle taskId from URL parameters
  useEffect(() => {
    const taskIdParam = searchParams.get('taskId');
    if (taskIdParam && !loading && tasks) {
      const taskId = parseInt(taskIdParam, 10);
      if (!isNaN(taskId)) {
        // Find the task in all columns
        const allTasks = Object.values(tasks).flat();
        const targetTask = allTasks.find(task => task.id === taskId);
        
        if (targetTask) {
          setSelectedTask(targetTask);
          setIsSheetOpen(true);
          // Remove taskId from URL to clean up
          const newSearchParams = new URLSearchParams(searchParams);
          newSearchParams.delete('taskId');
          setSearchParams(newSearchParams, { replace: true });
        }
      }
    }
  }, [loading, tasks, searchParams, setSearchParams]);



  const handleAddTask = () => {
    setIsCreateTaskSheetOpen(true);
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

  const handleTaskDeleted = () => {
    // Refresh kanban data after task deletion
    refreshKanbanData();
  };

  const handleCreateTask = async (taskData: TaskFormData, selectedEnvironment?: DevEnvironment) => {
    if (!projectId) return;

    try {
      setIsCreatingTask(true);

      // Prepare env_params based on selected environment and model
      let envParams = "{}";
      if (taskData.dev_environment_id && taskData.model && taskData.model !== "default" && selectedEnvironment) {
        if (selectedEnvironment.type === "claude-code") {
          envParams = JSON.stringify({ model: taskData.model });
        }
      }

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
        env_params: envParams,
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



  if (loading) {
    return (
      <div className="min-h-screen bg-background">
        <AppHeader>
          <AppHeaderContent>
            <div className={`flex items-center ${isMobile ? "gap-2" : "gap-4"}`}>
              {/* Logo */}
              <Logo className={`${isMobile ? "h-6 w-auto" : "h-8 w-auto"} flex-shrink-0`} />
              {/* Project Switcher Skeleton */}
              <Skeleton className={`h-10 ${isMobile ? "w-44" : "w-64"}`} />
            </div>
          </AppHeaderContent>
          <AppHeaderActions>
            <div className={`flex items-center ${isMobile ? "gap-1" : "gap-2"}`}>
              <Skeleton className={`h-8 ${isMobile ? "w-8" : "w-24"}`} />
              <Skeleton className={`h-8 ${isMobile ? "w-16" : "w-24"}`} />
            </div>
          </AppHeaderActions>
        </AppHeader>
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
      <AppHeader>
        <AppHeaderContent>
          <div className={`flex items-center ${isMobile ? "gap-2" : "gap-4"}`}>
            {/* Logo */}
            <Logo className={`${isMobile ? "h-6 w-auto" : "h-8 w-auto"} flex-shrink-0`} />
            
            {/* Project Switcher */}
            {project && (
              <ProjectSwitcher
                projects={projects}
                currentProject={project}
                onProjectChange={handleProjectChange}
              />
            )}
          </div>
        </AppHeaderContent>
        <AppHeaderActions>
          <div className={`flex items-center ${isMobile ? "gap-1" : "gap-2"}`}>
            <Button 
              variant="ghost" 
              size={isMobile ? "sm" : "sm"}
              onClick={() => navigate("/projects")}
              className={`flex items-center ${isMobile ? "gap-1 px-2" : "gap-2"}`}
            >
              <ArrowLeft className="h-4 w-4" />
              {!isMobile && t("navigation.projects")}
            </Button>
            <Button 
              onClick={handleAddTask} 
              size="sm"
              className={isMobile ? "px-2" : ""}
            >
              <Plus className="h-4 w-4 mr-2" />
{isMobile ? t("common.create") : t("tasks.addTask")}
            </Button>
          </div>
        </AppHeaderActions>
      </AppHeader>

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
          onTaskDeleted={handleTaskDeleted}
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
