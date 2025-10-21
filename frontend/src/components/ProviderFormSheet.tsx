import { useTranslation } from "react-i18next";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import { useProviderForm } from "@/hooks/useProviderForm";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ProviderConfig } from "@/components/forms/ProviderConfig";
import type { Provider } from "@/types/provider";

interface ProviderFormSheetProps {
  provider?: Provider;
  onSubmit: (provider: Provider) => Promise<void>;
  onCancel?: () => void;
  formId?: string;
}

export function ProviderFormSheet({
  provider,
  onSubmit,
  onCancel: _onCancel,
  formId = "provider-form-sheet",
}: ProviderFormSheetProps) {
  const { t } = useTranslation();

  // Use custom hook for all form logic
  const {
    formData,
    configVars,
    providerTypes,
    loading,
    loadingTypes,
    error,
    errors,
    isEdit,
    handleInputChange,
    handleSubmit,
    addConfigVar,
    removeConfigVar,
    updateConfigVar,
  } = useProviderForm(provider, onSubmit);

  return (
    <form id={formId} onSubmit={handleSubmit} className="my-4 space-y-6">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="space-y-6">
        {/* Name */}
        <div className="space-y-2">
          <Label htmlFor="provider-name">
            {t("provider.form.name")} <span className="text-destructive">*</span>
          </Label>
          <Input
            id="provider-name"
            type="text"
            placeholder={t("provider.form.name_placeholder")}
            value={formData.name}
            onChange={(e) => handleInputChange("name", e.target.value)}
            disabled={loading}
            className={errors.name ? "border-destructive" : ""}
          />
          {errors.name && (
            <p className="text-sm text-destructive">{errors.name}</p>
          )}
        </div>

        {/* Description */}
        <div className="space-y-2">
          <Label htmlFor="provider-description">
            {t("provider.form.description")}
          </Label>
          <Textarea
            id="provider-description"
            placeholder={t("provider.form.description_placeholder")}
            value={formData.description}
            onChange={(e) => handleInputChange("description", e.target.value)}
            disabled={loading}
            rows={3}
          />
        </div>

        {/* Type - Only show in create mode */}
        {!isEdit && (
          <div className="space-y-2">
            <Label htmlFor="provider-type">
              {t("provider.form.type")} <span className="text-destructive">*</span>
            </Label>
            <Select
              value={formData.type}
              onValueChange={(value) => handleInputChange("type", value)}
              disabled={loading || loadingTypes}
            >
              <SelectTrigger id="provider-type" className={errors.type ? "border-destructive" : ""}>
                <SelectValue placeholder={t("provider.form.type_placeholder")} />
              </SelectTrigger>
              <SelectContent>
                {providerTypes.map((type) => (
                  <SelectItem key={type} value={type}>
                    {t(`provider.types.${type}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.type && (
              <p className="text-sm text-destructive">{errors.type}</p>
            )}
            {isEdit && (
              <p className="text-sm text-muted-foreground">
                {t("provider.type_cannot_be_modified")}
              </p>
            )}
          </div>
        )}

        {/* Type - Display only in edit mode */}
        {isEdit && (
          <div className="space-y-2">
            <Label htmlFor="provider-type-display">
              {t("provider.form.type")}
            </Label>
            <Input
              id="provider-type-display"
              type="text"
              value={t(`provider.types.${formData.type}`)}
              disabled
              className="bg-muted"
            />
            <p className="text-sm text-muted-foreground">
              {t("provider.type_cannot_be_modified")}
            </p>
          </div>
        )}

        {/* Config */}
        <ProviderConfig
          configVars={configVars}
          onAddConfigVar={addConfigVar}
          onRemoveConfigVar={removeConfigVar}
          onUpdateConfigVar={updateConfigVar}
          disabled={loading}
        />
        {errors.config && (
          <p className="text-sm text-destructive">{errors.config}</p>
        )}
      </div>
    </form>
  );
}
