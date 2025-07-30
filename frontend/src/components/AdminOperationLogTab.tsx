import React, { useState, useEffect } from "react";
import { toast } from "sonner";
import { AdminOperationLogList } from "./AdminOperationLogList";
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
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [selectedLog, setSelectedLog] = useState<AdminOperationLog | null>(
    null
  );

  const pageSize = 10;

  const loadLogs = async (params?: AdminOperationLogListParams) => {
    try {
      setLoading(true);

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

  useEffect(() => {
    loadLogs();
  }, []);

  return (
    <>
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
