import { Badge } from "@/components/ui/badge";
import {
  KanbanBoardCard,
  KanbanBoardCardTitle,
} from "@/components/kanban";
import type { Task } from "@/types/task";

interface TaskCardProps {
  task: Task;
  onClick?: () => void;
}

export function TaskCard({ task, onClick }: TaskCardProps) {
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
