import { useState, useEffect, useCallback, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  DragOverlay,
  DndContext,
  PointerSensor,
  useSensor,
  useSensors,
  closestCorners,
} from "@dnd-kit/core";
import type {
  DragEndEvent,
  DragOverEvent,
  DragStartEvent,
} from "@dnd-kit/core";
import {
  arrayMove,
  SortableContext,
  horizontalListSortingStrategy,
} from "@dnd-kit/sortable";
import {
  useSortable,
  SortableContext as ItemSortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import {
  ArrowLeft,
  Plus,
  Settings,
  MoreHorizontal,
  Calendar,
  GitBranch,
  User,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
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
import { usePageTitle } from "@/hooks/usePageTitle";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { Task, TaskStatus } from "@/types/task";
import type { Project } from "@/types/project";

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

const COLUMN_ORDER_KEY = "kanban-column-order";

// Task Card Component
function TaskCard({ task }: { task: Task }) {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const handleClick = () => {
    navigate(`/projects/${task.project_id}/tasks/${task.id}/conversation`);
  };

  const getStatusBadgeVariant = (status: TaskStatus) => {
    switch (status) {
      case "todo":
        return "secondary";
      case "in_progress":
        return "default";
      case "done":
        return "default";
      case "cancelled":
        return "destructive";
      default:
        return "secondary";
    }
  };

  const getStatusBadgeClass = (status: TaskStatus) => {
    switch (status) {
      case "todo":
        return "bg-gray-100 text-gray-800 hover:bg-gray-200";
      case "in_progress":
        return "bg-blue-100 text-blue-800 hover:bg-blue-200";
      case "done":
        return "bg-green-100 text-green-800 hover:bg-green-200";
      case "cancelled":
        return "bg-red-100 text-red-800 hover:bg-red-200";
      default:
        return "bg-gray-100 text-gray-800 hover:bg-gray-200";
    }
  };

  return (
    <Card
      className="cursor-pointer hover:shadow-md transition-shadow mb-3"
      onClick={handleClick}
    >
      <CardContent className="p-4">
        <div className="space-y-3">
          <div className="flex items-start justify-between">
            <h4 className="text-sm font-medium line-clamp-2 flex-1 mr-2">
              {task.title}
            </h4>
            <DropdownMenu>
              <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                <Button variant="ghost" size="icon" className="h-6 w-6">
                  <MoreHorizontal className="h-3 w-3" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem>{t("common.edit")}</DropdownMenuItem>
                <DropdownMenuItem>{t("common.delete")}</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <div className="flex items-center space-x-1">
              <GitBranch className="h-3 w-3" />
              <span className="truncate max-w-20">{task.work_branch}</span>
            </div>
            {task.conversation_count > 0 && (
              <Badge variant="outline" className="text-xs">
                {task.conversation_count}
              </Badge>
            )}
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-1 text-xs text-muted-foreground">
              <User className="h-3 w-3" />
              <span>{task.created_by}</span>
            </div>
            <div className="flex items-center space-x-1 text-xs text-muted-foreground">
              <Calendar className="h-3 w-3" />
              <span>{new Date(task.created_at).toLocaleDateString()}</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

// Draggable Task Card
function DraggableTaskCard({ task }: { task: Task }) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: task.id,
    data: {
      type: "task",
      task,
    },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  if (isDragging) {
    return (
      <div
        ref={setNodeRef}
        style={style}
        className="opacity-50 bg-background border-2 border-dashed border-border rounded-lg h-32"
      />
    );
  }

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}>
      <TaskCard task={task} />
    </div>
  );
}

// Kanban Column Component
function KanbanColumn({
  title,
  status,
  tasks,
  onAddTask,
}: {
  title: string;
  status: TaskStatus;
  tasks: Task[];
  onAddTask: () => void;
}) {
  const { t } = useTranslation();
  const taskIds = tasks.map((task) => task.id);

  const getColumnColor = (status: TaskStatus) => {
    switch (status) {
      case "todo":
        return "border-l-gray-500";
      case "in_progress":
        return "border-l-blue-500";
      case "done":
        return "border-l-green-500";
      case "cancelled":
        return "border-l-red-500";
      default:
        return "border-l-gray-500";
    }
  };

  return (
    <div
      className={`flex flex-col h-full min-w-80 border-l-4 ${getColumnColor(
        status
      )} bg-muted/20 rounded-lg`}
    >
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-semibold flex items-center space-x-2">
            <span>{title}</span>
            <Badge variant="secondary" className="text-xs">
              {tasks.length}
            </Badge>
          </CardTitle>
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            onClick={onAddTask}
          >
            <Plus className="h-3 w-3" />
          </Button>
        </div>
      </CardHeader>

      <CardContent className="flex-1 pt-0 px-4 pb-4 overflow-y-auto">
        <ItemSortableContext
          items={taskIds}
          strategy={verticalListSortingStrategy}
        >
          <div className="space-y-3">
            {tasks.map((task) => (
              <DraggableTaskCard key={task.id} task={task} />
            ))}
            {tasks.length === 0 && (
              <div className="text-center text-muted-foreground text-sm py-8">
                {t("tasks.no_tasks")}
              </div>
            )}
          </div>
        </ItemSortableContext>
      </CardContent>
    </div>
  );
}

// Draggable Column
function DraggableColumn({
  title,
  status,
  tasks,
  onAddTask,
}: {
  title: string;
  status: TaskStatus;
  tasks: Task[];
  onAddTask: () => void;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: status,
    data: {
      type: "column",
      status,
    },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className={isDragging ? "opacity-50" : ""}
    >
      <KanbanColumn
        title={title}
        status={status}
        tasks={tasks}
        onAddTask={onAddTask}
      />
    </div>
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
  const [activeId, setActiveId] = useState<string | number | null>(null);
  const [activeTask, setActiveTask] = useState<Task | null>(null);

  // Column order management
  const [columnOrder, setColumnOrder] = useState<string[]>(() => {
    const saved = localStorage.getItem(COLUMN_ORDER_KEY);
    return saved ? JSON.parse(saved) : KANBAN_COLUMNS.map((col) => col.id);
  });

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 3,
      },
    })
  );

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

  // Save column order to localStorage
  const saveColumnOrder = useCallback((order: string[]) => {
    setColumnOrder(order);
    localStorage.setItem(COLUMN_ORDER_KEY, JSON.stringify(order));
  }, []);

  // Handle drag start
  const handleDragStart = useCallback((event: DragStartEvent) => {
    const { active } = event;
    setActiveId(active.id);

    if (active.data.current?.type === "task") {
      setActiveTask(active.data.current.task);
    }
  }, []);

  // Handle drag end
  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      const { active, over } = event;

      setActiveId(null);
      setActiveTask(null);

      if (!over) return;

      const activeId = active.id;
      const overId = over.id;

      // Handle column reordering
      if (
        active.data.current?.type === "column" &&
        over.data.current?.type === "column"
      ) {
        const oldIndex = columnOrder.indexOf(String(activeId));
        const newIndex = columnOrder.indexOf(String(overId));

        if (oldIndex !== newIndex) {
          const newOrder = arrayMove(columnOrder, oldIndex, newIndex);
          saveColumnOrder(newOrder);
        }
        return;
      }

      // Handle task movement
      if (active.data.current?.type === "task") {
        const task = active.data.current.task as Task;
        const targetStatus = over.data.current?.status || String(overId);

        if (task.status !== targetStatus) {
          try {
            // Update task status via API
            await apiService.tasks.batchUpdateStatus({
              task_ids: [task.id],
              status: targetStatus as TaskStatus,
            });

            // Update local state
            setTasks((prev) => {
              const newTasks = { ...prev };

              // Remove task from old column
              newTasks[task.status] = newTasks[task.status].filter(
                (t) => t.id !== task.id
              );

              // Add task to new column with updated status
              const updatedTask = {
                ...task,
                status: targetStatus as TaskStatus,
              };
              newTasks[targetStatus as TaskStatus] = [
                ...newTasks[targetStatus as TaskStatus],
                updatedTask,
              ];

              return newTasks;
            });
          } catch (error) {
            logError(error as Error, "Failed to update task status");
          }
        }
      }
    },
    [columnOrder, saveColumnOrder]
  );

  // Handle drag over (for better visual feedback)
  const handleDragOver = useCallback((event: DragOverEvent) => {
    const { active, over } = event;

    if (!over) return;

    // Only handle task over column
    if (
      active.data.current?.type === "task" &&
      over.data.current?.type === "column"
    ) {
      // Could add visual feedback here if needed
    }
  }, []);

  const orderedColumns = useMemo(() => {
    return columnOrder.map(
      (columnId) => KANBAN_COLUMNS.find((col) => col.id === columnId)!
    );
  }, [columnOrder]);

  const handleGoBack = () => {
    navigate(`/projects/${projectId}/tasks`);
  };

  const handleAddTask = () => {
    navigate(`/projects/${projectId}/tasks/create`);
  };

  const handleProjectSettings = () => {
    navigate(`/projects/${projectId}/edit`);
  };

  const handleProjectChange = (newProjectId: string) => {
    navigate(`/projects/${newProjectId}/kanban`);
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
      <header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="flex h-16 items-center px-6">
          <div className="flex items-center space-x-4">
            <Button
              variant="ghost"
              size="icon"
              onClick={handleGoBack}
              className="h-8 w-8"
            >
              <ArrowLeft className="h-4 w-4" />
            </Button>

            <Select value={projectId} onValueChange={handleProjectChange}>
              <SelectTrigger className="min-w-48 border-none shadow-none text-lg font-semibold bg-transparent">
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

          <div className="ml-auto flex items-center space-x-4">
            <Button onClick={handleAddTask} size="sm">
              <Plus className="h-4 w-4 mr-2" />
              {t("tasks.addTask")}
            </Button>

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
        <DndContext
          sensors={sensors}
          collisionDetection={closestCorners}
          onDragStart={handleDragStart}
          onDragOver={handleDragOver}
          onDragEnd={handleDragEnd}
        >
          <div className="flex space-x-6 h-[calc(100vh-8rem)] overflow-x-auto">
            <SortableContext
              items={columnOrder}
              strategy={horizontalListSortingStrategy}
            >
              {orderedColumns.map((column) => (
                <DraggableColumn
                  key={column.id}
                  title={t(`tasks.status.${column.status}`)}
                  status={column.status}
                  tasks={tasks[column.status] || []}
                  onAddTask={handleAddTask}
                />
              ))}
            </SortableContext>
          </div>

          <DragOverlay>
            {activeTask && <TaskCard task={activeTask} />}
          </DragOverlay>
        </DndContext>
      </main>
    </div>
  );
}
