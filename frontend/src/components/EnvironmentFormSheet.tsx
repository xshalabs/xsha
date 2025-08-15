import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
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
  FileText,
  Settings,
  Cpu,
  MemoryStick,
  Plus,
  Trash2,
  Container,
  Shield
} from "lucide-react";
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import { devEnvironmentsApi } from "@/lib/api/environments";
import { handleApiError, logError } from "@/lib/errors";
import type {
  DevEnvironmentDisplay,
  CreateDevEnvironmentRequest,
  UpdateDevEnvironmentRequest,
} from "@/types/dev-environment";

interface EnvironmentImageOption {
  image: string;
  name: string;
  type: string;
  description: string;
}

interface EnvironmentFormSheetProps {
  environment?: DevEnvironmentDisplay;
  onSubmit: (environment: DevEnvironmentDisplay) => Promise<void>;
  onCancel?: () => void;
  formId?: string;
}

const defaultResources = {
  cpu: 1.0,
  memory: 1024,
};

export function EnvironmentFormSheet({
  environment,
  onSubmit,
  onCancel,
  formId = "environment-form-sheet",
}: EnvironmentFormSheetProps) {
  const { t } = useTranslation();
  const isEdit = !!environment;

  const [environmentImages, setEnvironmentImages] = useState<EnvironmentImageOption[]>([]);
  const [loadingImages, setLoadingImages] = useState(true);

  const [formData, setFormData] = useState({
    name: environment?.name || "",
    description: environment?.description || "",
    system_prompt: environment?.system_prompt || "",
    docker_image: environment?.docker_image || "",
    cpu_limit: environment?.cpu_limit || 1.0,
    memory_limit: environment?.memory_limit || 1024,
  });

  const [envVars, setEnvVars] = useState<Record<string, string>>(
    environment?.env_vars_map || {}
  );
  const [newEnvKey, setNewEnvKey] = useState("");
  const [newEnvValue, setNewEnvValue] = useState("");

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    const loadEnvironmentImages = async () => {
      try {
        setLoadingImages(true);
        const response = await devEnvironmentsApi.getAvailableImages();
        const images: EnvironmentImageOption[] = response.images.map((item) => ({
          image: item.image,
          name: item.name,
          type: item.type,
          description: item.name,
        }));
        setEnvironmentImages(images);

        if (!isEdit && images.length > 0) {
          const defaultImage =
            images.find((img) => img.type === "claude-code") || images[0];
          setFormData((prev) => ({
            ...prev,
            docker_image: defaultImage.image,
          }));
        }
      } catch (error: any) {
        const errorMessage = handleApiError(error);
        console.error("Failed to load environment images:", errorMessage);
        const fallbackImage = {
          image: "claude-code:latest",
          name: "Claude Code",
          type: "claude-code",
          description: "Claude Code",
        };
        setEnvironmentImages([fallbackImage]);
        if (!isEdit) {
          setFormData((prev) => ({
            ...prev,
            docker_image: fallbackImage.image,
          }));
        }
      } finally {
        setLoadingImages(false);
      }
    };

    loadEnvironmentImages();
  }, [isEdit]);

  useEffect(() => {
    if (environment && isEdit) {
      setFormData({
        name: environment.name,
        description: environment.description,
        system_prompt: environment.system_prompt || "",
        docker_image: environment.docker_image,
        cpu_limit: environment.cpu_limit,
        memory_limit: environment.memory_limit,
      });
      setEnvVars(environment.env_vars_map || {});
    }
    setErrors({});
    setNewEnvKey("");
    setNewEnvValue("");
  }, [environment, isEdit]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = t("devEnvironments.validation.name_required");
    }

    if (!isEdit && !formData.docker_image.trim()) {
      newErrors.docker_image = t("devEnvironments.validation.docker_image_required");
    }

    if (formData.cpu_limit <= 0 || formData.cpu_limit > 16) {
      newErrors.cpu_limit = t("devEnvironments.validation.cpu_limit_invalid");
    }

    if (formData.memory_limit <= 0 || formData.memory_limit > 32768) {
      newErrors.memory_limit = t("devEnvironments.validation.memory_limit_invalid");
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleInputChange = (
    field: keyof typeof formData,
    value: string | number
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

  const handleDockerImageChange = (dockerImage: string) => {
    const selectedImage = environmentImages.find((img) => img.image === dockerImage);

    setFormData((prev) => ({
      ...prev,
      docker_image: dockerImage,
      cpu_limit: defaultResources.cpu,
      memory_limit: defaultResources.memory,
    }));
  };

  const addEnvVar = () => {
    if (!newEnvKey.trim()) {
      toast.error(t("devEnvironments.env_vars.key_required"));
      return;
    }

    if (envVars[newEnvKey]) {
      toast.error(t("devEnvironments.env_vars.key_exists"));
      return;
    }

    setEnvVars((prev) => ({
      ...prev,
      [newEnvKey]: newEnvValue,
    }));
    setNewEnvKey("");
    setNewEnvValue("");
  };

  const removeEnvVar = (key: string) => {
    setEnvVars((prev) => {
      const newVars = { ...prev };
      delete newVars[key];
      return newVars;
    });
  };

  const updateEnvVar = (key: string, value: string) => {
    setEnvVars((prev) => ({
      ...prev,
      [key]: value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      setLoading(true);
      setError(null);

      let result: DevEnvironmentDisplay;

      if (isEdit && environment) {
        const requestData: UpdateDevEnvironmentRequest = {
          name: formData.name,
          description: formData.description,
          system_prompt: formData.system_prompt,
          cpu_limit: formData.cpu_limit,
          memory_limit: formData.memory_limit,
          env_vars: envVars,
        };
        await apiService.devEnvironments.update(environment.id, requestData);
        
        // Get updated environment
        const response = await apiService.devEnvironments.get(environment.id);
        result = {
          ...response.environment,
          env_vars_map: envVars,
        } as DevEnvironmentDisplay;
      } else {
        const selectedImage = environmentImages.find(
          (img) => img.image === formData.docker_image
        );
        const envType = selectedImage?.type || "claude-code";

        const requestData: CreateDevEnvironmentRequest = {
          ...formData,
          type: envType,
          env_vars: envVars,
        };
        const response = await apiService.devEnvironments.create(requestData);
        result = {
          ...response.environment,
          env_vars_map: envVars,
        } as DevEnvironmentDisplay;
      }

      await onSubmit(result);
    } catch (error) {
      const errorMessage =
        error instanceof Error
          ? error.message
          : isEdit
          ? t("devEnvironments.update_error")
          : t("devEnvironments.create_error");
      setError(errorMessage);
      logError(
        error as Error,
        `Failed to ${isEdit ? "update" : "create"} environment`
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
        {/* Environment Name */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <FileText className="h-4 w-4 text-muted-foreground" />
            <Label htmlFor="name" className="text-sm font-medium">
              {t("devEnvironments.form.name")} <span className="text-red-500">*</span>
            </Label>
          </div>
          <Input
            id="name"
            value={formData.name}
            onChange={(e) => handleInputChange("name", e.target.value)}
            placeholder={t("devEnvironments.form.name_placeholder")}
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
              {t("devEnvironments.form.description")}
            </Label>
          </div>
          <Textarea
            id="description"
            value={formData.description}
            onChange={(e) => handleInputChange("description", e.target.value)}
            placeholder={t("devEnvironments.form.description_placeholder")}
            rows={3}
            disabled={loading}
          />
        </div>

        {/* System Prompt */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Settings className="h-4 w-4 text-muted-foreground" />
            <Label htmlFor="system_prompt" className="text-sm font-medium">
              {t("devEnvironments.form.system_prompt")}
            </Label>
          </div>
          <Textarea
            id="system_prompt"
            value={formData.system_prompt}
            onChange={(e) => handleInputChange("system_prompt", e.target.value)}
            placeholder={t("devEnvironments.form.system_prompt_placeholder")}
            rows={4}
            disabled={loading}
          />
        </div>

        {/* Docker Image (only for creation) */}
        {!isEdit && (
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Container className="h-4 w-4 text-muted-foreground" />
              <Label className="text-sm font-medium">
                {t("devEnvironments.form.docker_image")} <span className="text-red-500">*</span>
              </Label>
              {loadingImages && (
                <Loader2 className="h-3 w-3 animate-spin text-blue-500" />
              )}
            </div>
            <Select
              value={formData.docker_image}
              onValueChange={handleDockerImageChange}
              disabled={loadingImages || loading}
            >
              <SelectTrigger>
                <SelectValue
                  placeholder={
                    loadingImages
                      ? t("common.loading") + "..."
                      : t("devEnvironments.form.docker_image_placeholder")
                  }
                />
              </SelectTrigger>
              <SelectContent className="max-w-[400px]">
                {loadingImages ? (
                  <SelectItem value="loading" disabled>
                    <div className="flex items-center gap-2">
                      <Loader2 className="h-3 w-3 animate-spin" />
                      {t("common.loading")}...
                    </div>
                  </SelectItem>
                ) : environmentImages.length === 0 ? (
                  <SelectItem value="empty" disabled>
                    {t("devEnvironments.noImagesAvailable")}
                  </SelectItem>
                ) : (
                  environmentImages.map((imgOption) => (
                    <SelectItem key={imgOption.image} value={imgOption.image}>
                      <div className="flex items-center justify-between w-full">
                        <span className="font-medium truncate">{imgOption.name}</span>
                        <span className="text-xs text-muted-foreground ml-2">
                          {imgOption.type}
                        </span>
                      </div>
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
            {errors.docker_image && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.docker_image}
              </p>
            )}
          </div>
        )}

        {/* Resource Limits */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* CPU Limit */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Cpu className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="cpu_limit" className="text-sm font-medium">
                {t("devEnvironments.form.cpu_limit")} <span className="text-red-500">*</span>
                <span className="text-xs text-muted-foreground ml-1">
                  (0.1-16 {t("devEnvironments.stats.cores")})
                </span>
              </Label>
            </div>
            <Input
              id="cpu_limit"
              type="number"
              step="0.1"
              min="0.1"
              max="16"
              value={formData.cpu_limit}
              onChange={(e) => handleInputChange("cpu_limit", parseFloat(e.target.value))}
              className={errors.cpu_limit ? "border-red-500 focus-visible:ring-red-500" : ""}
              disabled={loading}
            />
            {errors.cpu_limit && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.cpu_limit}
              </p>
            )}
          </div>

          {/* Memory Limit */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <MemoryStick className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="memory_limit" className="text-sm font-medium">
                {t("devEnvironments.form.memory_limit")} <span className="text-red-500">*</span>
                <span className="text-xs text-muted-foreground ml-1">
                  (128-32768 MB)
                </span>
              </Label>
            </div>
            <Input
              id="memory_limit"
              type="number"
              min="128"
              max="32768"
              value={formData.memory_limit}
              onChange={(e) => handleInputChange("memory_limit", parseInt(e.target.value))}
              className={errors.memory_limit ? "border-red-500 focus-visible:ring-red-500" : ""}
              disabled={loading}
            />
            {errors.memory_limit && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.memory_limit}
              </p>
            )}
          </div>
        </div>

        {/* Environment Variables */}
        <div className="space-y-4">
          <div className="flex items-center gap-2">
            <Settings className="h-4 w-4 text-muted-foreground" />
            <Label className="text-sm font-medium">
              {t("devEnvironments.env_vars.title")}
            </Label>
          </div>
          <div className="space-y-3">
            {Object.keys(envVars).length === 0 && (
              <p className="text-sm text-muted-foreground">
                {t("devEnvironments.env_vars.empty_message")}
              </p>
            )}
            {Object.entries(envVars).map(([key, value]) => (
              <div key={key} className="grid gap-2 grid-cols-5">
                <Input
                  placeholder={t("devEnvironments.env_vars.key")}
                  className="col-span-2"
                  value={key}
                  onChange={(e) => {
                    const newKey = e.target.value;
                    if (newKey !== key) {
                      // Check if new key already exists
                      if (newKey && envVars[newKey]) {
                        toast.error(t("devEnvironments.env_vars.key_exists"));
                        return;
                      }
                      const newVars = { ...envVars };
                      delete newVars[key];
                      if (newKey) {
                        newVars[newKey] = value;
                      }
                      setEnvVars(newVars);
                    }
                  }}
                />
                <Input
                  placeholder={t("devEnvironments.env_vars.value")}
                  className="col-span-2"
                  value={value}
                  onChange={(e) => updateEnvVar(key, e.target.value)}
                />
                <Button
                  type="button"
                  size="icon"
                  variant="ghost"
                  onClick={() => removeEnvVar(key)}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>
          <div className="grid gap-2 grid-cols-5">
            <Input
              placeholder={t("devEnvironments.env_vars.key")}
              className="col-span-2"
              value={newEnvKey}
              onChange={(e) => setNewEnvKey(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && newEnvKey.trim()) {
                  e.preventDefault();
                  addEnvVar();
                }
              }}
            />
            <Input
              placeholder={t("devEnvironments.env_vars.value")}
              className="col-span-2"
              value={newEnvValue}
              onChange={(e) => setNewEnvValue(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && newEnvKey.trim()) {
                  e.preventDefault();
                  addEnvVar();
                }
              }}
            />
            <Button
              type="button"
              size="icon"
              variant="ghost"
              onClick={addEnvVar}
              disabled={!newEnvKey.trim()}
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>
          <div>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={addEnvVar}
              disabled={!newEnvKey.trim()}
            >
              <Plus className="h-4 w-4 mr-2" />
              {t("devEnvironments.env_vars.add")}
            </Button>
          </div>
        </div>

        {/* Configuration Help */}
        <Alert>
          <Shield className="h-4 w-4" />
          <AlertDescription className="text-xs">
            {t("devEnvironments.configurationHelp")}
          </AlertDescription>
        </Alert>
      </div>
    </form>
  );
}
