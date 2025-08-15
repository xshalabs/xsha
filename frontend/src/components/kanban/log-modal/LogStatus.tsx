import { memo } from 'react';
import { useTranslation } from 'react-i18next';
import { Wifi, WifiOff } from 'lucide-react';
import { LogMessage } from '@/hooks/useLogStreaming';

interface LogStatusProps {
  logs: LogMessage[];
  isConnected: boolean;
  isStreaming: boolean;
}

export const LogStatus = memo<LogStatusProps>(({
  logs,
  isConnected,
  isStreaming,
}) => {
  const { t } = useTranslation();

  return (
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
  );
});

LogStatus.displayName = 'LogStatus';
