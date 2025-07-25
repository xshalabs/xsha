
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useTranslation } from 'react-i18next';
import { 
  Edit, 
  Trash2, 
  Play, 
  CheckCircle, 
  Clock, 
  GitBranch,
  GitPullRequest,
  ChevronLeft, 
  ChevronRight,
  Plus,
  RefreshCw,
  Filter,
  X,
  MessageSquare
} from 'lucide-react';
import type { Task, TaskStatus } from '@/types/task';
import type { Project } from '@/types/project';

interface TaskListProps {
  tasks: Task[];
  projects: Project[];
  loading: boolean;
  currentPage: number;
  totalPages: number;
  total: number;
  statusFilter?: TaskStatus;
  projectFilter?: number;
  hideProjectFilter?: boolean; // 新增：是否隐藏项目过滤器
  onPageChange: (page: number) => void;
  onStatusFilterChange: (status: TaskStatus | undefined) => void;
  onProjectFilterChange: (projectId: number | undefined) => void;
  onEdit: (task: Task) => void;
  onDelete: (id: number) => void;
  onViewConversation?: (task: Task) => void;
  onRefresh: () => void;
  onCreateNew: () => void;
}

export function TaskList({
  tasks,
  projects,
  loading,
  currentPage,
  totalPages,
  total,
  statusFilter,
  projectFilter,
  hideProjectFilter = false,
  onPageChange,
  onStatusFilterChange,
  onProjectFilterChange,
  onEdit,
  onDelete,
  onViewConversation,
  onRefresh,
  onCreateNew
}: TaskListProps) {
  const { t } = useTranslation();

  const getStatusColor = (status: TaskStatus) => {
    switch (status) {
      case 'todo':
        return 'bg-gray-100 text-gray-800 border-gray-300';
      case 'in_progress':
        return 'bg-blue-100 text-blue-800 border-blue-300';
      case 'done':
        return 'bg-green-100 text-green-800 border-green-300';
      case 'cancelled':
        return 'bg-red-100 text-red-800 border-red-300';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-300';
    }
  };

  const getStatusIcon = (status: TaskStatus) => {
    switch (status) {
      case 'todo':
        return <Clock className="w-3 h-3" />;
      case 'in_progress':
        return <Play className="w-3 h-3" />;
      case 'done':
        return <CheckCircle className="w-3 h-3" />;
      case 'cancelled':
        return <X className="w-3 h-3" />;
      default:
        return <Clock className="w-3 h-3" />;
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };



  const handleDeleteClick = (task: Task) => {
    if (confirm(t('tasks.messages.deleteConfirm', { title: task.title }))) {
      onDelete(task.id);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <RefreshCw className="w-8 h-8 animate-spin mx-auto mb-2" />
          <p>{t('common.loading')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 头部工具栏 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <h2 className="text-2xl font-bold">{t('tasks.title')}</h2>
          <Badge variant="secondary">
            {t('tasks.totalCount', { count: total })}
          </Badge>
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            disabled={loading}
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            {t('common.refresh')}
          </Button>
          <Button onClick={onCreateNew}>
            <Plus className="w-4 h-4 mr-2" />
            {t('tasks.actions.create')}
          </Button>
        </div>
      </div>

      {/* 筛选器 */}
      <div className="flex items-center space-x-4 p-4 bg-gray-50 rounded-lg">
        <Filter className="w-4 h-4 text-gray-600" />
        
        {/* 项目筛选 */}
        {!hideProjectFilter && (
          <div className="flex items-center space-x-2">
            <label className="text-sm font-medium text-gray-700">
              {t('tasks.filters.project')}:
            </label>
            <Select
              value={projectFilter?.toString() || 'all'}
              onValueChange={(value) => 
                onProjectFilterChange(value === 'all' ? undefined : parseInt(value))
              }
            >
              <SelectTrigger className="w-[200px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('common.all')}</SelectItem>
                {projects.map((project) => (
                  <SelectItem key={project.id} value={project.id.toString()}>
                    {project.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}

        {/* 状态筛选 */}
        <div className="flex items-center space-x-2">
          <label className="text-sm font-medium text-gray-700">
            {t('tasks.filters.status')}:
          </label>
          <Select
            value={statusFilter || 'all'}
            onValueChange={(value) => 
              onStatusFilterChange(value === 'all' ? undefined : value as TaskStatus)
            }
          >
            <SelectTrigger className="w-[150px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{t('common.all')}</SelectItem>
              <SelectItem value="todo">{t('tasks.status.todo')}</SelectItem>
              <SelectItem value="in_progress">{t('tasks.status.in_progress')}</SelectItem>
              <SelectItem value="done">{t('tasks.status.done')}</SelectItem>
              <SelectItem value="cancelled">{t('tasks.status.cancelled')}</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* 任务列表 */}
      {tasks.length === 0 ? (
        <div className="text-center py-12">
          <div className="text-gray-500 mb-4">
            <Clock className="w-16 h-16 mx-auto mb-4 opacity-50" />
            <p className="text-lg">{t('tasks.empty.title')}</p>
            <p className="text-sm">{t('tasks.empty.description')}</p>
          </div>
          <Button onClick={onCreateNew}>
            <Plus className="w-4 h-4 mr-2" />
            {t('tasks.actions.create')}
          </Button>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {tasks.map((task) => (
            <Card key={task.id} className="relative">
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <CardTitle className="text-lg">{task.title}</CardTitle>
                    <CardDescription className="text-sm">
                      {task.project?.name && (
                        <span className="text-blue-600">
                          {task.project.name}
                        </span>
                      )}
                    </CardDescription>
                  </div>
                  <div className="flex items-center space-x-1">
                    <Badge 
                      variant="outline" 
                      className={`text-xs ${getStatusColor(task.status)}`}
                    >
                      {getStatusIcon(task.status)}
                      <span className="ml-1">{t(`tasks.status.${task.status}`)}</span>
                    </Badge>
                  </div>
                </div>
              </CardHeader>
              
              <CardContent className="space-y-4">
                {/* 分支信息 */}
                <div className="flex items-center text-sm text-gray-500">
                  <GitBranch className="w-4 h-4 mr-1" />
                  <span>{task.start_branch}</span>
                  {task.has_pull_request && (
                    <div className="ml-auto">
                      <Badge variant="outline" className="text-xs">
                        <GitPullRequest className="w-3 h-3 mr-1" />
                        PR
                      </Badge>
                    </div>
                  )}
                </div>

                {/* 开发环境信息 */}
                {task.dev_environment && (
                  <div className="flex items-center text-sm text-gray-500">
                    <div className="w-4 h-4 mr-1 rounded-full bg-blue-500 flex items-center justify-center">
                      <div className="w-2 h-2 bg-white rounded-full"></div>
                    </div>
                    <span>{task.dev_environment.name}</span>
                    <Badge 
                      variant="outline" 
                      className={`text-xs ml-2 ${
                        task.dev_environment.status === 'running' ? 'bg-green-100 text-green-800 border-green-300' :
                        task.dev_environment.status === 'stopped' ? 'bg-gray-100 text-gray-800 border-gray-300' :
                        task.dev_environment.status === 'error' ? 'bg-red-100 text-red-800 border-red-300' :
                        'bg-yellow-100 text-yellow-800 border-yellow-300'
                      }`}
                    >
                      {task.dev_environment.status}
                    </Badge>
                  </div>
                )}

                {/* 时间信息 */}
                <div className="text-xs text-gray-500">
                  <div>{t('common.createdAt')}: {formatDate(task.created_at)}</div>
                  <div>{t('common.updatedAt')}: {formatDate(task.updated_at)}</div>
                </div>

                {/* 操作按钮 */}
                <div className="flex items-center justify-between pt-2 border-t">
                  <div className="flex items-center space-x-1">
                    {/* 显示状态 - 只读 */}
                    <Badge 
                      variant="outline" 
                      className={`text-xs ${getStatusColor(task.status)}`}
                    >
                      {getStatusIcon(task.status)}
                      <span className="ml-1">
                        {t(`tasks.status.${task.status}`)}
                      </span>
                    </Badge>

                    {/* 显示PR状态 - 只读 */}
                    {task.has_pull_request && (
                      <Badge variant="secondary" className="text-xs">
                        <GitPullRequest className="w-3 h-3 mr-1" />
                        PR
                      </Badge>
                    )}
                  </div>

                  <div className="flex items-center space-x-1">
                    {onViewConversation && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onViewConversation(task)}
                        className="h-8 px-2"
                        title={t('tasks.actions.viewConversation')}
                      >
                        <MessageSquare className="w-3 h-3" />
                      </Button>
                    )}
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onEdit(task)}
                      className="h-8 px-2"
                    >
                      <Edit className="w-3 h-3" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleDeleteClick(task)}
                      className="h-8 px-2 text-red-600 hover:text-red-700"
                    >
                      <Trash2 className="w-3 h-3" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* 分页 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-gray-700">
            {t('common.pagination.info', { 
              start: (currentPage - 1) * 20 + 1,
              end: Math.min(currentPage * 20, total),
              total 
            })}
          </div>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(currentPage - 1)}
              disabled={currentPage <= 1}
            >
              <ChevronLeft className="w-4 h-4" />
              {t('common.pagination.previous')}
            </Button>
            
            <div className="flex items-center space-x-1">
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                const page = currentPage <= 3 ? i + 1 : currentPage - 2 + i;
                if (page > totalPages) return null;
                
                return (
                  <Button
                    key={page}
                    variant={page === currentPage ? "default" : "outline"}
                    size="sm"
                    onClick={() => onPageChange(page)}
                    className="w-8 h-8 p-0"
                  >
                    {page}
                  </Button>
                );
              })}
            </div>

            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(currentPage + 1)}
              disabled={currentPage >= totalPages}
            >
              {t('common.pagination.next')}
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
} 