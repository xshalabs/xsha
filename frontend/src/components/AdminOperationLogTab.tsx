import React, { useState, useEffect } from "react";
import { AdminOperationLogList } from "./AdminOperationLogList";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import { useTranslation } from "react-i18next";
import type {
  AdminOperationLog,
  AdminOperationLogListParams,
} from "@/types/admin-logs";

export const AdminOperationLogTab: React.FC = () => {
  const { t } = useTranslation();
  const [logs, setLogs] = useState<AdminOperationLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [filters, setFilters] = useState<AdminOperationLogListParams>({});

  const pageSize = 10;

  const loadLogs = async (params?: AdminOperationLogListParams) => {
    try {
      setLoading(true);
      
      // 使用传入的 params，如果没有则使用当前的 filters
      const requestParams = params ?? filters;
      
      const response = await apiService.adminLogs.getOperationLogs({
        page: params?.page ?? currentPage,
        page_size: pageSize,
        ...requestParams,
      });

      setLogs(response.logs);
      setTotal(response.total);
      setTotalPages(response.total_pages);
      if (params?.page) {
        setCurrentPage(params.page);
      }
    } catch (err: any) {
      logError(err, "Failed to load operation logs");
      console.error("Failed to load operation logs:", err);
    } finally {
      setLoading(false);
    }
  };

  const handlePageChange = (page: number) => {
    loadLogs({ ...filters, page });
  };

  const handleFiltersChange = (newFilters: AdminOperationLogListParams) => {
    setFilters(newFilters);
    loadLogs({ ...newFilters, page: 1 });
  };

  const handleViewDetail = async (id: number) => {
    try {
      const response = await apiService.adminLogs.getOperationLog(id);
      const logInfo = [
        `${t("adminLogs.operationLogs.columns.id")}: ${response.log.id}`,
        `${t("adminLogs.operationLogs.columns.operation")}: ${
          response.log.operation
        }`,
        `${t("adminLogs.operationLogs.columns.resource")}: ${
          response.log.resource || "N/A"
        }`,
        `${t("adminLogs.operationLogs.columns.username")}: ${
          response.log.username || "N/A"
        }`,
        `${t("adminLogs.operationLogs.columns.description")}: ${
          response.log.details || "N/A"
        }`,
        `${t("adminLogs.operationLogs.columns.time")}: ${new Date(
          response.log.operation_time
        ).toLocaleString()}`,
      ].join("\n\n");

      alert(logInfo);
    } catch (err: any) {
      logError(err, "Failed to load operation log detail");
      console.error("Failed to load operation log detail:", err);
    }
  };

  useEffect(() => {
    loadLogs();
  }, []);

  return (
    <AdminOperationLogList
      logs={logs}
      loading={loading}
      currentPage={currentPage}
      totalPages={totalPages}
      total={total}
      filters={filters}
      onPageChange={handlePageChange}
      onFiltersChange={handleFiltersChange}
      onRefresh={() => loadLogs()}
      onViewDetail={handleViewDetail}
    />
  );
}; 