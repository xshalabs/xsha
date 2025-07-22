import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

// 动态导入模块化语言文件
const loadModularTranslations = async (locale: string) => {
  try {
    // 动态导入各个模块的翻译文件
    const [
      common,
      auth,
      navigation,
      errors,
      api,
      dashboard,
      gitCredentials,
      projects,
      adminLogs,
      devEnvironments,
      tasks,
      taskConversations
    ] = await Promise.all([
      import(`./locales/${locale}/common.json`),
      import(`./locales/${locale}/auth.json`),
      import(`./locales/${locale}/navigation.json`),
      import(`./locales/${locale}/errors.json`),
      import(`./locales/${locale}/api.json`),
      import(`./locales/${locale}/dashboard.json`),
      import(`./locales/${locale}/git-credentials.json`),
      import(`./locales/${locale}/projects.json`),
      import(`./locales/${locale}/admin-logs.json`),
      import(`./locales/${locale}/dev-environments.json`),
      import(`./locales/${locale}/tasks.json`),
      import(`./locales/${locale}/task-conversations.json`)
    ]);

    // 合并所有模块的翻译内容
    return {
      common: common.default,
      auth: auth.default,
      navigation: navigation.default,
      errors: errors.default,
      validation: {
        required: errors.default.validation_required,
        invalid_format: errors.default.validation_invalid_format,
        too_short: errors.default.validation_too_short,
        too_long: errors.default.validation_too_long
      },
      api: api.default,
      health: {
        status_ok: api.default.health_status_ok
      },
      login: {
        success: auth.default.login_success,
        failed: auth.default.login_failed,
        invalid_request: auth.default.login_invalid_request,
        token_generate_error: auth.default.login_token_generate_error,
        rate_limit: auth.default.login_rate_limit
      },
      logout: {
        success: auth.default.logout_success,
        failed: auth.default.logout_failed,
        invalid_token: auth.default.logout_invalid_token,
        token_expired: auth.default.logout_token_expired
      },
      dashboard: dashboard.default,
      gitCredentials: gitCredentials.default,
      projects: projects.default,
      adminLogs: adminLogs.default,
      dev_environments: devEnvironments.default,
      tasks: tasks.default,
      taskConversation: taskConversations.default
    };
  } catch (error) {
    console.error(`Failed to load translations for locale: ${locale}`, error);
    throw error;
  }
};

// 初始化i18n
const initializeI18n = async () => {
  try {
    // 加载翻译资源
    const [zhCN, enUS] = await Promise.all([
      loadModularTranslations('zh-CN'),
      loadModularTranslations('en-US')
    ]);

    // 语言资源配置
    const resources = {
      'zh-CN': {
        translation: zhCN
      },
      'en-US': {
        translation: enUS
      }
    };

    await i18n
      .use(LanguageDetector) // 自动检测用户语言
      .use(initReactI18next) // 绑定react-i18next
      .init({
        resources,
        fallbackLng: 'zh-CN', // 默认语言
        
        // 语言检测配置
        detection: {
          order: ['localStorage', 'navigator', 'htmlTag'],
          caches: ['localStorage'],
          lookupLocalStorage: 'i18nextLng'
        },

        interpolation: {
          escapeValue: false, // React已经默认转义
        },

        // 调试模式（生产环境建议关闭）
        debug: import.meta.env.NODE_ENV === 'development',
      });

    console.log('✅ i18n initialized successfully with modular translations');
  } catch (error) {
    console.error('❌ Failed to initialize i18n:', error);
    throw error;
  }
};

// 导出初始化函数和i18n实例
export { initializeI18n };
export default i18n; 