import { useEffect, useRef, useState, useCallback } from 'react';
import { SSELogConnection } from '@/lib/api/sse-logs';
import type { LogMessage, SSEEventHandlers } from '@/lib/api/sse-logs';

interface UseSSELogsOptions {
  conversationId?: number;
  autoConnect?: boolean;
  onMessage?: (message: LogMessage) => void;
  onError?: (error: Event) => void;
  onOpen?: (event: Event) => void;
  onClose?: () => void;
}

interface UseSSELogsReturn {
  messages: LogMessage[];
  isConnected: boolean;
  connectionState: number;
  connect: () => void;
  disconnect: () => void;
  clearMessages: () => void;
  addMessage: (message: LogMessage) => void;
}

export const useSSELogs = (options: UseSSELogsOptions = {}): UseSSELogsReturn => {
  const {
    conversationId,
    autoConnect = false,
    onMessage,
    onError,
    onOpen,
    onClose,
  } = options;

  const [messages, setMessages] = useState<LogMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionState, setConnectionState] = useState<number>(EventSource.CLOSED);
  
  const connectionRef = useRef<SSELogConnection | null>(null);

  // 添加消息到列表
  const addMessage = useCallback((message: LogMessage) => {
    setMessages(prev => {
      // 限制消息数量，保留最新的1000条
      const newMessages = [...prev, message];
      return newMessages.length > 1000 ? newMessages.slice(-1000) : newMessages;
    });
  }, []);

  // 清空消息
  const clearMessages = useCallback(() => {
    setMessages([]);
  }, []);

  // 连接SSE
  const connect = useCallback(() => {
    if (connectionRef.current) {
      connectionRef.current.disconnect();
    }

    connectionRef.current = new SSELogConnection(conversationId);

    const handlers: SSEEventHandlers = {
      onMessage: (message: LogMessage) => {
        addMessage(message);
        onMessage?.(message);
      },
      onOpen: (event: Event) => {
        console.log('SSE连接已建立');
        setIsConnected(true);
        setConnectionState(EventSource.OPEN);
        onOpen?.(event);
      },
      onError: (event: Event) => {
        console.error('SSE连接错误');
        setIsConnected(false);
        setConnectionState(EventSource.CLOSED);
        onError?.(event);
      },
      onClose: () => {
        console.log('SSE连接已关闭');
        setIsConnected(false);
        setConnectionState(EventSource.CLOSED);
        onClose?.();
      },
    };

    connectionRef.current.connect(handlers);

    // 定时更新连接状态
    const statusInterval = setInterval(() => {
      if (connectionRef.current) {
        const state = connectionRef.current.getReadyState();
        setConnectionState(state);
        setIsConnected(connectionRef.current.isConnected());
      }
    }, 1000);

    return () => clearInterval(statusInterval);
  }, [conversationId, addMessage, onMessage, onOpen, onError, onClose]);

  // 断开连接
  const disconnect = useCallback(() => {
    if (connectionRef.current) {
      connectionRef.current.disconnect();
      connectionRef.current = null;
    }
    setIsConnected(false);
    setConnectionState(EventSource.CLOSED);
  }, []);

  // 自动连接
  useEffect(() => {
    if (autoConnect) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [autoConnect, connect, disconnect]);

  // 清理函数
  useEffect(() => {
    return () => {
      if (connectionRef.current) {
        connectionRef.current.disconnect();
      }
    };
  }, []);

  return {
    messages,
    isConnected,
    connectionState,
    connect,
    disconnect,
    clearMessages,
    addMessage,
  };
};

// 连接状态文本映射
export const getConnectionStateText = (state: number): string => {
  switch (state) {
    case EventSource.CONNECTING:
      return '连接中';
    case EventSource.OPEN:
      return '已连接';
    case EventSource.CLOSED:
    default:
      return '已断开';
  }
};

// 日志类型样式映射
export const getLogTypeStyle = (logType: string): string => {
  switch (logType) {
    case 'log':
      return 'text-foreground';
    case 'status':
      return 'text-blue-600 font-medium';
    case 'error':
      return 'text-red-600 font-medium';
    case 'connection':
      return 'text-green-600 font-medium';
    case 'heartbeat':
      return 'text-gray-400 text-xs';
    case 'test':
      return 'text-orange-600 font-medium';
    default:
      return 'text-foreground';
  }
}; 