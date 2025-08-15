import React from "react";
import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { 
  Key,
  Shield,
  AlertCircle
} from "lucide-react";
import type {
  GitCredentialFormData,
} from "@/types/credentials";
import { GitCredentialType as CredentialTypes } from "@/types/credentials";

interface CredentialSecretFieldsProps {
  formData: GitCredentialFormData;
  errors: Record<string, string>;
  loading: boolean;
  isEdit: boolean;
  onInputChange: (field: keyof GitCredentialFormData, value: string) => void;
}

export function CredentialSecretFields({
  formData,
  errors,
  loading,
  isEdit,
  onInputChange,
}: CredentialSecretFieldsProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      {/* Password Field */}
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
            onChange={(e) => onInputChange("password", e.target.value)}
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

      {/* Token Field */}
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
            onChange={(e) => onInputChange("token", e.target.value)}
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

      {/* SSH Key Fields */}
      {formData.type === CredentialTypes.SSH_KEY && (
        <>
          {/* Private Key */}
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
              onChange={(e) => onInputChange("private_key", e.target.value)}
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

          {/* Public Key */}
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
              onChange={(e) => onInputChange("public_key", e.target.value)}
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
  );
}
