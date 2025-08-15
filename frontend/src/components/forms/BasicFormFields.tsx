import React from "react";
import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { AlertCircle, FileText, Settings } from "lucide-react";

interface BasicFormFieldsProps {
  name: string;
  description: string;
  systemPrompt: string;
  onNameChange: (value: string) => void;
  onDescriptionChange: (value: string) => void;
  onSystemPromptChange: (value: string) => void;
  errors?: {
    name?: string;
  };
  disabled?: boolean;
}

export function BasicFormFields({
  name,
  description,
  systemPrompt,
  onNameChange,
  onDescriptionChange,
  onSystemPromptChange,
  errors = {},
  disabled = false,
}: BasicFormFieldsProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      {/* Environment Name */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <FileText className="h-4 w-4 text-muted-foreground" />
          <Label htmlFor="name" className="text-sm font-medium">
            {t("devEnvironments.form.name")} <span className="text-red-500">*</span>
          </Label>
        </div>
        <Input
          id="name"
          value={name}
          onChange={(e) => onNameChange(e.target.value)}
          placeholder={t("devEnvironments.form.name_placeholder")}
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
            {t("devEnvironments.form.description")}
          </Label>
        </div>
        <Textarea
          id="description"
          value={description}
          onChange={(e) => onDescriptionChange(e.target.value)}
          placeholder={t("devEnvironments.form.description_placeholder")}
          rows={3}
          disabled={disabled}
        />
      </div>

      {/* System Prompt */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Settings className="h-4 w-4 text-muted-foreground" />
          <Label htmlFor="system_prompt" className="text-sm font-medium">
            {t("devEnvironments.form.system_prompt")}
          </Label>
        </div>
        <Textarea
          id="system_prompt"
          value={systemPrompt}
          onChange={(e) => onSystemPromptChange(e.target.value)}
          placeholder={t("devEnvironments.form.system_prompt_placeholder")}
          rows={4}
          disabled={disabled}
        />
      </div>
    </div>
  );
}
