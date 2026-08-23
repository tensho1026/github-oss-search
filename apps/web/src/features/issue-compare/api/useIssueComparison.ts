import { useQueries } from "@tanstack/react-query";

import { getIssueDetail } from "../../../shared/api/issues";
import { queryKeys } from "../../../shared/query/query-keys";
import type { CompareLocation } from "../model/compare-location";

export function useIssueComparison(location: CompareLocation) {
  return useQueries({
    queries: location.references.map((reference) => ({
      enabled: location.valid,
      queryFn: ({ signal }: { signal: AbortSignal }) =>
        getIssueDetail(
          reference.owner,
          reference.repository,
          reference.issueNumber,
          location.skills,
          signal,
        ),
      queryKey: queryKeys.issues.detail(
        reference.owner,
        reference.repository,
        reference.issueNumber,
        location.skills,
      ),
    })),
  });
}
