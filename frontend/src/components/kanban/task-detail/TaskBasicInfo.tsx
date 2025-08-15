import { memo } from "react";
import { useTranslation } from "react-i18next";
import {
  GitBranch,
  User,
  FileText,
  Clock,
  Monitor,
  Calendar,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import type { Task } from "@/types/task";
import { getStatusBadgeClass, formatDate } from "./utils";

interface TaskBasicInfoProps {
  task: Task;
}

export const TaskBasicInfo = memo<TaskBasicInfoProps>(({ task }) => {
  const { t } = useTranslation();

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  return (
    <div className="space-y-4">
      <h3 className="font-medium text-foreground text-lg flex items-center gap-2">
        <FileText className="h-5 w-5" />
        {t("tasks.tabs.basic")}
      </h3>
      
      <div className="grid grid-cols-2 gap-4 text-sm">
        <div>
          <span className="font-medium text-foreground">
            {t("tasks.status.label")}:
          </span>
          <Badge
            className={`ml-2 ${getStatusBadgeClass(task.status)}`}
          >
            {t(`tasks.status.${task.status}`)}
          </Badge>
        </div>

        <div className="flex items-center">
          <GitBranch className="h-4 w-4 mr-1" />
          <span className="font-medium text-foreground">
            {t("tasks.workBranch")}:
          </span>
          <span className="ml-2 font-mono text-xs">
            {task.work_branch}
          </span>
        </div>

        <div className="flex items-center">
          <GitBranch className="h-4 w-4 mr-1" />
          <span className="font-medium text-foreground">
            {t("tasks.startBranch")}:
          </span>
          <span className="ml-2 font-mono text-xs">
            {task.start_branch}
          </span>
        </div>

        <div className="flex items-center">
          <Monitor className="h-4 w-4 mr-1" />
          <span className="font-medium text-foreground">
            {t("tasks.environment")}:
          </span>
          <span className="ml-2">
            {task.dev_environment?.name || "-"}
          </span>
        </div>

        <div className="flex items-center">
          <Clock className="h-4 w-4 mr-1" />
          <span className="font-medium text-foreground">
            {t("tasks.executionTime")}:
          </span>
          <span className="ml-2">
            {task.latest_execution_time 
              ? formatTime(task.latest_execution_time)
              : t("common.notSet")
            }
          </span>
        </div>

        <div className="flex items-center">
          <User className="h-4 w-4 mr-1" />
          <span className="font-medium text-foreground">
            {t("tasks.createdBy")}:
          </span>
          <span className="ml-2">{task.created_by}</span>
        </div>

        <div className="flex items-center">
          <Calendar className="h-4 w-4 mr-1" />
          <span className="font-medium text-foreground">
            {t("tasks.createdAt")}:
          </span>
          <span className="ml-2">
            {formatDate(task.created_at)}
          </span>
        </div>
      </div>
    </div>
  );
});

TaskBasicInfo.displayName = "TaskBasicInfo";
