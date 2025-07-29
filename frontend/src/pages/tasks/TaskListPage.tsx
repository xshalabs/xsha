import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useParams, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { usePageTitle } from "@/hooks/usePageTitle";
import { TaskList } from "@/components/TaskList";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { Task, TaskStatus } from "@/types/task";
import type { Project } from "@/types/project";
import { Plus } from "lucide-react";

const TaskListPage: React.FC = () => {
  const { t } = useTranslation();
  const { projectId } = useParams<{ projectId: string }>();
  const navigate = useNavigate();

  const [currentProject, setCurrentProject] = useState<Project | null>(null);

  const [tasks, setTasks] = useState<Task[]>([]);
  const [tasksLoading, setTasksLoading] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);
  const [statusFilter, setStatusFilter] = useState<TaskStatus | undefined>();

  const [projects, setProjects] = useState<Project[]>([]);

  usePageTitle(
    currentProject ? `${currentProject.name} - ${t("tasks.title")}` : t("tasks.title")
  );

  const pageSize = 20;

  const loadProjects = async () => {
    try {
      const response = await apiService.projects.list();
      setProjects(response.projects);
    } catch (error) {
      logError(error as Error, "Failed to load projects");
    }
  };

  const loadTasks = async (
    page = 1,
    status?: TaskStatus,
    projectId?: number
  ) => {
    try {
      setTasksLoading(true);
      const response = await apiService.tasks.list({
        page,
        page_size: pageSize,
        status,
        project_id: projectId,
      });

      setTasks(response.data.tasks);
      setTotalPages(Math.ceil(response.data.total / pageSize));
      setTotal(response.data.total);
      setCurrentPage(page);
    } catch (error) {
      logError(error as Error, "Failed to load tasks");
    } finally {
      setTasksLoading(false);
    }
  };

  useEffect(() => {
    loadProjects();
    loadTasks(1, statusFilter, projectId ? parseInt(projectId, 10) : undefined);
    if (projectId) {
      apiService.projects
        .get(parseInt(projectId, 10))
        .then((response) => {
          setCurrentProject(response.project);
        })
        .catch((error) => {
          logError(error as Error, "Failed to load project");
          navigate("/projects");
        });
    }
  }, [projectId]);

  const handleTaskCreate = () => {
    navigate(`/projects/${projectId}/tasks/create`);
  };

  const handleTaskEdit = (task: Task) => {
    navigate(`/projects/${projectId}/tasks/${task.id}/edit`);
  };

  const handleTaskDelete = async (id: number) => {
    try {
      await apiService.tasks.delete(id);
      loadTasks(
        currentPage,
        statusFilter,
        projectId ? parseInt(projectId, 10) : undefined
      );
    } catch (error) {
      logError(error as Error, "Failed to delete task");
      alert(
        error instanceof Error
          ? error.message
          : t("tasks.messages.deleteFailed")
      );
    }
  };

  const handleViewConversation = (task: Task) => {
    navigate(`/projects/${projectId}/tasks/${task.id}/conversation`);
  };

  const handlePageChange = (page: number) => {
    loadTasks(
      page,
      statusFilter,
      projectId ? parseInt(projectId, 10) : undefined
    );
  };

  const handleStatusFilterChange = (status: TaskStatus | undefined) => {
    setStatusFilter(status);
    loadTasks(1, status, projectId ? parseInt(projectId, 10) : undefined);
  };

  const handleProjectFilterChange = (_projectId: number | undefined) => {
    // This function is no longer needed as project filtering is handled by URL param
    // Keeping it for now, but it will not be called from the TaskList component
    // as the project filter is now a URL param.
  };

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              {t("tasks.title")}
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {t("tasks.page_description")}
            </p>
          </div>
          <div className="flex gap-2">
            <Button onClick={handleTaskCreate}>
              <Plus className="h-4 w-4 mr-2" />
              {t("tasks.create")}
            </Button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <TaskList
          tasks={tasks}
          projects={projects}
          loading={tasksLoading}
          currentPage={currentPage}
          totalPages={totalPages}
          total={total}
          statusFilter={statusFilter}
          projectFilter={projectId ? parseInt(projectId, 10) : undefined}
          hideProjectFilter={true}
          onPageChange={handlePageChange}
          onStatusFilterChange={handleStatusFilterChange}
          onProjectFilterChange={handleProjectFilterChange}
          onEdit={handleTaskEdit}
          onDelete={handleTaskDelete}
          onViewConversation={handleViewConversation}
          onCreateNew={handleTaskCreate}
        />
      </div>
    </div>
  );
};

export default TaskListPage;
