import { useQuery } from "@tanstack/react-query";

import { getPreferences } from "../../account/api/account";
import { useAuth } from "../../auth/auth-context";
import { queryKeys } from "../../../shared/query/query-keys";

export function usePreferredPageSize(fallback = 20): number {
  const { session } = useAuth();
  const query = useQuery({
    enabled: Boolean(session?.authenticated && session.csrfToken),
    queryFn: ({ signal }) => getPreferences(signal),
    queryKey: queryKeys.account.preferences,
  });
  return query.data?.data.resultsPerPage ?? fallback;
}
