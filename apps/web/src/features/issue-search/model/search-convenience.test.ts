import { describe, expect, it } from "vitest";

import {
  createDefaultSearchFilters,
  decodeSearchParams,
} from "./search-filters";
import { emptyIssueSearchActions } from "./search-empty-actions";
import {
  hasAdvancedIssueFilters,
  issueSearchFilterChips,
} from "./search-filter-chips";
import { suggestedIssueSearchName } from "./saved-search-name";

describe("issue search convenience models", () => {
  it("uses a preference page size when the URL omits perPage", () => {
    const decoded = decodeSearchParams(
      new URLSearchParams("username=octocat"),
      {
        perPage: 10,
      },
    );
    expect(decoded.filters.perPage).toBe(10);
    expect(
      decodeSearchParams(new URLSearchParams("username=octocat&perPage=50"), {
        perPage: 10,
      }).filters.perPage,
    ).toBe(50);
  });

  it("builds removable chips that reset pagination", () => {
    const filters = {
      ...createDefaultSearchFilters("octocat"),
      frameworks: ["React"],
      labels: ["bug"],
      languages: ["Go"],
      maximumEffort: "two_hours" as const,
      minimumStars: 50,
      page: 3,
    };
    const chips = issueSearchFilterChips(filters);
    expect(chips.map((chip) => chip.id)).toEqual([
      "language:Go",
      "framework:React",
      "label:bug",
      "effort",
      "stars",
    ]);
    expect(chips[0]?.nextFilters).toMatchObject({
      languages: [],
      page: 1,
    });
  });

  it("treats non-default advanced fields as collapsed-form content", () => {
    expect(hasAdvancedIssueFilters(createDefaultSearchFilters("octocat"))).toBe(
      false,
    );
    expect(
      hasAdvancedIssueFilters({
        ...createDefaultSearchFilters("octocat"),
        includeStale: true,
      }),
    ).toBe(true);
  });

  it("offers concrete empty-result searches from the current filters", () => {
    const actions = emptyIssueSearchActions({
      ...createDefaultSearchFilters("octocat"),
      frameworks: ["Gin"],
      maximumDifficulty: 3,
      maximumEffort: "two_hours",
      minimumStars: 10,
      updatedWithinDays: 180,
    });
    expect(actions.map((action) => action.id)).toEqual([
      "raise-difficulty",
      "relax-recency",
      "clear-stars",
      "clear-effort",
    ]);
    expect(actions[0]?.filters.maximumDifficulty).toBe(4);
    expect(actions[0]?.filters.page).toBe(1);
  });

  it("suggests a saved-search name from the leading filters", () => {
    expect(
      suggestedIssueSearchName({
        ...createDefaultSearchFilters("octocat"),
        labels: ["good first issue"],
        languages: ["Go"],
        maximumEffort: "two_hours",
      }),
    ).toBe("Go · good first issue · 2h · octocat");
  });
});
