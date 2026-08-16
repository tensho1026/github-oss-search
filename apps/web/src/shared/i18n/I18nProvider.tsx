import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";

import {
  I18nContext,
  supportedLocales,
  type Locale,
  type TranslationValues,
} from "./i18n-context";
import { enMessages, type MessageKey } from "./messages";

const localeStorageKey = "issuescout.locale";

function isLocale(value: unknown): value is Locale {
  return supportedLocales.includes(value as Locale);
}

function readStoredLocale(): Locale {
  try {
    const storedLocale = window.localStorage.getItem(localeStorageKey);
    return isLocale(storedLocale) ? storedLocale : "en";
  } catch {
    return "en";
  }
}

function interpolate(message: string, values?: TranslationValues): string {
  if (!values) {
    return message;
  }

  return message.replaceAll(/\{([^{}]+)\}/g, (placeholder, key: string) => {
    const value = values[key];
    return value === undefined ? placeholder : String(value);
  });
}

export function I18nProvider({ children }: PropsWithChildren) {
  const [locale, setActiveLocale] = useState<Locale>(readStoredLocale);
  const [messages, setMessages] =
    useState<Record<MessageKey, string>>(enMessages);

  useEffect(() => {
    let current = true;
    if (locale === "en") {
      return () => {
        current = false;
      };
    }

    void import("./ja-messages").then(({ jaMessages }) => {
      if (current) {
        setMessages(jaMessages);
      }
    });
    return () => {
      current = false;
    };
  }, [locale]);

  const setLocale = useCallback((nextLocale: Locale) => {
    if (nextLocale === "en") {
      setMessages(enMessages);
    }
    setActiveLocale(nextLocale);
  }, []);

  useEffect(() => {
    document.documentElement.lang = locale;
    try {
      window.localStorage.setItem(localeStorageKey, locale);
    } catch {
      // A denied storage write must never make localization unavailable.
    }
  }, [locale]);

  const value = useMemo(
    () => ({
      locale,
      setLocale,
      t: (key: MessageKey, values?: TranslationValues) =>
        interpolate(messages[key], values),
    }),
    [locale, messages, setLocale],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}
