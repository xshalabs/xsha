import { memo, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { GitBranch, Eye, Zap, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { Task } from "@/types/task";

interface TaskActionsProps {
  task: Task;
  onPushBranch: () => void;
  onViewGitDiff: () => void;
  onDelete: () => void;
}

export const TaskActions = memo<TaskActionsProps>(
  ({ task, onPushBranch, onViewGitDiff, onDelete }) => {
    const { t } = useTranslation();

    const handlePushBranch = useCallback(() => {
      onPushBranch();
    }, [onPushBranch]);

    const handleViewGitDiff = useCallback(() => {
      onViewGitDiff();
    }, [onViewGitDiff]);

    const handleDelete = useCallback(() => {
      onDelete();
    }, [onDelete]);

    return (
      <div className="border-y py-6 space-y-6 px-6">
        <h4 className="font-medium text-base text-foreground flex items-center gap-2">
          <Zap className="h-4 w-4" />
          {t("tasks.actions.title")}
        </h4>
        <div className="flex flex-wrap gap-3">
          <Button
            onClick={handlePushBranch}
            className="flex items-center gap-2"
            variant="outline"
            aria-label={t("tasks.actions.pushBranch")}
          >
            <GitBranch className="h-4 w-4" />
            {t("tasks.actions.pushBranch")}
          </Button>

          <Button
            onClick={handleViewGitDiff}
            variant="outline"
            className="flex items-center gap-2"
            aria-label={t("tasks.actions.viewGitDiff")}
          >
            <Eye className="h-4 w-4" />
            {t("tasks.actions.viewGitDiff")}
          </Button>

          <Button
            onClick={handleDelete}
            variant="destructive"
            className="flex items-center gap-2"
            aria-label={t("tasks.actions.delete")}
          >
            <Trash2 className="h-4 w-4" />
            {t("tasks.actions.delete")}
          </Button>
        </div>
      </div>
    );
  }
);

TaskActions.displayName = "TaskActions";
