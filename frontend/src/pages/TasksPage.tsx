import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from 'react-router-dom';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { usePageTitle } from '@/hooks/usePageTitle';
import { TaskList } from '@/components/TaskList';
import { TaskForm } from '@/components/TaskForm';
import { TaskConversation } from '@/components/TaskConversation';
import { apiService } from '@/lib/api/index';
import { logError } from '@/lib/errors';
import { ArrowLeft, CheckCircle, Clock, Play, X, BarChart3 } from 'lucide-react';
import type { 
  Task, 
  TaskStatus, 
  TaskFormData, 
  TaskStats 
} from '@/types/task';
import type { 
  TaskConversation as TaskConversationInterface, 
  ConversationFormData
} from '@/types/task-conversation';
import type { Project } from '@/types/project';

type ViewMode = 'list' | 'create' | 'edit' | 'conversation';

export function TasksPage() {
  const { t } = useTranslation();
  const { projectId } = useParams<{ projectId: string }>();
  const navigate = useNavigate();
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [selectedTask, setSelectedTask] = useState<Task | undefined>();
  
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
  
  // 对话相关状态
  const [conversations, setConversations] = useState<TaskConversationInterface[]>([]);
  const [conversationsLoading, setConversationsLoading] = useState(false);
  
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

  // 加载对话列表
  const loadConversations = async (taskId: number) => {
    try {
      setConversationsLoading(true);
      const response = await apiService.taskConversations.list({
        task_id: taskId,
        page: 1,
        page_size: 100 // 一次加载更多对话
      });
      
      setConversations(response.data.conversations);
    } catch (error) {
      logError(error as Error, 'Failed to load conversations');
    } finally {
      setConversationsLoading(false);
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
    setSelectedTask(undefined);
    setViewMode('create');
  };

  // 处理任务编辑
  const handleTaskEdit = (task: Task) => {
    setSelectedTask(task);
    setViewMode('edit');
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



  // 处理表单提交
  const handleFormSubmit = async (data: TaskFormData) => {
    try {
      if (selectedTask) {
        // 编辑任务
        await apiService.tasks.update(selectedTask.id, data);
        alert(t('tasks.messages.updateSuccess'));
      } else {
        // 创建任务
        const projectIdNum = projectId ? parseInt(projectId, 10) : undefined;
        if (!projectIdNum) {
          throw new Error('项目ID不能为空');
        }
        await apiService.tasks.create({ ...data, project_id: projectIdNum });
        alert(t('tasks.messages.createSuccess'));
      }
      
      setViewMode('list');
      setSelectedTask(undefined);
      loadTasks(currentPage, statusFilter, projectId ? parseInt(projectId, 10) : undefined);
    } catch (error) {
      logError(error as Error, 'Failed to submit task');
      throw error; // 让表单组件处理错误显示
    }
  };

  // 处理查看对话
  const handleViewConversation = (task: Task) => {
    setSelectedTask(task);
    setViewMode('conversation');
    loadConversations(task.id);
  };

  // 处理发送消息
  const handleSendMessage = async (data: ConversationFormData) => {
    if (!selectedTask) return;
    
    try {
      await apiService.taskConversations.create({
        task_id: selectedTask.id,
        content: data.content,
        role: data.role
      });
      
      // 重新加载对话列表
      loadConversations(selectedTask.id);
    } catch (error) {
      logError(error as Error, 'Failed to send message');
      throw error;
    }
  };



  // 处理返回列表
  const handleBackToList = () => {
    setViewMode('list');
    setSelectedTask(undefined);
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
    // If project filtering is re-introduced, this function will need to be updated.
  };

  // 刷新数据
  const handleRefresh = () => {
    loadTasks(currentPage, statusFilter, projectId ? parseInt(projectId, 10) : undefined);
  };

  const handleConversationRefresh = () => {
    if (selectedTask) {
      loadConversations(selectedTask.id);
    }
  };

  // 渲染统计卡片
  const renderStatsCards = () => {
    if (!stats) return null;

    const statItems = [
      { key: 'total', icon: BarChart3, color: 'text-gray-600' },
      { key: 'todo', icon: Clock, color: 'text-gray-600' },
      { key: 'in_progress', icon: Play, color: 'text-blue-600' },
      { key: 'done', icon: CheckCircle, color: 'text-green-600' },
      { key: 'cancelled', icon: X, color: 'text-red-600' },
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
                  <p className="text-xs text-gray-500">
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

  // 根据视图模式渲染内容
  const renderContent = () => {
    switch (viewMode) {
      case 'create':
      case 'edit':
        return (
          <div>
            <div className="mb-6">
              <Button
                variant="outline"
                onClick={handleBackToList}
                className="mb-4"
              >
                <ArrowLeft className="w-4 h-4 mr-2" />
                {t('common.back')}
              </Button>
            </div>
            <TaskForm
              task={selectedTask}
              projects={projects}
              onSubmit={handleFormSubmit}
              onCancel={handleBackToList}
            />
          </div>
        );

      case 'conversation':
        return (
          <div>
            <div className="mb-6">
              <Button
                variant="outline"
                onClick={handleBackToList}
                className="mb-4"
              >
                <ArrowLeft className="w-4 h-4 mr-2" />
                {t('common.back')}
              </Button>
            </div>
            {selectedTask && (
              <TaskConversation
                taskTitle={selectedTask.title}
                conversations={conversations}
                loading={conversationsLoading}
                onSendMessage={handleSendMessage}
                onRefresh={handleConversationRefresh}
              />
            )}
          </div>
        );

      default:
        return (
          <div>
            {renderStatsCards()}
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
    }
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
              <p className="text-gray-600">任务管理</p>
            </div>
          )}
        </div>
      </div>

      <Tabs defaultValue="tasks" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="tasks">{t('tasks.tabs.management')}</TabsTrigger>
          <TabsTrigger 
            value="conversations" 
            onClick={() => selectedTask && handleViewConversation(selectedTask)}
            disabled={!selectedTask}
          >
            {t('tasks.tabs.conversations')}
            {selectedTask && (
              <Badge variant="secondary" className="ml-2">
                {selectedTask.title}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="tasks" className="mt-6">
          {renderContent()}
        </TabsContent>

        <TabsContent value="conversations" className="mt-6">
          {selectedTask ? (
            <TaskConversation
              taskTitle={selectedTask.title}
              conversations={conversations}
              loading={conversationsLoading}
              onSendMessage={handleSendMessage}
              onRefresh={handleConversationRefresh}
            />
          ) : (
            <div className="text-center py-12 text-gray-500">
              <p>{t('taskConversation.selectTask')}</p>
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
} 