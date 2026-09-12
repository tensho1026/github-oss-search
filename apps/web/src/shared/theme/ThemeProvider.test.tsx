import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ThemeProvider } from "./ThemeProvider";
import { useTheme } from "./theme-context";

function Probe() {
  const { preference, resolvedTheme, setPreference } = useTheme();
  return (
    <>
      <p>{`${preference}:${resolvedTheme}`}</p>
      <button onClick={() => setPreference("light")} type="button">
        Use light
      </button>
    </>
  );
}

describe("ThemeProvider", () => {
  beforeEach(() => {
    window.localStorage.clear();
    document.documentElement.removeAttribute("data-theme");
  });

  it("persists anonymous choices and follows OS changes only in system mode", async () => {
    let listener: ((event: { matches: boolean }) => void) | undefined;
    vi.stubGlobal(
      "matchMedia",
      vi.fn(() => ({
        addEventListener: (_name: string, next: typeof listener) => {
          listener = next;
        },
        matches: false,
        removeEventListener: vi.fn(),
      })),
    );
    const user = userEvent.setup();
    render(
      <ThemeProvider>
        <Probe />
      </ThemeProvider>,
    );

    expect(screen.getByText("system:light")).toBeInTheDocument();
    act(() => listener?.({ matches: true }));
    expect(screen.getByText("system:dark")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Use light" }));
    act(() => listener?.({ matches: false }));
    expect(screen.getByText("light:light")).toBeInTheDocument();
    expect(document.documentElement.dataset.theme).toBe("light");
    expect(window.localStorage.getItem("issuescout.theme")).toBe("light");
  });
});
