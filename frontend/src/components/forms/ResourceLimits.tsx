import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { AlertCircle, Cpu, MemoryStick } from "lucide-react";

interface ResourceLimitsProps {
  cpuLimit: number;
  memoryLimit: number;
  onCpuLimitChange: (value: number) => void;
  onMemoryLimitChange: (value: number) => void;
  errors?: {
    cpu_limit?: string;
    memory_limit?: string;
  };
  disabled?: boolean;
}

export function ResourceLimits({
  cpuLimit,
  memoryLimit,
  onCpuLimitChange,
  onMemoryLimitChange,
  errors = {},
  disabled = false,
}: ResourceLimitsProps) {
  const { t } = useTranslation();

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {/* CPU Limit */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Cpu className="h-4 w-4 text-muted-foreground" />
          <Label htmlFor="cpu_limit" className="text-sm font-medium">
            {t("devEnvironments.form.cpu_limit")} <span className="text-red-500">*</span>
            <span className="text-xs text-muted-foreground ml-1">
              (0.1-16 {t("devEnvironments.stats.cores")})
            </span>
          </Label>
        </div>
        <Input
          id="cpu_limit"
          type="number"
          step="0.1"
          min="0.1"
          max="16"
          value={cpuLimit}
          onChange={(e) => onCpuLimitChange(parseFloat(e.target.value))}
          className={errors.cpu_limit ? "border-red-500 focus-visible:ring-red-500" : ""}
          disabled={disabled}
        />
        {errors.cpu_limit && (
          <p className="text-sm text-red-500 flex items-center gap-1">
            <AlertCircle className="h-3 w-3" />
            {errors.cpu_limit}
          </p>
        )}
      </div>

      {/* Memory Limit */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <MemoryStick className="h-4 w-4 text-muted-foreground" />
          <Label htmlFor="memory_limit" className="text-sm font-medium">
            {t("devEnvironments.form.memory_limit")} <span className="text-red-500">*</span>
            <span className="text-xs text-muted-foreground ml-1">
              (128-32768 MB)
            </span>
          </Label>
        </div>
        <Input
          id="memory_limit"
          type="number"
          min="128"
          max="32768"
          value={memoryLimit}
          onChange={(e) => onMemoryLimitChange(parseInt(e.target.value))}
          className={errors.memory_limit ? "border-red-500 focus-visible:ring-red-500" : ""}
          disabled={disabled}
        />
        {errors.memory_limit && (
          <p className="text-sm text-red-500 flex items-center gap-1">
            <AlertCircle className="h-3 w-3" />
            {errors.memory_limit}
          </p>
        )}
      </div>
    </div>
  );
}
