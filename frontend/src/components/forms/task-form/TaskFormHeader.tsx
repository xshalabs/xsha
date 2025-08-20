import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { FileText, AlertCircle } from "lucide-react";

interface TaskFormHeaderProps {
  title: string;
  onTitleChange: (title: string) => void;
  error?: string;
  disabled?: boolean;
}

export function TaskFormHeader({ title, onTitleChange, error, disabled }: TaskFormHeaderProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <FileText className="h-4 w-4 text-muted-foreground" />
        <Label htmlFor="title" className="text-sm font-medium">
          {t("tasks.fields.title")} <span className="text-red-500">*</span>
        </Label>
      </div>
      <Input
        id="title"
        type="text"
        value={title}
        onChange={(e) => onTitleChange(e.target.value)}
        placeholder={t("tasks.form.titlePlaceholder")}
        className={error ? "border-red-500 focus-visible:ring-red-500" : ""}
        disabled={disabled}
      />
      {error && (
        <p className="text-sm text-red-500 flex items-center gap-1">
          <AlertCircle className="h-3 w-3" />
          {error}
        </p>
      )}
    </div>
  );
}