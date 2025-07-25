import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';
import { usePageTitle } from '@/hooks/usePageTitle';
import { TaskConversation } from '@/components/TaskConversation';
import { apiService } from '@/lib/api/index';
import { logError } from '@/lib/errors';
import type { Task } from '@/types/task';
import type { 
  TaskConversation as TaskConversationInterface, 
  ConversationFormData
} from '@/types/task-conversation';

const TaskConversationPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, taskId } = useParams<{ projectId: string; taskId: string }>();
  
  const [task, setTask] = useState<Task | null>(null);
  const [conversations, setConversations] = useState<TaskConversationInterface[]>([]);
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
      loadConversations(task.id);
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

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="max-w-4xl mx-auto">
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
      <div className="max-w-4xl mx-auto">
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

        <TaskConversation
          taskTitle={task.title}
          conversations={conversations}
          loading={conversationsLoading}
          onSendMessage={handleSendMessage}
          onRefresh={handleConversationRefresh}
        />
      </div>
    </div>
  );
};

export default TaskConversationPage; 