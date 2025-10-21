import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import { devEnvironmentsApi } from "@/lib/api/environments";
import { handleApiError, logError } from "@/lib/errors";
import type {
  DevEnvironmentDisplay,
  CreateDevEnvironmentRequest,
  UpdateDevEnvironmentRequest,
} from "@/types/dev-environment";
import type { Provider } from "@/types/provider";

export interface EnvironmentImageOption {
  image: string;
  name: string;
  type: string;
  description: string;
}

export interface EnvironmentFormData {
  name: string;
  description: string;
  system_prompt: string;
  docker_image: string;
  cpu_limit: number;
  memory_limit: number;
  provider_id?: number;
}

export interface EnvVar {
  id: string;
  key: string;
  value: string;
}

const defaultResources = {
  cpu: 1.0,
  memory: 1024,
};

export function useEnvironmentForm(
  environment?: DevEnvironmentDisplay,
  onSubmit?: (environment: DevEnvironmentDisplay) => Promise<void>
) {
  const { t } = useTranslation();
  const isEdit = !!environment;

  // Environment images state
  const [environmentImages, setEnvironmentImages] = useState<EnvironmentImageOption[]>([]);
  const [loadingImages, setLoadingImages] = useState(true);

  // Providers state
  const [providers, setProviders] = useState<Provider[]>([]);
  const [loadingProviders, setLoadingProviders] = useState(true);

  // Form state
  const [formData, setFormData] = useState<EnvironmentFormData>({
    name: environment?.name || "",
    description: environment?.description || "",
    system_prompt: environment?.system_prompt || "",
    docker_image: environment?.docker_image || "",
    cpu_limit: environment?.cpu_limit || 1.0,
    memory_limit: environment?.memory_limit || 1024,
    provider_id: environment?.provider_id,
  });

  // Environment variables state
  const [envVars, setEnvVars] = useState<EnvVar[]>([]);

  // UI state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Load environment images
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

  // Load providers
  useEffect(() => {
    const loadProviders = async () => {
      try {
        setLoadingProviders(true);
        const response = await apiService.providers.list({ page_size: 100 });
        setProviders(response.providers);
      } catch (error: any) {
        const errorMessage = handleApiError(error);
        console.error("Failed to load providers:", errorMessage);
        // Providers are optional, so we just set empty array
        setProviders([]);
      } finally {
        setLoadingProviders(false);
      }
    };

    loadProviders();
  }, []);

  // Initialize form data from environment
  useEffect(() => {
    if (environment && isEdit) {
      setFormData({
        name: environment.name,
        description: environment.description,
        system_prompt: environment.system_prompt || "",
        docker_image: environment.docker_image,
        cpu_limit: environment.cpu_limit,
        memory_limit: environment.memory_limit,
        provider_id: environment.provider_id,
      });

      // Convert env_vars_map to array structure
      const envVarsArray = Object.entries(environment.env_vars_map || {}).map(([key, value], index) => ({
        id: `env-${index}-${Date.now()}`,
        key,
        value
      }));
      setEnvVars(envVarsArray);
    } else {
      setEnvVars([]);
    }
    setErrors({});
  }, [environment, isEdit]);

  // Validation
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = t("devEnvironments.validation.name_required");
    }

    if (!formData.docker_image.trim()) {
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

  // Handlers
  const handleInputChange = (field: keyof EnvironmentFormData, value: string | number) => {
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
    setFormData((prev) => ({
      ...prev,
      docker_image: dockerImage,
      cpu_limit: defaultResources.cpu,
      memory_limit: defaultResources.memory,
    }));
  };

  // Environment variables handlers
  const addEnvVar = (key: string, value: string) => {
    if (!key.trim()) {
      toast.error(t("devEnvironments.env_vars.key_required"));
      return false;
    }

    if (envVars.some(item => item.key === key)) {
      toast.error(t("devEnvironments.env_vars.key_exists"));
      return false;
    }

    setEnvVars((prev) => [
      ...prev,
      {
        id: `env-new-${Date.now()}-${Math.random()}`,
        key,
        value,
      }
    ]);
    return true;
  };

  const removeEnvVar = (id: string) => {
    setEnvVars((prev) => prev.filter(item => item.id !== id));
  };

  const updateEnvVar = (id: string, field: 'key' | 'value', newValue: string) => {
    if (field === 'key' && newValue) {
      const existingItem = envVars.find(item => item.id === id);
      if (existingItem && envVars.some(item => item.id !== id && item.key === newValue)) {
        toast.error(t("devEnvironments.env_vars.key_exists"));
        return false;
      }
    }

    setEnvVars((prev) =>
      prev.map(item =>
        item.id === id
          ? { ...item, [field]: newValue }
          : item
      )
    );
    return true;
  };

  // Submit handler
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      setLoading(true);
      setError(null);

      // Convert environment variables to object
      const envVarsObj: Record<string, string> = {};
      envVars.forEach(({ key, value }) => {
        if (key.trim()) {
          envVarsObj[key] = value;
        }
      });

      let result: DevEnvironmentDisplay;

      if (isEdit && environment) {
        const requestData: UpdateDevEnvironmentRequest = {
          name: formData.name,
          description: formData.description,
          system_prompt: formData.system_prompt,
          docker_image: formData.docker_image,
          cpu_limit: formData.cpu_limit,
          memory_limit: formData.memory_limit,
          env_vars: envVarsObj,
          provider_id: formData.provider_id,
        };
        await apiService.devEnvironments.update(environment.id, requestData);

        const response = await apiService.devEnvironments.get(environment.id);
        result = {
          ...response.environment,
          env_vars_map: envVarsObj,
        } as DevEnvironmentDisplay;
      } else {
        const selectedImage = environmentImages.find(
          (img) => img.image === formData.docker_image
        );
        const envType = selectedImage?.type || "claude-code";

        const requestData: CreateDevEnvironmentRequest = {
          ...formData,
          type: envType,
          env_vars: envVarsObj,
          provider_id: formData.provider_id,
        };
        const response = await apiService.devEnvironments.create(requestData);
        result = {
          ...response.environment,
          env_vars_map: envVarsObj,
        } as DevEnvironmentDisplay;
      }

      if (onSubmit) {
        await onSubmit(result);
      }
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
      throw error;
    } finally {
      setLoading(false);
    }
  };

  return {
    // State
    formData,
    envVars,
    environmentImages,
    providers,
    loading,
    loadingImages,
    loadingProviders,
    error,
    errors,
    isEdit,

    // Handlers
    handleInputChange,
    handleDockerImageChange,
    handleSubmit,
    addEnvVar,
    removeEnvVar,
    updateEnvVar,

    // Utilities
    validateForm,
  };
}
