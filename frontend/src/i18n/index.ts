import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

// 引入翻译文件
import zhCN from './locales/zh-CN.json';
import enUS from './locales/en-US.json';

// 语言资源配置
const resources = {
  'zh-CN': {
    translation: zhCN
  },
  'en-US': {
    translation: enUS
  }
};

i18n
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

export default i18n; 