import { keepPreviousData, useQuery } from "@tanstack/react-query";

import { searchGitHubIssues } from "../../../shared/api/issues";
import { queryKeys } from "../../../shared/query/query-keys";
import {
  createRelaxedSearchFilters,
  encodeSearchParams,
  toIssueSearchRequest,
  type DecodedSearchLocation,
} from "../model/search-filters";

export function useIssueSearch(location: DecodedSearchLocation) {
  const enabled = location.shouldSearch && location.valid;
  const canonicalSearch = enabled
    ? encodeSearchParams(location.filters).toString()
    : "disabled";

  return useQuery({
    enabled,
    placeholderData: keepPreviousData,
    queryFn: async ({ signal }) => {
      const envelope = await searchGitHubIssues(
        toIssueSearchRequest(location.filters),
        location.filters.page,
        location.filters.perPage,
        signal,
      );
      if (envelope.data.pagination.total > 0) {
        return [envelope, false] as const;
      }
      const relaxed = createRelaxedSearchFilters(location.filters);
      return [
        await searchGitHubIssues(
          toIssueSearchRequest(relaxed),
          relaxed.page,
          relaxed.perPage,
          signal,
        ),
        true,
      ] as const;
    },
    queryKey: queryKeys.issues.search(canonicalSearch),
  });
}
