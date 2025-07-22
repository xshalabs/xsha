import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';

/**
 * 自定义Hook用于设置页面标题
 * @param titleKey 翻译key，用于获取页面标题
 * @param fallback 后备标题，当翻译不存在时使用
 */
export const usePageTitle = (titleKey: string, fallback?: string) => {
  const { t } = useTranslation();

  useEffect(() => {
    // 如果是pageTitle相关的key，自动加上common前缀
    const fullTitleKey = titleKey.startsWith('pageTitle.') ? `common.${titleKey}` : titleKey;
    const title = t(fullTitleKey, fallback || titleKey);
    const appName = t('common.app.name', 'Sleep0');
    
    // 设置页面标题格式：页面标题 - 应用名称
    document.title = title === titleKey ? 
      (fallback ? `${fallback} - ${appName}` : appName) : 
      `${title} - ${appName}`;

    // 清理函数，组件卸载时恢复默认标题
    return () => {
      document.title = appName;
    };
  }, [titleKey, fallback, t]);
};

/**
 * 设置页面标题的简化版本，直接传入标题文本
 * @param title 页面标题文本
 */
export const useDirectPageTitle = (title: string) => {
  const { t } = useTranslation();

  useEffect(() => {
    const appName = t('common.app.name', 'Sleep0');
    document.title = `${title} - ${appName}`;

    return () => {
      document.title = appName;
    };
  }, [title, t]);
}; 