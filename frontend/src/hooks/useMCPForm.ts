import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { apiService } from "@/lib/api/index";
import { logError, ApiError, handleApiError } from "@/lib/errors";
import type { MCP, MCPFormData } from "@/types/mcp";

interface UseMCPFormOptions {
  mcp?: MCP;
  onSubmit: (mcp: MCP) => Promise<void>;
}

export function useMCPForm({ mcp, onSubmit }: UseMCPFormOptions) {
  const { t } = useTranslation();
  const isEdit = !!mcp;

  const [formData, setFormData] = useState<MCPFormData>({
    name: mcp?.name || "",
    description: mcp?.description || "",
    config: mcp?.config || "{}",
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Reset form to initial state
  const resetForm = useCallback(() => {
    setFormData({
      name: "",
      description: "",
      config: "{}",
    });
    setError("");
    setErrors({});
  }, []);

  // Initialize form data when editing
  useEffect(() => {
    if (mcp) {
      setFormData({
        name: mcp.name,
        description: mcp.description,
        config: mcp.config,
      });
      setError("");
      setErrors({});
    } else {
      // Reset form when no MCP is provided (create mode)
      resetForm();
    }
  }, [mcp, resetForm]);

  const handleInputChange = useCallback((
    field: keyof MCPFormData,
    value: string
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

  const handleConfigChange = useCallback((config: string) => {
    setFormData((prev) => ({
      ...prev,
      config,
    }));

    // Clear config error
    if (errors.config) {
      setErrors((prev) => ({
        ...prev,
        config: "",
      }));
    }

    // Clear general error
    if (error) {
      setError("");
    }
  }, [errors, error]);


  const validateForm = useCallback((): boolean => {
    const newErrors: Record<string, string> = {};

    // Validate basic fields
    if (!formData.name.trim()) {
      newErrors.name = t("mcp.form.fields.name.required");
    } else {
      // Validate name format: only allow letters, numbers, underscore, and hyphen
      const namePattern = /^[a-zA-Z0-9_-]+$/;
      if (!namePattern.test(formData.name.trim())) {
        newErrors.name = t("mcp.form.validation.invalidFormat");
      }
    }

    if (!formData.config.trim()) {
      newErrors.config = t("mcp.form.fields.config.required");
    } else {
      // Validate JSON format
      try {
        JSON.parse(formData.config);
      } catch {
        newErrors.config = t("mcp.form.validation.invalidJson");
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [formData, t]);

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
        config: formData.config,
        enabled: isEdit && mcp ? mcp.enabled : true, // Preserve status when editing, default to true when creating
      };

      if (isEdit && mcp) {
        await apiService.mcp.update(mcp.id, payload);
        // For edit, create a MCP object with updated fields
        const updatedMCP: MCP = {
          ...mcp,
          name: payload.name,
          description: payload.description,
          config: payload.config,
          enabled: payload.enabled,
        };
        await onSubmit(updatedMCP);
      } else {
        const result = await apiService.mcp.create(payload);
        // Convert CreateMCPResponse to MCP format
        const mcpResult: MCP = {
          ...result,
          config: result.config as string, // Ensure config is string
        } as MCP;
        await onSubmit(mcpResult);
      }
    } catch (error) {
      logError(error, "Failed to save MCP configuration");

      // Extract the actual error message from the server response
      const errorMessage = handleApiError(error);
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }, [formData, isEdit, mcp, onSubmit, t, validateForm]);

  return {
    formData,
    loading,
    error,
    errors,
    isEdit,
    handleInputChange,
    handleConfigChange,
    handleSubmit,
    resetForm,
  };
}