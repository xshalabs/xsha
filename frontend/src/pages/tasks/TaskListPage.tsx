import React, {
  useState,
  useEffect,
  useMemo,
  useCallback,
  useRef,
} from "react";
import { useTranslation } from "react-i18next";
import { useParams, useNavigate, useSearchParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { usePageTitle } from "@/hooks/usePageTitle";
import { PushBranchDialog } from "@/components/PushBranchDialog";

import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { Task, TaskStatus } from "@/types/task";
import type { Project } from "@/types/project";
import { toast } from "sonner";
import { usePageActions } from "@/contexts/PageActionsContext";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";
import { DataTable } from "@/components/ui/data-table/data-table";
import { createTaskColumns } from "@/components/data-table/tasks/columns";
import { TaskDataTableToolbar } from "@/components/data-table/tasks/data-table-toolbar";
import { TaskDataTableActionBar } from "@/components/data-table/tasks/data-table-action-bar";
import { DataTablePaginationServer } from "@/components/ui/data-table/data-table-pagination-server";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";
import {
  Section,
  SectionDescription,
  SectionGroup,
  SectionHeader,
  SectionTitle,
} from "@/components/content";
import { Plus } from "lucide-react";

const TaskListPage: React.FC = () => {
  const { t } = useTranslation();
  const { projectId } = useParams<{ projectId: string }>();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { setActions } = usePageActions();
  const { setItems } = useBreadcrumb();

  // Add request deduplication
  const lastRequestRef = useRef<string>("");
  const isRequestInProgress = useRef(false);

  const [currentProject, setCurrentProject] = useState<Project | null>(null);

  const [tasks, setTasks] = useState<Task[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);

  // New DataTable state management
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);

  const [loading, setLoading] = useState(true);

  const [pushDialogOpen, setPushDialogOpen] = useState(false);
  const [selectedTaskForPush, setSelectedTaskForPush] = useState<Task | null>(null);

  usePageTitle(
    currentProject
      ? `${currentProject.name} - ${t("tasks.title")}`
      : t("tasks.title")
  );

  const pageSize = 20;







  const loadTasksData = useCallback(
    async (page: number, filters: ColumnFiltersState, sorting: SortingState, updateUrl = true) => {
      // Create a unique request key for deduplication
      const requestKey = JSON.stringify({ page, filters, sorting, updateUrl, projectId });

      // Skip if same request is already in progress or just completed
      if (
        isRequestInProgress.current ||
        lastRequestRef.current === requestKey
      ) {
        return;
      }

      isRequestInProgress.current = true;
      lastRequestRef.current = requestKey;

      try {
        setLoading(true);

        // Convert DataTable filters and sorting to API parameters
        const apiParams: any = {
          page,
          page_size: pageSize,
          project_id: projectId ? parseInt(projectId, 10) : undefined,
        };

        // Handle sorting
        if (sorting.length > 0) {
          const sort = sorting[0];
          apiParams.sort_by = sort.id;
          apiParams.sort_direction = sort.desc ? "desc" : "asc";
        }

        // Handle column filters
        filters.forEach((filter) => {
          if (filter.id === "title" && filter.value) {
            apiParams.title = filter.value as string;
          } else if (filter.id === "status" && filter.value) {
            const statusArray = filter.value as string[];
            if (statusArray.length > 0) {
              apiParams.status = statusArray.join(","); // API now supports multiple statuses
            }
          } else if (filter.id === "start_branch" && filter.value) {
            apiParams.branch = filter.value as string;
          }
        });

        const response = await apiService.tasks.list(apiParams);
        setTasks(response.data.tasks);
        setTotalPages(Math.ceil(response.data.total / pageSize));
        setTotal(response.data.total);
        setCurrentPage(page);

        // Update URL parameters
        if (updateUrl) {
          const params = new URLSearchParams();

          // Add filter parameters
          filters.forEach((filter) => {
            if (filter.value) {
              params.set(filter.id, Array.isArray(filter.value) ? filter.value.join(',') : String(filter.value));
            }
          });

          // Add sorting parameters
          if (sorting.length > 0) {
            const sort = sorting[0];
            params.set("sort_by", sort.id);
            params.set("sort_direction", sort.desc ? "desc" : "asc");
          }

          // Add page parameter (only if not page 1)
          if (page > 1) {
            params.set("page", String(page));
          }

          // Update URL without causing navigation
          setSearchParams(params, { replace: true });
        }
      } catch (error) {
        logError(error as Error, "Failed to load tasks");
      } finally {
        setLoading(false);
        isRequestInProgress.current = false;

        // Clear the request key after a short delay to allow legitimate new requests
        setTimeout(() => {
          if (lastRequestRef.current === requestKey) {
            lastRequestRef.current = "";
          }
        }, 500); // Increase delay to prevent rapid duplicate requests
      }
    },
    [pageSize, projectId, setSearchParams]
  );

  // Initialize from URL on component mount (only once)
  const [isInitialized, setIsInitialized] = useState(false);

  useEffect(() => {
    // Get URL params directly to avoid dependency issues
    const titleParam = searchParams.get("title");
    const statusParam = searchParams.get("status");
    const branchParam = searchParams.get("start_branch");
    const pageParam = searchParams.get("page");
    const sortByParam = searchParams.get("sort_by");
    const sortDirectionParam = searchParams.get("sort_direction");

    const initialFilters: ColumnFiltersState = [];

    if (titleParam) {
      initialFilters.push({ id: "title", value: titleParam });
    }

    if (statusParam) {
      initialFilters.push({ id: "status", value: statusParam.split(',') });
    }

    if (branchParam) {
      initialFilters.push({ id: "start_branch", value: branchParam });
    }

    const initialPage = pageParam ? parseInt(pageParam, 10) : 1;

    // Initialize sorting from URL
    const initialSorting: SortingState = [];
    if (sortByParam) {
      initialSorting.push({
        id: sortByParam,
        desc: sortDirectionParam === 'desc'
      });
    }

    // Set state first
    setColumnFilters(initialFilters);
    setCurrentPage(initialPage);
    setSorting(initialSorting);

    // Load initial data using the unified function
    loadTasksData(initialPage, initialFilters, initialSorting, false).then(() => {
      setIsInitialized(true);
    });

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectId]); // Only run once on mount or projectId change

  // Handle filter and sorting changes (skip initial load)
  useEffect(() => {
    if (isInitialized) {
      loadTasksData(1, columnFilters, sorting); // Reset to page 1 when filtering or sorting
    }
  }, [columnFilters, sorting, isInitialized, loadTasksData]);

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
        { type: "link", label: t("navigation.projects"), href: "/projects" },
        { type: "page", label: currentProject.name }
      ]);
    }

    // Cleanup when component unmounts
    return () => {
      setActions(null);
      setItems([]);
    };
  }, [navigate, setActions, setItems, t, projectId, currentProject]);

  const handleTaskEdit = useCallback((task: Task) => {
    navigate(`/projects/${projectId}/tasks/${task.id}/edit`);
  }, [navigate, projectId]);

  const handleTaskDelete = useCallback(async (id: number) => {
    try {
      await apiService.tasks.delete(id);
      await loadTasksData(currentPage, columnFilters, sorting);
    } catch (error) {
      // Re-throw error to let QuickActions handle the user notification
      throw error;
    }
  }, [loadTasksData, currentPage, columnFilters, sorting]);

  const handleViewConversation = useCallback((task: Task) => {
    navigate(`/projects/${projectId}/tasks/${task.id}/conversation`);
  }, [navigate, projectId]);

  const handleViewGitDiff = useCallback((task: Task) => {
    navigate(`/projects/${projectId}/tasks/${task.id}/git-diff`);
  }, [navigate, projectId]);

  const handlePageChange = useCallback((page: number) => {
    loadTasksData(page, columnFilters, sorting);
  }, [loadTasksData, columnFilters, sorting]);



  const handleBatchUpdateStatus = useCallback(async (
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

      await loadTasksData(currentPage, columnFilters, sorting);
    } catch (error) {
      logError(error as Error, "Failed to batch update task status");
      toast.error(
        error instanceof Error
          ? error.message
          : t("tasks.messages.batchUpdateFailed")
      );
    }
  }, [loadTasksData, currentPage, columnFilters, t]);

  const handleBatchDelete = useCallback(async (taskIds: number[]) => {
    try {
      // Assuming there's a batch delete API
      for (const id of taskIds) {
        await apiService.tasks.delete(id);
      }
      toast.success(t("tasks.messages.deleteSuccess"));
      await loadTasksData(currentPage, columnFilters, sorting);
    } catch (error) {
      logError(error as Error, "Failed to batch delete tasks");
      toast.error(
        error instanceof Error
          ? error.message
          : t("tasks.messages.deleteFailed")
      );
    }
  }, [loadTasksData, currentPage, columnFilters, t]);

  const handlePushBranch = useCallback(async (task: Task) => {
    if (!task.work_branch) {
      toast.error(t("tasks.messages.push_failed"), {
        description: "Task has no work branch",
      });
      return;
    }

    setSelectedTaskForPush(task);
    setPushDialogOpen(true);
  }, [t]);

  const handlePushSuccess = useCallback(async () => {
    await loadTasksData(currentPage, columnFilters, sorting);
  }, [loadTasksData, currentPage, columnFilters, sorting]);

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

        </Section>
        <Section>
          <div className="space-y-4">
            <DataTable
              columns={useMemo(() => createTaskColumns({
                t,
                onEdit: handleTaskEdit,
                onDelete: handleTaskDelete,
                onViewConversation: handleViewConversation,
                onViewGitDiff: handleViewGitDiff,
                onPushBranch: handlePushBranch,
                hideProjectColumn: true,
              }), [t, handleTaskEdit, handleTaskDelete, handleViewConversation, handleViewGitDiff, handlePushBranch])}
              data={tasks}
              toolbarComponent={(props) => (
                <TaskDataTableToolbar 
                  {...props}
                />
              )}
              actionBar={(props) => (
                <TaskDataTableActionBar
                  {...props}
                  onBatchUpdateStatus={handleBatchUpdateStatus}
                  onBatchDelete={handleBatchDelete}
                />
              )}
              columnFilters={columnFilters}
              setColumnFilters={setColumnFilters}
              sorting={sorting}
              setSorting={setSorting}
              loading={loading}
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
