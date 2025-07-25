import { useState, useEffect, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
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
  ConversationRole, 
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
  const [messageRole, setMessageRole] = useState<ConversationRole>('user');
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 自动滚动到底部
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [conversations]);

  // 获取状态颜色
  const getStatusColor = (status: ConversationStatus) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800 border-yellow-300';
      case 'running':
        return 'bg-blue-100 text-blue-800 border-blue-300';
      case 'success':
        return 'bg-green-100 text-green-800 border-green-300';
      case 'failed':
        return 'bg-red-100 text-red-800 border-red-300';
      case 'cancelled':
        return 'bg-gray-100 text-gray-800 border-gray-300';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-300';
    }
  };

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

  // 获取角色图标
  const getRoleIcon = (role: ConversationRole) => {
    return role === 'user' ? <User className="w-4 h-4" /> : <Bot className="w-4 h-4" />;
  };

  // 获取角色颜色
  const getRoleColor = (role: ConversationRole) => {
    return role === 'user' 
      ? 'bg-blue-50 border-blue-200' 
      : 'bg-gray-50 border-gray-200';
  };

  // 处理发送消息
  const handleSendMessage = async () => {
    if (!newMessage.trim()) return;

    setSending(true);
    try {
      await onSendMessage({
        content: newMessage.trim(),
        role: messageRole
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

  return (
    <div className="space-y-6">
      {/* 头部 */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <h3 className="text-xl font-semibold flex items-center">
            <MessageSquare className="w-5 h-5 mr-2" />
            {t('taskConversation.title')}
          </h3>
          <p className="text-sm text-gray-600">
            {t('taskConversation.subtitle', { taskTitle })}
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={onRefresh}
          disabled={loading}
        >
          <RefreshCw className="w-4 h-4 mr-2" />
          {t('common.refresh')}
        </Button>
      </div>

      {/* 对话历史 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">
            {t('taskConversation.history')}
          </CardTitle>
          {conversations.length > 0 && (
            <CardDescription>
              {t('taskConversation.messageCount', { count: conversations.length })}
            </CardDescription>
          )}
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
                  <div
                    className={`p-4 rounded-lg border ${getRoleColor(conversation.role)}`}
                  >
                    {/* 消息头部 */}
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <div className={`p-1 rounded-full ${
                          conversation.role === 'user' ? 'bg-blue-100' : 'bg-gray-100'
                        }`}>
                          {getRoleIcon(conversation.role)}
                        </div>
                        <span className="font-medium">
                          {t(`taskConversation.role.${conversation.role}`)}
                        </span>
                        <span className="text-xs text-gray-500">
                          {conversation.created_by}
                        </span>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        {/* 显示状态 - 只读 */}
                        <Badge 
                          variant="outline" 
                          className={`text-xs ${getStatusColor(conversation.status)}`}
                        >
                          {getStatusIcon(conversation.status)}
                          <span className="ml-1">
                            {t(`taskConversation.status.${conversation.status}`)}
                          </span>
                        </Badge>
                      </div>
                    </div>

                    {/* 消息内容 */}
                    <div className="mb-2">
                      <p className="text-sm whitespace-pre-wrap">{conversation.content}</p>
                    </div>

                    {/* 时间信息 */}
                    <div className="text-xs text-gray-500">
                      {formatTime(conversation.created_at)}
                    </div>
                  </div>

                  {/* 执行日志 - 只对用户消息显示 */}
                  {conversation.role === 'user' && (
                    <TaskExecutionLog
                      conversationId={conversation.id}
                      conversationStatus={conversation.status}
                      onStatusChange={(newStatus) => {
                        onConversationStatusChange?.(conversation.id, newStatus);
                      }}
                    />
                  )}
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
            {/* 角色选择 */}
            <div className="flex items-center space-x-4">
              <label className="text-sm font-medium">
                {t('taskConversation.messageRole')}:
              </label>
              <Select
                value={messageRole}
                onValueChange={(value) => setMessageRole(value as ConversationRole)}
              >
                <SelectTrigger className="w-32">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="user">
                    <div className="flex items-center space-x-2">
                      <User className="w-4 h-4" />
                      <span>{t('taskConversation.role.user')}</span>
                    </div>
                  </SelectItem>
                  <SelectItem value="assistant">
                    <div className="flex items-center space-x-2">
                      <Bot className="w-4 h-4" />
                      <span>{t('taskConversation.role.assistant')}</span>
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* 消息输入 */}
            <div className="flex space-x-2">
              <textarea
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                placeholder={t('taskConversation.messagePlaceholder')}
                rows={3}
                className="flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-none"
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                    e.preventDefault();
                    handleSendMessage();
                  }
                }}
                disabled={sending}
              />
              <Button
                onClick={handleSendMessage}
                disabled={!newMessage.trim() || sending}
                className="self-end"
              >
                <Send className="w-4 h-4 mr-2" />
                {sending ? t('common.sending') : t('common.send')}
              </Button>
            </div>
            
            <p className="text-xs text-gray-500">
              {t('taskConversation.sendHint')}
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
} 