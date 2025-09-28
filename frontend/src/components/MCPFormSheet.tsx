import React, { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import {
  FormSheet,
  FormSheetContent,
  FormSheetHeader,
  FormSheetTitle,
  FormSheetDescription,
  FormSheetFooter,
  FormCardGroup,
} from "@/components/forms/form-sheet";
import { MCPBasicFields } from "@/components/forms/MCPBasicFields";
import { MCPConfigFields } from "@/components/forms/MCPConfigFields";
import { useMCPForm } from "@/hooks/useMCPForm";
import type { MCP } from "@/types/mcp";
import { FormCard, FormCardContent } from "./forms/form-card";

interface MCPFormSheetProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  mcp?: MCP;
}

export function MCPFormSheet({
  isOpen,
  onClose,
  onSuccess,
  mcp,
}: MCPFormSheetProps) {
  const { t } = useTranslation();

  const handleSubmit = async () => {
    onSuccess();
    onClose();
  };

  const {
    formData,
    loading,
    error,
    errors,
    isEdit,
    handleInputChange,
    handleConfigChange,
    handleSubmit: onFormSubmit,
    resetForm,
  } = useMCPForm({ mcp, onSubmit: handleSubmit });

  // Reset form when sheet opens in create mode
  useEffect(() => {
    if (isOpen && !mcp) {
      // Reset form when opening in create mode
      resetForm();
    }
  }, [isOpen, mcp, resetForm]);

  const handleFormSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await onFormSubmit();
  };

  return (
    <FormSheet open={isOpen} onOpenChange={onClose}>
      <FormSheetContent className="w-[600px] sm:max-w-[600px]">
        <FormSheetHeader>
          <FormSheetTitle>
            {isEdit
              ? t("mcp.form.edit.title")
              : t("mcp.form.create.title")}
          </FormSheetTitle>
          <FormSheetDescription>
            {isEdit
              ? t("mcp.form.edit.description")
              : t("mcp.form.create.description")}
          </FormSheetDescription>
        </FormSheetHeader>

        <FormCardGroup className="overflow-y-auto">
          <FormCard className="border-none overflow-auto">
            <FormCardContent>
              <form onSubmit={handleFormSubmit} className="my-4 space-y-6">
                {error && (
                  <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                )}

                <div className="space-y-6">
                  <MCPBasicFields
                    formData={formData}
                    errors={errors}
                    onInputChange={handleInputChange}
                  />

                  <MCPConfigFields
                    config={formData.config}
                    error={errors.config}
                    onConfigChange={handleConfigChange}
                  />
                </div>
              </form>
            </FormCardContent>
          </FormCard>
        </FormCardGroup>

        <FormSheetFooter>
          <Button type="submit" disabled={loading} onClick={handleFormSubmit}>
            {loading
              ? t("mcp.form.submitting")
              : t("mcp.form.submit")}
          </Button>
        </FormSheetFooter>
      </FormSheetContent>
    </FormSheet>
  );
}