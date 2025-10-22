import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { usePageTitleContext } from "@/contexts/PageTitleContext";

export const usePageTitle = (titleKey: string, fallback?: string) => {
  const { t } = useTranslation();
  const { setPageTitle } = usePageTitleContext();

  useEffect(() => {
    const title = t(titleKey, fallback || titleKey);
    const appName = t("common.app.name", "XSHA");

    // Set browser document title
    document.title =
      title === titleKey
        ? fallback
          ? `${fallback} - ${appName}`
          : appName
        : `${title} - ${appName}`;

    // Set page title in context for SiteHeader
    setPageTitle(title === titleKey ? (fallback || titleKey) : title);

    return () => {
      document.title = appName;
      setPageTitle(null);
    };
  }, [titleKey, fallback, t, setPageTitle]);
};
