import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type {
  Notifier,
  NotifierFormData,
  NotifierTypeInfo,
  NotifierType,
} from "@/types/notifier";

interface UseNotifierFormOptions {
  notifier?: Notifier;
  onSubmit: (notifier: Notifier) => Promise<void>;
}

export function useNotifierForm({ notifier, onSubmit }: UseNotifierFormOptions) {
  const { t } = useTranslation();
  const isEdit = !!notifier;

  const [formData, setFormData] = useState<NotifierFormData>({
    name: notifier?.name || "",
    description: notifier?.description || "",
    type: notifier?.type || ("" as NotifierType),
    config: notifier?.config
      ? (typeof notifier.config === "string" ? JSON.parse(notifier.config as string) : notifier.config)
      : {},
  });

  const [notifierTypes, setNotifierTypes] = useState<NotifierTypeInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [typesLoading, setTypesLoading] = useState(false);
  const [error, setError] = useState<string>("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Load notifier types
  const loadNotifierTypes = useCallback(async () => {
    try {
      setTypesLoading(true);
      const response = await apiService.notifiers.getTypes();
      setNotifierTypes(response.data);
    } catch (error) {
      logError(error, "Failed to load notifier types");
      setError(t("notifiers.errors.loadTypesFailed"));
    } finally {
      setTypesLoading(false);
    }
  }, [t]);

  useEffect(() => {
    loadNotifierTypes();
  }, [loadNotifierTypes]);

  // Reset form to initial state
  const resetForm = useCallback(() => {
    setFormData({
      name: "",
      description: "",
      type: "" as NotifierType,
      config: {},
    });
    setError("");
    setErrors({});
  }, []);

  // Initialize form data when editing
  useEffect(() => {
    if (notifier) {
      setFormData({
        name: notifier.name,
        description: notifier.description,
        type: notifier.type,
        config: typeof notifier.config === "string"
          ? JSON.parse(notifier.config as string)
          : notifier.config,
      });
      setError("");
      setErrors({});
    } else {
      // Reset form when no notifier is provided (create mode)
      resetForm();
    }
  }, [notifier, resetForm]);

  const handleInputChange = useCallback((
    field: keyof NotifierFormData,
    value: string | NotifierType | Record<string, unknown>
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));

    // Clear field error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({
        ...prev,
        [field]: "",
      }));
    }

    // Clear general error
    if (error) {
      setError("");
    }
  }, [errors, error]);

  const handleConfigChange = useCallback((
    configField: string,
    value: string | Record<string, unknown>
  ) => {
    setFormData((prev) => ({
      ...prev,
      config: {
        ...prev.config,
        [configField]: value,
      },
    }));

    // Clear config field error
    if (errors[`config.${configField}`]) {
      setErrors((prev) => ({
        ...prev,
        [`config.${configField}`]: "",
      }));
    }

    // Clear general error
    if (error) {
      setError("");
    }
  }, [errors, error]);

  const getSelectedTypeInfo = useCallback(() => {
    return notifierTypes.find(type => type.type === formData.type);
  }, [notifierTypes, formData.type]);

  const getFieldLabel = useCallback((fieldName: string) => {
    const labelMap: Record<string, string> = {
      webhook_url: t("notifiers.form.fields.webhookUrl.label"),
      url: t("notifiers.form.fields.url.label"),
      secret: t("notifiers.form.fields.secret.label"),
      method: t("notifiers.form.fields.method.label"),
      headers: t("notifiers.form.fields.headers.label"),
      body_template: t("notifiers.form.fields.bodyTemplate.label"),
    };
    return labelMap[fieldName] || fieldName;
  }, [t]);

  const validateForm = useCallback((): boolean => {
    const newErrors: Record<string, string> = {};

    // Validate basic fields
    if (!formData.name.trim()) {
      newErrors.name = t("notifiers.form.fields.name.required");
    }

    if (!formData.type) {
      newErrors.type = t("notifiers.form.fields.type.required");
    }

    // Validate config fields
    const typeInfo = getSelectedTypeInfo();
    if (typeInfo) {
      typeInfo.config_schema.forEach((fieldInfo) => {
        const fieldName = fieldInfo.name;

        if (fieldInfo.required && !formData.config?.[fieldName]) {
          newErrors[`config.${fieldName}`] = t("common.fieldRequired", {
            field: getFieldLabel(fieldName)
          });
        }
      });
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [formData, t, getSelectedTypeInfo, getFieldLabel]);

  const handleSubmit = useCallback(async (): Promise<void> => {
    setError("");

    if (!validateForm()) {
      return;
    }

    try {
      setLoading(true);

      const payload = {
        name: formData.name.trim(),
        description: formData.description.trim(),
        ...(!isEdit && { type: formData.type }), // Only include type when creating
        config: formData.config,
      };

      if (isEdit && notifier) {
        await apiService.notifiers.update(notifier.id, payload);
        // For edit, create a Notifier object with updated fields
        const updatedNotifier: Notifier = {
          ...notifier,
          name: payload.name,
          description: payload.description,
          config: JSON.stringify(payload.config), // Convert NotifierConfig to string
        };
        await onSubmit(updatedNotifier);
      } else {
        const result = await apiService.notifiers.create(payload as any);
        // Convert CreateNotifierResponse to Notifier format
        const notifierResult: Notifier = {
          ...result,
          config: JSON.stringify(result.config), // Convert NotifierConfig to string
        } as Notifier;
        await onSubmit(notifierResult);
      }
    } catch (error) {
      logError(error, "Failed to save notifier");
      setError(t("notifiers.errors.saveFailed"));
      throw error; // Re-throw to let caller handle
    } finally {
      setLoading(false);
    }
  }, [formData, isEdit, notifier, onSubmit, t, validateForm]);

  return {
    formData,
    notifierTypes,
    loading,
    typesLoading,
    error,
    errors,
    isEdit,
    handleInputChange,
    handleConfigChange,
    handleSubmit,
    getSelectedTypeInfo,
    getFieldLabel,
    resetForm,
  };
}