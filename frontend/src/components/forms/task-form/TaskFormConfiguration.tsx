import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, AlertCircle, GitBranch, RefreshCw } from "lucide-react";

interface TaskFormConfigurationProps {
  startBranch: string;
  onStartBranchChange: (branch: string) => void;
  availableBranches: string[];
  fetchingBranches: boolean;
  branchError?: string;
  onRefreshBranches: () => void;
  validationError?: string;
  disabled?: boolean;
}

export function TaskFormConfiguration({
  startBranch,
  onStartBranchChange,
  availableBranches,
  fetchingBranches,
  branchError,
  onRefreshBranches,
  validationError,
  disabled,
}: TaskFormConfigurationProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <GitBranch className="h-4 w-4 text-muted-foreground" />
        <Label htmlFor="start_branch" className="text-sm font-medium">
          {t("tasks.fields.startBranch")} <span className="text-red-500">*</span>
        </Label>
        {fetchingBranches && (
          <Loader2 className="h-3 w-3 animate-spin text-blue-500" />
        )}
      </div>
      <Select
        value={startBranch}
        onValueChange={onStartBranchChange}
        disabled={fetchingBranches || disabled}
      >
        <SelectTrigger
          className={validationError ? "border-red-500 focus:ring-red-500" : ""}
        >
          <SelectValue 
            placeholder={
              fetchingBranches 
                ? t("tasks.form.fetchingBranches") + "..."
                : t("tasks.form.selectBranch")
            } 
          />
        </SelectTrigger>
        <SelectContent>
          {fetchingBranches ? (
            <SelectItem value="loading" disabled>
              <div className="flex items-center gap-2">
                <Loader2 className="h-3 w-3 animate-spin" />
                {t("tasks.form.fetchingBranches")}...
              </div>
            </SelectItem>
          ) : availableBranches.length === 0 ? (
            <SelectItem value="empty" disabled>
              {branchError || t("tasks.form.noBranchesAvailable")}
            </SelectItem>
          ) : (
            availableBranches.map((branch) => (
              <SelectItem key={branch} value={branch}>
                <div className="flex items-center gap-2">
                  <GitBranch className="h-3 w-3" />
                  {branch}
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
      {branchError && !fetchingBranches && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription className="flex items-center justify-between">
            <span>{branchError}</span>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onRefreshBranches}
              className="h-6 px-2"
            >
              <RefreshCw className="h-3 w-3" />
            </Button>
          </AlertDescription>
        </Alert>
      )}
      <p className="text-xs text-muted-foreground">
        {t("tasks.form.branchFromRepository")}
      </p>
    </div>
  );
}