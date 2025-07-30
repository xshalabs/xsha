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
import type { DevEnvironment } from "@/types/dev-environment";
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
  const [titleFilter, setTitleFilter] = useState<string | undefined>();
  const [branchFilter, setBranchFilter] = useState<string | undefined>();
  const [devEnvironmentFilter, setDevEnvironmentFilter] = useState<
    number | undefined
  >();

  const [projects, setProjects] = useState<Project[]>([]);
  const [devEnvironments, setDevEnvironments] = useState<DevEnvironment[]>([]);

  usePageTitle(
    currentProject
      ? `${currentProject.name} - ${t("tasks.title")}`
      : t("tasks.title")
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

  const loadDevEnvironments = async () => {
    try {
      const response = await apiService.devEnvironments.list();
      setDevEnvironments(response.environments);
    } catch (error) {
      logError(error as Error, "Failed to load dev environments");
    }
  };

  const loadTasks = async (
    page = 1,
    status?: TaskStatus,
    projectId?: number,
    title?: string,
    branch?: string,
    devEnvId?: number
  ) => {
    try {
      setTasksLoading(true);
      const response = await apiService.tasks.list({
        page,
        page_size: pageSize,
        status,
        project_id: projectId,
        title,
        branch,
        dev_environment_id: devEnvId,
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
    loadDevEnvironments();
    loadTasks(
      1,
      statusFilter,
      projectId ? parseInt(projectId, 10) : undefined,
      titleFilter,
      branchFilter,
      devEnvironmentFilter
    );
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
        projectId ? parseInt(projectId, 10) : undefined,
        titleFilter,
        branchFilter,
        devEnvironmentFilter
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

  const handleViewGitDiff = (task: Task) => {
    navigate(`/projects/${projectId}/tasks/${task.id}/git-diff`);
  };

  const handlePageChange = (page: number) => {
    loadTasks(
      page,
      statusFilter,
      projectId ? parseInt(projectId, 10) : undefined,
      titleFilter,
      branchFilter,
      devEnvironmentFilter
    );
  };

  const handleStatusFilterChange = (status: TaskStatus | undefined) => {
    setStatusFilter(status);
    loadTasks(
      1,
      status,
      projectId ? parseInt(projectId, 10) : undefined,
      titleFilter,
      branchFilter,
      devEnvironmentFilter
    );
  };

  const handleTitleFilterChange = (title: string | undefined) => {
    setTitleFilter(title);
    loadTasks(
      1,
      statusFilter,
      projectId ? parseInt(projectId, 10) : undefined,
      title,
      branchFilter,
      devEnvironmentFilter
    );
  };

  const handleBranchFilterChange = (branch: string | undefined) => {
    setBranchFilter(branch);
    loadTasks(
      1,
      statusFilter,
      projectId ? parseInt(projectId, 10) : undefined,
      titleFilter,
      branch,
      devEnvironmentFilter
    );
  };

  const handleDevEnvironmentFilterChange = (envId: number | undefined) => {
    setDevEnvironmentFilter(envId);
    loadTasks(
      1,
      statusFilter,
      projectId ? parseInt(projectId, 10) : undefined,
      titleFilter,
      branchFilter,
      envId
    );
  };

  const handleProjectFilterChange = (_projectId: number | undefined) => {
    // This function is no longer needed as project filtering is handled by URL param.
    // Keeping it for now, but it will not be called from the TaskList component
    // as the project filter is now a URL param.
  };

  const handleFiltersApply = (filters: {
    status?: TaskStatus;
    project?: number;
    title?: string;
    branch?: string;
    devEnvironment?: number;
  }) => {
    setStatusFilter(filters.status);
    setTitleFilter(filters.title);
    setBranchFilter(filters.branch);
    setDevEnvironmentFilter(filters.devEnvironment);

    loadTasks(
      1,
      filters.status,
      projectId ? parseInt(projectId, 10) : undefined,
      filters.title,
      filters.branch,
      filters.devEnvironment
    );
  };

  const handleBatchUpdateStatus = async (
    taskIds: number[],
    status: TaskStatus
  ) => {
    try {
      const response = await apiService.tasks.batchUpdateStatus({
        task_ids: taskIds,
        status,
      });

      const { success_count, failed_count } = response.data;
      if (failed_count === 0) {
        alert(
          t("tasks.messages.batchUpdateSuccess", {
            success: success_count,
            failed: failed_count,
          })
        );
      } else {
        alert(
          t("tasks.messages.batchUpdateSuccess", {
            success: success_count,
            failed: failed_count,
          })
        );
      }

      loadTasks(
        currentPage,
        statusFilter,
        projectId ? parseInt(projectId, 10) : undefined,
        titleFilter,
        branchFilter,
        devEnvironmentFilter
      );
    } catch (error) {
      logError(error as Error, "Failed to batch update task status");
      alert(
        error instanceof Error
          ? error.message
          : t("tasks.messages.batchUpdateFailed")
      );
    }
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
          devEnvironments={devEnvironments}
          loading={tasksLoading}
          currentPage={currentPage}
          totalPages={totalPages}
          total={total}
          statusFilter={statusFilter}
          projectFilter={projectId ? parseInt(projectId, 10) : undefined}
          titleFilter={titleFilter}
          branchFilter={branchFilter}
          devEnvironmentFilter={devEnvironmentFilter}
          hideProjectFilter={true}
          onPageChange={handlePageChange}
          onStatusFilterChange={handleStatusFilterChange}
          onProjectFilterChange={handleProjectFilterChange}
          onTitleFilterChange={handleTitleFilterChange}
          onBranchFilterChange={handleBranchFilterChange}
          onDevEnvironmentFilterChange={handleDevEnvironmentFilterChange}
          onFiltersApply={handleFiltersApply}
          onEdit={handleTaskEdit}
          onDelete={handleTaskDelete}
          onViewConversation={handleViewConversation}
          onViewGitDiff={handleViewGitDiff}
          onCreateNew={handleTaskCreate}
          onBatchUpdateStatus={handleBatchUpdateStatus}
        />
      </div>
    </div>
  );
};

export default TaskListPage;
