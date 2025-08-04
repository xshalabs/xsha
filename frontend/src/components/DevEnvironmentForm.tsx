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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Plus, Trash2, Save, X } from "lucide-react";
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import { devEnvironmentsApi } from "@/lib/api/dev-environments";
import type {
  DevEnvironmentDisplay,
  CreateDevEnvironmentRequest,
  UpdateDevEnvironmentRequest,
} from "@/types/dev-environment";

interface DevEnvironmentFormProps {
  onClose: () => void;
  onSuccess: () => void;
  initialData?: DevEnvironmentDisplay | null;
  mode: "create" | "edit";
}

const defaultResources = {
  cpu: 1.0,
  memory: 1024,
};

interface EnvironmentImageOption {
  image: string;
  name: string;
  type: string;
  description: string;
}

const DevEnvironmentForm: React.FC<DevEnvironmentFormProps> = ({
  onClose,
  onSuccess,
  initialData,
  mode,
}) => {
  const { t } = useTranslation();

  const [environmentImages, setEnvironmentImages] = useState<
    EnvironmentImageOption[]
  >([]);
  const [loadingImages, setLoadingImages] = useState(true);

  const [formData, setFormData] = useState({
    name: "",
    description: "",
    docker_image: "",
    cpu_limit: 1.0,
    memory_limit: 1024,
  });

  const [envVars, setEnvVars] = useState<Record<string, string>>({});
  const [newEnvKey, setNewEnvKey] = useState("");
  const [newEnvValue, setNewEnvValue] = useState("");

  const [loading, setLoading] = useState(false);
  const [validationErrors, setValidationErrors] = useState<
    Record<string, string>
  >({});

  useEffect(() => {
    const loadEnvironmentImages = async () => {
      try {
        setLoadingImages(true);
        const response = await devEnvironmentsApi.getAvailableImages();
        const images: EnvironmentImageOption[] = response.images.map(
          (item) => ({
            image: item.image,
            name: item.name,
            type: item.type,
            description: `${item.name} - Docker Image: ${item.image}`,
          })
        );
        setEnvironmentImages(images);

        if (mode === "create" && images.length > 0) {
          const defaultImage =
            images.find((img) => img.type === "claude-code") || images[0];
          setFormData((prev) => ({
            ...prev,
            docker_image: defaultImage.image,
          }));
        }
      } catch (error: any) {
        toast.error(error.message || t("dev_environments.load_images_failed"));
        const fallbackImage = {
          image: "claude-code:latest",
          name: "Claude Code",
          type: "claude-code",
          description: "Claude Code - Docker Image: claude-code:latest",
        };
        setEnvironmentImages([fallbackImage]);
        if (mode === "create") {
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
  }, [mode, t]);

  useEffect(() => {
    if (initialData && mode === "edit") {
      setFormData({
        name: initialData.name,
        description: initialData.description,
        docker_image: initialData.docker_image,
        cpu_limit: initialData.cpu_limit,
        memory_limit: initialData.memory_limit,
      });
      setEnvVars(initialData.env_vars_map || {});
    } else {
      setFormData({
        name: "",
        description: "",
        docker_image: "",
        cpu_limit: 1.0,
        memory_limit: 1024,
      });
      setEnvVars({});
    }
    setValidationErrors({});
    setNewEnvKey("");
    setNewEnvValue("");
  }, [initialData, mode]);

  const validateForm = () => {
    const errors: Record<string, string> = {};

    if (!formData.name.trim()) {
      errors.name = t("dev_environments.validation.name_required");
    }

    if (!formData.docker_image.trim()) {
      errors.docker_image = t(
        "dev_environments.validation.docker_image_required"
      );
    }

    if (formData.cpu_limit <= 0 || formData.cpu_limit > 16) {
      errors.cpu_limit = t("dev_environments.validation.cpu_limit_invalid");
    }

    if (formData.memory_limit <= 0 || formData.memory_limit > 32768) {
      errors.memory_limit = t(
        "dev_environments.validation.memory_limit_invalid"
      );
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleFieldChange = (field: keyof typeof formData, value: any) => {
    setFormData((prev) => ({ ...prev, [field]: value }));

    if (validationErrors[field]) {
      setValidationErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  };

  const handleDockerImageChange = (dockerImage: string) => {
    const selectedImage = environmentImages.find(
      (img) => img.image === dockerImage
    );

    setFormData((prev) => ({
      ...prev,
      docker_image: dockerImage,
      cpu_limit: defaultResources.cpu,
      memory_limit: defaultResources.memory,
      type: selectedImage?.type || "claude-code",
    }));
  };

  const addEnvVar = () => {
    if (!newEnvKey.trim()) {
      toast.error(t("dev_environments.env_vars.key_required"));
      return;
    }

    if (envVars[newEnvKey]) {
      toast.error(t("dev_environments.env_vars.key_exists"));
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

  const handleSubmit = async () => {
    if (!validateForm()) {
      return;
    }

    setLoading(true);
    try {
      if (mode === "create") {
        // 根据选择的镜像找到对应的 type
        const selectedImage = environmentImages.find(
          (img) => img.image === formData.docker_image
        );
        const envType = selectedImage?.type || "claude-code";

        const requestData: CreateDevEnvironmentRequest = {
          ...formData,
          type: envType,
          env_vars: envVars,
        };
        await apiService.devEnvironments.create(requestData);
        toast.success(t("dev_environments.create_success"));
      } else {
        const requestData: UpdateDevEnvironmentRequest = {
          name: formData.name,
          description: formData.description,
          cpu_limit: formData.cpu_limit,
          memory_limit: formData.memory_limit,
          env_vars: envVars,
        };
        await apiService.devEnvironments.update(initialData!.id, requestData);
        toast.success(t("dev_environments.update_success"));
      }

      onSuccess();
    } catch (error: any) {
      console.error("Failed to save environment:", error);
      toast.error(
        mode === "create"
          ? t("dev_environments.create_failed")
          : t("dev_environments.update_failed")
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {mode === "create"
            ? t("dev_environments.create")
            : t("dev_environments.edit")}
        </CardTitle>
        <p className="text-muted-foreground">
          {mode === "create"
            ? t("dev_environments.create_description")
            : t("dev_environments.edit_description")}
        </p>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-6">
          <div className="space-y-4">
            <div className="flex flex-col gap-3">
              <Label htmlFor="name">{t("dev_environments.form.name")} *</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => handleFieldChange("name", e.target.value)}
                placeholder={t("dev_environments.form.name_placeholder")}
                className={validationErrors.name ? "border-destructive" : ""}
              />
              {validationErrors.name && (
                <p className="text-sm text-destructive">
                  {validationErrors.name}
                </p>
              )}
            </div>

            <div className="flex flex-col gap-3">
              <Label htmlFor="description">
                {t("dev_environments.form.description")}
              </Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                  handleFieldChange("description", e.target.value)
                }
                placeholder={t("dev_environments.form.description_placeholder")}
                rows={3}
              />
            </div>

            <div className="flex flex-col gap-3">
              <Label>{t("dev_environments.form.docker_image")} *</Label>
              <Select
                value={formData.docker_image}
                onValueChange={handleDockerImageChange}
                disabled={mode === "edit" || loadingImages}
              >
                <SelectTrigger>
                  <SelectValue
                    placeholder={t(
                      "dev_environments.form.docker_image_placeholder"
                    )}
                  >
                    {loadingImages ? t("common.loading") : undefined}
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {environmentImages.map((imgOption) => (
                    <SelectItem key={imgOption.image} value={imgOption.image}>
                      <div className="flex flex-col">
                        <span className="font-medium">{imgOption.image}</span>
                        <span className="text-xs text-muted-foreground">
                          {imgOption.description}
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {validationErrors.docker_image && (
                <p className="text-sm text-destructive">
                  {validationErrors.docker_image}
                </p>
              )}
              {mode === "edit" && (
                <p className="text-sm text-muted-foreground">
                  {t("dev_environments.form.docker_image_edit_disabled")}
                </p>
              )}
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <h3 className="text-lg font-medium">
              {t("dev_environments.form.resources")}
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex flex-col gap-3">
                <Label htmlFor="cpu_limit">
                  {t("dev_environments.form.cpu_limit")} * (0.1-16{" "}
                  {t("dev_environments.stats.cores")})
                </Label>
                <Input
                  id="cpu_limit"
                  type="number"
                  step="0.1"
                  min="0.1"
                  max="16"
                  value={formData.cpu_limit}
                  onChange={(e) =>
                    handleFieldChange("cpu_limit", parseFloat(e.target.value))
                  }
                  className={
                    validationErrors.cpu_limit ? "border-destructive" : ""
                  }
                />
                {validationErrors.cpu_limit && (
                  <p className="text-sm text-destructive">
                    {validationErrors.cpu_limit}
                  </p>
                )}
              </div>

              <div className="flex flex-col gap-3">
                <Label htmlFor="memory_limit">
                  {t("dev_environments.form.memory_limit")} * (128-32768 MB)
                </Label>
                <Input
                  id="memory_limit"
                  type="number"
                  min="128"
                  max="32768"
                  value={formData.memory_limit}
                  onChange={(e) =>
                    handleFieldChange("memory_limit", parseInt(e.target.value))
                  }
                  className={
                    validationErrors.memory_limit ? "border-destructive" : ""
                  }
                />
                {validationErrors.memory_limit && (
                  <p className="text-sm text-destructive">
                    {validationErrors.memory_limit}
                  </p>
                )}
              </div>
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <h3 className="text-lg font-medium">
              {t("dev_environments.form.env_vars")}
            </h3>

            <Card>
              <CardHeader>
                <CardTitle className="text-base">
                  {t("dev_environments.env_vars.add_new")}
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="flex flex-col gap-3">
                    <Label>{t("dev_environments.env_vars.key")}</Label>
                    <Input
                      value={newEnvKey}
                      onChange={(e) => setNewEnvKey(e.target.value)}
                      placeholder="VARIABLE_NAME"
                    />
                  </div>
                  <div className="flex flex-col gap-3">
                    <Label>{t("dev_environments.env_vars.value")}</Label>
                    <Input
                      value={newEnvValue}
                      onChange={(e) => setNewEnvValue(e.target.value)}
                      placeholder="variable_value"
                    />
                  </div>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  onClick={addEnvVar}
                  disabled={!newEnvKey.trim()}
                >
                  <Plus className="h-4 w-4 mr-2" />
                  {t("dev_environments.env_vars.add")}
                </Button>
              </CardContent>
            </Card>

            {Object.keys(envVars).length > 0 && (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">
                    {t("dev_environments.env_vars.current")}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="flex flex-col gap-3">
                    {Object.entries(envVars).map(([key, value]) => (
                      <div key={key} className="flex items-center gap-3">
                        <Badge
                          variant="outline"
                          className="min-w-0 flex-shrink-0"
                        >
                          {key}
                        </Badge>
                        <Input
                          value={value}
                          onChange={(e) => updateEnvVar(key, e.target.value)}
                          className="flex-1"
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => removeEnvVar(key)}
                          className="flex-shrink-0"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </div>

        <div className="flex gap-2 justify-end">
          <Button variant="outline" onClick={onClose} disabled={loading}>
            <X className="h-4 w-4 mr-2" />
            {t("common.cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={loading}>
            <Save className="h-4 w-4 mr-2" />
            {loading ? t("common.saving") : t("common.save")}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
};

export default DevEnvironmentForm;
