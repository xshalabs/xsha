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
import { NotifierBasicFields } from "@/components/forms/NotifierBasicFields";
import { NotifierConfigFields } from "@/components/forms/NotifierConfigFields";
import { useNotifierForm } from "@/hooks/useNotifierForm";
import type { Notifier } from "@/types/notifier";
import { FormCard, FormCardContent } from "./forms/form-card";

interface NotifierFormSheetProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  notifier?: Notifier;
}

export function NotifierFormSheet({
  isOpen,
  onClose,
  onSuccess,
  notifier,
}: NotifierFormSheetProps) {
  const { t } = useTranslation();

  const handleSubmit = async () => {
    onSuccess();
    onClose();
  };

  const {
    formData,
    notifierTypes,
    loading,
    typesLoading,
    error,
    errors,
    isEdit,
    handleInputChange,
    handleConfigChange,
    handleSubmit: onFormSubmit,
    getSelectedTypeInfo,
    getFieldLabel,
    resetForm,
  } = useNotifierForm({ notifier, onSubmit: handleSubmit });

  // Reset form when sheet opens in create mode
  useEffect(() => {
    if (isOpen && !notifier) {
      // Reset form when opening in create mode
      resetForm();
    }
  }, [isOpen, notifier, resetForm]);

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
              ? t("notifiers.form.edit.title")
              : t("notifiers.form.create.title")}
          </FormSheetTitle>
          <FormSheetDescription>
            {isEdit
              ? t("notifiers.form.edit.description")
              : t("notifiers.form.create.description")}
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
                  <NotifierBasicFields
                    formData={formData}
                    notifierTypes={notifierTypes}
                    errors={errors}
                    disabled={loading}
                    typesLoading={typesLoading}
                    isEdit={isEdit}
                    onChange={handleInputChange}
                  />

                  {formData.type && (
                    <NotifierConfigFields
                      formData={formData}
                      selectedTypeInfo={getSelectedTypeInfo()}
                      errors={errors}
                      disabled={loading}
                      onConfigChange={handleConfigChange}
                      getFieldLabel={getFieldLabel}
                    />
                  )}
                </div>
              </form>
            </FormCardContent>
          </FormCard>
        </FormCardGroup>

        <FormSheetFooter>
          <Button type="submit" disabled={loading} onClick={handleFormSubmit}>
            {loading
              ? t("notifiers.form.submitting")
              : t("notifiers.form.submit")}
          </Button>
        </FormSheetFooter>
      </FormSheetContent>
    </FormSheet>
  );
}
