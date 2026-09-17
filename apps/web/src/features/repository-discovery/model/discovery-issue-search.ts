import {
  createDefaultSearchFilters,
  encodeSearchParams,
  searchFilterOptions,
  type SearchFilters,
} from "../../issue-search/model/search-filters";
import type { RepositoryFilters } from "./repository-filters";

const allowedFrameworks = new Set(
  searchFilterOptions.frameworks.map((option) => option.value),
);
const allowedLabels = new Set(
  searchFilterOptions.labels.map((option) => option.value),
);
const allowedLanguages = new Set(
  searchFilterOptions.languages.map((option) => option.value),
);

export function issueSearchFiltersFromDiscovery(
  filters: RepositoryFilters,
  options: {
    labels?: readonly string[];
    language?: string | null;
    technologies?: readonly string[];
    username?: string;
  } = {},
): SearchFilters {
  const languages = (
    options.language && allowedLanguages.has(options.language)
      ? [options.language]
      : filters.languages
  ).filter((value) => allowedLanguages.has(value));
  const technologySource =
    filters.technologies.length > 0
      ? filters.technologies
      : [...(options.technologies ?? [])];
  const frameworks = technologySource.filter((value) =>
    allowedFrameworks.has(value),
  );
  const labels = (options.labels ?? []).filter((value) =>
    allowedLabels.has(value),
  );
  const defaults = createDefaultSearchFilters(options.username ?? "");
  return {
    ...defaults,
    excludeArchived: filters.excludeArchived,
    frameworks,
    labels: labels.length > 0 ? labels : defaults.labels,
    languages,
    maximumDifficulty: filters.maximumDifficulty,
    minimumStars: filters.minimumStars,
    updatedWithinDays: filters.updatedWithinDays,
  };
}

export function discoveryIssueSearchHref(
  filters: RepositoryFilters,
  options: {
    labels?: readonly string[];
    language?: string | null;
    technologies?: readonly string[];
    username?: string;
  } = {},
): string {
  const searchFilters = issueSearchFiltersFromDiscovery(filters, options);
  return `/search?${encodeSearchParams(
    searchFilters,
    Boolean(searchFilters.username.trim()),
  ).toString()}`;
}
