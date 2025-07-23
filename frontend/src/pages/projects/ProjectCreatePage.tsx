import React from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';
import { usePageTitle } from '@/hooks/usePageTitle';
import { ProjectForm } from '@/components/ProjectForm';
import type { Project } from '@/types/project';

const ProjectCreatePage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  
  usePageTitle(t('projects.create'));

  const handleSubmit = (project: Project) => {
    navigate('/projects');
  };

  const handleCancel = () => {
    navigate('/projects');
  };

  return (
    <div className="container mx-auto px-4 py-6">
      <div className="max-w-2xl mx-auto">
        <div className="mb-6">
          <Button 
            variant="outline" 
            onClick={() => navigate('/projects')}
            className="mb-4"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            {t('common.back')}
          </Button>
          <h1 className="text-2xl font-bold">{t('projects.create')}</h1>
          <p className="text-muted-foreground mt-2">
            {t('projects.create_description')}
          </p>
        </div>

        <ProjectForm
          onSubmit={handleSubmit}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
};

export default ProjectCreatePage; 