import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ArrowLeft, CheckCircle, Clock, Play, X, BarChart3, Plus } from 'lucide-react';
import { usePageTitle } from '@/hooks/usePageTitle';
import { TaskList } from '@/components/TaskList';
import { apiService } from '@/lib/api/index';
import { logError } from '@/lib/errors';
import type { 
  Task, 
  TaskStatus, 
  TaskStats 
} from '@/types/task';
import type { Project } from '@/types/project';

const TaskListPage: React.FC = () => {
  const { t } = useTranslation();
  const { projectId } = useParams<{ projectId: string }>();
  const navigate = useNavigate();
  
  // 当前项目信息
  const [currentProject, setCurrentProject] = useState<Project | null>(null);
  
  // 任务相关状态
  const [tasks, setTasks] = useState<Task[]>([]);
  const [tasksLoading, setTasksLoading] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);
  const [statusFilter, setStatusFilter] = useState<TaskStatus | undefined>();
  
  // 项目列表（用于任务创建/编辑时的下拉选择）
  const [projects, setProjects] = useState<Project[]>([]);
  
  // 统计数据
  const [stats, setStats] = useState<TaskStats | null>(null);

  // 设置页面标题
  usePageTitle(currentProject ? `${currentProject.name} - 任务管理` : '任务管理');

  const pageSize = 20;

  // 加载项目列表
  const loadProjects = async () => {
    try {
      const response = await apiService.projects.list();
      setProjects(response.projects);
    } catch (error) {
      logError(error as Error, 'Failed to load projects');
    }
  };

  // 加载任务列表
  const loadTasks = async (page = 1, status?: TaskStatus, projectId?: number) => {
    try {
      setTasksLoading(true);
      const response = await apiService.tasks.list({
        page,
        page_size: pageSize,
        status,
        project_id: projectId
      });
      
      setTasks(response.data.tasks);
      setTotalPages(Math.ceil(response.data.total / pageSize));
      setTotal(response.data.total);
      setCurrentPage(page);
    } catch (error) {
      logError(error as Error, 'Failed to load tasks');
    } finally {
      setTasksLoading(false);
    }
  };

  // 加载统计数据
  const loadStats = async (projectId?: number) => {
    if (!projectId) return;
    
    try {
      const response = await apiService.tasks.getStats(projectId);
      setStats(response.data);
    } catch (error) {
      logError(error as Error, 'Failed to load stats');
    }
  };

  // 初始化数据
  useEffect(() => {
    loadProjects();
    loadTasks(1, statusFilter, projectId ? parseInt(projectId, 10) : undefined);
    if (projectId) {
      apiService.projects.get(parseInt(projectId, 10)).then(response => {
        setCurrentProject(response.project);
      }).catch(error => {
        logError(error as Error, 'Failed to load project');
        navigate('/projects'); // Redirect to projects list if project not found
      });
    }
  }, [projectId]);

  // 当项目筛选改变时重新加载统计
  useEffect(() => {
    if (projectId) {
      loadStats(parseInt(projectId, 10));
    } else {
      setStats(null);
    }
  }, [projectId]);

  // 处理任务创建
  const handleTaskCreate = () => {
    navigate(`/projects/${projectId}/tasks/create`);
  };

  // 处理任务编辑
  const handleTaskEdit = (task: Task) => {
    navigate(`/projects/${projectId}/tasks/${task.id}/edit`);
  };

  // 处理任务删除
  const handleTaskDelete = async (id: number) => {
    try {
      await apiService.tasks.delete(id);
      loadTasks(currentPage, statusFilter, projectId ? parseInt(projectId, 10) : undefined);
    } catch (error) {
      logError(error as Error, 'Failed to delete task');
      alert(error instanceof Error ? error.message : t('tasks.messages.deleteFailed'));
    }
  };

  // 处理查看对话
  const handleViewConversation = (task: Task) => {
    navigate(`/projects/${projectId}/tasks/${task.id}/conversation`);
  };

  // 处理页面变化
  const handlePageChange = (page: number) => {
    loadTasks(page, statusFilter, projectId ? parseInt(projectId, 10) : undefined);
  };

  // 处理筛选变化
  const handleStatusFilterChange = (status: TaskStatus | undefined) => {
    setStatusFilter(status);
    loadTasks(1, status, projectId ? parseInt(projectId, 10) : undefined);
  };

  const handleProjectFilterChange = (_projectId: number | undefined) => {
    // This function is no longer needed as project filtering is handled by URL param
    // Keeping it for now, but it will not be called from the TaskList component
    // as the project filter is now a URL param.
  };

  // 刷新数据
  const handleRefresh = () => {
    loadTasks(currentPage, statusFilter, projectId ? parseInt(projectId, 10) : undefined);
  };

  // 渲染统计卡片
  const renderStatsCards = () => {
    if (!stats) return null;

    const statItems = [
      { key: 'total', icon: BarChart3, color: 'text-muted-foreground' },
      { key: 'todo', icon: Clock, color: 'text-muted-foreground' },
      { key: 'in_progress', icon: Play, color: 'text-primary' },
      { key: 'done', icon: CheckCircle, color: 'text-accent' },
      { key: 'cancelled', icon: X, color: 'text-destructive' },
    ];

    return (
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
        {statItems.map(({ key, icon: Icon, color }) => (
          <Card key={key}>
            <CardContent className="p-4">
              <div className="flex items-center space-x-2">
                <Icon className={`w-5 h-5 ${color}`} />
                <div>
                  <p className="text-2xl font-bold">{stats[key as keyof TaskStats]}</p>
                  <p className="text-xs text-muted-foreground">
                    {t(`tasks.stats.${key}`)}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    );
  };

  return (
    <div className="container mx-auto p-6">
      {/* 页面头部 */}
      <div className="mb-6">
        <div className="flex items-center gap-4 mb-4">
          <Button variant="outline" onClick={() => navigate('/projects')}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            返回项目列表
          </Button>
          {currentProject && (
            <div>
              <h1 className="text-2xl font-bold">{currentProject.name}</h1>
              <p className="text-muted-foreground">任务管理</p>
            </div>
          )}
        </div>
      </div>

      {/* 操作按钮 */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-xl font-semibold">{t('tasks.title')}</h2>
          <p className="text-muted-foreground">{t('tasks.description')}</p>
        </div>
        <Button onClick={handleTaskCreate}>
          <Plus className="w-4 h-4 mr-2" />
          {t('tasks.create')}
        </Button>
      </div>

      {/* 统计卡片 */}
      {renderStatsCards()}

      {/* 任务列表 */}
      <TaskList
        tasks={tasks}
        projects={projects}
        loading={tasksLoading}
        currentPage={currentPage}
        totalPages={totalPages}
        total={total}
        statusFilter={statusFilter}
        projectFilter={projectId ? parseInt(projectId, 10) : undefined}
        hideProjectFilter={true}
        onPageChange={handlePageChange}
        onStatusFilterChange={handleStatusFilterChange}
        onProjectFilterChange={handleProjectFilterChange}
        onEdit={handleTaskEdit}
        onDelete={handleTaskDelete}
        onViewConversation={handleViewConversation}
        onRefresh={handleRefresh}
        onCreateNew={handleTaskCreate}
      />
    </div>
  );
};

export default TaskListPage; 