import { memo, useCallback } from "react";
import {
  Dialog,
  DialogContent,
} from "@/components/ui/dialog";
import { useLogStreaming } from "@/hooks/useLogStreaming";
import { useAutoScroll } from "@/hooks/useAutoScroll";
import { useLogDownload } from "@/hooks/useLogDownload";
import {
  LogHeader,
  LogControls,
  LogContent,
  LogStatus,
} from "./log-modal";

interface ConversationLogModalProps {
  conversationId: number | null;
  isOpen: boolean;
  onClose: () => void;
}

export const ConversationLogModal = memo<ConversationLogModalProps>(({
  conversationId,
  isOpen,
  onClose,
}) => {
  // Custom hooks for separated concerns
  const {
    logs,
    connectionStatus,
    isStreaming,
    isConnected,
    hasAuthError,
    startStreaming,
    stopStreaming,
    refreshStream,
  } = useLogStreaming({ conversationId, isOpen });

  const { autoScroll, toggleAutoScroll } = useAutoScroll(true);
  
  const { downloadLogs } = useLogDownload(logs, conversationId);

  // Event handlers with better separation
  const handleManualStart = useCallback(() => {
    console.log('Manual start/retry button clicked');
    startStreaming();
  }, [startStreaming]);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl w-full h-[80vh] flex flex-col">
        <LogHeader conversationId={conversationId} />
        
        <LogControls
          connectionStatus={connectionStatus}
          isStreaming={isStreaming}
          hasAuthError={hasAuthError}
          conversationId={conversationId}
          logs={logs}
          autoScroll={autoScroll}
          onStart={handleManualStart}
          onStop={stopStreaming}
          onRefresh={refreshStream}
          onDownload={downloadLogs}
          onToggleAutoScroll={toggleAutoScroll}
        />

        <LogContent
          logs={logs}
          isStreaming={isStreaming}
          autoScroll={autoScroll}
        />

        <LogStatus
          logs={logs}
          isConnected={isConnected}
          isStreaming={isStreaming}
        />
      </DialogContent>
    </Dialog>
  );
});

ConversationLogModal.displayName = "ConversationLogModal";
