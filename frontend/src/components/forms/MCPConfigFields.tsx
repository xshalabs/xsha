import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { AlertCircle, Settings, FileCode } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useState, useEffect } from "react";

interface MCPConfigFieldsProps {
  config: string;
  error?: string;
  disabled?: boolean;
  onConfigChange: (config: string) => void;
}

export function MCPConfigFields({
  config,
  error,
  disabled = false,
  onConfigChange,
}: MCPConfigFieldsProps) {
  const { t } = useTranslation();
  const [isValidJson, setIsValidJson] = useState(true);

  // Validate JSON on config change
  useEffect(() => {
    if (!config) {
      setIsValidJson(true);
      return;
    }

    try {
      JSON.parse(config);
      setIsValidJson(true);
    } catch {
      setIsValidJson(false);
    }
  }, [config]);

  const handleConfigChange = (value: string) => {
    onConfigChange(value);
  };

  const formatJson = () => {
    try {
      if (config) {
        const parsed = JSON.parse(config);
        const formatted = JSON.stringify(parsed, null, 2);
        onConfigChange(formatted);
      }
    } catch {
      // Do nothing if invalid JSON
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Settings className="h-4 w-4 text-muted-foreground" />
        <h4 className="text-sm font-medium">{t("mcp.form.fields.config.label")}</h4>
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label htmlFor="config" className="text-sm font-medium">
            {t("mcp.form.fields.config.label")} <span className="text-red-500">*</span>
          </Label>
          {config && isValidJson && (
            <button
              type="button"
              onClick={formatJson}
              className="text-xs text-blue-600 hover:text-blue-800 underline"
              disabled={disabled}
            >
              Format JSON
            </button>
          )}
        </div>

        <div className="relative">
          <Textarea
            id="config"
            placeholder={t("mcp.form.fields.config.placeholder")}
            value={config}
            onChange={(e) => handleConfigChange(e.target.value)}
            disabled={disabled}
            rows={10}
            className={`font-mono text-sm ${error || !isValidJson ? "border-red-500 focus-visible:ring-red-500" : ""}`}
          />
          <div className="absolute top-2 right-2">
            <FileCode className={`h-4 w-4 ${isValidJson ? "text-green-500" : "text-red-500"}`} />
          </div>
        </div>

        {!isValidJson && (
          <p className="text-sm text-red-500 flex items-center gap-1">
            <AlertCircle className="h-3 w-3" />
            {t("mcp.form.validation.invalidJson")}
          </p>
        )}

        {error && (
          <p className="text-sm text-red-500 flex items-center gap-1">
            <AlertCircle className="h-3 w-3" />
            {error}
          </p>
        )}

        <p className="text-xs text-muted-foreground">
          {t("mcp.form.fields.config.help")}
        </p>
      </div>
    </div>
  );
}