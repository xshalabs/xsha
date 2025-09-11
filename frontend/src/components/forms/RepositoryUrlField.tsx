import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { AlertCircle, Link2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { ProjectFormData } from "@/types/project";

interface RepositoryUrlFieldProps {
  formData: ProjectFormData;
  errors: Record<string, string>;
  disabled?: boolean;
  onChange: (field: keyof ProjectFormData, value: string) => void;
}

export function RepositoryUrlField({
  formData,
  errors,
  disabled = false,
  onChange,
}: RepositoryUrlFieldProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <Link2 className="h-4 w-4 text-muted-foreground" />
        <Label htmlFor="repo_url" className="text-sm font-medium">
          {t("projects.repoUrl")} <span className="text-red-500">*</span>
        </Label>
      </div>
      <Input
        id="repo_url"
        type="text"
        value={formData.repo_url}
        onChange={(e) => onChange("repo_url", e.target.value)}
        placeholder={t("projects.placeholders.repoUrl")}
        className={errors.repo_url ? "border-red-500 focus-visible:ring-red-500" : ""}
        disabled={disabled}
      />
      {errors.repo_url && (
        <p className="text-sm text-red-500 flex items-center gap-1">
          <AlertCircle className="h-3 w-3" />
          {errors.repo_url}
        </p>
      )}
    </div>
  );
}
