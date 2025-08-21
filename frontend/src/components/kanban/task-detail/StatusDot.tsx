import { memo } from "react";
import { cn } from "@/lib/utils";

export type ConversationStatus = "pending" | "running" | "success" | "failed" | "cancelled";

interface StatusDotProps {
  status: ConversationStatus;
  isScheduled?: boolean;
  className?: string;
}

const statusConfig: Record<ConversationStatus, { bgColor: string; ringColor?: string; hasAnimation: boolean }> = {
  pending: {
    bgColor: "bg-yellow-500",
    ringColor: "ring-yellow-500/30",
    hasAnimation: true,
  },
  running: {
    bgColor: "bg-blue-500",
    ringColor: "ring-blue-500/30",
    hasAnimation: true,
  },
  success: {
    bgColor: "bg-green-500",
    hasAnimation: false,
  },
  failed: {
    bgColor: "bg-red-500",
    hasAnimation: false,
  },
  cancelled: {
    bgColor: "bg-gray-500",
    hasAnimation: false,
  },
};

export const StatusDot = memo<StatusDotProps>(({ status, isScheduled = false, className }) => {
  const config = statusConfig[status] || statusConfig.pending;
  
  // Override colors for scheduled pending tasks
  const bgColor = status === "pending" && isScheduled ? "bg-purple-500" : config.bgColor;
  const shadowColor = status === "pending" && isScheduled ? "shadow-purple-500/50" : `shadow-${status === "pending" ? "yellow" : "blue"}-500/50`;
  
  return (
    <div className={cn("relative flex items-center justify-center", className)}>
      <div
        className={cn(
          "w-3 h-3 rounded-full",
          bgColor,
          // Animation for pending and running states
          config.hasAnimation && [
            "animate-pulse",
            // Glowing ring effect
            "shadow-lg",
            shadowColor,
          ]
        )}
      />
      
      {/* Additional pulsing ring for pending and running states */}
      {config.hasAnimation && (
        <div
          className={cn(
            "absolute inset-0 w-3 h-3 rounded-full",
            bgColor,
            "animate-ping opacity-75"
          )}
        />
      )}
    </div>
  );
});

StatusDot.displayName = "StatusDot";
