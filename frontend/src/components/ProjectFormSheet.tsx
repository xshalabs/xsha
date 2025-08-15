import { useState, useEffect, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { 
  Save, 
  Loader2, 
  AlertCircle, 
  RefreshCw,
  FileText,
  GitBranch,
  Settings,
  Link2
} from "lucide-react";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import { useTranslation } from "react-i18next";
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

interface ProjectFormSheetProps {
  project?: Project;
  onSubmit: (project: Project) => Promise<void>;
  onCancel?: () => void;
  formId?: string;
}

export function ProjectFormSheet({
  project,
  onSubmit,
  onCancel,
  formId = "project-form-sheet",
}: ProjectFormSheetProps) {
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
  const [urlParsing, setUrlParsing] = useState(false);
  const [credentialValidating, setCredentialValidating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [urlParseTimeout, setUrlParseTimeout] = useState<NodeJS.Timeout | null>(null);
  const [accessValidated, setAccessValidated] = useState(false);
  const [accessError, setAccessError] = useState<string | null>(null);

  const loadCredentials = async (protocol: GitProtocolType) => {
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
  };

  const validateAccessAndFetchBranches = async (
    repoUrl: string,
    credentialId?: number
  ) => {
    if (!repoUrl.trim()) {
      setAccessValidated(false);
      setAccessError(null);
      return;
    }

    try {
      setCredentialValidating(true);
      setAccessError(null);
      setAccessValidated(false);

      const validateResponse = await apiService.projects.validateAccess({
        repo_url: repoUrl,
        credential_id: credentialId,
      });

      if (!validateResponse.can_access) {
        setAccessError(
          validateResponse.error || t("projects.messages.accessFailed")
        );
        return;
      }

      setAccessValidated(true);
    } catch (error) {
      const errorMessage =
        error instanceof Error
          ? error.message
          : t("projects.messages.validateAccessFailed");
      setAccessError(errorMessage);
      logError(error as Error, "Failed to validate repository access");
    } finally {
      setCredentialValidating(false);
    }
  };

  const parseRepositoryUrl = useCallback(
    async (url: string) => {
      if (!url.trim()) {
        return;
      }

      const gitUrlPattern = /^(https?:\/\/|git@|ssh:\/\/)/;
      if (!gitUrlPattern.test(url)) {
        return;
      }

      try {
        setUrlParsing(true);
        const response = await apiService.projects.parseUrl(url);

        if (response.result.is_valid) {
          const detectedProtocol = response.result.protocol as GitProtocolType;

          if (detectedProtocol !== formData.protocol) {
            setFormData((prev) => ({
              ...prev,
              protocol: detectedProtocol,
              credential_id: undefined,
            }));
          }
        }
      } catch (error) {
        logError(error as Error, "Failed to parse repository URL");
      } finally {
        setUrlParsing(false);
      }
    },
    [formData.protocol]
  );

  useEffect(() => {
    if (formData.protocol) {
      loadCredentials(formData.protocol);
    }
  }, [formData.protocol]);

  useEffect(() => {
    if (formData.repo_url && formData.credential_id) {
      validateAccessAndFetchBranches(formData.repo_url, formData.credential_id);
    } else if (formData.repo_url && !formData.credential_id) {
      validateAccessAndFetchBranches(formData.repo_url);
    } else {
      setAccessValidated(false);
      setAccessError(null);
    }
  }, [formData.repo_url, formData.credential_id]);

  useEffect(() => {
    return () => {
      if (urlParseTimeout) {
        clearTimeout(urlParseTimeout);
      }
    };
  }, [urlParseTimeout]);

  const validateForm = (): boolean => {
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
  };

  const handleInputChange = (
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

    if (field === "repo_url" && typeof value === "string") {
      if (urlParseTimeout) {
        clearTimeout(urlParseTimeout);
      }

      const timeoutId = setTimeout(() => {
        parseRepositoryUrl(value);
      }, 500);

      setUrlParseTimeout(timeoutId);
    }

    if (field === "credential_id") {
      setAccessValidated(false);
      setAccessError(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

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
          protocol: formData.protocol,
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
      throw error; // Re-throw to let parent component handle it
    } finally {
      setLoading(false);
    }
  };

  return (
    <form id={formId} onSubmit={handleSubmit} className="my-4 space-y-6">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Basic Information */}
      <div className="space-y-6">
        {/* Project Name */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <FileText className="h-4 w-4 text-muted-foreground" />
            <Label htmlFor="name" className="text-sm font-medium">
              {t("projects.name")} <span className="text-red-500">*</span>
            </Label>
          </div>
          <Input
            id="name"
            type="text"
            value={formData.name}
            onChange={(e) => handleInputChange("name", e.target.value)}
            placeholder={t("projects.placeholders.name")}
            className={errors.name ? "border-red-500 focus-visible:ring-red-500" : ""}
            disabled={loading}
          />
          {errors.name && (
            <p className="text-sm text-red-500 flex items-center gap-1">
              <AlertCircle className="h-3 w-3" />
              {errors.name}
            </p>
          )}
        </div>

        {/* Description */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <FileText className="h-4 w-4 text-muted-foreground" />
            <Label htmlFor="description" className="text-sm font-medium">
              {t("projects.description")}
            </Label>
          </div>
          <Input
            id="description"
            type="text"
            value={formData.description}
            onChange={(e) => handleInputChange("description", e.target.value)}
            placeholder={t("projects.placeholders.description")}
            disabled={loading}
          />
        </div>

        {/* System Prompt */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Settings className="h-4 w-4 text-muted-foreground" />
            <Label htmlFor="system_prompt" className="text-sm font-medium">
              {t("projects.systemPrompt")}
            </Label>
          </div>
          <Textarea
            id="system_prompt"
            value={formData.system_prompt}
            onChange={(e) => handleInputChange("system_prompt", e.target.value)}
            placeholder={t("projects.placeholders.systemPrompt")}
            rows={3}
            disabled={loading}
          />
          <p className="text-xs text-muted-foreground">
            {t("projects.systemPromptHelp")}
          </p>
        </div>

        {/* Repository URL */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Link2 className="h-4 w-4 text-muted-foreground" />
            <Label htmlFor="repo_url" className="text-sm font-medium">
              {t("projects.repoUrl")} <span className="text-red-500">*</span>
            </Label>
            {urlParsing && (
              <Loader2 className="h-3 w-3 animate-spin text-blue-500" />
            )}
          </div>
          <Input
            id="repo_url"
            type="text"
            value={formData.repo_url}
            onChange={(e) => handleInputChange("repo_url", e.target.value)}
            placeholder={t("projects.placeholders.repoUrl")}
            className={errors.repo_url ? "border-red-500 focus-visible:ring-red-500" : ""}
            disabled={loading}
          />
          {errors.repo_url && (
            <p className="text-sm text-red-500 flex items-center gap-1">
              <AlertCircle className="h-3 w-3" />
              {errors.repo_url}
            </p>
          )}
          {!errors.repo_url && formData.repo_url && (
            <p className="text-xs text-muted-foreground">
              {t("projects.protocolAutoDetected")}: {formData.protocol.toUpperCase()}
            </p>
          )}
        </div>

        {/* Git Credential */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <GitBranch className="h-4 w-4 text-muted-foreground" />
            <Label htmlFor="credential_id" className="text-sm font-medium">
              {t("projects.credential")}
            </Label>
            {credentialsLoading && (
              <Loader2 className="h-3 w-3 animate-spin text-blue-500" />
            )}
          </div>
          <Select
            onValueChange={(value) =>
              handleInputChange(
                "credential_id",
                value ? Number(value) : undefined
              )
            }
            value={formData.credential_id?.toString()}
            disabled={credentialsLoading || loading}
          >
            <SelectTrigger>
              <SelectValue
                placeholder={
                  credentialsLoading 
                    ? t("common.loading") + "..."
                    : t("projects.placeholders.selectCredential")
                }
              />
            </SelectTrigger>
            <SelectContent>
              {credentialsLoading ? (
                <SelectItem value="loading" disabled>
                  <div className="flex items-center gap-2">
                    <Loader2 className="h-3 w-3 animate-spin" />
                    {t("common.loading")}...
                  </div>
                </SelectItem>
              ) : credentials.length === 0 ? (
                <SelectItem value="empty" disabled>
                  {t("projects.noCredentialsAvailable")}
                </SelectItem>
              ) : (
                credentials.map((credential) => (
                  <SelectItem key={credential.id} value={credential.id.toString()}>
                    <div className="flex items-center justify-between w-full">
                      <span className="font-medium">{credential.name}</span>
                      <span className="text-xs text-muted-foreground ml-2">
                        {credential.type} - {credential.username}
                      </span>
                    </div>
                  </SelectItem>
                ))
              )}
            </SelectContent>
          </Select>
          <p className="text-xs text-muted-foreground">
            {t("projects.credentialHelp")}
          </p>

          {/* Access Validation Status */}
          {credentialValidating && (
            <div className="flex items-center space-x-2 text-sm text-blue-600">
              <Loader2 className="h-3 w-3 animate-spin" />
              <span>{t("projects.repository.validatingAccess")}</span>
            </div>
          )}

          {accessValidated && !credentialValidating && (
            <div className="text-sm text-green-600">
              ✓ {t("projects.repository.accessValidated")}
            </div>
          )}

          {accessError && !credentialValidating && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription className="flex items-center justify-between">
                <span>{accessError}</span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => validateAccessAndFetchBranches(formData.repo_url, formData.credential_id)}
                  className="h-6 px-2"
                >
                  <RefreshCw className="h-3 w-3" />
                </Button>
              </AlertDescription>
            </Alert>
          )}
        </div>
      </div>
    </form>
  );
}
