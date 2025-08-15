import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { AlertCircle, FileText, Settings } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { ProjectFormData } from "@/types/project";

interface ProjectBasicFieldsProps {
  formData: ProjectFormData;
  errors: Record<string, string>;
  disabled?: boolean;
  onChange: (field: keyof ProjectFormData, value: string) => void;
}

export function ProjectBasicFields({
  formData,
  errors,
  disabled = false,
  onChange,
}: ProjectBasicFieldsProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      {/* Project Name */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <FileText className="h-4 w-4 text-muted-foreground" />
          <Label htmlFor="name" className="text-sm font-medium">
            {t("projects.name")} <span className="text-red-500">*</span>
          </Label>
        </div>
        <Input
          id="name"
          type="text"
          value={formData.name}
          onChange={(e) => onChange("name", e.target.value)}
          placeholder={t("projects.placeholders.name")}
          className={errors.name ? "border-red-500 focus-visible:ring-red-500" : ""}
          disabled={disabled}
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
            {t("projects.description")}
          </Label>
        </div>
        <Input
          id="description"
          type="text"
          value={formData.description}
          onChange={(e) => onChange("description", e.target.value)}
          placeholder={t("projects.placeholders.description")}
          disabled={disabled}
        />
      </div>

      {/* System Prompt */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Settings className="h-4 w-4 text-muted-foreground" />
          <Label htmlFor="system_prompt" className="text-sm font-medium">
            {t("projects.systemPrompt")}
          </Label>
        </div>
        <Textarea
          id="system_prompt"
          value={formData.system_prompt}
          onChange={(e) => onChange("system_prompt", e.target.value)}
          placeholder={t("projects.placeholders.systemPrompt")}
          rows={3}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          {t("projects.systemPromptHelp")}
        </p>
      </div>
    </div>
  );
}
