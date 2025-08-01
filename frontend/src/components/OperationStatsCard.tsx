import React from "react";
import { useTranslation } from "react-i18next";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { AdminOperationStatsResponse } from "@/types/admin-logs";

interface OperationStatsCardProps {
  stats: AdminOperationStatsResponse;
}

export const OperationStatsCard: React.FC<OperationStatsCardProps> = ({
  stats,
}) => {
  const { t } = useTranslation();

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("adminLogs.stats.operationStats")}</CardTitle>
        <CardDescription>
          {t("adminLogs.stats.timeRange")}: {stats.start_time} ~{" "}
          {stats.end_time}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          {Object.entries(stats.operation_stats).map(([operation, count]) => (
            <div
              key={operation}
              className="text-center p-4 bg-muted rounded-lg"
            >
              <div className="text-2xl font-bold text-primary">{count}</div>
              <div className="text-sm text-muted-foreground">
                {t(`adminLogs.operationLogs.operations.${operation}`)}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
};
