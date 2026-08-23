import { keepPreviousData, useQuery } from "@tanstack/react-query";

import { searchGitHubRepositories } from "../../../shared/api/repositories";
import { queryKeys } from "../../../shared/query/query-keys";
import {
  createRelaxedRepositoryFilters,
  encodeRepositorySearchParams,
  toRepositoryDiscoveryRequest,
  type DecodedRepositoryLocation,
} from "../model/repository-filters";

export function useRepositoryDiscovery(location: DecodedRepositoryLocation) {
  const enabled = location.shouldSearch && location.valid;
  const canonicalSearch = enabled
    ? encodeRepositorySearchParams(location.filters).toString()
    : "disabled";

  return useQuery({
    enabled,
    placeholderData: keepPreviousData,
    queryFn: async ({ signal }) => {
      const envelope = await searchGitHubRepositories(
        toRepositoryDiscoveryRequest(location.filters),
        location.filters.page,
        location.filters.perPage,
        signal,
      );
      if (envelope.data.pagination.total > 0) {
        return [envelope, false] as const;
      }
      const relaxed = createRelaxedRepositoryFilters(location.filters);
      return [
        await searchGitHubRepositories(
          toRepositoryDiscoveryRequest(relaxed),
          relaxed.page,
          relaxed.perPage,
          signal,
        ),
        true,
      ] as const;
    },
    queryKey: queryKeys.repositories.search(canonicalSearch),
  });
}
