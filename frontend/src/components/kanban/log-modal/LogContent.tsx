import { memo, useRef, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Terminal } from 'lucide-react';
import { LogMessage } from '@/hooks/useLogStreaming';

interface LogContentProps {
  logs: LogMessage[];
  isStreaming: boolean;
  autoScroll: boolean;
}

export const LogContent = memo<LogContentProps>(({
  logs,
  isStreaming,
  autoScroll,
}) => {
  const { t } = useTranslation();
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = useCallback(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [autoScroll]);

  // Auto-scroll when new logs arrive
  useEffect(() => {
    scrollToBottom();
  }, [logs, scrollToBottom]);

  if (logs.length === 0) {
    return (
      <ScrollArea ref={scrollAreaRef} className="flex-1 border rounded-lg bg-background">
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
    );
  }

  return (
    <ScrollArea ref={scrollAreaRef} className="flex-1 border rounded-lg bg-background">
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
  );
});

LogContent.displayName = 'LogContent';
