import { useTranslation } from "react-i18next";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, AlertCircle, Zap } from "lucide-react";
import type { DevEnvironment } from "@/types/dev-environment";

interface TaskFormDevEnvironmentProps {
  devEnvironmentId?: number;
  onDevEnvironmentChange: (id: number | undefined) => void;
  devEnvironments: DevEnvironment[];
  loading: boolean;
  error?: string;
  validationError?: string;
  disabled?: boolean;
}

export function TaskFormDevEnvironment({
  devEnvironmentId,
  onDevEnvironmentChange,
  devEnvironments,
  loading,
  error,
  validationError,
  disabled,
}: TaskFormDevEnvironmentProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <Zap className="h-4 w-4 text-muted-foreground" />
        <Label htmlFor="dev_environment" className="text-sm font-medium">
          {t("tasks.fields.devEnvironment")} <span className="text-red-500">*</span>
        </Label>
      </div>
      <Select
        value={devEnvironmentId?.toString() || ""}
        onValueChange={(value) =>
          onDevEnvironmentChange(value ? parseInt(value) : undefined)
        }
        disabled={loading || disabled}
      >
        <SelectTrigger 
          className={validationError ? "border-red-500 focus:ring-red-500" : ""}
        >
          <SelectValue
            placeholder={
              loading 
                ? t("common.loading") + "..."
                : t("tasks.form.selectDevEnvironment")
            }
          />
        </SelectTrigger>
        <SelectContent>
          {loading ? (
            <SelectItem value="loading" disabled>
              <div className="flex items-center gap-2">
                <Loader2 className="h-3 w-3 animate-spin" />
                {t("common.loading")}...
              </div>
            </SelectItem>
          ) : devEnvironments.length === 0 ? (
            <SelectItem value="empty" disabled>
              {t("tasks.form.noDevEnvironmentsAvailable")}
            </SelectItem>
          ) : (
            devEnvironments.map((env) => (
              <SelectItem key={env.id} value={env.id.toString()}>
                <div className="flex items-center justify-between w-full">
                  <span className="font-medium">{env.name}</span>
                  <span className="text-xs text-muted-foreground ml-2">
                    {env.type}
                  </span>
                </div>
              </SelectItem>
            ))
          )}
        </SelectContent>
      </Select>
      {validationError && (
        <p className="text-sm text-red-500 flex items-center gap-1">
          <AlertCircle className="h-3 w-3" />
          {validationError}
        </p>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      <p className="text-xs text-muted-foreground">
        {t("tasks.form.devEnvironmentHint")}
      </p>
    </div>
  );
}