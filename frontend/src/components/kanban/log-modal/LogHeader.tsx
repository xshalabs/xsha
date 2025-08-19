import { memo } from 'react';
import { useTranslation } from 'react-i18next';
import {
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Terminal } from 'lucide-react';

interface LogHeaderProps {
  conversationId: number | null;
}

export const LogHeader = memo<LogHeaderProps>(({ conversationId }) => {
  const { t } = useTranslation();

  return (
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
  );
});

LogHeader.displayName = 'LogHeader';
