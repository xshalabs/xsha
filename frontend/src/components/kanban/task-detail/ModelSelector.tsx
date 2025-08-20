import { memo, useState, useCallback, useRef, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { X, Zap, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface ModelSelectorProps {
  model: string;
  disabled?: boolean;
  onChange: (model: string) => void;
  onCloseOtherControls: () => void;
}

export const ModelSelector = memo<ModelSelectorProps>(
  ({ model, disabled, onChange, onCloseOtherControls }) => {
    const { t } = useTranslation();
    const [isOpen, setIsOpen] = useState(false);
    const modelSelectorRef = useRef<HTMLDivElement>(null);

    const handleToggle = useCallback(() => {
      onCloseOtherControls(); // Close other controls first
      setIsOpen(!isOpen);
    }, [isOpen, onCloseOtherControls]);

    const handleModelChange = useCallback((newModel: string) => {
      onChange(newModel);
      setIsOpen(false); // Close after selection
    }, [onChange]);

    const handleClose = useCallback(() => {
      setIsOpen(false);
    }, []);

    // Handle click outside to close
    useEffect(() => {
      const handleClickOutside = (event: MouseEvent) => {
        const target = event.target as Element;
        
        // Only close if clicking completely outside our components and not on any portal/popup content
        const isClickOnPortal = target.closest('[data-radix-popper-content-wrapper], [data-radix-portal], [data-sonner-toaster]');
        const isClickOnModelSelector = modelSelectorRef.current?.contains(target as Node);
        
        if (!isClickOnPortal && !isClickOnModelSelector) {
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
      <div className="relative" ref={modelSelectorRef}>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={handleToggle}
          className={`h-7 w-7 p-0 rounded-md transition-colors ${
            model && model !== 'default'
              ? 'bg-purple-100 text-purple-600 hover:bg-purple-200 dark:bg-purple-900/50 dark:text-purple-400'
              : 'text-muted-foreground hover:text-foreground hover:bg-muted'
          }`}
          title={model ? t("taskConversations.selectModel") + ": " + model : t("taskConversations.selectModel")}
        >
          {model && model !== 'default' ? <Sparkles className="h-3.5 w-3.5" /> : <Zap className="h-3.5 w-3.5" />}
        </Button>
        
        {isOpen && (
          <div className="absolute bottom-full left-0 mb-2 p-3 bg-background border rounded-lg shadow-lg z-10 min-w-[200px]">
            <div className="flex items-center justify-between mb-2">
              <Label className="text-xs font-medium">{t("taskConversations.selectModel")}</Label>
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
              <Select
                value={model}
                onValueChange={handleModelChange}
                disabled={disabled}
              >
                <SelectTrigger className="h-8 text-xs">
                  <SelectValue placeholder={t("taskConversations.selectModel")} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="default">
                    <div className="flex flex-col items-start">
                      <span className="font-medium text-xs">{t("taskConversations.model.default")}</span>
                      <span className="text-xs text-muted-foreground">
                        {t("taskConversations.model.defaultDescription")}
                      </span>
                    </div>
                  </SelectItem>
                  <SelectItem value="sonnet">
                    <div className="flex flex-col items-start">
                      <span className="font-medium text-xs">{t("taskConversations.model.sonnet")}</span>
                      <span className="text-xs text-muted-foreground">
                        Sonnet
                      </span>
                    </div>
                  </SelectItem>
                  <SelectItem value="opus">
                    <div className="flex flex-col items-start">
                      <span className="font-medium text-xs">{t("taskConversations.model.opus")}</span>
                      <span className="text-xs text-muted-foreground">
                        Opus
                      </span>
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                {t("taskConversations.modelHint")}
              </p>
            </div>
          </div>
        )}
      </div>
    );
  }
);

ModelSelector.displayName = "ModelSelector";
