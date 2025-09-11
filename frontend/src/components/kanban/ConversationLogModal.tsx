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
  projectId?: number;
  taskId?: number;
  isOpen: boolean;
  onClose: () => void;
}

export const ConversationLogModal = memo<ConversationLogModalProps>(({
  conversationId,
  projectId,
  taskId,
  isOpen,
  onClose,
}) => {
  // Custom hooks for separated concerns
  const {
    logs,
    connectionStatus,
    isStreaming,
  } = useLogStreaming({ conversationId, projectId, taskId, isOpen });

  const { autoScroll, toggleAutoScroll } = useAutoScroll(true);
  
  const { downloadLogs } = useLogDownload(logs, conversationId);



  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="w-full max-w-[95vw] sm:max-w-[90vw] md:max-w-6xl lg:max-w-7xl xl:max-w-8xl h-[85vh] flex flex-col overflow-hidden">
        <LogHeader conversationId={conversationId} />
        
        <LogControls
          logs={logs}
          autoScroll={autoScroll}
          onDownload={downloadLogs}
          onToggleAutoScroll={toggleAutoScroll}
        />

        <div className="flex-1 min-h-0 overflow-hidden">
          <LogContent
            logs={logs}
            isStreaming={isStreaming}
            autoScroll={autoScroll}
            connectionStatus={connectionStatus}
          />
        </div>

        <LogStatus
          logs={logs}
          isStreaming={isStreaming}
        />
      </DialogContent>
    </Dialog>
  );
});

ConversationLogModal.displayName = "ConversationLogModal";
