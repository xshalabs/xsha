import { memo } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import {
  Download,
} from 'lucide-react';
import type { LogMessage } from '@/hooks/useLogStreaming';

interface LogControlsProps {
  logs: LogMessage[];
  autoScroll: boolean;
  onDownload: () => void;
  onToggleAutoScroll: () => void;
}

export const LogControls = memo<LogControlsProps>(({
  logs,
  autoScroll,
  onDownload,
  onToggleAutoScroll,
}) => {
  const { t } = useTranslation();



  return (
    <div className="flex items-center gap-2 p-3 bg-muted/50 rounded-lg flex-shrink-0">
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
