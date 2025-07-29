import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';
import { usePageTitle } from '@/hooks/usePageTitle';
import { TaskForm } from '@/components/TaskForm';
import { apiService } from '@/lib/api/index';
import { logError } from '@/lib/errors';
import type { Task, TaskFormData } from '@/types/task';

const TaskEditPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, taskId } = useParams<{ projectId: string; taskId: string }>();
  
  const [task, setTask] = useState<Task | null>(null);
  const [loading, setLoading] = useState(true);

  usePageTitle(task ? `${t('tasks.edit')} - ${task.title}` : t('tasks.edit'));

  // 加载数据
  useEffect(() => {
    const loadData = async () => {
      if (!taskId || !projectId) {
        logError(new Error('Task ID and Project ID are required'), 'Invalid IDs');
        navigate(`/projects/${projectId}/tasks`);
        return;
      }

      try {
        setLoading(true);
        
        // 加载任务
        const taskResponse = await apiService.tasks.get(parseInt(taskId, 10));
        setTask(taskResponse.task);
      } catch (error) {
        logError(error as Error, 'Failed to load task');
        alert(error instanceof Error ? error.message : t('tasks.messages.loadFailed'));
        navigate(`/projects/${projectId}/tasks`);
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [taskId, projectId, navigate, t]);

  const handleSubmit = async (data: TaskFormData | { title: string }) => {
    if (!taskId) return;
    
    try {
      // 确保传递正确的数据格式给API（编辑时只有title字段）
      await apiService.tasks.update(parseInt(taskId, 10), data as { title: string });
      navigate(`/projects/${projectId}/tasks`);
    } catch (error) {
      logError(error as Error, 'Failed to submit task');
      throw error; // 让表单组件处理错误显示
    }
  };

  const handleCancel = () => {
    navigate(`/projects/${projectId}/tasks`);
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="max-w-2xl mx-auto">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
              <p className="text-muted-foreground">{t('common.loading')}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!task) {
    return null;
  }

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
          <h1 className="text-2xl font-bold">{t('tasks.edit')}</h1>
          <p className="text-muted-foreground mt-2">
            {t('tasks.edit_description')} - {task.title}
          </p>
        </div>

        <TaskForm
          task={task}
          onSubmit={handleSubmit}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
};

export default TaskEditPage; 