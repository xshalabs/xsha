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

import {
  GitBranch,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Loader2,
  Copy,
  Eye,
  EyeOff,
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
  const [showFullOutput, setShowFullOutput] = useState(false);

  useEffect(() => {
    if (!open || !task) {
      setStep("confirm");
      setPushResult(null);
      setShowFullOutput(false);
    }
  }, [open, task]);

  const handleFirstConfirm = () => {
    setStep("final-confirm");
  };

  const handleFinalConfirm = async () => {
    if (!task) return;

    setStep("pushing");

    try {
      const response = await apiService.tasks.pushTaskBranch(task.id);

      setPushResult({
        success: true,
        message: response.message,
        output: response.data.output,
      });

      setStep("result");

      // 调用成功回调
      if (onSuccess) {
        onSuccess();
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      const errorDetails = error instanceof Error && 'details' in error 
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

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
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

            <div className="space-y-4">
              <div className="p-4 rounded-lg text-foreground">
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-sm font-medium">
                      {t("tasks.fields.title")}:
                    </span>
                    <span className="text-sm">{task?.title}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm font-medium">
                      {t("tasks.fields.work_branch")}:
                    </span>
                    <Badge variant="outline" className="font-mono">
                      {task?.work_branch}
                    </Badge>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm font-medium">
                      {t("tasks.fields.project")}:
                    </span>
                    <span className="text-sm text-muted-foreground">
                      {task?.project?.name}
                    </span>
                  </div>
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

            <div className="space-y-4">
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

            <div className="space-y-4">
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

            <div className="space-y-4">
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
                    <div className="text-sm text-red-700 mt-2">
                      <strong>{t("common.error")}:</strong> {pushResult.error}
                    </div>
                  )}
                  {pushResult?.details && (
                    <div className="text-sm text-red-700 mt-2 pt-2 border-t border-red-200">
                      <strong>{t("common.details")}:</strong> {pushResult.details}
                    </div>
                  )}
                </div>
              )}

              {pushResult?.output && (
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">
                      {t("tasks.push.output")}:
                    </span>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setShowFullOutput(!showFullOutput)}
                      >
                        {showFullOutput ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                        {showFullOutput ? t("common.hide") : t("common.show")}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => copyToClipboard(pushResult.output)}
                      >
                        <Copy className="h-4 w-4" />
                        {t("common.copy")}
                      </Button>
                    </div>
                  </div>

                  <Textarea
                    value={pushResult.output}
                    readOnly
                    className="font-mono text-xs bg-gray-50"
                    rows={showFullOutput ? 10 : 4}
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
        className="max-w-2xl max-h-[90vh] overflow-y-auto"
        showCloseButton={step !== "pushing"}
      >
        {renderContent()}
      </DialogContent>
    </Dialog>
  );
}
