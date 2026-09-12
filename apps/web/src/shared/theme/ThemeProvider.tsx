import { useEffect, useMemo, useState, type PropsWithChildren } from "react";

import type { ThemePreference } from "../api/generated";
import { ThemeContext } from "./theme-context";

const storageKey = "issuescout.theme";
function systemPrefersDark(): boolean {
  return (
    typeof window.matchMedia === "function" &&
    window.matchMedia("(prefers-color-scheme: dark)").matches
  );
}

function storedPreference(): ThemePreference {
  try {
    const value = window.localStorage.getItem(storageKey);
    if (value === "dark" || value === "light" || value === "system") {
      return value;
    }
  } catch {
    // Storage can be unavailable in privacy-restricted browser contexts.
  }
  return "system";
}

export function ThemeProvider({ children }: PropsWithChildren) {
  const [preference, setPreference] =
    useState<ThemePreference>(storedPreference);
  const [prefersDark, setPrefersDark] = useState(systemPrefersDark);
  const resolvedTheme =
    preference === "system" ? (prefersDark ? "dark" : "light") : preference;

  useEffect(() => {
    if (typeof window.matchMedia !== "function") return;
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const update = (event: MediaQueryListEvent) =>
      setPrefersDark(event.matches);
    media.addEventListener?.("change", update);
    return () => media.removeEventListener?.("change", update);
  }, []);

  useEffect(() => {
    document.documentElement.dataset.theme = resolvedTheme;
    try {
      window.localStorage.setItem(storageKey, preference);
    } catch {
      // The in-memory selection still works when persistence is unavailable.
    }
  }, [preference, resolvedTheme]);

  const value = useMemo(
    () => ({ preference, resolvedTheme, setPreference }),
    [preference, resolvedTheme],
  );

  return (
    <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
  );
}
