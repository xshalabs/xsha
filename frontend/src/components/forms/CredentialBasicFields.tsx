import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { 
  FileText,
  Key,
  Shield,
  Settings,
  AlertCircle
} from "lucide-react";
import type {
  GitCredentialType,
  GitCredentialFormData,
} from "@/types/credentials";
import { GitCredentialType as CredentialTypes } from "@/types/credentials";

interface CredentialBasicFieldsProps {
  formData: GitCredentialFormData;
  errors: Record<string, string>;
  loading: boolean;
  isEdit: boolean;
  onInputChange: (field: keyof GitCredentialFormData, value: string) => void;
}

export function CredentialBasicFields({
  formData,
  errors,
  loading,
  isEdit,
  onInputChange,
}: CredentialBasicFieldsProps) {
  const { t } = useTranslation();

  return (
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
          onChange={(e) => onInputChange("name", e.target.value)}
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
          onChange={(e) => onInputChange("description", e.target.value)}
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
              onInputChange("type", value as GitCredentialType)
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
              {/* SSH Key support could be enabled by uncommenting */}
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
            onChange={(e) => onInputChange("username", e.target.value)}
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
    </div>
  );
}
