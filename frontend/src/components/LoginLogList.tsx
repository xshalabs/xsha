import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  ChevronLeft,
  ChevronRight,
  Filter,
  RefreshCw,
  Calendar,
  User,
  Shield,
  CheckCircle,
  XCircle,
  Globe,
} from "lucide-react";
import type { LoginLog, LoginLogListParams } from "@/types/admin-logs";

interface LoginLogListProps {
  logs: LoginLog[];
  loading: boolean;
  currentPage: number;
  totalPages: number;
  total: number;
  filters: LoginLogListParams;
  onPageChange: (page: number) => void;
  onFiltersChange: (filters: LoginLogListParams) => void;
  onRefresh: () => void;
}

export const LoginLogList: React.FC<LoginLogListProps> = ({
  logs,
  loading,
  currentPage,
  totalPages,
  total,
  filters,
  onPageChange,
  onFiltersChange,
  onRefresh,
}) => {
  const { t } = useTranslation();
  const [showFilters, setShowFilters] = useState(false);
  const [localFilters, setLocalFilters] = useState<LoginLogListParams>(filters);

  const handleFilterChange = (
    key: keyof LoginLogListParams,
    value: string | undefined
  ) => {
    setLocalFilters((prev) => ({
      ...prev,
      [key]: value === "" ? undefined : value,
    }));
  };

  const applyFilters = () => {
    onFiltersChange({
      ...localFilters,
      page: 1, // 重置到第一页
    });
  };

  const resetFilters = () => {
    const emptyFilters: LoginLogListParams = {};
    setLocalFilters(emptyFilters);
    onFiltersChange(emptyFilters);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const getStatusIcon = (success: boolean) => {
    return success ? (
      <CheckCircle className="w-4 h-4 text-green-500" />
    ) : (
      <XCircle className="w-4 h-4 text-red-500" />
    );
  };

  const getReasonText = (reason: string) => {
    if (!reason) return "";
    const reasonKey = `adminLogs.loginLogs.reasons.${reason}`;
    const translatedReason = t(reasonKey);
    return translatedReason === reasonKey ? reason : translatedReason;
  };

  if (loading && logs.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">{t("common.loading")}</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div className="text-sm text-foreground">
          {t("adminLogs.common.total")} {total} {t("adminLogs.common.items")}
        </div>
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="ghost"
            className="text-foreground"
            onClick={() => setShowFilters(!showFilters)}
          >
            <Filter className="w-4 h-4 mr-2" />
            {t("adminLogs.common.search")}
          </Button>
          <Button
            onClick={onRefresh}
            disabled={loading}
            size="sm"
            variant="ghost"
            className="text-foreground"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            {t("adminLogs.common.refresh")}
          </Button>
        </div>
      </div>

      {/* 筛选器 */}
      {showFilters && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">
              {t("adminLogs.loginLogs.filters.all")}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="flex flex-col gap-3">
                <Label htmlFor="username">
                  {t("adminLogs.loginLogs.filters.username")}
                </Label>
                <Input
                  id="username"
                  value={localFilters.username || ""}
                  onChange={(e) =>
                    handleFilterChange("username", e.target.value)
                  }
                  placeholder={t("adminLogs.loginLogs.filters.username")}
                />
              </div>
            </div>

            <div className="flex gap-2 mt-4">
              <Button onClick={applyFilters}>
                {t("adminLogs.common.apply")}
              </Button>
              <Button variant="outline" onClick={resetFilters}>
                {t("adminLogs.common.reset")}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 日志列表 */}
      <div className="space-y-2">
        {logs.length === 0 ? (
          <Card>
            <CardContent className="text-center py-8">
              <p className="text-gray-500">
                {t("adminLogs.loginLogs.messages.noLogs")}
              </p>
            </CardContent>
          </Card>
        ) : (
          logs.map((log) => (
            <Card key={log.id} className="hover:shadow-md transition-shadow">
              <CardContent className="p-4">
                <div className="flex flex-col space-y-3 lg:flex-row lg:items-center lg:space-y-0">
                  <div className="flex flex-col space-y-2 lg:flex-row lg:items-center lg:space-x-4 lg:space-y-0 flex-1">
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(log.success)}
                      <span
                        className={`font-medium ${
                          log.success ? "text-green-600" : "text-red-600"
                        }`}
                      >
                        {log.success
                          ? t("adminLogs.loginLogs.status.success")
                          : t("adminLogs.loginLogs.status.failed")}
                      </span>
                    </div>

                    <div className="flex items-center space-x-2">
                      <User className="w-4 h-4 text-gray-400" />
                      <span className="text-sm font-medium break-all">
                        {log.username}
                      </span>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Globe className="w-4 h-4 text-gray-400" />
                      <span className="text-sm text-gray-600 break-all">{log.ip}</span>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Calendar className="w-4 h-4 text-gray-400" />
                      <span className="text-sm text-gray-600">
                        {formatDate(log.login_time)}
                      </span>
                    </div>
                  </div>
                </div>

                {!log.success && log.reason && (
                  <div className="mt-3 text-sm text-red-600 bg-red-50 p-3 rounded break-words">
                    <strong>{t("adminLogs.loginLogs.columns.reason")}:</strong>{" "}
                    {getReasonText(log.reason)}
                  </div>
                )}

                {log.user_agent && (
                  <div className="mt-3 text-xs text-gray-500 break-words">
                    <strong>
                      {t("adminLogs.loginLogs.columns.userAgent")}:
                    </strong>{" "}
                    {log.user_agent}
                  </div>
                )}
              </CardContent>
            </Card>
          ))
        )}
      </div>

      {/* 分页 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-gray-600">
            {t("adminLogs.common.page")} {currentPage} / {totalPages}
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              className="text-foreground"
              onClick={() => onPageChange(currentPage - 1)}
              disabled={currentPage <= 1}
            >
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="text-foreground"
              onClick={() => onPageChange(currentPage + 1)}
              disabled={currentPage >= totalPages}
            >
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};
