import React from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { ProjectForm } from "@/components/ProjectForm";
import type { Project } from "@/types/project";

const ProjectCreatePage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  usePageTitle(t("projects.create"));

  const handleSubmit = (_project: Project) => {
    navigate("/projects");
  };

  const handleCancel = () => {
    navigate("/projects");
  };

  return (
    <div className="container mx-auto p-6">
      <div className="max-w-2xl mx-auto">
        <div className="mb-6">
          <Button
            variant="default"
            onClick={() => navigate("/projects")}
            className="mb-4"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            {t("common.back")}
          </Button>
        </div>

        <ProjectForm onSubmit={handleSubmit} onCancel={handleCancel} />
      </div>
    </div>
  );
};

export default ProjectCreatePage;
