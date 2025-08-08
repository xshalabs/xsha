import React from "react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { AdminOperationLogTab } from "@/components/AdminOperationLogTab";
import { SectionGroup } from "@/components/content/section";

export const OperationLogsPage: React.FC = () => {
  usePageTitle("common.pageTitle.operationLogs");

  return (
    <SectionGroup>
      <AdminOperationLogTab />
    </SectionGroup>
  );
};
