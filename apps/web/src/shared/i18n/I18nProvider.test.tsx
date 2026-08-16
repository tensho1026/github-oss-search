import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it } from "vitest";

import { I18nProvider } from "./I18nProvider";
import { useI18n } from "./i18n-context";

function LocaleHarness() {
  const { locale, setLocale, t } = useI18n();

  return (
    <>
      <p>{locale}</p>
      <p>{t("nav.findIssues")}</p>
      <button onClick={() => setLocale("ja")} type="button">
        Japanese
      </button>
    </>
  );
}

afterEach(() => {
  window.localStorage.clear();
  document.documentElement.lang = "en";
});

describe("I18nProvider", () => {
  it("switches to Japanese, persists the locale, and updates document language", async () => {
    const user = userEvent.setup();
    render(
      <I18nProvider>
        <LocaleHarness />
      </I18nProvider>,
    );

    expect(screen.getByText("Find issues")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Japanese" }));

    expect(await screen.findByText("Issueを探す")).toBeInTheDocument();
    expect(document.documentElement).toHaveAttribute("lang", "ja");
    expect(window.localStorage.getItem("issuescout.locale")).toBe("ja");
  });

  it("restores a supported locale and rejects an invalid stored value", async () => {
    window.localStorage.setItem("issuescout.locale", "ja");
    const { unmount } = render(
      <I18nProvider>
        <LocaleHarness />
      </I18nProvider>,
    );

    expect(await screen.findByText("Issueを探す")).toBeInTheDocument();
    unmount();

    window.localStorage.setItem("issuescout.locale", "invalid");
    render(
      <I18nProvider>
        <LocaleHarness />
      </I18nProvider>,
    );

    expect(screen.getByText("Find issues")).toBeInTheDocument();
  });
});
