import React, { useState, useEffect, useMemo } from "react";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import { useTranslation } from "react-i18next";
import {
  Section,
  SectionHeader,
  SectionTitle,
  SectionDescription,
} from "@/components/content/section";
import {
  MetricCardGroup,
  MetricCardHeader,
  MetricCardTitle,
  MetricCardValue,
  MetricCardButton,
} from "@/components/metric/metric-card";
import { DataTable } from "@/components/ui/data-table";
import { useLoginLogColumns } from "@/components/data-table/login-logs/columns";
import { LoginLogDataTableToolbar } from "@/components/data-table/login-logs/data-table-toolbar";
import { CustomPagination } from "@/components/data-table/login-logs/custom-pagination";
import { CheckCircle, Filter, Shield } from "lucide-react";
import type { LoginLog } from "@/types/admin-logs";
import type { ColumnFiltersState } from "@tanstack/react-table";

export const LoginLogTab: React.FC = () => {
  const { t } = useTranslation();
  const [logs, setLogs] = useState<LoginLog[]>([]);
  
  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 20;
  
  // DataTable state
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);

  // Calculate metrics from current page logs
  const metrics = useMemo(() => {
    const successCount = logs.filter(log => log.success).length;
    const failedCount = logs.filter(log => !log.success).length;
    const totalCurrentPage = logs.length;

    return [
      {
        title: t("adminLogs.loginLogs.status.success"),
        value: successCount,
        variant: "success" as const,
        type: "status-filter" as const,
        filterKey: "true",
      },
      {
        title: t("adminLogs.loginLogs.status.failed"),
        value: failedCount,
        variant: "destructive" as const,
        type: "status-filter" as const,
        filterKey: "false",
      },
      {
        title: t("adminLogs.loginLogs.metrics.totalAttempts"),
        value: totalCurrentPage,
        variant: "default" as const,
        type: "info" as const,
        filterKey: "",
      },
    ];
  }, [logs, t]);

  const loadLogs = async (page = currentPage, filters = columnFilters) => {
    try {
      // Convert DataTable filters to API parameters
      const apiParams: any = {
        page,
        page_size: pageSize,
      };

      // Handle column filters
      filters.forEach((filter) => {
        if (filter.id === "username" && filter.value) {
          apiParams.username = filter.value;
        } else if (filter.id === "ip" && filter.value) {
          apiParams.ip = filter.value;
        } else if (filter.id === "success" && Array.isArray(filter.value) && filter.value.length > 0) {
          apiParams.success = filter.value[0] === "true";
        } else if (filter.id === "login_time" && filter.value) {
          const dateRange = filter.value as { startDate?: Date; endDate?: Date };
          if (dateRange.startDate) {
            apiParams.start_time = dateRange.startDate.toISOString().split('T')[0];
          }
          if (dateRange.endDate) {
            apiParams.end_time = dateRange.endDate.toISOString().split('T')[0];
          }
        }
      });

      const response = await apiService.adminLogs.getLoginLogs(apiParams);

      setLogs(response.logs);
      setTotal(response.total);
      setTotalPages(response.total_pages);
      setCurrentPage(page);
    } catch (err: any) {
      logError(err, "Failed to load login logs");
      console.error("Failed to load login logs:", err);
    }
  };

  const handleMetricCardClick = (metric: typeof metrics[0]) => {
    if (metric.type !== "status-filter") return;
    
    let newColumnFilters = [...columnFilters];
    
    const currentFilter = columnFilters.find(f => f.id === "success");
    const currentValues = (currentFilter?.value as string[]) || [];
    const isActive = currentValues.includes(metric.filterKey);
    
    newColumnFilters = columnFilters.filter(f => f.id !== "success");
    if (!isActive) {
      newColumnFilters.push({ id: "success", value: [metric.filterKey] });
    }
    
    setColumnFilters(newColumnFilters);
    loadLogs(1, newColumnFilters); // Reset to page 1 when filtering
  };

  const handlePageChange = (page: number) => {
    loadLogs(page);
  };

  // Create columns
  const columns = useLoginLogColumns();

  // Handle column filter changes from DataTable toolbar (excluding initial empty state)
  const [isInitialized, setIsInitialized] = useState(false);
  
  useEffect(() => {
    if (isInitialized) {
      loadLogs(1, columnFilters);
    }
  }, [columnFilters, isInitialized]);

  useEffect(() => {
    loadLogs().then(() => setIsInitialized(true));
  }, []);

  return (
    <>
      <Section>
        <SectionHeader>
          <SectionTitle>{t("adminLogs.loginLogs.overview.title")}</SectionTitle>
          <SectionDescription>
            {t("adminLogs.loginLogs.overview.description")}
          </SectionDescription>
        </SectionHeader>
        <MetricCardGroup>
          {metrics.map((metric) => {
            const currentFilter = columnFilters.find(f => f.id === "success");
            const currentValues = (currentFilter?.value as string[]) || [];
            const isFilterActive = currentValues.includes(metric.filterKey);

            const Icon = metric.type === "status-filter" 
              ? (isFilterActive ? CheckCircle : Filter)
              : Shield;

            return (
              <MetricCardButton
                key={metric.title}
                variant={metric.variant}
                onClick={() => handleMetricCardClick(metric)}
                disabled={metric.type === "info"}
              >
                <MetricCardHeader className="flex justify-between items-center gap-2 w-full">
                  <MetricCardTitle className="truncate">
                    {metric.title}
                  </MetricCardTitle>
                  <Icon className="size-4" />
                </MetricCardHeader>
                <MetricCardValue>{metric.value}</MetricCardValue>
              </MetricCardButton>
            );
          })}
        </MetricCardGroup>
      </Section>

      <Section>
        <div className="space-y-4">
          <DataTable
            columns={columns}
            data={logs}
            toolbarComponent={LoginLogDataTableToolbar}
            columnFilters={columnFilters}
            setColumnFilters={setColumnFilters}
          />
          <CustomPagination
            currentPage={currentPage}
            totalPages={totalPages}
            total={total}
            onPageChange={handlePageChange}
          />
        </div>
      </Section>
    </>
  );
};
