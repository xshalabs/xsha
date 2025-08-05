import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { apiService } from "@/lib/api/index";
import { toast } from "sonner";
import { handleApiError } from "@/lib/errors";
import DevEnvironmentForm from "@/components/DevEnvironmentForm";
import type { DevEnvironmentDisplay } from "@/types/dev-environment";

const DevEnvironmentEditPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();

  const [environment, setEnvironment] = useState<DevEnvironmentDisplay | null>(
    null
  );
  const [loading, setLoading] = useState(true);

  usePageTitle(
    environment
      ? `${t("devEnvironments.edit")} - ${environment.name}`
      : t("devEnvironments.edit")
  );

  useEffect(() => {
    const loadEnvironment = async () => {
      if (!id) {
        toast.error(t("devEnvironments.invalid_id"));
        navigate("/dev-environments");
        return;
      }

      try {
        setLoading(true);
        const response = await apiService.devEnvironments.get(parseInt(id, 10));

        let envVarsMap: Record<string, string> = {};
        try {
          if (response.environment.env_vars) {
            envVarsMap = JSON.parse(response.environment.env_vars);
          }
        } catch (error) {
          console.warn("Failed to parse env_vars:", error);
        }

        setEnvironment({
          ...response.environment,
          env_vars_map: envVarsMap,
        });
      } catch (error) {
        console.error("Failed to load environment:", error);
        const errorMessage = handleApiError(error);
        toast.error(errorMessage);
        navigate("/dev-environments");
      } finally {
        setLoading(false);
      }
    };

    loadEnvironment();
  }, [id, navigate, t]);

  const handleSuccess = () => {
    navigate("/dev-environments");
  };

  const handleCancel = () => {
    navigate("/dev-environments");
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-6">
        <div className="max-w-2xl mx-auto">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
              <p className="text-muted-foreground">{t("common.loading")}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!environment) {
    return null;
  }

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
          initialData={environment}
          mode="edit"
        />
      </div>
    </div>
  );
};

export default DevEnvironmentEditPage;
