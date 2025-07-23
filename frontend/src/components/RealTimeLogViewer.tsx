import React, { useEffect, useRef, useState } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useSSELogs, getConnectionStateText, getLogTypeStyle } from '@/hooks/useSSELogs';
import { sseLogsApi } from '@/lib/api/sse-logs';
import type { LogMessage } from '@/lib/api/sse-logs';

interface RealTimeLogViewerProps {
  conversationId?: number;
  autoConnect?: boolean;
  height?: string;
  showControls?: boolean;
  title?: string;
}

export const RealTimeLogViewer: React.FC<RealTimeLogViewerProps> = ({
  conversationId,
  autoConnect = true,
  height = "400px",
  showControls = true,
  title = "实时日志"
}) => {
  const [stats, setStats] = useState<{ connected_clients: number; timestamp: string } | null>(null);
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);

  const {
    messages,
    isConnected,
    connectionState,
    connect,
    disconnect,
    clearMessages,
  } = useSSELogs({
    conversationId,
    autoConnect,
    onMessage: (message: LogMessage) => {
      console.log('收到日志消息:', message);
      // 自动滚动到底部
      if (autoScroll && scrollAreaRef.current) {
        setTimeout(() => {
          if (scrollAreaRef.current) {
            scrollAreaRef.current.scrollTop = scrollAreaRef.current.scrollHeight;
          }
        }, 50);
      }
    },
    onOpen: () => {
      console.log('SSE连接已打开');
    },
    onError: (error) => {
      console.error('SSE连接错误:', error);
    },
    onClose: () => {
      console.log('SSE连接已关闭');
    },
  });

  // 获取连接统计信息
  useEffect(() => {
    const fetchStats = async () => {
      try {
        const statsData = await sseLogsApi.getStats();
        setStats(statsData);
      } catch (error) {
        console.error('获取统计信息失败:', error);
      }
    };

    fetchStats();
    const interval = setInterval(fetchStats, 10000); // 每10秒更新一次

    return () => clearInterval(interval);
  }, []);

  // 发送测试消息
  const handleSendTestMessage = async () => {
    if (!conversationId) return;
    
    try {
      await sseLogsApi.sendTestMessage(conversationId);
    } catch (error) {
      console.error('发送测试消息失败:', error);
    }
  };

  // 格式化消息时间
  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  // 过滤掉心跳消息（可选）
  const filteredMessages = messages.filter(msg => msg.log_type !== 'heartbeat');

  return (
    <Card className="w-full">
      <div className="p-4 border-b">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <h3 className="font-semibold">{title}</h3>
            {conversationId && (
              <Badge variant="outline">对话 #{conversationId}</Badge>
            )}
          </div>
          <div className="flex items-center gap-2">
            <Badge 
              variant={isConnected ? "default" : "destructive"}
              className={isConnected ? "bg-green-100 text-green-700 border-green-300" : ""}
            >
              {getConnectionStateText(connectionState)}
            </Badge>
            {stats && (
              <Badge variant="outline">
                连接数: {stats.connected_clients}
              </Badge>
            )}
          </div>
        </div>

        {showControls && (
          <div className="flex items-center gap-2 mt-3">
            <Button
              size="sm"
              variant={isConnected ? "destructive" : "default"}
              onClick={isConnected ? disconnect : connect}
            >
              {isConnected ? "断开连接" : "连接"}
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={clearMessages}
              disabled={messages.length === 0}
            >
              清空日志
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => setAutoScroll(!autoScroll)}
            >
              自动滚动: {autoScroll ? "开" : "关"}
            </Button>
            {conversationId && (
              <Button
                size="sm"
                variant="outline"
                onClick={handleSendTestMessage}
                disabled={!isConnected}
              >
                发送测试
              </Button>
            )}
          </div>
        )}
      </div>

      <div 
        ref={scrollAreaRef} 
        className="overflow-y-auto" 
        style={{ height }}
      >
        <div className="p-4 space-y-1">
          {filteredMessages.length === 0 ? (
            <div className="text-center text-muted-foreground py-8">
              {isConnected ? "等待日志消息..." : "请先连接到日志流"}
            </div>
          ) : (
            filteredMessages.map((message, index) => (
              <div
                key={index}
                className="flex items-start gap-2 text-sm font-mono py-1 px-2 rounded hover:bg-muted/50"
              >
                <span className="text-muted-foreground text-xs w-16 flex-shrink-0">
                  {formatTime(message.timestamp)}
                </span>
                <Badge 
                  variant="outline" 
                  className="text-xs flex-shrink-0"
                >
                  {message.log_type}
                </Badge>
                <span className={`flex-1 whitespace-pre-wrap ${getLogTypeStyle(message.log_type)}`}>
                  {message.content}
                </span>
              </div>
            ))
          )}
        </div>
      </div>

      <div className="px-4 py-2 border-t text-xs text-muted-foreground">
        共 {filteredMessages.length} 条消息
        {isConnected && " • 实时更新中"}
        {connectionState === EventSource.CONNECTING && " • 连接中..."}
      </div>
    </Card>
  );
}; 