import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";

import {
  GitBranch,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Loader2,
} from "lucide-react";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { Task } from "@/types/task";

interface PushBranchDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  task: Task | null;
  onSuccess?: () => void;
}

type PushStep = "confirm" | "final-confirm" | "pushing" | "result";

interface PushResult {
  success: boolean;
  message: string;
  output: string;
  error?: string;
  details?: string;
}

export function PushBranchDialog({
  open,
  onOpenChange,
  task,
  onSuccess,
}: PushBranchDialogProps) {
  const { t } = useTranslation();
  const [step, setStep] = useState<PushStep>("confirm");
  const [pushResult, setPushResult] = useState<PushResult | null>(null);
  const [forcePush, setForcePush] = useState(false);

  useEffect(() => {
    if (!open || !task) {
      setStep("confirm");
      setPushResult(null);
      setForcePush(false);
    }
  }, [open, task]);

  const handleFirstConfirm = () => {
    setStep("final-confirm");
  };

  const handleFinalConfirm = async () => {
    if (!task) return;

    setStep("pushing");

    try {
      const response = await apiService.tasks.pushTaskBranch(
        task.id,
        forcePush
      );

      setPushResult({
        success: true,
        message: response.message,
        output: response.data.output,
      });

      setStep("result");

      if (onSuccess) {
        onSuccess();
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      const errorDetails =
        error instanceof Error && "details" in error
          ? (error as any).details
          : undefined;
      logError(error as Error, "Failed to push task branch");

      setPushResult({
        success: false,
        message: t("tasks.messages.push_failed"),
        output: "",
        error: errorMessage,
        details: errorDetails,
      });

      setStep("result");
    }
  };

  const handleCancel = () => {
    if (step === "pushing") {
      return;
    }
    onOpenChange(false);
  };

  const handleClose = () => {
    onOpenChange(false);
  };

  const renderContent = () => {
    switch (step) {
      case "confirm":
        return (
          <>
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2 text-foreground">
                <GitBranch className="h-5 w-5" />
                {t("tasks.push.confirm_title")}
              </DialogTitle>
              <DialogDescription className="text-muted-foreground">
                {t("tasks.push.confirm_description")}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 my-4">
              <div className="rounded-lg text-foreground space-y-3">
                <div className="flex justify-between items-start">
                  <span className="text-sm font-medium flex-shrink-0">
                    {t("tasks.fields.title")}:
                  </span>
                  <span className="text-sm break-words text-right ml-2">
                    {task?.title}
                  </span>
                </div>
                <div className="flex justify-between items-start">
                  <span className="text-sm font-medium flex-shrink-0">
                    {t("tasks.fields.work_branch")}:
                  </span>
                  <Badge
                    variant="outline"
                    className="font-mono text-xs break-all ml-2"
                  >
                    {task?.work_branch}
                  </Badge>
                </div>
                <div className="flex justify-between items-start">
                  <span className="text-sm font-medium flex-shrink-0">
                    {t("tasks.fields.project")}:
                  </span>
                  <span className="text-sm text-muted-foreground break-words text-right ml-2">
                    {task?.project?.name}
                  </span>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-amber-50 border border-amber-200 rounded-lg">
                <AlertTriangle className="h-5 w-5 text-amber-600 mt-0.5 flex-shrink-0" />
                <div className="text-sm">
                  <p className="font-medium text-amber-800">
                    {t("tasks.push.warning_title")}
                  </p>
                  <p className="text-amber-700 mt-1">
                    {t("tasks.push.warning_description")}
                  </p>
                </div>
              </div>

              <div className="flex items-center space-x-2 p-3 border border-border rounded-lg">
                <Checkbox
                  id="force-push"
                  checked={forcePush}
                  onCheckedChange={(checked) => setForcePush(!!checked)}
                />
                <label
                  htmlFor="force-push"
                  className="text-sm font-medium text-foreground leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  {t("tasks.push.force_push")}
                </label>
              </div>

              {forcePush && (
                <div className="flex items-start gap-3 p-3 bg-red-50 border border-red-200 rounded-lg">
                  <AlertTriangle className="h-5 w-5 text-red-600 mt-0.5 flex-shrink-0" />
                  <div className="text-sm">
                    <p className="font-medium text-red-800">
                      {t("tasks.push.force_push_warning_title")}
                    </p>
                    <p className="text-red-700 mt-1">
                      {t("tasks.push.force_push_warning_description")}
                    </p>
                  </div>
                </div>
              )}
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                className="text-foreground hover:text-foreground"
                onClick={handleCancel}
              >
                {t("common.cancel")}
              </Button>
              <Button onClick={handleFirstConfirm}>
                {t("common.continue")}
              </Button>
            </DialogFooter>
          </>
        );

      case "final-confirm":
        return (
          <>
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2 text-red-600">
                <AlertTriangle className="h-5 w-5" />
                {t("tasks.push.final_confirm_title")}
              </DialogTitle>
              <DialogDescription>
                {t("tasks.push.final_confirm_description")}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 my-4">
              <div className="bg-red-50 border border-red-200 p-4 rounded-lg">
                <div className="text-center space-y-3">
                  <div className="text-lg font-semibold text-red-800">
                    {t("tasks.push.final_warning")}
                  </div>
                  <div className="text-sm text-red-700">
                    {t("tasks.push.final_warning_details", {
                      branch: task?.work_branch,
                      repository: task?.project?.name,
                    })}
                  </div>
                </div>
              </div>
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                className="text-foreground hover:text-foreground"
                onClick={() => setStep("confirm")}
              >
                {t("common.back")}
              </Button>
              <Button variant="destructive" onClick={handleFinalConfirm}>
                {t("tasks.push.confirm_push")}
              </Button>
            </DialogFooter>
          </>
        );

      case "pushing":
        return (
          <>
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2 text-foreground">
                <Loader2 className="h-5 w-5 animate-spin text-primary" />
                {t("tasks.push.pushing_title")}
              </DialogTitle>
              <DialogDescription>
                {t("tasks.push.pushing_description")}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 my-4">
              <div className="text-center py-8">
                <Loader2 className="h-12 w-12 animate-spin mx-auto text-primary mb-4" />
                <div className="text-lg font-medium mb-2 text-foreground">
                  {t("tasks.push.pushing_to", { branch: task?.work_branch })}
                </div>
                <div className="text-sm text-muted-foreground">
                  {t("tasks.push.please_wait")}
                </div>
              </div>
            </div>
          </>
        );

      case "result":
        return (
          <>
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2 text-foreground">
                {pushResult?.success ? (
                  <CheckCircle className="h-5 w-5 text-green-600" />
                ) : (
                  <XCircle className="h-5 w-5 text-red-600" />
                )}
                {pushResult?.success
                  ? t("tasks.push.success_title")
                  : t("tasks.push.error_title")}
              </DialogTitle>
              <DialogDescription className="text-muted-foreground">
                {pushResult?.message}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 my-4">
              {pushResult?.success ? (
                <div className="bg-green-50 border border-green-200 p-4 rounded-lg">
                  <div className="flex items-center gap-2 text-green-800 font-medium mb-2">
                    <CheckCircle className="h-4 w-4" />
                    {t("tasks.push.success_message")}
                  </div>
                  <div className="text-sm text-green-700">
                    {t("tasks.push.success_details", {
                      branch: task?.work_branch,
                      repository: task?.project?.name,
                    })}
                  </div>
                </div>
              ) : (
                <div className="bg-red-50 border border-red-200 p-4 rounded-lg">
                  <div className="flex items-center gap-2 text-red-800 font-medium mb-2">
                    <XCircle className="h-4 w-4" />
                    {t("tasks.push.error_message")}
                  </div>
                  {pushResult?.error && (
                    <div className="text-sm text-red-700 mt-2 break-words">
                      <strong>{t("common.error")}:</strong>{" "}
                      <span className="whitespace-pre-wrap">
                        {pushResult.error}
                      </span>
                    </div>
                  )}
                  {pushResult?.details && (
                    <div className="text-sm text-red-700 mt-2 pt-2 border-t border-red-200 break-words">
                      <strong>{t("common.details")}:</strong>{" "}
                      <span className="whitespace-pre-wrap">
                        {pushResult.details}
                      </span>
                    </div>
                  )}
                </div>
              )}

              {pushResult?.output && (
                <div className="space-y-2">
                  <div className="flex items-center">
                    <span className="text-sm font-medium text-foreground">
                      {t("tasks.push.output")}:
                    </span>
                  </div>

                  <Textarea
                    value={pushResult.output}
                    readOnly
                    className="font-mono text-xs text-muted-foreground resize-none overflow-auto"
                    rows={8}
                  />
                </div>
              )}
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                className="text-foreground hover:text-foreground"
                onClick={handleClose}
              >
                {t("common.close")}
              </Button>
            </DialogFooter>
          </>
        );

      default:
        return null;
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={step === "pushing" ? undefined : onOpenChange}
    >
      <DialogContent
        className="max-w-2xl max-h-[90vh] flex flex-col"
        showCloseButton={step !== "pushing"}
      >
        <div className="flex-1 overflow-y-auto">{renderContent()}</div>
      </DialogContent>
    </Dialog>
  );
}
