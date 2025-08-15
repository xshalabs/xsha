import { useTranslation } from "react-i18next";
import { Badge } from "@/components/ui/badge";
import {
  KanbanBoardColumn,
  KanbanBoardColumnHeader,
  KanbanBoardColumnTitle,
  KanbanColorCircle,
  KanbanBoardColumnList,
  KanbanBoardColumnListItem,
} from "@/components/kanban";
import { TaskCard } from "./TaskCard";
import type { Task, TaskStatus } from "@/types/task";

interface KanbanColumnProps {
  title: string;
  status: TaskStatus;
  tasks: Task[];
  onTaskClick: (task: Task) => void;
  onDropOverColumn: (dataTransferData: string) => void;
}

export function KanbanColumn({
  title,
  status,
  tasks,
  onTaskClick,
  onDropOverColumn,
}: KanbanColumnProps) {
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
              // For now, treat dropping over a task the same as dropping over the column
              // This ensures cross-column drops work when dropping over existing tasks
              onDropOverColumn(dataTransferData);
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
