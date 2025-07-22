import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
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
  ConversationFormData, 
  ConversationStatus 
} from '@/types/task-conversation';
import type { Project } from '@/types/project';

type ViewMode = 'list' | 'create' | 'edit' | 'conversation';

export function TasksPage() {
  const { t } = useTranslation();
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [selectedTask, setSelectedTask] = useState<Task | undefined>();
  
  // 任务相关状态
  const [tasks, setTasks] = useState<Task[]>([]);
  const [tasksLoading, setTasksLoading] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);
  const [statusFilter, setStatusFilter] = useState<TaskStatus | undefined>();
  const [projectFilter, setProjectFilter] = useState<number | undefined>();
  
  // 项目列表
  const [projects, setProjects] = useState<Project[]>([]);
  
  // 对话相关状态
  const [conversations, setConversations] = useState<TaskConversationInterface[]>([]);
  const [conversationsLoading, setConversationsLoading] = useState(false);
  
  // 统计数据
  const [stats, setStats] = useState<TaskStats | null>(null);

  // 设置页面标题
  usePageTitle('common.pageTitle.tasks');

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
    loadTasks(1, statusFilter, projectFilter);
  }, []);

  // 当项目筛选改变时重新加载统计
  useEffect(() => {
    if (projectFilter) {
      loadStats(projectFilter);
    } else {
      setStats(null);
    }
  }, [projectFilter]);

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
      loadTasks(currentPage, statusFilter, projectFilter);
    } catch (error) {
      logError(error as Error, 'Failed to delete task');
      alert(error instanceof Error ? error.message : t('tasks.messages.deleteFailed'));
    }
  };

  // 处理任务状态更新
  const handleTaskStatusUpdate = async (id: number, status: TaskStatus) => {
    try {
      await apiService.tasks.updateStatus(id, { status });
      loadTasks(currentPage, statusFilter, projectFilter);
    } catch (error) {
      logError(error as Error, 'Failed to update task status');
      alert(error instanceof Error ? error.message : t('tasks.messages.updateStatusFailed'));
    }
  };

  // 处理PR状态切换
  const handlePRToggle = async (id: number, hasPR: boolean) => {
    try {
      await apiService.tasks.updatePullRequestStatus(id, { has_pull_request: hasPR });
      loadTasks(currentPage, statusFilter, projectFilter);
    } catch (error) {
      logError(error as Error, 'Failed to update PR status');
      alert(error instanceof Error ? error.message : t('tasks.messages.updatePRFailed'));
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
        await apiService.tasks.create(data);
        alert(t('tasks.messages.createSuccess'));
      }
      
      setViewMode('list');
      setSelectedTask(undefined);
      loadTasks(currentPage, statusFilter, projectFilter);
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

  // 处理对话状态更新
  const handleConversationStatusUpdate = async (conversationId: number, status: ConversationStatus) => {
    try {
      await apiService.taskConversations.updateStatus(conversationId, { status });
      
      // 重新加载对话列表
      if (selectedTask) {
        loadConversations(selectedTask.id);
      }
    } catch (error) {
      logError(error as Error, 'Failed to update conversation status');
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
    loadTasks(page, statusFilter, projectFilter);
  };

  // 处理筛选变化
  const handleStatusFilterChange = (status: TaskStatus | undefined) => {
    setStatusFilter(status);
    loadTasks(1, status, projectFilter);
  };

  const handleProjectFilterChange = (projectId: number | undefined) => {
    setProjectFilter(projectId);
    loadTasks(1, statusFilter, projectId);
  };

  // 刷新数据
  const handleRefresh = () => {
    loadTasks(currentPage, statusFilter, projectFilter);
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
                taskId={selectedTask.id}
                taskTitle={selectedTask.title}
                conversations={conversations}
                loading={conversationsLoading}
                onSendMessage={handleSendMessage}
                onUpdateStatus={handleConversationStatusUpdate}
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
               projectFilter={projectFilter}
               onPageChange={handlePageChange}
               onStatusFilterChange={handleStatusFilterChange}
               onProjectFilterChange={handleProjectFilterChange}
               onEdit={handleTaskEdit}
               onDelete={handleTaskDelete}
               onUpdateStatus={handleTaskStatusUpdate}
               onTogglePR={handlePRToggle}
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
              taskId={selectedTask.id}
              taskTitle={selectedTask.title}
              conversations={conversations}
              loading={conversationsLoading}
              onSendMessage={handleSendMessage}
              onUpdateStatus={handleConversationStatusUpdate}
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