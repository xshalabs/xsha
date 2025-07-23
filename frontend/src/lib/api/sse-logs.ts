import { request } from './request';

// SSE日志消息类型
export interface LogMessage {
  conversation_id: number;
  content: string;
  timestamp: string;
  log_type: 'log' | 'status' | 'error' | 'connection' | 'heartbeat' | 'test';
}

// SSE连接统计信息
export interface LogStats {
  connected_clients: number;
  timestamp: string;
}

// SSE日志API
export const sseLogsApi = {
  // 获取SSE连接统计
  getStats: async (): Promise<LogStats> => {
    return request<LogStats>('/api/v1/logs/stats');
  },

  // 发送测试消息
  sendTestMessage: async (conversationId: number): Promise<{ message: string; content: string }> => {
    return request<{ message: string; content: string }>(`/api/v1/logs/test/${conversationId}`, {
      method: 'POST',
    });
  },

  // 创建SSE连接
  createEventSource: (conversationId?: number): EventSource => {
    const token = localStorage.getItem('token');
    let url = '/api/v1/logs/stream';
    
    // 添加参数
    const params = new URLSearchParams();
    if (conversationId) {
      params.append('conversationId', conversationId.toString());
    }
    if (token) {
      params.append('token', token);
    }
    
    if (params.toString()) {
      url += '?' + params.toString();
    }

    return new EventSource(url);
  },
};

// SSE事件监听器类型
export interface SSEEventHandlers {
  onMessage?: (message: LogMessage) => void;
  onOpen?: (event: Event) => void;
  onError?: (event: Event) => void;
  onClose?: () => void;
}

// SSE连接管理器
export class SSELogConnection {
  private eventSource: EventSource | null = null;
  private handlers: SSEEventHandlers = {};
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000; // 1秒
  private conversationId?: number;

  constructor(conversationId?: number) {
    this.conversationId = conversationId;
  }

  // 连接SSE
  connect(handlers: SSEEventHandlers): void {
    this.handlers = handlers;
    this.createConnection();
  }

  // 创建连接
  private createConnection(): void {
    try {
      this.eventSource = sseLogsApi.createEventSource(this.conversationId);

      this.eventSource.onopen = (event) => {
        console.log('SSE连接已建立');
        this.reconnectAttempts = 0;
        this.handlers.onOpen?.(event);
      };

      this.eventSource.onmessage = (event) => {
        try {
          const message: LogMessage = JSON.parse(event.data);
          this.handlers.onMessage?.(message);
        } catch (error) {
          console.error('解析SSE消息失败:', error);
        }
      };

      this.eventSource.onerror = (event) => {
        console.error('SSE连接错误:', event);
        this.handlers.onError?.(event);
        
        // 自动重连
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
          this.reconnectAttempts++;
          setTimeout(() => {
            console.log(`尝试重连SSE (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            this.disconnect();
            this.createConnection();
          }, this.reconnectDelay * this.reconnectAttempts);
        }
      };

    } catch (error) {
      console.error('创建SSE连接失败:', error);
      this.handlers.onError?.(new Event('error'));
    }
  }

  // 断开连接
  disconnect(): void {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
      console.log('SSE连接已断开');
      this.handlers.onClose?.();
    }
  }

  // 获取连接状态
  getReadyState(): number {
    return this.eventSource?.readyState ?? EventSource.CLOSED;
  }

  // 是否已连接
  isConnected(): boolean {
    return this.eventSource?.readyState === EventSource.OPEN;
  }
} 