import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';
import { usePageTitle } from '@/hooks/usePageTitle';
import { TaskConversation } from '@/components/TaskConversation';
import { TaskExecutionLog } from '@/components/TaskExecutionLog';
import { apiService } from '@/lib/api/index';
import { logError } from '@/lib/errors';
import type { Task } from '@/types/task';
import type { 
  TaskConversation as TaskConversationInterface, 
  ConversationFormData,
  ConversationStatus
} from '@/types/task-conversation';

const TaskConversationPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, taskId } = useParams<{ projectId: string; taskId: string }>();
  
  const [task, setTask] = useState<Task | null>(null);
  const [conversations, setConversations] = useState<TaskConversationInterface[]>([]);
  const [selectedConversationId, setSelectedConversationId] = useState<number | null>(null);
  const [conversationsLoading, setConversationsLoading] = useState(false);
  const [loading, setLoading] = useState(true);

  usePageTitle(task ? `${t('tasks.conversation')} - ${task.title}` : t('tasks.conversation'));

  // 加载任务信息
  useEffect(() => {
    const loadTask = async () => {
      if (!taskId) {
        logError(new Error('Task ID is required'), 'Invalid task ID');
        navigate(`/projects/${projectId}/tasks`);
        return;
      }

      try {
        setLoading(true);
        const response = await apiService.tasks.get(parseInt(taskId, 10));
        setTask(response.data);
      } catch (error) {
        logError(error as Error, 'Failed to load task');
        alert(error instanceof Error ? error.message : t('tasks.messages.loadFailed'));
        navigate(`/projects/${projectId}/tasks`);
      } finally {
        setLoading(false);
      }
    };

    loadTask();
  }, [taskId, projectId, navigate, t]);

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
      
      // 默认选中最新一条对话（如果没有选中的话）
      if (response.data.conversations.length > 0 && !selectedConversationId) {
        setSelectedConversationId(response.data.conversations[0].id);
      }
    } catch (error) {
      logError(error as Error, 'Failed to load conversations');
    } finally {
      setConversationsLoading(false);
    }
  };

  // 当任务加载完成后，加载对话
  useEffect(() => {
    if (task) {
      loadConversations(task.id);
    }
  }, [task]);

  // 处理发送消息
  const handleSendMessage = async (data: ConversationFormData) => {
    if (!task) return;
    
    try {
      await apiService.taskConversations.create({
        task_id: task.id,
        content: data.content
      });
      
      // 重新加载对话列表
      await loadConversations(task.id);
    } catch (error) {
      logError(error as Error, 'Failed to send message');
      throw error;
    }
  };

  const handleConversationRefresh = () => {
    if (task) {
      loadConversations(task.id);
    }
  };

  // 处理删除对话
  const handleDeleteConversation = async (conversationId: number) => {
    try {
      await apiService.taskConversations.delete(conversationId);
      
      // 如果删除的是当前选中的对话，选择另一个对话
      if (selectedConversationId === conversationId) {
        const remainingConversations = conversations.filter(c => c.id !== conversationId);
        setSelectedConversationId(remainingConversations.length > 0 ? remainingConversations[0].id : null);
      }
      
      // 重新加载对话列表
      if (task) {
        loadConversations(task.id);
      }
    } catch (error) {
      logError(error as Error, 'Failed to delete conversation');
      throw error;
    }
  };

  // 处理对话状态变化
  const handleConversationStatusChange = (conversationId: number, newStatus: ConversationStatus) => {
    setConversations(prev => 
      prev.map(conv => 
        conv.id === conversationId 
          ? { ...conv, status: newStatus }
          : conv
      )
    );
  };

  // 获取当前选中的对话
  const selectedConversation = conversations.find(c => c.id === selectedConversationId);

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="max-w-7xl mx-auto">
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
      <div className="max-w-7xl mx-auto">
        <div className="mb-6">
          <Button 
            variant="outline" 
            onClick={() => navigate(`/projects/${projectId}/tasks`)}
            className="mb-4"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            {t('common.back')}
          </Button>
          <h1 className="text-2xl font-bold">{t('tasks.conversation')}</h1>
          <p className="text-muted-foreground mt-2">
            {task.title}
          </p>
        </div>

        {/* 2栏布局 */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 h-[calc(100vh-200px)]">
          {/* 左侧：对话列表和新消息 */}
          <div className="flex flex-col">
            <TaskConversation
              taskTitle={task.title}
              conversations={conversations}
              selectedConversationId={selectedConversationId}
              loading={conversationsLoading}
              onSendMessage={handleSendMessage}
              onRefresh={handleConversationRefresh}
              onDeleteConversation={handleDeleteConversation}
              onSelectConversation={setSelectedConversationId}
              onConversationStatusChange={handleConversationStatusChange}
            />
          </div>

          {/* 右侧：执行日志 */}
          <div className="flex flex-col">
            {selectedConversation ? (
              <div className="h-full">
                <TaskExecutionLog
                  conversationId={selectedConversation.id}
                  conversationStatus={selectedConversation.status}
                  conversation={selectedConversation}
                  onStatusChange={(newStatus) => 
                    handleConversationStatusChange(selectedConversation.id, newStatus)
                  }
                />
              </div>
            ) : (
              <div className="flex items-center justify-center h-full bg-muted rounded-lg border-2 border-dashed border-border">
                <div className="text-center text-muted-foreground">
                  <p className="text-lg font-medium mb-2">{t('taskConversation.noSelection.title')}</p>
                  <p className="text-sm">{t('taskConversation.noSelection.description')}</p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default TaskConversationPage; 