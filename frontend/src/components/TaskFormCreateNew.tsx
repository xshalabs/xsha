import { useState, useEffect, useCallback, useRef } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Separator } from "@/components/ui/separator";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { 
  Loader2, 
  AlertCircle, 
  RefreshCw,
  Calendar,
  GitBranch,
  FileText,
  Zap,
  Clock,
  Sparkles,
  X,
  Paperclip
} from "lucide-react";
import type { TaskFormData } from "@/types/task";
import type { Project } from "@/types/project";
import type { DevEnvironment } from "@/types/dev-environment";
import { devEnvironmentsApi } from "@/lib/api/environments";
import { projectsApi } from "@/lib/api/projects";
import { AttachmentUploader } from "@/components/AttachmentUploader";
import { AttachmentSection } from "@/components/kanban/task-detail/AttachmentSection";
import { useAttachments } from "@/hooks/useAttachments";
import type { Attachment } from "@/lib/api/attachments";

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
  const { t } = useTranslation();

  // Attachment state
  const {
    attachments,
    uploading: uploadingAttachments,
    removeAttachment,
    clearAttachments,
    getAttachmentIds,
  } = useAttachments();

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

  // UI state
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  
  // Dev environments state
  const [devEnvironments, setDevEnvironments] = useState<DevEnvironment[]>([]);
  const [loadingDevEnvs, setLoadingDevEnvs] = useState(false);
  const [devEnvsError, setDevEnvsError] = useState<string>("");
  
  // Branch state
  const [availableBranches, setAvailableBranches] = useState<string[]>([]);
  const [fetchingBranches, setFetchingBranches] = useState(false);
  const [branchError, setBranchError] = useState<string>("");

  // UI state for inline controls
  const [isTimePickerOpen, setIsTimePickerOpen] = useState(false);
  const [isModelSelectorOpen, setIsModelSelectorOpen] = useState(false);
  const timePickerRef = useRef<HTMLDivElement>(null);
  const modelSelectorRef = useRef<HTMLDivElement>(null);

  // Load development environments
  useEffect(() => {
    const loadDevEnvironments = async () => {
      try {
        setLoadingDevEnvs(true);
        setDevEnvsError("");
        
        const allEnvironments: DevEnvironment[] = [];
        let currentPage = 1;
        let hasMorePages = true;
        
        while (hasMorePages) {
          const response = await devEnvironmentsApi.list({ 
            page: currentPage, 
            page_size: 100 
          });
          
          if (response.environments && response.environments.length > 0) {
            allEnvironments.push(...response.environments);
            hasMorePages = currentPage < response.total_pages;
            currentPage++;
          } else {
            hasMorePages = false;
          }
        }
        
        setDevEnvironments(allEnvironments);
      } catch (error) {
        console.error("Failed to load dev environments:", error);
        setDevEnvsError(t("tasks.errors.loadDevEnvironmentsFailed"));
      } finally {
        setLoadingDevEnvs(false);
      }
    };

    loadDevEnvironments();
  }, [t]);

  // Fetch project branches
  const fetchProjectBranches = useCallback(async () => {
    if (!currentProject) return;

    try {
      setFetchingBranches(true);
      setBranchError("");
      setAvailableBranches([]);

      const response = await projectsApi.fetchBranches({
        repo_url: currentProject.repo_url,
        credential_id: currentProject.credential_id || undefined,
      });

      if (response.result.can_access && response.result.branches && response.result.branches.length > 0) {
        setAvailableBranches(response.result.branches);
        
        // Auto-select default branch
        setFormData((prev) => {
          const currentBranch = prev.start_branch;
          if (!response.result.branches.includes(currentBranch)) {
            const defaultBranch = response.result.branches.includes("main")
              ? "main"
              : response.result.branches.includes("master")
              ? "master"
              : response.result.branches[0] || "main";
            return { ...prev, start_branch: defaultBranch };
          }
          return prev;
        });
      } else {
        const errorMsg = response.result.error_message || 
          (response.result.can_access ? t("tasks.errors.noBranchesFound") : t("tasks.errors.fetchBranchesFailed"));
        setBranchError(errorMsg);
      }
    } catch (error) {
      console.error("Failed to fetch branches:", error);
      setBranchError(t("tasks.errors.fetchBranchesFailed"));
    } finally {
      setFetchingBranches(false);
    }
  }, [currentProject, t]);

  // Load branches when project changes
  useEffect(() => {
    if (currentProject) {
      fetchProjectBranches();
    }
  }, [currentProject, fetchProjectBranches]);

  // Update project ID when defaultProjectId changes
  useEffect(() => {
    if (defaultProjectId && defaultProjectId !== formData.project_id) {
      setFormData((prev) => ({ ...prev, project_id: defaultProjectId }));
    }
  }, [defaultProjectId, formData.project_id]);

  // Form validation
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.title.trim()) {
      newErrors.title = t("tasks.validation.titleRequired");
    }

    if (!formData.start_branch.trim()) {
      newErrors.start_branch = t("tasks.validation.branchRequired");
    }

    if (!formData.requirement_desc?.trim()) {
      newErrors.requirement_desc = t("tasks.validation.requirementDescRequired");
    }

    if (!formData.dev_environment_id) {
      newErrors.dev_environment_id = t("tasks.validation.devEnvironmentRequired");
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setSubmitting(true);
    try {
      // Find the selected environment from loaded environments
      const selectedEnvironment = formData.dev_environment_id 
        ? devEnvironments.find(env => env.id === formData.dev_environment_id)
        : undefined;
      
      // Include attachment IDs in the form data
      const submitData: TaskFormData = { 
        ...formData, 
        include_branches: true,
        attachment_ids: getAttachmentIds()
      };
      
      await onSubmit(submitData, selectedEnvironment);
      
      // Clear attachments after successful submission
      clearAttachments();
    } catch (error) {
      console.error("Failed to submit task:", error);
    } finally {
      setSubmitting(false);
    }
  };

  // Handle field changes
  const handleChange = (
    field: keyof TaskFormData,
    value: string | number | Date | undefined
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));

    // Clear error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({
        ...prev,
        [field]: "",
      }));
    }
  };

  // Inline control handlers (similar to NewMessageForm)
  const handleTimePickerToggle = useCallback(() => {
    setIsTimePickerOpen(!isTimePickerOpen);
    setIsModelSelectorOpen(false); // Close model selector when opening time picker
  }, [isTimePickerOpen]);

  const handleModelSelectorToggle = useCallback(() => {
    setIsModelSelectorOpen(!isModelSelectorOpen);
    setIsTimePickerOpen(false); // Close time picker when opening model selector
  }, [isModelSelectorOpen]);

  const handleTimeChange = useCallback((time: Date | undefined) => {
    handleChange("execution_time", time);
    // Don't auto-close to allow multiple time adjustments
  }, []);

  const handleModelChange = useCallback((newModel: string) => {
    handleChange("model", newModel);
    setIsModelSelectorOpen(false); // Close after selection
  }, []);

  // Attachment handlers
  const handleAttachmentRemove = useCallback(async (attachment: Attachment) => {
    try {
      await removeAttachment(attachment);
    } catch (error) {
      console.error("Failed to remove attachment:", error);
    }
  }, [removeAttachment]);

  // Get selected environment
  const selectedEnvironment = formData.dev_environment_id 
    ? devEnvironments.find(env => env.id === formData.dev_environment_id)
    : undefined;

  // Handle click outside to close popups
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      
      // Only close if clicking completely outside our components and not on any portal/popup content
      const isClickOnPortal = target.closest('[data-radix-popper-content-wrapper], [data-radix-portal], [data-sonner-toaster]');
      const isClickOnTimePicker = timePickerRef.current?.contains(target as Node);
      const isClickOnModelSelector = modelSelectorRef.current?.contains(target as Node);
      
      if (!isClickOnPortal && !isClickOnTimePicker && !isClickOnModelSelector) {
        setIsTimePickerOpen(false);
        setIsModelSelectorOpen(false);
      }
    };

    // Use a timeout to avoid immediate closure
    const timeoutId = setTimeout(() => {
      document.addEventListener('mousedown', handleClickOutside);
    }, 100);

    return () => {
      clearTimeout(timeoutId);
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  return (
    <form id={formId} onSubmit={handleSubmit} className="my-4 space-y-6">
      {/* Basic Information */}
      <div className="space-y-6">
          {/* Title */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <FileText className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="title" className="text-sm font-medium">
                {t("tasks.fields.title")} <span className="text-red-500">*</span>
              </Label>
            </div>
            <Input
              id="title"
              type="text"
              value={formData.title}
              onChange={(e) => handleChange("title", e.target.value)}
              placeholder={t("tasks.form.titlePlaceholder")}
              className={errors.title ? "border-red-500 focus-visible:ring-red-500" : ""}
              disabled={submitting}
            />
            {errors.title && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.title}
              </p>
            )}
          </div>

          {/* Development Environment */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Zap className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="dev_environment" className="text-sm font-medium">
                {t("tasks.fields.devEnvironment")} <span className="text-red-500">*</span>
              </Label>
            </div>
            <Select
              value={formData.dev_environment_id?.toString() || ""}
              onValueChange={(value) =>
                handleChange(
                  "dev_environment_id",
                  value ? parseInt(value) : undefined
                )
              }
              disabled={loadingDevEnvs || submitting}
            >
              <SelectTrigger 
                className={errors.dev_environment_id ? "border-red-500 focus:ring-red-500" : ""}
              >
                <SelectValue
                  placeholder={
                    loadingDevEnvs 
                      ? t("common.loading") + "..."
                      : t("tasks.form.selectDevEnvironment")
                  }
                />
              </SelectTrigger>
              <SelectContent>
                {loadingDevEnvs ? (
                  <SelectItem value="loading" disabled>
                    <div className="flex items-center gap-2">
                      <Loader2 className="h-3 w-3 animate-spin" />
                      {t("common.loading")}...
                    </div>
                  </SelectItem>
                ) : devEnvironments.length === 0 ? (
                  <SelectItem value="empty" disabled>
                    {t("tasks.form.noDevEnvironmentsAvailable")}
                  </SelectItem>
                ) : (
                  devEnvironments.map((env) => (
                    <SelectItem key={env.id} value={env.id.toString()}>
                      <div className="flex items-center justify-between w-full">
                        <span className="font-medium">{env.name}</span>
                        <span className="text-xs text-muted-foreground ml-2">
                          {env.type}
                        </span>
                      </div>
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
            {errors.dev_environment_id && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.dev_environment_id}
              </p>
            )}
            {devEnvsError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{devEnvsError}</AlertDescription>
              </Alert>
            )}
            <p className="text-xs text-muted-foreground">
              {t("tasks.form.devEnvironmentHint")}
            </p>
          </div>

          {/* Description */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <FileText className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="requirement_desc" className="text-sm font-medium">
                {t("tasks.fields.requirementDesc")} <span className="text-red-500">*</span>
              </Label>
            </div>
            <div className="relative">
              <Textarea
                id="requirement_desc"
                value={formData.requirement_desc || ""}
                onChange={(e) => handleChange("requirement_desc", e.target.value)}
                placeholder={t("tasks.form.requirementDescPlaceholder")}
                rows={4}
                className={`min-h-[120px] resize-none pr-4 pb-16 ${errors.requirement_desc ? "border-red-500 focus-visible:ring-red-500" : ""}`}
                disabled={submitting}
              />
              
              {/* Interactive Controls positioned at the bottom left of the textarea */}
              <div className="absolute bottom-3 left-3 right-3 flex items-end gap-3">
                {/* Execution Time Control */}
                <div className="relative" ref={timePickerRef}>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={handleTimePickerToggle}
                    className={`h-7 w-7 p-0 rounded-md transition-colors ${
                      formData.execution_time 
                        ? 'bg-blue-100 text-blue-600 hover:bg-blue-200 dark:bg-blue-900/50 dark:text-blue-400' 
                        : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                    }`}
                    title={formData.execution_time ? t("tasks.fields.executionTime") + ": " + formData.execution_time.toLocaleString() : t("tasks.fields.executionTime")}
                  >
                    {formData.execution_time ? <Calendar className="h-3.5 w-3.5" /> : <Clock className="h-3.5 w-3.5" />}
                  </Button>
                  
                  {isTimePickerOpen && (
                    <div className="absolute bottom-full left-0 mb-2 p-3 bg-background border rounded-lg shadow-lg z-10 min-w-[200px]">
                      <div className="flex items-center justify-between mb-2">
                        <Label className="text-xs font-medium">{t("tasks.fields.executionTime")}</Label>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setIsTimePickerOpen(false)}
                          className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground"
                        >
                          <X className="h-3 w-3" />
                        </Button>
                      </div>
                      <div className="space-y-2">
                        <DateTimePicker
                          value={formData.execution_time}
                          onChange={handleTimeChange}
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
                      onClick={handleModelSelectorToggle}
                      className={`h-7 w-7 p-0 rounded-md transition-colors ${
                        formData.model && formData.model !== 'default'
                          ? 'bg-purple-100 text-purple-600 hover:bg-purple-200 dark:bg-purple-900/50 dark:text-purple-400'
                          : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                      }`}
                      title={formData.model ? t("tasks.fields.model") + ": " + formData.model : t("tasks.fields.model")}
                    >
                      {formData.model && formData.model !== 'default' ? <Sparkles className="h-3.5 w-3.5" /> : <Zap className="h-3.5 w-3.5" />}
                    </Button>
                    
                    {isModelSelectorOpen && (
                      <div className="absolute bottom-full left-0 mb-2 p-3 bg-background border rounded-lg shadow-lg z-10 min-w-[200px]">
                        <div className="flex items-center justify-between mb-2">
                          <Label className="text-xs font-medium">{t("tasks.fields.model")}</Label>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setIsModelSelectorOpen(false)}
                            className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground"
                          >
                            <X className="h-3 w-3" />
                          </Button>
                        </div>
                        <div className="space-y-2">
                          <Select
                            value={formData.model || "default"}
                            onValueChange={handleModelChange}
                            disabled={submitting}
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
              </div>
            </div>
            
            {errors.requirement_desc && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.requirement_desc}
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

          {/* Attachments Section */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Paperclip className="h-4 w-4 text-muted-foreground" />
              <Label className="text-sm font-medium">
                {t("tasks.fields.attachments", "Attachments")}
              </Label>
            </div>
            
            {/* Display existing attachments if any */}
            {attachments.length > 0 && (
              <AttachmentSection
                attachments={attachments}
                onRemove={handleAttachmentRemove}
              />
            )}
            
            {/* Attachment uploader */}
            <AttachmentUploader
              existingAttachments={attachments}
              onUploadSuccess={() => {
                // Attachment is automatically added to the hook state
              }}
              onUploadError={(error) => {
                console.error("Upload error:", error);
              }}
              disabled={submitting || uploadingAttachments}
            />
            
            <p className="text-xs text-muted-foreground">
              {t("tasks.form.attachmentsHint", "Upload images or PDF files to provide additional context for your task (max 10MB per file)")}
            </p>
          </div>
        </div>

      <Separator />

      {/* Configuration */}
      <div className="space-y-6">
          {/* Start Branch */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <GitBranch className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="start_branch" className="text-sm font-medium">
                {t("tasks.fields.startBranch")} <span className="text-red-500">*</span>
              </Label>
              {fetchingBranches && (
                <Loader2 className="h-3 w-3 animate-spin text-blue-500" />
              )}
            </div>
            <Select
              value={formData.start_branch}
              onValueChange={(value) => handleChange("start_branch", value)}
              disabled={fetchingBranches || submitting}
            >
              <SelectTrigger
                className={errors.start_branch ? "border-red-500 focus:ring-red-500" : ""}
              >
                <SelectValue 
                  placeholder={
                    fetchingBranches 
                      ? t("tasks.form.fetchingBranches") + "..."
                      : t("tasks.form.selectBranch")
                  } 
                />
              </SelectTrigger>
              <SelectContent>
                {fetchingBranches ? (
                  <SelectItem value="loading" disabled>
                    <div className="flex items-center gap-2">
                      <Loader2 className="h-3 w-3 animate-spin" />
                      {t("tasks.form.fetchingBranches")}...
                    </div>
                  </SelectItem>
                ) : availableBranches.length === 0 ? (
                  <SelectItem value="empty" disabled>
                    {branchError || t("tasks.form.noBranchesAvailable")}
                  </SelectItem>
                ) : (
                  availableBranches.map((branch) => (
                    <SelectItem key={branch} value={branch}>
                      <div className="flex items-center gap-2">
                        <GitBranch className="h-3 w-3" />
                        {branch}
                      </div>
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
            {errors.start_branch && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.start_branch}
              </p>
            )}
            {branchError && !fetchingBranches && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription className="flex items-center justify-between">
                  <span>{branchError}</span>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={fetchProjectBranches}
                    className="h-6 px-2"
                  >
                    <RefreshCw className="h-3 w-3" />
                  </Button>
                </AlertDescription>
              </Alert>
            )}
            <p className="text-xs text-muted-foreground">
              {t("tasks.form.branchFromRepository")}
            </p>
          </div>
        </div>
    </form>
  );
}
