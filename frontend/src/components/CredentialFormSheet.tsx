import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { 
  Save, 
  Loader2, 
  AlertCircle, 
  FileText,
  Key,
  Shield,
  Settings
} from "lucide-react";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type {
  GitCredential,
  GitCredentialType,
  GitCredentialFormData,
} from "@/types/credentials";
import { GitCredentialType as CredentialTypes } from "@/types/credentials";

interface CredentialFormSheetProps {
  credential?: GitCredential;
  onSubmit: (credential: GitCredential) => Promise<void>;
  onCancel?: () => void;
  formId?: string;
}

export function CredentialFormSheet({
  credential,
  onSubmit,
  onCancel,
  formId = "credential-form-sheet",
}: CredentialFormSheetProps) {
  const { t } = useTranslation();
  const isEdit = !!credential;

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

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

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

  useEffect(() => {
    if (formData.type === CredentialTypes.SSH_KEY) {
      setFormData((prev) => ({ ...prev, username: "" }));
    }
  }, [formData.type]);

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

      await onSubmit(result);
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
        {/* Credential Name */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <FileText className="h-4 w-4 text-muted-foreground" />
            <Label htmlFor="name" className="text-sm font-medium">
              {t("gitCredentials.name")} <span className="text-red-500">*</span>
            </Label>
          </div>
          <Input
            id="name"
            type="text"
            value={formData.name}
            onChange={(e) => handleInputChange("name", e.target.value)}
            placeholder={t("gitCredentials.placeholders.name")}
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
              {t("gitCredentials.description")}
            </Label>
          </div>
          <Input
            id="description"
            type="text"
            value={formData.description}
            onChange={(e) => handleInputChange("description", e.target.value)}
            placeholder={t("gitCredentials.placeholders.description")}
            disabled={loading}
          />
        </div>

        {/* Credential Type */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Settings className="h-4 w-4 text-muted-foreground" />
            <Label htmlFor="type" className="text-sm font-medium">
              {t("gitCredentials.type")} <span className="text-red-500">*</span>
            </Label>
          </div>
          {isEdit ? (
            <div className="flex items-center h-9 px-3 py-2 border border-input rounded-md bg-muted text-muted-foreground">
              {formData.type === CredentialTypes.PASSWORD && t("gitCredentials.types.password")}
              {formData.type === CredentialTypes.TOKEN && t("gitCredentials.types.token")}
              {formData.type === CredentialTypes.SSH_KEY && t("gitCredentials.types.ssh_key")}
            </div>
          ) : (
            <Select
              value={formData.type}
              onValueChange={(value) =>
                handleInputChange("type", value as GitCredentialType)
              }
              disabled={loading}
            >
              <SelectTrigger>
                <SelectValue placeholder={t("gitCredentials.type")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={CredentialTypes.PASSWORD}>
                  <div className="flex items-center gap-2">
                    <Key className="h-4 w-4" />
                    {t("gitCredentials.types.password")}
                  </div>
                </SelectItem>
                <SelectItem value={CredentialTypes.TOKEN}>
                  <div className="flex items-center gap-2">
                    <Shield className="h-4 w-4" />
                    {t("gitCredentials.types.token")}
                  </div>
                </SelectItem>
                {/* <SelectItem value={CredentialTypes.SSH_KEY}>
                  <div className="flex items-center gap-2">
                    <Key className="h-4 w-4" />
                    {t("gitCredentials.types.ssh_key")}
                  </div>
                </SelectItem> */}
              </SelectContent>
            </Select>
          )}
          {errors.type && (
            <p className="text-sm text-red-500 flex items-center gap-1">
              <AlertCircle className="h-3 w-3" />
              {errors.type}
            </p>
          )}
        </div>

        {/* Username (not for SSH keys) */}
        {formData.type !== CredentialTypes.SSH_KEY && (
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Key className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="username" className="text-sm font-medium">
                {t("gitCredentials.username")} <span className="text-red-500">*</span>
              </Label>
            </div>
            <Input
              id="username"
              type="text"
              value={formData.username}
              onChange={(e) => handleInputChange("username", e.target.value)}
              placeholder={t("gitCredentials.placeholders.username")}
              className={errors.username ? "border-red-500 focus-visible:ring-red-500" : ""}
              disabled={loading}
            />
            {errors.username && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.username}
              </p>
            )}
          </div>
        )}

        {/* Secret Fields */}
        {formData.type === CredentialTypes.PASSWORD && (
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Shield className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="password" className="text-sm font-medium">
                {t("gitCredentials.password")}
                {!isEdit && <span className="text-red-500"> *</span>}
              </Label>
            </div>
            <Input
              id="password"
              type="password"
              value={formData.password}
              onChange={(e) => handleInputChange("password", e.target.value)}
              placeholder={
                isEdit
                  ? t("gitCredentials.placeholders.passwordOptional")
                  : t("gitCredentials.placeholders.password")
              }
              className={errors.password ? "border-red-500 focus-visible:ring-red-500" : ""}
              disabled={loading}
            />
            {errors.password && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.password}
              </p>
            )}
          </div>
        )}

        {formData.type === CredentialTypes.TOKEN && (
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Shield className="h-4 w-4 text-muted-foreground" />
              <Label htmlFor="token" className="text-sm font-medium">
                {t("gitCredentials.token")}
                {!isEdit && <span className="text-red-500"> *</span>}
              </Label>
            </div>
            <Input
              id="token"
              type="password"
              value={formData.token}
              onChange={(e) => handleInputChange("token", e.target.value)}
              placeholder={
                isEdit
                  ? t("gitCredentials.placeholders.tokenOptional")
                  : t("gitCredentials.placeholders.token")
              }
              className={errors.token ? "border-red-500 focus-visible:ring-red-500" : ""}
              disabled={loading}
            />
            {errors.token && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.token}
              </p>
            )}
          </div>
        )}

        {formData.type === CredentialTypes.SSH_KEY && (
          <>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Key className="h-4 w-4 text-muted-foreground" />
                <Label htmlFor="private_key" className="text-sm font-medium">
                  {t("gitCredentials.privateKey")}
                  {!isEdit && <span className="text-red-500"> *</span>}
                </Label>
              </div>
              <Textarea
                id="private_key"
                value={formData.private_key}
                onChange={(e) => handleInputChange("private_key", e.target.value)}
                placeholder={
                  isEdit
                    ? t("gitCredentials.placeholders.privateKeyOptional")
                    : t("gitCredentials.placeholders.privateKey")
                }
                className={`min-h-[120px] font-mono text-sm ${
                  errors.private_key ? "border-red-500 focus-visible:ring-red-500" : ""
                }`}
                disabled={loading}
              />
              {errors.private_key && (
                <p className="text-sm text-red-500 flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.private_key}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Key className="h-4 w-4 text-muted-foreground" />
                <Label htmlFor="public_key" className="text-sm font-medium">
                  {t("gitCredentials.publicKey")}
                </Label>
              </div>
              <Textarea
                id="public_key"
                value={formData.public_key}
                onChange={(e) => handleInputChange("public_key", e.target.value)}
                placeholder={
                  isEdit
                    ? t("gitCredentials.placeholders.publicKeyOptional")
                    : t("gitCredentials.placeholders.publicKey")
                }
                className="min-h-[80px] font-mono text-sm"
                disabled={loading}
              />
            </div>
          </>
        )}

        {/* Security Notice */}
        <Alert>
          <Shield className="h-4 w-4" />
          <AlertDescription className="text-xs">
            {t("gitCredentials.securityNotice")}
          </AlertDescription>
        </Alert>
      </div>
    </form>
  );
}
