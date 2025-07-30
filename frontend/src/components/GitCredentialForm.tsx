import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import type {
  GitCredential,
  GitCredentialType,
  GitCredentialFormData,
} from "@/types/git-credentials";
import { GitCredentialType as CredentialTypes } from "@/types/git-credentials";

interface GitCredentialFormProps {
  credential?: GitCredential;
  onSuccess: () => void;
  onCancel: () => void;
}

export const GitCredentialForm: React.FC<GitCredentialFormProps> = ({
  credential,
  onSuccess,
  onCancel,
}) => {
  const { t } = useTranslation();
  const [formData, setFormData] = useState<GitCredentialFormData>({
    name: "",
    description: "",
    type: CredentialTypes.PASSWORD,
    username: "",
    password: "",
    token: "",
    private_key: "",
    public_key: "",
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEditing = !!credential;

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

  const updateField = (field: keyof GitCredentialFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const validateForm = (): string | null => {
    if (!formData.name.trim())
      return t("gitCredentials.validation.nameRequired");
    if (!formData.type) return t("gitCredentials.validation.typeRequired");

    if (formData.type !== CredentialTypes.SSH_KEY && !formData.username.trim())
      return t("gitCredentials.validation.usernameRequired");

    if (!isEditing) {
      switch (formData.type) {
        case CredentialTypes.PASSWORD:
          if (!formData.password)
            return t("gitCredentials.validation.passwordRequired");
          break;
        case CredentialTypes.TOKEN:
          if (!formData.token)
            return t("gitCredentials.validation.tokenRequired");
          break;
        case CredentialTypes.SSH_KEY:
          if (!formData.private_key)
            return t("gitCredentials.validation.privateKeyRequired");
          break;
      }
    }

    return null;
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

    const validationError = validateForm();
    if (validationError) {
      setError(validationError);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const submitData = buildSubmitData();

      if (isEditing && credential) {
        await apiService.gitCredentials.update(credential.id, submitData);
        toast.success(t("gitCredentials.messages.updateSuccess"));
      } else {
        await apiService.gitCredentials.create(submitData);
        toast.success(t("gitCredentials.messages.createSuccess"));
      }

      onSuccess();
    } catch (err: any) {
      setError(
        err.message ||
          t(
            isEditing
              ? "gitCredentials.messages.updateFailed"
              : "gitCredentials.messages.createFailed"
          )
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {t(isEditing ? "gitCredentials.edit" : "gitCredentials.create")}
        </CardTitle>
        <CardDescription>
          {t(
            isEditing
              ? "gitCredentials.editDescription"
              : "gitCredentials.createDescription",
            isEditing
              ? "Update existing Git repository access credential"
              : "Create new Git repository access credential"
          )}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-red-600 text-sm">{error}</p>
            </div>
          )}

          <div className="space-y-4">
            <div className="flex flex-col gap-3">
              <Label htmlFor="name">{t("gitCredentials.name")} *</Label>
              <Input
                id="name"
                type="text"
                value={formData.name}
                onChange={(e) => updateField("name", e.target.value)}
                placeholder={t("gitCredentials.placeholders.name")}
                required
              />
            </div>

            <div className="flex flex-col gap-3">
              <Label htmlFor="description">
                {t("gitCredentials.description")}
              </Label>
              <Input
                id="description"
                type="text"
                value={formData.description}
                onChange={(e) => updateField("description", e.target.value)}
                placeholder={t("gitCredentials.placeholders.description")}
              />
            </div>

            <div className="flex flex-col gap-3">
              <Label htmlFor="type">{t("gitCredentials.type")} *</Label>
              <Select
                value={formData.type}
                onValueChange={(value) =>
                  updateField("type", value as GitCredentialType)
                }
                disabled={isEditing}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder={t("gitCredentials.type")} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={CredentialTypes.PASSWORD}>
                    {t("gitCredentials.types.password")}
                  </SelectItem>
                  <SelectItem value={CredentialTypes.TOKEN}>
                    {t("gitCredentials.types.token")}
                  </SelectItem>
                  <SelectItem value={CredentialTypes.SSH_KEY}>
                    {t("gitCredentials.types.ssh_key")}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            {formData.type !== CredentialTypes.SSH_KEY && (
              <div className="flex flex-col gap-3">
                <Label htmlFor="username">
                  {t("gitCredentials.username")} *
                </Label>
                <Input
                  id="username"
                  type="text"
                  value={formData.username}
                  onChange={(e) => updateField("username", e.target.value)}
                  placeholder={t("gitCredentials.placeholders.username")}
                  required
                />
              </div>
            )}
          </div>

          <div className="space-y-4">
            <h3 className="text-lg font-medium">
              {t("gitCredentials.credentialInfo", "Credential Information")}
              {isEditing && (
                <span className="text-sm font-normal text-gray-500 ml-2">
                  (
                  {t(
                    "gitCredentials.optionalWhenEditing",
                    "Optional when editing - leave blank to keep current"
                  )}
                  )
                </span>
              )}
            </h3>

            {formData.type === CredentialTypes.PASSWORD && (
              <div className="flex flex-col gap-3">
                <Label htmlFor="password">
                  {t("gitCredentials.password")}
                  {!isEditing && " *"}
                </Label>
                <Input
                  id="password"
                  type="password"
                  value={formData.password}
                  onChange={(e) => updateField("password", e.target.value)}
                  placeholder={
                    isEditing
                      ? t(
                          "gitCredentials.placeholders.passwordOptional",
                          "Leave blank to keep current password"
                        )
                      : t("gitCredentials.placeholders.password")
                  }
                  required={!isEditing}
                />
              </div>
            )}

            {formData.type === CredentialTypes.TOKEN && (
              <div className="flex flex-col gap-3">
                <Label htmlFor="token">
                  {t("gitCredentials.token")}
                  {!isEditing && " *"}
                </Label>
                <Input
                  id="token"
                  type="password"
                  value={formData.token}
                  onChange={(e) => updateField("token", e.target.value)}
                  placeholder={
                    isEditing
                      ? t(
                          "gitCredentials.placeholders.tokenOptional",
                          "Leave blank to keep current token"
                        )
                      : t("gitCredentials.placeholders.token")
                  }
                  required={!isEditing}
                />
              </div>
            )}

            {formData.type === CredentialTypes.SSH_KEY && (
              <>
                <div className="flex flex-col gap-3">
                  <Label htmlFor="private_key">
                    {t("gitCredentials.privateKey")}
                    {!isEditing && " *"}
                  </Label>
                  <Textarea
                    id="private_key"
                    value={formData.private_key}
                    onChange={(e) => updateField("private_key", e.target.value)}
                    placeholder={
                      isEditing
                        ? t(
                            "gitCredentials.placeholders.privateKeyOptional",
                            "Leave blank to keep current private key"
                          )
                        : t("gitCredentials.placeholders.privateKey")
                    }
                    className="min-h-[120px] font-mono text-sm"
                    required={!isEditing}
                  />
                </div>
                <div className="flex flex-col gap-3">
                  <Label htmlFor="public_key">
                    {t("gitCredentials.publicKey")}
                  </Label>
                  <Textarea
                    id="public_key"
                    value={formData.public_key}
                    onChange={(e) => updateField("public_key", e.target.value)}
                    placeholder={
                      isEditing
                        ? t(
                            "gitCredentials.placeholders.publicKeyOptional",
                            "Leave blank to keep current public key"
                          )
                        : t("gitCredentials.placeholders.publicKey")
                    }
                    className="min-h-[80px] font-mono text-sm"
                  />
                </div>
              </>
            )}
          </div>

          <div className="flex justify-end space-x-3">
            <Button type="button" variant="outline" onClick={onCancel}>
              {t("common.cancel")}
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? t("common.saving", "Saving...") : t("common.save")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
};

export default GitCredentialForm;
