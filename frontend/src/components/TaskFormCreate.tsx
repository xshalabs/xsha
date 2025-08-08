import { useState, useEffect, useCallback } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
import { useTranslation } from "react-i18next";
import { Save, X, Loader2 } from "lucide-react";
import type { TaskFormData } from "@/types/task";
import type { Project } from "@/types/project";
import type { DevEnvironment } from "@/types/dev-environment";
import { devEnvironmentsApi } from "@/lib/api/environments";
import { projectsApi } from "@/lib/api/projects";

interface TaskFormCreateProps {
  defaultProjectId?: number;
  currentProject?: Project;
  loading?: boolean;
  onSubmit: (data: TaskFormData) => Promise<void>;
  onCancel: () => void;
}

export function TaskFormCreate({
  defaultProjectId,
  currentProject,
  loading = false,
  onSubmit,
  onCancel,
}: TaskFormCreateProps) {
  const { t } = useTranslation();

  const [formData, setFormData] = useState<TaskFormData>({
    title: "",
    start_branch: "main",
    project_id: defaultProjectId || 0,
    dev_environment_id: undefined,
    requirement_desc: "",
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [devEnvironments, setDevEnvironments] = useState<DevEnvironment[]>([]);
  const [loadingDevEnvs, setLoadingDevEnvs] = useState(false);
  const [availableBranches, setAvailableBranches] = useState<string[]>([]);
  const [branchError, setBranchError] = useState<string>("");
  const [fetchingBranches, setFetchingBranches] = useState(false);
  const [branchFetchError, setBranchFetchError] = useState<string>("");

  useEffect(() => {
    const loadDevEnvironments = async () => {
      try {
        setLoadingDevEnvs(true);
        const response = await devEnvironmentsApi.list();
        setDevEnvironments(response.environments || []);
      } catch (error) {
        console.error("Failed to load dev environments:", error);
        setDevEnvironments([]);
      } finally {
        setLoadingDevEnvs(false);
      }
    };

    loadDevEnvironments();
  }, []);

  const fetchProjectBranches = useCallback(async () => {
    if (!currentProject) return;

    try {
      setFetchingBranches(true);
      setBranchError("");
      setBranchFetchError("");
      setAvailableBranches([]);

      const response = await projectsApi.fetchBranches({
        repo_url: currentProject.repo_url,
        credential_id: currentProject.credential_id || undefined,
      });

      if (response.result.can_access && response.result.branches && response.result.branches.length > 0) {
        setAvailableBranches(response.result.branches);
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
        setFetchingBranches(false);
      } else {
        const errorMsg = response.result.error_message || 
          (response.result.can_access ? t("tasks.errors.noBranchesFound") : t("tasks.errors.fetchBranchesFailed"));
        setBranchError(errorMsg);
        setBranchFetchError(errorMsg);
        setFetchingBranches(false);
      }
    } catch (error) {
      console.error("Failed to fetch branches:", error);
      const errorMsg = t("tasks.errors.fetchBranchesFailed");
      setBranchError(errorMsg);
      setBranchFetchError(errorMsg);
      setFetchingBranches(false);
    }
  }, [currentProject, t]);

  useEffect(() => {
    if (defaultProjectId && defaultProjectId !== formData.project_id) {
      setFormData((prev) => ({ ...prev, project_id: defaultProjectId }));
    }
  }, [defaultProjectId, formData.project_id]);

  useEffect(() => {
    if (currentProject) {
      fetchProjectBranches();
    }
  }, [currentProject, fetchProjectBranches]);

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

  const handleChange = (
    field: keyof TaskFormData,
    value: string | number | undefined
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));

    if (errors[field]) {
      setErrors((prev) => ({
        ...prev,
        [field]: "",
      }));
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <Card className="relative">
        {fetchingBranches && (
          <div className="absolute inset-0 bg-white/10 backdrop-blur-sm z-10 rounded-lg flex items-center justify-center">
            <div className="flex flex-col items-center space-y-3">
              <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
              <p className="text-sm text-foreground">{t("tasks.form.fetchingBranches")}</p>
            </div>
          </div>
        )}
        
        {branchFetchError && !fetchingBranches && (
          <div className="absolute inset-0 bg-white/10 backdrop-blur-sm z-10 rounded-lg flex items-center justify-center">
            <div className="flex flex-col items-center space-y-4 max-w-md mx-auto p-6 text-center">
              <div className="w-12 h-12 rounded-full bg-red-100 flex items-center justify-center">
                <X className="h-6 w-6 text-red-600" />
              </div>
              <div>
                <h3 className="text-lg font-medium text-foreground mb-2">
                  {t("tasks.errors.fetchBranchesFailedTitle")}
                </h3>
                <p className="text-sm text-red-600 mb-4">{branchFetchError}</p>
                <Button 
                  onClick={() => {
                    setBranchFetchError("");
                    fetchProjectBranches();
                  }}
                  size="sm"
                >
                  {t("common.retry")}
                </Button>
              </div>
            </div>
          </div>
        )}

        <CardHeader>
          <CardTitle>{t("tasks.actions.create")}</CardTitle>
          <CardDescription>
            {t("tasks.form.createDescription")}
          </CardDescription>
        </CardHeader>

        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="flex flex-col gap-3">
              <Label htmlFor="title">
                {t("tasks.fields.title")}{" "}
                <span className="text-red-500">*</span>
              </Label>
              <Input
                id="title"
                type="text"
                value={formData.title}
                onChange={(e) => handleChange("title", e.target.value)}
                placeholder={t("tasks.form.titlePlaceholder")}
                className={errors.title ? "border-red-500" : ""}
              />
              {errors.title && (
                <p className="text-sm text-red-500">{errors.title}</p>
              )}
            </div>

            <div className="flex flex-col gap-3">
              <Label htmlFor="requirement_desc">
                {t("tasks.fields.requirementDesc")}{" "}
                <span className="text-red-500">*</span>
              </Label>
              <Textarea
                id="requirement_desc"
                value={formData.requirement_desc || ""}
                onChange={(e) =>
                  handleChange("requirement_desc", e.target.value)
                }
                placeholder={t("tasks.form.requirementDescPlaceholder")}
                rows={4}
                className={errors.requirement_desc ? "border-red-500" : ""}
              />
              {errors.requirement_desc && (
                <p className="text-sm text-red-500">{errors.requirement_desc}</p>
              )}
              <p className="text-sm text-gray-500">
                {t("tasks.form.requirementDescHint")}
              </p>
            </div>

            <div className="flex flex-col gap-3">
              <Label htmlFor="dev_environment">
                {t("tasks.fields.devEnvironment")}{" "}
                <span className="text-red-500">*</span>
              </Label>
              <Select
                value={formData.dev_environment_id?.toString() || ""}
                onValueChange={(value) =>
                  handleChange(
                    "dev_environment_id",
                    value ? parseInt(value) : undefined
                  )
                }
              >
                <SelectTrigger className={errors.dev_environment_id ? "border-red-500" : ""}>
                  <SelectValue
                    placeholder={t("tasks.form.selectDevEnvironment")}
                  />
                </SelectTrigger>
                <SelectContent>
                  {loadingDevEnvs ? (
                    <SelectItem value="loading" disabled>
                      {t("common.loading")}...
                    </SelectItem>
                  ) : (
                    devEnvironments.map((env) => (
                      <SelectItem key={env.id} value={env.id.toString()}>
                        <div className="flex items-center space-x-2">
                          <span>{env.name}</span>
                          <span className="text-xs text-gray-500">
                            ({env.type})
                          </span>
                        </div>
                      </SelectItem>
                    ))
                  )}
                </SelectContent>
              </Select>
              {errors.dev_environment_id && (
                <p className="text-sm text-red-500">{errors.dev_environment_id}</p>
              )}
              <p className="text-sm text-gray-500">
                {t("tasks.form.devEnvironmentHint")}
              </p>
            </div>

            <div className="flex flex-col gap-3">
              <Label htmlFor="start_branch">
                {t("tasks.fields.startBranch")}{" "}
                <span className="text-red-500">*</span>
              </Label>
              <Select
                value={formData.start_branch}
                onValueChange={(value) =>
                  handleChange("start_branch", value)
                }
              >
                <SelectTrigger
                  className={errors.start_branch ? "border-red-500" : ""}
                >
                  <SelectValue placeholder={t("tasks.form.selectBranch")} />
                </SelectTrigger>
                <SelectContent>
                  {availableBranches.map((branch) => (
                    <SelectItem key={branch} value={branch}>
                      {branch}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.start_branch && (
                <p className="text-sm text-red-500">{errors.start_branch}</p>
              )}
              {branchError && (
                <p className="text-sm text-orange-500">{branchError}</p>
              )}
              <p className="text-sm text-gray-500">
                {t("tasks.form.branchFromRepository")}
              </p>
            </div>

            <div className="flex items-center justify-end space-x-4 pt-6">
              <Button
                type="button"
                variant="outline"
                onClick={onCancel}
                disabled={submitting}
              >
                <X className="w-4 h-4 mr-2" />
                {t("common.cancel")}
              </Button>
              <Button type="submit" disabled={submitting || loading}>
                <Save className="w-4 h-4 mr-2" />
                {submitting
                  ? t("common.saving")
                  : t("tasks.actions.create")}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}