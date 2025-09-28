import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { AlertCircle, FileText, Settings2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { MCPFormData } from "@/types/mcp";

interface MCPBasicFieldsProps {
  formData: MCPFormData;
  errors: Record<string, string>;
  disabled?: boolean;
  onInputChange: (field: keyof MCPFormData, value: string) => void;
}

export function MCPBasicFields({
  formData,
  errors,
  disabled = false,
  onInputChange,
}: MCPBasicFieldsProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      {/* MCP Name */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Settings2 className="h-4 w-4 text-muted-foreground" />
          <Label htmlFor="name" className="text-sm font-medium">
            {t("mcp.form.fields.name.label")} <span className="text-red-500">*</span>
          </Label>
        </div>
        <Input
          id="name"
          type="text"
          value={formData.name}
          onChange={(e) => onInputChange("name", e.target.value)}
          placeholder={t("mcp.form.fields.name.placeholder")}
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
            {t("mcp.form.fields.description.label")}
            <span className="text-muted-foreground"> ({t("common.optional")})</span>
          </Label>
        </div>
        <Textarea
          id="description"
          value={formData.description}
          onChange={(e) => onInputChange("description", e.target.value)}
          placeholder={t("mcp.form.fields.description.placeholder")}
          rows={3}
          disabled={disabled}
        />
      </div>

    </div>
  );
}