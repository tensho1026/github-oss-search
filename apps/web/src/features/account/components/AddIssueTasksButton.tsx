import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ListPlus } from "lucide-react";
import { useState } from "react";

import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "../../../components/ui/alert";
import { Button } from "../../../components/ui/button";
import { Icon } from "../../../components/ui/icon";
import { ApiError } from "../../../shared/api/client";
import { queryKeys } from "../../../shared/query/query-keys";
import { useI18n } from "../../../shared/i18n/i18n-context";
import type { CompareReference } from "../../issue-compare/model/compare-location";
import { useAuth } from "../../auth/auth-context";
import { upsertIssueClaim } from "../api/account";

export function AddIssueTasksButton({
  label,
  references,
  returnTo,
}: {
  label?: string;
  references: readonly CompareReference[];
  returnTo: string;
}) {
  const { t } = useI18n();
  const { markSessionExpired, session, signIn } = useAuth();
  const queryClient = useQueryClient();
  const [saved, setSaved] = useState(false);
  const mutation = useMutation({
    mutationFn: async () => {
      if (!session?.authenticated || !session.csrfToken) {
        throw new ApiError({
          code: "AUTHENTICATION_REQUIRED",
          message: "Authentication is required.",
          status: 401,
        });
      }
      await Promise.all(
        references.map((reference) =>
          upsertIssueClaim(
            {
              issueNumber: reference.issueNumber,
              repositoryName: reference.repository,
              repositoryOwner: reference.owner,
            },
            session.csrfToken!,
          ),
        ),
      );
    },
    async onError(error) {
      if (
        error instanceof ApiError &&
        error.code === "AUTHENTICATION_REQUIRED"
      ) {
        signIn(returnTo);
        return;
      }
      if (error instanceof ApiError && error.status === 401) {
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

  if (references.length === 0) {
    return null;
  }

  return (
    <div className="grid gap-2">
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
            : (label ?? t("selection.addTasks"))}
      </Button>
      {mutation.error &&
      !(
        mutation.error instanceof ApiError &&
        mutation.error.code === "AUTHENTICATION_REQUIRED"
      ) ? (
        <Alert variant="warning">
          <AlertTitle>{t("selection.error")}</AlertTitle>
          <AlertDescription>{t("selection.errorDescription")}</AlertDescription>
        </Alert>
      ) : null}
    </div>
  );
}
