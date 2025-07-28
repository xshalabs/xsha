import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { OperationStatsCard } from "./OperationStatsCard";
import { ResourceStatsCard } from "./ResourceStatsCard";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { AdminOperationStatsResponse } from "@/types/admin-logs";

export const AdminStatsTab: React.FC = () => {
  const { t } = useTranslation();
  const [stats, setStats] = useState<AdminOperationStatsResponse | null>(null);
  const [loading, setLoading] = useState(false);

  const loadStats = async () => {
    try {
      setLoading(true);
      const response = await apiService.adminLogs.getOperationStats();
      setStats(response);
    } catch (err: any) {
      logError(err, "Failed to load stats");
      console.error("Failed to load stats:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadStats();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">{t("common.loading")}</div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="text-center py-8">
        <p className="text-muted-foreground">
          {t("adminLogs.stats.noStatsAvailable")}
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <OperationStatsCard stats={stats} />
      <ResourceStatsCard stats={stats} />
    </div>
  );
}; 