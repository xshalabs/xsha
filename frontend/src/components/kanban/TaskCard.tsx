import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { AlarmClock } from "lucide-react";
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

  // Check if latest_execution_time is in the future
  const hasFutureExecutionTime = () => {
    if (!task.latest_execution_time) return false;
    const executionTime = new Date(task.latest_execution_time);
    const now = new Date();
    return executionTime > now;
  };

  const showAlarmIcon = hasFutureExecutionTime();

  return (
    <KanbanBoardCard data={{ id: task.id.toString() }} onClick={handleClick}>
      <KanbanBoardCardTitle>{task.title}</KanbanBoardCardTitle>
      {showAlarmIcon && (
        <div className="flex items-center justify-between mt-2">
          <div className="flex items-center space-x-2">
            <Tooltip>
              <TooltipTrigger>
                <div className="flex items-center">
                  <AlarmClock className="h-3 w-3 text-orange-500" />
                </div>
              </TooltipTrigger>
              <TooltipContent>
                <p>Scheduled execution time</p>
              </TooltipContent>
            </Tooltip>
          </div>
        </div>
      )}
    </KanbanBoardCard>
  );
}
