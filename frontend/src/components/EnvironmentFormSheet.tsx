import { useTranslation } from "react-i18next";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle, Shield } from "lucide-react";
import { useEnvironmentForm } from "@/hooks/useEnvironmentForm";
import { BasicFormFields } from "@/components/forms/BasicFormFields";
import { DockerImageSelector } from "@/components/forms/DockerImageSelector";
import { ResourceLimits } from "@/components/forms/ResourceLimits";
import { EnvironmentVariables } from "@/components/forms/EnvironmentVariables";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { DevEnvironmentDisplay } from "@/types/dev-environment";

interface EnvironmentFormSheetProps {
  environment?: DevEnvironmentDisplay;
  onSubmit: (environment: DevEnvironmentDisplay) => Promise<void>;
  onCancel?: () => void;
  formId?: string;
}

export function EnvironmentFormSheet({
  environment,
  onSubmit,
  onCancel: _onCancel,
  formId = "environment-form-sheet",
}: EnvironmentFormSheetProps) {
  const { t } = useTranslation();

  // Use custom hook for all form logic
  const {
    formData,
    envVars,
    environmentImages,
    providers,
    loading,
    loadingImages,
    loadingProviders,
    error,
    errors,
    isEdit: _isEdit,
    handleInputChange,
    handleDockerImageChange,
    handleSubmit,
    addEnvVar,
    removeEnvVar,
    updateEnvVar,
  } = useEnvironmentForm(environment, onSubmit);

  return (
    <form id={formId} onSubmit={handleSubmit} className="my-4 space-y-6">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="space-y-6">
        {/* Basic Form Fields */}
        <BasicFormFields
          name={formData.name}
          description={formData.description}
          systemPrompt={formData.system_prompt}
          onNameChange={(value) => handleInputChange("name", value)}
          onDescriptionChange={(value) => handleInputChange("description", value)}
          onSystemPromptChange={(value) => handleInputChange("system_prompt", value)}
          errors={{ name: errors.name }}
          disabled={loading}
        />

        {/* Docker Image Selection */}
        <DockerImageSelector
          dockerImage={formData.docker_image}
          environmentImages={environmentImages}
          onDockerImageChange={handleDockerImageChange}
          loadingImages={loadingImages}
          error={errors.docker_image}
          disabled={loading}
        />

        {/* Provider Selection (Optional) */}
        <div className="space-y-2">
          <Label htmlFor="provider-select">
            {t("devEnvironments.form.provider", "Provider")} ({t("common.optional", "Optional")})
          </Label>
          <Select
            value={formData.provider_id?.toString() || "none"}
            onValueChange={(value) =>
              handleInputChange("provider_id", value === "none" ? undefined : parseInt(value))
            }
            disabled={loading || loadingProviders}
          >
            <SelectTrigger id="provider-select">
              <SelectValue placeholder={t("devEnvironments.form.provider_placeholder", "Select a provider")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="none">{t("devEnvironments.form.no_provider", "No Provider")}</SelectItem>
              {providers.map((provider) => (
                <SelectItem key={provider.id} value={provider.id.toString()}>
                  {provider.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="text-sm text-muted-foreground">
            {t("devEnvironments.form.provider_help", "Optional: Select a provider configuration for this environment")}
          </p>
        </div>

        {/* Resource Limits */}
        <ResourceLimits
          cpuLimit={formData.cpu_limit}
          memoryLimit={formData.memory_limit}
          onCpuLimitChange={(value) => handleInputChange("cpu_limit", value)}
          onMemoryLimitChange={(value) => handleInputChange("memory_limit", value)}
          errors={{
            cpu_limit: errors.cpu_limit,
            memory_limit: errors.memory_limit,
          }}
          disabled={loading}
        />

        {/* Environment Variables */}
        <EnvironmentVariables
          envVars={envVars}
          onAddEnvVar={addEnvVar}
          onRemoveEnvVar={removeEnvVar}
          onUpdateEnvVar={updateEnvVar}
          disabled={loading}
        />

        {/* Configuration Help */}
        <Alert>
          <Shield className="h-4 w-4" />
          <AlertDescription className="text-xs">
            {t("devEnvironments.configurationHelp")}
          </AlertDescription>
        </Alert>
      </div>
    </form>
  );
}
