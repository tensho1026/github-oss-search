import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { profileAnalysisFixture } from "../../../test/profile-fixtures";
import { ProfileExtendedAnalytics } from "./ProfileExtendedAnalytics";

describe("ProfileExtendedAnalytics", () => {
  it("renders contribution, recency, proficiency, and repository evidence", () => {
    render(
      <ProfileExtendedAnalytics
        analysis={profileAnalysisFixture.data}
        showPortfolio
      />,
    );

    expect(
      screen.getByRole("heading", {
        name: "Private contribution portfolio preview",
      }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "View canonical PR" }),
    ).toHaveAttribute("href", "https://github.com/community/project/pull/42");
    expect(screen.getByText("Observed merged")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "OSS Journey" }),
    ).toBeInTheDocument();
    expect(
      screen.getByText("First observed Go contribution"),
    ).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: "View evidence" })).toHaveLength(
      3,
    );
    expect(
      screen.getByRole("heading", { name: "Contribution Streak" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Evidence" })).toHaveAttribute(
      "href",
      "https://github.com/community/project/pull/42",
    );
    expect(
      screen.getByRole("heading", { name: "OSS Quest" }),
    ).toBeInTheDocument();
    expect(screen.getByText(/Respond to review feedback/)).toBeVisible();

    expect(
      screen.getByRole("heading", { name: "Public contribution activity" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Public contribution calendar" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("img", {
        name: "2 public contributions on 2026-07-22",
      }),
    ).toBeInTheDocument();
    expect(
      screen.getByLabelText("Contribution intensity legend from less to more"),
    ).toHaveTextContent("LessMore");
    expect(screen.getByText("18")).toBeInTheDocument();
    expect(screen.getAllByText("Sampled").length).toBeGreaterThan(0);
    expect(
      screen.getByRole("heading", {
        name: "Recently observed technologies",
      }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("progressbar", {
        name: "TypeScript diagnostic level 2 of 5",
      }),
    ).toHaveAttribute("aria-valuenow", "2");
    expect(
      screen.getByRole("heading", { name: "Repository source evidence" }),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/starred totals can be privacy-ambiguous/i),
    ).toBeVisible();
  });

  it("distinguishes unavailable evidence from zero activity", () => {
    render(
      <ProfileExtendedAnalytics
        analysis={{
          ...profileAnalysisFixture.data,
          contributions: {
            ...profileAnalysisFixture.data.contributions,
            commits: { status: "unavailable", value: 0 },
          },
          contributionCalendar: {
            status: "unavailable",
            total: 0,
            weeks: [],
          },
          proficiency: [],
          recentTechnologies: [],
          repositoryEvidence: {
            ...profileAnalysisFixture.data.repositoryEvidence,
            starred: {
              ...profileAnalysisFixture.data.repositoryEvidence.starred,
              observed: 0,
              status: "unavailable",
            },
          },
        }}
      />,
    );

    expect(screen.getAllByText("Unavailable").length).toBeGreaterThan(1);
    expect(
      screen.getByText(/did not provide a public daily calendar/i),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/No five-level diagnostic could be supported/i),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/No recently observed technology evidence/i),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/GitHub did not provide this public segment/i),
    ).toBeInTheDocument();
  });
});
