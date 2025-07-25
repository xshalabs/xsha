import { useState, useEffect, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useTranslation } from 'react-i18next';
import { 
  Send, 
  User, 
  Bot, 
  Clock, 
  CheckCircle, 
  XCircle, 
  RotateCcw,
  RefreshCw,
  MessageSquare,
  Play
} from 'lucide-react';
import type { 
  TaskConversation as TaskConversationInterface, 
  ConversationStatus, 
  ConversationFormData 
} from '@/types/task-conversation';
import { TaskExecutionLog } from './TaskExecutionLog';

interface TaskConversationProps {
  taskTitle: string;
  conversations: TaskConversationInterface[];
  loading: boolean;
  onSendMessage: (data: ConversationFormData) => Promise<void>;
  onRefresh: () => void;
  onConversationStatusChange?: (conversationId: number, newStatus: ConversationStatus) => void;
}

export function TaskConversation({
  taskTitle,
  conversations,
  loading,
  onSendMessage,
  onRefresh,
  onConversationStatusChange
}: TaskConversationProps) {
  const { t } = useTranslation();
  const [newMessage, setNewMessage] = useState('');
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [conversations]);

  // 获取状态图标
  const getStatusIcon = (status: ConversationStatus) => {
    switch (status) {
      case 'pending':
        return <Clock className="w-3 h-3" />;
      case 'running':
        return <Play className="w-3 h-3" />;
      case 'success':
        return <CheckCircle className="w-3 h-3" />;
      case 'failed':
        return <XCircle className="w-3 h-3" />;
      case 'cancelled':
        return <RotateCcw className="w-3 h-3" />;
      default:
        return <Clock className="w-3 h-3" />;
    }
  };

  // 处理发送消息
  const handleSendMessage = async () => {
    if (!newMessage.trim()) return;

    setSending(true);
    try {
      await onSendMessage({
        content: newMessage.trim()
      });
      setNewMessage('');
    } catch (error) {
      console.error('Failed to send message:', error);
    } finally {
      setSending(false);
    }
  };



  // 格式化时间
  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  // 获取状态颜色
  const getStatusColor = (status: ConversationStatus) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'running':
        return 'bg-blue-100 text-blue-800';
      case 'success':
        return 'bg-green-100 text-green-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      case 'cancelled':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="space-y-6">
      {/* 对话列表 */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <div>
            <CardTitle className="text-xl">
              {t('taskConversation.list.title')}
            </CardTitle>
            <CardDescription>
              {t('taskConversation.list.description', { taskTitle })}
            </CardDescription>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            disabled={loading}
            className="flex items-center space-x-2"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            <span>{t('common.refresh')}</span>
          </Button>
        </CardHeader>
        
        <CardContent>
          <div className="space-y-4 max-h-96 overflow-y-auto">
            {conversations.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <MessageSquare className="w-12 h-12 mx-auto mb-4 opacity-50" />
                <p>{t('taskConversation.empty.title')}</p>
                <p className="text-sm">{t('taskConversation.empty.description')}</p>
              </div>
            ) : (
              conversations.map((conversation) => (
                <div key={conversation.id} className="space-y-4">
                  {/* 对话消息 */}
                  <div className="p-4 rounded-lg border bg-gray-50 border-gray-200">
                    {/* 消息头部 */}
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <div className="p-1 rounded-full bg-gray-100">
                          <User className="w-4 h-4" />
                        </div>
                        <span className="font-medium">
                          {t('taskConversation.message')}
                        </span>
                        <span className="text-xs text-gray-500">
                          {conversation.created_by}
                        </span>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        <Badge className={getStatusColor(conversation.status)}>
                          <div className="flex items-center space-x-1">
                            {getStatusIcon(conversation.status)}
                            <span>{t(`taskConversation.status.${conversation.status}`)}</span>
                          </div>
                        </Badge>
                      </div>
                    </div>

                    {/* 消息内容 */}
                    <div className="mt-2 whitespace-pre-wrap text-sm">
                      {conversation.content}
                    </div>

                    {/* 时间戳 */}
                    <div className="text-xs text-gray-500">
                      {formatTime(conversation.created_at)}
                    </div>
                  </div>

                  {/* 执行日志 */}
                  <TaskExecutionLog
                    conversationId={conversation.id}
                    conversationStatus={conversation.status}
                    onStatusChange={(newStatus) => {
                      onConversationStatusChange?.(conversation.id, newStatus);
                    }}
                  />
                </div>
              ))
            )}
            <div ref={messagesEndRef} />
          </div>
        </CardContent>
      </Card>

      {/* 新消息输入 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">
            {t('taskConversation.newMessage')}
          </CardTitle>
        </CardHeader>
        
        <CardContent>
          <div className="space-y-4">
            {/* 消息输入 */}
            <div className="space-y-2">
              <label className="text-sm font-medium">
                {t('taskConversation.content')}:
              </label>
              <textarea
                className="w-full min-h-[120px] p-3 border rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder={t('taskConversation.contentPlaceholder')}
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
                    e.preventDefault();
                    handleSendMessage();
                  }
                }}
              />
            </div>

            {/* 发送按钮 */}
            <div className="flex justify-end">
              <Button
                onClick={handleSendMessage}
                disabled={!newMessage.trim() || sending}
                className="flex items-center space-x-2"
              >
                <Send className="w-4 h-4" />
                <span>{sending ? t('common.sending') : t('common.send')}</span>
              </Button>
            </div>

            {/* 快捷键提示 */}
            <div className="text-xs text-gray-500">
              {t('taskConversation.shortcut')}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
} 