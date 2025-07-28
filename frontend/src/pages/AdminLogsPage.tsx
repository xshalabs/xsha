import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { usePageTitle } from "@/hooks/usePageTitle";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { AdminOperationLogTab } from "@/components/AdminOperationLogTab";
import { LoginLogTab } from "@/components/LoginLogTab";
import { AdminStatsTab } from "@/components/AdminStatsTab";
import { FileText, Shield, TrendingUp } from "lucide-react";

type TabType = "operationLogs" | "loginLogs" | "stats";

export const AdminLogsPage: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<TabType>("operationLogs");

  usePageTitle("common.pageTitle.adminLogs");

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
              <AdminOperationLogTab />
            </TabsContent>

            <TabsContent value="loginLogs">
              <LoginLogTab />
            </TabsContent>

            <TabsContent value="stats">
              <AdminStatsTab />
            </TabsContent>
          </Tabs>
        </div>
      </div>
    </div>
  );
};
