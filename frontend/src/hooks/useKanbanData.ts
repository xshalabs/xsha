import { useState, useEffect, useCallback } from "react";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { Task, TaskStatus } from "@/types/task";
import type { Project } from "@/types/project";

interface KanbanTasks {
  todo: Task[];
  in_progress: Task[];
  done: Task[];
  cancelled: Task[];
}

export function useKanbanData(projectId: string | undefined) {
  const [project, setProject] = useState<Project | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [tasks, setTasks] = useState<KanbanTasks>({
    todo: [],
    in_progress: [],
    done: [],
    cancelled: [],
  });
  const [loading, setLoading] = useState(true);

  const loadData = useCallback(async () => {
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
  }, [projectId]);

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

  const refreshKanbanData = useCallback(async () => {
    if (!projectId) return;

    try {
      const kanbanResponse = await apiService.tasks.getKanbanTasks(
        parseInt(projectId)
      );
      setTasks(kanbanResponse.data);
    } catch (error) {
      console.error("Failed to refresh kanban data:", error);
      logError(error as Error, "Failed to refresh kanban data");
    }
  }, [projectId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  return {
    project,
    projects,
    tasks,
    loading,
    handleDropOverColumn,
    refreshKanbanData,
  };
}
