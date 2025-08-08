import React from "react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { AdminStatsTab } from "@/components/AdminStatsTab";
import { SectionGroup } from "@/components/content/section";

export const AuditStatsPage: React.FC = () => {
  usePageTitle("adminLogs.stats.title");

  return (
    <SectionGroup>
      <AdminStatsTab />
    </SectionGroup>
  );
};
