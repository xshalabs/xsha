import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle, GitBranch, Loader2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { ProjectFormData } from "@/types/project";

interface CredentialOption {
  id: number;
  name: string;
  type: string;
  username: string;
}

interface ProjectCredentialSelectorProps {
  formData: ProjectFormData;
  credentials: CredentialOption[];
  disabled?: boolean;
  credentialsLoading?: boolean;
  credentialValidating?: boolean;
  accessValidated?: boolean;
  accessError?: string | null;
  onChange: (field: keyof ProjectFormData, value: number | undefined) => void;
}

export function ProjectCredentialSelector({
  formData,
  credentials,
  disabled = false,
  credentialsLoading = false,
  credentialValidating = false,
  accessValidated = false,
  accessError = null,
  onChange,
}: ProjectCredentialSelectorProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <GitBranch className="h-4 w-4 text-muted-foreground" />
        <Label htmlFor="credential_id" className="text-sm font-medium">
          {t("projects.credential")}
        </Label>
        {credentialsLoading && (
          <Loader2 className="h-3 w-3 animate-spin text-blue-500" />
        )}
      </div>

      <Select
        onValueChange={(value) =>
          onChange("credential_id", value ? Number(value) : undefined)
        }
        value={formData.credential_id?.toString()}
        disabled={credentialsLoading || disabled}
      >
        <SelectTrigger>
          <SelectValue
            placeholder={
              credentialsLoading
                ? t("common.loading") + "..."
                : t("projects.placeholders.selectCredential")
            }
          />
        </SelectTrigger>
        <SelectContent>
          {credentialsLoading ? (
            <SelectItem value="loading" disabled>
              <div className="flex items-center gap-2">
                <Loader2 className="h-3 w-3 animate-spin" />
                {t("common.loading")}...
              </div>
            </SelectItem>
          ) : credentials.length === 0 ? (
            <SelectItem value="empty" disabled>
              {t("projects.noCredentialsAvailable")}
            </SelectItem>
          ) : (
            credentials.map((credential) => (
              <SelectItem key={credential.id} value={credential.id.toString()}>
                <div className="flex items-center justify-between w-full">
                  <span className="font-medium">{credential.name}</span>
                  <span className="text-xs text-muted-foreground ml-2">
                    {credential.type} - {credential.username}
                  </span>
                </div>
              </SelectItem>
            ))
          )}
        </SelectContent>
      </Select>

      <p className="text-xs text-muted-foreground">
        {t("projects.credentialHelp")}
      </p>

      {/* Access Validation Status */}
      {credentialValidating && (
        <div className="flex items-center space-x-2 text-sm text-blue-600">
          <Loader2 className="h-3 w-3 animate-spin" />
          <span>{t("projects.repository.validatingAccess")}</span>
        </div>
      )}

      {accessValidated && !credentialValidating && (
        <div className="text-sm text-green-600">
          ✓ {t("projects.repository.accessValidated")}
        </div>
      )}

      {accessError && !credentialValidating && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            {accessError}
          </AlertDescription>
        </Alert>
      )}
    </div>
  );
}
