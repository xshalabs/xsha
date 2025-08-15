import { memo } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import {
  Play,
  Square,
  RefreshCcw,
  Download,
} from 'lucide-react';
import { ConnectionStatus, LogMessage } from '@/hooks/useLogStreaming';

interface LogControlsProps {
  connectionStatus: ConnectionStatus;
  isStreaming: boolean;
  hasAuthError: boolean;
  conversationId: number | null;
  logs: LogMessage[];
  autoScroll: boolean;
  onStart: () => void;
  onStop: () => void;
  onRefresh: () => void;
  onDownload: () => void;
  onToggleAutoScroll: () => void;
}

export const LogControls = memo<LogControlsProps>(({
  connectionStatus,
  isStreaming,
  hasAuthError,
  conversationId,
  logs,
  autoScroll,
  onStart,
  onStop,
  onRefresh,
  onDownload,
  onToggleAutoScroll,
}) => {
  const { t } = useTranslation();

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
          onClick={onToggleAutoScroll}
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
            onClick={onStart}
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
            onClick={onStop}
          >
            <Square className="w-4 h-4 mr-1" />
            {t('taskConversations.logs.stop')}
          </Button>
        )}

        <Button
          variant="outline"
          size="sm"
          onClick={onRefresh}
          disabled={!conversationId}
        >
          <RefreshCcw className="w-4 h-4 mr-1" />
          {t('common.refresh')}
        </Button>

        <Button
          variant="outline"
          size="sm"
          onClick={onDownload}
          disabled={logs.length === 0}
        >
          <Download className="w-4 h-4 mr-1" />
          {t('taskConversations.logs.download')}
        </Button>
      </div>
    </div>
  );
});

LogControls.displayName = 'LogControls';
