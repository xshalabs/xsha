import React from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { GitCredentialForm } from "@/components/GitCredentialForm";

const GitCredentialCreatePage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  usePageTitle(t("gitCredentials.create"));

  const handleSuccess = () => {
    navigate("/git-credentials");
  };

  const handleCancel = () => {
    navigate("/git-credentials");
  };

  return (
    <div className="container mx-auto p-6">
      <div className="max-w-2xl mx-auto">
        <div className="mb-6">
          <Button
            variant="default"
            onClick={() => navigate("/git-credentials")}
            className="mb-4"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            {t("common.back")}
          </Button>
        </div>

        <GitCredentialForm onSuccess={handleSuccess} onCancel={handleCancel} />
      </div>
    </div>
  );
};

export default GitCredentialCreatePage;
