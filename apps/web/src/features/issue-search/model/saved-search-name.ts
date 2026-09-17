import type { SearchFilters } from "./search-filters";
import type { RepositoryFilters } from "../../repository-discovery/model/repository-filters";

const maximumNameLength = 80;

const effortLabels: Record<
  Exclude<SearchFilters["maximumEffort"], "">,
  string
> = {
  half_day: "half day",
  one_day: "1 day",
  thirty_minutes: "30m",
  three_days: "3 days",
  two_hours: "2h",
};

export function suggestedIssueSearchName(filters: SearchFilters): string {
  return joinName([
    filters.languages[0],
    filters.labels[0],
    filters.maximumEffort ? effortLabels[filters.maximumEffort] : undefined,
    filters.username,
  ]);
}

export function suggestedRepositorySearchName(
  filters: RepositoryFilters,
): string {
  return joinName([
    filters.languages[0],
    filters.technologies[0],
    filters.licenses[0],
    `${filters.minimumStars}+ stars`,
  ]);
}

export function duplicatedSavedSearchName(name: string): string {
  const suffix = " copy";
  const base = name.trim() || "Saved search";
  if (`${base}${suffix}`.length <= maximumNameLength) {
    return `${base}${suffix}`;
  }
  return `${base.slice(0, maximumNameLength - suffix.length)}${suffix}`;
}

function joinName(parts: Array<string | undefined>): string {
  const name = parts.filter((part) => Boolean(part?.trim())).join(" · ");
  return name.slice(0, maximumNameLength) || "Saved search";
}
