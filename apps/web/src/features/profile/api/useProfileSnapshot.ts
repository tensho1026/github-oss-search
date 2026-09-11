import { useQuery } from "@tanstack/react-query";
import { useCallback } from "react";

import { getProfileSnapshot } from "../../../shared/api/profile";
import type {
  GitHubUser,
  Meta,
  ProfileAnalysis,
} from "../../../shared/api/generated";
import { queryKeys } from "../../../shared/query/query-keys";
import { prioritizedProfileError } from "../model/profile-error";

export type ProfileSnapshot = {
  analysis: ProfileAnalysis;
  analysisMeta: Meta;
  user: GitHubUser;
  userMeta: Meta;
};

export function useProfileSnapshot(username: string, enabled = true) {
  const query = useQuery({
    enabled,
    queryFn: ({ signal }) => getProfileSnapshot(username, signal),
    queryKey: queryKeys.profile.snapshot(username),
  });

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  const snapshot = query.data
    ? {
        analysis: query.data.data.analysis,
        analysisMeta: query.data.meta,
        user: query.data.data.user,
        userMeta: query.data.meta,
      }
    : undefined;

  return {
    error: prioritizedProfileError([query.error]),
    isFetching: query.isFetching,
    isPending: query.isPending,
    refetch,
    snapshot,
  };
}
