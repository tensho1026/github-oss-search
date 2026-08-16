import { createContext, useContext } from "react";

import { enMessages, type MessageKey } from "./messages";

export const supportedLocales = ["en", "ja"] as const;

export type Locale = (typeof supportedLocales)[number];
export type TranslationValues = Readonly<Record<string, number | string>>;

export interface I18nContextValue {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: MessageKey, values?: TranslationValues) => string;
}

function translateEnglish(key: MessageKey, values?: TranslationValues): string {
  const message = enMessages[key];
  if (!values) {
    return message;
  }

  return message.replaceAll(/\{([^{}]+)\}/g, (placeholder, name: string) => {
    const value = values[name];
    return value === undefined ? placeholder : String(value);
  });
}

const englishFallback: I18nContextValue = {
  locale: "en",
  setLocale: () => undefined,
  t: translateEnglish,
};

export const I18nContext = createContext<I18nContextValue>(englishFallback);

export function useI18n(): I18nContextValue {
  return useContext(I18nContext);
}
