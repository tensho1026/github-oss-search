import type { SearchFilters } from "./search-filters";

export const emptyIssueSearchActionIds = [
  "raise-difficulty",
  "relax-recency",
  "clear-stars",
  "clear-effort",
  "clear-frameworks",
] as const;

export type EmptyIssueSearchActionId =
  (typeof emptyIssueSearchActionIds)[number];

export type EmptyIssueSearchAction = {
  filters: SearchFilters;
  id: EmptyIssueSearchActionId;
};

export function emptyIssueSearchActions(
  filters: SearchFilters,
): EmptyIssueSearchAction[] {
  const actions: EmptyIssueSearchAction[] = [];
  if (filters.maximumDifficulty < 5) {
    actions.push({
      id: "raise-difficulty",
      filters: withPageReset({
        ...filters,
        maximumDifficulty: filters.maximumDifficulty + 1,
      }),
    });
  }
  if (filters.updatedWithinDays < 365) {
    actions.push({
      id: "relax-recency",
      filters: withPageReset({ ...filters, updatedWithinDays: 365 }),
    });
  } else if (filters.updatedWithinDays < 3650) {
    actions.push({
      id: "relax-recency",
      filters: withPageReset({ ...filters, updatedWithinDays: 3650 }),
    });
  }
  if (filters.minimumStars > 0) {
    actions.push({
      id: "clear-stars",
      filters: withPageReset({ ...filters, minimumStars: 0 }),
    });
  }
  if (filters.maximumEffort) {
    actions.push({
      id: "clear-effort",
      filters: withPageReset({ ...filters, maximumEffort: "" }),
    });
  }
  if (filters.frameworks.length > 0) {
    actions.push({
      id: "clear-frameworks",
      filters: withPageReset({ ...filters, frameworks: [] }),
    });
  }
  return actions.slice(0, 4);
}

function withPageReset(filters: SearchFilters): SearchFilters {
  return { ...filters, page: 1 };
}
