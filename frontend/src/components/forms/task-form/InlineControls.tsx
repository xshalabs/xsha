import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import {
  Paperclip,
  Calendar,
  Clock,
  Sparkles,
  X,
  FileText,
} from "lucide-react";
import type { Attachment } from "@/lib/api/attachments";
import type { DevEnvironment } from "@/types/dev-environment";

interface InlineControlsProps {
  attachments: Attachment[];
  uploadingAttachments: boolean;
  onFileSelect: () => void;
  executionTime?: Date;
  onExecutionTimeChange: (time: Date | undefined) => void;
  model: string;
  onModelChange: (model: string) => void;
  isPlanMode?: boolean;
  onPlanModeChange: (isPlanMode: boolean) => void;
  selectedEnvironment?: DevEnvironment;
  disabled?: boolean;
  isTimePickerOpen: boolean;
  isModelSelectorOpen: boolean;
  isPlanModeSelectorOpen: boolean;
  timePickerRef: React.RefObject<HTMLDivElement | null>;
  modelSelectorRef: React.RefObject<HTMLDivElement | null>;
  planModeSelectorRef: React.RefObject<HTMLDivElement | null>;
  onTimePickerToggle: () => void;
  onModelSelectorToggle: () => void;
  onPlanModeSelectorToggle: () => void;
  onTimePickerClose: () => void;
  onModelSelectorClose: () => void;
  onPlanModeSelectorClose: () => void;
}

export function InlineControls({
  attachments,
  uploadingAttachments,
  onFileSelect,
  executionTime,
  onExecutionTimeChange,
  model,
  onModelChange,
  isPlanMode,
  onPlanModeChange,
  selectedEnvironment,
  disabled,
  isTimePickerOpen,
  isModelSelectorOpen,
  isPlanModeSelectorOpen,
  timePickerRef,
  modelSelectorRef,
  planModeSelectorRef,
  onTimePickerToggle,
  onModelSelectorToggle,
  onPlanModeSelectorToggle,
  onTimePickerClose,
  onModelSelectorClose,
  onPlanModeSelectorClose,
}: InlineControlsProps) {
  const { t } = useTranslation();

  return (
    /* Interactive Controls positioned at the bottom left of the textarea */
    <div className="absolute bottom-3 left-3 right-3 flex items-end gap-3">
      {/* File attachment button */}
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={onFileSelect}
        className={`h-7 w-7 p-0 rounded-md transition-colors ${
          attachments.length > 0
            ? 'bg-green-100 text-green-600 hover:bg-green-200 dark:bg-green-900/50 dark:text-green-400'
            : 'text-muted-foreground hover:text-foreground hover:bg-muted'
        }`}
        title={attachments.length > 0 ? `${attachments.length} attachment(s)` : "Attach files"}
        disabled={uploadingAttachments || disabled}
      >
        <Paperclip className="h-3.5 w-3.5" />
      </Button>

      {/* Execution Time Control */}
      <div className="relative" ref={timePickerRef}>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onTimePickerToggle}
          className={`h-7 w-7 p-0 rounded-md transition-colors ${
            executionTime 
              ? 'bg-blue-100 text-blue-600 hover:bg-blue-200 dark:bg-blue-900/50 dark:text-blue-400' 
              : 'text-muted-foreground hover:text-foreground hover:bg-muted'
          }`}
          title={executionTime ? t("tasks.fields.executionTime") + ": " + executionTime.toLocaleString() : t("tasks.fields.executionTime")}
        >
          {executionTime ? <Calendar className="h-3.5 w-3.5" /> : <Clock className="h-3.5 w-3.5" />}
        </Button>
        
        {isTimePickerOpen && (
          <div className="absolute bottom-full left-0 mb-2 p-3 bg-background border rounded-lg shadow-lg z-10 min-w-[200px]">
            <div className="flex items-center justify-between mb-2">
              <Label className="text-xs font-medium">{t("tasks.fields.executionTime")}</Label>
              <Button
                variant="ghost"
                size="sm"
                onClick={onTimePickerClose}
                className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground"
              >
                <X className="h-3 w-3" />
              </Button>
            </div>
            <div className="space-y-2">
              <DateTimePicker
                value={executionTime}
                onChange={onExecutionTimeChange}
                placeholder={t("tasks.form.executionTimePlaceholder")}
                label=""
                className="h-8 text-xs"
              />
              <p className="text-xs text-muted-foreground">
                {t("tasks.form.executionTimeHint")}
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Model Selection - Only show when environment is selected */}
      {selectedEnvironment && (
        <div className="relative" ref={modelSelectorRef}>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onModelSelectorToggle}
            disabled={isPlanMode}
            className={`h-7 w-7 p-0 rounded-md transition-colors ${
              isPlanMode
                ? 'bg-orange-100 text-orange-600 dark:bg-orange-900/50 dark:text-orange-400 opacity-75 cursor-not-allowed'
                : model && model !== 'default'
                ? 'bg-purple-100 text-purple-600 hover:bg-purple-200 dark:bg-purple-900/50 dark:text-purple-400'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            }`}
            title={isPlanMode ? t("taskConversations.modelLockedInPlanMode") : (model ? t("tasks.fields.model") + ": " + model : t("tasks.fields.model"))}
          >
            <Sparkles className="h-3.5 w-3.5" />
          </Button>
          
          {isModelSelectorOpen && (
            <div className="absolute bottom-full left-0 mb-2 p-3 bg-background border rounded-lg shadow-lg z-10 min-w-[200px]">
              <div className="flex items-center justify-between mb-2">
                <Label className="text-xs font-medium">{t("tasks.fields.model")}</Label>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={onModelSelectorClose}
                  className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground"
                >
                  <X className="h-3 w-3" />
                </Button>
              </div>
              <div className="space-y-2">
                <Select
                  value={model || "default"}
                  onValueChange={onModelChange}
                  disabled={disabled}
                >
                  <SelectTrigger className="h-8 text-xs">
                    <SelectValue placeholder={t("tasks.form.selectModel")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="default">
                      <div className="flex flex-col items-start">
                        <span className="font-medium text-xs">{t("tasks.model.default")}</span>
                        <span className="text-xs text-muted-foreground">
                          {t("tasks.model.defaultDescription")}
                        </span>
                      </div>
                    </SelectItem>
                    {selectedEnvironment.type === "claude-code" && (
                      <>
                        <SelectItem value="sonnet">
                          <div className="flex flex-col items-start">
                            <span className="font-medium text-xs">{t("tasks.model.sonnet")}</span>
                            <span className="text-xs text-muted-foreground">
                              Sonnet
                            </span>
                          </div>
                        </SelectItem>
                        <SelectItem value="opus">
                          <div className="flex flex-col items-start">
                            <span className="font-medium text-xs">{t("tasks.model.opus")}</span>
                            <span className="text-xs text-muted-foreground">
                              Opus
                            </span>
                          </div>
                        </SelectItem>
                      </>
                    )}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  {t("tasks.form.modelHint")}
                </p>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Plan Mode Selection - Only show when environment is claude-code */}
      {selectedEnvironment && selectedEnvironment.type === "claude-code" && (
        <div className="relative" ref={planModeSelectorRef}>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onPlanModeSelectorToggle}
            className={`h-7 w-7 p-0 rounded-md transition-colors ${
              isPlanMode
                ? 'bg-orange-100 text-orange-600 hover:bg-orange-200 dark:bg-orange-900/50 dark:text-orange-400'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            }`}
            title={isPlanMode ? t("tasks.fields.planModeEnabled") : t("tasks.fields.planMode")}
          >
            <FileText className="h-3.5 w-3.5" />
          </Button>
          
          {isPlanModeSelectorOpen && (
            <div className="absolute bottom-full left-0 mb-2 p-3 bg-background border rounded-lg shadow-lg z-10 min-w-[200px]">
              <div className="flex items-center justify-between mb-2">
                <Label className="text-xs font-medium">{t("tasks.fields.planMode")}</Label>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={onPlanModeSelectorClose}
                  className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground"
                >
                  <X className="h-3 w-3" />
                </Button>
              </div>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="plan-mode-switch" className="text-xs">
                    {t("tasks.fields.enablePlanMode")}
                  </Label>
                  <Switch
                    id="plan-mode-switch"
                    checked={isPlanMode || false}
                    onCheckedChange={onPlanModeChange}
                    disabled={disabled}
                  />
                </div>
                <p className="text-xs text-muted-foreground">
                  {t("tasks.form.planModeHint")}
                </p>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}