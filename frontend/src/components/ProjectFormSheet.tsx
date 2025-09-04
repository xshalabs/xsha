import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import { useProjectForm } from "@/hooks/useProjectForm";
import { 
  ProjectBasicFields,
  RepositoryUrlField,
  ProjectCredentialSelector 
} from "@/components/forms";
import type { Project } from "@/types/project";

interface ProjectFormSheetProps {
  project?: Project;
  onSubmit: (project: Project) => Promise<void>;
  onCancel?: () => void;
  formId?: string;
}

export function ProjectFormSheet({
  project,
  onSubmit,
  formId = "project-form-sheet",
}: ProjectFormSheetProps) {
  
  const {
    formData,
    loading,
    credentialsLoading,
    credentialValidating,
    error,
    errors,
    accessError,
    accessValidated,
    credentials,
    handleInputChange,
    handleSubmit: onFormSubmit,
  } = useProjectForm({ project, onSubmit });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await onFormSubmit();
  };


  return (
    <form id={formId} onSubmit={handleSubmit} className="my-4 space-y-6">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="space-y-6">
        <ProjectBasicFields
          formData={formData}
          errors={errors}
          disabled={loading}
          onChange={handleInputChange}
        />

        <RepositoryUrlField
          formData={formData}
          errors={errors}
          disabled={loading}
          onChange={handleInputChange}
        />

        <ProjectCredentialSelector
          formData={formData}
          credentials={credentials}
          disabled={loading}
          credentialsLoading={credentialsLoading}
          credentialValidating={credentialValidating}
          accessValidated={accessValidated}
          accessError={accessError}
          onChange={handleInputChange}
        />
      </div>
    </form>
  );
}
