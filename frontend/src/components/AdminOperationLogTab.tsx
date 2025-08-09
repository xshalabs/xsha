import React, { useState, useEffect, useMemo } from "react";
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import { useAdminOperationLogColumns } from "@/components/data-table/admin-logs/columns";
import { AdminOperationLogDataTableToolbar } from "@/components/data-table/admin-logs/data-table-toolbar";
import { CustomPagination } from "@/components/data-table/admin-logs/custom-pagination";
import { CheckCircle, Filter } from "lucide-react";
import type {
  AdminOperationLog,
} from "@/types/admin-logs";
import type { ColumnFiltersState } from "@tanstack/react-table";

export const AdminOperationLogTab: React.FC = () => {
  const { t } = useTranslation();
  const [logs, setLogs] = useState<AdminOperationLog[]>([]);
  const [_loading, setLoading] = useState(false);

  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [selectedLog, setSelectedLog] = useState<AdminOperationLog | null>(
    null
  );
  
  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 20;
  
  // DataTable state
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);

  // Calculate metrics from current page logs (for now, could be enhanced with global stats API)
  const metrics = useMemo(() => {
    const operationCounts = {
      create: logs.filter(log => log.operation === 'create').length,
      read: logs.filter(log => log.operation === 'read').length,
      update: logs.filter(log => log.operation === 'update').length,
      delete: logs.filter(log => log.operation === 'delete').length,
      login: logs.filter(log => log.operation === 'login').length,
      logout: logs.filter(log => log.operation === 'logout').length,
    };

    const successCount = logs.filter(log => log.success).length;
    const failedCount = logs.filter(log => !log.success).length;

    return [
      {
        title: t("adminLogs.operationLogs.operations.create"),
        value: operationCounts.create,
        variant: "success" as const,
        type: "filter" as const,
        filterKey: "create",
      },
      {
        title: t("adminLogs.operationLogs.operations.update"),
        value: operationCounts.update,
        variant: "warning" as const,
        type: "filter" as const,
        filterKey: "update",
      },
      {
        title: t("adminLogs.operationLogs.operations.delete"),
        value: operationCounts.delete,
        variant: "destructive" as const,
        type: "filter" as const,
        filterKey: "delete",
      },
      {
        title: t("adminLogs.operationLogs.status.success"),
        value: successCount,
        variant: "success" as const,
        type: "status-filter" as const,
        filterKey: "true",
      },
      {
        title: t("adminLogs.operationLogs.status.failed"),
        value: failedCount,
        variant: "destructive" as const,
        type: "status-filter" as const,
        filterKey: "false",
      },
    ];
  }, [logs, t]);

  const loadLogs = async (page = currentPage, filters = columnFilters) => {
    try {
      setLoading(true);

      // Convert DataTable filters to API parameters
      const apiParams: any = {
        page,
        page_size: pageSize,
      };

      // Handle column filters
      filters.forEach((filter) => {
        if (filter.id === "username" && filter.value) {
          apiParams.username = filter.value;
        } else if (filter.id === "operation" && Array.isArray(filter.value) && filter.value.length > 0) {
          apiParams.operation = filter.value[0]; // API expects single operation
        } else if (filter.id === "success" && Array.isArray(filter.value) && filter.value.length > 0) {
          apiParams.success = filter.value[0] === "true";
        } else if (filter.id === "operation_time" && filter.value) {
          const { startDate, endDate } = filter.value as { startDate?: Date; endDate?: Date };
          if (startDate) {
            apiParams.start_time = startDate.toISOString().split('T')[0];
          }
          if (endDate) {
            apiParams.end_time = endDate.toISOString().split('T')[0];
          }
        }
      });

      const response = await apiService.adminLogs.getOperationLogs(apiParams);

      setLogs(response.logs);
      setTotal(response.total);
      setTotalPages(response.total_pages);
      setCurrentPage(page);
    } catch (err: any) {
      logError(err, "Failed to load operation logs");
      console.error("Failed to load operation logs:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleMetricCardClick = (metric: typeof metrics[0]) => {
    let newColumnFilters = [...columnFilters];
    
    if (metric.type === "filter") {
      const currentFilter = columnFilters.find(f => f.id === "operation");
      const currentValues = (currentFilter?.value as string[]) || [];
      const isActive = currentValues.includes(metric.filterKey);
      
      newColumnFilters = columnFilters.filter(f => f.id !== "operation");
      if (!isActive) {
        newColumnFilters.push({ id: "operation", value: [metric.filterKey] });
      }
    } else if (metric.type === "status-filter") {
      const currentFilter = columnFilters.find(f => f.id === "success");
      const currentValues = (currentFilter?.value as string[]) || [];
      const isActive = currentValues.includes(metric.filterKey);
      
      newColumnFilters = columnFilters.filter(f => f.id !== "success");
      if (!isActive) {
        newColumnFilters.push({ id: "success", value: [metric.filterKey] });
      }
    }
    
    setColumnFilters(newColumnFilters);
    loadLogs(1, newColumnFilters); // Reset to page 1 when filtering
  };

  const handlePageChange = (page: number) => {
    loadLogs(page);
  };

  const handleViewDetail = async (id: number) => {
    try {
      const response = await apiService.adminLogs.getOperationLog(id);
      setSelectedLog(response.log);
      setDetailDialogOpen(true);
    } catch (err: any) {
      logError(err, "Failed to load operation log detail");
      console.error("Failed to load operation log detail:", err);
      toast.error(t("adminLogs.operationLogs.messages.loadDetailFailed"));
    }
  };

  const handleCloseDetail = () => {
    setDetailDialogOpen(false);
    setSelectedLog(null);
  };

  // Create columns with the view detail handler
  const columns = useAdminOperationLogColumns({ onViewDetail: handleViewDetail });

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
          <SectionTitle>{t("adminLogs.operationLogs.overview.title")}</SectionTitle>
          <SectionDescription>
            {t("adminLogs.operationLogs.overview.description")}
          </SectionDescription>
        </SectionHeader>
        <MetricCardGroup>
          {metrics.map((metric) => {
            const currentFilter = columnFilters.find(f => 
              f.id === (metric.type === "filter" ? "operation" : "success")
            );
            const currentValues = (currentFilter?.value as string[]) || [];
            const isFilterActive = currentValues.includes(metric.filterKey);

            const Icon = isFilterActive ? CheckCircle : Filter;

            return (
              <MetricCardButton
                key={metric.title}
                variant={metric.variant}
                onClick={() => handleMetricCardClick(metric)}
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
            toolbarComponent={AdminOperationLogDataTableToolbar}
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

      <Dialog open={detailDialogOpen} onOpenChange={setDetailDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="text-foreground">
              {t("adminLogs.operationLogs.detail.title")}
            </DialogTitle>
            <DialogDescription className="text-muted-foreground">
              {t("adminLogs.operationLogs.detail.description")}
            </DialogDescription>
          </DialogHeader>

          {selectedLog && (
            <div className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium text-foreground">
                    {t("adminLogs.operationLogs.columns.id")}
                  </label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {selectedLog.id}
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium text-foreground">
                    {t("adminLogs.operationLogs.columns.operation")}
                  </label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {selectedLog.operation}
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium text-foreground">
                    {t("adminLogs.operationLogs.columns.resource")}
                  </label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {selectedLog.resource || "N/A"}
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium text-foreground">
                    {t("adminLogs.operationLogs.columns.username")}
                  </label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {selectedLog.username || "N/A"}
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium text-foreground">
                    {t("adminLogs.operationLogs.columns.time")}
                  </label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {new Date(selectedLog.operation_time).toLocaleString()}
                  </p>
                </div>
              </div>

              <div>
                <label className="text-sm font-medium text-foreground">
                  {t("adminLogs.operationLogs.columns.description")}
                </label>
                <p className="text-sm text-muted-foreground mt-1 whitespace-pre-wrap">
                  {selectedLog.details || "N/A"}
                </p>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              className="text-foreground hover:text-foreground"
              onClick={handleCloseDetail}
            >
              {t("common.close")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
};
