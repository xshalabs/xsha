import { useRef } from "react";
import { useTranslation } from "react-i18next";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { AlertCircle, FileText } from "lucide-react";
import { AttachmentSection } from "@/components/kanban/task-detail/AttachmentSection";
import { InlineControls } from "./InlineControls";
import type { Attachment } from "@/lib/api/attachments";
import type { DevEnvironment } from "@/types/dev-environment";

interface TaskFormDescriptionProps {
  requirementDesc: string;
  onRequirementDescChange: (value: string) => void;
  onPaste: (e: React.ClipboardEvent<HTMLTextAreaElement>) => void;
  attachments: Attachment[];
  onAttachmentRemove: (attachment: Attachment) => void;
  uploadingAttachments: boolean;
  onFileSelect: () => void;
  onFileInputChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  executionTime?: Date;
  onExecutionTimeChange: (time: Date | undefined) => void;
  model: string;
  onModelChange: (model: string) => void;
  isPlanMode?: boolean;
  onPlanModeChange: (isPlanMode: boolean) => void;
  selectedEnvironment?: DevEnvironment;
  error?: string;
  disabled?: boolean;
  // Inline controls state
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

export function TaskFormDescription({
  requirementDesc,
  onRequirementDescChange,
  onPaste,
  attachments,
  onAttachmentRemove,
  uploadingAttachments,
  onFileSelect,
  onFileInputChange,
  executionTime,
  onExecutionTimeChange,
  model,
  onModelChange,
  isPlanMode,
  onPlanModeChange,
  selectedEnvironment,
  error,
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
}: TaskFormDescriptionProps) {
  const { t } = useTranslation();
  const fileInputRef = useRef<HTMLInputElement>(null);

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <FileText className="h-4 w-4 text-muted-foreground" />
        <Label htmlFor="requirement_desc" className="text-sm font-medium">
          {t("tasks.fields.requirementDesc")} <span className="text-red-500">*</span>
        </Label>
      </div>
      
      {/* Uploaded Attachments Display - Above the textarea */}
      {attachments.length > 0 && (
        <AttachmentSection
          attachments={attachments}
          onRemove={onAttachmentRemove}
        />
      )}
      
      <div className="relative">
        <Textarea
          id="requirement_desc"
          value={requirementDesc}
          onChange={(e) => onRequirementDescChange(e.target.value)}
          onPaste={onPaste}
          placeholder={t("tasks.form.requirementDescPlaceholder")}
          rows={4}
          className={`min-h-[120px] resize-none pr-4 pb-16 ${error ? "border-red-500 focus-visible:ring-red-500" : ""}`}
          disabled={disabled}
          aria-describedby="requirement-desc-shortcut-hint"
        />
        
        <InlineControls
          attachments={attachments}
          uploadingAttachments={uploadingAttachments}
          onFileSelect={onFileSelect}
          executionTime={executionTime}
          onExecutionTimeChange={onExecutionTimeChange}
          model={model}
          onModelChange={onModelChange}
          isPlanMode={isPlanMode}
          onPlanModeChange={onPlanModeChange}
          selectedEnvironment={selectedEnvironment}
          disabled={disabled}
          isTimePickerOpen={isTimePickerOpen}
          isModelSelectorOpen={isModelSelectorOpen}
          isPlanModeSelectorOpen={isPlanModeSelectorOpen}
          timePickerRef={timePickerRef}
          modelSelectorRef={modelSelectorRef}
          planModeSelectorRef={planModeSelectorRef}
          onTimePickerToggle={onTimePickerToggle}
          onModelSelectorToggle={onModelSelectorToggle}
          onPlanModeSelectorToggle={onPlanModeSelectorToggle}
          onTimePickerClose={onTimePickerClose}
          onModelSelectorClose={onModelSelectorClose}
          onPlanModeSelectorClose={onPlanModeSelectorClose}
        />
        
        {/* Hidden file input */}
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*,.pdf"
          multiple
          onChange={onFileInputChange}
          className="hidden"
        />
      </div>
      
      {error && (
        <p className="text-sm text-red-500 flex items-center gap-1">
          <AlertCircle className="h-3 w-3" />
          {error}
        </p>
      )}
      
      <p className="text-xs text-muted-foreground">
        {t("tasks.form.requirementDescHint")}
      </p>
      
      {/* Hint for interactive controls */}
      <div className="text-xs text-muted-foreground">
        {t("tasks.form.clickIconsToConfigureHint", "Click icons in the text area to configure execution settings")}
      </div>
    </div>
  );
}