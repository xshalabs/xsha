import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { AlertCircle, Settings } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { NotifierFormData, NotifierTypeInfo } from "@/types/notifier";

interface NotifierConfigFieldsProps {
  formData: NotifierFormData;
  selectedTypeInfo?: NotifierTypeInfo;
  errors: Record<string, string>;
  disabled?: boolean;
  onConfigChange: (field: string, value: string | Record<string, unknown>) => void;
  getFieldLabel: (fieldName: string) => string;
}

export function NotifierConfigFields({
  formData,
  selectedTypeInfo,
  errors,
  disabled = false,
  onConfigChange,
  getFieldLabel,
}: NotifierConfigFieldsProps) {
  const { t } = useTranslation();

  if (!selectedTypeInfo) {
    return null;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Settings className="h-4 w-4 text-muted-foreground" />
        <h4 className="text-sm font-medium">{t("notifiers.form.fields.configuration")}</h4>
      </div>

      <div className="space-y-6">
        {selectedTypeInfo.config_schema.map((fieldInfo) => {
          // Add defensive check for undefined fieldInfo
          if (!fieldInfo) {
            console.warn(`Field info missing`);
            return null;
          }

          const fieldName = fieldInfo.name;
          const value = formData.config?.[fieldName] || "";
          const stringValue = typeof value === "string" ? value : String(value);
          const errorKey = `config.${fieldName}`;

          if (fieldName === "headers" || fieldName === "body_template") {
            return (
              <div key={fieldName} className="space-y-2">
                <Label htmlFor={fieldName} className="text-sm font-medium">
                  {getFieldLabel(fieldName)}
                  {fieldInfo.required && <span className="text-red-500"> *</span>}
                  {!fieldInfo.required && (
                    <span className="text-muted-foreground"> ({t("common.optional")})</span>
                  )}
                </Label>
                <Textarea
                  id={fieldName}
                  placeholder={fieldInfo.description}
                  value={typeof value === "object" ? JSON.stringify(value, null, 2) : stringValue}
                  onChange={(e) => {
                    let parsedValue: string | Record<string, unknown> = e.target.value;
                    if (fieldName === "headers" && parsedValue) {
                      try {
                        parsedValue = JSON.parse(parsedValue);
                      } catch {
                        // Keep as string if invalid JSON
                      }
                    }
                    onConfigChange(fieldName, parsedValue);
                  }}
                  disabled={disabled}
                  rows={fieldName === "body_template" ? 4 : 3}
                  className={errors[errorKey] ? "border-red-500 focus-visible:ring-red-500" : ""}
                />
                {errors[errorKey] && (
                  <p className="text-sm text-red-500 flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {errors[errorKey]}
                  </p>
                )}
                {fieldInfo.description && (
                  <p className="text-xs text-muted-foreground">{fieldInfo.description}</p>
                )}
              </div>
            );
          }

          return (
            <div key={fieldName} className="space-y-2">
              <Label htmlFor={fieldName} className="text-sm font-medium">
                {getFieldLabel(fieldName)}
                {fieldInfo.required && <span className="text-red-500"> *</span>}
                {!fieldInfo.required && (
                  <span className="text-muted-foreground"> ({t("common.optional")})</span>
                )}
              </Label>
              <Input
                id={fieldName}
                type={fieldName.includes("secret") ? "password" : "text"}
                placeholder={fieldInfo.default || fieldInfo.description}
                value={stringValue}
                onChange={(e) => onConfigChange(fieldName, e.target.value)}
                disabled={disabled}
                className={errors[errorKey] ? "border-red-500 focus-visible:ring-red-500" : ""}
              />
              {errors[errorKey] && (
                <p className="text-sm text-red-500 flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors[errorKey]}
                </p>
              )}
              {fieldInfo.description && (
                <p className="text-xs text-muted-foreground">{fieldInfo.description}</p>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}