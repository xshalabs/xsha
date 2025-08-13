import React from "react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { AdminOperationLogList } from "@/components/AdminOperationLogList";
import { SectionGroup } from "@/components/content/section";

export const OperationLogsPage: React.FC = () => {
  usePageTitle("adminLogs.operationLogs.title");

  return (
    <SectionGroup>
      <AdminOperationLogList />
    </SectionGroup>
  );
};
