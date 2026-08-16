import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import { App } from "./App";

describe("App", () => {
  it("renders the IssueScout application shell", async () => {
    render(<App />);

    expect(
      await screen.findByRole("heading", {
        level: 1,
        name: /your next contribution, decoded/i,
      }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "IssueScout home" }),
    ).toBeInTheDocument();
  });

  it("switches the application shell to Japanese", async () => {
    const user = userEvent.setup();
    render(<App />);

    const [languageSwitcher] = await screen.findAllByRole("combobox", {
      name: "Language",
    });
    expect(languageSwitcher).toBeDefined();
    languageSwitcher!.focus();
    await user.keyboard("{ArrowDown}{ArrowDown}{Enter}");

    expect(
      await screen.findByRole("link", { name: "IssueScout ホーム" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Issueを探す" })).toBeVisible();
  });
});
