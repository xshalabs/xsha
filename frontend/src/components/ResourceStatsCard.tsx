import React from "react";
import { useTranslation } from "react-i18next";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { AdminOperationStatsResponse } from "@/types/admin-logs";

interface ResourceStatsCardProps {
  stats: AdminOperationStatsResponse;
}

export const ResourceStatsCard: React.FC<ResourceStatsCardProps> = ({
  stats,
}) => {
  const { t } = useTranslation();

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("adminLogs.stats.resourceStats")}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {Object.entries(stats.resource_stats).map(([resource, count]) => (
            <div
              key={resource}
              className="text-center p-4 bg-muted rounded-lg"
            >
              <div className="text-2xl font-bold text-accent">{count}</div>
              <div className="text-sm text-muted-foreground capitalize">
                {resource}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}; 