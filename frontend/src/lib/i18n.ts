import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import en from './locales/en.json';
import sw from './locales/sw.json';

i18n
    .use(initReactI18next)
    .init({
        resources: {
            en: { translation: en },
            sw: { translation: sw },
        },
        lng: 'en',
        fallbackLng: 'en',
        supportedLngs: ['en', 'sw'],
        interpolation: {
            escapeValue: false,
        }
    });

export default i18n;
