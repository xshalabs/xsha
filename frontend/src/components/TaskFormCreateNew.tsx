import { useState, useEffect, useCallback } from "react";
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
  Save, 
  Loader2, 
  AlertCircle, 
  RefreshCw,
  Calendar,
  GitBranch,
  Settings,
  FileText,
  Zap
} from "lucide-react";
import type { TaskFormData } from "@/types/task";
import type { Project } from "@/types/project";
import type { DevEnvironment } from "@/types/dev-environment";
import { devEnvironmentsApi } from "@/lib/api/environments";
import { projectsApi } from "@/lib/api/projects";

interface TaskFormCreateNewProps {
  defaultProjectId?: number;
  currentProject?: Project;
  onSubmit: (data: TaskFormData) => Promise<void>;
  onCancel?: () => void;
  formId?: string;
}

export function TaskFormCreateNew({
  defaultProjectId,
  currentProject,
  onSubmit,
  onCancel,
  formId = "new-task-create-form",
}: TaskFormCreateNewProps) {
  const { t } = useTranslation();

  // Form state
  const [formData, setFormData] = useState<TaskFormData>({
    title: "",
    start_branch: "main",
    project_id: defaultProjectId || 0,
    dev_environment_id: undefined,
    requirement_desc: "",
    execution_time: undefined,
    include_branches: true,
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
      await onSubmit({ ...formData, include_branches: true });
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

          {/* Description */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <FileText className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="requirement_desc" className="text-sm font-medium">
                {t("tasks.fields.requirementDesc")} <span className="text-red-500">*</span>
              </Label>
            </div>
            <Textarea
              id="requirement_desc"
              value={formData.requirement_desc || ""}
              onChange={(e) => handleChange("requirement_desc", e.target.value)}
              placeholder={t("tasks.form.requirementDescPlaceholder")}
              rows={4}
              className={errors.requirement_desc ? "border-red-500 focus-visible:ring-red-500" : ""}
              disabled={submitting}
            />
            {errors.requirement_desc && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.requirement_desc}
              </p>
            )}
            <p className="text-xs text-muted-foreground">
              {t("tasks.form.requirementDescHint")}
            </p>
          </div>

          {/* Execution Time */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Calendar className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="execution_time" className="text-sm font-medium">
                {t("tasks.fields.executionTime")}
              </Label>
            </div>
            <DateTimePicker
              id="execution_time"
              value={formData.execution_time}
              onChange={(date) => handleChange("execution_time", date)}
              placeholder={t("tasks.form.executionTimePlaceholder")}
              label=""
              className={errors.execution_time ? "border-red-500" : ""}
              disabled={submitting}
            />
            {errors.execution_time && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.execution_time}
              </p>
            )}
            <p className="text-xs text-muted-foreground">
              {t("tasks.form.executionTimeHint")}
            </p>
          </div>
        </div>

      <Separator />

      {/* Configuration */}
      <div className="space-y-6">
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
