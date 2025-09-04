import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type {
  Project,
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectFormData,
  GitProtocolType,
} from "@/types/project";

interface CredentialOption {
  id: number;
  name: string;
  type: string;
  username: string;
}

interface UseProjectFormOptions {
  project?: Project;
  onSubmit: (project: Project) => Promise<void>;
}

export function useProjectForm({ project, onSubmit }: UseProjectFormOptions) {
  const { t } = useTranslation();
  const isEdit = !!project;

  const [formData, setFormData] = useState<ProjectFormData>({
    name: project?.name || "",
    description: project?.description || "",
    system_prompt: project?.system_prompt || "",
    repo_url: project?.repo_url || "",
    protocol: project?.protocol || "https",
    credential_id: project?.credential_id,
  });

  const [credentials, setCredentials] = useState<CredentialOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [credentialsLoading, setCredentialsLoading] = useState(false);
  const [credentialValidating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [accessValidated, setAccessValidated] = useState(false);
  const [accessError, setAccessError] = useState<string | null>(null);

  const loadCredentials = useCallback(async (protocol: GitProtocolType) => {
    try {
      setCredentialsLoading(true);
      const response = await apiService.projects.getCompatibleCredentials(protocol);
      setCredentials(response.credentials);
    } catch (error) {
      logError(error as Error, "Failed to load credentials");
      setCredentials([]);
    } finally {
      setCredentialsLoading(false);
    }
  }, []);



  const validateForm = useCallback((): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = t("projects.validation.nameRequired");
    }

    if (!formData.repo_url.trim()) {
      newErrors.repo_url = t("projects.validation.repoUrlRequired");
    } else {
      const urlPattern = /^(https?:\/\/|git@)/;
      if (!urlPattern.test(formData.repo_url)) {
        newErrors.repo_url = t("projects.validation.invalidRepoUrl");
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [formData.name, formData.repo_url, t]);

  const handleInputChange = useCallback((
    field: keyof ProjectFormData,
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


    if (field === "credential_id") {
      setAccessValidated(false);
      setAccessError(null);
    }
  }, [errors]);

  const handleSubmit = useCallback(async () => {
    if (!validateForm()) {
      return;
    }

    try {
      setLoading(true);
      setError(null);

      let result: Project;

      if (isEdit && project) {
        const updateData: UpdateProjectRequest = {
          name: formData.name,
          description: formData.description,
          system_prompt: formData.system_prompt,
          repo_url: formData.repo_url,
          protocol: formData.protocol,
          credential_id: formData.credential_id,
        };

        await apiService.projects.update(project.id, updateData);
        const response = await apiService.projects.get(project.id);
        result = response.project;
      } else {
        const createData: CreateProjectRequest = {
          name: formData.name,
          description: formData.description,
          system_prompt: formData.system_prompt,
          repo_url: formData.repo_url,
          credential_id: formData.credential_id,
        };

        const response = await apiService.projects.create(createData);
        result = response.project;
      }

      await onSubmit(result);
    } catch (error) {
      const errorMessage =
        error instanceof Error
          ? error.message
          : isEdit
          ? t("projects.messages.updateFailed")
          : t("projects.messages.createFailed");
      setError(errorMessage);
      logError(
        error as Error,
        `Failed to ${isEdit ? "update" : "create"} project`
      );
      throw error;
    } finally {
      setLoading(false);
    }
  }, [
    validateForm,
    formData,
    isEdit,
    project,
    onSubmit,
    t
  ]);

  // Effects
  useEffect(() => {
    if (formData.protocol) {
      loadCredentials(formData.protocol);
    }
  }, [formData.protocol, loadCredentials]);



  return {
    // Form data
    formData,
    isEdit,

    // Loading states
    loading,
    credentialsLoading,
    credentialValidating,

    // Error states
    error,
    errors,
    accessError,

    // Validation states
    accessValidated,

    // Data
    credentials,

    // Actions
    handleInputChange,
    handleSubmit,
  };
}
