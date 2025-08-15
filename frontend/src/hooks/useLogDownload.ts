import { useCallback } from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { LogMessage } from './useLogStreaming';

export interface UseLogDownloadReturn {
  downloadLogs: () => void;
}

export const useLogDownload = (
  logs: LogMessage[], 
  conversationId: number | null
): UseLogDownloadReturn => {
  const { t } = useTranslation();

  const downloadLogs = useCallback(() => {
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

  return { downloadLogs };
};
