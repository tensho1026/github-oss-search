import { Menu, MoonStar, Sun } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, Outlet } from "react-router";

import { appConfig, appRoutes } from "../../shared/config/app-config";
import { useI18n } from "../../shared/i18n/i18n-context";
import { AccountControl } from "../../features/auth/components/AccountControl";
import { AuthFeedback } from "../../features/auth/components/AuthFeedback";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Icon } from "../ui/icon";
import { LanguageSwitcher } from "./LanguageSwitcher";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";

type Theme = "dark" | "light";

function preferredTheme(): Theme {
  return typeof window.matchMedia === "function" &&
    window.matchMedia("(prefers-color-scheme: light)").matches
    ? "light"
    : "dark";
}

function Brand() {
  const { t } = useI18n();

  return (
    <Link
      aria-label={t("brand.home", { productName: appConfig.productName })}
      className="inline-flex shrink-0 items-center gap-3 whitespace-nowrap rounded-lg font-semibold tracking-[-0.02em] text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring"
      to={appRoutes.home}
    >
      <span
        aria-hidden="true"
        className="relative grid size-9 place-items-center overflow-hidden rounded-xl border border-accent/25 bg-accent-soft text-xs font-bold text-accent-soft-foreground shadow-sm"
      >
        <span className="absolute -top-3 -right-2 size-7 rounded-full bg-accent/20" />
        IS
      </span>
      <span className="hidden min-[400px]:inline">{appConfig.productName}</span>
    </Link>
  );
}

function ThemeToggle({
  onChange,
  theme,
}: {
  onChange: () => void;
  theme: Theme;
}) {
  const { t } = useI18n();
  const nextTheme = theme === "dark" ? "light" : "dark";
  const label = t(nextTheme === "light" ? "theme.useLight" : "theme.useDark");
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          aria-label={label}
          onClick={onChange}
          size="icon"
          variant="ghost"
        >
          <Icon icon={theme === "dark" ? Sun : MoonStar} />
        </Button>
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  );
}

function NavigationLinks() {
  const { t } = useI18n();

  return (
    <>
      <Link
        className="rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground outline-none transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        to="/#how-it-works"
      >
        {t("nav.howItWorks")}
      </Link>
      <Link
        className="rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground outline-none transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        to={appRoutes.search}
      >
        {t("nav.findIssues")}
      </Link>
      <Link
        className="rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground outline-none transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        to={appRoutes.repositories}
      >
        {t("nav.discoverRepositories")}
      </Link>
      <Button asChild size="small" variant="outline">
        <Link to="/#analyze">{t("nav.analyzeProfile")}</Link>
      </Button>
    </>
  );
}

export function AppShell() {
  const { t } = useI18n();
  const [theme, setTheme] = useState<Theme>(preferredTheme);

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
  }, [theme]);

  return (
    <div className="app-backdrop min-h-screen bg-background text-foreground">
      <a
        className="fixed top-3 left-3 z-[100] -translate-y-20 rounded-lg bg-foreground px-4 py-2 text-sm font-semibold text-background transition-transform focus:translate-y-0"
        href="#main-content"
      >
        {t("a11y.skipToContent")}
      </a>
      <header className="sticky top-0 z-40 border-b border-border/75 bg-background/78 backdrop-blur-xl">
        <div className="mx-auto flex h-18 w-full max-w-7xl items-center justify-between px-5 sm:px-8 lg:px-10">
          <Brand />
          <nav
            aria-label={t("nav.primary")}
            className="hidden items-center gap-1 whitespace-nowrap xl:flex"
          >
            <NavigationLinks />
            <LanguageSwitcher />
            <ThemeToggle
              onChange={() =>
                setTheme((current) => (current === "dark" ? "light" : "dark"))
              }
              theme={theme}
            />
            <AccountControl />
          </nav>
          <div className="flex min-w-0 items-center gap-1 xl:hidden">
            <LanguageSwitcher compact />
            <AccountControl />
            <ThemeToggle
              onChange={() =>
                setTheme((current) => (current === "dark" ? "light" : "dark"))
              }
              theme={theme}
            />
            <Dialog>
              <DialogTrigger asChild>
                <Button aria-label={t("nav.open")} size="icon" variant="ghost">
                  <Icon icon={Menu} />
                </Button>
              </DialogTrigger>
              <DialogContent className="top-4 right-4 left-auto w-[min(88vw,22rem)] translate-x-0 translate-y-0">
                <DialogHeader>
                  <DialogTitle>{t("nav.dialogTitle")}</DialogTitle>
                  <DialogDescription>
                    {t("nav.dialogDescription")}
                  </DialogDescription>
                </DialogHeader>
                <nav
                  aria-label={t("nav.mobile")}
                  className="grid items-stretch gap-2"
                >
                  <NavigationLinks />
                </nav>
              </DialogContent>
            </Dialog>
          </div>
        </div>
      </header>
      <AuthFeedback />
      <main id="main-content">
        <Outlet />
      </main>
      <footer className="border-t border-border/75">
        <div className="mx-auto flex w-full max-w-7xl flex-col gap-3 px-5 py-8 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between sm:px-8 lg:px-10">
          <p>{t("footer.evidence")}</p>
          <p>{t("footer.anonymous")}</p>
        </div>
      </footer>
    </div>
  );
}
