import {
  createDefaultRepositoryFilters,
  type RepositoryFilters,
} from "./repository-filters";
import type { FilterChip } from "../../issue-search/model/search-filter-chips";

export function repositoryFilterChips(
  filters: RepositoryFilters,
  defaults: RepositoryFilters = createDefaultRepositoryFilters({
    perPage: filters.perPage,
  }),
): Array<FilterChip<RepositoryFilters>> {
  const chips: Array<FilterChip<RepositoryFilters>> = [];
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
  for (const technology of filters.technologies) {
    chips.push({
      id: `technology:${technology}`,
      label: technology,
      nextFilters: withPageReset({
        ...filters,
        technologies: filters.technologies.filter(
          (value) => value !== technology,
        ),
      }),
    });
  }
  for (const license of filters.licenses) {
    chips.push({
      id: `license:${license}`,
      label: license,
      nextFilters: withPageReset({
        ...filters,
        licenses: filters.licenses.filter((value) => value !== license),
      }),
    });
  }
  for (const category of filters.categories) {
    chips.push({
      id: `category:${category}`,
      label: category,
      nextFilters: withPageReset({
        ...filters,
        categories: filters.categories.filter((value) => value !== category),
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
  if (!filters.excludeArchived) {
    chips.push({
      id: "archived",
      label: "archived",
      nextFilters: withPageReset({ ...filters, excludeArchived: true }),
    });
  }
  return chips;
}

export function hasAdvancedRepositoryFilters(
  filters: RepositoryFilters,
  defaults: RepositoryFilters = createDefaultRepositoryFilters({
    perPage: filters.perPage,
  }),
): boolean {
  return (
    filters.licenses.length > 0 ||
    filters.categories.length > 0 ||
    filters.minimumStars !== defaults.minimumStars ||
    filters.minimumForks !== defaults.minimumForks ||
    filters.minimumOpenIssues !== defaults.minimumOpenIssues ||
    filters.maximumOpenIssues !== defaults.maximumOpenIssues ||
    filters.minimumReadiness !== defaults.minimumReadiness ||
    filters.maximumDifficulty !== defaults.maximumDifficulty ||
    filters.perPage !== defaults.perPage ||
    filters.forkPolicy !== defaults.forkPolicy ||
    filters.hasJapaneseReadme !== defaults.hasJapaneseReadme ||
    !filters.excludeArchived
  );
}

export const emptyRepositorySearchActionIds = [
  "raise-difficulty",
  "relax-recency",
  "clear-stars",
  "clear-technologies",
] as const;

export type EmptyRepositorySearchActionId =
  (typeof emptyRepositorySearchActionIds)[number];

export type EmptyRepositorySearchAction = {
  filters: RepositoryFilters;
  id: EmptyRepositorySearchActionId;
};

export function emptyRepositorySearchActions(
  filters: RepositoryFilters,
): EmptyRepositorySearchAction[] {
  const actions: EmptyRepositorySearchAction[] = [];
  if (filters.maximumDifficulty < 5) {
    actions.push({
      id: "raise-difficulty",
      filters: withPageReset({
        ...filters,
        maximumDifficulty: filters.maximumDifficulty + 1,
      }),
    });
  }
  if (filters.updatedWithinDays < 3650) {
    actions.push({
      id: "relax-recency",
      filters: withPageReset({
        ...filters,
        updatedWithinDays: Math.min(3650, filters.updatedWithinDays * 2),
      }),
    });
  }
  if (filters.minimumStars > 0) {
    actions.push({
      id: "clear-stars",
      filters: withPageReset({ ...filters, minimumStars: 0 }),
    });
  }
  if (filters.technologies.length > 0) {
    actions.push({
      id: "clear-technologies",
      filters: withPageReset({ ...filters, technologies: [] }),
    });
  }
  return actions.slice(0, 4);
}

function withPageReset(filters: RepositoryFilters): RepositoryFilters {
  return { ...filters, page: 1 };
}
