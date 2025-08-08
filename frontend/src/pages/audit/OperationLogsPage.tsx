import React from "react";
import { useTranslation } from "react-i18next";
import { usePageTitle } from "@/hooks/usePageTitle";
import { AdminOperationLogTab } from "@/components/AdminOperationLogTab";

export const OperationLogsPage: React.FC = () => {
  const { t } = useTranslation();

  usePageTitle("common.pageTitle.operationLogs");

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              {t("navigation.audit.operationLogs")}
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {t("adminLogs.operationLogs.description")}
            </p>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="px-4 sm:px-0">
          <AdminOperationLogTab />
        </div>
      </div>
    </div>
  );
};
