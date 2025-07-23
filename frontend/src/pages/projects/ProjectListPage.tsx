import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Plus, RefreshCw } from 'lucide-react';
import { usePageTitle } from '@/hooks/usePageTitle';
import { apiService } from '@/lib/api/index';
import { logError } from '@/lib/errors';
import { ROUTES } from '@/lib/constants';
import { ProjectList } from '@/components/ProjectList';
import type { Project } from '@/types/project';

const ProjectListPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  usePageTitle(t('common.pageTitle.projects'));

  const handleEdit = (project: Project) => {
    navigate(`/projects/${project.id}/edit`);
  };

  const handleDelete = async (id: number) => {
    if (!confirm(t('projects.messages.deleteConfirm'))) {
      return;
    }

    try {
      await apiService.projects.delete(id);
      // ProjectList 组件会自动刷新数据
    } catch (error) {
      logError(error as Error, 'Failed to delete project');
      alert(error instanceof Error ? error.message : t('projects.messages.deleteFailed'));
    }
  };

  const handleCreateNew = () => {
    navigate('/projects/create');
  };

  return (
    <div className="container mx-auto px-4 py-6">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">{t('navigation.projects')}</h1>
          <p className="text-muted-foreground">
            {t('projects.page_description')}
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={handleCreateNew}>
            <Plus className="h-4 w-4 mr-2" />
            {t('projects.create')}
          </Button>
        </div>
      </div>

      <ProjectList
        onEdit={handleEdit}
        onDelete={handleDelete}
        onCreateNew={handleCreateNew}
      />
    </div>
  );
};

export default ProjectListPage; 