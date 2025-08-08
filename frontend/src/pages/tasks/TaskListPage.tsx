import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useParams, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { usePageTitle } from "@/hooks/usePageTitle";
import { TaskList } from "@/components/TaskList";
import { PushBranchDialog } from "@/components/PushBranchDialog";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { Task, TaskStatus } from "@/types/task";
import type { Project } from "@/types/project";
import type { DevEnvironment } from "@/types/dev-environment";
import { toast } from "sonner";
import { usePageActions } from "@/contexts/PageActionsContext";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";
import { DataTablePaginationServer } from "@/components/ui/data-table/data-table-pagination-server";
import {
  Section,
  SectionDescription,
  SectionGroup,
  SectionHeader,
  SectionTitle,
} from "@/components/content";
import {
  MetricCardGroup,
  MetricCardHeader,
  MetricCardTitle,
  MetricCardValue,
  MetricCardButton,
} from "@/components/metric";

import { Plus, CheckCircle, Clock, Play, X, ListFilter } from "lucide-react";

const TaskListPage: React.FC = () => {
  const { t } = useTranslation();
  const { projectId } = useParams<{ projectId: string }>();
  const navigate = useNavigate();
  const { setActions } = usePageActions();
  const { setItems } = useBreadcrumb();

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

  const [pushDialogOpen, setPushDialogOpen] = useState(false);
  const [selectedTaskForPush, setSelectedTaskForPush] = useState<Task | null>(null);

  usePageTitle(
    currentProject
      ? `${currentProject.name} - ${t("tasks.title")}`
      : t("tasks.title")
  );

  const pageSize = 20;

  // Metrics for task status
  const metrics = React.useMemo(() => [
    {
      title: t("tasks.status.todo"),
      value: tasks.filter((task) => task.status === "todo").length,
      variant: "default" as const,
      type: "filter" as const,
      status: "todo" as TaskStatus,
    },
    {
      title: t("tasks.status.in_progress"),
      value: tasks.filter((task) => task.status === "in_progress").length,
      variant: "warning" as const,
      type: "filter" as const,
      status: "in_progress" as TaskStatus,
    },
    {
      title: t("tasks.status.done"),
      value: tasks.filter((task) => task.status === "done").length,
      variant: "success" as const,
      type: "filter" as const,
      status: "done" as TaskStatus,
    },
    {
      title: t("tasks.status.cancelled"),
      value: tasks.filter((task) => task.status === "cancelled").length,
      variant: "destructive" as const,
      type: "filter" as const,
      status: "cancelled" as TaskStatus,
    },
    {
      title: t("tasks.metrics.total"),
      value: total,
      variant: "ghost" as const,
      type: "info" as const,
    },
  ], [tasks, total, t]);

  const icons = {
    filter: {
      active: CheckCircle,
      inactive: ListFilter,
    },
    info: {
      active: ListFilter,
      inactive: ListFilter,
    },
  };

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

  // Set page actions (Create button in header) and breadcrumb
  useEffect(() => {
    const handleCreateNew = () => {
      navigate(`/projects/${projectId}/tasks/create`);
    };

    setActions(
      <Button onClick={handleCreateNew} size="sm">
        <Plus className="h-4 w-4 mr-2" />
        {t("tasks.create")}
      </Button>
    );

    // Set breadcrumb items
    if (currentProject) {
      setItems([
        { label: t("navigation.projects"), href: "/projects" },
        { label: currentProject.name, href: `/projects/${projectId}/tasks` }
      ]);
    }

    // Cleanup when component unmounts
    return () => {
      setActions(null);
      setItems([]);
    };
  }, [navigate, setActions, setItems, t, projectId, currentProject]);

  const handleTaskCreate = () => {
    navigate(`/projects/${projectId}/tasks/create`);
  };

  const handleTaskEdit = (task: Task) => {
    navigate(`/projects/${projectId}/tasks/${task.id}/edit`);
  };

  const handleTaskDelete = async (id: number) => {
    try {
      await apiService.tasks.delete(id);
      toast.success(t("tasks.messages.deleteSuccess"));
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
      toast.error(
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
    setCurrentPage(page);
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
    setCurrentPage(1); // Reset to first page when filtering
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
    setCurrentPage(1);
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
    setCurrentPage(1);
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
    setCurrentPage(1);
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
    setCurrentPage(1);

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
        toast.success(
          t("tasks.messages.batchUpdateSuccess", {
            success: success_count,
            failed: failed_count,
          })
        );
      } else {
        toast.warning(
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
      toast.error(
        error instanceof Error
          ? error.message
          : t("tasks.messages.batchUpdateFailed")
      );
    }
  };

  const handlePushBranch = async (task: Task) => {
    if (!task.work_branch) {
      toast.error(t("tasks.messages.push_failed"), {
        description: "Task has no work branch",
      });
      return;
    }

    setSelectedTaskForPush(task);
    setPushDialogOpen(true);
  };

  const handlePushSuccess = async () => {
    await loadTasks(
      currentPage,
      statusFilter,
      projectId ? parseInt(projectId, 10) : undefined,
      titleFilter,
      branchFilter,
      devEnvironmentFilter
    );
  };

  return (
    <div className="min-h-screen bg-background">
      <SectionGroup>
        <Section>
          <SectionHeader>
            <SectionTitle>
              {currentProject ? `${currentProject.name} - ${t("tasks.title")}` : t("tasks.title")}
            </SectionTitle>
            <SectionDescription>
              {t("tasks.page_description")}
            </SectionDescription>
          </SectionHeader>
          <MetricCardGroup>
            {metrics.map((metric) => {
              const isFilterActive = statusFilter === metric.status;
              const isActive = metric.type === "filter" ? isFilterActive : false;
              const Icon = icons[metric.type][isActive ? "active" : "inactive"];

              return (
                <MetricCardButton
                  key={metric.title}
                  variant={metric.variant}
                  onClick={() => {
                    if (metric.type === "filter" && metric.status) {
                      if (!isFilterActive) {
                        handleStatusFilterChange(metric.status);
                      } else {
                        handleStatusFilterChange(undefined);
                      }
                    }
                  }}
                  disabled={metric.type === "info"}
                  className={metric.type === "info" ? "cursor-default" : ""}
                >
                  <MetricCardHeader className="flex justify-between items-center gap-2 w-full">
                    <MetricCardTitle className="truncate">
                      {metric.title}
                    </MetricCardTitle>
                    {metric.type === "filter" && <Icon className="size-4" />}
                  </MetricCardHeader>
                  <MetricCardValue>{metric.value}</MetricCardValue>
                </MetricCardButton>
              );
            })}
          </MetricCardGroup>
        </Section>
        <Section>
          <div className="space-y-4">
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
              onPushBranch={handlePushBranch}
              onCreateNew={handleTaskCreate}
              onBatchUpdateStatus={handleBatchUpdateStatus}
              hidePagination={true}
            />
            <DataTablePaginationServer
              currentPage={currentPage}
              totalPages={totalPages}
              total={total}
              onPageChange={handlePageChange}
            />
          </div>
        </Section>
      </SectionGroup>

      <PushBranchDialog
        open={pushDialogOpen}
        onOpenChange={setPushDialogOpen}
        task={selectedTaskForPush}
        onSuccess={handlePushSuccess}
      />
    </div>
  );
};

export default TaskListPage;
