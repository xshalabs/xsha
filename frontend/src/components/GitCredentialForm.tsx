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

  // 初始化表单数据
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

  // 表单字段更新
  const updateField = (field: keyof GitCredentialFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  // 表单验证
  const validateForm = (): string | null => {
    if (!formData.name.trim())
      return t("gitCredentials.validation.nameRequired");
    if (!formData.type) return t("gitCredentials.validation.typeRequired");
    if (!formData.username.trim())
      return t("gitCredentials.validation.usernameRequired");

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

    return null;
  };

  // 构建提交数据
  const buildSubmitData = () => {
    const secretData: Record<string, string> = {};

    switch (formData.type) {
      case CredentialTypes.PASSWORD:
        if (formData.password) secretData.password = formData.password;
        break;
      case CredentialTypes.TOKEN:
        if (formData.token) secretData.password = formData.token; // token存储在password字段
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
      username: formData.username.trim(),
      secret_data: secretData,
    };
  };

  // 提交表单
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
        // 编辑模式
        await apiService.gitCredentials.update(credential.id, submitData);
      } else {
        // 创建模式
        await apiService.gitCredentials.create(submitData);
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
            <div>
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

            <div>
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

            <div>
              <Label htmlFor="type">{t("gitCredentials.type")} *</Label>
              <select
                id="type"
                value={formData.type}
                onChange={(e) =>
                  updateField("type", e.target.value as GitCredentialType)
                }
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
                disabled={isEditing} // 编辑时不允许修改类型
              >
                <option value={CredentialTypes.PASSWORD}>
                  {t("gitCredentials.types.password")}
                </option>
                <option value={CredentialTypes.TOKEN}>
                  {t("gitCredentials.types.token")}
                </option>
                <option value={CredentialTypes.SSH_KEY}>
                  {t("gitCredentials.types.ssh_key")}
                </option>
              </select>
            </div>

            <div>
              <Label htmlFor="username">{t("gitCredentials.username")} *</Label>
              <Input
                id="username"
                type="text"
                value={formData.username}
                onChange={(e) => updateField("username", e.target.value)}
                placeholder={t("gitCredentials.placeholders.username")}
                required
              />
            </div>
          </div>

          <div className="space-y-4">
            <h3 className="text-lg font-medium">
              {t("gitCredentials.credentialInfo", "Credential Information")}
            </h3>

            {formData.type === CredentialTypes.PASSWORD && (
              <div>
                <Label htmlFor="password">
                  {t("gitCredentials.password")} *
                </Label>
                <Input
                  id="password"
                  type="password"
                  value={formData.password}
                  onChange={(e) => updateField("password", e.target.value)}
                  placeholder={t("gitCredentials.placeholders.password")}
                  required
                />
              </div>
            )}

            {formData.type === CredentialTypes.TOKEN && (
              <div>
                <Label htmlFor="token">{t("gitCredentials.token")} *</Label>
                <Input
                  id="token"
                  type="password"
                  value={formData.token}
                  onChange={(e) => updateField("token", e.target.value)}
                  placeholder={t("gitCredentials.placeholders.token")}
                  required
                />
              </div>
            )}

            {formData.type === CredentialTypes.SSH_KEY && (
              <>
                <div>
                  <Label htmlFor="private_key">
                    {t("gitCredentials.privateKey")} *
                  </Label>
                  <textarea
                    id="private_key"
                    value={formData.private_key}
                    onChange={(e) => updateField("private_key", e.target.value)}
                    placeholder={t("gitCredentials.placeholders.privateKey")}
                    className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500 min-h-[120px] font-mono text-sm"
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="public_key">
                    {t("gitCredentials.publicKey")}
                  </Label>
                  <textarea
                    id="public_key"
                    value={formData.public_key}
                    onChange={(e) => updateField("public_key", e.target.value)}
                    placeholder={t("gitCredentials.placeholders.publicKey")}
                    className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500 min-h-[80px] font-mono text-sm"
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
              {loading
                ? t("common.saving", "Saving...")
                : t(isEditing ? "common.save" : "common.submit")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
};

export default GitCredentialForm;
