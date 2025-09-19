import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import LanguageDetector from "i18next-browser-languagedetector";

const loadModularTranslations = async (locale: string) => {
  try {
    const [
      common,
      auth,
      navigation,
      errors,
      dashboard,
      gitCredentials,
      projects,
      admin,
      adminLogs,
      devEnvironments,
      tasks,
      taskConversations,
      gitDiff,
      systemConfig,
      user,
      notifiers,
    ] = await Promise.all([
      import(`./locales/${locale}/common.json`),
      import(`./locales/${locale}/auth.json`),
      import(`./locales/${locale}/navigation.json`),
      import(`./locales/${locale}/errors.json`),
      import(`./locales/${locale}/dashboard.json`),
      import(`./locales/${locale}/gitCredentials.json`),
      import(`./locales/${locale}/projects.json`),
      import(`./locales/${locale}/admin.json`),
      import(`./locales/${locale}/adminLogs.json`),
      import(`./locales/${locale}/devEnvironments.json`),
      import(`./locales/${locale}/tasks.json`),
      import(`./locales/${locale}/taskConversations.json`),
      import(`./locales/${locale}/gitDiff.json`),
      import(`./locales/${locale}/systemConfig.json`),
      import(`./locales/${locale}/user.json`),
      import(`./locales/${locale}/notifiers.json`),
    ]);

    return {
      common: common.default,
      auth: auth.default,
      navigation: navigation.default,
      errors: errors.default,
      dashboard: dashboard.default,
      gitCredentials: gitCredentials.default,
      projects: projects.default,
      admin: admin.default,
      adminLogs: adminLogs.default,
      devEnvironments: devEnvironments.default,
      tasks: tasks.default,
      taskConversations: taskConversations.default,
      gitDiff: gitDiff.default,
      systemConfig: systemConfig.default,
      user: user.default,
      notifiers: notifiers.default,
    };
  } catch (error) {
    console.error(`Failed to load translations for locale: ${locale}`, error);
    throw error;
  }
};

const initializeI18n = async () => {
  try {
    const [zhCN, enUS] = await Promise.all([
      loadModularTranslations("zh-CN"),
      loadModularTranslations("en-US"),
    ]);

    const resources = {
      "zh-CN": {
        translation: zhCN,
      },
      "en-US": {
        translation: enUS,
      },
    };

    await i18n
      .use(LanguageDetector)
      .use(initReactI18next)
      .init({
        resources,
        fallbackLng: "en-US",
        detection: {
          order: ["localStorage", "navigator", "htmlTag"],
          caches: ["localStorage"],
          lookupLocalStorage: "i18nextLng",
        },
        interpolation: {
          escapeValue: false,
        },
        debug: import.meta.env.NODE_ENV === "development",
      });

    console.log("✅ i18n initialized successfully with modular translations");
  } catch (error) {
    console.error("❌ Failed to initialize i18n:", error);
    throw error;
  }
};

export { initializeI18n };
export default i18n;
