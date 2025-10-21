import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Settings, Plus, Trash2 } from "lucide-react";
import type { ConfigVar } from "@/hooks/useProviderForm";

interface ProviderConfigProps {
  configVars: ConfigVar[];
  onAddConfigVar: (key: string, value: string) => boolean;
  onRemoveConfigVar: (id: string) => void;
  onUpdateConfigVar: (id: string, field: 'key' | 'value', newValue: string) => boolean;
  disabled?: boolean;
}

export function ProviderConfig({
  configVars,
  onAddConfigVar,
  onRemoveConfigVar,
  onUpdateConfigVar,
  disabled = false,
}: ProviderConfigProps) {
  const { t } = useTranslation();
  const [newConfigKey, setNewConfigKey] = useState("");
  const [newConfigValue, setNewConfigValue] = useState("");

  const handleAddConfigVar = () => {
    const success = onAddConfigVar(newConfigKey, newConfigValue);
    if (success) {
      setNewConfigKey("");
      setNewConfigValue("");
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && newConfigKey.trim()) {
      e.preventDefault();
      handleAddConfigVar();
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Settings className="h-4 w-4 text-muted-foreground" />
        <Label className="text-sm font-medium">
          {t("provider.config_vars.title")}
        </Label>
      </div>

      {/* Existing configuration variables */}
      <div className="space-y-3">
        {configVars.length === 0 && (
          <p className="text-sm text-muted-foreground">
            {t("provider.config_vars.empty_message")}
          </p>
        )}
        {configVars.map(({ id, key, value }) => (
          <div key={id} className="grid gap-2 grid-cols-5">
            <Input
              placeholder={t("provider.config_vars.key")}
              className="col-span-2"
              value={key}
              onChange={(e) => onUpdateConfigVar(id, 'key', e.target.value)}
              disabled={disabled}
            />
            <Input
              placeholder={t("provider.config_vars.value")}
              className="col-span-2"
              value={value}
              onChange={(e) => onUpdateConfigVar(id, 'value', e.target.value)}
              disabled={disabled}
            />
            <Button
              type="button"
              size="icon"
              variant="ghost"
              onClick={() => onRemoveConfigVar(id)}
              disabled={disabled}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ))}
      </div>

      {/* Add new configuration variable */}
      <div className="grid gap-2 grid-cols-5">
        <Input
          placeholder={t("provider.config_vars.key")}
          className="col-span-2"
          value={newConfigKey}
          onChange={(e) => setNewConfigKey(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={disabled}
        />
        <Input
          placeholder={t("provider.config_vars.value")}
          className="col-span-2"
          value={newConfigValue}
          onChange={(e) => setNewConfigValue(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={disabled}
        />
        <Button
          type="button"
          size="icon"
          variant="ghost"
          onClick={handleAddConfigVar}
          disabled={!newConfigKey.trim() || disabled}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>

      {/* Add button (alternative) */}
      <div>
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={handleAddConfigVar}
          disabled={!newConfigKey.trim() || disabled}
        >
          <Plus className="h-4 w-4 mr-2" />
          {t("provider.config_vars.add")}
        </Button>
      </div>
    </div>
  );
}
