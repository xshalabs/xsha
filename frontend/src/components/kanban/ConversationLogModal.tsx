import { memo } from "react";
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
  } = useLogStreaming({ conversationId, isOpen });

  const { autoScroll, toggleAutoScroll } = useAutoScroll(true);
  
  const { downloadLogs } = useLogDownload(logs, conversationId);



  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="w-full max-w-[95vw] sm:max-w-[90vw] md:max-w-6xl lg:max-w-7xl xl:max-w-8xl h-[80vh] flex flex-col">
        <LogHeader conversationId={conversationId} />
        
        <LogControls
          logs={logs}
          autoScroll={autoScroll}
          onDownload={downloadLogs}
          onToggleAutoScroll={toggleAutoScroll}
        />

        <LogContent
          logs={logs}
          isStreaming={isStreaming}
          autoScroll={autoScroll}
          connectionStatus={connectionStatus}
        />

        <LogStatus
          logs={logs}
          isStreaming={isStreaming}
        />
      </DialogContent>
    </Dialog>
  );
});

ConversationLogModal.displayName = "ConversationLogModal";
