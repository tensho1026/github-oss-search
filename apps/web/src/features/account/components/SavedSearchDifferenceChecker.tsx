import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "../../../components/ui/alert";
import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { ApiError } from "../../../shared/api/client";
import type {
  IssueSearchRequest,
  RepositoryDiscoveryRequest,
  SavedSearch,
} from "../../../shared/api/generated";
import { searchGitHubIssues } from "../../../shared/api/issues";
import { searchGitHubRepositories } from "../../../shared/api/repositories";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { queryKeys } from "../../../shared/query/query-keys";
import { updateSavedSearchSnapshot } from "../api/account";
import { AccountRequestAlert } from "./AccountRequestAlert";

type Difference = { added: string[]; baseline: boolean; removed: string[] };

export function SavedSearchDifferenceChecker({
  csrfToken,
  onSessionExpired,
  search,
}: {
  csrfToken: string;
  onSessionExpired: () => Promise<void>;
  search: SavedSearch;
}) {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const [difference, setDifference] = useState<Difference>();
  const check = useMutation({
    mutationFn: async () => {
      const keys =
        search.searchType === "issue"
          ? (
              await searchGitHubIssues(
                search.filters as IssueSearchRequest,
                1,
                50,
              )
            ).data.items.map((item) =>
              `${item.repository.owner}/${item.repository.name}#${item.issue.number}`.toLowerCase(),
            )
          : (
              await searchGitHubRepositories(
                search.filters as RepositoryDiscoveryRequest,
                1,
                50,
              )
            ).data.items.map((item) => item.repository.fullName.toLowerCase());
      const current = [...new Set(keys)].sort();
      const previous = new Set(search.resultKeys);
      const currentSet = new Set(current);
      const result = {
        added: current.filter((key) => !previous.has(key)),
        baseline: !search.lastCheckedAt,
        removed: search.resultKeys.filter((key) => !currentSet.has(key)),
      };
      await updateSavedSearchSnapshot(
        search.id,
        { resultKeys: current, version: search.version },
        csrfToken,
      );
      return result;
    },
    async onError(error) {
      if (error instanceof ApiError && error.status === 401) {
        await onSessionExpired();
      }
    },
    async onSuccess(result) {
      setDifference(result);
      await queryClient.invalidateQueries({
        queryKey: queryKeys.account.savedSearches,
      });
    },
  });

  return (
    <div className="grid gap-3 lg:col-span-2">
      <div>
        <Button
          disabled={check.isPending}
          onClick={() => check.mutate()}
          size="small"
          variant="outline"
        >
          {check.isPending ? t("savedDiff.checking") : t("savedDiff.check")}
        </Button>
      </div>
      {difference ? (
        difference.baseline ? (
          <Alert variant="success">
            <AlertTitle>{t("savedDiff.baseline")}</AlertTitle>
            <AlertDescription>
              {t("savedDiff.baselineDescription")}
            </AlertDescription>
          </Alert>
        ) : (
          <Alert
            variant={
              difference.added.length || difference.removed.length
                ? "warning"
                : "success"
            }
          >
            <AlertTitle>
              {t("savedDiff.summary", {
                added: difference.added.length,
                removed: difference.removed.length,
              })}
            </AlertTitle>
            <AlertDescription>
              <DifferenceList
                label={t("savedDiff.added")}
                values={difference.added}
              />
              <DifferenceList
                label={t("savedDiff.removed")}
                values={difference.removed}
              />
            </AlertDescription>
          </Alert>
        )
      ) : null}
      {check.error ? <AccountRequestAlert error={check.error} /> : null}
    </div>
  );
}

function DifferenceList({
  label,
  values,
}: {
  label: string;
  values: string[];
}) {
  if (values.length === 0) return null;
  return (
    <div className="mt-2">
      <strong>{label}</strong>
      <div className="mt-1 flex flex-wrap gap-1">
        {values.map((value) => (
          <Badge key={value} variant="neutral">
            {value}
          </Badge>
        ))}
      </div>
    </div>
  );
}
