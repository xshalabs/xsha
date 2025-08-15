import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type {
  GitCredential,
  GitCredentialType,
  GitCredentialFormData,
} from "@/types/credentials";
import { GitCredentialType as CredentialTypes } from "@/types/credentials";

export interface UseCredentialFormProps {
  credential?: GitCredential;
  onSubmit?: (credential: GitCredential) => Promise<void>;
}

export function useCredentialForm({
  credential,
  onSubmit,
}: UseCredentialFormProps) {
  const { t } = useTranslation();
  const isEdit = !!credential;

  // Form state
  const [formData, setFormData] = useState<GitCredentialFormData>({
    name: credential?.name || "",
    description: credential?.description || "",
    type: credential?.type || CredentialTypes.PASSWORD,
    username: credential?.username || "",
    password: "",
    token: "",
    private_key: "",
    public_key: credential?.public_key || "",
  });

  // UI state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Initialize form data when credential changes
  useEffect(() => {
    if (credential) {
      setFormData({
        name: credential.name,
        description: credential.description,
        type: credential.type,
        username: credential.username,
        password: "",
        token: "",
        private_key: "",
        public_key: credential.public_key || "",
      });
    }
  }, [credential]);

  // Clear username for SSH key type
  useEffect(() => {
    if (formData.type === CredentialTypes.SSH_KEY) {
      setFormData((prev) => ({ ...prev, username: "" }));
    }
  }, [formData.type]);

  // Form validation
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = t("gitCredentials.validation.nameRequired");
    }
    if (!formData.type) {
      newErrors.type = t("gitCredentials.validation.typeRequired");
    }

    if (formData.type !== CredentialTypes.SSH_KEY && !formData.username.trim()) {
      newErrors.username = t("gitCredentials.validation.usernameRequired");
    }

    if (!isEdit) {
      switch (formData.type) {
        case CredentialTypes.PASSWORD:
          if (!formData.password) {
            newErrors.password = t("gitCredentials.validation.passwordRequired");
          }
          break;
        case CredentialTypes.TOKEN:
          if (!formData.token) {
            newErrors.token = t("gitCredentials.validation.tokenRequired");
          }
          break;
        case CredentialTypes.SSH_KEY:
          if (!formData.private_key) {
            newErrors.private_key = t("gitCredentials.validation.privateKeyRequired");
          }
          break;
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Input change handler
  const handleInputChange = (
    field: keyof GitCredentialFormData,
    value: string
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

  // Build submit data
  const buildSubmitData = () => {
    const secretData: Record<string, string> = {};

    switch (formData.type) {
      case CredentialTypes.PASSWORD:
        if (formData.password) secretData.password = formData.password;
        break;
      case CredentialTypes.TOKEN:
        if (formData.token) secretData.password = formData.token;
        break;
      case CredentialTypes.SSH_KEY:
        if (formData.private_key) secretData.private_key = formData.private_key;
        if (formData.public_key) secretData.public_key = formData.public_key;
        break;
    }

    return {
      name: formData.name.trim(),
      description: formData.description.trim(),
      type: formData.type,
      username:
        formData.type === CredentialTypes.SSH_KEY
          ? ""
          : formData.username.trim(),
      secret_data: secretData,
    };
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

      const submitData = buildSubmitData();
      let result: GitCredential;

      if (isEdit && credential) {
        await apiService.gitCredentials.update(credential.id, submitData);
        // Get updated credential
        const response = await apiService.gitCredentials.get(credential.id);
        result = response.credential;
      } else {
        const response = await apiService.gitCredentials.create(submitData);
        result = response.credential;
      }

      if (onSubmit) {
        await onSubmit(result);
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error
          ? error.message
          : isEdit
          ? t("gitCredentials.messages.updateFailed")
          : t("gitCredentials.messages.createFailed");
      setError(errorMessage);
      logError(
        error as Error,
        `Failed to ${isEdit ? "update" : "create"} credential`
      );
      throw error; // Re-throw to let parent component handle it
    } finally {
      setLoading(false);
    }
  };

  return {
    // State
    formData,
    loading,
    error,
    errors,
    isEdit,

    // Handlers
    handleInputChange,
    handleSubmit,

    // Utilities
    validateForm,
    buildSubmitData,
  };
}
