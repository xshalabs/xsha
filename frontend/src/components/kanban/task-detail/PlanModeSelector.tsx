import { memo, useState, useCallback, useRef, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { X, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

interface PlanModeSelectorProps {
  isPlanMode?: boolean;
  disabled?: boolean;
  onChange: (isPlanMode: boolean) => void;
  onModelChange?: (model: string) => void;
  onCloseOtherControls: () => void;
}

export const PlanModeSelector = memo<PlanModeSelectorProps>(
  ({ isPlanMode, disabled, onChange, onModelChange, onCloseOtherControls }) => {
    const { t } = useTranslation();
    const [isOpen, setIsOpen] = useState(false);
    const planModeSelectorRef = useRef<HTMLDivElement>(null);

    const handleToggle = useCallback(() => {
      onCloseOtherControls(); // Close other controls first
      setIsOpen(!isOpen);
    }, [isOpen, onCloseOtherControls]);

    const handlePlanModeChange = useCallback((checked: boolean) => {
      onChange(checked);
      // When enabling plan mode, automatically set model to opus
      if (checked && onModelChange) {
        onModelChange('opus');
      }
    }, [onChange, onModelChange]);

    const handleClose = useCallback(() => {
      setIsOpen(false);
    }, []);

    // Handle click outside to close
    useEffect(() => {
      const handleClickOutside = (event: MouseEvent) => {
        const target = event.target as Element;
        
        // Only close if clicking completely outside our components and not on any portal/popup content
        const isClickOnPortal = target.closest('[data-radix-popper-content-wrapper], [data-radix-portal], [data-sonner-toaster]');
        const isClickOnPlanModeSelector = planModeSelectorRef.current?.contains(target as Node);
        
        if (!isClickOnPortal && !isClickOnPlanModeSelector) {
          setIsOpen(false);
        }
      };

      if (isOpen) {
        // Use a timeout to avoid immediate closure
        const timeoutId = setTimeout(() => {
          document.addEventListener('mousedown', handleClickOutside);
        }, 100);

        return () => {
          clearTimeout(timeoutId);
          document.removeEventListener('mousedown', handleClickOutside);
        };
      }
    }, [isOpen]);

    return (
      <div className="relative" ref={planModeSelectorRef}>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={handleToggle}
          className={`h-7 w-7 p-0 rounded-md transition-colors ${
            isPlanMode
              ? 'bg-orange-100 text-orange-600 hover:bg-orange-200 dark:bg-orange-900/50 dark:text-orange-400'
              : 'text-muted-foreground hover:text-foreground hover:bg-muted'
          }`}
          title={isPlanMode ? t("taskConversations.planModeEnabled") : t("taskConversations.planMode")}
        >
          <FileText className="h-3.5 w-3.5" />
        </Button>
        
        {isOpen && (
          <div className="absolute bottom-full left-0 mb-2 p-3 bg-background border rounded-lg shadow-lg z-10 min-w-[200px]">
            <div className="flex items-center justify-between mb-2">
              <Label className="text-xs font-medium">{t("taskConversations.planMode")}</Label>
              <Button
                variant="ghost"
                size="sm"
                onClick={handleClose}
                className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground"
              >
                <X className="h-3 w-3" />
              </Button>
            </div>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="plan-mode-switch" className="text-xs">
                  {t("taskConversations.enablePlanMode")}
                </Label>
                <Switch
                  id="plan-mode-switch"
                  checked={isPlanMode || false}
                  onCheckedChange={handlePlanModeChange}
                  disabled={disabled}
                />
              </div>
              <p className="text-xs text-muted-foreground">
                {t("taskConversations.planModeHint")}
              </p>
            </div>
          </div>
        )}
      </div>
    );
  }
);

PlanModeSelector.displayName = "PlanModeSelector";