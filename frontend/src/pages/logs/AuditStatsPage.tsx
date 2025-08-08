import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { usePageTitle } from "@/hooks/usePageTitle";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import {
  Section,
  SectionDescription,
  SectionGroup,
  SectionHeader,
  SectionTitle,
} from "@/components/content/section";
import {
  MetricCard,
  MetricCardGroup,
  MetricCardHeader,
  MetricCardTitle,
  MetricCardValue,
} from "@/components/metric/metric-card";
import { Button } from "@/components/ui/button";
import { Calendar, RefreshCw, X } from "lucide-react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import type { AdminOperationStatsResponse } from "@/types/admin-logs";

export const AuditStatsPage: React.FC = () => {
  const { t } = useTranslation();
  usePageTitle("adminLogs.stats.title");

  const [stats, setStats] = useState<AdminOperationStatsResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const dateRange = "Last 7 days";

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

  const getOperationText = (operation: string) => {
    const operationMap = {
      create: t("adminLogs.operationLogs.operations.create"),
      read: t("adminLogs.operationLogs.operations.read"),
      update: t("adminLogs.operationLogs.operations.update"),
      delete: t("adminLogs.operationLogs.operations.delete"),
      login: t("adminLogs.operationLogs.operations.login"),
      logout: t("adminLogs.operationLogs.operations.logout"),
    } as const;
    return operationMap[operation as keyof typeof operationMap] || operation;
  };

  const getOperationVariant = (operation: string) => {
    const variantMap = {
      create: "success" as const,
      read: "default" as const,
      update: "warning" as const,
      delete: "destructive" as const,
      login: "default" as const,
      logout: "ghost" as const,
    };
    return variantMap[operation as keyof typeof variantMap] || "default";
  };

  const getTotalOperations = () => {
    if (!stats) return 0;
    return Object.values(stats.operation_stats).reduce((sum, count) => sum + count, 0);
  };

  const getTotalResources = () => {
    if (!stats) return 0;
    return Object.values(stats.resource_stats).reduce((sum, count) => sum + count, 0);
  };

  if (loading) {
    return (
      <SectionGroup>
        <div className="flex items-center justify-center h-64">
          <div className="text-muted-foreground">{t("common.loading")}</div>
        </div>
      </SectionGroup>
    );
  }

  if (!stats) {
    return (
      <SectionGroup>
        <div className="text-center py-8">
          <p className="text-muted-foreground">
            {t("adminLogs.stats.noStatsAvailable")}
          </p>
        </div>
      </SectionGroup>
    );
  }

  return (
    <SectionGroup>
      <Section>
        <SectionHeader>
          <SectionTitle>{t("adminLogs.stats.title")}</SectionTitle>
          <SectionDescription>
            {t("adminLogs.stats.timeRange")}: {stats.start_time} ~ {stats.end_time}
          </SectionDescription>
        </SectionHeader>
        <div className="flex flex-wrap gap-2">
          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline" size="sm">
                <Calendar className="w-4 h-4 mr-2" />
                {dateRange}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
              <div className="p-3">
                <p className="text-sm text-muted-foreground">
                  {t("adminLogs.stats.dateFilterNotImplemented")}
                </p>
              </div>
            </PopoverContent>
          </Popover>
          <Button variant="ghost" size="sm" onClick={loadStats}>
            <RefreshCw className="w-4 h-4 mr-2" />
            {t("common.refresh")}
          </Button>
          <Button variant="ghost" size="sm">
            <X className="w-4 h-4" />
            {t("common.reset")}
          </Button>
        </div>
        <MetricCardGroup>
          <MetricCard variant="default">
            <MetricCardHeader>
              <MetricCardTitle>{t("adminLogs.stats.totalOperations")}</MetricCardTitle>
            </MetricCardHeader>
            <MetricCardValue>{getTotalOperations()}</MetricCardValue>
          </MetricCard>
          <MetricCard variant="default">
            <MetricCardHeader>
              <MetricCardTitle>{t("adminLogs.stats.totalResources")}</MetricCardTitle>
            </MetricCardHeader>
            <MetricCardValue>{getTotalResources()}</MetricCardValue>
          </MetricCard>
          <MetricCard variant="success">
            <MetricCardHeader>
              <MetricCardTitle>{t("adminLogs.operationLogs.operations.create")}</MetricCardTitle>
            </MetricCardHeader>
            <MetricCardValue>{stats.operation_stats.create || 0}</MetricCardValue>
          </MetricCard>
          <MetricCard variant="default">
            <MetricCardHeader>
              <MetricCardTitle>{t("adminLogs.operationLogs.operations.read")}</MetricCardTitle>
            </MetricCardHeader>
            <MetricCardValue>{stats.operation_stats.read || 0}</MetricCardValue>
          </MetricCard>
          <MetricCard variant="warning">
            <MetricCardHeader>
              <MetricCardTitle>{t("adminLogs.operationLogs.operations.update")}</MetricCardTitle>
            </MetricCardHeader>
            <MetricCardValue>{stats.operation_stats.update || 0}</MetricCardValue>
          </MetricCard>
        </MetricCardGroup>
      </Section>

      <Section>
        <SectionHeader>
          <SectionTitle>{t("adminLogs.stats.operationStats")}</SectionTitle>
          <SectionDescription>
            {t("adminLogs.stats.operationBreakdown")}
          </SectionDescription>
        </SectionHeader>
        <MetricCardGroup>
          {Object.entries(stats.operation_stats).map(([operation, count]) => (
            <MetricCard key={operation} variant={getOperationVariant(operation)}>
              <MetricCardHeader>
                <MetricCardTitle className="truncate">
                  {getOperationText(operation)}
                </MetricCardTitle>
              </MetricCardHeader>
              <MetricCardValue>{count}</MetricCardValue>
            </MetricCard>
          ))}
        </MetricCardGroup>
      </Section>

      <Section>
        <SectionHeader>
          <SectionTitle>{t("adminLogs.stats.resourceStats")}</SectionTitle>
          <SectionDescription>
            {t("adminLogs.stats.resourceBreakdown")}
          </SectionDescription>
        </SectionHeader>
        <MetricCardGroup>
          {Object.entries(stats.resource_stats).map(([resource, count]) => (
            <MetricCard key={resource} variant="default">
              <MetricCardHeader>
                <MetricCardTitle className="truncate capitalize">
                  {resource}
                </MetricCardTitle>
              </MetricCardHeader>
              <MetricCardValue>{count}</MetricCardValue>
            </MetricCard>
          ))}
        </MetricCardGroup>
      </Section>
    </SectionGroup>
  );
};
