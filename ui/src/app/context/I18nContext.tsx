import React, { createContext, useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

/**
 * I18nContext 提供语言切换功能
 * 管理当前语言状态，与 react-i18next 集成
 */
interface I18nContextType {
  currentLanguage: string;
  availableLanguages: Array<{ code: string; name: string; nativeName: string }>;
  changeLanguage: (lang: string) => Promise<void>;
}

const I18nContext = createContext<I18nContextType | undefined>(undefined);

/**
 * I18nProvider 组件
 * 提供语言切换功能给子组件
 */
export function I18nProvider({ children }: { children: React.ReactNode }) {
  const { i18n } = useTranslation();
  const [currentLanguage, setCurrentLanguage] = useState(i18n.language || 'en');

  // 可用语言列表
  const availableLanguages = [
    { code: 'en', name: 'English', nativeName: 'English' },
    { code: 'zh', name: 'Chinese', nativeName: '中文' },
  ];

  // 监听语言变化
  useEffect(() => {
    const handleLanguageChanged = (lng: string) => {
      setCurrentLanguage(lng);
    };

    i18n.on('languageChanged', handleLanguageChanged);
    setCurrentLanguage(i18n.language || 'en');

    return () => {
      i18n.off('languageChanged', handleLanguageChanged);
    };
  }, [i18n]);

  /**
   * 切换语言
   * @param lang 语言代码 (en | zh)
   */
  const changeLanguage = async (lang: string) => {
    if (availableLanguages.some(l => l.code === lang)) {
      await i18n.changeLanguage(lang);
      setCurrentLanguage(lang);
    }
  };

  return (
    <I18nContext.Provider
      value={{
        currentLanguage,
        availableLanguages,
        changeLanguage,
      }}
    >
      {children}
    </I18nContext.Provider>
  );
}

/**
 * useI18n hook
 * 用于在组件中访问语言切换功能
 */
export function useI18n() {
  const context = useContext(I18nContext);
  if (context === undefined) {
    throw new Error('useI18n must be used within an I18nProvider');
  }
  return context;
}

