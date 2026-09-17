import { describe, expect, it } from "vitest";

import {
  discoveryIssueSearchHref,
  issueSearchFiltersFromDiscovery,
} from "./discovery-issue-search";
import { createDefaultRepositoryFilters } from "./repository-filters";
import {
  emptyRepositorySearchActions,
  hasAdvancedRepositoryFilters,
  repositoryFilterChips,
} from "./repository-filter-chips";

describe("repository discovery convenience models", () => {
  it("builds removable chips that reset pagination", () => {
    const filters = {
      ...createDefaultRepositoryFilters(),
      languages: ["Go"],
      minimumStars: 50,
      page: 3,
      technologies: ["React"],
    };
    const chips = repositoryFilterChips(filters);
    expect(chips.map((chip) => chip.id)).toEqual([
      "language:Go",
      "technology:React",
      "stars",
    ]);
    expect(chips[0]?.nextFilters).toMatchObject({
      languages: [],
      page: 1,
    });
  });

  it("treats non-default advanced fields as collapsed-form content", () => {
    expect(hasAdvancedRepositoryFilters(createDefaultRepositoryFilters())).toBe(
      false,
    );
    expect(
      hasAdvancedRepositoryFilters({
        ...createDefaultRepositoryFilters(),
        minimumStars: 50,
      }),
    ).toBe(true);
  });

  it("offers concrete empty-result searches from the current filters", () => {
    const actions = emptyRepositorySearchActions({
      ...createDefaultRepositoryFilters(),
      maximumDifficulty: 3,
      minimumStars: 10,
      technologies: ["Gin"],
      updatedWithinDays: 365,
    });
    expect(actions.map((action) => action.id)).toEqual([
      "raise-difficulty",
      "relax-recency",
      "clear-stars",
      "clear-technologies",
    ]);
    expect(actions[0]?.filters.maximumDifficulty).toBe(4);
    expect(actions[0]?.filters.page).toBe(1);
  });

  it("hands starter-issue searches the current discovery filters", () => {
    const filters = issueSearchFiltersFromDiscovery(
      {
        ...createDefaultRepositoryFilters(),
        excludeArchived: false,
        maximumDifficulty: 2,
        minimumStars: 25,
        updatedWithinDays: 90,
      },
      {
        labels: ["good first issue", "not-a-label"],
        language: "TypeScript",
        technologies: ["React"],
        username: "octocat",
      },
    );
    expect(filters).toMatchObject({
      excludeArchived: false,
      frameworks: ["React"],
      labels: ["good first issue"],
      languages: ["TypeScript"],
      maximumDifficulty: 2,
      minimumStars: 25,
      updatedWithinDays: 90,
      username: "octocat",
    });
    expect(
      discoveryIssueSearchHref(createDefaultRepositoryFilters(), {
        language: "Go",
        labels: ["good first issue"],
      }),
    ).toMatch(/^\/search\?.*language=Go.*label=good(\+|%20)first(\+|%20)issue/);
    expect(
      discoveryIssueSearchHref(createDefaultRepositoryFilters(), {
        language: "Go",
      }),
    ).not.toContain("search=1");
  });
});
