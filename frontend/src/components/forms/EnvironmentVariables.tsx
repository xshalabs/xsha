import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Settings, Plus, Trash2, ChevronDown, ChevronRight } from "lucide-react";
import * as Collapsible from "@radix-ui/react-collapsible";
import type { EnvVar } from "@/hooks/useEnvironmentForm";

interface EnvironmentVariablesProps {
  envVars: EnvVar[];
  onAddEnvVar: (key: string, value: string) => boolean;
  onRemoveEnvVar: (id: string) => void;
  onUpdateEnvVar: (id: string, field: 'key' | 'value', newValue: string) => boolean;
  disabled?: boolean;
}

export function EnvironmentVariables({
  envVars,
  onAddEnvVar,
  onRemoveEnvVar,
  onUpdateEnvVar,
  disabled = false,
}: EnvironmentVariablesProps) {
  const { t } = useTranslation();
  const [newEnvKey, setNewEnvKey] = useState("");
  const [newEnvValue, setNewEnvValue] = useState("");
  const [isOpen, setIsOpen] = useState(false);

  const handleAddEnvVar = () => {
    const success = onAddEnvVar(newEnvKey, newEnvValue);
    if (success) {
      setNewEnvKey("");
      setNewEnvValue("");
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && newEnvKey.trim()) {
      e.preventDefault();
      handleAddEnvVar();
    }
  };

  return (
    <Collapsible.Root open={isOpen} onOpenChange={setIsOpen}>
      <div className="space-y-4">
        <Collapsible.Trigger asChild>
          <button
            type="button"
            className="flex items-center gap-2 w-full text-left hover:opacity-70 transition-opacity"
          >
            {isOpen ? (
              <ChevronDown className="h-4 w-4 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            )}
            <Settings className="h-4 w-4 text-muted-foreground" />
            <Label className="text-sm font-medium cursor-pointer">
              {t("devEnvironments.env_vars.title")}
            </Label>
            <span className="text-sm text-muted-foreground">
              ({envVars.length} {t("devEnvironments.env_vars.count", "variables")})
            </span>
          </button>
        </Collapsible.Trigger>

        <Collapsible.Content className="space-y-3">
          {/* Existing environment variables */}
          <div className="space-y-3">
            {envVars.length === 0 && (
              <p className="text-sm text-muted-foreground">
                {t("devEnvironments.env_vars.empty_message")}
              </p>
            )}
            {envVars.map(({ id, key, value }) => (
              <div key={id} className="grid gap-2 grid-cols-5">
                <Input
                  placeholder={t("devEnvironments.env_vars.key")}
                  className="col-span-2"
                  value={key}
                  onChange={(e) => onUpdateEnvVar(id, 'key', e.target.value)}
                  disabled={disabled}
                />
                <Input
                  placeholder={t("devEnvironments.env_vars.value")}
                  className="col-span-2"
                  value={value}
                  onChange={(e) => onUpdateEnvVar(id, 'value', e.target.value)}
                  disabled={disabled}
                />
                <Button
                  type="button"
                  size="icon"
                  variant="ghost"
                  onClick={() => onRemoveEnvVar(id)}
                  disabled={disabled}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>

          {/* Add new environment variable */}
          <div className="grid gap-2 grid-cols-5">
            <Input
              placeholder={t("devEnvironments.env_vars.key")}
              className="col-span-2"
              value={newEnvKey}
              onChange={(e) => setNewEnvKey(e.target.value)}
              onKeyDown={handleKeyDown}
              disabled={disabled}
            />
            <Input
              placeholder={t("devEnvironments.env_vars.value")}
              className="col-span-2"
              value={newEnvValue}
              onChange={(e) => setNewEnvValue(e.target.value)}
              onKeyDown={handleKeyDown}
              disabled={disabled}
            />
            <Button
              type="button"
              size="icon"
              variant="ghost"
              onClick={handleAddEnvVar}
              disabled={!newEnvKey.trim() || disabled}
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
              onClick={handleAddEnvVar}
              disabled={!newEnvKey.trim() || disabled}
            >
              <Plus className="h-4 w-4 mr-2" />
              {t("devEnvironments.env_vars.add")}
            </Button>
          </div>
        </Collapsible.Content>
      </div>
    </Collapsible.Root>
  );
}
