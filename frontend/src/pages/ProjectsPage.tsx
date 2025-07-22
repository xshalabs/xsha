import { useState } from 'react';
import { ProjectList } from '@/components/ProjectList';
import { ProjectForm } from '@/components/ProjectForm';
import { Button } from '@/components/ui/button';
import { apiService } from '@/lib/api';
import { logError } from '@/lib/errors';
import { useTranslation } from 'react-i18next';
import type { Project } from '@/types/project';

type ViewMode = 'list' | 'create' | 'edit';

export function ProjectsPage() {
  const { t } = useTranslation();
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [editingProject, setEditingProject] = useState<Project | undefined>();
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const handleCreateNew = () => {
    setEditingProject(undefined);
    setViewMode('create');
  };

  const handleEdit = (project: Project) => {
    setEditingProject(project);
    setViewMode('edit');
  };

  const handleDelete = async (id: number) => {
    if (!confirm(t('projects.messages.deleteConfirm'))) {
      return;
    }

    try {
      await apiService.projects.delete(id);
      // 触发列表刷新
      setRefreshTrigger(prev => prev + 1);
    } catch (error) {
      logError(error as Error, 'Failed to delete project');
      alert(error instanceof Error ? error.message : t('projects.messages.deleteFailed'));
    }
  };



  const handleUse = async (id: number) => {
    try {
      const response = await apiService.projects.use(id);
      // 可以在这里处理使用项目的逻辑，比如显示项目详情或跳转到其他页面
      console.log('Using project:', response.project);
      alert(t('projects.messages.useSuccess'));
    } catch (error) {
      logError(error as Error, 'Failed to use project');
      alert(error instanceof Error ? error.message : t('projects.messages.useFailed'));
    }
  };

  const handleFormSubmit = (_project: Project) => {
    // 回到列表视图并触发刷新
    setViewMode('list');
    setEditingProject(undefined);
    setRefreshTrigger(prev => prev + 1);

    // 显示成功消息
    const message = editingProject 
      ? t('projects.messages.updateSuccess') 
      : t('projects.messages.createSuccess');
    
    // 可以使用 toast 或其他通知方式
    alert(message);
  };

  const handleFormCancel = () => {
    setViewMode('list');
    setEditingProject(undefined);
  };

  const handleBackToList = () => {
    setViewMode('list');
    setEditingProject(undefined);
  };

  return (
    <div className="container mx-auto px-4 py-8">
      {viewMode === 'list' ? (
        <ProjectList
          key={refreshTrigger} // 强制重新渲染以刷新数据
          onEdit={handleEdit}
          onDelete={handleDelete}
          onUse={handleUse}
          onCreateNew={handleCreateNew}
        />
      ) : (
        <div className="space-y-6">
          {/* 返回列表按钮 */}
          <div className="flex items-center gap-4">
            <Button variant="outline" onClick={handleBackToList}>
              ← {t('projects.backToList')}
            </Button>
            <h1 className="text-2xl font-bold">
              {viewMode === 'create' ? t('projects.create') : t('projects.edit')}
            </h1>
          </div>

          {/* 项目表单 */}
          <ProjectForm
            project={editingProject}
            onSubmit={handleFormSubmit}
            onCancel={handleFormCancel}
          />
        </div>
      )}
    </div>
  );
} 