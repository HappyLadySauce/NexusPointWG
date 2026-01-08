import i18n from 'i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import { initReactI18next } from 'react-i18next';

// Import translation resources
import enCommon from './locales/en/common.json';
import enDashboard from './locales/en/dashboard.json';
import enIPPools from './locales/en/ipPools.json';
import enLogin from './locales/en/login.json';
import enPeers from './locales/en/peers.json';
import enSettings from './locales/en/settings.json';
import enUsers from './locales/en/users.json';

import zhCommon from './locales/zh/common.json';
import zhDashboard from './locales/zh/dashboard.json';
import zhIPPools from './locales/zh/ipPools.json';
import zhLogin from './locales/zh/login.json';
import zhPeers from './locales/zh/peers.json';
import zhSettings from './locales/zh/settings.json';
import zhUsers from './locales/zh/users.json';

// Language detection configuration
const languageDetectorOptions = {
    // Order of detection methods
    order: ['localStorage', 'navigator', 'htmlTag'],

    // Keys to lookup language from
    lookupLocalStorage: 'i18nextLng',

    // Cache user language
    caches: ['localStorage'],

    // Only detect language, don't cache
    excludeCacheFor: ['cimode'],
};

// Normalize language code
const normalizeLanguage = (lng: string | undefined): string => {
    if (!lng) return 'en';

    // Map Chinese variants to 'zh'
    if (lng.startsWith('zh')) {
        return 'zh';
    }

    // Map English variants to 'en'
    if (lng.startsWith('en')) {
        return 'en';
    }

    // Default to English for other languages
    return 'en';
};

i18n
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
        // Resources
        resources: {
            en: {
                common: enCommon,
                dashboard: enDashboard,
                login: enLogin,
                peers: enPeers,
                users: enUsers,
                ipPools: enIPPools,
                settings: enSettings,
            },
            zh: {
                common: zhCommon,
                dashboard: zhDashboard,
                login: zhLogin,
                peers: zhPeers,
                users: zhUsers,
                ipPools: zhIPPools,
                settings: zhSettings,
            },
        },

        // Default namespace
        defaultNS: 'common',

        // Fallback namespace
        fallbackNS: 'common',

        // Supported languages
        supportedLngs: ['en', 'zh'],

        // Default language (will be overridden by detector)
        lng: 'en',

        // Fallback language
        fallbackLng: 'en',

        // Language detection options
        detection: languageDetectorOptions,

        // Interpolation options
        interpolation: {
            escapeValue: false, // React already escapes values
        },

        // React options
        react: {
            useSuspense: false,
        },

        // Normalize language code
        load: 'languageOnly',

        // Custom language normalization
        cleanCode: true,
    });

// Custom language normalization after detection
const detectedLng = i18n.language || 'en';
const normalizedLng = normalizeLanguage(detectedLng);
if (normalizedLng !== detectedLng) {
    i18n.changeLanguage(normalizedLng);
}

// Update HTML lang attribute when language changes
i18n.on('languageChanged', (lng) => {
    document.documentElement.lang = lng;
});

// Set initial HTML lang attribute
document.documentElement.lang = normalizedLng;

export default i18n;

