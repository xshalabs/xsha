import { useState, useRef, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { tokenManager } from '@/lib/api/token';

export interface LogMessage {
  line: string;
  timestamp: number;
}

export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error' | 'unauthorized';

export interface UseLogStreamingOptions {
  conversationId: number | null;
  isOpen: boolean;
}

export interface UseLogStreamingReturn {
  logs: LogMessage[];
  connectionStatus: ConnectionStatus;
  isStreaming: boolean;
  isConnected: boolean;
  hasAuthError: boolean;
  startStreaming: () => void;
  stopStreaming: () => void;
  refreshStream: () => void;
  clearLogs: () => void;
}

export const useLogStreaming = ({ 
  conversationId, 
  isOpen 
}: UseLogStreamingOptions): UseLogStreamingReturn => {
  const { t } = useTranslation();
  const [logs, setLogs] = useState<LogMessage[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('disconnected');
  const [isStreaming, setIsStreaming] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [hasAuthError, setHasAuthError] = useState(false);
  
  const eventSourceRef = useRef<EventSource | null>(null);
  const isFinishedRef = useRef(false);
  const currentConversationRef = useRef<number | null>(null);

  const closeConnection = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
  }, []);

  const resetState = useCallback(() => {
    setIsStreaming(false);
    setIsConnected(false);
    setConnectionStatus('disconnected');
    setHasAuthError(false);
    isFinishedRef.current = false;
  }, []);

  const stopStreaming = useCallback(() => {
    closeConnection();
    resetState();
    currentConversationRef.current = null;
  }, [closeConnection, resetState]);

  const clearLogs = useCallback(() => {
    setLogs([]);
  }, []);

  const setupEventListeners = useCallback((eventSource: EventSource) => {
    eventSource.onopen = () => {
      setIsConnected(true);
      setConnectionStatus('connected');
    };

    eventSource.addEventListener('connected', (event: MessageEvent) => {
      try {
        if (event.data) {
          const data = JSON.parse(event.data);
          console.log('Connected:', data);
        }
        setIsConnected(true);
        setConnectionStatus('connected');
      } catch (error) {
        console.error('Error parsing connected event:', error);
        setIsConnected(true);
        setConnectionStatus('connected');
      }
    });

    eventSource.addEventListener('log', (event: MessageEvent) => {
      try {
        if (event.data) {
          const data = JSON.parse(event.data);
          setLogs(prev => [...prev, {
            line: data.line,
            timestamp: data.timestamp,
          }]);
        }
      } catch (error) {
        console.error('Error parsing log event:', error, 'data:', event.data);
      }
    });

    eventSource.addEventListener('finished', (event: MessageEvent) => {
      try {
        if (event.data) {
          const data = JSON.parse(event.data);
          console.log('Finished:', data);
        }
        
        isFinishedRef.current = true;
        closeConnection();
        
        console.log('Conversation', currentConversationRef.current, 'finished successfully');
        setIsStreaming(false);
        setConnectionStatus('disconnected');
      } catch (error) {
        console.error('Error parsing finished event:', error);
        isFinishedRef.current = true;
        closeConnection();
        setIsStreaming(false);
        setConnectionStatus('disconnected');
      }
    });

    eventSource.addEventListener('error', (event: MessageEvent) => {
      try {
        let errorMessage = t('taskConversations.logs.error');
        if (event.data) {
          const data = JSON.parse(event.data);
          errorMessage = data.message || errorMessage;
        }
        setConnectionStatus('error');
      } catch (error) {
        console.error('Error parsing error event:', error);
        setConnectionStatus('error');
      }
    });

    eventSource.onerror = (error) => {
      // Don't treat normal connection close after finished event as error
      if (isFinishedRef.current) {
        console.log('Connection closed normally after finished event');
        return;
      }
      
      console.error('SSE connection error:', error);
      setIsConnected(false);
      setIsStreaming(false);
      
      if (eventSource.readyState === EventSource.CLOSED) {
        setConnectionStatus('unauthorized');
        setHasAuthError(true);
        closeConnection();
      } else {
        setConnectionStatus('error');
      }
    };
  }, [t, closeConnection]);

  const startStreaming = useCallback(() => {
    if (!conversationId || isStreaming || hasAuthError) return;
    
    // Prevent duplicate connections for the same conversation
    if (currentConversationRef.current === conversationId && isFinishedRef.current) {
      console.log('Conversation', conversationId, 'already processed and finished, preventing reconnection');
      return;
    }
    
    // Prevent automatic reconnection if already finished
    if (isFinishedRef.current && currentConversationRef.current === conversationId) {
      console.log('Conversation already finished, preventing automatic reconnection');
      return;
    }

    console.log('Starting streaming for conversation:', conversationId);

    const token = tokenManager.getToken();
    if (!token) {
      setConnectionStatus('unauthorized');
      setHasAuthError(true);
      return;
    }

    // Set current conversation and reset states
    currentConversationRef.current = conversationId;
    setIsStreaming(true);
    setConnectionStatus('connecting');
    setLogs([]);
    setHasAuthError(false);
    isFinishedRef.current = false;

    const url = `/api/v1/conversations/${conversationId}/logs/stream?token=${encodeURIComponent(token)}`;
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

    setupEventListeners(eventSource);
  }, [conversationId, isStreaming, hasAuthError, t, setupEventListeners]);

  const refreshStream = useCallback(() => {
    console.log('Manual refresh triggered');
    // Reset states to allow manual reconnection
    isFinishedRef.current = false;
    currentConversationRef.current = null;
    stopStreaming();
    setTimeout(() => {
      startStreaming();
    }, 500);
  }, [stopStreaming, startStreaming]);

  // Auto-start streaming when modal opens
  useEffect(() => {
    console.log('Stream effect triggered:', { isOpen, conversationId, isFinished: isFinishedRef.current });
    
    if (isOpen && conversationId) {
      console.log('Modal opened, attempting to start streaming...');
      startStreaming();
    } else {
      console.log('Modal closed or no conversation, stopping streaming...');
      stopStreaming();
      clearLogs();
      isFinishedRef.current = false;
      currentConversationRef.current = null;
    }
  }, [isOpen, conversationId]); // Only depend on isOpen and conversationId

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      console.log('useLogStreaming cleanup');
      closeConnection();
      isFinishedRef.current = false;
      currentConversationRef.current = null;
    };
  }, [closeConnection]);

  return {
    logs,
    connectionStatus,
    isStreaming,
    isConnected,
    hasAuthError,
    startStreaming,
    stopStreaming,
    refreshStream,
    clearLogs,
  };
};
