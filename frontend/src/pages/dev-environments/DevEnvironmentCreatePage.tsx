import React, { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { usePageTitle } from "@/hooks/usePageTitle";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";
import DevEnvironmentForm from "@/components/DevEnvironmentForm";
import {
  EmptyStateContainer,
  EmptyStateTitle,
  EmptyStateDescription,
} from "@/components/content/empty-state";
import {
  Section,
  SectionGroup,
  SectionHeader,
  SectionTitle,
} from "@/components/content/section";
import type { DevEnvironmentDisplay } from "@/types/dev-environment";

const DevEnvironmentCreatePage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { setItems } = useBreadcrumb();

  usePageTitle(t("devEnvironments.create"));

  // Set breadcrumb navigation
  useEffect(() => {
    setItems([
      {
        type: "link",
        label: t("devEnvironments.list"),
        href: "/dev-environments",
      },
      {
        type: "page",
        label: t("devEnvironments.create"),
      },
    ]);

    // Cleanup when component unmounts
    return () => {
      setItems([]);
    };
  }, [setItems, t]);

  const handleSubmit = (_environment: DevEnvironmentDisplay) => {
    navigate("/dev-environments");
  };

  return (
    <SectionGroup>
      <Section>
        <SectionHeader>
          <SectionTitle>{t("devEnvironments.create")}</SectionTitle>
        </SectionHeader>
        <DevEnvironmentForm onSubmit={handleSubmit} />
      </Section>
      <Section>
        <EmptyStateContainer>
          <EmptyStateTitle>{t("devEnvironments.createAndCustomize")}</EmptyStateTitle>
          <EmptyStateDescription>
            {t("devEnvironments.createHelpText")}
          </EmptyStateDescription>
        </EmptyStateContainer>
      </Section>
    </SectionGroup>
  );
};

export default DevEnvironmentCreatePage;
