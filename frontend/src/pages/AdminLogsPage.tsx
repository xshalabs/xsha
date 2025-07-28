import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { usePageTitle } from "@/hooks/usePageTitle";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { AdminOperationLogList } from "@/components/AdminOperationLogList";
import { LoginLogList } from "@/components/LoginLogList";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import { FileText, Shield, TrendingUp } from "lucide-react";
import type {
  AdminOperationLog,
  LoginLog,
  AdminOperationLogListParams,
  LoginLogListParams,
  AdminOperationStatsResponse,
} from "@/types/admin-logs";

type TabType = "operationLogs" | "loginLogs" | "stats";

export const AdminLogsPage: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<TabType>("operationLogs");

  usePageTitle("common.pageTitle.adminLogs");

  const [operationLogs, setOperationLogs] = useState<AdminOperationLog[]>([]);
  const [operationLoading, setOperationLoading] = useState(false);
  const [operationCurrentPage, setOperationCurrentPage] = useState(1);
  const [operationTotalPages, setOperationTotalPages] = useState(1);
  const [operationTotal, setOperationTotal] = useState(0);
  const [operationFilters, setOperationFilters] =
    useState<AdminOperationLogListParams>({});

  const [loginLogs, setLoginLogs] = useState<LoginLog[]>([]);
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginCurrentPage, setLoginCurrentPage] = useState(1);
  const [loginTotalPages, setLoginTotalPages] = useState(1);
  const [loginTotal, setLoginTotal] = useState(0);
  const [loginFilters, setLoginFilters] = useState<LoginLogListParams>({});

  const [stats, setStats] = useState<AdminOperationStatsResponse | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);

  const pageSize = 10;

  const loadOperationLogs = async (params?: AdminOperationLogListParams) => {
    try {
      setOperationLoading(true);
      const response = await apiService.adminLogs.getOperationLogs({
        page: operationCurrentPage,
        page_size: pageSize,
        ...operationFilters,
        ...params,
      });

      setOperationLogs(response.logs);
      setOperationTotal(response.total);
      setOperationTotalPages(response.total_pages);
      if (params?.page) {
        setOperationCurrentPage(params.page);
      }
    } catch (err: any) {
      logError(err, "Failed to load operation logs");
      console.error("Failed to load operation logs:", err);
    } finally {
      setOperationLoading(false);
    }
  };

  const loadLoginLogs = async (params?: LoginLogListParams) => {
    try {
      setLoginLoading(true);
      const response = await apiService.adminLogs.getLoginLogs({
        page: loginCurrentPage,
        page_size: pageSize,
        ...loginFilters,
        ...params,
      });

      setLoginLogs(response.logs);
      setLoginTotal(response.total);
      setLoginTotalPages(response.total_pages);
      if (params?.page) {
        setLoginCurrentPage(params.page);
      }
    } catch (err: any) {
      logError(err, "Failed to load login logs");
      console.error("Failed to load login logs:", err);
    } finally {
      setLoginLoading(false);
    }
  };

  const loadStats = async () => {
    try {
      setStatsLoading(true);
      const response = await apiService.adminLogs.getOperationStats();
      setStats(response);
    } catch (err: any) {
      logError(err, "Failed to load stats");
      console.error("Failed to load stats:", err);
    } finally {
      setStatsLoading(false);
    }
  };

  const handleOperationPageChange = (page: number) => {
    loadOperationLogs({ ...operationFilters, page });
  };

  const handleOperationFiltersChange = (
    filters: AdminOperationLogListParams
  ) => {
    setOperationFilters(filters);
    loadOperationLogs({ ...filters, page: 1 });
  };

  const handleLoginPageChange = (page: number) => {
    loadLoginLogs({ ...loginFilters, page });
  };

  const handleLoginFiltersChange = (filters: LoginLogListParams) => {
    setLoginFilters(filters);
    loadLoginLogs({ ...filters, page: 1 });
  };

  const handleViewOperationDetail = async (id: number) => {
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
    switch (activeTab) {
      case "operationLogs":
        if (operationLogs.length === 0) {
          loadOperationLogs();
        }
        break;
      case "loginLogs":
        if (loginLogs.length === 0) {
          loadLoginLogs();
        }
        break;
      case "stats":
        if (!stats) {
          loadStats();
        }
        break;
    }
  }, [activeTab]);

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              {t("adminLogs.title")}
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {t("adminLogs.description")}
            </p>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
        <div className="px-4 sm:px-0">
          <Tabs
            value={activeTab}
            onValueChange={(value) => setActiveTab(value as TabType)}
          >
            <TabsList className="mb-6">
              <TabsTrigger
                value="operationLogs"
                className="flex items-center gap-2"
              >
                <FileText className="w-4 h-4" />
                {t("adminLogs.operationLogs.title")}
              </TabsTrigger>
              <TabsTrigger
                value="loginLogs"
                className="flex items-center gap-2"
              >
                <Shield className="w-4 h-4" />
                {t("adminLogs.loginLogs.title")}
              </TabsTrigger>
              <TabsTrigger value="stats" className="flex items-center gap-2">
                <TrendingUp className="w-4 h-4" />
                {t("adminLogs.stats.title")}
              </TabsTrigger>
            </TabsList>

            <TabsContent value="operationLogs">
              <AdminOperationLogList
                logs={operationLogs}
                loading={operationLoading}
                currentPage={operationCurrentPage}
                totalPages={operationTotalPages}
                total={operationTotal}
                filters={operationFilters}
                onPageChange={handleOperationPageChange}
                onFiltersChange={handleOperationFiltersChange}
                onRefresh={() => loadOperationLogs()}
                onViewDetail={handleViewOperationDetail}
              />
            </TabsContent>

            <TabsContent value="loginLogs">
              <LoginLogList
                logs={loginLogs}
                loading={loginLoading}
                currentPage={loginCurrentPage}
                totalPages={loginTotalPages}
                total={loginTotal}
                filters={loginFilters}
                onPageChange={handleLoginPageChange}
                onFiltersChange={handleLoginFiltersChange}
                onRefresh={() => loadLoginLogs()}
              />
            </TabsContent>

            <TabsContent value="stats">
              <div className="space-y-6">
                {statsLoading ? (
                  <div className="flex items-center justify-center h-64">
                    <div className="text-muted-foreground">
                      {t("common.loading")}
                    </div>
                  </div>
                ) : stats ? (
                  <div className="space-y-6">
                    <Card>
                      <CardHeader>
                        <CardTitle>
                          {t("adminLogs.stats.operationStats")}
                        </CardTitle>
                        <CardDescription>
                          {t("adminLogs.stats.timeRange")}: {stats.start_time} ~{" "}
                          {stats.end_time}
                        </CardDescription>
                      </CardHeader>
                      <CardContent>
                        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
                          {Object.entries(stats.operation_stats).map(
                            ([operation, count]) => (
                              <div
                                key={operation}
                                className="text-center p-4 bg-muted rounded-lg"
                              >
                                <div className="text-2xl font-bold text-primary">
                                  {count}
                                </div>
                                <div className="text-sm text-muted-foreground">
                                  {t(
                                    `adminLogs.operationLogs.operations.${operation}`
                                  )}
                                </div>
                              </div>
                            )
                          )}
                        </div>
                      </CardContent>
                    </Card>

                    <Card>
                      <CardHeader>
                        <CardTitle>
                          {t("adminLogs.stats.resourceStats")}
                        </CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                          {Object.entries(stats.resource_stats).map(
                            ([resource, count]) => (
                              <div
                                key={resource}
                                className="text-center p-4 bg-muted rounded-lg"
                              >
                                <div className="text-2xl font-bold text-accent">
                                  {count}
                                </div>
                                <div className="text-sm text-muted-foreground capitalize">
                                  {resource}
                                </div>
                              </div>
                            )
                          )}
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <p className="text-muted-foreground">
                      {t("adminLogs.stats.noStatsAvailable")}
                    </p>
                  </div>
                )}
              </div>
            </TabsContent>
          </Tabs>
        </div>
      </div>
    </div>
  );
};
