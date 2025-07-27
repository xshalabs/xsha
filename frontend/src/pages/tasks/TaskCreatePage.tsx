import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';
import { usePageTitle } from '@/hooks/usePageTitle';
import { TaskForm } from '@/components/TaskForm';
import { apiService } from '@/lib/api/index';
import { logError } from '@/lib/errors';
import type { TaskFormData } from '@/types/task';
import type { Project } from '@/types/project';

const TaskCreatePage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId } = useParams<{ projectId: string }>();
  
  const [currentProject, setCurrentProject] = useState<Project | null>(null);

  usePageTitle(t('tasks.create'));

  // 加载当前项目信息
  useEffect(() => {
    const loadCurrentProject = async () => {
      if (projectId) {
        try {
          const projectResponse = await apiService.projects.get(parseInt(projectId, 10));
          setCurrentProject(projectResponse.project);
        } catch (error) {
          logError(error as Error, 'Failed to load current project');
        }
      }
    };

    loadCurrentProject();
  }, [projectId]);

  const handleSubmit = async (data: TaskFormData | { title: string }) => {
    try {
      const projectIdNum = projectId ? parseInt(projectId, 10) : undefined;
      if (!projectIdNum) {
        throw new Error(t('errors.project_id_required'));
      }
      
      // 类型守卫，确保data包含必要的字段
      if ('start_branch' in data && 'project_id' in data) {
        await apiService.tasks.create({ ...data, project_id: projectIdNum });
      } else {
        // 如果只有title，需要其他必需字段的默认值
        throw new Error(t('errors.task_fields_required'));
      }
      navigate(`/projects/${projectId}/tasks`);
    } catch (error) {
      logError(error as Error, 'Failed to submit task');
      throw error; // 让表单组件处理错误显示
    }
  };

  const handleCancel = () => {
    navigate(`/projects/${projectId}/tasks`);
  };

  return (
    <div className="container mx-auto p-6">
      <div className="max-w-2xl mx-auto">
        <div className="mb-6">
          <Button 
            variant="outline" 
            onClick={() => navigate(`/projects/${projectId}/tasks`)}
            className="mb-4"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            {t('common.back')}
          </Button>
          <h1 className="text-2xl font-bold">{t('tasks.create')}</h1>
          <p className="text-muted-foreground mt-2">
            {currentProject && `${t('tasks.create_description')} - ${currentProject.name}`}
          </p>
        </div>

        <TaskForm
          defaultProjectId={projectId ? parseInt(projectId, 10) : undefined}
          currentProject={currentProject || undefined}
          onSubmit={handleSubmit}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
};

export default TaskCreatePage; 