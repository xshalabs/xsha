import { memo, useState, useCallback, useRef, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { X, Clock, Calendar } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { DateTimePicker } from "@/components/ui/datetime-picker";

interface ExecutionTimeControlProps {
  executionTime?: Date;
  onChange: (time: Date | undefined) => void;
  onCloseOtherControls: () => void;
}

export const ExecutionTimeControl = memo<ExecutionTimeControlProps>(
  ({ executionTime, onChange, onCloseOtherControls }) => {
    const { t } = useTranslation();
    const [isOpen, setIsOpen] = useState(false);
    const timePickerRef = useRef<HTMLDivElement>(null);

    const handleToggle = useCallback(() => {
      onCloseOtherControls(); // Close other controls first
      setIsOpen(!isOpen);
    }, [isOpen, onCloseOtherControls]);

    const handleTimeChange = useCallback((time: Date | undefined) => {
      onChange(time);
      // Don't auto-close to allow multiple time adjustments
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
        const isClickOnTimePicker = timePickerRef.current?.contains(target as Node);
        
        if (!isClickOnPortal && !isClickOnTimePicker) {
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
      <div className="relative" ref={timePickerRef}>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={handleToggle}
          className={`h-7 w-7 p-0 rounded-md transition-colors ${
            executionTime 
              ? 'bg-blue-100 text-blue-600 hover:bg-blue-200 dark:bg-blue-900/50 dark:text-blue-400' 
              : 'text-muted-foreground hover:text-foreground hover:bg-muted'
          }`}
          title={executionTime ? t("taskConversations.executionTime") + ": " + executionTime.toLocaleString() : t("taskConversations.executionTime")}
        >
          {executionTime ? <Calendar className="h-3.5 w-3.5" /> : <Clock className="h-3.5 w-3.5" />}
        </Button>
        
        {isOpen && (
          <div className="absolute bottom-full left-0 mb-2 p-3 bg-background border rounded-lg shadow-lg z-10 min-w-[200px]">
            <div className="flex items-center justify-between mb-2">
              <Label className="text-xs font-medium">{t("taskConversations.executionTime")}</Label>
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
              <DateTimePicker
                value={executionTime}
                onChange={handleTimeChange}
                placeholder={t("taskConversations.executionTimePlaceholder")}
                label=""
                className="h-8 text-xs"
              />
              <p className="text-xs text-muted-foreground">
                {t("taskConversations.executionTimeHint")}
              </p>
            </div>
          </div>
        )}
      </div>
    );
  }
);

ExecutionTimeControl.displayName = "ExecutionTimeControl";
