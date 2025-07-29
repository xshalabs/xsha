import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import { ProjectForm } from "@/components/ProjectForm";
import type { Project } from "@/types/project";

const ProjectEditPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();

  const [project, setProject] = useState<Project | null>(null);
  const [loading, setLoading] = useState(true);

  usePageTitle(
    project ? `${t("projects.edit")} - ${project.name}` : t("projects.edit")
  );

  // 加载项目数据
  useEffect(() => {
    const loadProject = async () => {
      if (!id) {
        logError(new Error("Project ID is required"), "Invalid project ID");
        navigate("/projects");
        return;
      }

      try {
        setLoading(true);
        const response = await apiService.projects.get(parseInt(id, 10));
        setProject(response.project);
      } catch (error) {
        logError(error as Error, "Failed to load project");
        alert(
          error instanceof Error
            ? error.message
            : t("projects.messages.loadFailed")
        );
        navigate("/projects");
      } finally {
        setLoading(false);
      }
    };

    loadProject();
  }, [id, navigate, t]);

  const handleSubmit = (_project: Project) => {
    navigate("/projects");
  };

  const handleCancel = () => {
    navigate("/projects");
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

  if (!project) {
    return null;
  }

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

        <ProjectForm
          project={project}
          onSubmit={handleSubmit}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
};

export default ProjectEditPage;
