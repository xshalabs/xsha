import { useEffect } from "react";
import { useTranslation } from "react-i18next";

export const usePageTitle = (titleKey: string, fallback?: string) => {
  const { t } = useTranslation();

  useEffect(() => {
    const title = t(titleKey, fallback || titleKey);
    const appName = t("common.app.name", "XSHA");

    document.title =
      title === titleKey
        ? fallback
          ? `${fallback} - ${appName}`
          : appName
        : `${title} - ${appName}`;

    return () => {
      document.title = appName;
    };
  }, [titleKey, fallback, t]);
};

export const useDirectPageTitle = (title: string) => {
  const { t } = useTranslation();

  useEffect(() => {
    const appName = t("common.app.name", "XSHA");
    document.title = `${title} - ${appName}`;

    return () => {
      document.title = appName;
    };
  }, [title, t]);
};
