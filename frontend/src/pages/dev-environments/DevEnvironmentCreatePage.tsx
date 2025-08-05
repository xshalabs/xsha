import React from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
import DevEnvironmentForm from "@/components/DevEnvironmentForm";

const DevEnvironmentCreatePage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  usePageTitle(t("devEnvironments.create"));

  const handleSuccess = () => {
    navigate("/dev-environments");
  };

  const handleCancel = () => {
    navigate("/dev-environments");
  };

  return (
    <div className="container mx-auto px-4 py-6">
      <div className="max-w-2xl mx-auto">
        <div className="mb-6">
          <Button
            variant="default"
            onClick={() => navigate("/dev-environments")}
            className="mb-4"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            {t("common.back")}
          </Button>
        </div>

        <DevEnvironmentForm
          onClose={handleCancel}
          onSuccess={handleSuccess}
          mode="create"
        />
      </div>
    </div>
  );
};

export default DevEnvironmentCreatePage;
