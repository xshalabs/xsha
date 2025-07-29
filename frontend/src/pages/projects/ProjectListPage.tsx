import React from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import { ProjectList } from "@/components/ProjectList";
import type { Project } from "@/types/project";

const ProjectListPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  usePageTitle(t("common.pageTitle.projects"));

  const handleEdit = (project: Project) => {
    navigate(`/projects/${project.id}/edit`);
  };

  const handleDelete = async (id: number) => {
    if (!confirm(t("projects.messages.deleteConfirm"))) {
      return;
    }

    try {
      await apiService.projects.delete(id);
    } catch (error) {
      logError(error as Error, "Failed to delete project");
      alert(
        error instanceof Error
          ? error.message
          : t("projects.messages.deleteFailed")
      );
    }
  };

  const handleCreateNew = () => {
    navigate("/projects/create");
  };

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              {t("navigation.projects")}
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {t("projects.page_description")}
            </p>
          </div>
          <div className="flex gap-2">
            <Button onClick={handleCreateNew}>
              <Plus className="h-4 w-4 mr-2" />
              {t("projects.create")}
            </Button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <ProjectList
          onEdit={handleEdit}
          onDelete={handleDelete}
          onCreateNew={handleCreateNew}
        />
      </div>
    </div>
  );
};

export default ProjectListPage;
