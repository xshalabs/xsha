import { useState, useEffect, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useTranslation } from 'react-i18next';
import { 
  Send, 
  User, 
  Clock, 
  CheckCircle, 
  XCircle, 
  RotateCcw,
  RefreshCw,
  MessageSquare,
  Play,
  Trash2
} from 'lucide-react';
import type { 
  TaskConversation as TaskConversationInterface, 
  ConversationStatus, 
  ConversationFormData 
} from '@/types/task-conversation';

interface TaskConversationProps {
  taskTitle: string;
  conversations: TaskConversationInterface[];
  selectedConversationId: number | null;
  loading: boolean;
  onSendMessage: (data: ConversationFormData) => Promise<void>;
  onRefresh: () => void;
  onSelectConversation: (conversationId: number) => void;
  onConversationStatusChange?: (conversationId: number, newStatus: ConversationStatus) => void;
  onDeleteConversation?: (conversationId: number) => Promise<void>;
}

export function TaskConversation({
  taskTitle,
  conversations,
  selectedConversationId,
  loading,
  onSendMessage,
  onRefresh,
  onSelectConversation,
  onConversationStatusChange,
  onDeleteConversation
}: TaskConversationProps) {
  const { t } = useTranslation();
  const [newMessage, setNewMessage] = useState('');
  const [sending, setSending] = useState(false);
  const [deletingId, setDeletingId] = useState<number | null>(null);
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

  // 检查是否有正在处理的对话
  const hasPendingOrRunningConversations = () => {
    return conversations.some(conv => 
      conv.status === 'pending' || conv.status === 'running'
    );
  };

  // 处理发送消息
  const handleSendMessage = async () => {
    if (!newMessage.trim()) return;
    if (hasPendingOrRunningConversations()) return;

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

  // 删除对话
  const handleDeleteConversation = async (conversationId: number) => {
    if (!onDeleteConversation) return;
    
    if (window.confirm(t('taskConversation.deleteConfirm'))) {
      try {
        setDeletingId(conversationId);
        await onDeleteConversation(conversationId);
      } catch (error) {
        console.error('Failed to delete conversation:', error);
      } finally {
        setDeletingId(null);
      }
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
    <div className="space-y-6 h-full flex flex-col">
      {/* 对话列表 */}
      <Card className="flex-1 flex flex-col">
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
        
        <CardContent className="flex-1 flex flex-col">
          <div className="space-y-3 flex-1 overflow-y-auto max-h-[400px]">
            {conversations.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <MessageSquare className="w-12 h-12 mx-auto mb-4 opacity-50" />
                <p>{t('taskConversation.empty.title')}</p>
                <p className="text-sm">{t('taskConversation.empty.description')}</p>
              </div>
            ) : (
              conversations.map((conversation) => (
                <div 
                  key={conversation.id} 
                  className={`p-4 rounded-lg border cursor-pointer transition-colors ${
                    selectedConversationId === conversation.id 
                      ? 'border-blue-500 bg-blue-50' 
                      : 'border-gray-200 bg-white hover:bg-gray-50'
                  }`}
                  onClick={() => onSelectConversation(conversation.id)}
                >
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
                      {/* 删除按钮 - 只对非running状态的对话显示 */}
                      {conversation.status !== 'running' && onDeleteConversation && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleDeleteConversation(conversation.id);
                          }}
                          disabled={deletingId === conversation.id}
                          className="h-6 w-6 p-0 text-red-500 hover:text-red-700 hover:bg-red-50"
                        >
                          <Trash2 className="w-3 h-3" />
                        </Button>
                      )}
                    </div>
                  </div>

                  {/* 消息内容 */}
                  <div className="mt-2 text-sm line-clamp-3">
                    {conversation.content}
                  </div>

                  {/* 时间戳 */}
                  <div className="text-xs text-gray-400 mt-2">
                    {formatTime(conversation.created_at)}
                  </div>
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
                disabled={!newMessage.trim() || sending || hasPendingOrRunningConversations()}
                className="flex items-center space-x-2"
              >
                <Send className="w-4 h-4" />
                <span>{sending ? t('common.sending') : t('common.send')}</span>
              </Button>
            </div>

            {/* 状态提示 */}
            {hasPendingOrRunningConversations() && (
              <div className="text-sm text-amber-600 bg-amber-50 p-3 rounded-lg border border-amber-200">
                {t('taskConversation.cannotSendWhileProcessing')}
              </div>
            )}

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