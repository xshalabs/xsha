import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import { useCredentialForm } from "@/hooks/useCredentialForm";
import { CredentialBasicFields } from "@/components/forms/CredentialBasicFields";
import { CredentialSecretFields } from "@/components/forms/CredentialSecretFields";
import type { GitCredential } from "@/types/credentials";

interface CredentialFormSheetProps {
  credential?: GitCredential;
  onSubmit: (credential: GitCredential) => Promise<void>;
  onCancel?: () => void;
  formId?: string;
}

export function CredentialFormSheet({
  credential,
  onSubmit,
  formId = "credential-form-sheet",
}: CredentialFormSheetProps) {
  const {
    formData,
    loading,
    error,
    errors,
    isEdit,
    handleInputChange,
    handleSubmit,
  } = useCredentialForm({
    credential,
    onSubmit,
  });

  return (
    <form id={formId} onSubmit={handleSubmit} className="my-4 space-y-6">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Basic Information */}
      <CredentialBasicFields
        formData={formData}
        errors={errors}
        loading={loading}
        isEdit={isEdit}
        onInputChange={handleInputChange}
      />

      {/* Secret Fields */}
      <CredentialSecretFields
        formData={formData}
        errors={errors}
        loading={loading}
        isEdit={isEdit}
        onInputChange={handleInputChange}
      />
    </form>
  );
}
