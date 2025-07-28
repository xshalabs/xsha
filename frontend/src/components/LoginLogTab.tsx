import React, { useState, useEffect } from "react";
import { LoginLogList } from "./LoginLogList";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { LoginLog, LoginLogListParams } from "@/types/admin-logs";

export const LoginLogTab: React.FC = () => {
  const [logs, setLogs] = useState<LoginLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [filters, setFilters] = useState<LoginLogListParams>({});

  const pageSize = 10;

  const loadLogs = async (params?: LoginLogListParams) => {
    try {
      setLoading(true);
      
      // 使用传入的 params，如果没有则使用当前的 filters
      const requestParams = params ?? filters;
      
      const response = await apiService.adminLogs.getLoginLogs({
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
      logError(err, "Failed to load login logs");
      console.error("Failed to load login logs:", err);
    } finally {
      setLoading(false);
    }
  };

  const handlePageChange = (page: number) => {
    loadLogs({ ...filters, page });
  };

  const handleFiltersChange = (newFilters: LoginLogListParams) => {
    setFilters(newFilters);
    loadLogs({ ...newFilters, page: 1 });
  };

  useEffect(() => {
    loadLogs();
  }, []);

  return (
    <LoginLogList
      logs={logs}
      loading={loading}
      currentPage={currentPage}
      totalPages={totalPages}
      total={total}
      filters={filters}
      onPageChange={handlePageChange}
      onFiltersChange={handleFiltersChange}
      onRefresh={() => loadLogs()}
    />
  );
}; 