import {
  createDefaultSearchFilters,
  type SearchFilters,
} from "./search-filters";

export type FilterChip<TFilters> = {
  id: string;
  label: string;
  nextFilters: TFilters;
};

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

export function issueSearchFilterChips(
  filters: SearchFilters,
  defaults: SearchFilters = createDefaultSearchFilters(filters.username),
): Array<FilterChip<SearchFilters>> {
  const chips: Array<FilterChip<SearchFilters>> = [];
  for (const language of filters.languages) {
    chips.push({
      id: `language:${language}`,
      label: language,
      nextFilters: withPageReset({
        ...filters,
        languages: filters.languages.filter((value) => value !== language),
      }),
    });
  }
  for (const framework of filters.frameworks) {
    chips.push({
      id: `framework:${framework}`,
      label: framework,
      nextFilters: withPageReset({
        ...filters,
        frameworks: filters.frameworks.filter((value) => value !== framework),
      }),
    });
  }
  for (const label of filters.labels) {
    chips.push({
      id: `label:${label}`,
      label,
      nextFilters: withPageReset({
        ...filters,
        labels: filters.labels.filter((value) => value !== label),
      }),
    });
  }
  if (filters.maximumEffort) {
    chips.push({
      id: "effort",
      label: effortLabels[filters.maximumEffort],
      nextFilters: withPageReset({ ...filters, maximumEffort: "" }),
    });
  }
  if (filters.maximumDifficulty !== defaults.maximumDifficulty) {
    chips.push({
      id: "difficulty",
      label: `≤ ${filters.maximumDifficulty}`,
      nextFilters: withPageReset({
        ...filters,
        maximumDifficulty: defaults.maximumDifficulty,
      }),
    });
  }
  if (filters.minimumStars !== defaults.minimumStars) {
    chips.push({
      id: "stars",
      label: `${filters.minimumStars}+ stars`,
      nextFilters: withPageReset({
        ...filters,
        minimumStars: defaults.minimumStars,
      }),
    });
  }
  if (filters.updatedWithinDays !== defaults.updatedWithinDays) {
    chips.push({
      id: "recency",
      label: `${filters.updatedWithinDays}d`,
      nextFilters: withPageReset({
        ...filters,
        updatedWithinDays: defaults.updatedWithinDays,
      }),
    });
  }
  if (filters.includeDocumentation) {
    chips.push({
      id: "documentation",
      label: "docs",
      nextFilters: withPageReset({
        ...filters,
        includeDocumentation: false,
      }),
    });
  }
  if (filters.includeStale) {
    chips.push({
      id: "stale",
      label: "stale",
      nextFilters: withPageReset({ ...filters, includeStale: false }),
    });
  }
  if (!filters.includeEnglish) {
    chips.push({
      id: "english",
      label: "non-English",
      nextFilters: withPageReset({ ...filters, includeEnglish: true }),
    });
  }
  if (!filters.excludeArchived) {
    chips.push({
      id: "archived",
      label: "archived",
      nextFilters: withPageReset({ ...filters, excludeArchived: true }),
    });
  }
  return chips;
}

export function hasAdvancedIssueFilters(
  filters: SearchFilters,
  defaults: SearchFilters = createDefaultSearchFilters(filters.username, {
    perPage: filters.perPage,
  }),
): boolean {
  return (
    filters.frameworks.length > 0 ||
    filters.minimumStars !== defaults.minimumStars ||
    filters.maximumDifficulty !== defaults.maximumDifficulty ||
    filters.perPage !== defaults.perPage ||
    filters.includeDocumentation ||
    filters.includeStale ||
    !filters.includeEnglish ||
    !filters.excludeArchived
  );
}

function withPageReset(filters: SearchFilters): SearchFilters {
  return { ...filters, page: 1 };
}
