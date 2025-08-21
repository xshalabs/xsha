import { useState, useEffect, useCallback } from "react";
import { Separator } from "@/components/ui/separator";
import type { TaskFormData } from "@/types/task";
import type { Project } from "@/types/project";
import type { DevEnvironment } from "@/types/dev-environment";
import { useDevEnvironments } from "@/hooks/useDevEnvironments";
import { useProjectBranches } from "@/hooks/useProjectBranches";
import { useTaskFormValidation } from "@/hooks/useTaskFormValidation";
import { useTaskFormFileHandling } from "@/hooks/useTaskFormFileHandling";
import { useInlineControls } from "@/hooks/useInlineControls";
import { TaskFormHeader } from "@/components/forms/task-form/TaskFormHeader";
import { TaskFormDevEnvironment } from "@/components/forms/task-form/TaskFormDevEnvironment";
import { TaskFormDescription } from "@/components/forms/task-form/TaskFormDescription";
import { TaskFormConfiguration } from "@/components/forms/task-form/TaskFormConfiguration";

interface TaskFormCreateNewProps {
  defaultProjectId?: number;
  currentProject?: Project;
  onSubmit: (data: TaskFormData, selectedEnvironment?: DevEnvironment) => Promise<void>;
  formId?: string;
}

export function TaskFormCreateNew({
  defaultProjectId,
  currentProject,
  onSubmit,
  formId = "new-task-create-form",
}: TaskFormCreateNewProps) {
  // Form state
  const [formData, setFormData] = useState<TaskFormData>({
    title: "",
    start_branch: "main",
    project_id: defaultProjectId || 0,
    dev_environment_id: undefined,
    requirement_desc: "",
    execution_time: undefined,
    include_branches: true,
    model: "default",
  });

  const [submitting, setSubmitting] = useState(false);

  // Custom hooks
  const { devEnvironments, loading: loadingDevEnvs, error: devEnvsError } = useDevEnvironments();
  
  const { errors, validateForm, clearFieldError } = useTaskFormValidation();
  
  const {
    isTimePickerOpen,
    isModelSelectorOpen,
    timePickerRef,
    modelSelectorRef,
    handleTimePickerToggle,
    handleModelSelectorToggle,
    closeTimePickerManual,
    closeModelSelectorManual,
  } = useInlineControls();

  const { 
    availableBranches, 
    fetching: fetchingBranches, 
    error: branchError, 
    fetchProjectBranches 
  } = useProjectBranches({
    currentProject,
  });

  // Define handleChange first before using it in callbacks
  const handleChange = useCallback((
    field: keyof TaskFormData,
    value: string | number | Date | undefined
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
    clearFieldError(field);
  }, [clearFieldError]);

  const requirementDescChangeCallback = useCallback((value: string) => {
    handleChange("requirement_desc", value);
  }, [handleChange]);

  const {
    attachments,
    uploadingAttachments,
    handlePaste,
    handleFileInputChange,
    handleAttachmentRemove,
    clearAttachments,
    getAttachmentIds,
  } = useTaskFormFileHandling({
    requirementDesc: formData.requirement_desc || "",
    onRequirementDescChange: requirementDescChangeCallback,
  });

  // Handle branch auto-selection separately to avoid circular dependency
  useEffect(() => {
    if (availableBranches.length > 0 && formData.start_branch) {
      if (!availableBranches.includes(formData.start_branch)) {
        const defaultBranch = availableBranches.includes("main")
          ? "main"
          : availableBranches.includes("master")
          ? "master"
          : availableBranches[0];
        setFormData(prev => ({ ...prev, start_branch: defaultBranch }));
      }
    }
  }, [availableBranches, formData.start_branch]);

  useEffect(() => {
    if (currentProject) {
      fetchProjectBranches();
    }
  }, [currentProject, fetchProjectBranches]);

  useEffect(() => {
    if (defaultProjectId && defaultProjectId !== formData.project_id) {
      setFormData((prev) => ({ ...prev, project_id: defaultProjectId }));
    }
  }, [defaultProjectId, formData.project_id]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm(formData)) {
      return;
    }

    setSubmitting(true);
    try {
      const selectedEnvironment = formData.dev_environment_id 
        ? devEnvironments.find(env => env.id === formData.dev_environment_id)
        : undefined;
      
      const submitData: TaskFormData = { 
        ...formData, 
        include_branches: true,
        attachment_ids: getAttachmentIds()
      };
      
      await onSubmit(submitData, selectedEnvironment);
      clearAttachments();
    } catch (error) {
      console.error("Failed to submit task:", error);
    } finally {
      setSubmitting(false);
    }
  };

  const handleTimeChange = useCallback((time: Date | undefined) => {
    handleChange("execution_time", time);
  }, [handleChange]);

  const handleModelChange = useCallback((newModel: string) => {
    handleChange("model", newModel);
    closeModelSelectorManual();
  }, [handleChange, closeModelSelectorManual]);

  const handleFileSelect = useCallback(() => {
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    fileInput?.click();
  }, []);


  const selectedEnvironment = formData.dev_environment_id 
    ? devEnvironments.find(env => env.id === formData.dev_environment_id)
    : undefined;

  return (
    <form id={formId} onSubmit={handleSubmit} className="my-4 space-y-6">
      {/* Basic Information */}
      <div className="space-y-6">
        <TaskFormHeader
          title={formData.title}
          onTitleChange={(title) => handleChange("title", title)}
          error={errors.title}
          disabled={submitting}
        />

        <TaskFormDevEnvironment
          devEnvironmentId={formData.dev_environment_id}
          onDevEnvironmentChange={(id) => handleChange("dev_environment_id", id)}
          devEnvironments={devEnvironments}
          loading={loadingDevEnvs}
          error={devEnvsError}
          validationError={errors.dev_environment_id}
          disabled={submitting}
        />

        <TaskFormDescription
          requirementDesc={formData.requirement_desc || ""}
          onRequirementDescChange={(value) => handleChange("requirement_desc", value)}
          onPaste={handlePaste}
          attachments={attachments}
          onAttachmentRemove={handleAttachmentRemove}
          uploadingAttachments={uploadingAttachments}
          onFileSelect={handleFileSelect}
          onFileInputChange={handleFileInputChange}
          executionTime={formData.execution_time}
          onExecutionTimeChange={handleTimeChange}
          model={formData.model || "default"}
          onModelChange={handleModelChange}
          selectedEnvironment={selectedEnvironment}
          error={errors.requirement_desc}
          disabled={submitting}
          isTimePickerOpen={isTimePickerOpen}
          isModelSelectorOpen={isModelSelectorOpen}
          timePickerRef={timePickerRef}
          modelSelectorRef={modelSelectorRef}
          onTimePickerToggle={handleTimePickerToggle}
          onModelSelectorToggle={handleModelSelectorToggle}
          onTimePickerClose={closeTimePickerManual}
          onModelSelectorClose={closeModelSelectorManual}
        />
      </div>

      <Separator />

      {/* Configuration */}
      <div className="space-y-6">
        <TaskFormConfiguration
          startBranch={formData.start_branch}
          onStartBranchChange={(branch) => handleChange("start_branch", branch)}
          availableBranches={availableBranches}
          fetchingBranches={fetchingBranches}
          branchError={branchError}
          onRefreshBranches={fetchProjectBranches}
          validationError={errors.start_branch}
          disabled={submitting}
        />
      </div>
    </form>
  );
}
