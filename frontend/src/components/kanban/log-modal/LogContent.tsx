import { memo, useRef, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Terminal } from 'lucide-react';
import type { LogMessage, ConnectionStatus } from '@/hooks/useLogStreaming';

interface LogContentProps {
  logs: LogMessage[];
  isStreaming: boolean;
  autoScroll: boolean;
  connectionStatus: ConnectionStatus;
}

export const LogContent = memo<LogContentProps>(({
  logs,
  isStreaming,
  autoScroll,
  connectionStatus,
}) => {
  const { t } = useTranslation();
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = useCallback(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [autoScroll]);

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

  // Auto-scroll when new logs arrive
  useEffect(() => {
    scrollToBottom();
  }, [logs, scrollToBottom]);

  if (logs.length === 0) {
    return (
      <div className="h-full border rounded-lg bg-background overflow-hidden">
        <ScrollArea className="h-full">
          <div className="p-4">
            <div className="flex flex-col items-center justify-center h-40 text-muted-foreground">
              <Terminal className="w-12 h-12 mb-4 opacity-50" />
              <p className="text-sm">
                {isStreaming 
                  ? t('taskConversations.logs.waiting') 
                  : t('taskConversations.logs.noLogs')
                }
              </p>
            </div>
          </div>
        </ScrollArea>
      </div>
    );
  }

  return (
    <div className="h-full border rounded-lg bg-background overflow-hidden relative">
      <ScrollArea ref={scrollAreaRef} className="h-full">
        <div className="p-4">
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
        </div>
      </ScrollArea>
      
      {/* 连接状态显示在左下角 */}
      <div className="absolute bottom-2 left-2 flex items-center gap-2 bg-background/90 backdrop-blur-sm border rounded-md px-2 py-1 shadow-sm">
        <div className={`w-2 h-2 rounded-full ${getConnectionStatusColor()}`} />
        <span className="text-xs text-muted-foreground">
          {getConnectionStatusText()}
        </span>
      </div>
    </div>
  );
});

LogContent.displayName = 'LogContent';
