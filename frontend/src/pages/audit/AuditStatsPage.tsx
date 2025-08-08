import React from "react";
import { useTranslation } from "react-i18next";
import { usePageTitle } from "@/hooks/usePageTitle";
import { AdminStatsTab } from "@/components/AdminStatsTab";

export const AuditStatsPage: React.FC = () => {
  const { t } = useTranslation();

  usePageTitle("common.pageTitle.auditStats");

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              {t("navigation.audit.stats")}
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {t("adminLogs.stats.description")}
            </p>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="px-4 sm:px-0">
          <AdminStatsTab />
        </div>
      </div>
    </div>
  );
};
