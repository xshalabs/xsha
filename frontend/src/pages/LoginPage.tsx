import React from "react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { LoginForm } from "@/components/LoginForm";
import { Logo } from "@/components/Logo";

export const LoginPage: React.FC = () => {
  usePageTitle("common.pageTitle.login");

  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <a
          href="#"
          className="flex items-center justify-center self-center"
        >
          <Logo />
        </a>
        <LoginForm />
      </div>
    </div>
  );
};
