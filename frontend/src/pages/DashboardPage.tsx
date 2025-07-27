import React from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useAuth } from "@/contexts/AuthContext";
import { usePageTitle } from "@/hooks/usePageTitle";
import { Button } from "@/components/ui/button";
import { ROUTES } from "@/lib/constants";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export const DashboardPage: React.FC = () => {
  const { t } = useTranslation();
  const { user } = useAuth();
  const navigate = useNavigate();

  usePageTitle("common.pageTitle.dashboard");

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              {t("dashboard.title")}
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {t("dashboard.welcome")}, {user}!
            </p>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <Card
              className="cursor-pointer hover:shadow-lg transition-shadow"
              onClick={() => navigate(ROUTES.projects)}
            >
              <CardHeader>
                <CardTitle>{t("dashboard.projectManagement.title")}</CardTitle>
                <CardDescription>
                  {t("dashboard.projectManagement.description")}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  {t("dashboard.projectManagement.content")}
                </p>
                <Button
                  className="mt-4"
                  onClick={(e) => {
                    e.stopPropagation();
                    navigate(ROUTES.projects);
                  }}
                >
                  {t("projects.title")}
                </Button>
              </CardContent>
            </Card>

            <Card
              className="cursor-pointer hover:shadow-lg transition-shadow"
              onClick={() => navigate(ROUTES.gitCredentials)}
            >
              <CardHeader>
                <CardTitle>{t("dashboard.gitCredentials.title")}</CardTitle>
                <CardDescription>
                  {t("dashboard.gitCredentials.description")}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  {t("dashboard.gitCredentials.content")}
                </p>
                <Button
                  className="mt-4"
                  onClick={(e) => {
                    e.stopPropagation();
                    navigate(ROUTES.gitCredentials);
                  }}
                >
                  {t("dashboard.gitCredentials.manage")}
                </Button>
              </CardContent>
            </Card>

            <Card
              className="cursor-pointer hover:shadow-lg transition-shadow"
              onClick={() => navigate(ROUTES.adminLogs)}
            >
              <CardHeader>
                <CardTitle>{t("dashboard.adminLogs.title")}</CardTitle>
                <CardDescription>
                  {t("dashboard.adminLogs.description")}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  {t("dashboard.adminLogs.content")}
                </p>
                <Button
                  className="mt-4"
                  onClick={(e) => {
                    e.stopPropagation();
                    navigate(ROUTES.adminLogs);
                  }}
                >
                  {t("dashboard.adminLogs.manage")}
                </Button>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
};
