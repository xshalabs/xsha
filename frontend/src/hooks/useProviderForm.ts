import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import { handleApiError, logError } from "@/lib/errors";
import type {
  Provider,
  CreateProviderRequest,
  UpdateProviderRequest,
  ProviderType,
} from "@/types/provider";

export interface ProviderFormData {
  name: string;
  description: string;
  type: ProviderType;
  config: string;
}

export interface ConfigVar {
  id: string;
  key: string;
  value: string;
}

export function useProviderForm(
  provider?: Provider,
  onSubmit?: (provider: Provider) => Promise<void>
) {
  const { t } = useTranslation();
  const isEdit = !!provider;

  // Provider types state
  const [providerTypes, setProviderTypes] = useState<ProviderType[]>([]);
  const [loadingTypes, setLoadingTypes] = useState(true);

  // Form state
  const [formData, setFormData] = useState<ProviderFormData>({
    name: provider?.name || "",
    description: provider?.description || "",
    type: provider?.type || "claude-code",
    config: provider?.config || "",
  });

  // Config variables state
  const [configVars, setConfigVars] = useState<ConfigVar[]>([]);

  // UI state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Load provider types
  useEffect(() => {
    const loadProviderTypes = async () => {
      try {
        setLoadingTypes(true);
        const response = await apiService.providers.getTypes();
        setProviderTypes(response.types);

        // Set default type if creating new provider
        if (!isEdit && response.types.length > 0) {
          setFormData((prev) => ({
            ...prev,
            type: response.types[0],
          }));
        }
      } catch (error: any) {
        const errorMessage = handleApiError(error);
        console.error("Failed to load provider types:", errorMessage);
        // Fallback to default type
        setProviderTypes(["claude-code"]);
        if (!isEdit) {
          setFormData((prev) => ({
            ...prev,
            type: "claude-code",
          }));
        }
      } finally {
        setLoadingTypes(false);
      }
    };

    loadProviderTypes();
  }, [isEdit]);

  // Initialize form data from provider
  useEffect(() => {
    if (provider && isEdit) {
      setFormData({
        name: provider.name,
        description: provider.description,
        type: provider.type,
        config: provider.config,
      });

      // Convert config JSON string to array structure
      try {
        const configObj = JSON.parse(provider.config || "{}");
        const configVarsArray = Object.entries(configObj).map(([key, value], index) => ({
          id: `config-${index}-${Date.now()}`,
          key,
          value: String(value),
        }));
        setConfigVars(configVarsArray);
      } catch (e) {
        console.error("Failed to parse provider config:", e);
        setConfigVars([]);
      }
    } else {
      setConfigVars([]);
    }
    setErrors({});
  }, [provider, isEdit]);

  // Validation
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = t("provider.validation.name_required");
    }

    if (!formData.type) {
      newErrors.type = t("provider.validation.type_required");
    }

    if (configVars.length === 0) {
      newErrors.config = t("provider.validation.config_required");
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Handlers
  const handleInputChange = (field: keyof ProviderFormData, value: string) => {
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

  // Config variables handlers
  const addConfigVar = (key: string, value: string) => {
    if (!key.trim()) {
      toast.error(t("provider.config_vars.key_required"));
      return false;
    }

    if (configVars.some(item => item.key === key)) {
      toast.error(t("provider.config_vars.key_exists"));
      return false;
    }

    setConfigVars((prev) => [
      ...prev,
      {
        id: `config-new-${Date.now()}-${Math.random()}`,
        key,
        value,
      }
    ]);
    return true;
  };

  const removeConfigVar = (id: string) => {
    setConfigVars((prev) => prev.filter(item => item.id !== id));
  };

  const updateConfigVar = (id: string, field: 'key' | 'value', newValue: string) => {
    if (field === 'key' && newValue) {
      const existingItem = configVars.find(item => item.id === id);
      if (existingItem && configVars.some(item => item.id !== id && item.key === newValue)) {
        toast.error(t("provider.config_vars.key_exists"));
        return false;
      }
    }

    setConfigVars((prev) =>
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

      // Convert config variables to JSON string
      const configObj: Record<string, string> = {};
      configVars.forEach(({ key, value }) => {
        if (key.trim()) {
          configObj[key] = value;
        }
      });
      const configString = JSON.stringify(configObj);

      let result: Provider;

      if (isEdit && provider) {
        const requestData: UpdateProviderRequest = {
          name: formData.name,
          description: formData.description,
          config: configString,
        };
        await apiService.providers.update(provider.id, requestData);

        const response = await apiService.providers.get(provider.id);
        result = response.provider;
      } else {
        const requestData: CreateProviderRequest = {
          name: formData.name,
          description: formData.description,
          type: formData.type,
          config: configString,
        };
        const response = await apiService.providers.create(requestData);
        result = response.provider;
      }

      if (onSubmit) {
        await onSubmit(result);
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error
          ? error.message
          : isEdit
          ? t("provider.update_error")
          : t("provider.create_error");
      setError(errorMessage);
      logError(
        error as Error,
        `Failed to ${isEdit ? "update" : "create"} provider`
      );
      throw error;
    } finally {
      setLoading(false);
    }
  };

  return {
    // State
    formData,
    configVars,
    providerTypes,
    loading,
    loadingTypes,
    error,
    errors,
    isEdit,

    // Handlers
    handleInputChange,
    handleSubmit,
    addConfigVar,
    removeConfigVar,
    updateConfigVar,

    // Utilities
    validateForm,
  };
}
