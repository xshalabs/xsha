import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useAuth } from "@/contexts/AuthContext";
import { usePageTitle } from "@/hooks/usePageTitle";
import { Button } from "@/components/ui/button";
import { ROUTES } from "@/lib/constants";
import {
  SectionGroup,
  Section,
  SectionHeader,
  SectionHeaderRow,
  SectionTitle,
  SectionDescription,
  EmptyStateContainer,
  EmptyStateTitle,
  EmptyStateDescription,
} from "@/components/content";
import {
  MetricCard,
  MetricCardGroup,
  MetricCardHeader,
  MetricCardTitle,
  MetricCardValue,
} from "@/components/metric";
import {
  Folder,
  Key,
  Container,
  Activity,
  Plus,
  Clock,
} from "lucide-react";

export const DashboardPage: React.FC = () => {
  const { t } = useTranslation();
  const { user } = useAuth();
  const navigate = useNavigate();

  usePageTitle("common.pageTitle.dashboard");

  // Mock metrics data - in real app this would come from API
  const metrics = [
    {
      title: t("dashboard.metrics.totalProjects"),
      value: "12",
      href: ROUTES.projects,
      variant: "default" as const,
      icon: Folder,
    },
    {
      title: t("dashboard.metrics.activeEnvironments"),
      value: "8",
      href: "/environments",
      variant: "default" as const,
      icon: Container,
    },
    {
      title: t("dashboard.metrics.gitCredentials"),
      value: "5",
      href: ROUTES.gitCredentials,
      variant: "default" as const,
      icon: Key,
    },
    {
      title: t("dashboard.metrics.recentTasks"),
      value: "23",
      href: "/tasks",
      variant: "default" as const,
      icon: Activity,
    },
  ];

  const quickActions = [
    {
      title: t("dashboard.quickActions.newProject"),
      description: t("dashboard.quickActions.newProjectDesc"),
      icon: Folder,
      action: () => navigate("/projects/create"),
    },
    {
      title: t("dashboard.quickActions.newEnvironment"),
      description: t("dashboard.quickActions.newEnvironmentDesc"),
      icon: Container,
      action: () => navigate("/environments/create"),
    },
    {
      title: t("dashboard.quickActions.addCredential"),
      description: t("dashboard.quickActions.addCredentialDesc"),
      icon: Key,
      action: () => navigate("/credentials/create"),
    },
  ];

  return (
    <SectionGroup>
      <Section>
        <SectionHeader>
          <SectionTitle>{t("dashboard.title")}</SectionTitle>
          <SectionDescription>
            {t("dashboard.welcome")}, {user}! {t("dashboard.description")}
          </SectionDescription>
        </SectionHeader>
        <MetricCardGroup>
          {metrics.map((metric) => (
            <Link to={metric.href} key={metric.title}>
              <MetricCard variant={metric.variant}>
                <MetricCardHeader className="flex justify-between items-center gap-2">
                  <MetricCardTitle className="truncate">
                    {metric.title}
                  </MetricCardTitle>
                  <metric.icon className="size-4" />
                </MetricCardHeader>
                <MetricCardValue>{metric.value}</MetricCardValue>
              </MetricCard>
            </Link>
          ))}
        </MetricCardGroup>
      </Section>

      <Section>
        <SectionHeaderRow>
          <SectionHeader>
            <SectionTitle>{t("dashboard.quickActions.title")}</SectionTitle>
            <SectionDescription>
              {t("dashboard.quickActions.description")}
            </SectionDescription>
          </SectionHeader>
        </SectionHeaderRow>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {quickActions.map((action) => (
            <div
              key={action.title}
              className="border border-border rounded-lg p-4 hover:border-primary/50 transition-colors cursor-pointer"
              onClick={action.action}
            >
              <div className="flex items-center gap-3 mb-2">
                <div className="flex items-center justify-center w-8 h-8 rounded-lg bg-primary/10">
                  <action.icon className="size-4 text-primary" />
                </div>
                <h3 className="font-medium">{action.title}</h3>
              </div>
              <p className="text-sm text-muted-foreground mb-3">
                {action.description}
              </p>
              <Button size="sm" variant="outline" className="w-full">
                <Plus className="size-4 mr-2" />
                {t("common.create")}
              </Button>
            </div>
          ))}
        </div>
      </Section>

      <Section>
        <SectionHeaderRow>
          <SectionHeader>
            <SectionTitle>{t("dashboard.recentActivity.title")}</SectionTitle>
            <SectionDescription>
              {t("dashboard.recentActivity.description")}
            </SectionDescription>
          </SectionHeader>

        </SectionHeaderRow>
        <EmptyStateContainer className="min-h-[200px]">
          <Clock className="size-8 text-muted-foreground mb-2" />
          <EmptyStateTitle>{t("dashboard.recentActivity.empty")}</EmptyStateTitle>
          <EmptyStateDescription>
            {t("dashboard.recentActivity.emptyDesc")}
          </EmptyStateDescription>
        </EmptyStateContainer>
      </Section>


    </SectionGroup>
  );
};
