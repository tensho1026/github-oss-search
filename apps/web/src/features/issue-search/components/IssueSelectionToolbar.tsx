import { useMutation, useQueryClient } from "@tanstack/react-query";
import { GitCompareArrows, ListPlus, X } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router";

import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "../../../components/ui/alert";
import { Button } from "../../../components/ui/button";
import { Icon } from "../../../components/ui/icon";
import type { IssueSearchItem } from "../../../shared/api/generated";
import { appRoutes } from "../../../shared/config/app-config";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { queryKeys } from "../../../shared/query/query-keys";
import { useAuth } from "../../auth/auth-context";
import { upsertIssueClaim } from "../../account/api/account";
import { encodeCompareLocation } from "../../issue-compare/model/compare-location";

export function IssueSelectionToolbar({
  items,
  onClear,
  returnTo,
  skills,
}: {
  items: readonly IssueSearchItem[];
  onClear: () => void;
  returnTo: string;
  skills: readonly string[];
}) {
  const { t } = useI18n();
  const { markSessionExpired, session, signIn } = useAuth();
  const queryClient = useQueryClient();
  const [saved, setSaved] = useState(false);
  const references = items.map((item) => ({
    issueNumber: item.issue.number,
    owner: item.repository.owner,
    repository: item.repository.name,
  }));
  const compareParameters = encodeCompareLocation(references, skills, returnTo);
  const comparePath = `${appRoutes.compare}?${compareParameters.toString()}`;
  const mutation = useMutation({
    mutationFn: async () => {
      if (!session?.authenticated || !session.csrfToken) {
        throw new Error("AUTHENTICATION_REQUIRED");
      }
      await Promise.all(
        items.map((item) =>
          upsertIssueClaim(
            {
              issueNumber: item.issue.number,
              repositoryName: item.repository.name,
              repositoryOwner: item.repository.owner,
            },
            session.csrfToken!,
          ),
        ),
      );
    },
    async onError(error) {
      if (
        error instanceof Error &&
        error.message === "AUTHENTICATION_REQUIRED"
      ) {
        signIn(returnTo);
        return;
      }
      if (
        typeof error === "object" &&
        error &&
        "status" in error &&
        error.status === 401
      ) {
        await markSessionExpired();
      }
    },
    async onSuccess() {
      setSaved(true);
      await queryClient.invalidateQueries({
        queryKey: queryKeys.account.issueClaims,
      });
    },
  });

  if (items.length === 0) return null;
  return (
    <div className="sticky bottom-4 z-20 rounded-2xl border border-accent/40 bg-surface/95 p-4 shadow-lg backdrop-blur">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <strong>{t("selection.count", { count: items.length })}</strong>
        <div className="flex flex-wrap gap-2">
          <Button
            asChild={items.length >= 2}
            disabled={items.length < 2}
            size="small"
          >
            {items.length >= 2 ? (
              <Link to={comparePath}>
                <Icon icon={GitCompareArrows} />
                {t("selection.compare")}
              </Link>
            ) : (
              <span>
                <Icon icon={GitCompareArrows} />
                {t("selection.compare")}
              </span>
            )}
          </Button>
          <Button
            disabled={mutation.isPending || saved}
            onClick={() => mutation.mutate()}
            size="small"
            variant="outline"
          >
            <Icon icon={ListPlus} />
            {mutation.isPending
              ? t("selection.adding")
              : saved
                ? t("selection.added")
                : t("selection.addTasks")}
          </Button>
          <Button
            aria-label={t("selection.clear")}
            onClick={onClear}
            size="small"
            variant="ghost"
          >
            <Icon icon={X} />
          </Button>
        </div>
      </div>
      {items.length < 2 ? (
        <p className="mt-2 text-xs text-muted-foreground">
          {t("selection.compareHint")}
        </p>
      ) : null}
      {mutation.error &&
      mutation.error instanceof Error &&
      mutation.error.message !== "AUTHENTICATION_REQUIRED" ? (
        <Alert className="mt-3" variant="warning">
          <AlertTitle>{t("selection.error")}</AlertTitle>
          <AlertDescription>{t("selection.errorDescription")}</AlertDescription>
        </Alert>
      ) : null}
    </div>
  );
}
