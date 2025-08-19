import { memo } from "react";
import { useTranslation } from "react-i18next";
import {
  GitBranch,
  User,
  FileText,
  Monitor,
  Calendar,
  Activity,
  Copy,
} from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import type { Task } from "@/types/task";
import { getStatusBadgeClass, formatDate } from "./utils";

interface TaskBasicInfoProps {
  task: Task;
}

export const TaskBasicInfo = memo<TaskBasicInfoProps>(({ task }) => {
  const { t } = useTranslation();

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast.success(t("common.copied_to_clipboard"));
    } catch (err) {
      console.error("Failed to copy text: ", err);
      toast.error(t("common.copy_failed"));

      try {
        const textarea = document.createElement("textarea");
        textarea.value = text;
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand("copy");
        document.body.removeChild(textarea);
        toast.success(t("common.copied_to_clipboard"));
      } catch (fallbackErr) {
        console.error("Fallback copy also failed:", fallbackErr);
        toast.error(t("common.copy_not_supported"));
      }
    }
  };

  return (
    <div className="space-y-6 px-6">
      <h3 className="font-medium text-foreground text-base flex items-center gap-2">
        <FileText className="h-4 w-4" />
        {t("tasks.tabs.basic")}
      </h3>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4 text-sm">
        <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-0">
          <span className="font-medium text-foreground flex items-center gap-1 text-sm">
            <Activity className="h-3 w-3" />
            {t("tasks.status.label")}:
          </span>
          <Badge
            className={`sm:ml-2 w-fit ${getStatusBadgeClass(task.status)}`}
          >
            {t(`tasks.status.${task.status}`)}
          </Badge>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-0 min-w-0 sm:col-span-2 lg:col-span-1">
          <span className="font-medium text-foreground flex items-center gap-1 flex-shrink-0 text-sm">
            <GitBranch className="h-3 w-3" />
            {t("tasks.workBranch")}:
          </span>
          <div className="flex items-center gap-2 sm:ml-2 min-w-0 flex-1">
            <span
              className="font-mono text-xs bg-muted px-2 py-1 rounded truncate flex-1 min-w-0"
              title={task.work_branch}
            >
              {task.work_branch}
            </span>
            <button
              onClick={() => copyToClipboard(task.work_branch)}
              className="p-1 hover:bg-muted rounded transition-colors flex-shrink-0"
              title="复制工作分支"
            >
              <Copy className="h-3 w-3 text-muted-foreground hover:text-foreground" />
            </button>
          </div>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-0 min-w-0">
          <span className="font-medium text-foreground flex items-center gap-1 flex-shrink-0 text-sm">
            <GitBranch className="h-3 w-3" />
            {t("tasks.startBranch")}:
          </span>
          <span
            className="sm:ml-2 font-mono text-xs bg-muted px-2 py-1 rounded truncate"
            title={task.start_branch}
          >
            {task.start_branch}
          </span>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-0 min-w-0">
          <span className="font-medium text-foreground flex items-center gap-1 flex-shrink-0 text-sm">
            <Monitor className="h-3 w-3" />
            {t("tasks.environment")}:
          </span>
          <span
            className="sm:ml-2 truncate"
            title={task.dev_environment?.name || "-"}
          >
            {task.dev_environment?.name || "-"}
          </span>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-0 min-w-0">
          <span className="font-medium text-foreground flex items-center gap-1 flex-shrink-0 text-sm">
            <User className="h-3 w-3" />
            {t("tasks.createdBy")}:
          </span>
          <span className="sm:ml-2 truncate" title={task.created_by}>
            {task.created_by}
          </span>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-0 min-w-0">
          <span className="font-medium text-foreground flex items-center gap-1 flex-shrink-0 text-sm">
            <Calendar className="h-3 w-3" />
            {t("tasks.createdAt")}:
          </span>
          <span
            className="sm:ml-2 text-xs truncate"
            title={formatDate(task.created_at)}
          >
            {formatDate(task.created_at)}
          </span>
        </div>
      </div>
    </div>
  );
});

TaskBasicInfo.displayName = "TaskBasicInfo";
