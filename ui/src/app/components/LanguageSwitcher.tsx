import React from 'react';
import { Globe } from 'lucide-react';
import { useI18n } from '../context/I18nContext';
import { useTranslation } from 'react-i18next';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select';
import { Label } from './ui/label';

/**
 * LanguageSwitcher 组件
 * 提供语言切换功能的下拉选择器
 */
export function LanguageSwitcher() {
  const { currentLanguage, availableLanguages, changeLanguage } = useI18n();
  const { t } = useTranslation('settings');

  const handleLanguageChange = (value: string) => {
    changeLanguage(value);
  };

  return (
    <div className="space-y-2">
      <Label htmlFor="language-select">{t('language.title')}</Label>
      <Select value={currentLanguage} onValueChange={handleLanguageChange}>
        <SelectTrigger id="language-select" className="w-full">
          <div className="flex items-center gap-2">
            <Globe className="h-4 w-4 text-muted-foreground" />
            <SelectValue>
              {availableLanguages.find(lang => lang.code === currentLanguage)?.nativeName || currentLanguage}
            </SelectValue>
          </div>
        </SelectTrigger>
        <SelectContent>
          {availableLanguages.map((lang) => (
            <SelectItem key={lang.code} value={lang.code}>
              {lang.nativeName}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <p className="text-xs text-muted-foreground">{t('language.description')}</p>
    </div>
  );
}

