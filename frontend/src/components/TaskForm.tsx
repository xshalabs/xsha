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
import { Save, X } from "lucide-react";
import type { Task, TaskFormData } from "@/types/task";
import type { Project } from "@/types/project";
import type { DevEnvironment } from "@/types/dev-environment";
import { devEnvironmentsApi } from "@/lib/api/dev-environments";
import { projectsApi } from "@/lib/api/projects";

interface TaskFormProps {
  task?: Task;
  defaultProjectId?: number;
  currentProject?: Project;
  loading?: boolean;
  onSubmit: (data: TaskFormData | { title: string }) => Promise<void>;
  onCancel: () => void;
}

export function TaskForm({
  task,
  defaultProjectId,
  currentProject,
  loading = false,
  onSubmit,
  onCancel,
}: TaskFormProps) {
  const { t } = useTranslation();
  const isEdit = !!task;

  const [formData, setFormData] = useState<TaskFormData>({
    title: task?.title || "",
    start_branch: task?.start_branch || "main",
    project_id: task?.project_id || defaultProjectId || 0,
    dev_environment_id: task?.dev_environment_id || undefined,
    requirement_desc: "",
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [devEnvironments, setDevEnvironments] = useState<DevEnvironment[]>([]);
  const [loadingDevEnvs, setLoadingDevEnvs] = useState(false);
  const [availableBranches, setAvailableBranches] = useState<string[]>([]);
  const [loadingBranches, setLoadingBranches] = useState(false);
  const [branchError, setBranchError] = useState<string>("");

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
    if (!currentProject || isEdit) return;

    try {
      setLoadingBranches(true);
      setBranchError("");
      setAvailableBranches([]);

      const response = await projectsApi.fetchBranches({
        repo_url: currentProject.repo_url,
        credential_id: currentProject.credential_id || undefined,
      });

      if (response.result.can_access) {
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
      } else {
        setBranchError(
          response.result.error_message || t("tasks.errors.fetchBranchesFailed")
        );
      }
    } catch (error) {
      console.error("Failed to fetch branches:", error);
      setBranchError(t("tasks.errors.fetchBranchesFailed"));
    } finally {
      setLoadingBranches(false);
    }
  }, [currentProject, isEdit, t]);

  useEffect(() => {
    if (!isEdit && !task && defaultProjectId) {
      if (defaultProjectId !== formData.project_id) {
        setFormData((prev) => ({ ...prev, project_id: defaultProjectId }));
      }
    }
  }, [defaultProjectId, isEdit, task, formData.project_id]);

  useEffect(() => {
    if (currentProject && !isEdit) {
      fetchProjectBranches();
    }
  }, [currentProject, isEdit, fetchProjectBranches]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.title.trim()) {
      newErrors.title = t("tasks.validation.titleRequired");
    }

    if (!isEdit) {
      if (!formData.start_branch.trim()) {
        newErrors.start_branch = t("tasks.validation.branchRequired");
      }

      if (!formData.requirement_desc?.trim()) {
        newErrors.requirement_desc = t("tasks.validation.requirementDescRequired");
      }

      if (!formData.dev_environment_id) {
        newErrors.dev_environment_id = t("tasks.validation.devEnvironmentRequired");
      }
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
      const submitData = isEdit
        ? { title: formData.title }
        : { ...formData, include_branches: true };

      await onSubmit(submitData);
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
      <Card>
        <CardHeader>
          <CardTitle>
            {isEdit ? t("tasks.actions.edit") : t("tasks.actions.create")}
          </CardTitle>
          <CardDescription>
            {isEdit
              ? t("tasks.form.editDescription")
              : t("tasks.form.createDescription")}
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

            {!isEdit && (
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
            )}

            {!isEdit && (
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
            )}

            {!isEdit && (
              <div className="flex flex-col gap-3">
                <Label htmlFor="start_branch">
                  {t("tasks.fields.startBranch")}{" "}
                  <span className="text-red-500">*</span>
                </Label>
                {availableBranches.length > 0 ? (
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
                ) : loadingBranches ? (
                  <div className="flex items-center space-x-2 px-3 py-2 border border-gray-300 rounded-md bg-gray-50">
                    <span className="text-sm text-gray-500">
                      {t("common.loading")}...
                    </span>
                  </div>
                ) : (
                  <Input
                    id="start_branch"
                    type="text"
                    value={formData.start_branch}
                    onChange={(e) =>
                      handleChange("start_branch", e.target.value)
                    }
                    placeholder={t("tasks.form.branchPlaceholder")}
                    className={errors.start_branch ? "border-red-500" : ""}
                  />
                )}
                {errors.start_branch && (
                  <p className="text-sm text-red-500">{errors.start_branch}</p>
                )}
                {branchError && (
                  <p className="text-sm text-orange-500">{branchError}</p>
                )}
                <p className="text-sm text-gray-500">
                  {availableBranches.length > 0
                    ? t("tasks.form.branchFromRepository")
                    : t("tasks.form.branchHint")}
                </p>
              </div>
            )}

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
                  : isEdit
                  ? t("common.save")
                  : t("tasks.actions.create")}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
