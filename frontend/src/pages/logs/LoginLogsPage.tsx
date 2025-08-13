import React from "react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { LoginLogList } from "@/components/LoginLogList";
import { SectionGroup } from "@/components/content/section";

export const LoginLogsPage: React.FC = () => {
  usePageTitle("adminLogs.loginLogs.title");

  return (
    <SectionGroup>
      <LoginLogList />
    </SectionGroup>
  );
};
