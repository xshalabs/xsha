import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { AlertCircle, FileText, Settings, Loader2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { NotifierFormData, NotifierTypeInfo, NotifierType } from "@/types/notifier";

interface NotifierBasicFieldsProps {
  formData: NotifierFormData;
  notifierTypes: NotifierTypeInfo[];
  errors: Record<string, string>;
  disabled?: boolean;
  typesLoading?: boolean;
  isEdit?: boolean;
  onChange: (field: keyof NotifierFormData, value: string | NotifierType) => void;
}

export function NotifierBasicFields({
  formData,
  notifierTypes,
  errors,
  disabled = false,
  typesLoading = false,
  isEdit = false,
  onChange,
}: NotifierBasicFieldsProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      {/* Notifier Type */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Settings className="h-4 w-4 text-muted-foreground" />
          <Label htmlFor="type" className="text-sm font-medium">
            {t("notifiers.form.fields.type.label")} <span className="text-red-500">*</span>
          </Label>
        </div>
        <Select
          value={formData.type}
          onValueChange={(value) => {
            onChange("type", value as NotifierType);
            // Reset config when type changes (only if not editing)
            if (!isEdit) {
              onChange("config", "" as any);
            }
          }}
          disabled={isEdit || disabled || typesLoading}
        >
          <SelectTrigger className={errors.type ? "border-red-500 focus-visible:ring-red-500" : ""}>
            {typesLoading ? (
              <div className="flex items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>{t("common.loading")}</span>
              </div>
            ) : (
              <SelectValue placeholder={t("notifiers.form.fields.type.placeholder")} />
            )}
          </SelectTrigger>
          <SelectContent>
            {notifierTypes.map((type) => (
              <SelectItem key={type.type} value={type.type}>
                {type.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {errors.type && (
          <p className="text-sm text-red-500 flex items-center gap-1">
            <AlertCircle className="h-3 w-3" />
            {errors.type}
          </p>
        )}
        {isEdit && (
          <p className="text-xs text-muted-foreground">
            {t("notifiers.form.fields.type.cannotModify")}
          </p>
        )}
      </div>

      {/* Notifier Name */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <FileText className="h-4 w-4 text-muted-foreground" />
          <Label htmlFor="name" className="text-sm font-medium">
            {t("notifiers.form.fields.name.label")} <span className="text-red-500">*</span>
          </Label>
        </div>
        <Input
          id="name"
          type="text"
          value={formData.name}
          onChange={(e) => onChange("name", e.target.value)}
          placeholder={t("notifiers.form.fields.name.placeholder")}
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
            {t("notifiers.form.fields.description.label")}
            <span className="text-muted-foreground"> ({t("common.optional")})</span>
          </Label>
        </div>
        <Textarea
          id="description"
          value={formData.description}
          onChange={(e) => onChange("description", e.target.value)}
          placeholder={t("notifiers.form.fields.description.placeholder")}
          rows={3}
          disabled={disabled}
        />
      </div>
    </div>
  );
}