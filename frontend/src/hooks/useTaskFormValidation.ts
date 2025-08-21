import { useState, useCallback } from "react";
import { useTranslation } from "react-i18next";
import type { TaskFormData } from "@/types/task";

export function useTaskFormValidation() {
  const { t } = useTranslation();
  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = useCallback((formData: TaskFormData): boolean => {
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
  }, [t]);

  const clearFieldError = useCallback((field: keyof TaskFormData) => {
    if (errors[field]) {
      setErrors((prev) => ({
        ...prev,
        [field]: "",
      }));
    }
  }, [errors]);

  const clearAllErrors = useCallback(() => {
    setErrors({});
  }, []);

  return {
    errors,
    validateForm,
    clearFieldError,
    clearAllErrors,
  };
}