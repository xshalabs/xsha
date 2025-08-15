import { memo, useEffect, useRef, useState, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import {
  Play,
  Square,
  RefreshCcw,
  Download,
  Terminal,
  WifiOff,
  Wifi,
} from "lucide-react";
import { toast } from "sonner";
import { tokenManager } from "@/lib/api/token";

interface ConversationLogModalProps {
  conversationId: number | null;
  isOpen: boolean;
  onClose: () => void;
}

interface LogMessage {
  line: string;
  timestamp: number;
}

export const ConversationLogModal = memo<ConversationLogModalProps>(({
  conversationId,
  isOpen,
  onClose,
}) => {
  const { t } = useTranslation();
  const [logs, setLogs] = useState<LogMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isStreaming, setIsStreaming] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<'disconnected' | 'connecting' | 'connected' | 'error' | 'unauthorized'>('disconnected');
  const [autoScroll, setAutoScroll] = useState(true);
  const [hasAuthError, setHasAuthError] = useState(false);
  const eventSourceRef = useRef<EventSource | null>(null);
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const isFinishedRef = useRef(false);
  const currentConversationRef = useRef<number | null>(null);

  const scrollToBottom = useCallback(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [autoScroll]);

  const handleStartStreaming = useCallback(() => {
    if (!conversationId || isStreaming || hasAuthError) return;
    
    // Prevent duplicate connections for the same conversation
    if (currentConversationRef.current === conversationId && isFinishedRef.current) {
      console.log('Conversation', conversationId, 'already processed and finished, preventing reconnection');
      return;
    }
    
    // Prevent automatic reconnection if already finished (unless manually triggered)
    if (isFinishedRef.current && currentConversationRef.current === conversationId) {
      console.log('Conversation already finished, preventing automatic reconnection');
      return;
    }

    console.log('Starting streaming for conversation:', conversationId);

    // Get token for authentication
    const token = tokenManager.getToken();
    if (!token) {
      setConnectionStatus('unauthorized');
      setHasAuthError(true);
      toast.error(t('taskConversations.logs.authRequired'));
      return;
    }

    // Set current conversation and reset finished state
    currentConversationRef.current = conversationId;
    setIsStreaming(true);
    setConnectionStatus('connecting');
    setLogs([]);
    setHasAuthError(false);
    isFinishedRef.current = false;

    // Add token as URL parameter for SSE authentication
    const url = `/api/v1/conversations/${conversationId}/logs/stream?token=${encodeURIComponent(token)}`;
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

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
        toast.success(t('taskConversations.logs.connected'));
      } catch (error) {
        console.error('Error parsing connected event:', error);
        setIsConnected(true);
        setConnectionStatus('connected');
        toast.success(t('taskConversations.logs.connected'));
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
        
        // Mark as finished to prevent onerror from treating connection close as error
        isFinishedRef.current = true;
        
        // Close the EventSource connection to prevent onerror
        if (eventSourceRef.current) {
          eventSourceRef.current.close();
          eventSourceRef.current = null;
        }
        
        console.log('Conversation', currentConversationRef.current, 'finished successfully');
        setIsStreaming(false);
        setConnectionStatus('disconnected');
        toast.info(t('taskConversations.logs.finished'));
      } catch (error) {
        console.error('Error parsing finished event:', error);
        isFinishedRef.current = true;
        
        // Close connection even on error
        if (eventSourceRef.current) {
          eventSourceRef.current.close();
          eventSourceRef.current = null;
        }
        
        setIsStreaming(false);
        setConnectionStatus('disconnected');
        toast.info(t('taskConversations.logs.finished'));
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
        toast.error(errorMessage);
      } catch (error) {
        console.error('Error parsing error event:', error);
        setConnectionStatus('error');
        toast.error(t('taskConversations.logs.error'));
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
      
      // Check if it's an authentication error by checking the readyState
      if (eventSource.readyState === EventSource.CLOSED) {
        // EventSource automatically closes on 401, 403, etc.
        // We can infer this is likely an auth error
        setConnectionStatus('unauthorized');
        setHasAuthError(true);
        toast.error(t('taskConversations.logs.authError'));
        
        // Close the EventSource to prevent auto-reconnection
        if (eventSourceRef.current) {
          eventSourceRef.current.close();
          eventSourceRef.current = null;
        }
      } else {
        setConnectionStatus('error');
        toast.error(t('taskConversations.logs.connectionError'));
      }
    };
  }, [conversationId, isStreaming, hasAuthError, t]);

  const handleStopStreaming = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    setIsStreaming(false);
    setIsConnected(false);
    setConnectionStatus('disconnected');
    setHasAuthError(false);
    isFinishedRef.current = false;
  }, []);

  const handleRefresh = useCallback(() => {
    console.log('Manual refresh triggered');
    // Reset finished state and conversation ref to allow manual reconnection
    isFinishedRef.current = false;
    currentConversationRef.current = null;
    handleStopStreaming();
    setTimeout(() => {
      handleStartStreaming();
    }, 500);
  }, [handleStopStreaming, handleStartStreaming]);

  const handleDownloadLogs = useCallback(() => {
    if (logs.length === 0) {
      toast.warning(t('taskConversations.logs.noLogsToDownload'));
      return;
    }

    const logContent = logs.map(log => log.line).join('\n');
    const blob = new Blob([logContent], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `conversation_${conversationId}_logs.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    toast.success(t('taskConversations.logs.downloaded'));
  }, [logs, conversationId, t]);

  const toggleAutoScroll = useCallback(() => {
    setAutoScroll(prev => !prev);
  }, []);

  // Auto-scroll when new logs arrive
  useEffect(() => {
    scrollToBottom();
  }, [logs, scrollToBottom]);

  // Cleanup on unmount or conversation change
  useEffect(() => {
    return () => {
      console.log('Component unmounting, cleaning up...');
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      isFinishedRef.current = false;
      currentConversationRef.current = null;
    };
  }, []);

  // Reset state when modal opens/closes or conversation changes
  useEffect(() => {
    console.log('Modal useEffect triggered:', { isOpen, conversationId, isFinished: isFinishedRef.current });
    
    if (isOpen && conversationId) {
      console.log('Modal opened, attempting to start streaming...');
      // Start streaming automatically when modal opens
      handleStartStreaming();
    } else {
      console.log('Modal closed or no conversation, stopping streaming...');
      handleStopStreaming();
      setLogs([]);
      isFinishedRef.current = false;
      currentConversationRef.current = null;
    }
  }, [isOpen, conversationId]); // Remove function dependencies to prevent infinite loop

  const getConnectionStatusColor = () => {
    switch (connectionStatus) {
      case 'connected':
        return 'bg-green-500';
      case 'connecting':
        return 'bg-yellow-500';
      case 'unauthorized':
        return 'bg-orange-500';
      case 'error':
        return 'bg-red-500';
      default:
        return 'bg-gray-500';
    }
  };

  const getConnectionStatusText = () => {
    switch (connectionStatus) {
      case 'connected':
        return t('taskConversations.logs.status.connected');
      case 'connecting':
        return t('taskConversations.logs.status.connecting');
      case 'unauthorized':
        return t('taskConversations.logs.status.unauthorized');
      case 'error':
        return t('taskConversations.logs.status.error');
      default:
        return t('taskConversations.logs.status.disconnected');
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl w-full h-[80vh] flex flex-col">
        <DialogHeader className="flex-shrink-0">
          <DialogTitle className="flex items-center gap-2">
            <Terminal className="w-5 h-5" />
            {t('taskConversations.logs.title')}
            {conversationId && (
              <Badge variant="outline" className="text-xs">
                ID: {conversationId}
              </Badge>
            )}
          </DialogTitle>
          <DialogDescription>
            {t('taskConversations.logs.description')}
          </DialogDescription>
        </DialogHeader>

        {/* Control Bar */}
        <div className="flex items-center gap-2 p-3 bg-muted/50 rounded-lg flex-shrink-0">
          <div className="flex items-center gap-2">
            <div className={`w-2 h-2 rounded-full ${getConnectionStatusColor()}`} />
            <span className="text-sm text-muted-foreground">
              {getConnectionStatusText()}
            </span>
          </div>

          <div className="flex-1" />

          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={toggleAutoScroll}
              className={`${autoScroll ? 'bg-primary/10' : ''}`}
            >
              <span className="text-xs">
                {autoScroll ? t('taskConversations.logs.autoScroll.on') : t('taskConversations.logs.autoScroll.off')}
              </span>
            </Button>

            {!isStreaming ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  console.log('Manual start/retry button clicked');
                  // Reset finished state and conversation ref when manually starting
                  isFinishedRef.current = false;
                  currentConversationRef.current = null;
                  handleStartStreaming();
                }}
                disabled={!conversationId}
                className={hasAuthError ? 'border-orange-500 text-orange-600' : ''}
              >
                <Play className="w-4 h-4 mr-1" />
                {hasAuthError ? t('taskConversations.logs.retry') : t('taskConversations.logs.start')}
              </Button>
            ) : (
              <Button
                variant="outline"
                size="sm"
                onClick={handleStopStreaming}
              >
                <Square className="w-4 h-4 mr-1" />
                {t('taskConversations.logs.stop')}
              </Button>
            )}

            <Button
              variant="outline"
              size="sm"
              onClick={handleRefresh}
              disabled={!conversationId}
            >
              <RefreshCcw className="w-4 h-4 mr-1" />
              {t('common.refresh')}
            </Button>

            <Button
              variant="outline"
              size="sm"
              onClick={handleDownloadLogs}
              disabled={logs.length === 0}
            >
              <Download className="w-4 h-4 mr-1" />
              {t('taskConversations.logs.download')}
            </Button>
          </div>
        </div>

        {/* Log Content */}
        <ScrollArea ref={scrollAreaRef} className="flex-1 border rounded-lg bg-background">
          <div className="p-4">
            {logs.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-40 text-muted-foreground">
                <Terminal className="w-12 h-12 mb-4 opacity-50" />
                <p className="text-sm">
                  {isStreaming 
                    ? t('taskConversations.logs.waiting') 
                    : t('taskConversations.logs.noLogs')
                  }
                </p>
              </div>
            ) : (
              <div className="space-y-1 font-mono text-xs">
                {logs.map((log, index) => (
                  <div key={index} className="flex gap-2 hover:bg-muted/30 px-2 py-1 rounded">
                    <span className="text-muted-foreground whitespace-nowrap">
                      {new Date(log.timestamp * 1000).toLocaleTimeString()}
                    </span>
                    <span className="text-foreground whitespace-pre-wrap break-all">
                      {log.line}
                    </span>
                  </div>
                ))}
                <div ref={bottomRef} />
              </div>
            )}
          </div>
        </ScrollArea>

        {/* Status Bar */}
        <div className="flex items-center justify-between text-xs text-muted-foreground p-2 border-t flex-shrink-0">
          <div className="flex items-center gap-4">
            <span>{t('taskConversations.logs.totalLines', { count: logs.length })}</span>
            <div className="flex items-center gap-1">
              {isConnected ? <Wifi className="w-3 h-3" /> : <WifiOff className="w-3 h-3" />}
              <span>
                {isConnected 
                  ? t('taskConversations.logs.realTime') 
                  : t('taskConversations.logs.offline')
                }
              </span>
            </div>
          </div>
          <div>
            {isStreaming && (
              <span className="animate-pulse">
                {t('taskConversations.logs.streaming')}
              </span>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
});

ConversationLogModal.displayName = "ConversationLogModal";
