import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { AppProviders } from "../../app/AppProviders";
import { CopyLinkButton } from "./copy-link-button";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("CopyLinkButton", () => {
  it("copies the provided href", async () => {
    const user = userEvent.setup();
    const writeText = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal("navigator", {
      ...navigator,
      clipboard: { writeText },
    });
    render(
      <AppProviders>
        <CopyLinkButton href="http://127.0.0.1:5173/search?search=1" />
      </AppProviders>,
    );

    await user.click(screen.getByRole("button", { name: "Copy search link" }));
    expect(writeText).toHaveBeenCalledWith(
      "http://127.0.0.1:5173/search?search=1",
    );
    expect(
      await screen.findByRole("button", { name: "Link copied" }),
    ).toBeInTheDocument();
  });
});
