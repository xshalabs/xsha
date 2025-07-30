import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DatePicker } from "@/components/ui/date-picker";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  ChevronLeft,
  ChevronRight,
  Filter,
  RefreshCw,
  Calendar,
  User,
  Activity,
  CheckCircle,
  XCircle,
  Eye,
} from "lucide-react";
import type {
  AdminOperationLog,
  AdminOperationLogListParams,
  AdminOperationType,
} from "@/types/admin-logs";

interface AdminOperationLogListProps {
  logs: AdminOperationLog[];
  loading: boolean;
  currentPage: number;
  totalPages: number;
  total: number;
  filters: AdminOperationLogListParams;
  onPageChange: (page: number) => void;
  onFiltersChange: (filters: AdminOperationLogListParams) => void;
  onRefresh: () => void;
  onViewDetail: (id: number) => void;
}

export const AdminOperationLogList: React.FC<AdminOperationLogListProps> = ({
  logs,
  loading,
  currentPage,
  totalPages,
  total,
  filters,
  onPageChange,
  onFiltersChange,
  onRefresh,
  onViewDetail,
}) => {
  const { t } = useTranslation();
  const [showFilters, setShowFilters] = useState(false);
  const [localFilters, setLocalFilters] =
    useState<AdminOperationLogListParams>(filters);

  const handleFilterChange = (
    key: keyof AdminOperationLogListParams,
    value: string | boolean | undefined
  ) => {
    setLocalFilters((prev) => ({
      ...prev,
      [key]: value === "" ? undefined : value,
    }));
  };

  const applyFilters = () => {
    onFiltersChange({
      ...localFilters,
      page: 1,
    });
  };

  const resetFilters = () => {
    const emptyFilters: AdminOperationLogListParams = {};
    setLocalFilters(emptyFilters);
    onFiltersChange(emptyFilters);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const getOperationIcon = (operation: AdminOperationType) => {
    switch (operation) {
      case "create":
        return <div className="w-2 h-2 bg-green-500 rounded-full" />;
      case "read":
        return <div className="w-2 h-2 bg-blue-500 rounded-full" />;
      case "update":
        return <div className="w-2 h-2 bg-yellow-500 rounded-full" />;
      case "delete":
        return <div className="w-2 h-2 bg-red-500 rounded-full" />;
      case "login":
        return <div className="w-2 h-2 bg-purple-500 rounded-full" />;
      case "logout":
        return <div className="w-2 h-2 bg-gray-500 rounded-full" />;
      default:
        return <div className="w-2 h-2 bg-gray-400 rounded-full" />;
    }
  };

  const getStatusIcon = (success: boolean) => {
    return success ? (
      <CheckCircle className="w-4 h-4 text-green-500" />
    ) : (
      <XCircle className="w-4 h-4 text-red-500" />
    );
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

      {showFilters && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">
              {t("adminLogs.operationLogs.filters.all")}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <div className="flex flex-col gap-3">
                <Label htmlFor="username">
                  {t("adminLogs.operationLogs.filters.username")}
                </Label>
                <Input
                  id="username"
                  value={localFilters.username || ""}
                  onChange={(e) =>
                    handleFilterChange("username", e.target.value)
                  }
                  placeholder={t("adminLogs.operationLogs.filters.username")}
                />
              </div>

              <div className="flex flex-col gap-3">
                <Label htmlFor="resource">
                  {t("adminLogs.operationLogs.filters.resource")}
                </Label>
                <Input
                  id="resource"
                  value={localFilters.resource || ""}
                  onChange={(e) =>
                    handleFilterChange("resource", e.target.value)
                  }
                  placeholder={t("adminLogs.operationLogs.filters.resource")}
                />
              </div>

              <div className="flex flex-col gap-3">
                <Label htmlFor="operation">
                  {t("adminLogs.operationLogs.filters.operation")}
                </Label>
                <Select
                  value={localFilters.operation || "all"}
                  onValueChange={(value) =>
                    handleFilterChange(
                      "operation",
                      value === "all"
                        ? undefined
                        : (value as AdminOperationType)
                    )
                  }
                >
                  <SelectTrigger className="w-full">
                    <SelectValue
                      placeholder={t("adminLogs.operationLogs.filters.all")}
                    />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">
                      {t("adminLogs.operationLogs.filters.all")}
                    </SelectItem>
                    <SelectItem value="create">
                      {t("adminLogs.operationLogs.operations.create")}
                    </SelectItem>
                    <SelectItem value="read">
                      {t("adminLogs.operationLogs.operations.read")}
                    </SelectItem>
                    <SelectItem value="update">
                      {t("adminLogs.operationLogs.operations.update")}
                    </SelectItem>
                    <SelectItem value="delete">
                      {t("adminLogs.operationLogs.operations.delete")}
                    </SelectItem>
                    <SelectItem value="login">
                      {t("adminLogs.operationLogs.operations.login")}
                    </SelectItem>
                    <SelectItem value="logout">
                      {t("adminLogs.operationLogs.operations.logout")}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex flex-col gap-3">
                <Label htmlFor="success">
                  {t("adminLogs.operationLogs.filters.success")}
                </Label>
                <Select
                  value={
                    localFilters.success === undefined
                      ? "all"
                      : localFilters.success.toString()
                  }
                  onValueChange={(value) =>
                    handleFilterChange(
                      "success",
                      value === "all" ? undefined : value === "true"
                    )
                  }
                >
                  <SelectTrigger className="w-full">
                    <SelectValue
                      placeholder={t("adminLogs.operationLogs.filters.all")}
                    />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">
                      {t("adminLogs.operationLogs.filters.all")}
                    </SelectItem>
                    <SelectItem value="true">
                      {t("adminLogs.operationLogs.status.success")}
                    </SelectItem>
                    <SelectItem value="false">
                      {t("adminLogs.operationLogs.status.failed")}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div>
                <DatePicker
                  id="start_time"
                  label={t("adminLogs.operationLogs.filters.startDate")}
                  placeholder={t("adminLogs.operationLogs.filters.startDate")}
                  value={
                    localFilters.start_time
                      ? new Date(localFilters.start_time)
                      : undefined
                  }
                  onChange={(date) =>
                    handleFilterChange(
                      "start_time",
                      date
                        ? `${date.getFullYear()}-${String(
                            date.getMonth() + 1
                          ).padStart(2, "0")}-${String(date.getDate()).padStart(
                            2,
                            "0"
                          )}`
                        : ""
                    )
                  }
                />
              </div>

              <div>
                <DatePicker
                  id="end_time"
                  label={t("adminLogs.operationLogs.filters.endDate")}
                  placeholder={t("adminLogs.operationLogs.filters.endDate")}
                  value={
                    localFilters.end_time
                      ? new Date(localFilters.end_time)
                      : undefined
                  }
                  onChange={(date) =>
                    handleFilterChange(
                      "end_time",
                      date
                        ? `${date.getFullYear()}-${String(
                            date.getMonth() + 1
                          ).padStart(2, "0")}-${String(date.getDate()).padStart(
                            2,
                            "0"
                          )}`
                        : ""
                    )
                  }
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

      <div className="space-y-2">
        {logs.length === 0 ? (
          <Card>
            <CardContent className="text-center py-8">
              <p className="text-gray-500">
                {t("adminLogs.operationLogs.messages.noLogs")}
              </p>
            </CardContent>
          </Card>
        ) : (
          logs.map((log) => (
            <Card key={log.id} className="hover:shadow-md transition-shadow">
              <CardContent className="p-4">
                <div className="flex flex-col space-y-3 lg:flex-row lg:items-center lg:justify-between lg:space-y-0">
                  <div className="flex flex-col space-y-2 lg:flex-row lg:items-center lg:space-x-4 lg:space-y-0 flex-1">
                    <div className="flex items-center space-x-2">
                      {getOperationIcon(log.operation)}
                      <span className="font-medium">
                        {t(
                          `adminLogs.operationLogs.operations.${log.operation}`
                        )}
                      </span>
                    </div>

                    <div className="flex items-center space-x-2">
                      <User className="w-4 h-4 text-gray-400" />
                      <span className="text-sm break-all">{log.username}</span>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Activity className="w-4 h-4 text-gray-400" />
                      <span className="text-sm break-all">{log.resource}</span>
                    </div>

                    <div className="flex items-center space-x-2">
                      {getStatusIcon(log.success)}
                      <span
                        className={`text-sm ${
                          log.success ? "text-green-600" : "text-red-600"
                        }`}
                      >
                        {log.success
                          ? t("adminLogs.operationLogs.status.success")
                          : t("adminLogs.operationLogs.status.failed")}
                      </span>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Calendar className="w-4 h-4 text-gray-400" />
                      <span className="text-sm text-gray-600">
                        {formatDate(log.operation_time)}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center justify-end lg:justify-start">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onViewDetail(log.id)}
                    >
                      <Eye className="w-4 h-4 mr-1" />
                      {t("adminLogs.common.detail")}
                    </Button>
                  </div>
                </div>

                {log.description && (
                  <div className="mt-3 text-sm text-gray-600 break-words">
                    {log.description}
                  </div>
                )}

                {!log.success && log.error_msg && (
                  <div className="mt-3 text-sm text-red-600 bg-red-50 p-3 rounded break-words">
                    {log.error_msg}
                  </div>
                )}
              </CardContent>
            </Card>
          ))
        )}
      </div>

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
